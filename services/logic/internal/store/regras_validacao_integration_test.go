//go:build integration

// Teste de integração de RegraValidacao contra Postgres real (RF10/RF11, RN01,
// RN06). Requer o Postgres do docker-compose com as migrações aplicadas.
// Executar:
//
//	DATABASE_URL=postgres://mach:mach@localhost:5432/machv4?sslmode=disable \
//	  go test -tags integration ./services/logic/internal/store/...
package store

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	commonv1 "github.com/machv4/platform/gen/go/construtor/common/v1"
	"github.com/machv4/platform/pkg/tenantctx"
)

const appRoleRegrasValidacao = "machapp_logic_regras_validacao_rls_test"

func dsnRegrasValidacao() string {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v
	}
	return "postgres://mach:mach@localhost:5432/machv4?sslmode=disable"
}

// setupRegrasValidacao cria (como superusuário) dois tenants com um sistema
// cada (sem regras de validação pré-existentes) e devolve o Store sobre um
// pool com a role comum (RLS efetiva).
func setupRegrasValidacao(t *testing.T) (s *Store, ctxA context.Context, sistemaA string, ctxB context.Context, sistemaB string) {
	t.Helper()
	bg := context.Background()

	admin, err := pgx.Connect(bg, dsnRegrasValidacao())
	if err != nil {
		t.Skipf("Postgres indisponível (%v)", err)
	}
	defer admin.Close(bg)

	if _, err := admin.Exec(bg,
		`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname='`+appRoleRegrasValidacao+`') THEN CREATE ROLE `+appRoleRegrasValidacao+` NOLOGIN NOSUPERUSER; END IF; END $$;`); err != nil {
		t.Fatalf("setup role: %v", err)
	}
	if _, err := admin.Exec(bg, `GRANT USAGE ON SCHEMA public TO `+appRoleRegrasValidacao); err != nil {
		t.Fatalf("grant usage: %v", err)
	}
	if _, err := admin.Exec(bg, `GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO `+appRoleRegrasValidacao); err != nil {
		t.Fatalf("grant tables: %v", err)
	}

	criar := func(nomeTenant string) (tenantID, sistemaID string) {
		if err := admin.QueryRow(bg,
			`INSERT INTO tenants (nome, tipo, chave_blind_index) VALUES ($1,'dono','\x6b') RETURNING id`, nomeTenant,
		).Scan(&tenantID); err != nil {
			t.Fatalf("tenant %s: %v", nomeTenant, err)
		}
		if err := admin.QueryRow(bg,
			`INSERT INTO sistemas (tenant_id, nome) VALUES ($1,'sistema-itg-rv') RETURNING id`, tenantID,
		).Scan(&sistemaID); err != nil {
			t.Fatalf("sistema de %s: %v", nomeTenant, err)
		}
		t.Cleanup(func() {
			c, err := pgx.Connect(context.Background(), dsnRegrasValidacao())
			if err != nil {
				return
			}
			defer c.Close(context.Background())
			_, _ = c.Exec(context.Background(), `DELETE FROM tenants WHERE id=$1`, tenantID)
		})
		return tenantID, sistemaID
	}

	tenantA, sistemaA := criar("itg-regras-validacao-a")
	tenantB, sistemaB := criar("itg-regras-validacao-b")

	cfg, err := pgxpool.ParseConfig(dsnRegrasValidacao())
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}
	cfg.AfterConnect = func(ctx context.Context, c *pgx.Conn) error {
		_, err := c.Exec(ctx, "SET ROLE "+appRoleRegrasValidacao)
		return err
	}
	pool, err := pgxpool.NewWithConfig(bg, cfg)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)

	s = New(pool)
	ctxA = tenantctx.NewContext(bg, &commonv1.TenantContext{TenantId: tenantA, Tipo: "dono"})
	ctxB = tenantctx.NewContext(bg, &commonv1.TenantContext{TenantId: tenantB, Tipo: "dono"})
	return s, ctxA, sistemaA, ctxB, sistemaB
}

func TestCriarRegraValidacao_ListarRegrasValidacao_RoundTrip(t *testing.T) {
	s, ctxA, sistemaA, _, _ := setupRegrasValidacao(t)

	criada, err := s.CriarRegraValidacao(ctxA, RegraValidacao{
		SistemaID:    sistemaA,
		BlindIndexes: []string{"bi-1", "bi-2"},
		Tipo:         "regex",
		Parametros:   []byte(`{"padrao":"^[0-9]+$"}`),
	})
	if err != nil {
		t.Fatalf("erro inesperado ao criar: %v", err)
	}
	if criada.ID == "" {
		t.Fatal("esperava id preenchido após criar")
	}

	regras, err := s.ListarRegrasValidacao(ctxA, sistemaA)
	if err != nil {
		t.Fatalf("erro inesperado ao listar: %v", err)
	}
	if len(regras) != 1 {
		t.Fatalf("esperava 1 regra; got %d", len(regras))
	}
	if regras[0].ID != criada.ID || regras[0].Tipo != "regex" || len(regras[0].BlindIndexes) != 2 {
		t.Fatalf("regra inesperada: %+v", regras[0])
	}
}

func TestListarRegrasValidacao_IsolamentoEntreTenants(t *testing.T) {
	s, ctxA, _, ctxB, sistemaB := setupRegrasValidacao(t)

	if _, err := s.CriarRegraValidacao(ctxB, RegraValidacao{
		SistemaID:    sistemaB,
		BlindIndexes: []string{"bi-1"},
		Tipo:         "obrigatorio",
		Parametros:   []byte(`{}`),
	}); err != nil {
		t.Fatalf("erro inesperado ao criar regra do tenant B: %v", err)
	}

	// Tenant A não deve ver as regras do sistema do tenant B, mesmo sabendo o id.
	regras, err := s.ListarRegrasValidacao(ctxA, sistemaB)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(regras) != 0 {
		t.Fatalf("RLS deveria impedir acesso cross-tenant; got %d regras", len(regras))
	}

	// Contraprova: o próprio tenant B enxerga sua regra.
	regrasB, err := s.ListarRegrasValidacao(ctxB, sistemaB)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(regrasB) != 1 {
		t.Fatalf("esperava 1 regra para o tenant B; got %d", len(regrasB))
	}
}
