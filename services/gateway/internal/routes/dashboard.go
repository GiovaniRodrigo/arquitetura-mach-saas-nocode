package routes

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	iamv1 "github.com/machv4/platform/gen/go/construtor/iam/v1"
	"github.com/machv4/platform/services/gateway/internal/web"
)

// DashboardCliente é o subconjunto do IAMServiceClient usado pelas rotas do
// Dashboard (spec 004, RF04/RF05/RF06).
type DashboardCliente interface {
	ListarUltimosAcessos(ctx context.Context, in *iamv1.ListarUltimosAcessosRequest, opts ...grpc.CallOption) (*iamv1.ListarUltimosAcessosResponse, error)
	ListarFeedback(ctx context.Context, in *iamv1.ListarFeedbackRequest, opts ...grpc.CallOption) (*iamv1.ListarFeedbackResponse, error)
	AtualizarStatusFeedback(ctx context.Context, in *iamv1.AtualizarStatusFeedbackRequest, opts ...grpc.CallOption) (*iamv1.Feedback, error)
	ResumoFinanceiro(ctx context.Context, in *iamv1.ResumoFinanceiroRequest, opts ...grpc.CallOption) (*iamv1.ResumoFinanceiroResponse, error)
}

// eventoLoginResp espelha um EventoLogin no corpo JSON (services/frontend/src/api/types.ts:69-73).
type eventoLoginResp struct {
	UsuarioNome string `json:"usuario_nome"`
	TenantNome  string `json:"tenant_nome"`
	CriadoEm    string `json:"criado_em"`
}

// feedbackResp espelha um Feedback no corpo JSON (services/frontend/src/api/types.ts:79-85).
type feedbackResp struct {
	ID         string `json:"id"`
	TenantNome string `json:"tenant_nome"`
	Mensagem   string `json:"mensagem"`
	Status     string `json:"status"`
	CriadoEm   string `json:"criado_em"`
}

// atualizarStatusFeedbackReq é o corpo de PATCH /api/v1/dashboard/feedback/{id}.
type atualizarStatusFeedbackReq struct {
	Status string `json:"status"`
}

// resumoFinanceiroResp espelha ResumoFinanceiro no corpo JSON (services/frontend/src/api/types.ts:88-92).
type resumoFinanceiroResp struct {
	ReceitaTotalCentavos int64  `json:"receita_total_centavos"`
	Moeda                string `json:"moeda"`
	Competencia          string `json:"competencia"`
}

// ListarUltimosAcessos serve GET /api/v1/dashboard/ultimos-acessos (spec 004,
// RF04, RN02): os 10 logins mais recentes agregados dos tenants vinculados ao
// tenant do contexto.
func ListarUltimosAcessos(iam DashboardCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		out, err := iam.ListarUltimosAcessos(r.Context(), &iamv1.ListarUltimosAcessosRequest{})
		if err != nil {
			writeDashboardError(w, err)
			return
		}
		eventos := make([]eventoLoginResp, 0, len(out.GetEventos()))
		for _, e := range out.GetEventos() {
			eventos = append(eventos, eventoLoginResp{
				UsuarioNome: e.GetUsuarioNome(),
				TenantNome:  e.GetTenantNome(),
				CriadoEm:    e.GetCriadoEm(),
			})
		}
		web.JSON(w, http.StatusOK, map[string]any{"eventos": eventos})
	}
}

// ListarFeedback serve GET /api/v1/dashboard/feedback (spec 004, RF05): as
// mensagens de feedback dos tenants vinculados, opcionalmente filtradas por
// `?status=`.
func ListarFeedback(iam DashboardCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		out, err := iam.ListarFeedback(r.Context(), &iamv1.ListarFeedbackRequest{Status: r.URL.Query().Get("status")})
		if err != nil {
			writeDashboardError(w, err)
			return
		}
		itens := make([]feedbackResp, 0, len(out.GetItens()))
		for _, f := range out.GetItens() {
			itens = append(itens, feedbackResp{
				ID:         f.GetId(),
				TenantNome: f.GetTenantNome(),
				Mensagem:   f.GetMensagem(),
				Status:     f.GetStatus(),
				CriadoEm:   f.GetCriadoEm(),
			})
		}
		web.JSON(w, http.StatusOK, map[string]any{"itens": itens})
	}
}

// AtualizarStatusFeedback serve PATCH /api/v1/dashboard/feedback/{id} (spec
// 004, RF05, RN03): a única transição aceita é pendente→respondido.
func AtualizarStatusFeedback(iam DashboardCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body atualizarStatusFeedbackReq
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			web.Error(w, http.StatusBadRequest, "INVALID_BODY", "corpo JSON inválido")
			return
		}
		out, err := iam.AtualizarStatusFeedback(r.Context(), &iamv1.AtualizarStatusFeedbackRequest{
			Id:     chi.URLParam(r, "id"),
			Status: body.Status,
		})
		if err != nil {
			writeDashboardError(w, err)
			return
		}
		web.JSON(w, http.StatusOK, feedbackResp{
			ID:         out.GetId(),
			TenantNome: out.GetTenantNome(),
			Mensagem:   out.GetMensagem(),
			Status:     out.GetStatus(),
			CriadoEm:   out.GetCriadoEm(),
		})
	}
}

// ResumoFinanceiro serve GET /api/v1/dashboard/resumo-financeiro (spec 004,
// RF06, RN04): a receita agregada do mês corrente dos tenants vinculados.
func ResumoFinanceiro(iam DashboardCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		out, err := iam.ResumoFinanceiro(r.Context(), &iamv1.ResumoFinanceiroRequest{})
		if err != nil {
			writeDashboardError(w, err)
			return
		}
		web.JSON(w, http.StatusOK, resumoFinanceiroResp{
			ReceitaTotalCentavos: out.GetReceitaTotalCentavos(),
			Moeda:                out.GetMoeda(),
			Competencia:          out.GetCompetencia(),
		})
	}
}

// writeDashboardError traduz os códigos gRPC do IAM para HTTP nas rotas do
// Dashboard: InvalidArgument → 400; Unauthenticated → 401; NotFound → 404;
// default → 500.
func writeDashboardError(w http.ResponseWriter, err error) {
	switch status.Code(err) {
	case codes.InvalidArgument:
		web.Error(w, http.StatusBadRequest, "INVALID_PARAM", "parâmetro inválido")
	case codes.Unauthenticated:
		web.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "não autenticado")
	case codes.NotFound:
		web.Error(w, http.StatusNotFound, "NOT_FOUND", "não encontrado")
	default:
		web.Error(w, http.StatusInternalServerError, "INTERNAL", "erro interno")
	}
}
