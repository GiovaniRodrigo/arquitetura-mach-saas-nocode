// Package routes — auto cadastro e login por e-mail/senha (spec 006).
//
// Rotas públicas (fora do grupo autenticado), registradas sob /api/v1/auth/*
// para reaproveitar o mesmo proxy que /api/* já tem no Vite dev server e no
// Nginx de produção (decisão técnica 4.2 do plan.md) — diferente de /auth/*
// (OAuth), que é navegação de browser, não fetch do SPA.
package routes

import (
	"context"
	"encoding/json"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	iamv1 "github.com/machv4/platform/gen/go/construtor/iam/v1"
	"github.com/machv4/platform/services/gateway/internal/web"
)

// AuthCliente é o subconjunto do IAMServiceClient usado por estas rotas.
type AuthCliente interface {
	RegistrarUsuario(ctx context.Context, in *iamv1.RegistrarUsuarioRequest, opts ...grpc.CallOption) (*iamv1.RegistrarUsuarioResponse, error)
	AutenticarSenha(ctx context.Context, in *iamv1.AutenticarSenhaRequest, opts ...grpc.CallOption) (*iamv1.AutenticarSenhaResponse, error)
}

// authResp espelha a identidade autenticada no corpo JSON — mesmo formato
// devolvido pelo cadastro e pelo login por senha (contracts/api.md).
type authResp struct {
	Jwt      string `json:"jwt"`
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Tipo     string `json:"tipo"`
}

// registrarReq é o corpo de POST /api/v1/auth/registro.
type registrarReq struct {
	Nome       string `json:"nome"`
	Email      string `json:"email"`
	Senha      string `json:"senha"`
	NomeTenant string `json:"nome_tenant"`
}

// loginReq é o corpo de POST /api/v1/auth/login.
type loginReq struct {
	Email string `json:"email"`
	Senha string `json:"senha"`
}

// RegistrarUsuario serve POST /api/v1/auth/registro (spec 006, RF04): cria um
// tenant próprio (dono) e a conta de senha do usuário, devolvendo-o já
// autenticado. Validação de campos (nome/e-mail/senha≥8/nome_tenant) acontece
// no IAM (RegistrarUsuario) — aqui só se traduz o corpo e o erro.
func RegistrarUsuario(iam AuthCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body registrarReq
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			web.Error(w, http.StatusBadRequest, "INVALID_BODY", "corpo JSON inválido")
			return
		}
		out, err := iam.RegistrarUsuario(r.Context(), &iamv1.RegistrarUsuarioRequest{
			Nome:       body.Nome,
			Email:      body.Email,
			Senha:      body.Senha,
			NomeTenant: body.NomeTenant,
		})
		if err != nil {
			writeAuthError(w, err)
			return
		}
		web.JSON(w, http.StatusCreated, authResp{Jwt: out.GetJwt(), UserID: out.GetUserId(), TenantID: out.GetTenantId(), Tipo: out.GetTipo()})
	}
}

// Login serve POST /api/v1/auth/login (spec 006, RF06): autentica uma conta de
// senha e devolve o mesmo formato de identidade do cadastro/OAuth. E-mail
// inexistente e senha incorreta produzem a mesma resposta (RN04) — o IAM já
// garante isso devolvendo sempre o mesmo erro Unauthenticated.
func Login(iam AuthCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body loginReq
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			web.Error(w, http.StatusBadRequest, "INVALID_BODY", "corpo JSON inválido")
			return
		}
		out, err := iam.AutenticarSenha(r.Context(), &iamv1.AutenticarSenhaRequest{Email: body.Email, Senha: body.Senha})
		if err != nil {
			writeAuthError(w, err)
			return
		}
		web.JSON(w, http.StatusOK, authResp{Jwt: out.GetJwt(), UserID: out.GetUserId(), TenantID: out.GetTenantId(), Tipo: out.GetTipo()})
	}
}

// writeAuthError traduz os códigos gRPC do IAM para HTTP nas rotas de
// cadastro/login: InvalidArgument (campo obrigatório/senha curta) → 422;
// AlreadyExists (e-mail duplicado, RN02) → 409; Unauthenticated (credenciais
// inválidas, RN04) → 401. A mensagem nunca inclui a senha (RNF05).
func writeAuthError(w http.ResponseWriter, err error) {
	switch status.Code(err) {
	case codes.InvalidArgument:
		web.Error(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "dados de cadastro/login inválidos")
	case codes.AlreadyExists:
		web.Error(w, http.StatusConflict, "EMAIL_DUPLICADO", "e-mail já cadastrado")
	case codes.Unauthenticated:
		web.Error(w, http.StatusUnauthorized, "CREDENCIAIS_INVALIDAS", "credenciais inválidas")
	default:
		web.Error(w, http.StatusInternalServerError, "INTERNAL", "erro interno")
	}
}
