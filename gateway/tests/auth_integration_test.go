//go:build integration

// Teste ponta-a-ponta do Gateway + IAM (RN01, critério de aceite 1). Requer o
// Postgres do docker-compose com as migrações aplicadas. Executar:
//
//	DATABASE_URL=postgres://mach:mach@localhost:5432/machv4?sslmode=disable \
//	  go test -tags integration ./gateway/tests/...
package tests

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	iamv1 "github.com/machv4/platform/gen/go/construtor/iam/v1"
	gatewayapp "github.com/machv4/platform/gateway/internal/app"
	"github.com/machv4/platform/gateway/internal/middleware"
	"github.com/machv4/platform/pkg/tenantctx"
	iamapp "github.com/machv4/platform/services/iam/app"
	"github.com/machv4/platform/services/iam/auth"
)

func dsn() string {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v
	}
	return "postgres://mach:mach@localhost:5432/machv4?sslmode=disable"
}

// harness monta Postgres + IAM (bufconn) + Gateway e devolve o handler HTTP, um
// emissor de JWT e os ids dos dois tenants semeados.
func harness(t *testing.T) (handler http.Handler, iss *auth.Issuer, tenantA, tenantB string) {
	t.Helper()
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn())
	if err != nil {
		t.Skipf("Postgres indisponível (%v)", err)
	}
	t.Cleanup(pool.Close)

	var idA, idB string
	if err := pool.QueryRow(ctx, `INSERT INTO tenants (nome, tipo, chave_blind_index) VALUES ('itg-gw-A','dono','\x6b41') RETURNING id`).Scan(&idA); err != nil {
		t.Fatalf("tenant A: %v", err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO tenants (nome, tipo, chave_blind_index) VALUES ('itg-gw-B','dono','\x6b42') RETURNING id`).Scan(&idB); err != nil {
		t.Fatalf("tenant B: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM tenants WHERE id = ANY($1)`, []string{idA, idB})
	})

	// A tem view+click em bi-1; B tem apenas view — assim a decisão distingue os tenants.
	if _, err := pool.Exec(ctx, `INSERT INTO permissoes (tenant_id, blind_index, papel, view, click) VALUES ($1,'bi-1','dono',true,true)`, idA); err != nil {
		t.Fatalf("perm A: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO permissoes (tenant_id, blind_index, papel, view, click) VALUES ($1,'bi-1','dono',true,false)`, idB); err != nil {
		t.Fatalf("perm B: %v", err)
	}

	// Par de chaves partilhado entre emissor (teste) e validador (IAM).
	priv, err := auth.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	iamSrv := iamapp.NewServer(&priv.PublicKey, pool)

	// IAM sobre bufconn, com o interceptor que extrai o TenantContext do Metadata.
	lis := bufconn.Listen(1 << 20)
	grpcSrv := grpc.NewServer(grpc.ChainUnaryInterceptor(tenantctx.UnaryServerInterceptor()))
	iamv1.RegisterIAMServiceServer(grpcSrv, iamSrv)
	go func() { _ = grpcSrv.Serve(lis) }()
	t.Cleanup(grpcSrv.Stop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(tenantctx.UnaryClientInterceptor()),
	)
	if err != nil {
		t.Fatalf("iam client: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	// Este teste exercita apenas /permissoes; os demais serviços não são dialados aqui.
	handler = gatewayapp.NewRouter(iamv1.NewIAMServiceClient(conn), nil, nil, nil, nil, middleware.NewRateLimiter(1000, 1000))
	return handler, auth.NewIssuer(priv, time.Hour), idA, idB
}

func TestGateway_SemJWT_401(t *testing.T) {
	handler, _, _, _ := harness(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/permissoes?sistema_id=s1&bi=bi-1", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("sem JWT deveria dar 401; got=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGateway_TokenInvalido_401(t *testing.T) {
	handler, _, _, _ := harness(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/permissoes?sistema_id=s1&bi=bi-1", nil)
	req.Header.Set("Authorization", "Bearer nao-e-um-jwt")
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("token inválido deveria dar 401; got=%d", rec.Code)
	}
}

// O núcleo do critério 1: o token do tenant A jamais recebe a decisão do tenant B.
func TestGateway_IsolamentoEntreTenants(t *testing.T) {
	handler, iss, tenantA, tenantB := harness(t)

	decisao := func(tenantID string) (view, click bool) {
		token, err := iss.Issue("user-x", tenantID, "dono")
		if err != nil {
			t.Fatal(err)
		}
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/permissoes?sistema_id=s1&bi=bi-1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("esperava 200 para tenant %s; got=%d body=%s", tenantID, rec.Code, rec.Body.String())
		}
		var body struct {
			Permissions map[string]struct {
				View  bool `json:"view"`
				Click bool `json:"click"`
			} `json:"permissions"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("json: %v", err)
		}
		p := body.Permissions["bi-1"]
		return p.View, p.Click
	}

	aView, aClick := decisao(tenantA)
	if !aView || !aClick {
		t.Fatalf("tenant A deveria ter view+click; got view=%v click=%v", aView, aClick)
	}

	bView, bClick := decisao(tenantB)
	if !bView || bClick {
		t.Fatalf("tenant B deveria ter só view (dado próprio, não o de A); got view=%v click=%v", bView, bClick)
	}
}
