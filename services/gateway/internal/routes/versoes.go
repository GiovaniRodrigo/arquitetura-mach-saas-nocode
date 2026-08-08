package routes

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"

	deployv1 "github.com/machv4/platform/gen/go/construtor/deploy/v1"
	"github.com/machv4/platform/services/gateway/internal/web"
)

// VersoesCliente é o subconjunto do DeployEngineServiceClient usado pela fachada
// REST de versoes/{id} (spec 004, RF12) — sobre a mesma lógica de
// publicar/rollback já servida por DeployCliente em deploy.go.
type VersoesCliente interface {
	ListarVersoes(ctx context.Context, in *deployv1.ListarVersoesRequest, opts ...grpc.CallOption) (*deployv1.ListarVersoesResponse, error)
	PublicarVersao(ctx context.Context, in *deployv1.PublicarVersaoRequest, opts ...grpc.CallOption) (*deployv1.PublicarResponse, error)
	ReverterVersao(ctx context.Context, in *deployv1.ReverterVersaoRequest, opts ...grpc.CallOption) (*deployv1.ReverterVersaoResponse, error)
}

// ListarVersoes serve GET /api/v1/sistemas/{sistema_id}/versoes (RF12).
func ListarVersoes(deploy VersoesCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		out, err := deploy.ListarVersoes(r.Context(), &deployv1.ListarVersoesRequest{SistemaId: chi.URLParam(r, "sistema_id")})
		if err != nil {
			writeDeployError(w, err)
			return
		}
		versoes := make([]map[string]any, 0, len(out.GetVersoes()))
		for _, v := range out.GetVersoes() {
			versoes = append(versoes, map[string]any{
				"id":        v.GetId(),
				"numero":    v.GetNumero(),
				"ativa":     v.GetAtiva(),
				"criado_em": v.GetCriadoEm(),
			})
		}
		web.JSON(w, http.StatusOK, map[string]any{"versoes": versoes})
	}
}

// PublicarVersaoRota serve POST /api/v1/sistemas/{sistema_id}/versoes/{versao_id}/publicar
// (RF12). O {versao_id} do path é ignorado de propósito: hoje não existe o
// conceito de "versão inativa aguardando publicação" no schema — toda linha de
// versoes_sistema já foi ativa em algum momento, e Publicar sempre cria uma
// versão NOVA a partir do estado atual (nunca reaproveita uma linha existente).
// Então esta rota apenas delega para a mesma lógica de Manager.Publicar já usada
// por POST /sistemas/{sistema_id}/publicar.
func PublicarVersaoRota(deploy VersoesCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		out, err := deploy.PublicarVersao(r.Context(), &deployv1.PublicarVersaoRequest{SistemaId: chi.URLParam(r, "sistema_id")})
		if err != nil {
			writeDeployError(w, err)
			return
		}
		web.JSON(w, http.StatusCreated, map[string]any{"versao_ativa": out.GetNumero()})
	}
}

// ReverterVersaoRota serve POST /api/v1/sistemas/{sistema_id}/versoes/{versao_id}/reverter
// (RF12). Diferente de PublicarVersaoRota, aqui {versao_id} é o identificador
// (UUID) da versão-alvo: é resolvido para o número correspondente e delega para
// Manager.Rollback (mesma semântica de RN05) via RollbackPorID.
func ReverterVersaoRota(deploy VersoesCliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		out, err := deploy.ReverterVersao(r.Context(), &deployv1.ReverterVersaoRequest{
			SistemaId: chi.URLParam(r, "sistema_id"),
			VersaoId:  chi.URLParam(r, "versao_id"),
		})
		if err != nil {
			writeDeployError(w, err)
			return
		}
		web.JSON(w, http.StatusOK, map[string]any{"versao_ativa": out.GetVersaoAtiva()})
	}
}
