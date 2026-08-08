package server

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonv1 "github.com/machv4/platform/gen/go/construtor/common/v1"
	iamv1 "github.com/machv4/platform/gen/go/construtor/iam/v1"
	"github.com/machv4/platform/pkg/tenantctx"
	"github.com/machv4/platform/services/iam/auth"
	"github.com/machv4/platform/services/iam/internal/permissions"
	"github.com/machv4/platform/services/iam/internal/store"
)

type fakeStore struct {
	perms []permissions.Permissao
	err   error
	// resposta do upsert de identidade
	userID, tenantID, tipo string
	upsertErr              error
	// tenants (ListarFilhos/CriarTenant)
	filhos       []store.Tenant
	filhosErr    error
	criarTenant  store.Tenant
	criarErr     error
	nomeCriado   string
	tipoCriado   string
	parentCriado *string
	// tenants (ObterTenant/AtualizarTenant/ExcluirTenant)
	obterTenant     store.Tenant
	obterErr        error
	atualizarTenant store.Tenant
	atualizarErr    error
	idAtualizado    string
	nomeAtualizado  string
	excluirErr      error
	idExcluido      string
	// cadastro por senha (CriarTenantEUsuarioComSenha)
	registrarUserID, registrarTenantID                   string
	registrarErr                                         error
	nomeRegistrado, emailRegistrado, senhaHashRegistrada string
	nomeTenantRegistrado                                 string
	// login por senha (ObterUsuarioPorEmailSenha)
	obterSenhaUserID, obterSenhaTenantID, obterSenhaTipo, obterSenhaHash string
	obterSenhaErr                                                        error
	emailConsultado                                                      string
	// Dashboard: evento de login (RegistrarEventoLogin)
	registrarEventoErr              error
	eventoUsuarioID, eventoTenantID string
	// Dashboard: últimos acessos (UltimosAcessos)
	ultimosAcessos       []store.EventoLogin
	ultimosAcessosErr    error
	ultimosAcessosTenant string
	ultimosAcessosLimite int
	// Dashboard: feedback (ListarFeedback/AtualizarStatusFeedback)
	listarFeedback                                 []store.Feedback
	listarFeedbackErr                              error
	listarFeedbackTenant                           string
	listarFeedbackStatus                           *string
	atualizarFeedback                              store.Feedback
	atualizarFeedbackErr                           error
	feedbackIDAtualizado, feedbackStatusAtualizado string
	// Dashboard: resumo financeiro (ResumoFinanceiro)
	resumoFinanceiro       store.ResumoFinanceiro
	resumoFinanceiroErr    error
	resumoFinanceiroTenant string

	// Conta/Configuração (spec 004, RF14-RF18)
	perfilNome, perfilFotoURL, perfilUserIDChamado                         string
	perfilErr                                                              error
	emailUsuario                                                           string
	emailUsuarioErr                                                        error
	hashSenha                                                              string
	hashSenhaErr                                                           error
	senhaAtualizadaUserID, senhaAtualizadaHash                             string
	atualizarSenhaErr                                                      error
	iniciarMfaUserID                                                       string
	iniciarMfaSegredo                                                      []byte
	iniciarMfaErr                                                          error
	segredoMfaPendente                                                     []byte
	ultimoCodigoMfa                                                        string
	ultimoCodigoMfaEm                                                      *time.Time
	obterSegredoMfaErr                                                     error
	confirmarMfaUserID, confirmarMfaCodigo                                 string
	confirmarMfaErr                                                        error
	desativarMfaUserID                                                     string
	desativarMfaErr                                                        error
	excluirContaUserID, excluirContaTenantID                               string
	excluirContaErr                                                        error
	solicitarTrocaUserID, solicitarTrocaNovoEmail, solicitarTrocaTokenHash string
	solicitarTrocaExpira                                                   time.Time
	solicitarTrocaErr                                                      error
	confirmarTrocaTokenHash                                                string
	confirmarTrocaErr                                                      error
}

func (f *fakeStore) RegistrarEventoLogin(_ context.Context, usuarioID, tenantID string) error {
	f.eventoUsuarioID, f.eventoTenantID = usuarioID, tenantID
	return f.registrarEventoErr
}

func (f *fakeStore) UltimosAcessos(_ context.Context, tenantID string, limite int) ([]store.EventoLogin, error) {
	f.ultimosAcessosTenant, f.ultimosAcessosLimite = tenantID, limite
	return f.ultimosAcessos, f.ultimosAcessosErr
}

func (f *fakeStore) ListarFeedback(_ context.Context, tenantID string, status *string) ([]store.Feedback, error) {
	f.listarFeedbackTenant, f.listarFeedbackStatus = tenantID, status
	return f.listarFeedback, f.listarFeedbackErr
}

func (f *fakeStore) AtualizarStatusFeedback(_ context.Context, id, novoStatus string) (store.Feedback, error) {
	f.feedbackIDAtualizado, f.feedbackStatusAtualizado = id, novoStatus
	return f.atualizarFeedback, f.atualizarFeedbackErr
}

func (f *fakeStore) ResumoFinanceiro(_ context.Context, tenantID string) (store.ResumoFinanceiro, error) {
	f.resumoFinanceiroTenant = tenantID
	return f.resumoFinanceiro, f.resumoFinanceiroErr
}

func (f *fakeStore) PermissoesDe(context.Context, []string) ([]permissions.Permissao, error) {
	return f.perms, f.err
}

func (f *fakeStore) UpsertUsuarioThirdParty(context.Context, string, string, string, string) (string, string, string, error) {
	return f.userID, f.tenantID, f.tipo, f.upsertErr
}

func (f *fakeStore) ListarFilhos(context.Context, string) ([]store.Tenant, error) {
	return f.filhos, f.filhosErr
}

func (f *fakeStore) CriarTenant(_ context.Context, nome, tipo string, parentID *string, _ []byte) (store.Tenant, error) {
	f.nomeCriado, f.tipoCriado, f.parentCriado = nome, tipo, parentID
	return f.criarTenant, f.criarErr
}

func (f *fakeStore) ObterTenant(context.Context, string) (store.Tenant, error) {
	return f.obterTenant, f.obterErr
}

func (f *fakeStore) AtualizarTenant(_ context.Context, id, nome string) (store.Tenant, error) {
	f.idAtualizado, f.nomeAtualizado = id, nome
	return f.atualizarTenant, f.atualizarErr
}

func (f *fakeStore) ExcluirTenant(_ context.Context, id string) error {
	f.idExcluido = id
	return f.excluirErr
}

func (f *fakeStore) CriarTenantEUsuarioComSenha(_ context.Context, nomeUsuario, email, senhaHash, nomeTenant string) (string, string, error) {
	f.nomeRegistrado, f.emailRegistrado, f.senhaHashRegistrada, f.nomeTenantRegistrado = nomeUsuario, email, senhaHash, nomeTenant
	return f.registrarUserID, f.registrarTenantID, f.registrarErr
}

func (f *fakeStore) ObterUsuarioPorEmailSenha(_ context.Context, email string) (string, string, string, string, error) {
	f.emailConsultado = email
	return f.obterSenhaUserID, f.obterSenhaTenantID, f.obterSenhaTipo, f.obterSenhaHash, f.obterSenhaErr
}

func (f *fakeStore) AtualizarPerfil(_ context.Context, userID, nome, fotoURL string) error {
	f.perfilUserIDChamado, f.perfilNome, f.perfilFotoURL = userID, nome, fotoURL
	return f.perfilErr
}

func (f *fakeStore) ObterEmailUsuario(context.Context, string) (string, error) {
	return f.emailUsuario, f.emailUsuarioErr
}

func (f *fakeStore) ObterHashSenha(context.Context, string) (string, error) {
	return f.hashSenha, f.hashSenhaErr
}

func (f *fakeStore) AtualizarSenha(_ context.Context, userID, novoHash string) error {
	f.senhaAtualizadaUserID, f.senhaAtualizadaHash = userID, novoHash
	return f.atualizarSenhaErr
}

func (f *fakeStore) IniciarMfa(_ context.Context, userID string, segredoCifrado []byte) error {
	f.iniciarMfaUserID, f.iniciarMfaSegredo = userID, segredoCifrado
	return f.iniciarMfaErr
}

func (f *fakeStore) ObterSegredoMfaPendente(context.Context, string) ([]byte, string, *time.Time, error) {
	return f.segredoMfaPendente, f.ultimoCodigoMfa, f.ultimoCodigoMfaEm, f.obterSegredoMfaErr
}

func (f *fakeStore) ConfirmarMfa(_ context.Context, userID, codigoUsado string, _ time.Time) error {
	f.confirmarMfaUserID, f.confirmarMfaCodigo = userID, codigoUsado
	return f.confirmarMfaErr
}

func (f *fakeStore) DesativarMfa(_ context.Context, userID string) error {
	f.desativarMfaUserID = userID
	return f.desativarMfaErr
}

func (f *fakeStore) ExcluirConta(_ context.Context, userID, tenantID string) error {
	f.excluirContaUserID, f.excluirContaTenantID = userID, tenantID
	return f.excluirContaErr
}

func (f *fakeStore) SolicitarTrocaEmail(_ context.Context, userID, novoEmail, tokenHash string, expiraEm time.Time) error {
	f.solicitarTrocaUserID, f.solicitarTrocaNovoEmail, f.solicitarTrocaTokenHash, f.solicitarTrocaExpira = userID, novoEmail, tokenHash, expiraEm
	return f.solicitarTrocaErr
}

func (f *fakeStore) ConfirmarTrocaEmail(_ context.Context, tokenHash string) error {
	f.confirmarTrocaTokenHash = tokenHash
	return f.confirmarTrocaErr
}

func newServer(t *testing.T, store Store) (*IAMServer, *auth.Issuer) {
	t.Helper()
	priv, err := auth.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	iss := auth.NewIssuer(priv, time.Hour)
	return New(auth.NewValidator(&priv.PublicKey), iss, store), iss
}

// chaveMfaTeste é a chave AES-256-GCM fixa usada pelos testes de MFA — nunca
// use uma chave fixa em produção (ver services/iam/cmd/main.go, loadMfaKey).
var chaveMfaTeste = [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}

func newServerComMfa(t *testing.T, store Store) (*IAMServer, *auth.Issuer) {
	t.Helper()
	srv, iss := newServer(t, store)
	return srv.ComChaveMfa(chaveMfaTeste), iss
}

// ctxComUsuario monta um TenantContext completo (tenant + user_id) — as rotas
// de Conta/Configuração exigem user_id, diferente das rotas de tenants.
func ctxComUsuario(tenantID, userID, tipo string) context.Context {
	return tenantctx.NewContext(context.Background(), &commonv1.TenantContext{TenantId: tenantID, UserId: userID, Tipo: tipo})
}

func TestValidarToken_ValidoEInvalido(t *testing.T) {
	srv, iss := newServer(t, &fakeStore{})
	token, _ := iss.Issue("user-1", "tenant-A", "dono")

	resp, _ := srv.ValidarToken(context.Background(), &iamv1.ValidarTokenRequest{Jwt: token})
	if !resp.GetValido() || resp.GetTenantId() != "tenant-A" || resp.GetUserId() != "user-1" {
		t.Fatalf("token válido mal interpretado: %+v", resp)
	}

	bad, _ := srv.ValidarToken(context.Background(), &iamv1.ValidarTokenRequest{Jwt: "lixo"})
	if bad.GetValido() {
		t.Fatal("token inválido não deveria validar")
	}
}

func TestAutenticarThirdParty_EmiteTokenValido(t *testing.T) {
	store := &fakeStore{userID: "user-9", tenantID: "tenant-padrao", tipo: "cliente"}
	srv, _ := newServer(t, store)

	resp, err := srv.AutenticarThirdParty(context.Background(), &iamv1.AutenticarThirdPartyRequest{
		Provedor: "google", ExternalId: "g-123", Email: "a@b.com", Nome: "Ana",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.GetUserId() != "user-9" || resp.GetTenantId() != "tenant-padrao" || resp.GetTipo() != "cliente" {
		t.Fatalf("identidade inesperada: %+v", resp)
	}
	// O JWT emitido deve validar no mesmo servidor.
	val, _ := srv.ValidarToken(context.Background(), &iamv1.ValidarTokenRequest{Jwt: resp.GetJwt()})
	if !val.GetValido() || val.GetUserId() != "user-9" || val.GetTipo() != "cliente" {
		t.Fatalf("token emitido não validou: %+v", val)
	}
}

func TestAutenticarThirdParty_RejeitaCamposObrigatorios(t *testing.T) {
	srv, _ := newServer(t, &fakeStore{})
	if _, err := srv.AutenticarThirdParty(context.Background(), &iamv1.AutenticarThirdPartyRequest{Provedor: "google"}); err == nil {
		t.Fatal("external_id ausente deveria falhar")
	}
}

func TestAvaliarPermissoes_SemTenantContext(t *testing.T) {
	srv, _ := newServer(t, &fakeStore{})
	if _, err := srv.AvaliarPermissoes(context.Background(), &iamv1.AvaliarPermissoesRequest{}); err == nil {
		t.Fatal("sem TenantContext deveria retornar erro Unauthenticated")
	}
}

func TestAvaliarPermissoes_MapaBooleano(t *testing.T) {
	loader := &fakeStore{perms: []permissions.Permissao{
		{BlindIndex: "bi-1", Papel: "dono", View: true, Click: true},
	}}
	srv, _ := newServer(t, loader)

	ctx := tenantctx.NewContext(context.Background(), &commonv1.TenantContext{TenantId: "tenant-A", Tipo: "dono"})
	resp, err := srv.AvaliarPermissoes(ctx, &iamv1.AvaliarPermissoesRequest{BlindIndexes: []string{"bi-1", "bi-2"}})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if !resp.Permissions["bi-1"].GetView() || !resp.Permissions["bi-1"].GetClick() {
		t.Fatalf("bi-1 deveria estar liberado: %+v", resp.Permissions["bi-1"])
	}
	// bi-2 sem permissão → fail-closed.
	if resp.Permissions["bi-2"].GetView() || resp.Permissions["bi-2"].GetClick() {
		t.Fatalf("bi-2 deveria estar negado: %+v", resp.Permissions["bi-2"])
	}
}

func TestListarTenants_SemTenantContext(t *testing.T) {
	srv, _ := newServer(t, &fakeStore{})
	if _, err := srv.ListarTenants(context.Background(), &iamv1.ListarTenantsRequest{}); err == nil {
		t.Fatal("sem TenantContext deveria retornar erro Unauthenticated")
	}
}

func TestListarTenants_DevolveFilhosDoTenantDoContexto(t *testing.T) {
	fs := &fakeStore{filhos: []store.Tenant{{ID: "t1", Nome: "Acme", Tipo: "cliente"}}}
	srv, _ := newServer(t, fs)

	ctx := tenantctx.NewContext(context.Background(), &commonv1.TenantContext{TenantId: "tenant-A", Tipo: "dono"})
	resp, err := srv.ListarTenants(ctx, &iamv1.ListarTenantsRequest{})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if len(resp.GetTenants()) != 1 || resp.GetTenants()[0].GetNome() != "Acme" {
		t.Fatalf("tenants inesperados: %+v", resp.GetTenants())
	}
}

func TestCriarTenant_SemTenantContext(t *testing.T) {
	srv, _ := newServer(t, &fakeStore{})
	if _, err := srv.CriarTenant(context.Background(), &iamv1.CriarTenantRequest{Nome: "Acme"}); err == nil {
		t.Fatal("sem TenantContext deveria retornar erro Unauthenticated")
	}
}

func TestCriarTenant_ClienteFinalNegado(t *testing.T) {
	srv, _ := newServer(t, &fakeStore{})
	ctx := tenantctx.NewContext(context.Background(), &commonv1.TenantContext{TenantId: "tenant-A", Tipo: "cliente"})
	if _, err := srv.CriarTenant(ctx, &iamv1.CriarTenantRequest{Nome: "Acme"}); err == nil {
		t.Fatal("cliente final não deveria poder criar tenant")
	}
}

func TestCriarTenant_NomeVazio(t *testing.T) {
	srv, _ := newServer(t, &fakeStore{})
	ctx := tenantctx.NewContext(context.Background(), &commonv1.TenantContext{TenantId: "tenant-A", Tipo: "dono"})
	if _, err := srv.CriarTenant(ctx, &iamv1.CriarTenantRequest{}); err == nil {
		t.Fatal("nome vazio deveria falhar")
	}
}

func TestCriarTenant_DonoCriaSobPropioTenant(t *testing.T) {
	fs := &fakeStore{criarTenant: store.Tenant{ID: "t9", Nome: "Acme", Tipo: "cliente"}}
	srv, _ := newServer(t, fs)

	ctx := tenantctx.NewContext(context.Background(), &commonv1.TenantContext{TenantId: "tenant-A", Tipo: "dono"})
	resp, err := srv.CriarTenant(ctx, &iamv1.CriarTenantRequest{Nome: "Acme"})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.GetId() != "t9" || resp.GetNome() != "Acme" {
		t.Fatalf("tenant inesperado: %+v", resp)
	}
	if fs.nomeCriado != "Acme" || fs.tipoCriado != "cliente" {
		t.Fatalf("store chamado com args inesperados: nome=%q tipo=%q", fs.nomeCriado, fs.tipoCriado)
	}
	if fs.parentCriado == nil || *fs.parentCriado != "tenant-A" {
		t.Fatalf("parent_id deveria ser o tenant do contexto: %+v", fs.parentCriado)
	}
}

func tenantACtx() context.Context {
	return tenantctx.NewContext(context.Background(), &commonv1.TenantContext{TenantId: "tenant-A", Tipo: "dono"})
}

func TestObterTenant_SemTenantContext(t *testing.T) {
	srv, _ := newServer(t, &fakeStore{})
	if _, err := srv.ObterTenant(context.Background(), &iamv1.ObterTenantRequest{Id: "t1"}); err == nil {
		t.Fatal("sem TenantContext deveria retornar erro Unauthenticated")
	}
}

func TestObterTenant_ClienteFinalNegado(t *testing.T) {
	srv, _ := newServer(t, &fakeStore{})
	ctx := tenantctx.NewContext(context.Background(), &commonv1.TenantContext{TenantId: "tenant-A", Tipo: "cliente"})
	if _, err := srv.ObterTenant(ctx, &iamv1.ObterTenantRequest{Id: "t1"}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("esperava PermissionDenied; got=%v", err)
	}
}

func TestObterTenant_ForaDaHierarquia_NotFound(t *testing.T) {
	parent := "tenant-B"
	fs := &fakeStore{obterTenant: store.Tenant{ID: "t1", Nome: "Acme", Tipo: "cliente", ParentID: &parent}}
	srv, _ := newServer(t, fs)
	if _, err := srv.ObterTenant(tenantACtx(), &iamv1.ObterTenantRequest{Id: "t1"}); status.Code(err) != codes.NotFound {
		t.Fatalf("tenant fora da hierarquia deveria ser NotFound; got=%v", err)
	}
}

func TestObterTenant_Inexistente_NotFound(t *testing.T) {
	fs := &fakeStore{obterErr: store.ErrNaoEncontrado}
	srv, _ := newServer(t, fs)
	if _, err := srv.ObterTenant(tenantACtx(), &iamv1.ObterTenantRequest{Id: "t1"}); status.Code(err) != codes.NotFound {
		t.Fatalf("esperava NotFound; got=%v", err)
	}
}

func TestObterTenant_FilhoDoContexto_OK(t *testing.T) {
	parent := "tenant-A"
	fs := &fakeStore{obterTenant: store.Tenant{ID: "t1", Nome: "Acme", Tipo: "cliente", ParentID: &parent}}
	srv, _ := newServer(t, fs)
	resp, err := srv.ObterTenant(tenantACtx(), &iamv1.ObterTenantRequest{Id: "t1"})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.GetId() != "t1" || resp.GetNome() != "Acme" {
		t.Fatalf("tenant inesperado: %+v", resp)
	}
}

func TestAtualizarTenant_NomeVazio(t *testing.T) {
	srv, _ := newServer(t, &fakeStore{})
	if _, err := srv.AtualizarTenant(tenantACtx(), &iamv1.AtualizarTenantRequest{Id: "t1"}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("nome vazio deveria falhar; got=%v", err)
	}
}

func TestAtualizarTenant_ForaDaHierarquia_NotFound(t *testing.T) {
	parent := "tenant-B"
	fs := &fakeStore{obterTenant: store.Tenant{ID: "t1", ParentID: &parent}}
	srv, _ := newServer(t, fs)
	if _, err := srv.AtualizarTenant(tenantACtx(), &iamv1.AtualizarTenantRequest{Id: "t1", Nome: "Novo"}); status.Code(err) != codes.NotFound {
		t.Fatalf("esperava NotFound; got=%v", err)
	}
	if fs.idAtualizado != "" {
		t.Fatal("store.AtualizarTenant não deveria ser chamado fora da hierarquia")
	}
}

func TestAtualizarTenant_FilhoDoContexto_OK(t *testing.T) {
	parent := "tenant-A"
	fs := &fakeStore{
		obterTenant:     store.Tenant{ID: "t1", ParentID: &parent},
		atualizarTenant: store.Tenant{ID: "t1", Nome: "Novo Nome", Tipo: "cliente"},
	}
	srv, _ := newServer(t, fs)
	resp, err := srv.AtualizarTenant(tenantACtx(), &iamv1.AtualizarTenantRequest{Id: "t1", Nome: "Novo Nome"})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.GetNome() != "Novo Nome" {
		t.Fatalf("nome inesperado: %+v", resp)
	}
	if fs.idAtualizado != "t1" || fs.nomeAtualizado != "Novo Nome" {
		t.Fatalf("store chamado com args inesperados: id=%q nome=%q", fs.idAtualizado, fs.nomeAtualizado)
	}
}

func TestExcluirTenant_ClienteFinalNegado(t *testing.T) {
	srv, _ := newServer(t, &fakeStore{})
	ctx := tenantctx.NewContext(context.Background(), &commonv1.TenantContext{TenantId: "tenant-A", Tipo: "cliente"})
	if _, err := srv.ExcluirTenant(ctx, &iamv1.ExcluirTenantRequest{Id: "t1"}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("esperava PermissionDenied; got=%v", err)
	}
}

func TestExcluirTenant_ForaDaHierarquia_NotFound(t *testing.T) {
	parent := "tenant-B"
	fs := &fakeStore{obterTenant: store.Tenant{ID: "t1", ParentID: &parent}}
	srv, _ := newServer(t, fs)
	if _, err := srv.ExcluirTenant(tenantACtx(), &iamv1.ExcluirTenantRequest{Id: "t1"}); status.Code(err) != codes.NotFound {
		t.Fatalf("esperava NotFound; got=%v", err)
	}
	if fs.idExcluido != "" {
		t.Fatal("store.ExcluirTenant não deveria ser chamado fora da hierarquia")
	}
}

func TestExcluirTenant_FilhoDoContexto_OK(t *testing.T) {
	parent := "tenant-A"
	fs := &fakeStore{obterTenant: store.Tenant{ID: "t1", ParentID: &parent}}
	srv, _ := newServer(t, fs)
	if _, err := srv.ExcluirTenant(tenantACtx(), &iamv1.ExcluirTenantRequest{Id: "t1"}); err != nil {
		t.Fatalf("erro: %v", err)
	}
	if fs.idExcluido != "t1" {
		t.Fatalf("store chamado com id inesperado: %q", fs.idExcluido)
	}
}

func TestRegistrarUsuario_Sucesso(t *testing.T) {
	fs := &fakeStore{registrarUserID: "user-1", registrarTenantID: "tenant-1"}
	srv, _ := newServer(t, fs)

	resp, err := srv.RegistrarUsuario(context.Background(), &iamv1.RegistrarUsuarioRequest{
		Nome: "Ana", Email: "ana@example.com", Senha: "12345678", NomeTenant: "Ana LTDA",
	})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.GetUserId() != "user-1" || resp.GetTenantId() != "tenant-1" || resp.GetTipo() != "dono" {
		t.Fatalf("identidade inesperada: %+v", resp)
	}
	val, _ := srv.ValidarToken(context.Background(), &iamv1.ValidarTokenRequest{Jwt: resp.GetJwt()})
	if !val.GetValido() || val.GetTenantId() != "tenant-1" || val.GetTipo() != "dono" {
		t.Fatalf("token emitido não validou: %+v", val)
	}

	if fs.nomeRegistrado != "Ana" || fs.emailRegistrado != "ana@example.com" || fs.nomeTenantRegistrado != "Ana LTDA" {
		t.Fatalf("store chamado com args inesperados: nome=%q email=%q tenant=%q", fs.nomeRegistrado, fs.emailRegistrado, fs.nomeTenantRegistrado)
	}
	// A senha nunca deve ser persistida em texto claro (RN03, RNF01).
	if fs.senhaHashRegistrada == "12345678" {
		t.Fatal("senha não deveria ir em texto claro ao store")
	}
	if bcrypt.CompareHashAndPassword([]byte(fs.senhaHashRegistrada), []byte("12345678")) != nil {
		t.Fatalf("hash gerado não corresponde à senha original: %q", fs.senhaHashRegistrada)
	}
}

func TestRegistrarUsuario_CamposObrigatorios(t *testing.T) {
	srv, _ := newServer(t, &fakeStore{})
	casos := []*iamv1.RegistrarUsuarioRequest{
		{Email: "a@b.com", Senha: "12345678", NomeTenant: "T"}, // sem nome
		{Nome: "Ana", Senha: "12345678", NomeTenant: "T"},      // sem email
		{Nome: "Ana", Email: "a@b.com", NomeTenant: "T"},       // sem senha
		{Nome: "Ana", Email: "a@b.com", Senha: "12345678"},     // sem nome_tenant
	}
	for _, c := range casos {
		if _, err := srv.RegistrarUsuario(context.Background(), c); status.Code(err) != codes.InvalidArgument {
			t.Fatalf("esperava InvalidArgument para %+v; got=%v", c, err)
		}
	}
}

func TestRegistrarUsuario_SenhaCurta(t *testing.T) {
	srv, _ := newServer(t, &fakeStore{})
	_, err := srv.RegistrarUsuario(context.Background(), &iamv1.RegistrarUsuarioRequest{
		Nome: "Ana", Email: "a@b.com", Senha: "1234567", NomeTenant: "T",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("senha com menos de 8 caracteres deveria falhar; got=%v", err)
	}
}

func TestRegistrarUsuario_EmailDuplicado(t *testing.T) {
	fs := &fakeStore{registrarErr: store.ErrEmailJaCadastrado}
	srv, _ := newServer(t, fs)
	_, err := srv.RegistrarUsuario(context.Background(), &iamv1.RegistrarUsuarioRequest{
		Nome: "Ana", Email: "a@b.com", Senha: "12345678", NomeTenant: "T",
	})
	if status.Code(err) != codes.AlreadyExists {
		t.Fatalf("e-mail duplicado deveria ser AlreadyExists; got=%v", err)
	}
}

func TestAutenticarSenha_Sucesso(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("12345678"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	fs := &fakeStore{obterSenhaUserID: "user-1", obterSenhaTenantID: "tenant-1", obterSenhaTipo: "dono", obterSenhaHash: string(hash)}
	srv, _ := newServer(t, fs)

	resp, err := srv.AutenticarSenha(context.Background(), &iamv1.AutenticarSenhaRequest{Email: "ana@example.com", Senha: "12345678"})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.GetUserId() != "user-1" || resp.GetTenantId() != "tenant-1" || resp.GetTipo() != "dono" {
		t.Fatalf("identidade inesperada: %+v", resp)
	}
	if fs.emailConsultado != "ana@example.com" {
		t.Fatalf("store consultado com e-mail inesperado: %q", fs.emailConsultado)
	}
}

// RN04: e-mail inexistente e senha incorreta devem devolver exatamente o mesmo
// erro, para não permitir enumeração de e-mails cadastrados.
func TestAutenticarSenha_EmailInexistenteESenhaIncorreta_MesmoErro(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("senha-correta"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}

	srvInexistente, _ := newServer(t, &fakeStore{obterSenhaErr: store.ErrNaoEncontrado})
	_, errInexistente := srvInexistente.AutenticarSenha(context.Background(), &iamv1.AutenticarSenhaRequest{Email: "fantasma@example.com", Senha: "qualquer"})

	srvSenhaErrada, _ := newServer(t, &fakeStore{obterSenhaUserID: "user-1", obterSenhaTenantID: "tenant-1", obterSenhaTipo: "dono", obterSenhaHash: string(hash)})
	_, errSenhaErrada := srvSenhaErrada.AutenticarSenha(context.Background(), &iamv1.AutenticarSenhaRequest{Email: "ana@example.com", Senha: "senha-errada"})

	if status.Code(errInexistente) != codes.Unauthenticated || status.Code(errSenhaErrada) != codes.Unauthenticated {
		t.Fatalf("ambos deveriam ser Unauthenticated; inexistente=%v senhaErrada=%v", errInexistente, errSenhaErrada)
	}
	if errInexistente.Error() != errSenhaErrada.Error() {
		t.Fatalf("mensagens deveriam ser idênticas (RN04); inexistente=%q senhaErrada=%q", errInexistente.Error(), errSenhaErrada.Error())
	}
}

// TestAutenticarSenha_RegistraEventoDeLogin cobre a telemetria adicionada ao
// login por senha (spec 004, RF04): o evento é gravado com o usuário/tenant
// autenticados.
func TestAutenticarSenha_RegistraEventoDeLogin(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("12345678"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	fs := &fakeStore{obterSenhaUserID: "user-1", obterSenhaTenantID: "tenant-1", obterSenhaTipo: "dono", obterSenhaHash: string(hash)}
	srv, _ := newServer(t, fs)

	if _, err := srv.AutenticarSenha(context.Background(), &iamv1.AutenticarSenhaRequest{Email: "ana@example.com", Senha: "12345678"}); err != nil {
		t.Fatalf("erro: %v", err)
	}
	if fs.eventoUsuarioID != "user-1" || fs.eventoTenantID != "tenant-1" {
		t.Fatalf("evento de login não registrado com a identidade correta: usuario=%q tenant=%q", fs.eventoUsuarioID, fs.eventoTenantID)
	}
}

// TestAutenticarSenha_FalhaAoRegistrarEventoNaoImpedeLogin: a telemetria é
// auxiliar — login é o fluxo crítico e nunca deve falhar por causa dela.
func TestAutenticarSenha_FalhaAoRegistrarEventoNaoImpedeLogin(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("12345678"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	fs := &fakeStore{
		obterSenhaUserID: "user-1", obterSenhaTenantID: "tenant-1", obterSenhaTipo: "dono", obterSenhaHash: string(hash),
		registrarEventoErr: errors.New("db indisponível"),
	}
	srv, _ := newServer(t, fs)

	resp, err := srv.AutenticarSenha(context.Background(), &iamv1.AutenticarSenhaRequest{Email: "ana@example.com", Senha: "12345678"})
	if err != nil {
		t.Fatalf("erro ao registrar evento de login não deveria impedir o login: %v", err)
	}
	if resp.GetUserId() != "user-1" {
		t.Fatalf("resposta inesperada: %+v", resp)
	}
}

// TestAutenticarThirdParty_RegistraEventoDeLogin cobre a mesma telemetria no
// login social.
func TestAutenticarThirdParty_RegistraEventoDeLogin(t *testing.T) {
	fs := &fakeStore{userID: "user-9", tenantID: "tenant-padrao", tipo: "cliente"}
	srv, _ := newServer(t, fs)

	if _, err := srv.AutenticarThirdParty(context.Background(), &iamv1.AutenticarThirdPartyRequest{
		Provedor: "google", ExternalId: "g-123", Email: "a@b.com", Nome: "Ana",
	}); err != nil {
		t.Fatalf("erro: %v", err)
	}
	if fs.eventoUsuarioID != "user-9" || fs.eventoTenantID != "tenant-padrao" {
		t.Fatalf("evento de login não registrado com a identidade correta: usuario=%q tenant=%q", fs.eventoUsuarioID, fs.eventoTenantID)
	}
}

func TestListarUltimosAcessos_SemTenantContext(t *testing.T) {
	srv, _ := newServer(t, &fakeStore{})
	if _, err := srv.ListarUltimosAcessos(context.Background(), &iamv1.ListarUltimosAcessosRequest{}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("esperava Unauthenticated; got=%v", err)
	}
}

func TestListarUltimosAcessos_DevolveEventosFormatados(t *testing.T) {
	quando := time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC)
	fs := &fakeStore{ultimosAcessos: []store.EventoLogin{
		{UsuarioNome: "Ana", TenantNome: "Acme", CriadoEm: quando},
	}}
	srv, _ := newServer(t, fs)

	resp, err := srv.ListarUltimosAcessos(tenantACtx(), &iamv1.ListarUltimosAcessosRequest{})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if len(resp.GetEventos()) != 1 || resp.GetEventos()[0].GetUsuarioNome() != "Ana" {
		t.Fatalf("eventos inesperados: %+v", resp.GetEventos())
	}
	if resp.GetEventos()[0].GetCriadoEm() != quando.Format(time.RFC3339) {
		t.Fatalf("criado_em não formatado como RFC3339: %q", resp.GetEventos()[0].GetCriadoEm())
	}
	if fs.ultimosAcessosTenant != "tenant-A" || fs.ultimosAcessosLimite != 10 {
		t.Fatalf("store chamado com args inesperados: tenant=%q limite=%d", fs.ultimosAcessosTenant, fs.ultimosAcessosLimite)
	}
}

func TestListarFeedback_SemTenantContext(t *testing.T) {
	srv, _ := newServer(t, &fakeStore{})
	if _, err := srv.ListarFeedback(context.Background(), &iamv1.ListarFeedbackRequest{}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("esperava Unauthenticated; got=%v", err)
	}
}

func TestListarFeedback_StatusInvalido(t *testing.T) {
	srv, _ := newServer(t, &fakeStore{})
	if _, err := srv.ListarFeedback(tenantACtx(), &iamv1.ListarFeedbackRequest{Status: "arquivado"}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("esperava InvalidArgument; got=%v", err)
	}
}

func TestListarFeedback_SemFiltro_PassaNilAoStore(t *testing.T) {
	fs := &fakeStore{}
	srv, _ := newServer(t, fs)
	if _, err := srv.ListarFeedback(tenantACtx(), &iamv1.ListarFeedbackRequest{}); err != nil {
		t.Fatalf("erro: %v", err)
	}
	if fs.listarFeedbackStatus != nil {
		t.Fatalf("status vazio deveria virar nil (sem filtro); got=%v", *fs.listarFeedbackStatus)
	}
	if fs.listarFeedbackTenant != "tenant-A" {
		t.Fatalf("tenant inesperado: %q", fs.listarFeedbackTenant)
	}
}

func TestListarFeedback_ComFiltro_PassaPonteiroAoStore(t *testing.T) {
	fs := &fakeStore{}
	srv, _ := newServer(t, fs)
	if _, err := srv.ListarFeedback(tenantACtx(), &iamv1.ListarFeedbackRequest{Status: "pendente"}); err != nil {
		t.Fatalf("erro: %v", err)
	}
	if fs.listarFeedbackStatus == nil || *fs.listarFeedbackStatus != "pendente" {
		t.Fatalf("filtro de status não propagado: %v", fs.listarFeedbackStatus)
	}
}

func TestListarFeedback_DevolveItensFormatados(t *testing.T) {
	quando := time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC)
	fs := &fakeStore{listarFeedback: []store.Feedback{
		{ID: "f1", TenantNome: "Acme", Mensagem: "ajuda", Status: "pendente", CriadoEm: quando},
	}}
	srv, _ := newServer(t, fs)

	resp, err := srv.ListarFeedback(tenantACtx(), &iamv1.ListarFeedbackRequest{})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if len(resp.GetItens()) != 1 || resp.GetItens()[0].GetId() != "f1" || resp.GetItens()[0].GetStatus() != "pendente" {
		t.Fatalf("itens inesperados: %+v", resp.GetItens())
	}
}

// TestAtualizarStatusFeedback_RejeitaTransicaoParaPendente cobre RN03: a
// transição respondido→pendente é proibida — rejeitada antes mesmo de tocar
// o store.
func TestAtualizarStatusFeedback_RejeitaTransicaoParaPendente(t *testing.T) {
	fs := &fakeStore{}
	srv, _ := newServer(t, fs)
	if _, err := srv.AtualizarStatusFeedback(tenantACtx(), &iamv1.AtualizarStatusFeedbackRequest{Id: "f1", Status: "pendente"}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("esperava InvalidArgument; got=%v", err)
	}
	if fs.feedbackIDAtualizado != "" {
		t.Fatal("store.AtualizarStatusFeedback não deveria ser chamado para status='pendente'")
	}
}

func TestAtualizarStatusFeedback_RejeitaStatusDesconhecido(t *testing.T) {
	srv, _ := newServer(t, &fakeStore{})
	if _, err := srv.AtualizarStatusFeedback(tenantACtx(), &iamv1.AtualizarStatusFeedbackRequest{Id: "f1", Status: "arquivado"}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("esperava InvalidArgument; got=%v", err)
	}
}

func TestAtualizarStatusFeedback_Sucesso(t *testing.T) {
	fs := &fakeStore{atualizarFeedback: store.Feedback{ID: "f1", TenantNome: "Acme", Mensagem: "ajuda", Status: "respondido"}}
	srv, _ := newServer(t, fs)

	resp, err := srv.AtualizarStatusFeedback(tenantACtx(), &iamv1.AtualizarStatusFeedbackRequest{Id: "f1", Status: "respondido"})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.GetId() != "f1" || resp.GetStatus() != "respondido" {
		t.Fatalf("resposta inesperada: %+v", resp)
	}
	if fs.feedbackIDAtualizado != "f1" || fs.feedbackStatusAtualizado != "respondido" {
		t.Fatalf("store chamado com args inesperados: id=%q status=%q", fs.feedbackIDAtualizado, fs.feedbackStatusAtualizado)
	}
}

func TestAtualizarStatusFeedback_NaoEncontrado(t *testing.T) {
	fs := &fakeStore{atualizarFeedbackErr: store.ErrNaoEncontrado}
	srv, _ := newServer(t, fs)
	if _, err := srv.AtualizarStatusFeedback(tenantACtx(), &iamv1.AtualizarStatusFeedbackRequest{Id: "f1", Status: "respondido"}); status.Code(err) != codes.NotFound {
		t.Fatalf("esperava NotFound; got=%v", err)
	}
}

func TestResumoFinanceiro_SemTenantContext(t *testing.T) {
	srv, _ := newServer(t, &fakeStore{})
	if _, err := srv.ResumoFinanceiro(context.Background(), &iamv1.ResumoFinanceiroRequest{}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("esperava Unauthenticated; got=%v", err)
	}
}

func TestResumoFinanceiro_DevolveValoresDoStore(t *testing.T) {
	fs := &fakeStore{resumoFinanceiro: store.ResumoFinanceiro{ReceitaTotalCentavos: 12345, Moeda: "BRL", Competencia: "2026-08"}}
	srv, _ := newServer(t, fs)

	resp, err := srv.ResumoFinanceiro(tenantACtx(), &iamv1.ResumoFinanceiroRequest{})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.GetReceitaTotalCentavos() != 12345 || resp.GetMoeda() != "BRL" || resp.GetCompetencia() != "2026-08" {
		t.Fatalf("resposta inesperada: %+v", resp)
	}
	if fs.resumoFinanceiroTenant != "tenant-A" {
		t.Fatalf("tenant inesperado: %q", fs.resumoFinanceiroTenant)
	}
}

// ─────────────────────────────────────────────────────────────────────────
// Área "Conta/Configuração" (spec 004, RF14-RF18).
// ─────────────────────────────────────────────────────────────────────────

func TestAtualizarPerfil_SemTenantContext(t *testing.T) {
	srv, _ := newServer(t, &fakeStore{})
	if _, err := srv.AtualizarPerfil(context.Background(), &iamv1.AtualizarPerfilRequest{Nome: "Ana"}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("esperava Unauthenticated; got=%v", err)
	}
}

func TestAtualizarPerfil_NomeVazio(t *testing.T) {
	srv, _ := newServer(t, &fakeStore{})
	ctx := ctxComUsuario("tenant-A", "user-1", "dono")
	if _, err := srv.AtualizarPerfil(ctx, &iamv1.AtualizarPerfilRequest{}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("nome vazio deveria falhar; got=%v", err)
	}
}

func TestAtualizarPerfil_Sucesso(t *testing.T) {
	fs := &fakeStore{}
	srv, _ := newServer(t, fs)
	ctx := ctxComUsuario("tenant-A", "user-1", "dono")
	if _, err := srv.AtualizarPerfil(ctx, &iamv1.AtualizarPerfilRequest{Nome: "Ana", FotoUrl: "https://x/y.png"}); err != nil {
		t.Fatalf("erro: %v", err)
	}
	if fs.perfilUserIDChamado != "user-1" || fs.perfilNome != "Ana" || fs.perfilFotoURL != "https://x/y.png" {
		t.Fatalf("store chamado com args inesperados: %+v", fs)
	}
}

func TestAtualizarSenha_SenhaCurta(t *testing.T) {
	srv, _ := newServer(t, &fakeStore{})
	ctx := ctxComUsuario("tenant-A", "user-1", "dono")
	_, err := srv.AtualizarSenha(ctx, &iamv1.AtualizarSenhaRequest{SenhaAtual: "atual123", SenhaNova: "curta"})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("senha nova curta deveria falhar; got=%v", err)
	}
}

func TestAtualizarSenha_SenhaAtualErrada(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("senha-correta"), bcrypt.DefaultCost)
	fs := &fakeStore{hashSenha: string(hash)}
	srv, _ := newServer(t, fs)
	ctx := ctxComUsuario("tenant-A", "user-1", "dono")

	_, err := srv.AtualizarSenha(ctx, &iamv1.AtualizarSenhaRequest{SenhaAtual: "senha-errada", SenhaNova: "12345678"})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("senha atual errada deveria dar Unauthenticated; got=%v", err)
	}
	if fs.senhaAtualizadaHash != "" {
		t.Fatal("senha não deveria ser trocada quando senha_atual está errada")
	}
}

func TestAtualizarSenha_Sucesso(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("senha-correta"), bcrypt.DefaultCost)
	fs := &fakeStore{hashSenha: string(hash)}
	srv, _ := newServer(t, fs)
	ctx := ctxComUsuario("tenant-A", "user-1", "dono")

	if _, err := srv.AtualizarSenha(ctx, &iamv1.AtualizarSenhaRequest{SenhaAtual: "senha-correta", SenhaNova: "nova-senha-12345"}); err != nil {
		t.Fatalf("erro: %v", err)
	}
	if fs.senhaAtualizadaUserID != "user-1" {
		t.Fatalf("userID inesperado: %q", fs.senhaAtualizadaUserID)
	}
	if bcrypt.CompareHashAndPassword([]byte(fs.senhaAtualizadaHash), []byte("nova-senha-12345")) != nil {
		t.Fatal("hash persistido não corresponde à nova senha")
	}
}

func TestAtivarMfa_SemTenantContext(t *testing.T) {
	srv, _ := newServer(t, &fakeStore{})
	if _, err := srv.AtivarMfa(context.Background(), &iamv1.AtivarMfaRequest{}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("esperava Unauthenticated; got=%v", err)
	}
}

func TestAtivarMfa_Sucesso(t *testing.T) {
	fs := &fakeStore{emailUsuario: "ana@example.com"}
	srv, _ := newServerComMfa(t, fs)
	ctx := ctxComUsuario("tenant-A", "user-1", "dono")

	resp, err := srv.AtivarMfa(ctx, &iamv1.AtivarMfaRequest{})
	if err != nil {
		t.Fatalf("erro: %v", err)
	}
	if resp.GetSegredoOtpAuthUri() == "" || !strings.HasPrefix(resp.GetSegredoOtpAuthUri(), "otpauth://") {
		t.Fatalf("uri otpauth inesperada: %q", resp.GetSegredoOtpAuthUri())
	}
	if fs.iniciarMfaUserID != "user-1" || len(fs.iniciarMfaSegredo) == 0 {
		t.Fatalf("store não recebeu segredo cifrado: %+v", fs)
	}
	// O segredo persistido nunca é o texto claro — deve decifrar de volta.
	if _, err := auth.DecifrarSegredo(chaveMfaTeste, fs.iniciarMfaSegredo); err != nil {
		t.Fatalf("segredo cifrado não decifra com a mesma chave: %v", err)
	}
}

func TestAtivarMfa_JaAtivo(t *testing.T) {
	fs := &fakeStore{emailUsuario: "ana@example.com", iniciarMfaErr: store.ErrMfaJaAtivo}
	srv, _ := newServerComMfa(t, fs)
	ctx := ctxComUsuario("tenant-A", "user-1", "dono")

	if _, err := srv.AtivarMfa(ctx, &iamv1.AtivarMfaRequest{}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("mfa já ativo deveria dar FailedPrecondition; got=%v", err)
	}
}

// gerarSegredoCifrado cria um segredo TOTP válido e devolve o texto claro e a
// versão cifrada com chaveMfaTeste — usado pelos testes de ConfirmarMfa.
func gerarSegredoCifrado(t *testing.T) (claro string, cifrado []byte) {
	t.Helper()
	chave, err := totp.Generate(totp.GenerateOpts{Issuer: "MACH V4", AccountName: "ana@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	cifrado, err = auth.CifrarSegredo(chaveMfaTeste, chave.Secret())
	if err != nil {
		t.Fatal(err)
	}
	return chave.Secret(), cifrado
}

func TestConfirmarMfa_SemAtivacaoPendente(t *testing.T) {
	fs := &fakeStore{}
	srv, _ := newServerComMfa(t, fs)
	ctx := ctxComUsuario("tenant-A", "user-1", "dono")

	if _, err := srv.ConfirmarMfa(ctx, &iamv1.ConfirmarMfaRequest{Codigo: "123456"}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("sem ativação pendente deveria dar FailedPrecondition; got=%v", err)
	}
}

func TestConfirmarMfa_CodigoInvalido(t *testing.T) {
	_, cifrado := gerarSegredoCifrado(t)
	fs := &fakeStore{segredoMfaPendente: cifrado}
	srv, _ := newServerComMfa(t, fs)
	ctx := ctxComUsuario("tenant-A", "user-1", "dono")

	if _, err := srv.ConfirmarMfa(ctx, &iamv1.ConfirmarMfaRequest{Codigo: "000000"}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("código inválido deveria dar Unauthenticated; got=%v", err)
	}
	if fs.confirmarMfaCodigo != "" {
		t.Fatal("mfa não deveria ser confirmado com código inválido")
	}
}

func TestConfirmarMfa_Sucesso(t *testing.T) {
	claro, cifrado := gerarSegredoCifrado(t)
	codigo, err := totp.GenerateCode(claro, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	fs := &fakeStore{segredoMfaPendente: cifrado}
	srv, _ := newServerComMfa(t, fs)
	ctx := ctxComUsuario("tenant-A", "user-1", "dono")

	if _, err := srv.ConfirmarMfa(ctx, &iamv1.ConfirmarMfaRequest{Codigo: codigo}); err != nil {
		t.Fatalf("erro: %v", err)
	}
	if fs.confirmarMfaUserID != "user-1" || fs.confirmarMfaCodigo != codigo {
		t.Fatalf("store não confirmado com os args esperados: %+v", fs)
	}
}

// Anti-replay (RF15): o mesmo código TOTP não pode confirmar o MFA duas vezes
// seguidas, mesmo sendo criptograficamente válido dentro da janela de tempo.
func TestConfirmarMfa_AntiReplay(t *testing.T) {
	claro, cifrado := gerarSegredoCifrado(t)
	codigo, err := totp.GenerateCode(claro, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	agora := time.Now()
	fs := &fakeStore{segredoMfaPendente: cifrado, ultimoCodigoMfa: codigo, ultimoCodigoMfaEm: &agora}
	srv, _ := newServerComMfa(t, fs)
	ctx := ctxComUsuario("tenant-A", "user-1", "dono")

	if _, err := srv.ConfirmarMfa(ctx, &iamv1.ConfirmarMfaRequest{Codigo: codigo}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("código repetido deveria ser rejeitado; got=%v", err)
	}
	if fs.confirmarMfaCodigo != "" {
		t.Fatal("mfa não deveria ser confirmado com código repetido")
	}
}

func TestDesativarMfa_SenhaErrada(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("senha-correta"), bcrypt.DefaultCost)
	fs := &fakeStore{hashSenha: string(hash)}
	srv, _ := newServer(t, fs)
	ctx := ctxComUsuario("tenant-A", "user-1", "dono")

	if _, err := srv.DesativarMfa(ctx, &iamv1.DesativarMfaRequest{SenhaAtual: "errada"}); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("senha errada deveria dar Unauthenticated; got=%v", err)
	}
	if fs.desativarMfaUserID != "" {
		t.Fatal("mfa não deveria ser desativado com senha errada")
	}
}

func TestDesativarMfa_Sucesso(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("senha-correta"), bcrypt.DefaultCost)
	fs := &fakeStore{hashSenha: string(hash)}
	srv, _ := newServer(t, fs)
	ctx := ctxComUsuario("tenant-A", "user-1", "dono")

	if _, err := srv.DesativarMfa(ctx, &iamv1.DesativarMfaRequest{SenhaAtual: "senha-correta"}); err != nil {
		t.Fatalf("erro: %v", err)
	}
	if fs.desativarMfaUserID != "user-1" {
		t.Fatalf("userID inesperado: %q", fs.desativarMfaUserID)
	}
}

func TestExcluirConta_TenantVinculado(t *testing.T) {
	fs := &fakeStore{filhos: []store.Tenant{{ID: "t1", Nome: "Acme"}}}
	srv, _ := newServer(t, fs)
	ctx := ctxComUsuario("tenant-A", "user-1", "dono")

	_, err := srv.ExcluirConta(ctx, &iamv1.ExcluirContaRequest{SenhaAtual: "qualquer"})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("tenant vinculado deveria dar FailedPrecondition; got=%v", err)
	}
	if fs.excluirContaUserID != "" {
		t.Fatal("conta não deveria ser excluída com tenant vinculado")
	}
}

func TestExcluirConta_SenhaErrada(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("senha-correta"), bcrypt.DefaultCost)
	fs := &fakeStore{hashSenha: string(hash)}
	srv, _ := newServer(t, fs)
	ctx := ctxComUsuario("tenant-A", "user-1", "dono")

	_, err := srv.ExcluirConta(ctx, &iamv1.ExcluirContaRequest{SenhaAtual: "errada"})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("senha errada deveria dar Unauthenticated; got=%v", err)
	}
	if fs.excluirContaUserID != "" {
		t.Fatal("conta não deveria ser excluída com senha errada")
	}
}

func TestExcluirConta_Sucesso(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("senha-correta"), bcrypt.DefaultCost)
	fs := &fakeStore{hashSenha: string(hash)}
	srv, _ := newServer(t, fs)
	ctx := ctxComUsuario("tenant-A", "user-1", "dono")

	if _, err := srv.ExcluirConta(ctx, &iamv1.ExcluirContaRequest{SenhaAtual: "senha-correta"}); err != nil {
		t.Fatalf("erro: %v", err)
	}
	if fs.excluirContaUserID != "user-1" || fs.excluirContaTenantID != "tenant-A" {
		t.Fatalf("store chamado com args inesperados: %+v", fs)
	}
}

func TestSolicitarTrocaEmail_EmailInvalido(t *testing.T) {
	srv, _ := newServer(t, &fakeStore{})
	ctx := ctxComUsuario("tenant-A", "user-1", "dono")
	if _, err := srv.SolicitarTrocaEmail(ctx, &iamv1.SolicitarTrocaEmailRequest{NovoEmail: "sem-arroba"}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("e-mail inválido deveria falhar; got=%v", err)
	}
}

func TestSolicitarTrocaEmail_Sucesso(t *testing.T) {
	fs := &fakeStore{}
	srv, _ := newServer(t, fs)
	ctx := ctxComUsuario("tenant-A", "user-1", "dono")

	if _, err := srv.SolicitarTrocaEmail(ctx, &iamv1.SolicitarTrocaEmailRequest{NovoEmail: "novo@example.com"}); err != nil {
		t.Fatalf("erro: %v", err)
	}
	if fs.solicitarTrocaUserID != "user-1" || fs.solicitarTrocaNovoEmail != "novo@example.com" {
		t.Fatalf("store chamado com args inesperados: %+v", fs)
	}
	if len(fs.solicitarTrocaTokenHash) != 64 { // sha256 em hex
		t.Fatalf("tokenHash deveria ter 64 chars hex; got=%d", len(fs.solicitarTrocaTokenHash))
	}
	if fs.solicitarTrocaExpira.Before(time.Now().Add(50 * time.Minute)) {
		t.Fatalf("expiração deveria ser ~1h no futuro; got=%v", fs.solicitarTrocaExpira)
	}
}

func TestSolicitarTrocaEmail_EmailDuplicado(t *testing.T) {
	fs := &fakeStore{solicitarTrocaErr: store.ErrEmailJaCadastrado}
	srv, _ := newServer(t, fs)
	ctx := ctxComUsuario("tenant-A", "user-1", "dono")

	_, err := srv.SolicitarTrocaEmail(ctx, &iamv1.SolicitarTrocaEmailRequest{NovoEmail: "ja-existe@example.com"})
	if status.Code(err) != codes.AlreadyExists {
		t.Fatalf("e-mail duplicado deveria dar AlreadyExists; got=%v", err)
	}
}

func TestConfirmarTrocaEmail_TokenVazio(t *testing.T) {
	srv, _ := newServer(t, &fakeStore{})
	if _, err := srv.ConfirmarTrocaEmail(context.Background(), &iamv1.ConfirmarTrocaEmailRequest{}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("token vazio deveria falhar; got=%v", err)
	}
}

func TestConfirmarTrocaEmail_TokenInvalidoOuExpirado(t *testing.T) {
	fs := &fakeStore{confirmarTrocaErr: store.ErrNaoEncontrado}
	srv, _ := newServer(t, fs)
	if _, err := srv.ConfirmarTrocaEmail(context.Background(), &iamv1.ConfirmarTrocaEmailRequest{Token: "lixo"}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("token inválido/expirado deveria falhar; got=%v", err)
	}
}

func TestConfirmarTrocaEmail_Sucesso(t *testing.T) {
	fs := &fakeStore{}
	srv, _ := newServer(t, fs)
	if _, err := srv.ConfirmarTrocaEmail(context.Background(), &iamv1.ConfirmarTrocaEmailRequest{Token: "token-valido"}); err != nil {
		t.Fatalf("erro: %v", err)
	}
	if len(fs.confirmarTrocaTokenHash) != 64 {
		t.Fatalf("tokenHash deveria ter 64 chars hex; got=%d", len(fs.confirmarTrocaTokenHash))
	}
}
