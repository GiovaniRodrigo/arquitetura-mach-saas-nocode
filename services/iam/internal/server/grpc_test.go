package server

import (
	"context"
	"testing"
	"time"

	iamv1 "github.com/machv4/platform/gen/go/construtor/iam/v1"
	commonv1 "github.com/machv4/platform/gen/go/construtor/common/v1"
	"github.com/machv4/platform/pkg/tenantctx"
	"github.com/machv4/platform/services/iam/auth"
	"github.com/machv4/platform/services/iam/internal/permissions"
)

type fakeStore struct {
	perms []permissions.Permissao
	err   error
	// resposta do upsert de identidade
	userID, tenantID, tipo string
	upsertErr              error
}

func (f fakeStore) PermissoesDe(context.Context, []string) ([]permissions.Permissao, error) {
	return f.perms, f.err
}

func (f fakeStore) UpsertUsuarioThirdParty(context.Context, string, string, string, string) (string, string, string, error) {
	return f.userID, f.tenantID, f.tipo, f.upsertErr
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
	srv, iss := newServer(t, fakeStore{})
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
	store := fakeStore{userID: "user-9", tenantID: "tenant-padrao", tipo: "cliente"}
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
	srv, _ := newServer(t, fakeStore{})
	if _, err := srv.AutenticarThirdParty(context.Background(), &iamv1.AutenticarThirdPartyRequest{Provedor: "google"}); err == nil {
		t.Fatal("external_id ausente deveria falhar")
	}
}

func TestAvaliarPermissoes_SemTenantContext(t *testing.T) {
	srv, _ := newServer(t, fakeStore{})
	if _, err := srv.AvaliarPermissoes(context.Background(), &iamv1.AvaliarPermissoesRequest{}); err == nil {
		t.Fatal("sem TenantContext deveria retornar erro Unauthenticated")
	}
}

func TestAvaliarPermissoes_MapaBooleano(t *testing.T) {
	loader := fakeStore{perms: []permissions.Permissao{
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
