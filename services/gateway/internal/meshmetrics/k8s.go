// Package meshmetrics lê recursos/telemetria dos 8 serviços da plataforma
// diretamente da infraestrutura Kubernetes + Linkerd (spec 009: migração para
// Kubernetes + service mesh), substituindo o polling gRPC customizado
// desenhado originalmente na spec 008. Duas fontes:
//   - metrics.k8s.io (metrics-server): CPU/memória reais do pod.
//   - Prometheus do linkerd-viz: RPS/taxa de sucesso/latência do sidecar,
//     capturados automaticamente pelo mesh, sem instrumentar cada serviço.
package meshmetrics

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	serviceAccountDir = "/var/run/secrets/kubernetes.io/serviceaccount"
	kubeAPIBase       = "https://kubernetes.default.svc"
)

// RecursoPod é o uso de CPU/memória de um serviço, lido do metrics-server.
type RecursoPod struct {
	CPUMillicores int64
	MemoriaBytes  int64
}

// K8sClient consulta metrics.k8s.io usando as credenciais in-cluster
// (ServiceAccount montado automaticamente em todo pod — RBAC concedido em
// infra/k8s/05-gateway-rbac.yaml).
type K8sClient struct {
	http      *http.Client
	token     string
	namespace string
}

// NewK8sClient monta o client a partir do ServiceAccount montado no pod. Fora
// de um cluster (ex.: testes unitários), use NewK8sClientFromToken.
func NewK8sClient(namespace string) (*K8sClient, error) {
	tokenBytes, err := os.ReadFile(serviceAccountDir + "/token")
	if err != nil {
		return nil, fmt.Errorf("meshmetrics: token do ServiceAccount: %w", err)
	}
	caBytes, err := os.ReadFile(serviceAccountDir + "/ca.crt")
	if err != nil {
		return nil, fmt.Errorf("meshmetrics: CA do ServiceAccount: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caBytes) {
		return nil, fmt.Errorf("meshmetrics: CA do ServiceAccount inválida")
	}
	return &K8sClient{
		http: &http.Client{
			Timeout:   5 * time.Second,
			Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}},
		},
		token:     strings.TrimSpace(string(tokenBytes)),
		namespace: namespace,
	}, nil
}

type podMetricsList struct {
	Items []struct {
		Metadata struct {
			Name   string            `json:"name"`
			Labels map[string]string `json:"labels"`
		} `json:"metadata"`
		Containers []struct {
			Name  string `json:"name"`
			Usage struct {
				CPU    string `json:"cpu"`
				Memory string `json:"memory"`
			} `json:"usage"`
		} `json:"containers"`
	} `json:"items"`
}

// PodResourceUsage devolve o uso de CPU/memória por serviço (chave = valor da
// label "app", que nos manifests em infra/k8s/ é sempre igual ao nome do
// container de negócio — exclui o container linkerd-proxy do sidecar, que
// não é o processo monitorado).
func (c *K8sClient) PodResourceUsage(ctx context.Context) (map[string]RecursoPod, error) {
	url := fmt.Sprintf("%s/apis/metrics.k8s.io/v1beta1/namespaces/%s/pods", kubeAPIBase, c.namespace)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("meshmetrics: metrics.k8s.io respondeu %d", resp.StatusCode)
	}

	var list podMetricsList
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, err
	}

	out := make(map[string]RecursoPod)
	for _, item := range list.Items {
		app := item.Metadata.Labels["app"]
		if app == "" {
			continue
		}
		for _, ctr := range item.Containers {
			if ctr.Name != app {
				continue // pula o sidecar linkerd-proxy/linkerd-init
			}
			cpu, _ := parseCPU(ctr.Usage.CPU)
			mem, _ := parseMemoria(ctr.Usage.Memory)
			out[app] = RecursoPod{CPUMillicores: cpu, MemoriaBytes: mem}
		}
	}
	return out, nil
}

// parseCPU converte a quantidade do metrics-server (ex.: "3346240n", "12m",
// "1") para millicores.
func parseCPU(v string) (int64, error) {
	switch {
	case strings.HasSuffix(v, "n"):
		n, err := strconv.ParseInt(strings.TrimSuffix(v, "n"), 10, 64)
		return n / 1_000_000, err // nanocores -> millicores
	case strings.HasSuffix(v, "u"):
		n, err := strconv.ParseInt(strings.TrimSuffix(v, "u"), 10, 64)
		return n / 1_000, err // microcores -> millicores
	case strings.HasSuffix(v, "m"):
		n, err := strconv.ParseInt(strings.TrimSuffix(v, "m"), 10, 64)
		return n, err
	default:
		n, err := strconv.ParseFloat(v, 64)
		return int64(n * 1000), err // cores -> millicores
	}
}

// parseMemoria converte a quantidade do metrics-server (ex.: "170440Ki",
// "512Mi") para bytes.
func parseMemoria(v string) (int64, error) {
	units := map[string]int64{
		"Ki": 1024, "Mi": 1024 * 1024, "Gi": 1024 * 1024 * 1024,
		"K": 1000, "M": 1_000_000, "G": 1_000_000_000,
	}
	for suffix, mult := range units {
		if strings.HasSuffix(v, suffix) {
			n, err := strconv.ParseInt(strings.TrimSuffix(v, suffix), 10, 64)
			return n * mult, err
		}
	}
	return strconv.ParseInt(v, 10, 64)
}
