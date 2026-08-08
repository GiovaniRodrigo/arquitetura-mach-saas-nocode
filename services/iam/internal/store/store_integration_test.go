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
