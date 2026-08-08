// Package routes — área "Conta/Configuração" (spec 004-reestruturacao-ia-
// navegacao, RF14-RF18): perfil, senha, MFA TOTP em duas etapas, exclusão de
// conta (LGPD, anonimização) e troca de e-mail em duas etapas. Todas as rotas
// vivem no grupo autenticado (middleware.Auth) — o usuário vem do
// TenantContext, nunca de um id no corpo (ver router.go).
package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	iamv1 "github.com/machv4/platform/gen/go/construtor/iam/v1"
	"github.com/machv4/platform/services/gateway/internal/web"
)

// ContaCliente é o subconjunto do IAMServiceClient usado pelas rotas de Conta.
type ContaCliente interface {
	AtualizarPerfil(ctx context.Context, in *iamv1.AtualizarPerfilRequest, opts ...grpc.CallOption) (*iamv1.AtualizarPerfilResponse, error)
	AtualizarSenha(ctx context.Context, in *iamv1.AtualizarSenhaRequest, opts ...grpc.CallOption) (*iamv1.AtualizarSenhaResponse, error)
	AtivarMfa(ctx context.Context, in *iamv1.AtivarMfaRequest, opts ...grpc.CallOption) (*iamv1.AtivarMfaResponse, error)
	ConfirmarMfa(ctx context.Context, in *iamv1.ConfirmarMfaRequest, opts ...grpc.CallOption) (*iamv1.ConfirmarMfaResponse, error)
	DesativarMfa(ctx context.Context, in *iamv1.DesativarMfaRequest, opts ...grpc.CallOption) (*iamv1.DesativarMfaResponse, error)
	ExcluirConta(ctx context.Context, in *iamv1.ExcluirContaRequest, opts ...grpc.CallOption) (*iamv1.ExcluirContaResponse, error)
	SolicitarTrocaEmail(ctx context.Context, in *iamv1.SolicitarTrocaEmailRequest, opts ...grpc.CallOption) (*iamv1.SolicitarTrocaEmailResponse, error)
	ConfirmarTrocaEmail(ctx context.Context, in *iamv1.ConfirmarTrocaEmailRequest, opts ...grpc.CallOption) (*iamv1.ConfirmarTrocaEmailResponse, error)
}

// atualizarPerfilReq é o corpo de PATCH /api/v1/conta/perfil.
type atualizarPerfilReq struct {
	Nome    string `json:"nome"`
	FotoURL string `json:"foto_url"`
}

// AtualizarPerfil serve PATCH /api/v1/conta/perfil (RF17).
func AtualizarPerfil(iam ContaCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body atualizarPerfilReq
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			web.Error(w, http.StatusBadRequest, "INVALID_BODY", "corpo JSON inválido")
			return
		}
		if _, err := iam.AtualizarPerfil(r.Context(), &iamv1.AtualizarPerfilRequest{Nome: body.Nome, FotoUrl: body.FotoURL}); err != nil {
			writeContaError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// atualizarSenhaReq é o corpo de PUT /api/v1/conta/senha.
type atualizarSenhaReq struct {
	SenhaAtual string `json:"senha_atual"`
	SenhaNova  string `json:"senha_nova"`
}

// AtualizarSenha serve PUT /api/v1/conta/senha (RF14, RNF02 — senha_atual é o
// mecanismo de reautenticação, validado no IAM antes de qualquer troca).
func AtualizarSenha(iam ContaCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body atualizarSenhaReq
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			web.Error(w, http.StatusBadRequest, "INVALID_BODY", "corpo JSON inválido")
			return
		}
		if _, err := iam.AtualizarSenha(r.Context(), &iamv1.AtualizarSenhaRequest{
			SenhaAtual: body.SenhaAtual,
			SenhaNova:  body.SenhaNova,
		}); err != nil {
			writeContaError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// ativarMfaResp espelha a otpauth:// URI de exibição única (RNF01).
type ativarMfaResp struct {
	SegredoOtpAuthURI string `json:"segredo_otp_auth_uri"`
}

// AtivarMfa serve POST /api/v1/conta/mfa/ativar (RF15, etapa 1 de 2: gera e
// mostra o segredo, ainda não liga o MFA).
func AtivarMfa(iam ContaCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		out, err := iam.AtivarMfa(r.Context(), &iamv1.AtivarMfaRequest{})
		if err != nil {
			writeContaError(w, err)
			return
		}
		web.JSON(w, http.StatusOK, ativarMfaResp{SegredoOtpAuthURI: out.GetSegredoOtpAuthUri()})
	}
}

// confirmarMfaReq é o corpo de POST /api/v1/conta/mfa/confirmar.
type confirmarMfaReq struct {
	Codigo string `json:"codigo"`
}

// ConfirmarMfa serve POST /api/v1/conta/mfa/confirmar (RF15, etapa 2 de 2:
// valida o código do app autenticador e liga o MFA).
func ConfirmarMfa(iam ContaCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body confirmarMfaReq
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			web.Error(w, http.StatusBadRequest, "INVALID_BODY", "corpo JSON inválido")
			return
		}
		if _, err := iam.ConfirmarMfa(r.Context(), &iamv1.ConfirmarMfaRequest{Codigo: body.Codigo}); err != nil {
			writeContaError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// senhaAtualReq é o corpo compartilhado por DELETE /api/v1/conta/mfa e DELETE
// /api/v1/conta — ambos exigem reautenticação por senha (RNF02).
type senhaAtualReq struct {
	SenhaAtual string `json:"senha_atual"`
}

// DesativarMfa serve DELETE /api/v1/conta/mfa (RF15). Decisão de segurança que
// diverge do contrato assumido em contracts/api.md (que não pedia
// reautenticação nesta rota): exige senha_atual no corpo — desativar 2FA sem
// confirmar a senha é superfície de abuso de sessão sequestrada.
func DesativarMfa(iam ContaCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body senhaAtualReq
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			web.Error(w, http.StatusBadRequest, "INVALID_BODY", "corpo JSON inválido")
			return
		}
		if _, err := iam.DesativarMfa(r.Context(), &iamv1.DesativarMfaRequest{SenhaAtual: body.SenhaAtual}); err != nil {
			writeContaError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// ExcluirConta serve DELETE /api/v1/conta (RF16, RN07). Também diverge do
// contrato assumido (que não enviava senha nenhuma): exige senha_atual no
// corpo pelo mesmo racional de RNF02 — client.ts e SegurancaForm.tsx foram
// atualizados para pedir a senha ao usuário antes de excluir.
func ExcluirConta(iam ContaCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body senhaAtualReq
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			web.Error(w, http.StatusBadRequest, "INVALID_BODY", "corpo JSON inválido")
			return
		}
		if _, err := iam.ExcluirConta(r.Context(), &iamv1.ExcluirContaRequest{SenhaAtual: body.SenhaAtual}); err != nil {
			writeContaError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// solicitarTrocaEmailReq é o corpo de POST /api/v1/conta/email.
type solicitarTrocaEmailReq struct {
	NovoEmail string `json:"novo_email"`
}

// SolicitarTrocaEmail serve POST /api/v1/conta/email (RF18, RN08): envia o
// link/token de confirmação ao novo endereço sem alterar o e-mail de login.
func SolicitarTrocaEmail(iam ContaCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body solicitarTrocaEmailReq
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			web.Error(w, http.StatusBadRequest, "INVALID_BODY", "corpo JSON inválido")
			return
		}
		if _, err := iam.SolicitarTrocaEmail(r.Context(), &iamv1.SolicitarTrocaEmailRequest{NovoEmail: body.NovoEmail}); err != nil {
			writeContaError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// confirmarTrocaEmailReq é o corpo de POST /api/v1/conta/email/confirmar.
type confirmarTrocaEmailReq struct {
	Token string `json:"token"`
}

// ConfirmarTrocaEmail serve POST /api/v1/conta/email/confirmar (RF18, RN08):
// efetiva a troca com o token recebido no e-mail.
func ConfirmarTrocaEmail(iam ContaCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body confirmarTrocaEmailReq
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			web.Error(w, http.StatusBadRequest, "INVALID_BODY", "corpo JSON inválido")
			return
		}
		if _, err := iam.ConfirmarTrocaEmail(r.Context(), &iamv1.ConfirmarTrocaEmailRequest{Token: body.Token}); err != nil {
			writeContaError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// writeContaError traduz os códigos gRPC das rotas de Conta para HTTP.
//
// Unauthenticated cobre dois casos distintos que o IAM sinaliza com o mesmo
// código gRPC — diferenciados aqui pela mensagem (decisão documentada, ver
// grpc.go): erroReautenticacaoNecessaria (senha_atual errada/ausente em
// troca de senha, desativar MFA ou excluir conta) vira 401
// REAUTENTICACAO_NECESSARIA; qualquer outro Unauthenticated (sessão ausente,
// código MFA errado/repetido) vira 401 UNAUTHORIZED genérico.
//
// FailedPrecondition também cobre dois casos: bloqueio de exclusão por tenant
// vinculado (RN07) vira 409 TENANT_ATIVO_VINCULADO; estado inválido de MFA
// (já ativo, ou confirmar sem ativação pendente) vira 409 MFA_INVALIDO —
// diferenciados pela presença da palavra "tenant" na mensagem.
func writeContaError(w http.ResponseWriter, err error) {
	st := status.Convert(err)
	msg := st.Message()

	switch st.Code() {
	case codes.InvalidArgument:
		if strings.Contains(msg, "obrigat") {
			web.Error(w, http.StatusBadRequest, "MISSING_PARAM", msg)
		} else {
			web.Error(w, http.StatusBadRequest, "INVALID_PARAM", msg)
		}
	case codes.Unauthenticated:
		if strings.Contains(msg, "reautenticação") {
			web.Error(w, http.StatusUnauthorized, "REAUTENTICACAO_NECESSARIA", msg)
		} else {
			web.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", msg)
		}
	case codes.FailedPrecondition:
		if strings.Contains(msg, "tenant") {
			web.Error(w, http.StatusConflict, "TENANT_ATIVO_VINCULADO", msg)
		} else {
			web.Error(w, http.StatusConflict, "MFA_INVALIDO", msg)
		}
	case codes.AlreadyExists:
		web.Error(w, http.StatusConflict, "EMAIL_JA_CADASTRADO", msg)
	default:
		web.Error(w, http.StatusInternalServerError, "INTERNAL", "erro interno")
	}
}
