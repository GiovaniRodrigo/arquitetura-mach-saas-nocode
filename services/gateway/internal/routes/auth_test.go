package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	iamv1 "github.com/machv4/platform/gen/go/construtor/iam/v1"
)

// fakeIamAuth implementa AuthCliente.
type fakeIamAuth struct {
	registrarResp *iamv1.RegistrarUsuarioResponse
	registrarErr  error
	registrarReq  *iamv1.RegistrarUsuarioRequest

	loginResp *iamv1.AutenticarSenhaResponse
	loginErr  error
	loginReq  *iamv1.AutenticarSenhaRequest
}

func (f *fakeIamAuth) RegistrarUsuario(_ context.Context, in *iamv1.RegistrarUsuarioRequest, _ ...grpc.CallOption) (*iamv1.RegistrarUsuarioResponse, error) {
	f.registrarReq = in
	return f.registrarResp, f.registrarErr
}

func (f *fakeIamAuth) AutenticarSenha(_ context.Context, in *iamv1.AutenticarSenhaRequest, _ ...grpc.CallOption) (*iamv1.AutenticarSenhaResponse, error) {
	f.loginReq = in
	return f.loginResp, f.loginErr
}

func TestRegistrarUsuario_OK(t *testing.T) {
	fi := &fakeIamAuth{registrarResp: &iamv1.RegistrarUsuarioResponse{Jwt: "tok", UserId: "u1", TenantId: "t1", Tipo: "dono"}}
	rec := httptest.NewRecorder()
	corpo := `{"nome":"Ana","email":"ana@example.com","senha":"12345678","nome_tenant":"Ana LTDA"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/registro", strings.NewReader(corpo))

	RegistrarUsuario(fi)(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status: %d, body: %s", rec.Code, rec.Body.String())
	}
	var corpoResp struct {
		Jwt      string `json:"jwt"`
		UserID   string `json:"user_id"`
		TenantID string `json:"tenant_id"`
		Tipo     string `json:"tipo"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&corpoResp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if corpoResp.Jwt != "tok" || corpoResp.UserID != "u1" || corpoResp.TenantID != "t1" || corpoResp.Tipo != "dono" {
		t.Fatalf("corpo inesperado: %+v", corpoResp)
	}
	if fi.registrarReq.GetNome() != "Ana" || fi.registrarReq.GetEmail() != "ana@example.com" ||
		fi.registrarReq.GetSenha() != "12345678" || fi.registrarReq.GetNomeTenant() != "Ana LTDA" {
		t.Fatalf("request ao IAM inesperado: %+v", fi.registrarReq)
	}
}

func TestRegistrarUsuario_CorpoInvalido(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/registro", strings.NewReader("{lixo"))
	RegistrarUsuario(&fakeIamAuth{})(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: %d", rec.Code)
	}
}

func TestRegistrarUsuario_EmailDuplicado(t *testing.T) {
	fi := &fakeIamAuth{registrarErr: status.Error(codes.AlreadyExists, "e-mail já cadastrado")}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/registro", strings.NewReader(`{"nome":"Ana","email":"a@b.com","senha":"12345678","nome_tenant":"T"}`))

	RegistrarUsuario(fi)(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status: %d, body: %s", rec.Code, rec.Body.String())
	}
	var corpo struct {
		Codigo string `json:"codigo"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&corpo)
	if corpo.Codigo != "EMAIL_DUPLICADO" {
		t.Fatalf("código de erro inesperado: %q", corpo.Codigo)
	}
}

func TestRegistrarUsuario_CamposObrigatorios_ViraValidationError(t *testing.T) {
	fi := &fakeIamAuth{registrarErr: status.Error(codes.InvalidArgument, "nome, email, senha e nome_tenant são obrigatórios")}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/registro", strings.NewReader(`{}`))

	RegistrarUsuario(fi)(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status: %d, body: %s", rec.Code, rec.Body.String())
	}
}

func TestLogin_OK(t *testing.T) {
	fi := &fakeIamAuth{loginResp: &iamv1.AutenticarSenhaResponse{Jwt: "tok", UserId: "u1", TenantId: "t1", Tipo: "dono"}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"ana@example.com","senha":"12345678"}`))

	Login(fi)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d, body: %s", rec.Code, rec.Body.String())
	}
	if fi.loginReq.GetEmail() != "ana@example.com" || fi.loginReq.GetSenha() != "12345678" {
		t.Fatalf("request ao IAM inesperado: %+v", fi.loginReq)
	}
}

// RN04 / Critério de Aceitação 4: e-mail inexistente e senha incorreta devem
// produzir a mesma resposta HTTP (status + corpo), nunca distinguíveis.
func TestLogin_EmailInexistenteESenhaIncorreta_MesmaResposta(t *testing.T) {
	erroGenerico := status.Error(codes.Unauthenticated, "credenciais inválidas")

	fiInexistente := &fakeIamAuth{loginErr: erroGenerico}
	recInexistente := httptest.NewRecorder()
	Login(fiInexistente)(recInexistente, httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"fantasma@example.com","senha":"qualquer"}`)))

	fiSenhaErrada := &fakeIamAuth{loginErr: erroGenerico}
	recSenhaErrada := httptest.NewRecorder()
	Login(fiSenhaErrada)(recSenhaErrada, httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"ana@example.com","senha":"errada"}`)))

	if recInexistente.Code != http.StatusUnauthorized || recSenhaErrada.Code != http.StatusUnauthorized {
		t.Fatalf("ambos deveriam ser 401; inexistente=%d senhaErrada=%d", recInexistente.Code, recSenhaErrada.Code)
	}
	if recInexistente.Body.String() != recSenhaErrada.Body.String() {
		t.Fatalf("corpos deveriam ser idênticos (RN04); inexistente=%q senhaErrada=%q", recInexistente.Body.String(), recSenhaErrada.Body.String())
	}
}

func TestLogin_CorpoInvalido(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader("{lixo"))
	Login(&fakeIamAuth{})(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: %d", rec.Code)
	}
}
