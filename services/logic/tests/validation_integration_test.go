//go:build integration

// Teste de integração do Logic Engine + Postgres (RN08, RNF08, critério 2).
// Requer o Postgres do docker-compose com as migrações aplicadas. Executar:
//
//	DATABASE_URL=postgres://mach:mach@localhost:5432/machv4?sslmode=disable \
//	  go test -tags integration ./services/logic/tests/...
package tests

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	commonv1 "github.com/machv4/platform/gen/go/construtor/common/v1"
	logicv1 "github.com/machv4/platform/gen/go/construtor/logic/v1"
	"github.com/machv4/platform/pkg/tenantctx"
	logicapp "github.com/machv4/platform/services/logic/app"
)

const appRole = "machapp_logic_rls_test"

func dsn() string {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v
	}
	return "postgres://mach:mach@localhost:5432/machv4?sslmode=disable"
}

// setup semeia (como superusuário) a role comum, um tenant, um sistema e o schema
// de campos, e devolve o servidor Logic sobre um pool que assume a role comum (RLS
// efetiva) mais o contexto de tenant e o sistema_id.
func setup(t *testing.T) (srv logicv1.LogicEngineServiceServer, ctx context.Context, pool *pgxpool.Pool, sistemaID string) {
	t.Helper()
	bg := context.Background()

	admin, err := pgx.Connect(bg, dsn())
	if err != nil {
		t.Skipf("Postgres indisponível (%v)", err)
	}
	defer admin.Close(bg)

	for _, stmt := range []string{
		`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname='` + appRole + `') THEN CREATE ROLE ` + appRole + ` NOLOGIN NOSUPERUSER; END IF; END $$;`,
		`GRANT USAGE ON SCHEMA public TO ` + appRole + `;`,
		`GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO ` + appRole + `;`,
	} {
		if _, err := admin.Exec(bg, stmt); err != nil {
			t.Fatalf("setup role: %v (%s)", err, stmt)
		}
	}

	var tenantID string
	if err := admin.QueryRow(bg,
		`INSERT INTO tenants (nome, tipo, chave_blind_index) VALUES ('itg-logic','dono','\x6b') RETURNING id`).Scan(&tenantID); err != nil {
		t.Fatalf("tenant: %v", err)
	}
	t.Cleanup(func() {
		c, err := pgx.Connect(context.Background(), dsn())
		if err != nil {
			return
		}
		defer c.Close(context.Background())
		_, _ = c.Exec(context.Background(), `DELETE FROM tenants WHERE id=$1`, tenantID)
	})

	if err := admin.QueryRow(bg,
		`INSERT INTO sistemas (tenant_id, nome) VALUES ($1,'sistema-itg') RETURNING id`, tenantID).Scan(&sistemaID); err != nil {
		t.Fatalf("sistema: %v", err)
	}

	// Schema: bi-nome (string, obrigatório) e bi-idade (number, 0..120).
	for _, ins := range []struct {
		bi, tipo string
		obrig    bool
		limites  string
	}{
		{"bi-nome", "string", true, `{"max_length":20}`},
		{"bi-idade", "number", true, `{"min":0,"max":120}`},
	} {
		if _, err := admin.Exec(bg,
			`INSERT INTO campos_definicao (blind_index, sistema_id, tenant_id, tipo, obrigatorio, limites)
			 VALUES ($1,$2,$3,$4,$5,$6)`,
			ins.bi, sistemaID, tenantID, ins.tipo, ins.obrig, ins.limites); err != nil {
			t.Fatalf("campo %s: %v", ins.bi, err)
		}
	}

	cfg, err := pgxpool.ParseConfig(dsn())
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}
	cfg.AfterConnect = func(ctx context.Context, c *pgx.Conn) error {
		_, err := c.Exec(ctx, "SET ROLE "+appRole)
		return err
	}
	pool, err = pgxpool.NewWithConfig(bg, cfg)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)

	ctx = tenantctx.NewContext(bg, &commonv1.TenantContext{TenantId: tenantID, Tipo: "dono"})
	return logicapp.NewServer(pool), ctx, pool, sistemaID
}

// contagemDados conta via conexão superusuário (bypassa a RLS) — o pool da
// aplicação exigiria app.tenant_id fixado, que só existe dentro de WithTenant.
func contagemDados(t *testing.T, sistemaID string) int {
	t.Helper()
	c, err := pgx.Connect(context.Background(), dsn())
	if err != nil {
		t.Fatalf("conexão admin: %v", err)
	}
	defer c.Close(context.Background())
	var n int
	if err := c.QueryRow(context.Background(),
		`SELECT count(*) FROM dados_operacionais WHERE sistema_id=$1`, sistemaID).Scan(&n); err != nil {
		t.Fatalf("contagem: %v", err)
	}
	return n
}

// Critério 2: submissão que contorna o cliente (campo injetado + valor fora do
// limite) é rejeitada, com erros só por blind_index e sem nomes reais, e nada é
// persistido.
func TestSubmissaoMaliciosa_Rejeitada(t *testing.T) {
	srv, ctx, _, sistemaID := setup(t)

	resp, err := srv.SalvarFormulario(ctx, &logicv1.SalvarFormularioRequest{
		SistemaId: sistemaID,
		DadosFormulario: map[string]string{
			"bi-nome":     "Ana",
			"bi-idade":    "999",              // acima do máximo
			"bi-injetado": "'; DROP TABLE x;", // chave fora do schema (contorna o cliente)
		},
	})
	if err != nil {
		t.Fatalf("erro gRPC inesperado: %v", err)
	}
	if resp.GetSucesso() {
		t.Fatal("submissão maliciosa não deveria ter sucesso")
	}

	erros := resp.GetErrosValidacao()
	if _, ok := erros["bi-idade"]; !ok {
		t.Fatalf("esperava erro em bi-idade; got %v", erros)
	}
	if _, ok := erros["bi-injetado"]; !ok {
		t.Fatalf("esperava rejeição da chave injetada; got %v", erros)
	}
	// As mensagens jamais podem conter nome real de coluna/tabela — só termos
	// genéricos (RNF08). As chaves são blind_indexes (do cliente), não colunas.
	for k, v := range erros {
		for _, proibido := range []string{"sistema_id", "dados_operacionais", "campos_definicao", "regras_negocio", "valores"} {
			if strings.Contains(strings.ToLower(v), proibido) {
				t.Fatalf("mensagem vazou nome real %q em (%q:%q)", proibido, k, v)
			}
		}
	}

	if n := contagemDados(t, sistemaID); n != 0 {
		t.Fatalf("nada deveria ser persistido; encontrei %d linhas", n)
	}
}

// Contraprova: submissão válida persiste exatamente uma linha.
func TestSubmissaoValida_Persiste(t *testing.T) {
	srv, ctx, _, sistemaID := setup(t)

	resp, err := srv.SalvarFormulario(ctx, &logicv1.SalvarFormularioRequest{
		SistemaId:       sistemaID,
		DadosFormulario: map[string]string{"bi-nome": "Ana", "bi-idade": "30"},
	})
	if err != nil {
		t.Fatalf("erro gRPC inesperado: %v", err)
	}
	if !resp.GetSucesso() {
		t.Fatalf("submissão válida deveria ter sucesso; erros=%v", resp.GetErrosValidacao())
	}
	if n := contagemDados(t, sistemaID); n != 1 {
		t.Fatalf("esperava 1 linha persistida; encontrei %d", n)
	}
}
