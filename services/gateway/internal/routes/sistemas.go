package routes

import (
	"encoding/json"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	designv1 "github.com/machv4/platform/gen/go/construtor/design/v1"
	"github.com/machv4/platform/services/gateway/internal/web"
)

// sistemaResp espelha um sistema no corpo JSON.
type sistemaResp struct {
	ID   string `json:"id"`
	Nome string `json:"nome"`
}

// criarSistemaReq é o corpo de POST /api/v1/sistemas.
type criarSistemaReq struct {
	Nome string `json:"nome"`
}

// ListarSistemas serve GET /api/v1/sistemas (RN01): devolve os sistemas do tenant
// do contexto. O tenant vem do TenantContext (posto pelo Auth) e é propagado ao
// Design Engine via Metadata gRPC.
func ListarSistemas(design DesignCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		out, err := design.ListarSistemas(r.Context(), &designv1.ListarSistemasRequest{})
		if err != nil {
			writeSistemaError(w, err)
			return
		}
		lista := make([]sistemaResp, 0, len(out.GetSistemas()))
		for _, s := range out.GetSistemas() {
			lista = append(lista, sistemaResp{ID: s.GetId(), Nome: s.GetNome()})
		}
		web.JSON(w, http.StatusOK, map[string]any{"sistemas": lista})
	}
}

// CriarSistema serve POST /api/v1/sistemas (RN01, RN02): cria um sistema no tenant
// corrente. Restrito a dono/parceiro no Design Engine → PermissionDenied vira 403.
func CriarSistema(design DesignCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body criarSistemaReq
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			web.Error(w, http.StatusBadRequest, "INVALID_BODY", "corpo JSON inválido")
			return
		}
		if body.Nome == "" {
			web.Error(w, http.StatusBadRequest, "MISSING_PARAM", "nome obrigatório")
			return
		}
		out, err := design.CriarSistema(r.Context(), &designv1.CriarSistemaRequest{Nome: body.Nome})
		if err != nil {
			writeSistemaError(w, err)
			return
		}
		web.JSON(w, http.StatusCreated, sistemaResp{ID: out.GetId(), Nome: out.GetNome()})
	}
}

// writeSistemaError traduz os códigos gRPC do Design Engine para HTTP nas rotas
// de sistemas: PermissionDenied (cliente final) → 403; InvalidArgument → 400.
func writeSistemaError(w http.ResponseWriter, err error) {
	switch status.Code(err) {
	case codes.InvalidArgument:
		web.Error(w, http.StatusBadRequest, "MISSING_PARAM", "nome obrigatório")
	case codes.PermissionDenied:
		web.Error(w, http.StatusForbidden, "FORBIDDEN", "sem permissão para criar sistema")
	case codes.Unauthenticated:
		web.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "não autenticado")
	default:
		web.Error(w, http.StatusInternalServerError, "INTERNAL", "erro interno")
	}
}
