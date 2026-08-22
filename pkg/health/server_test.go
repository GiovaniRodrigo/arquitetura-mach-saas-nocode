package health_test

import (
	"context"
	"testing"
	"time"

	healthv1 "github.com/machv4/platform/gen/go/construtor/health/v1"
	"github.com/machv4/platform/pkg/health"
)

func TestObterStatus_ServicoServindo(t *testing.T) {
	inicio := time.Now().Add(-5 * time.Second)
	srv := health.NewServer("iam", inicio)

	got, err := srv.ObterStatus(context.Background(), &healthv1.ObterStatusRequest{})
	if err != nil {
		t.Fatalf("ObterStatus retornou erro: %v", err)
	}

	if got.Nome != "iam" {
		t.Errorf("Nome = %q, want %q", got.Nome, "iam")
	}
	if got.Tipo != "grpc" {
		t.Errorf("Tipo = %q, want %q", got.Tipo, "grpc")
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
	if got.MensagemErro != "" {
		t.Errorf("MensagemErro = %q, want empty", got.MensagemErro)
	}
}
