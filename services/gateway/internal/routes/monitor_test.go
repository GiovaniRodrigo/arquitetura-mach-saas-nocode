package routes

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/machv4/platform/services/gateway/internal/meshmetrics"
)

// fakeRecursosClient implementa recursosClient para os testes de
// ObterRecursos — a implementação real (meshmetrics.Client) fala HTTP de
// verdade com o metrics-server/Prometheus do cluster.
type fakeRecursosClient struct {
	servicos []meshmetrics.ServicoRecurso
	err      error
}

func (f *fakeRecursosClient) ObterRecursos(context.Context) ([]meshmetrics.ServicoRecurso, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.servicos, nil
}

func TestObterRecursos_OK(t *testing.T) {
	fake := &fakeRecursosClient{servicos: []meshmetrics.ServicoRecurso{
		{
			Nome: "iam", Status: "servindo",
			CPUMillicores: 12, MemoriaBytes: 15728640,
			RequisicoesPorSegundo: 0.3, TaxaSucessoPercent: 100, LatenciaP99Ms: 1.2,
		},
		{Nome: "logic", Status: "indisponivel"},
	}}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/monitor/recursos", nil)
	ObterRecursos(fake)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d (%s)", rec.Code, rec.Body.String())
	}

	var corpo struct {
		Servicos []struct {
			Nome                  string  `json:"nome"`
			Status                string  `json:"status"`
			CPUMillicores         int64   `json:"cpu_millicores"`
			MemoriaBytes          int64   `json:"memoria_bytes"`
			RequisicoesPorSegundo float64 `json:"requisicoes_por_segundo"`
			TaxaSucessoPercent    float64 `json:"taxa_sucesso_percent"`
			LatenciaP99Ms         float64 `json:"latencia_p99_ms"`
		} `json:"servicos"`
		ColetadoEmUnix int64 `json:"coletado_em_unix"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &corpo); err != nil {
		t.Fatalf("json: %v", err)
	}

	if len(corpo.Servicos) != 2 {
		t.Fatalf("esperava 2 serviços; got %d", len(corpo.Servicos))
	}
	if corpo.ColetadoEmUnix == 0 {
		t.Fatalf("coletado_em_unix não preenchido")
	}

	iam := corpo.Servicos[0]
	if iam.Nome != "iam" || iam.Status != "servindo" || iam.CPUMillicores != 12 ||
		iam.MemoriaBytes != 15728640 || iam.TaxaSucessoPercent != 100 {
		t.Fatalf("serviço iam inesperado: %+v", iam)
	}

	logic := corpo.Servicos[1]
	if logic.Nome != "logic" || logic.Status != "indisponivel" {
		t.Fatalf("serviço logic inesperado: %+v", logic)
	}
}

// Um erro do client (cluster/metrics-server fora do ar) vira um único erro de
// tela — nunca 8 "cards" individuais quebrados.
func TestObterRecursos_ErroDoClient_502(t *testing.T) {
	fake := &fakeRecursosClient{err: errors.New("metrics-server indisponível")}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/monitor/recursos", nil)
	ObterRecursos(fake)(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("esperava 502; got %d (%s)", rec.Code, rec.Body.String())
	}

	var corpo struct {
		Codigo   string `json:"codigo"`
		Mensagem string `json:"mensagem"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &corpo); err != nil {
		t.Fatalf("json: %v", err)
	}
	if corpo.Codigo != "RECURSOS_INDISPONIVEIS" {
		t.Fatalf("codigo inesperado: %q", corpo.Codigo)
	}
}
