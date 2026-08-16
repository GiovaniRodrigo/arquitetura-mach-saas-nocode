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

// --- Dashboard (spec 004, RF04/RF05/RF06) ---------------------------------

// inserirUsuarioDireto cria uma linha em users diretamente por SQL (provedor
// 'senha'), sem passar pelo fluxo de cadastro — útil aqui só para ter um
// usuário vinculado a um tenant já existente (CriarTenantEUsuarioComSenha
// sempre cria um tenant raiz novo, o que não serve para testar um usuário no
// tenant FILHO de uma hierarquia já montada).
func inserirUsuarioDireto(t *testing.T, p *pgxpool.Pool, email, nome, tenantID, tipo string) string {
	t.Helper()
	var id string
	err := p.QueryRow(context.Background(),
		`INSERT INTO users (provedor, external_id, email, nome, tenant_id, tipo)
		 VALUES ('senha', $1, $1, $2, $3, $4) RETURNING id`,
		email, nome, tenantID, tipo,
	).Scan(&id)
	if err != nil {
		t.Fatalf("inserir usuário direto: %v", err)
	}
	t.Cleanup(func() { _, _ = p.Exec(context.Background(), `DELETE FROM users WHERE id=$1`, id) })
	return id
}

func TestUltimosAcessos_OrdemLimiteEIncluiFilhos(t *testing.T) {
	p := pool(t)
	s := New(p)
	ctx := context.Background()

	dono, err := s.CriarTenant(ctx, "itg-dash-dono", "dono", nil, []byte("k"))
	if err != nil {
		t.Fatalf("criar dono: %v", err)
	}
	t.Cleanup(func() { _, _ = p.Exec(context.Background(), `DELETE FROM tenants WHERE id=$1`, dono.ID) })

	filho, err := s.CriarTenant(ctx, "itg-dash-filho", "cliente", &dono.ID, []byte("k"))
	if err != nil {
		t.Fatalf("criar filho: %v", err)
	}
	t.Cleanup(func() { _, _ = p.Exec(context.Background(), `DELETE FROM tenants WHERE id=$1`, filho.ID) })

	outro, err := s.CriarTenant(ctx, "itg-dash-outro", "dono", nil, []byte("k"))
	if err != nil {
		t.Fatalf("criar outro: %v", err)
	}
	t.Cleanup(func() { _, _ = p.Exec(context.Background(), `DELETE FROM tenants WHERE id=$1`, outro.ID) })

	userDono := inserirUsuarioDireto(t, p, "itg-dash-dono@example.com", "Ana", dono.ID, "dono")
	userFilho := inserirUsuarioDireto(t, p, "itg-dash-filho@example.com", "Bia", filho.ID, "cliente")
	userOutro := inserirUsuarioDireto(t, p, "itg-dash-outro@example.com", "Carla", outro.ID, "dono")

	base := time.Now().Add(-time.Hour)
	inserirEvento := func(usuarioID, tenantID string, quando time.Time) {
		_, err := p.Exec(ctx,
			`INSERT INTO eventos_login (usuario_id, tenant_id, criado_em) VALUES ($1,$2,$3)`,
			usuarioID, tenantID, quando)
		if err != nil {
			t.Fatalf("inserir evento de login: %v", err)
		}
	}
	// 5 eventos vinculados (dono+filho), em ordem crescente de tempo, e 1 de
	// um tenant não vinculado (outro) que nunca deve aparecer.
	inserirEvento(userDono, dono.ID, base)
	inserirEvento(userFilho, filho.ID, base.Add(1*time.Minute))
	inserirEvento(userDono, dono.ID, base.Add(2*time.Minute))
	inserirEvento(userFilho, filho.ID, base.Add(3*time.Minute))
	inserirEvento(userDono, dono.ID, base.Add(4*time.Minute)) // mais recente
	inserirEvento(userOutro, outro.ID, base.Add(5*time.Minute))

	todos, err := s.UltimosAcessos(ctx, dono.ID, 10)
	if err != nil {
		t.Fatalf("últimos acessos: %v", err)
	}
	if len(todos) != 5 {
		t.Fatalf("esperava 5 eventos (dono+filho, sem outro); got=%d (%+v)", len(todos), todos)
	}
	// timestamptz do Postgres tem precisão de microssegundos (vs. nanossegundos
	// do time.Now() em Go), então a comparação usa tolerância em vez de Equal.
	if todos[0].UsuarioNome != "Ana" || todos[0].CriadoEm.Sub(base.Add(4*time.Minute)).Abs() > time.Millisecond {
		t.Fatalf("primeiro evento deveria ser o mais recente (dono, +4min); got=%+v", todos[0])
	}
	for i := 1; i < len(todos); i++ {
		if todos[i-1].CriadoEm.Before(todos[i].CriadoEm) {
			t.Fatalf("eventos fora de ordem (deveria ser desc): %+v", todos)
		}
	}

	limitados, err := s.UltimosAcessos(ctx, dono.ID, 2)
	if err != nil {
		t.Fatalf("últimos acessos com limite: %v", err)
	}
	if len(limitados) != 2 {
		t.Fatalf("esperava 2 eventos com limite=2; got=%d", len(limitados))
	}
}

func TestListarFeedback_FiltroEHierarquia(t *testing.T) {
	p := pool(t)
	s := New(p)
	ctx := context.Background()

	dono, err := s.CriarTenant(ctx, "itg-fb-dono", "dono", nil, []byte("k"))
	if err != nil {
		t.Fatalf("criar dono: %v", err)
	}
	t.Cleanup(func() { _, _ = p.Exec(context.Background(), `DELETE FROM tenants WHERE id=$1`, dono.ID) })

	filho, err := s.CriarTenant(ctx, "itg-fb-filho", "cliente", &dono.ID, []byte("k"))
	if err != nil {
		t.Fatalf("criar filho: %v", err)
	}
	t.Cleanup(func() { _, _ = p.Exec(context.Background(), `DELETE FROM tenants WHERE id=$1`, filho.ID) })

	outro, err := s.CriarTenant(ctx, "itg-fb-outro", "dono", nil, []byte("k"))
	if err != nil {
		t.Fatalf("criar outro: %v", err)
	}
	t.Cleanup(func() { _, _ = p.Exec(context.Background(), `DELETE FROM tenants WHERE id=$1`, outro.ID) })

	inserirFeedback := func(tenantID, mensagem, status string) string {
		var id string
		if err := p.QueryRow(ctx,
			`INSERT INTO feedback (tenant_id, mensagem, status) VALUES ($1,$2,$3) RETURNING id`,
			tenantID, mensagem, status,
		).Scan(&id); err != nil {
			t.Fatalf("inserir feedback: %v", err)
		}
		return id
	}
	inserirFeedback(dono.ID, "msg-dono-pendente", "pendente")
	respondidoID := inserirFeedback(dono.ID, "msg-dono-respondido", "respondido")
	inserirFeedback(filho.ID, "msg-filho-pendente", "pendente")
	inserirFeedback(outro.ID, "msg-outro-pendente", "pendente") // fora da hierarquia

	todos, err := s.ListarFeedback(ctx, dono.ID, nil)
	if err != nil {
		t.Fatalf("listar feedback sem filtro: %v", err)
	}
	if len(todos) != 3 {
		t.Fatalf("esperava 3 itens (dono+filho, sem outro); got=%d (%+v)", len(todos), todos)
	}

	pendente := "pendente"
	filtrados, err := s.ListarFeedback(ctx, dono.ID, &pendente)
	if err != nil {
		t.Fatalf("listar feedback com filtro: %v", err)
	}
	if len(filtrados) != 2 {
		t.Fatalf("esperava 2 itens pendentes (dono+filho); got=%d (%+v)", len(filtrados), filtrados)
	}
	for _, f := range filtrados {
		if f.Status != "pendente" {
			t.Fatalf("filtro de status vazou item não-pendente: %+v", f)
		}
	}

	// RN03: pendente → respondido é a única transição válida.
	pendenteID := filtrados[0].ID
	atualizado, err := s.AtualizarStatusFeedback(ctx, pendenteID, "respondido")
	if err != nil {
		t.Fatalf("atualizar pendente->respondido: %v", err)
	}
	if atualizado.Status != "respondido" {
		t.Fatalf("status não atualizado: %+v", atualizado)
	}

	// RN03: respondido → pendente é proibida. A cláusula WHERE do UPDATE só
	// atinge linhas com status='pendente', então tentar "voltar" um item já
	// respondido não afeta nenhuma linha (ErrNaoEncontrado) e o status
	// permanece 'respondido'.
	if _, err := s.AtualizarStatusFeedback(ctx, respondidoID, "pendente"); err != ErrNaoEncontrado {
		t.Fatalf("esperava ErrNaoEncontrado ao tentar respondido->pendente; got=%v", err)
	}
	inalterado, err := s.ListarFeedback(ctx, dono.ID, nil)
	if err != nil {
		t.Fatalf("relistar feedback: %v", err)
	}
	for _, f := range inalterado {
		if f.ID == respondidoID && f.Status != "respondido" {
			t.Fatalf("transição respondido->pendente não deveria ter efeito; got=%+v", f)
		}
	}

	if _, err := s.AtualizarStatusFeedback(ctx, "00000000-0000-0000-0000-000000000099", "respondido"); err != ErrNaoEncontrado {
		t.Fatalf("esperava ErrNaoEncontrado para id inexistente; got=%v", err)
	}
}

func TestResumoFinanceiro_SomaMesCorrenteEHierarquia(t *testing.T) {
	p := pool(t)
	s := New(p)
	ctx := context.Background()

	dono, err := s.CriarTenant(ctx, "itg-fin-dono", "dono", nil, []byte("k"))
	if err != nil {
		t.Fatalf("criar dono: %v", err)
	}
	t.Cleanup(func() { _, _ = p.Exec(context.Background(), `DELETE FROM tenants WHERE id=$1`, dono.ID) })

	filho, err := s.CriarTenant(ctx, "itg-fin-filho", "cliente", &dono.ID, []byte("k"))
	if err != nil {
		t.Fatalf("criar filho: %v", err)
	}
	t.Cleanup(func() { _, _ = p.Exec(context.Background(), `DELETE FROM tenants WHERE id=$1`, filho.ID) })

	outro, err := s.CriarTenant(ctx, "itg-fin-outro", "dono", nil, []byte("k"))
	if err != nil {
		t.Fatalf("criar outro: %v", err)
	}
	t.Cleanup(func() { _, _ = p.Exec(context.Background(), `DELETE FROM tenants WHERE id=$1`, outro.ID) })

	semAssinatura, err := s.CriarTenant(ctx, "itg-fin-sem-assinatura", "dono", nil, []byte("k"))
	if err != nil {
		t.Fatalf("criar tenant sem assinatura: %v", err)
	}
	t.Cleanup(func() { _, _ = p.Exec(context.Background(), `DELETE FROM tenants WHERE id=$1`, semAssinatura.ID) })

	inserirAssinatura := func(tenantID string, centavos int64, competencia string) {
		_, err := p.Exec(ctx,
			`INSERT INTO assinaturas_tenant (tenant_id, valor_centavos, competencia) VALUES ($1,$2,$3::date)`,
			tenantID, centavos, competencia+"-01")
		if err != nil {
			t.Fatalf("inserir assinatura: %v", err)
		}
	}
	mesCorrente := time.Now().Format("2006-01")
	mesPassado := time.Now().AddDate(0, -1, 0).Format("2006-01")

	inserirAssinatura(dono.ID, 10000, mesCorrente)
	inserirAssinatura(filho.ID, 5000, mesCorrente)
	inserirAssinatura(dono.ID, 7000, mesPassado)     // fora do mês corrente
	inserirAssinatura(outro.ID, 999999, mesCorrente) // fora da hierarquia

	resumo, err := s.ResumoFinanceiro(ctx, dono.ID)
	if err != nil {
		t.Fatalf("resumo financeiro: %v", err)
	}
	if resumo.ReceitaTotalCentavos != 15000 {
		t.Fatalf("receita total inesperada: got=%d, want=15000 (%+v)", resumo.ReceitaTotalCentavos, resumo)
	}
	if resumo.Moeda != "BRL" {
		t.Fatalf("moeda inesperada: %q", resumo.Moeda)
	}
	if resumo.Competencia != mesCorrente {
		t.Fatalf("competência inesperada: got=%q want=%q", resumo.Competencia, mesCorrente)
	}

	semDados, err := s.ResumoFinanceiro(ctx, semAssinatura.ID)
	if err != nil {
		t.Fatalf("resumo financeiro (sem dados): %v", err)
	}
	if semDados.ReceitaTotalCentavos != 0 {
		t.Fatalf("tenant sem assinatura deveria somar 0; got=%d", semDados.ReceitaTotalCentavos)
	}
}

func TestRegistrarEventoLogin(t *testing.T) {
	p := pool(t)
	s := New(p)
	ctx := context.Background()

	dono, err := s.CriarTenant(ctx, "itg-evt-dono", "dono", nil, []byte("k"))
	if err != nil {
		t.Fatalf("criar dono: %v", err)
	}
	t.Cleanup(func() { _, _ = p.Exec(context.Background(), `DELETE FROM tenants WHERE id=$1`, dono.ID) })

	userID := inserirUsuarioDireto(t, p, "itg-evt@example.com", "Ana", dono.ID, "dono")

	if err := s.RegistrarEventoLogin(ctx, userID, dono.ID); err != nil {
		t.Fatalf("registrar evento de login: %v", err)
	}

	eventos, err := s.UltimosAcessos(ctx, dono.ID, 10)
	if err != nil {
		t.Fatalf("últimos acessos: %v", err)
	}
	if len(eventos) != 1 || eventos[0].UsuarioNome != "Ana" {
		t.Fatalf("evento registrado não apareceu em últimos acessos: %+v", eventos)
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

// TestAcessosPorMes_ZeroPreenchidoEIncluiHierarquia cobre o gráfico "Acessos
// por mês" do Dashboard: 6 meses (mais antigo primeiro), zero-preenchidos, com
// eventos de fora da janela e de fora da hierarquia excluídos.
func TestAcessosPorMes_ZeroPreenchidoEIncluiHierarquia(t *testing.T) {
	p := pool(t)
	s := New(p)
	ctx := context.Background()

	dono, err := s.CriarTenant(ctx, "itg-acessosmes-dono", "dono", nil, []byte("k"))
	if err != nil {
		t.Fatalf("criar dono: %v", err)
	}
	t.Cleanup(func() { _, _ = p.Exec(context.Background(), `DELETE FROM tenants WHERE id=$1`, dono.ID) })

	filho, err := s.CriarTenant(ctx, "itg-acessosmes-filho", "cliente", &dono.ID, []byte("k"))
	if err != nil {
		t.Fatalf("criar filho: %v", err)
	}
	t.Cleanup(func() { _, _ = p.Exec(context.Background(), `DELETE FROM tenants WHERE id=$1`, filho.ID) })

	outro, err := s.CriarTenant(ctx, "itg-acessosmes-outro", "dono", nil, []byte("k"))
	if err != nil {
		t.Fatalf("criar outro: %v", err)
	}
	t.Cleanup(func() { _, _ = p.Exec(context.Background(), `DELETE FROM tenants WHERE id=$1`, outro.ID) })

	userDono := inserirUsuarioDireto(t, p, "itg-acessosmes-dono@example.com", "Ana", dono.ID, "dono")
	userFilho := inserirUsuarioDireto(t, p, "itg-acessosmes-filho@example.com", "Bia", filho.ID, "cliente")
	userOutro := inserirUsuarioDireto(t, p, "itg-acessosmes-outro@example.com", "Carla", outro.ID, "dono")

	inserirEvento := func(usuarioID, tenantID string, quando time.Time) {
		_, err := p.Exec(ctx,
			`INSERT INTO eventos_login (usuario_id, tenant_id, criado_em) VALUES ($1,$2,$3)`,
			usuarioID, tenantID, quando)
		if err != nil {
			t.Fatalf("inserir evento de login: %v", err)
		}
	}
	agora := time.Now()
	inserirEvento(userDono, dono.ID, agora)                   // mês corrente
	inserirEvento(userFilho, filho.ID, agora)                 // mês corrente (via filho)
	inserirEvento(userDono, dono.ID, agora.AddDate(0, -5, 0)) // mês mais antigo da janela
	inserirEvento(userDono, dono.ID, agora.AddDate(0, -7, 0)) // fora da janela de 6 meses
	inserirEvento(userOutro, outro.ID, agora)                 // fora da hierarquia

	pontos, err := s.AcessosPorMes(ctx, dono.ID)
	if err != nil {
		t.Fatalf("acessos por mês: %v", err)
	}
	if len(pontos) != 6 {
		t.Fatalf("esperava 6 pontos (zero-preenchidos); got=%d (%+v)", len(pontos), pontos)
	}

	mesCorrente := agora.Format("2006-01")
	mesMaisAntigo := agora.AddDate(0, -5, 0).Format("2006-01")
	if pontos[0].Competencia != mesMaisAntigo {
		t.Fatalf("primeiro ponto deveria ser o mês mais antigo da janela; got=%q want=%q", pontos[0].Competencia, mesMaisAntigo)
	}
	if pontos[len(pontos)-1].Competencia != mesCorrente {
		t.Fatalf("último ponto deveria ser o mês corrente; got=%q want=%q", pontos[len(pontos)-1].Competencia, mesCorrente)
	}

	totais := map[string]int32{}
	for _, p := range pontos {
		totais[p.Competencia] = p.Total
	}
	if totais[mesCorrente] != 2 {
		t.Fatalf("mês corrente deveria somar 2 (dono+filho); got=%d", totais[mesCorrente])
	}
	if totais[mesMaisAntigo] != 1 {
		t.Fatalf("mês mais antigo da janela deveria somar 1; got=%d", totais[mesMaisAntigo])
	}
}

// TestReceitaPorMes_ZeroPreenchidoEIncluiHierarquia cobre o gráfico "Receita
// de assinatura" do Dashboard: 6 meses (mais antigo primeiro), zero-
// preenchidos, com assinaturas de fora da janela e de fora da hierarquia
// excluídas.
func TestReceitaPorMes_ZeroPreenchidoEIncluiHierarquia(t *testing.T) {
	p := pool(t)
	s := New(p)
	ctx := context.Background()

	dono, err := s.CriarTenant(ctx, "itg-receitames-dono", "dono", nil, []byte("k"))
	if err != nil {
		t.Fatalf("criar dono: %v", err)
	}
	t.Cleanup(func() { _, _ = p.Exec(context.Background(), `DELETE FROM tenants WHERE id=$1`, dono.ID) })

	filho, err := s.CriarTenant(ctx, "itg-receitames-filho", "cliente", &dono.ID, []byte("k"))
	if err != nil {
		t.Fatalf("criar filho: %v", err)
	}
	t.Cleanup(func() { _, _ = p.Exec(context.Background(), `DELETE FROM tenants WHERE id=$1`, filho.ID) })

	outro, err := s.CriarTenant(ctx, "itg-receitames-outro", "dono", nil, []byte("k"))
	if err != nil {
		t.Fatalf("criar outro: %v", err)
	}
	t.Cleanup(func() { _, _ = p.Exec(context.Background(), `DELETE FROM tenants WHERE id=$1`, outro.ID) })

	inserirAssinatura := func(tenantID string, centavos int64, competencia string) {
		_, err := p.Exec(ctx,
			`INSERT INTO assinaturas_tenant (tenant_id, valor_centavos, competencia) VALUES ($1,$2,$3::date)`,
			tenantID, centavos, competencia+"-01")
		if err != nil {
			t.Fatalf("inserir assinatura: %v", err)
		}
	}
	agora := time.Now()
	mesCorrente := agora.Format("2006-01")
	mesMaisAntigo := agora.AddDate(0, -5, 0).Format("2006-01")
	foraDaJanela := agora.AddDate(0, -7, 0).Format("2006-01")

	inserirAssinatura(dono.ID, 10000, mesCorrente)
	inserirAssinatura(filho.ID, 5000, mesCorrente)
	inserirAssinatura(dono.ID, 3000, mesMaisAntigo)
	inserirAssinatura(dono.ID, 999999, foraDaJanela)
	inserirAssinatura(outro.ID, 888888, mesCorrente)

	pontos, err := s.ReceitaPorMes(ctx, dono.ID)
	if err != nil {
		t.Fatalf("receita por mês: %v", err)
	}
	if len(pontos) != 6 {
		t.Fatalf("esperava 6 pontos (zero-preenchidos); got=%d (%+v)", len(pontos), pontos)
	}
	if pontos[0].Competencia != mesMaisAntigo {
		t.Fatalf("primeiro ponto deveria ser o mês mais antigo da janela; got=%q want=%q", pontos[0].Competencia, mesMaisAntigo)
	}
	if pontos[len(pontos)-1].Competencia != mesCorrente {
		t.Fatalf("último ponto deveria ser o mês corrente; got=%q want=%q", pontos[len(pontos)-1].Competencia, mesCorrente)
	}

	totais := map[string]int64{}
	for _, p := range pontos {
		totais[p.Competencia] = p.ValorCentavos
	}
	if totais[mesCorrente] != 15000 {
		t.Fatalf("mês corrente deveria somar 15000 (dono+filho); got=%d", totais[mesCorrente])
	}
	if totais[mesMaisAntigo] != 3000 {
		t.Fatalf("mês mais antigo da janela deveria somar 3000; got=%d", totais[mesMaisAntigo])
	}
}
