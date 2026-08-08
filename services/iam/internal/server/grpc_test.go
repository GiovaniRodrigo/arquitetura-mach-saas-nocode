package server

import (
	"context"
	"testing"
	"time"

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
	registrarUserID, registrarTenantID                        string
	registrarErr                                              error
	nomeRegistrado, emailRegistrado, senhaHashRegistrada       string
	nomeTenantRegistrado                                       string
	// login por senha (ObterUsuarioPorEmailSenha)
	obterSenhaUserID, obterSenhaTenantID, obterSenhaTipo, obterSenhaHash string
	obterSenhaErr                                                       error
	emailConsultado                                                     string
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

func newServer(t *testing.T, store Store) (*IAMServer, *auth.Issuer) {
	t.Helper()
	priv, err := auth.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	iss := auth.NewIssuer(priv, time.Hour)
	return New(auth.NewValidator(&priv.PublicKey), iss, store), iss
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
		{Email: "a@b.com", Senha: "12345678", NomeTenant: "T"},           // sem nome
		{Nome: "Ana", Senha: "12345678", NomeTenant: "T"},                // sem email
		{Nome: "Ana", Email: "a@b.com", NomeTenant: "T"},                 // sem senha
		{Nome: "Ana", Email: "a@b.com", Senha: "12345678"},               // sem nome_tenant
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
