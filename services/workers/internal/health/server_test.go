package health_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/machv4/platform/services/workers/internal/health"
)

type respostaHealth struct {
	Status              string `json:"status"`
	UptimeSegundos      int64  `json:"uptime_segundos"`
	MemoriaAlocadaBytes int64  `json:"memoria_alocada_bytes"`
	MemoriaSistemaBytes int64  `json:"memoria_sistema_bytes"`
	Goroutines          int32  `json:"goroutines"`
}

func TestNovoHandler_GetHealth_Retorna200ComCorpoEsperado(t *testing.T) {
	inicio := time.Now().Add(-5 * time.Second)
	handler := health.NovoHandler(inicio)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" && ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var got respostaHealth
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("corpo não é JSON válido: %v (body=%s)", err, rec.Body.String())
	}

	if got.Status != "servindo" {
		t.Errorf("Status = %q, want %q", got.Status, "servindo")
	}
	if got.UptimeSegundos < 5 {
		t.Errorf("UptimeSegundos = %d, want >= 5", got.UptimeSegundos)
	}
	if got.MemoriaAlocadaBytes <= 0 {
		t.Errorf("MemoriaAlocadaBytes = %d, want > 0", got.MemoriaAlocadaBytes)
	}
	if got.MemoriaSistemaBytes <= 0 {
		t.Errorf("MemoriaSistemaBytes = %d, want > 0", got.MemoriaSistemaBytes)
	}
	if got.Goroutines <= 0 {
		t.Errorf("Goroutines = %d, want > 0", got.Goroutines)
	}
}
