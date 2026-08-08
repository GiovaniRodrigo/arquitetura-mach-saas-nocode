package routes

import (
	"encoding/json"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	designv1 "github.com/machv4/platform/gen/go/construtor/design/v1"
	"github.com/machv4/platform/services/gateway/internal/web"
)

// whiteLabelResp espelha a personalização de marca do tenant no corpo JSON.
type whiteLabelResp struct {
	LogoURL         string `json:"logo_url"`
	CorPrimaria     string `json:"cor_primaria"`
	CorSecundaria   string `json:"cor_secundaria"`
	DominioProprio  string `json:"dominio_proprio"`
	DominioValidado bool   `json:"dominio_validado"`
}

// atualizarWhiteLabelReq é o corpo de PUT /api/v1/configuracao/white-label.
type atualizarWhiteLabelReq struct {
	LogoURL        string `json:"logo_url"`
	CorPrimaria    string `json:"cor_primaria"`
	CorSecundaria  string `json:"cor_secundaria"`
	DominioProprio string `json:"dominio_proprio"`
}

// AtualizarWhiteLabel serve PUT /api/v1/configuracao/white-label (spec 004,
// RF13, RNF03): atualiza logo/cores/domínio próprio do tenant do contexto. A
// validação do domínio é assíncrona e fora de escopo — dominio_validado
// devolve sempre false. Responde 202 quando dominio_proprio veio preenchido no
// request (validação ainda pendente); 200 caso contrário.
func AtualizarWhiteLabel(design DesignCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body atualizarWhiteLabelReq
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			web.Error(w, http.StatusBadRequest, "INVALID_BODY", "corpo JSON inválido")
			return
		}
		out, err := design.AtualizarWhiteLabel(r.Context(), &designv1.AtualizarWhiteLabelRequest{
			LogoUrl:        body.LogoURL,
			CorPrimaria:    body.CorPrimaria,
			CorSecundaria:  body.CorSecundaria,
			DominioProprio: body.DominioProprio,
		})
		if err != nil {
			writeWhiteLabelError(w, err)
			return
		}
		resp := whiteLabelResp{
			LogoURL:         out.GetLogoUrl(),
			CorPrimaria:     out.GetCorPrimaria(),
			CorSecundaria:   out.GetCorSecundaria(),
			DominioProprio:  out.GetDominioProprio(),
			DominioValidado: out.GetDominioValidado(),
		}
		httpStatus := http.StatusOK
		if body.DominioProprio != "" {
			httpStatus = http.StatusAccepted
		}
		web.JSON(w, httpStatus, resp)
	}
}

// writeWhiteLabelError traduz os códigos gRPC do Design Engine para HTTP na
// rota de white-label.
func writeWhiteLabelError(w http.ResponseWriter, err error) {
	switch status.Code(err) {
	case codes.InvalidArgument:
		web.Error(w, http.StatusBadRequest, "INVALID_BODY", "corpo inválido")
	case codes.Unauthenticated:
		web.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "não autenticado")
	default:
		web.Error(w, http.StatusInternalServerError, "INTERNAL", "erro interno")
	}
}
