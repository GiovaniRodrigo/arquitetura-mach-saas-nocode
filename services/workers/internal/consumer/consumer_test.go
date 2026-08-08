package consumer

import (
	"context"
	"errors"
	"testing"

	"github.com/machv4/platform/pkg/eventbus"
	"github.com/machv4/platform/services/workers/internal/handlers"
)

func roteador(err error, max int) *Roteador {
	h := handlers.HandlerFunc(func(context.Context, eventbus.Corpo) error { return err })
	return NewRoteador(map[string]handlers.Handler{eventbus.TipoWebhook: h}, max)
}

func TestProcessar_Sucesso_Ack(t *testing.T) {
	d, err := roteador(nil, 3).Processar(context.Background(), eventbus.Corpo{Tipo: eventbus.TipoWebhook}, 0)
	if d != Ack || err != nil {
		t.Fatalf("esperava Ack; got %v err=%v", d, err)
	}
}

func TestProcessar_TipoDesconhecido_DeadLetter(t *testing.T) {
	d, _ := roteador(nil, 3).Processar(context.Background(), eventbus.Corpo{Tipo: "desconhecido"}, 0)
	if d != DeadLetter {
		t.Fatalf("tipo sem handler deveria ir à DLQ; got %v", d)
	}
}

func TestProcessar_FalhaComTentativasRestantes_Retry(t *testing.T) {
	d, err := roteador(errors.New("boom"), 3).Processar(context.Background(), eventbus.Corpo{Tipo: eventbus.TipoWebhook}, 0)
	if d != Retry || err == nil {
		t.Fatalf("esperava Retry; got %v err=%v", d, err)
	}
}

func TestProcessar_FalhaEsgotada_DeadLetter(t *testing.T) {
	// tentativas=2, +1 = 3 >= max(3) → DLQ
	d, err := roteador(errors.New("boom"), 3).Processar(context.Background(), eventbus.Corpo{Tipo: eventbus.TipoWebhook}, 2)
	if d != DeadLetter || err == nil {
		t.Fatalf("esperava DeadLetter na exaustão; got %v err=%v", d, err)
	}
}
