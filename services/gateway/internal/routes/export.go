package routes

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"

	exportv1 "github.com/machv4/platform/gen/go/construtor/export/v1"
	"github.com/machv4/platform/services/gateway/internal/web"
)

// ExportCliente é o subconjunto do ExportEngineServiceClient usado pelas rotas
// (RF05). O streaming ColetarDados é interno ao serviço, não exposto no Gateway.
type ExportCliente interface {
	CriarJob(ctx context.Context, in *exportv1.CriarJobRequest, opts ...grpc.CallOption) (*exportv1.Job, error)
	ObterJob(ctx context.Context, in *exportv1.ObterJobRequest, opts ...grpc.CallOption) (*exportv1.Job, error)
}

// CriarExportacao serve POST /api/v1/exportacoes — cria o Job e responde 202 já
// (RF05). O trabalho de coleta/upload corre em segundo plano no Export Engine.
func CriarExportacao(export ExportCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			SistemaID string `json:"sistema_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			web.Error(w, http.StatusBadRequest, "INVALID_BODY", "corpo JSON inválido")
			return
		}
		if body.SistemaID == "" {
			web.Error(w, http.StatusBadRequest, "MISSING_PARAM", "sistema_id obrigatório")
			return
		}
		out, err := export.CriarJob(r.Context(), &exportv1.CriarJobRequest{SistemaId: body.SistemaID})
		if err != nil {
			writeGRPCError(w, err)
			return
		}
		web.JSON(w, http.StatusAccepted, map[string]string{
			"job_id": out.GetId(),
			"status": out.GetStatus(),
		})
	}
}

// ObterExportacao serve GET /api/v1/exportacoes/{job_id} — estado e, se pronto, o
// Presigned URL de expiração curta (RF05, critério 7).
func ObterExportacao(export ExportCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		out, err := export.ObterJob(r.Context(), &exportv1.ObterJobRequest{Id: chi.URLParam(r, "job_id")})
		if err != nil {
			writeGRPCError(w, err)
			return
		}
		resp := map[string]string{"job_id": out.GetId(), "status": out.GetStatus()}
		if out.GetArquivoUrl() != "" {
			resp["arquivo_url"] = out.GetArquivoUrl()
		}
		if out.GetExpiraEm() != "" {
			resp["expira_em"] = out.GetExpiraEm()
		}
		web.JSON(w, http.StatusOK, resp)
	}
}
