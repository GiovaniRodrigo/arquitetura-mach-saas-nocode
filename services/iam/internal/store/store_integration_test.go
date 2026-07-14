//go:build integration

// Requer o Postgres do docker-compose com as migrações aplicadas. Executar:
//
//	DATABASE_URL=postgres://mach:mach@localhost:5432/machv4?sslmode=disable \
//	  go test -tags integration ./services/iam/internal/store/...
package store

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	commonv1 "github.com/machv4/platform/gen/go/construtor/common/v1"
	"github.com/machv4/platform/pkg/tenantctx"
)

func dsn() string {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v
	}
	return "postgres://mach:mach@localhost:5432/machv4?sslmode=disable"
}

func pool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	p, err := pgxpool.New(context.Background(), dsn())
	if err != nil {
		t.Skipf("Postgres indisponível (%v)", err)
	}
	t.Cleanup(p.Close)
	return p
}

func TestHierarquiaTenants(t *testing.T) {
	p := pool(t)
	s := New(p)
	ctx := context.Background()

	dono, err := s.CriarTenant(ctx, "itg-dono", "dono", nil, []byte("k"))
	if err != nil {
		t.Fatalf("criar dono: %v", err)
	}
	t.Cleanup(func() { _, _ = p.Exec(context.Background(), `DELETE FROM tenants WHERE id=$1`, dono.ID) })

	parceiro, err := s.CriarTenant(ctx, "itg-parceiro", "parceiro", &dono.ID, []byte("k"))
	if err != nil {
		t.Fatalf("criar parceiro: %v", err)
	}

	got, err := s.ObterTenant(ctx, parceiro.ID)
	if err != nil {
		t.Fatalf("obter: %v", err)
	}
	if got.ParentID == nil || *got.ParentID != dono.ID {
		t.Fatalf("parent_id do parceiro deveria ser o dono; got=%v", got.ParentID)
	}

	filhos, err := s.ListarFilhos(ctx, dono.ID)
	if err != nil {
		t.Fatalf("listar filhos: %v", err)
	}
	if len(filhos) != 1 || filhos[0].ID != parceiro.ID {
		t.Fatalf("dono deveria ter 1 filho (parceiro); got=%+v", filhos)
	}
}

func TestPermissoesDe_FiltraPorTenant(t *testing.T) {
	p := pool(t)
	s := New(p)
	ctx := context.Background()

	ta, _ := s.CriarTenant(ctx, "itg-permA", "dono", nil, []byte("k"))
	tb, _ := s.CriarTenant(ctx, "itg-permB", "dono", nil, []byte("k"))
	t.Cleanup(func() {
		_, _ = p.Exec(context.Background(), `DELETE FROM tenants WHERE id = ANY($1)`, []string{ta.ID, tb.ID})
	})

	_, _ = p.Exec(ctx, `INSERT INTO permissoes (tenant_id, blind_index, papel, view, click) VALUES ($1,'bi-1','editor',true,true)`, ta.ID)
	_, _ = p.Exec(ctx, `INSERT INTO permissoes (tenant_id, blind_index, papel, view, click) VALUES ($1,'bi-1','editor',true,false)`, tb.ID)

	ctxA := tenantctx.NewContext(ctx, &commonv1.TenantContext{TenantId: ta.ID})
	perms, err := s.PermissoesDe(ctxA, []string{"bi-1"})
	if err != nil {
		t.Fatalf("permissões: %v", err)
	}
	if len(perms) != 1 {
		t.Fatalf("tenant A deveria ver apenas a própria permissão; got=%d", len(perms))
	}
	if !perms[0].Click {
		t.Fatal("deveria ter carregado a permissão do tenant A (click=true), não a de B")
	}
}

func TestPermissoesDe_SemTenant(t *testing.T) {
	if _, err := New(pool(t)).PermissoesDe(context.Background(), []string{"bi-1"}); err != ErrSemTenant {
		t.Fatalf("esperava ErrSemTenant; got=%v", err)
	}
}
