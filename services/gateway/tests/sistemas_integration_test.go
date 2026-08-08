//go:build integration

// Teste ponta-a-ponta das rotas de sistemas: Gateway → Design Engine (gRPC) →
// Postgres, com a cadeia real de autenticação (JWT → TenantContext → Metadata) e
// a RLS efetiva (pool sob role não-superusuário). Requer o Postgres do
// docker-compose com as migrações aplicadas. Executar:
//
//	DATABASE_URL=postgres://mach:mach@localhost:5432/machv4?sslmode=disable \
//	  go test -tags integration ./services/gateway/tests/...
package tests

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	designv1 "github.com/machv4/platform/gen/go/construtor/design/v1"
	iamv1 "github.com/machv4/platform/gen/go/construtor/iam/v1"
	gatewayapp "github.com/machv4/platform/services/gateway/internal/app"
	"github.com/machv4/platform/services/gateway/internal/middleware"
	"github.com/machv4/platform/pkg/tenantctx"
	designapp "github.com/machv4/platform/services/design/app"
	iamapp "github.com/machv4/platform/services/iam/app"
	"github.com/machv4/platform/services/iam/auth"
)

const sisAppRole = "machapp_gw_sis_rls_test"

// sistemasHarness monta Postgres + IAM + Design Engine (ambos em bufconn) e o
// Gateway, devolvendo o handler HTTP, um emissor de JWT e os ids de dois tenants.
func sistemasHarness(t *testing.T) (http.Handler, *auth.Issuer, string, string) {
	t.Helper()
	ctx := context.Background()

	admin, err := pgx.Connect(ctx, dsn())
	if err != nil {
		t.Skipf("Postgres indisponível (%v)", err)
	}
	defer admin.Close(ctx)

	for _, stmt := range []string{
		`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname='` + sisAppRole + `') THEN CREATE ROLE ` + sisAppRole + ` NOLOGIN NOSUPERUSER; END IF; END $$;`,
		`GRANT USAGE ON SCHEMA public TO ` + sisAppRole + `;`,
		`GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO ` + sisAppRole + `;`,
	} {
		if _, err := admin.Exec(ctx, stmt); err != nil {
			t.Fatalf("setup role: %v (%s)", err, stmt)
		}
	}

	var idA, idB string
	if err := admin.QueryRow(ctx, `INSERT INTO tenants (nome, tipo, chave_blind_index) VALUES ('itg-sis-A','dono','\x6b41') RETURNING id`).Scan(&idA); err != nil {
		t.Fatalf("tenant A: %v", err)
	}
	if err := admin.QueryRow(ctx, `INSERT INTO tenants (nome, tipo, chave_blind_index) VALUES ('itg-sis-B','dono','\x6b42') RETURNING id`).Scan(&idB); err != nil {
		t.Fatalf("tenant B: %v", err)
	}
	t.Cleanup(func() {
		c, err := pgx.Connect(context.Background(), dsn())
		if err != nil {
			return
		}
		defer c.Close(context.Background())
		_, _ = c.Exec(context.Background(), `DELETE FROM tenants WHERE id = ANY($1)`, []string{idA, idB})
	})

	// Pool comum (não-superusuário) → RLS efetiva no Design Engine.
	cfg, err := pgxpool.ParseConfig(dsn())
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}
	cfg.AfterConnect = func(ctx context.Context, c *pgx.Conn) error {
		_, err := c.Exec(ctx, "SET ROLE "+sisAppRole)
		return err
	}
	appPool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("app pool: %v", err)
	}
	t.Cleanup(appPool.Close)

	// Pool admin (superusuário) para o IAM validar o token (não toca RLS).
	adminPool, err := pgxpool.New(ctx, dsn())
	if err != nil {
		t.Fatalf("admin pool: %v", err)
	}
	t.Cleanup(adminPool.Close)

	priv, err := auth.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	// IAM + Design sobre bufconn, com o interceptor de TenantContext.
	lis := bufconn.Listen(1 << 20)
	grpcSrv := grpc.NewServer(grpc.ChainUnaryInterceptor(tenantctx.UnaryServerInterceptor()))
	iamv1.RegisterIAMServiceServer(grpcSrv, iamapp.NewServer(priv, time.Hour, adminPool))
	designv1.RegisterDesignEngineServiceServer(grpcSrv, designapp.NewServer(appPool))
	go func() { _ = grpcSrv.Serve(lis) }()
	t.Cleanup(grpcSrv.Stop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(tenantctx.UnaryClientInterceptor()),
	)
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	handler := gatewayapp.NewRouter(
		iamv1.NewIAMServiceClient(conn),
		designv1.NewDesignEngineServiceClient(conn),
		nil, nil, nil,
		middleware.NewRateLimiter(1000, 1000), nil,
	)
	return handler, auth.NewIssuer(priv, time.Hour), idA, idB
}

func tokenDe(t *testing.T, iss *auth.Issuer, tenantID, tipo string) string {
	t.Helper()
	tok, err := iss.Issue("user-x", tenantID, tipo)
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func TestSistemas_DonoCriaEListaNaCadeiaReal(t *testing.T) {
	handler, iss, tenantA, _ := sistemasHarness(t)
	tok := tokenDe(t, iss, tenantA, "dono")

	// POST cria o sistema.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sistemas", strings.NewReader(`{"nome":"E2E Demo"}`))
	req.Header.Set("Authorization", "Bearer "+tok)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST esperava 201; got=%d body=%s", rec.Code, rec.Body.String())
	}
	var criado struct{ ID, Nome string }
	if err := json.Unmarshal(rec.Body.Bytes(), &criado); err != nil {
		t.Fatalf("json criado: %v", err)
	}
	if criado.ID == "" || criado.Nome != "E2E Demo" {
		t.Fatalf("resposta de criação inesperada: %+v", criado)
	}

	// GET lista e o sistema criado aparece.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/sistemas", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET esperava 200; got=%d", rec.Code)
	}
	var lista struct {
		Sistemas []struct{ ID, Nome string } `json:"sistemas"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &lista); err != nil {
		t.Fatalf("json lista: %v", err)
	}
	var achou bool
	for _, s := range lista.Sistemas {
		if s.ID == criado.ID {
			achou = true
		}
	}
	if !achou {
		t.Fatalf("sistema criado não apareceu na listagem: %+v", lista.Sistemas)
	}
}

func TestSistemas_ClienteNaoCria_403(t *testing.T) {
	handler, iss, tenantA, _ := sistemasHarness(t)
	tok := tokenDe(t, iss, tenantA, "cliente")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sistemas", strings.NewReader(`{"nome":"Proibido"}`))
	req.Header.Set("Authorization", "Bearer "+tok)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("cliente final deveria receber 403; got=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestSistemas_IsolamentoEntreTenants(t *testing.T) {
	handler, iss, tenantA, tenantB := sistemasHarness(t)

	// A cria um sistema.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sistemas", strings.NewReader(`{"nome":"So de A"}`))
	req.Header.Set("Authorization", "Bearer "+tokenDe(t, iss, tenantA, "dono"))
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST A esperava 201; got=%d body=%s", rec.Code, rec.Body.String())
	}
	var criado struct{ ID string }
	_ = json.Unmarshal(rec.Body.Bytes(), &criado)

	// B lista e nunca enxerga o sistema de A (RLS por tenant).
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/sistemas", nil)
	req.Header.Set("Authorization", "Bearer "+tokenDe(t, iss, tenantB, "dono"))
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET B esperava 200; got=%d", rec.Code)
	}
	var lista struct {
		Sistemas []struct{ ID string } `json:"sistemas"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &lista)
	for _, s := range lista.Sistemas {
		if s.ID == criado.ID {
			t.Fatalf("B não deveria ver o sistema de A (%s)", criado.ID)
		}
	}
}
