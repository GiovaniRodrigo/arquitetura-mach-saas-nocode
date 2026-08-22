package meshmetrics

import (
	"context"
	"errors"
	"time"
)

// ErrIndisponivel é devolvido quando o Client não foi montado (ex.: o
// Gateway não conseguiu ler o ServiceAccount do cluster no boot).
var ErrIndisponivel = errors.New("meshmetrics: client não inicializado")

// ServicoRecurso é o estado consolidado de um serviço monitorado — a fusão de
// RecursoPod (metrics-server) e MeshStat (Prometheus/linkerd-viz) para um dos
// 8 serviços fixos da plataforma (spec 008/009).
type ServicoRecurso struct {
	Nome                  string
	Status                string // "servindo" | "indisponivel"
	CPUMillicores         int64
	MemoriaBytes          int64
	RequisicoesPorSegundo float64
	TaxaSucessoPercent    float64
	LatenciaP99Ms         float64
}

// Servicos é a lista fixa dos 8 serviços da plataforma monitorados pela tela
// Monitor de Recursos — mesmos nomes dos Deployments em infra/k8s/04-services.yaml.
var Servicos = []string{"iam", "design", "logic", "deploy", "export", "workers", "collab", "gateway"}

// Client agrega as duas fontes de telemetria do cluster.
type Client struct {
	k8s        *K8sClient
	prometheus *PrometheusClient
}

func NewClient(k8s *K8sClient, prometheus *PrometheusClient) *Client {
	return &Client{k8s: k8s, prometheus: prometheus}
}

// ObterRecursos consulta metrics-server e Prometheus em paralelo (RNF01 da
// spec 008 original, mesmo espírito) e funde os dois por nome de serviço. Um
// serviço sem entrada no metrics-server (pod não existe/não roda) vira
// status "indisponivel" — a ausência de tráfego recente no Prometheus, por si
// só, NÃO torna um serviço indisponível (pode estar ocioso e saudável).
func (c *Client) ObterRecursos(ctx context.Context) ([]ServicoRecurso, error) {
	if c == nil {
		return nil, ErrIndisponivel
	}
	type podResult struct {
		usage map[string]RecursoPod
		err   error
	}
	type meshResult struct {
		stats map[string]MeshStat
		err   error
	}
	podCh := make(chan podResult, 1)
	meshCh := make(chan meshResult, 1)

	go func() {
		u, err := c.k8s.PodResourceUsage(ctx)
		podCh <- podResult{u, err}
	}()
	go func() {
		s, err := c.prometheus.MeshStats(ctx)
		meshCh <- meshResult{s, err}
	}()

	pods := <-podCh
	mesh := <-meshCh
	if pods.err != nil {
		return nil, pods.err
	}
	// Prometheus fora do ar não derruba a tela inteira (RNF02 original): os
	// serviços aparecem com CPU/memória e sem RPS/latência/sucesso.
	meshStats := mesh.stats
	if meshStats == nil {
		meshStats = map[string]MeshStat{}
	}

	out := make([]ServicoRecurso, 0, len(Servicos))
	for _, nome := range Servicos {
		usage, ok := pods.usage[nome]
		if !ok {
			out = append(out, ServicoRecurso{Nome: nome, Status: "indisponivel"})
			continue
		}
		stat := meshStats[nome]
		out = append(out, ServicoRecurso{
			Nome:                  nome,
			Status:                "servindo",
			CPUMillicores:         usage.CPUMillicores,
			MemoriaBytes:          usage.MemoriaBytes,
			RequisicoesPorSegundo: stat.RequisicoesPorSegundo,
			TaxaSucessoPercent:    stat.TaxaSucessoPercent,
			LatenciaP99Ms:         stat.LatenciaP99Ms,
		})
	}
	return out, nil
}

// ColetadoEm é usado pelo handler HTTP para o campo coletado_em_unix.
func ColetadoEm() int64 { return time.Now().Unix() }
