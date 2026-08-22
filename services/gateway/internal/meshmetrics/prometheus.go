package meshmetrics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// MeshStat é o sinal dourado (golden signal) de um serviço, capturado pelo
// sidecar Linkerd e agregado no Prometheus do linkerd-viz — sem nenhuma
// instrumentação própria do serviço.
type MeshStat struct {
	RequisicoesPorSegundo float64
	TaxaSucessoPercent    float64
	LatenciaP99Ms         float64
}

// PrometheusClient consulta o Prometheus instalado pelo linkerd-viz
// (prometheus.linkerd-viz.svc.cluster.local:9090), alcançável via DNS interno
// do cluster a partir de qualquer pod — inclusive o Gateway.
type PrometheusClient struct {
	http      *http.Client
	baseURL   string
	namespace string
}

func NewPrometheusClient(baseURL, namespace string) *PrometheusClient {
	return &PrometheusClient{
		http:      &http.Client{Timeout: 5 * time.Second},
		baseURL:   baseURL,
		namespace: namespace,
	}
}

type promResponse struct {
	Data struct {
		Result []struct {
			Metric map[string]string `json:"metric"`
			Value  [2]any            `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

func (c *PrometheusClient) query(ctx context.Context, promQL string) (*promResponse, error) {
	u := c.baseURL + "/api/v1/query?" + url.Values{"query": {promQL}}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("meshmetrics: prometheus respondeu %d", resp.StatusCode)
	}
	var out promResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

// MeshStats devolve RPS/sucesso/latência por serviço (chave = label
// "deployment" do Linkerd, igual ao nome do Deployment em infra/k8s/).
// Serviços sem tráfego na janela de amostragem simplesmente não aparecem no
// mapa — quem chama trata a ausência como "0 requisições", não como erro.
func (c *PrometheusClient) MeshStats(ctx context.Context) (map[string]MeshStat, error) {
	out := make(map[string]MeshStat)

	rps, err := c.query(ctx, fmt.Sprintf(
		`sum(rate(request_total{namespace="%s",direction="inbound"}[30s])) by (deployment)`, c.namespace))
	if err != nil {
		return nil, err
	}
	for _, r := range rps.Data.Result {
		dep := r.Metric["deployment"]
		v, _ := strconv.ParseFloat(r.Value[1].(string), 64)
		stat := out[dep]
		stat.RequisicoesPorSegundo = v
		out[dep] = stat
	}

	sucesso, err := c.query(ctx, fmt.Sprintf(
		`sum(rate(response_total{namespace="%s",direction="inbound",classification="success"}[30s])) by (deployment) `+
			`/ sum(rate(response_total{namespace="%s",direction="inbound"}[30s])) by (deployment) * 100`,
		c.namespace, c.namespace))
	if err != nil {
		return nil, err
	}
	for _, r := range sucesso.Data.Result {
		dep := r.Metric["deployment"]
		v, _ := strconv.ParseFloat(r.Value[1].(string), 64)
		stat := out[dep]
		stat.TaxaSucessoPercent = v
		out[dep] = stat
	}

	for dep := range out {
		latencia, err := c.query(ctx, fmt.Sprintf(
			`histogram_quantile(0.99, sum(rate(response_latency_ms_bucket{namespace="%s",direction="inbound",deployment="%s"}[30s])) by (le))`,
			c.namespace, dep))
		if err != nil {
			continue // latência é um extra — não derruba a coleta dos demais (RN01 da spec 008 original, mesmo espírito)
		}
		if len(latencia.Data.Result) == 0 {
			continue
		}
		v, _ := strconv.ParseFloat(latencia.Data.Result[0].Value[1].(string), 64)
		stat := out[dep]
		stat.LatenciaP99Ms = v
		out[dep] = stat
	}

	return out, nil
}
