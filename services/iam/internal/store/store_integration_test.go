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

func TestAtualizarEExcluirTenant(t *testing.T) {
	p := pool(t)
	s := New(p)
	ctx := context.Background()

	dono, err := s.CriarTenant(ctx, "itg-dono-upd", "dono", nil, []byte("k"))
	if err != nil {
		t.Fatalf("criar dono: %v", err)
	}
	t.Cleanup(func() { _, _ = p.Exec(context.Background(), `DELETE FROM tenants WHERE id=$1`, dono.ID) })

	cliente, err := s.CriarTenant(ctx, "itg-cliente-upd", "cliente", &dono.ID, []byte("k"))
	if err != nil {
		t.Fatalf("criar cliente: %v", err)
	}

	atualizado, err := s.AtualizarTenant(ctx, cliente.ID, "itg-cliente-renomeado")
	if err != nil {
		t.Fatalf("atualizar: %v", err)
	}
	if atualizado.Nome != "itg-cliente-renomeado" {
		t.Fatalf("nome não atualizado: %+v", atualizado)
	}

	if _, err := s.AtualizarTenant(ctx, "00000000-0000-0000-0000-000000000099", "x"); err != ErrNaoEncontrado {
		t.Fatalf("esperava ErrNaoEncontrado; got=%v", err)
	}

	if err := s.ExcluirTenant(ctx, cliente.ID); err != nil {
		t.Fatalf("excluir: %v", err)
	}
	if _, err := s.ObterTenant(ctx, cliente.ID); err != ErrNaoEncontrado {
		t.Fatalf("esperava ErrNaoEncontrado após excluir; got=%v", err)
	}
	if err := s.ExcluirTenant(ctx, cliente.ID); err != ErrNaoEncontrado {
		t.Fatalf("excluir de novo deveria devolver ErrNaoEncontrado; got=%v", err)
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

func TestCriarTenantEUsuarioComSenha(t *testing.T) {
	p := pool(t)
	s := New(p)
	ctx := context.Background()

	email := "itg-cadastro@example.com"
	t.Cleanup(func() {
		_, _ = p.Exec(context.Background(), `DELETE FROM users WHERE email=$1`, email)
	})

	userID, tenantID, err := s.CriarTenantEUsuarioComSenha(ctx, "Ana", email, "hash-bcrypt-fake", "Ana LTDA")
	if err != nil {
		t.Fatalf("criar tenant e usuário: %v", err)
	}
	t.Cleanup(func() { _, _ = p.Exec(context.Background(), `DELETE FROM tenants WHERE id=$1`, tenantID) })

	if userID == "" || tenantID == "" {
		t.Fatalf("esperava ids não vazios; got user=%q tenant=%q", userID, tenantID)
	}

	tenant, err := s.ObterTenant(ctx, tenantID)
	if err != nil {
		t.Fatalf("obter tenant criado: %v", err)
	}
	if tenant.Tipo != "dono" {
		t.Fatalf("tenant do cadastro deveria ser 'dono'; got=%q", tenant.Tipo)
	}
	if tenant.ParentID != nil {
		t.Fatalf("tenant do cadastro deveria ser raiz (parent_id nil); got=%v", tenant.ParentID)
	}

	uid, tid, tipo, hash, err := s.ObterUsuarioPorEmailSenha(ctx, email)
	if err != nil {
		t.Fatalf("obter usuário por e-mail/senha: %v", err)
	}
	if uid != userID || tid != tenantID {
		t.Fatalf("ids inconsistentes: cadastro(user=%s,tenant=%s) obtido(user=%s,tenant=%s)", userID, tenantID, uid, tid)
	}
	if tipo != "dono" {
		t.Fatalf("tipo do usuário do cadastro deveria ser 'dono'; got=%q", tipo)
	}
	if hash != "hash-bcrypt-fake" {
		t.Fatalf("senha_hash não persistido corretamente; got=%q", hash)
	}
}

func TestCriarTenantEUsuarioComSenha_EmailDuplicadoNaoDeixaTenantOrfao(t *testing.T) {
	p := pool(t)
	s := New(p)
	ctx := context.Background()

	email := "itg-duplicado@example.com"
	t.Cleanup(func() {
		_, _ = p.Exec(context.Background(), `DELETE FROM users WHERE email=$1`, email)
	})

	_, tenantID1, err := s.CriarTenantEUsuarioComSenha(ctx, "Ana", email, "hash1", "Ana LTDA")
	if err != nil {
		t.Fatalf("primeiro cadastro: %v", err)
	}
	t.Cleanup(func() { _, _ = p.Exec(context.Background(), `DELETE FROM tenants WHERE id=$1`, tenantID1) })

	var antes int
	if err := p.QueryRow(ctx, `SELECT count(*) FROM tenants`).Scan(&antes); err != nil {
		t.Fatalf("contar tenants antes: %v", err)
	}

	_, _, err = s.CriarTenantEUsuarioComSenha(ctx, "Ana Duplicada", email, "hash2", "Outra Empresa")
	if err != ErrEmailJaCadastrado {
		t.Fatalf("esperava ErrEmailJaCadastrado; got=%v", err)
	}

	var depois int
	if err := p.QueryRow(ctx, `SELECT count(*) FROM tenants`).Scan(&depois); err != nil {
		t.Fatalf("contar tenants depois: %v", err)
	}
	if depois != antes {
		t.Fatalf("cadastro com e-mail duplicado deixou tenant órfão: antes=%d depois=%d", antes, depois)
	}
}

func TestObterUsuarioPorEmailSenha_NaoEncontrado(t *testing.T) {
	if _, _, _, _, err := New(pool(t)).ObterUsuarioPorEmailSenha(context.Background(), "inexistente@example.com"); err != ErrNaoEncontrado {
		t.Fatalf("esperava ErrNaoEncontrado; got=%v", err)
	}
}
