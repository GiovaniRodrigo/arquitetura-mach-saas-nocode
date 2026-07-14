package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/machv4/platform/pkg/eventbus"
)

// notificacaoPayload é o corpo esperado de um evento notificacao.envio.
type notificacaoPayload struct {
	Canal       string `json:"canal"`       // email | sms | push
	Destinatario string `json:"destinatario"`
	Mensagem    string `json:"mensagem"`
}

// Sink entrega a notificação ao canal externo (email/SMS/push). Abstraído para
// permitir teste e futura integração com provedores reais.
type Sink interface {
	Enviar(ctx context.Context, canal, destinatario, mensagem string) error
}

// Notificacao processa eventos de notificação.
type Notificacao struct {
	sink Sink
}

// NewNotificacao cria o handler. Sem sink, usa um sink de log (dev/local).
func NewNotificacao(sink Sink) *Notificacao {
	if sink == nil {
		sink = logSink{}
	}
	return &Notificacao{sink: sink}
}

// Handle valida o payload e delega ao sink configurado.
func (h *Notificacao) Handle(ctx context.Context, corpo eventbus.Corpo) error {
	var p notificacaoPayload
	if err := json.Unmarshal(corpo.Payload, &p); err != nil {
		return fmt.Errorf("notificacao: payload inválido: %w", err)
	}
	if p.Destinatario == "" || p.Canal == "" {
		return fmt.Errorf("notificacao: canal/destinatario ausentes")
	}
	return h.sink.Enviar(ctx, p.Canal, p.Destinatario, p.Mensagem)
}

type logSink struct{}

func (logSink) Enviar(_ context.Context, canal, destinatario, _ string) error {
	log.Printf("notificacao: enviada via %s para %s", canal, destinatario)
	return nil
}
