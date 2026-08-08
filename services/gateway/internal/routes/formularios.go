package routes

import (
	"context"
	"encoding/json"
	"net/http"

	"google.golang.org/grpc"

	logicv1 "github.com/machv4/platform/gen/go/construtor/logic/v1"
	"github.com/machv4/platform/services/gateway/internal/web"
)

// FormulariosCliente é o subconjunto do LogicEngineServiceClient usado pela rota
// de submissão de formulários (RF07).
type FormulariosCliente interface {
	SalvarFormulario(ctx context.Context, in *logicv1.SalvarFormularioRequest, opts ...grpc.CallOption) (*logicv1.SalvarFormularioResponse, error)
}

type formularioReq struct {
	SistemaID       string            `json:"sistema_id"`
	DadosFormulario map[string]string `json:"dados_formulario"`
}

type formularioResp struct {
	Sucesso        bool              `json:"sucesso"`
	ErrosValidacao map[string]string `json:"erros_validacao,omitempty"`
	MensagemStatus string            `json:"mensagem_status"`
}

// SalvarFormulario serve POST /api/v1/formularios (RF07, RN08). Uma falha de
// validação não é erro de transporte: o Logic Engine devolve sucesso=false com o
// mapa de erros por blind_index, que respondemos como 422 sem expor nomes reais
// (RNF08, critério 2).
func SalvarFormulario(logic FormulariosCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body formularioReq
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			web.Error(w, http.StatusBadRequest, "INVALID_BODY", "corpo JSON inválido")
			return
		}
		if body.SistemaID == "" {
			web.Error(w, http.StatusBadRequest, "MISSING_PARAM", "sistema_id obrigatório")
			return
		}

		resp, err := logic.SalvarFormulario(r.Context(), &logicv1.SalvarFormularioRequest{
			SistemaId:       body.SistemaID,
			DadosFormulario: body.DadosFormulario,
		})
		if err != nil {
			writeGRPCError(w, err)
			return
		}

		out := formularioResp{
			Sucesso:        resp.GetSucesso(),
			ErrosValidacao: resp.GetErrosValidacao(),
			MensagemStatus: resp.GetMensagemStatus(),
		}
		if !resp.GetSucesso() {
			web.JSON(w, http.StatusUnprocessableEntity, out)
			return
		}
		web.JSON(w, http.StatusOK, out)
	}
}
