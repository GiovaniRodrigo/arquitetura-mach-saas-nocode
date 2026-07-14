//go:build integration

// Testes de integração da RLS multi-tenant. Requerem o Postgres do
// docker-compose de pé com as migrações aplicadas. Executar com:
//
//	DATABASE_URL=postgres://mach:mach@localhost:5432/machv4?sslmode=disable \
//	  go test -tags integration ./pkg/database/...
package database

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	commonv1 "github.com/machv4/platform/gen/go/construtor/common/v1"
	"github.com/machv4/platform/pkg/tenantctx"
)

const appRole = "machapp_rls_test"

func dsn() string {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v
	}
	return "postgres://mach:mach@localhost:5432/machv4?sslmode=disable"
}

func ctxFor(tenantID string) context.Context {
	return tenantctx.NewContext(context.Background(), &commonv1.TenantContext{TenantId: tenantID})
}

// setup cria a role não-superusuário (para a RLS valer), dois tenants e um sistema
// cada. As inserções ocorrem como superusuário, que ignora a RLS de propósito.
func setup(t *testing.T) (pool *pgxpool.Pool, tenantA, tenantB, sisA, sisB string) {
	t.Helper()
	ctx := context.Background()

	admin, err := pgx.Connect(ctx, dsn())
	if err != nil {
		t.Skipf("Postgres indisponível (%v) — teste de integração ignorado", err)
	}
	defer admin.Close(ctx)

	for _, stmt := range []string{
		`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname='` + appRole + `') THEN CREATE ROLE ` + appRole + ` NOLOGIN NOSUPERUSER; END IF; END $$;`,
		`GRANT USAGE ON SCHEMA public TO ` + appRole + `;`,
		`GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO ` + appRole + `;`,
	} {
		if _, err := admin.Exec(ctx, stmt); err != nil {
			t.Fatalf("setup role: %v (%s)", err, stmt)
		}
	}

	err = admin.QueryRow(ctx,
		`INSERT INTO tenants (nome, tipo, chave_blind_index) VALUES ('itg-A','dono','\x00') RETURNING id`).Scan(&tenantA)
	if err != nil {
		t.Fatalf("insert tenant A: %v", err)
	}
	err = admin.QueryRow(ctx,
		`INSERT INTO tenants (nome, tipo, chave_blind_index) VALUES ('itg-B','dono','\x00') RETURNING id`).Scan(&tenantB)
	if err != nil {
		t.Fatalf("insert tenant B: %v", err)
	}
	if err := admin.QueryRow(ctx, `INSERT INTO sistemas (tenant_id, nome) VALUES ($1,'sA') RETURNING id`, tenantA).Scan(&sisA); err != nil {
		t.Fatalf("insert sistema A: %v", err)
	}
	if err := admin.QueryRow(ctx, `INSERT INTO sistemas (tenant_id, nome) VALUES ($1,'sB') RETURNING id`, tenantB).Scan(&sisB); err != nil {
		t.Fatalf("insert sistema B: %v", err)
	}

	t.Cleanup(func() {
		c, err := pgx.Connect(context.Background(), dsn())
		if err != nil {
			return
		}
		defer c.Close(context.Background())
		_, _ = c.Exec(context.Background(), `DELETE FROM tenants WHERE id = ANY($1)`, []string{tenantA, tenantB})
	})

	// Pool que assume a role comum em cada conexão — assim a RLS é efetiva.
	cfg, err := pgxpool.ParseConfig(dsn())
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}
	cfg.AfterConnect = func(ctx context.Context, c *pgx.Conn) error {
		_, err := c.Exec(ctx, "SET ROLE "+appRole)
		return err
	}
	pool, err = pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool, tenantA, tenantB, sisA, sisB
}

// O tenant A só enxerga o próprio sistema; o de B fica invisível (RN01).
func TestRLS_LeituraIsoladaPorTenant(t *testing.T) {
	pool, tenantA, tenantB, sisA, sisB := setup(t)
	db := New(pool)

	var visto string
	err := db.WithTenant(ctxFor(tenantA), func(ctx context.Context, tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT string_agg(id::text, ',') FROM sistemas WHERE id = ANY($1)`,
			[]string{sisA, sisB}).Scan(&visto)
	})
	if err != nil {
		t.Fatalf("leitura tenant A: %v", err)
	}
	if visto != sisA {
		t.Fatalf("tenant A deveria ver apenas %s; viu %q", sisA, visto)
	}

	err = db.WithTenant(ctxFor(tenantB), func(ctx context.Context, tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT string_agg(id::text, ',') FROM sistemas WHERE id = ANY($1)`,
			[]string{sisA, sisB}).Scan(&visto)
	})
	if err != nil {
		t.Fatalf("leitura tenant B: %v", err)
	}
	if visto != sisB {
		t.Fatalf("tenant B deveria ver apenas %s; viu %q", sisB, visto)
	}
}

// A política WITH CHECK impede inserir dados carimbados com o tenant de outro.
func TestRLS_InsercaoCrossTenantBloqueada(t *testing.T) {
	pool, tenantA, tenantB, _, _ := setup(t)
	db := New(pool)

	err := db.WithTenant(ctxFor(tenantA), func(ctx context.Context, tx pgx.Tx) error {
		_, e := tx.Exec(ctx, `INSERT INTO sistemas (tenant_id, nome) VALUES ($1, 'intruso')`, tenantB)
		return e
	})
	if err == nil {
		t.Fatal("inserção com tenant_id de outro tenant deveria falhar (WITH CHECK)")
	}
}
