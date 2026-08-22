package routes

import (
	"context"
	"net/http"

	"github.com/machv4/platform/services/gateway/internal/meshmetrics"
	"github.com/machv4/platform/services/gateway/internal/web"
)

// recursosClient é satisfeita por *meshmetrics.Client em produção; os testes
// usam um fake, já que Client fala HTTP de verdade com o cluster.
type recursosClient interface {
	ObterRecursos(ctx context.Context) ([]meshmetrics.ServicoRecurso, error)
}

// servicoRecursoResp é o corpo JSON de um serviço em GET /api/v1/monitor/recursos
// (spec 009 — substitui o polling gRPC customizado da spec 008 original por
// telemetria lida direto do Kubernetes/Linkerd).
type servicoRecursoResp struct {
	Nome                  string  `json:"nome"`
	Status                string  `json:"status"`
	CPUMillicores         int64   `json:"cpu_millicores"`
	MemoriaBytes          int64   `json:"memoria_bytes"`
	RequisicoesPorSegundo float64 `json:"requisicoes_por_segundo"`
	TaxaSucessoPercent    float64 `json:"taxa_sucesso_percent"`
	LatenciaP99Ms         float64 `json:"latencia_p99_ms"`
}

type obterRecursosResp struct {
	Servicos       []servicoRecursoResp `json:"servicos"`
	ColetadoEmUnix int64                `json:"coletado_em_unix"`
}

// ObterRecursos serve GET /api/v1/monitor/recursos: CPU/memória vêm do
// metrics-server (metrics.k8s.io), RPS/taxa de sucesso/latência vêm do
// Prometheus do linkerd-viz — nenhum serviço precisa se auto-instrumentar, o
// sidecar do mesh já captura tudo.
//
// Falha ao falar com o metrics-server (cluster fora do ar) vira um único erro
// HTTP de tela; falha só do Prometheus não derruba a resposta (o Client já
// degrada normalmente — ver meshmetrics.Client.ObterRecursos).
func ObterRecursos(client recursosClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		servicos, err := client.ObterRecursos(r.Context())
		if err != nil {
			web.Error(w, http.StatusBadGateway, "RECURSOS_INDISPONIVEIS", "não foi possível consultar os recursos do cluster")
			return
		}

		resp := obterRecursosResp{
			Servicos:       make([]servicoRecursoResp, 0, len(servicos)),
			ColetadoEmUnix: meshmetrics.ColetadoEm(),
		}
		for _, s := range servicos {
			resp.Servicos = append(resp.Servicos, servicoRecursoResp{
				Nome:                  s.Nome,
				Status:                s.Status,
				CPUMillicores:         s.CPUMillicores,
				MemoriaBytes:          s.MemoriaBytes,
				RequisicoesPorSegundo: s.RequisicoesPorSegundo,
				TaxaSucessoPercent:    s.TaxaSucessoPercent,
				LatenciaP99Ms:         s.LatenciaP99Ms,
			})
		}

		web.JSON(w, http.StatusOK, resp)
	}
}
