package routes

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	logicv1 "github.com/machv4/platform/gen/go/construtor/logic/v1"
	"github.com/machv4/platform/gateway/internal/web"
)

// RegrasCliente é o subconjunto do LogicEngineServiceClient usado pelas rotas de
// regras (RF02).
type RegrasCliente interface {
	CriarRegra(ctx context.Context, in *logicv1.Regra, opts ...grpc.CallOption) (*logicv1.Regra, error)
	ObterRegra(ctx context.Context, in *logicv1.ObterRegraRequest, opts ...grpc.CallOption) (*logicv1.Regra, error)
	AtualizarRegra(ctx context.Context, in *logicv1.Regra, opts ...grpc.CallOption) (*logicv1.Regra, error)
	RemoverRegra(ctx context.Context, in *logicv1.ObterRegraRequest, opts ...grpc.CallOption) (*logicv1.RemoverResponse, error)
}

type regraReq struct {
	SistemaID     string          `json:"sistema_id"`
	ArvoreDecisao json.RawMessage `json:"arvore_decisao"`
}

type regraResp struct {
	ID            string          `json:"id"`
	SistemaID     string          `json:"sistema_id"`
	ArvoreDecisao json.RawMessage `json:"arvore_decisao"`
}

func regraRespFrom(r *logicv1.Regra) regraResp {
	return regraResp{ID: r.GetId(), SistemaID: r.GetSistemaId(), ArvoreDecisao: r.GetArvoreDecisao()}
}

// CriarRegra serve POST /api/v1/regras (RF02).
func CriarRegra(logic RegrasCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body regraReq
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			web.Error(w, http.StatusBadRequest, "INVALID_BODY", "corpo JSON inválido")
			return
		}
		if body.SistemaID == "" {
			web.Error(w, http.StatusBadRequest, "MISSING_PARAM", "sistema_id obrigatório")
			return
		}
		out, err := logic.CriarRegra(r.Context(), &logicv1.Regra{
			SistemaId:     body.SistemaID,
			ArvoreDecisao: body.ArvoreDecisao,
		})
		if err != nil {
			writeRegraError(w, err)
			return
		}
		web.JSON(w, http.StatusCreated, map[string]string{"regra_id": out.GetId()})
	}
}

// ObterRegra serve GET /api/v1/regras/{id} (RF02).
func ObterRegra(logic RegrasCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		out, err := logic.ObterRegra(r.Context(), &logicv1.ObterRegraRequest{Id: chi.URLParam(r, "id")})
		if err != nil {
			writeRegraError(w, err)
			return
		}
		web.JSON(w, http.StatusOK, regraRespFrom(out))
	}
}

// AtualizarRegra serve PUT /api/v1/regras/{id} (RF02).
func AtualizarRegra(logic RegrasCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body regraReq
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			web.Error(w, http.StatusBadRequest, "INVALID_BODY", "corpo JSON inválido")
			return
		}
		out, err := logic.AtualizarRegra(r.Context(), &logicv1.Regra{
			Id:            chi.URLParam(r, "id"),
			SistemaId:     body.SistemaID,
			ArvoreDecisao: body.ArvoreDecisao,
		})
		if err != nil {
			writeRegraError(w, err)
			return
		}
		web.JSON(w, http.StatusOK, regraRespFrom(out))
	}
}

// RemoverRegra serve DELETE /api/v1/regras/{id} (RF02).
func RemoverRegra(logic RegrasCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := logic.RemoverRegra(r.Context(), &logicv1.ObterRegraRequest{Id: chi.URLParam(r, "id")}); err != nil {
			writeRegraError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// writeRegraError traduz códigos gRPC para HTTP — InvalidArgument (árvore de
// decisão malformada) vira 422 INVALID_DECISION_TREE conforme api.md.
func writeRegraError(w http.ResponseWriter, err error) {
	switch status.Code(err) {
	case codes.InvalidArgument:
		web.Error(w, http.StatusUnprocessableEntity, "INVALID_DECISION_TREE", "nó lógico desconhecido ou ciclo detectado")
	case codes.NotFound:
		web.Error(w, http.StatusNotFound, "NOT_FOUND", "regra não encontrada")
	case codes.Unauthenticated:
		web.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "não autenticado")
	default:
		web.Error(w, http.StatusInternalServerError, "INTERNAL", "erro interno")
	}
}
