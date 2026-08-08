// Package consumer roteia cada mensagem ao handler do seu tipo e decide o
// desfecho (ack, retry ou dead-letter), isolando essa lógica do transporte AMQP
// para a tornar testável (RF08, RN09, RNF06).
package consumer

import (
	"context"

	"github.com/machv4/platform/pkg/eventbus"
	"github.com/machv4/platform/services/workers/internal/handlers"
)

// Desfecho é a decisão do consumidor sobre uma mensagem.
type Desfecho int

const (
	// Ack — processada com sucesso.
	Ack Desfecho = iota
	// Retry — falhou, mas ainda há tentativas: re-publicar com backoff.
	Retry
	// DeadLetter — tipo desconhecido ou tentativas esgotadas: enviar à DLQ.
	DeadLetter
)

// Roteador mapeia tipos de evento a handlers e aplica o teto de tentativas.
type Roteador struct {
	handlers      map[string]handlers.Handler
	maxTentativas int
}

// NewRoteador cria o roteador. maxTentativas é o teto antes da DLQ (RNF06).
func NewRoteador(hs map[string]handlers.Handler, maxTentativas int) *Roteador {
	return &Roteador{handlers: hs, maxTentativas: maxTentativas}
}

// Processar executa o handler do tipo e devolve o desfecho. `tentativas` é o
// número de falhas já acumuladas (0 na primeira entrega). Um tipo sem handler
// vai direto à DLQ (mensagem inválida, não re-tentável).
func (r *Roteador) Processar(ctx context.Context, corpo eventbus.Corpo, tentativas int) (Desfecho, error) {
	h, ok := r.handlers[corpo.Tipo]
	if !ok {
		return DeadLetter, nil
	}

	if err := h.Handle(ctx, corpo); err != nil {
		if tentativas+1 >= r.maxTentativas {
			return DeadLetter, err
		}
		return Retry, err
	}
	return Ack, nil
}
