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
	"github.com/machv4/platform/services/gateway/internal/web"
)

// RegrasNegocioCliente é o subconjunto do LogicEngineServiceClient usado pelas
// rotas de regras de validação de componente (RF10/RF11, RN06). Entidade
// distinta de RegrasCliente (árvore de decisão do motor de ações).
type RegrasNegocioCliente interface {
	ListarRegrasValidacao(ctx context.Context, in *logicv1.ListarRegrasValidacaoRequest, opts ...grpc.CallOption) (*logicv1.ListarRegrasValidacaoResponse, error)
	CriarRegraValidacao(ctx context.Context, in *logicv1.RegraValidacao, opts ...grpc.CallOption) (*logicv1.RegraValidacao, error)
}

// regraNegocioReq é o corpo de POST /api/v1/sistemas/{sistema_id}/regras-negocio.
type regraNegocioReq struct {
	BlindIndexes []string        `json:"blind_indexes"`
	Tipo         string          `json:"tipo"`
	Parametros   json.RawMessage `json:"parametros"`
}

// regraNegocioResp espelha uma regra de validação de componente no corpo JSON
// (services/frontend/src/api/types.ts: RegraNegocio).
type regraNegocioResp struct {
	ID           string         `json:"id"`
	BlindIndexes []string       `json:"blind_indexes"`
	Tipo         string         `json:"tipo"`
	Parametros   map[string]any `json:"parametros"`
}

func regraNegocioRespFrom(r *logicv1.RegraValidacao) regraNegocioResp {
	parametros := map[string]any{}
	if p := r.GetParametros(); len(p) > 0 {
		_ = json.Unmarshal(p, &parametros)
	}
	return regraNegocioResp{
		ID:           r.GetId(),
		BlindIndexes: r.GetBlindIndexes(),
		Tipo:         r.GetTipo(),
		Parametros:   parametros,
	}
}

// ListarRegrasNegocio serve GET /api/v1/sistemas/{sistema_id}/regras-negocio
// (RF10/RF11): devolve as regras de validação de estado de componente do
// sistema, restritas ao tenant do contexto.
func ListarRegrasNegocio(logic RegrasNegocioCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sistemaID := chi.URLParam(r, "sistema_id")
		out, err := logic.ListarRegrasValidacao(r.Context(), &logicv1.ListarRegrasValidacaoRequest{SistemaId: sistemaID})
		if err != nil {
			writeRegraNegocioError(w, err)
			return
		}
		lista := make([]regraNegocioResp, 0, len(out.GetRegras()))
		for _, r := range out.GetRegras() {
			lista = append(lista, regraNegocioRespFrom(r))
		}
		web.JSON(w, http.StatusOK, map[string]any{"regras": lista})
	}
}

// CriarRegraNegocio serve POST /api/v1/sistemas/{sistema_id}/regras-negocio
// (RF10/RF11, RN06): cria uma regra de validação (regex/tamanho/obrigatorio)
// para um ou mais componentes do sistema. Valida no Gateway também —
// blind_indexes não-vazio e parametros como JSON válido — antes de repassar ao
// gRPC como bytes (o servidor revalida de qualquer forma, nunca confiando no
// client).
func CriarRegraNegocio(logic RegrasNegocioCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sistemaID := chi.URLParam(r, "sistema_id")

		var body regraNegocioReq
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			web.Error(w, http.StatusBadRequest, "INVALID_BODY", "corpo JSON inválido")
			return
		}
		if len(body.BlindIndexes) == 0 {
			web.Error(w, http.StatusBadRequest, "MISSING_PARAM", "blind_indexes obrigatório")
			return
		}
		parametros := body.Parametros
		if len(parametros) == 0 {
			parametros = json.RawMessage(`{}`)
		}
		if !json.Valid(parametros) {
			web.Error(w, http.StatusBadRequest, "INVALID_BODY", "parametros deve ser um JSON válido")
			return
		}

		out, err := logic.CriarRegraValidacao(r.Context(), &logicv1.RegraValidacao{
			SistemaId:    sistemaID,
			BlindIndexes: body.BlindIndexes,
			Tipo:         body.Tipo,
			Parametros:   parametros,
		})
		if err != nil {
			writeRegraNegocioError(w, err)
			return
		}
		web.JSON(w, http.StatusCreated, regraNegocioRespFrom(out))
	}
}

// writeRegraNegocioError traduz códigos gRPC para HTTP nas rotas de regras de
// validação de componente: InvalidArgument → 400; PermissionDenied → 403;
// NotFound → 404; Unauthenticated → 401 (mesmo padrão de writeSistemaError /
// writeTenantError).
func writeRegraNegocioError(w http.ResponseWriter, err error) {
	switch status.Code(err) {
	case codes.InvalidArgument:
		web.Error(w, http.StatusBadRequest, "INVALID_PARAM", "blind_indexes ou tipo inválido")
	case codes.PermissionDenied:
		web.Error(w, http.StatusForbidden, "FORBIDDEN", "sem permissão para gerenciar regras de negócio")
	case codes.NotFound:
		web.Error(w, http.StatusNotFound, "NOT_FOUND", "sistema não encontrado")
	case codes.Unauthenticated:
		web.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "não autenticado")
	default:
		web.Error(w, http.StatusInternalServerError, "INTERNAL", "erro interno")
	}
}
