// Package routes traduz as rotas REST do Gateway em chamadas gRPC internas.
package routes

import (
	"context"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	iamv1 "github.com/machv4/platform/gen/go/construtor/iam/v1"
	"github.com/machv4/platform/gateway/internal/web"
)

// PermissoesAvaliador é o subconjunto do IAMServiceClient usado por esta rota.
type PermissoesAvaliador interface {
	AvaliarPermissoes(ctx context.Context, in *iamv1.AvaliarPermissoesRequest, opts ...grpc.CallOption) (*iamv1.AvaliarPermissoesResponse, error)
}

type permissaoDTO struct {
	View  bool `json:"view"`
	Click bool `json:"click"`
}

type permissoesResp struct {
	Permissions map[string]permissaoDTO `json:"permissions"`
}

// Permissoes serve GET /api/v1/permissoes?sistema_id={uuid}&bi=...&bi=...
//
// O tenant vem do TenantContext no contexto (posto pelo Auth) e é propagado ao
// IAM via Metadata gRPC — nunca do query string (RNF02). Os blind_indexes são,
// nesta fase, informados explicitamente; quando o Deploy Engine existir (Fase 5)
// virão da versão ativa do sistema.
func Permissoes(iam PermissoesAvaliador) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		sistemaID := q.Get("sistema_id")
		if sistemaID == "" {
			web.Error(w, http.StatusBadRequest, "MISSING_PARAM", "sistema_id obrigatório")
			return
		}

		resp, err := iam.AvaliarPermissoes(r.Context(), &iamv1.AvaliarPermissoesRequest{
			SistemaId:    sistemaID,
			BlindIndexes: q["bi"],
		})
		if err != nil {
			writeGRPCError(w, err)
			return
		}

		out := permissoesResp{Permissions: make(map[string]permissaoDTO, len(resp.GetPermissions()))}
		for bi, p := range resp.GetPermissions() {
			out.Permissions[bi] = permissaoDTO{View: p.GetView(), Click: p.GetClick()}
		}
		web.JSON(w, http.StatusOK, out)
	}
}

// writeGRPCError traduz códigos gRPC para HTTP conforme api.md.
func writeGRPCError(w http.ResponseWriter, err error) {
	switch status.Code(err) {
	case codes.Unauthenticated:
		web.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "não autenticado")
	case codes.PermissionDenied:
		web.Error(w, http.StatusForbidden, "FORBIDDEN", "acesso negado")
	case codes.NotFound:
		web.Error(w, http.StatusNotFound, "NOT_FOUND", "recurso inexistente")
	default:
		web.Error(w, http.StatusInternalServerError, "INTERNAL", "erro interno")
	}
}
