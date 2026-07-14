package routes

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	deployv1 "github.com/machv4/platform/gen/go/construtor/deploy/v1"
	"github.com/machv4/platform/gateway/internal/web"
)

// DeployCliente é o subconjunto do DeployEngineServiceClient usado pelas rotas de
// publicação/rollback (RF04).
type DeployCliente interface {
	Publicar(ctx context.Context, in *deployv1.PublicarRequest, opts ...grpc.CallOption) (*deployv1.PublicarResponse, error)
	Rollback(ctx context.Context, in *deployv1.RollbackRequest, opts ...grpc.CallOption) (*deployv1.RollbackResponse, error)
	ObterVersaoAtiva(ctx context.Context, in *deployv1.ObterVersaoAtivaRequest, opts ...grpc.CallOption) (*deployv1.VersaoAtiva, error)
}

// Publicar serve POST /api/v1/sistemas/{sistema_id}/publicar (RF04, RN04).
func Publicar(deploy DeployCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		out, err := deploy.Publicar(r.Context(), &deployv1.PublicarRequest{SistemaId: chi.URLParam(r, "sistema_id")})
		if err != nil {
			writeDeployError(w, err)
			return
		}
		web.JSON(w, http.StatusCreated, map[string]any{
			"versao_id": out.GetVersaoId(),
			"numero":    out.GetNumero(),
			"ativa":     true,
		})
	}
}

// Rollback serve POST /api/v1/sistemas/{sistema_id}/rollback (RF04, RN05). O corpo
// é opcional; versao_numero omitido/0 = versão imediatamente anterior.
func Rollback(deploy DeployCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			VersaoNumero int32 `json:"versao_numero"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil && !errors.Is(err, io.EOF) {
			web.Error(w, http.StatusBadRequest, "INVALID_BODY", "corpo JSON inválido")
			return
		}
		out, err := deploy.Rollback(r.Context(), &deployv1.RollbackRequest{
			SistemaId:    chi.URLParam(r, "sistema_id"),
			VersaoNumero: body.VersaoNumero,
		})
		if err != nil {
			writeDeployError(w, err)
			return
		}
		web.JSON(w, http.StatusOK, map[string]int32{"versao_ativa": out.GetVersaoAtiva()})
	}
}

// VersaoAtiva serve GET /api/v1/sistemas/{sistema_id}/versao-ativa (RN04).
func VersaoAtiva(deploy DeployCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		out, err := deploy.ObterVersaoAtiva(r.Context(), &deployv1.ObterVersaoAtivaRequest{SistemaId: chi.URLParam(r, "sistema_id")})
		if err != nil {
			writeDeployError(w, err)
			return
		}
		web.JSON(w, http.StatusOK, map[string]any{
			"versao_id": out.GetVersaoId(),
			"numero":    out.GetNumero(),
			"definicao": json.RawMessage(out.GetDefinicaoJson()),
		})
	}
}

// writeDeployError traduz os códigos gRPC do Deploy Engine para HTTP conforme
// api.md: Aborted → 409 PUBLISH_IN_PROGRESS; FailedPrecondition → 422
// NO_PREVIOUS_VERSION.
func writeDeployError(w http.ResponseWriter, err error) {
	switch status.Code(err) {
	case codes.Aborted:
		web.Error(w, http.StatusConflict, "PUBLISH_IN_PROGRESS", "outra publicação em andamento para o sistema")
	case codes.FailedPrecondition:
		web.Error(w, http.StatusUnprocessableEntity, "NO_PREVIOUS_VERSION", "não há versão anterior para reverter")
	case codes.NotFound:
		web.Error(w, http.StatusNotFound, "NOT_FOUND", "recurso inexistente")
	case codes.InvalidArgument:
		web.Error(w, http.StatusBadRequest, "MISSING_PARAM", "parâmetro obrigatório ausente")
	case codes.Unauthenticated:
		web.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "não autenticado")
	default:
		web.Error(w, http.StatusInternalServerError, "INTERNAL", "erro interno")
	}
}
