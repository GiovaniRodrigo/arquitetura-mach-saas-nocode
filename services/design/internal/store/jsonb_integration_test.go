//go:build integration

// Requer o Postgres do docker-compose com as migrações aplicadas (incl. 0011).
// Executar:
//
//	DATABASE_URL=postgres://mach:mach@localhost:5432/machv4?sslmode=disable \
//	  go test -tags integration ./services/design/internal/store/...
package store

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	commonv1 "github.com/machv4/platform/gen/go/construtor/common/v1"
	designv1 "github.com/machv4/platform/gen/go/construtor/design/v1"
	"github.com/machv4/platform/pkg/database"
	"github.com/machv4/platform/pkg/tenantctx"
)

// appRole é uma role não-superusuário: só sob ela a RLS (migração 0010/0011) é
// efetiva, já que superusuários a ignoram de propósito.
const appRole = "machapp_design_rls_test"

func dsn() string {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v
	}
	return "postgres://mach:mach@localhost:5432/machv4?sslmode=disable"
}

func ctxFor(tenantID string) context.Context {
	return tenantctx.NewContext(context.Background(), &commonv1.TenantContext{TenantId: tenantID, Tipo: "dono"})
}

// setup semeia (como superusuário) a role comum e dois pares tenant+sistema, e
// devolve um pool que assume a role comum em cada conexão — tornando a RLS efetiva.
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

	seed := func(nome string) (tid, sid string) {
		if err := admin.QueryRow(ctx,
			`INSERT INTO tenants (nome, tipo, chave_blind_index) VALUES ($1,'dono','\x6b') RETURNING id`, nome,
		).Scan(&tid); err != nil {
			t.Fatalf("tenant %s: %v", nome, err)
		}
		if err := admin.QueryRow(ctx,
			`INSERT INTO sistemas (tenant_id, nome) VALUES ($1,$2) RETURNING id`, tid, nome,
		).Scan(&sid); err != nil {
			t.Fatalf("sistema %s: %v", nome, err)
		}
		return tid, sid
	}
	tenantA, sisA = seed("itg-design-A")
	tenantB, sisB = seed("itg-design-B")

	t.Cleanup(func() {
		c, err := pgx.Connect(context.Background(), dsn())
		if err != nil {
			return
		}
		defer c.Close(context.Background())
		_, _ = c.Exec(context.Background(), `DELETE FROM tenants WHERE id = ANY($1)`, []string{tenantA, tenantB})
	})

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

func arvore() *designv1.Componente {
	c := &designv1.Componente{BlindIndex: "bi-root", Tipo: "container",
		ComponenteFilhos: []*designv1.Componente{{BlindIndex: "bi-a", Tipo: "input_texto"}}}
	c.ComponenteFilhos[0].Propriedades = []byte(`{"label":"Nome"}`)
	return c
}

func TestCriarObterRoundTrip(t *testing.T) {
	pool, tenantA, _, sisA, _ := setup(t)
	s := New(pool)
	ctx := ctxFor(tenantA)

	id, err := s.Criar(ctx, &designv1.Design{SistemaId: sisA, Nome: "Tela Login", Arvore: arvore()})
	if err != nil {
		t.Fatalf("criar: %v", err)
	}

	got, err := s.Obter(ctx, id)
	if err != nil {
		t.Fatalf("obter: %v", err)
	}
	if got.GetNome() != "Tela Login" || got.GetArvore().GetBlindIndex() != "bi-root" {
		t.Fatalf("design não preservado: %+v", got)
	}
	if len(got.GetArvore().GetComponenteFilhos()) != 1 {
		t.Fatalf("filhos não preservados: %+v", got.GetArvore())
	}
	// O JSONB reformata o texto (espaços), então comparamos semanticamente.
	var props map[string]string
	if err := json.Unmarshal(got.GetArvore().GetComponenteFilhos()[0].GetPropriedades(), &props); err != nil {
		t.Fatalf("propriedades ilegíveis: %v", err)
	}
	if props["label"] != "Nome" {
		t.Fatalf("propriedades não preservadas: %v", props)
	}
}

func TestSemTenant_Rejeitado(t *testing.T) {
	pool, _, _, sisA, _ := setup(t)
	s := New(pool)
	// Contexto sem TenantContext → ScopedDB recusa antes de tocar o banco (RN01).
	_, err := s.Criar(context.Background(), &designv1.Design{SistemaId: sisA, Arvore: arvore()})
	if !errors.Is(err, database.ErrNoTenant) {
		t.Fatalf("esperava ErrNoTenant; got %v", err)
	}
}

func TestIsolamentoEntreTenants(t *testing.T) {
	pool, tenantA, tenantB, sisA, _ := setup(t)
	s := New(pool)

	id, err := s.Criar(ctxFor(tenantA), &designv1.Design{SistemaId: sisA, Nome: "Privado de A", Arvore: arvore()})
	if err != nil {
		t.Fatalf("criar A: %v", err)
	}

	// O tenant B nunca enxerga o design de A (RLS): ErrNaoEncontrado.
	if _, err := s.Obter(ctxFor(tenantB), id); !errors.Is(err, ErrNaoEncontrado) {
		t.Fatalf("B não deveria ver design de A; got %v", err)
	}
}

func TestSalvarUpsert(t *testing.T) {
	pool, tenantA, _, sisA, _ := setup(t)
	s := New(pool)
	ctx := ctxFor(tenantA)

	id, err := s.Criar(ctx, &designv1.Design{SistemaId: sisA, Nome: "v1", Arvore: arvore()})
	if err != nil {
		t.Fatalf("criar: %v", err)
	}

	// Write-behind da colaboração: mesmo id, novo estado consolidado.
	nova := &designv1.Componente{BlindIndex: "bi-root", Tipo: "container"}
	if err := s.Salvar(ctx, &designv1.Design{Id: id, SistemaId: sisA, Nome: "v2", Arvore: nova}); err != nil {
		t.Fatalf("salvar: %v", err)
	}

	got, err := s.Obter(ctx, id)
	if err != nil {
		t.Fatalf("obter: %v", err)
	}
	if got.GetNome() != "v2" || len(got.GetArvore().GetComponenteFilhos()) != 0 {
		t.Fatalf("upsert não aplicado: %+v", got)
	}
}
