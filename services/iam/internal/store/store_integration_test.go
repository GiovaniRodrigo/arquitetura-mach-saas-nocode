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
	"time"

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

// ─────────────────────────────────────────────────────────────────────────
// Área "Conta/Configuração" (spec 004, RF14-RF18).
// ─────────────────────────────────────────────────────────────────────────

// criarUsuarioSenha cria (via CriarTenantEUsuarioComSenha) um tenant raiz +
// usuário de senha para os testes desta seção, e registra a limpeza de ambos.
func criarUsuarioSenha(t *testing.T, s *Store, email string) (userID, tenantID string) {
	t.Helper()
	userID, tenantID, err := s.CriarTenantEUsuarioComSenha(context.Background(), "Itg Conta", email, "hash-bcrypt-fake", "Itg Conta LTDA")
	if err != nil {
		t.Fatalf("criar usuário de senha: %v", err)
	}
	// users.tenant_id não tem ON DELETE CASCADE (diferente de tenants.parent_id,
	// que é SET NULL) — apagar o tenant antes do usuário violaria a FK. Um único
	// Cleanup, na ordem certa, evita o vazamento de dados de teste.
	t.Cleanup(func() {
		ctx := context.Background()
		_, _ = s.db.Exec(ctx, `DELETE FROM users WHERE id=$1`, userID)
		_, _ = s.db.Exec(ctx, `DELETE FROM tenants WHERE id=$1`, tenantID)
	})
	return userID, tenantID
}

func TestAtualizarPerfil_store(t *testing.T) {
	p := pool(t)
	s := New(p)
	userID, _ := criarUsuarioSenha(t, s, "itg-perfil@example.com")

	if err := s.AtualizarPerfil(context.Background(), userID, "Novo Nome", "https://x/foto.png"); err != nil {
		t.Fatalf("atualizar perfil: %v", err)
	}

	var nome, foto string
	if err := p.QueryRow(context.Background(), `SELECT nome, foto_url FROM users WHERE id=$1`, userID).Scan(&nome, &foto); err != nil {
		t.Fatalf("verificar perfil: %v", err)
	}
	if nome != "Novo Nome" || foto != "https://x/foto.png" {
		t.Fatalf("perfil não persistido corretamente: nome=%q foto=%q", nome, foto)
	}

	if err := s.AtualizarPerfil(context.Background(), "00000000-0000-0000-0000-000000000099", "X", ""); err != ErrNaoEncontrado {
		t.Fatalf("usuário inexistente deveria dar ErrNaoEncontrado; got=%v", err)
	}
}

func TestObterHashSenhaEAtualizarSenha(t *testing.T) {
	p := pool(t)
	s := New(p)
	userID, _ := criarUsuarioSenha(t, s, "itg-senha@example.com")

	hash, err := s.ObterHashSenha(context.Background(), userID)
	if err != nil {
		t.Fatalf("obter hash: %v", err)
	}
	if hash != "hash-bcrypt-fake" {
		t.Fatalf("hash inesperado: %q", hash)
	}

	if err := s.AtualizarSenha(context.Background(), userID, "novo-hash-bcrypt"); err != nil {
		t.Fatalf("atualizar senha: %v", err)
	}
	hash2, err := s.ObterHashSenha(context.Background(), userID)
	if err != nil {
		t.Fatalf("obter hash após atualizar: %v", err)
	}
	if hash2 != "novo-hash-bcrypt" {
		t.Fatalf("senha não persistida corretamente: %q", hash2)
	}
}

// TestFluxoMfa cobre ativar → confirmar → anti-replay → desativar (RF15): MFA
// não fica ativo antes de confirmar, e o mesmo código não confirma duas vezes.
func TestFluxoMfa(t *testing.T) {
	p := pool(t)
	s := New(p)
	userID, _ := criarUsuarioSenha(t, s, "itg-mfa@example.com")
	ctx := context.Background()

	if err := s.IniciarMfa(ctx, userID, []byte("segredo-cifrado-1")); err != nil {
		t.Fatalf("iniciar mfa: %v", err)
	}

	segredo, ultimoCodigo, ultimoCodigoEm, err := s.ObterSegredoMfaPendente(ctx, userID)
	if err != nil {
		t.Fatalf("obter segredo pendente: %v", err)
	}
	if string(segredo) != "segredo-cifrado-1" {
		t.Fatalf("segredo pendente inesperado: %q", segredo)
	}
	if ultimoCodigo != "" || ultimoCodigoEm != nil {
		t.Fatalf("não deveria haver anti-replay antes de qualquer confirmação: codigo=%q em=%v", ultimoCodigo, ultimoCodigoEm)
	}

	// MFA não fica ativo antes de confirmar.
	var ativoAntes bool
	if err := p.QueryRow(ctx, `SELECT mfa_ativo FROM users WHERE id=$1`, userID).Scan(&ativoAntes); err != nil {
		t.Fatalf("consultar mfa_ativo: %v", err)
	}
	if ativoAntes {
		t.Fatal("mfa_ativo deveria ser false antes de ConfirmarMfa")
	}

	// IniciarMfa de novo enquanto ainda não confirmado deve ser permitido
	// (reconfigurar antes de terminar o enrollment não está bloqueado).
	if err := s.IniciarMfa(ctx, userID, []byte("segredo-cifrado-2")); err != nil {
		t.Fatalf("reiniciar mfa antes de confirmar: %v", err)
	}

	agora := time.Now().Truncate(time.Second)
	if err := s.ConfirmarMfa(ctx, userID, "123456", agora); err != nil {
		t.Fatalf("confirmar mfa: %v", err)
	}

	var ativoDepois bool
	if err := p.QueryRow(ctx, `SELECT mfa_ativo FROM users WHERE id=$1`, userID).Scan(&ativoDepois); err != nil {
		t.Fatalf("consultar mfa_ativo: %v", err)
	}
	if !ativoDepois {
		t.Fatal("mfa_ativo deveria ser true após ConfirmarMfa")
	}

	// IniciarMfa deve recusar com ErrMfaJaAtivo agora que está confirmado.
	if err := s.IniciarMfa(ctx, userID, []byte("segredo-cifrado-3")); err != ErrMfaJaAtivo {
		t.Fatalf("esperava ErrMfaJaAtivo com mfa já confirmado; got=%v", err)
	}

	// Anti-replay: o mesmo código usado persiste em mfa_ultimo_codigo_usado.
	_, ultimoCodigoDepois, ultimoCodigoEmDepois, err := s.ObterSegredoMfaPendente(ctx, userID)
	if err != nil {
		t.Fatalf("obter segredo após confirmar: %v", err)
	}
	if ultimoCodigoDepois != "123456" || ultimoCodigoEmDepois == nil {
		t.Fatalf("estado de anti-replay não persistido: codigo=%q em=%v", ultimoCodigoDepois, ultimoCodigoEmDepois)
	}

	if err := s.DesativarMfa(ctx, userID); err != nil {
		t.Fatalf("desativar mfa: %v", err)
	}
	var ativoFinal bool
	var segredoFinal []byte
	if err := p.QueryRow(ctx, `SELECT mfa_ativo, mfa_segredo_cifrado FROM users WHERE id=$1`, userID).Scan(&ativoFinal, &segredoFinal); err != nil {
		t.Fatalf("consultar estado final do mfa: %v", err)
	}
	if ativoFinal || segredoFinal != nil {
		t.Fatalf("mfa deveria estar totalmente desligado após DesativarMfa: ativo=%v segredo=%v", ativoFinal, segredoFinal)
	}
}

// TestExcluirConta_BloqueadaComTenantFilho cobre RN07: a exclusão não roda
// (nem parcialmente) se o tenant do usuário tiver 1+ filhos vinculados.
func TestExcluirConta_BloqueadaComTenantFilho(t *testing.T) {
	p := pool(t)
	s := New(p)
	ctx := context.Background()
	userID, tenantID := criarUsuarioSenha(t, s, "itg-excl-bloqueada@example.com")

	filho, err := s.CriarTenant(ctx, "itg-excl-filho", "cliente", &tenantID, []byte("k"))
	if err != nil {
		t.Fatalf("criar tenant filho: %v", err)
	}
	t.Cleanup(func() { _, _ = p.Exec(context.Background(), `DELETE FROM tenants WHERE id=$1`, filho.ID) })

	if err := s.ExcluirConta(ctx, userID, tenantID); err != ErrTenantAtivoVinculado {
		t.Fatalf("esperava ErrTenantAtivoVinculado; got=%v", err)
	}

	// Nada deve ter sido anonimizado (nunca excluir parcialmente).
	var email string
	if err := p.QueryRow(ctx, `SELECT email FROM users WHERE id=$1`, userID).Scan(&email); err != nil {
		t.Fatalf("verificar usuário após bloqueio: %v", err)
	}
	if email != "itg-excl-bloqueada@example.com" {
		t.Fatalf("usuário não deveria ter sido alterado; email=%q", email)
	}
}

// TestExcluirConta_Anonimiza cobre RF16: exclusão sem tenant vinculado
// anonimiza a conta (LGPD) — não deixa e-mail/nome originais.
func TestExcluirConta_Anonimiza(t *testing.T) {
	p := pool(t)
	s := New(p)
	ctx := context.Background()
	userID, tenantID := criarUsuarioSenha(t, s, "itg-excl-ok@example.com")

	if err := s.ExcluirConta(ctx, userID, tenantID); err != nil {
		t.Fatalf("excluir conta: %v", err)
	}

	var nome, email, provedor string
	var senhaHash *string
	if err := p.QueryRow(ctx, `SELECT nome, email, provedor, senha_hash FROM users WHERE id=$1`, userID).
		Scan(&nome, &email, &provedor, &senhaHash); err != nil {
		t.Fatalf("verificar usuário anonimizado: %v", err)
	}
	if nome != "" {
		t.Fatalf("nome deveria estar vazio; got=%q", nome)
	}
	if email == "itg-excl-ok@example.com" {
		t.Fatal("e-mail original não deveria sobreviver à anonimização")
	}
	if provedor != "removido" {
		t.Fatalf("provedor deveria ser 'removido'; got=%q", provedor)
	}
	if senhaHash != nil {
		t.Fatalf("senha_hash deveria ser NULL; got=%v", *senhaHash)
	}
}

func TestSolicitarEConfirmarTrocaEmail(t *testing.T) {
	p := pool(t)
	s := New(p)
	ctx := context.Background()
	userID, _ := criarUsuarioSenha(t, s, "itg-email-antigo@example.com")

	if err := s.SolicitarTrocaEmail(ctx, userID, "itg-email-novo@example.com", "hash-token-1", time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("solicitar troca de e-mail: %v", err)
	}

	var emailPendente string
	if err := p.QueryRow(ctx, `SELECT email_pendente FROM users WHERE id=$1`, userID).Scan(&emailPendente); err != nil {
		t.Fatalf("verificar email_pendente: %v", err)
	}
	if emailPendente != "itg-email-novo@example.com" {
		t.Fatalf("email_pendente inesperado: %q", emailPendente)
	}

	if err := s.ConfirmarTrocaEmail(ctx, "hash-token-1"); err != nil {
		t.Fatalf("confirmar troca de e-mail: %v", err)
	}

	var email string
	var pendenteDepois, tokenHashDepois *string
	if err := p.QueryRow(ctx, `SELECT email, email_pendente, email_token_hash FROM users WHERE id=$1`, userID).
		Scan(&email, &pendenteDepois, &tokenHashDepois); err != nil {
		t.Fatalf("verificar e-mail após confirmação: %v", err)
	}
	if email != "itg-email-novo@example.com" {
		t.Fatalf("email não trocado; got=%q", email)
	}
	if pendenteDepois != nil || tokenHashDepois != nil {
		t.Fatalf("estado pendente deveria ser limpo após confirmar: pendente=%v tokenHash=%v", pendenteDepois, tokenHashDepois)
	}
}

func TestSolicitarTrocaEmail_EmailJaCadastrado(t *testing.T) {
	p := pool(t)
	s := New(p)
	ctx := context.Background()
	_, _ = criarUsuarioSenha(t, s, "itg-email-existente@example.com")
	userID, _ := criarUsuarioSenha(t, s, "itg-email-solicitante@example.com")

	if err := s.SolicitarTrocaEmail(ctx, userID, "itg-email-existente@example.com", "hash-token-2", time.Now().Add(time.Hour)); err != ErrEmailJaCadastrado {
		t.Fatalf("esperava ErrEmailJaCadastrado; got=%v", err)
	}
}

// TestConfirmarTrocaEmail_TokenExpirado_ErroGenerico cobre RN08/RN04: um token
// expirado devolve o mesmo erro genérico de um token inexistente.
func TestConfirmarTrocaEmail_TokenExpirado_ErroGenerico(t *testing.T) {
	p := pool(t)
	s := New(p)
	ctx := context.Background()
	userID, _ := criarUsuarioSenha(t, s, "itg-email-expirado@example.com")

	if err := s.SolicitarTrocaEmail(ctx, userID, "itg-email-expirado-novo@example.com", "hash-token-3", time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("solicitar troca de e-mail: %v", err)
	}
	// Força a expiração diretamente no banco (SolicitarTrocaEmail sempre grava
	// now()+1h; simulamos o relógio ter avançado).
	if _, err := p.Exec(ctx, `UPDATE users SET email_token_expira_em = now() - interval '1 minute' WHERE id=$1`, userID); err != nil {
		t.Fatalf("expirar token manualmente: %v", err)
	}

	errExpirado := s.ConfirmarTrocaEmail(ctx, "hash-token-3")
	errInexistente := s.ConfirmarTrocaEmail(ctx, "hash-token-jamais-existiu")

	if errExpirado != ErrNaoEncontrado || errInexistente != ErrNaoEncontrado {
		t.Fatalf("token expirado e token inexistente deveriam devolver o mesmo ErrNaoEncontrado; expirado=%v inexistente=%v", errExpirado, errInexistente)
	}
}
