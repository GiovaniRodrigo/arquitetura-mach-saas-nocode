// Package events publica tarefas assíncronas (webhooks, notificações) do Logic
// Engine no RabbitMQ, desacopladas do fluxo síncrono (RF08). Cada mensagem leva
// o tenant_id, o component_blind_index e o traceparent nos cabeçalhos (RNF02,
// RNF04); a routing key inclui o tenant para fair queuing (RN09).
package events

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"

	"github.com/machv4/platform/pkg/eventbus"
	"github.com/machv4/platform/pkg/tenantctx"
)

// Canal é o subconjunto de *amqp.Channel usado pelo publisher (facilita testes).
type Canal interface {
	PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
}

// Evento é uma tarefa a publicar. ComponentBlindIndex vai no cabeçalho (RNF04),
// não no corpo.
type Evento struct {
	Tipo                string
	ComponentBlindIndex string
	Payload             json.RawMessage
}

// Publisher publica eventos no exchange partilhado.
type Publisher struct {
	ch Canal
}

// NewPublisher cria um Publisher sobre um canal AMQP.
func NewPublisher(ch Canal) *Publisher { return &Publisher{ch: ch} }

// Publicar envia o evento. O tenant vem do contexto (nunca do corpo); a routing
// key é `<tipo>.<tenant_id>` (RN09) e o traceparent é injetado do contexto OTel.
func (p *Publisher) Publicar(ctx context.Context, ev Evento) error {
	tid := tenantctx.TenantID(ctx)
	if tid == "" {
		return fmt.Errorf("events: publicação sem tenant no contexto (RN01)")
	}

	body, err := json.Marshal(eventbus.Corpo{Tipo: ev.Tipo, Payload: ev.Payload})
	if err != nil {
		return fmt.Errorf("events: serializar evento: %w", err)
	}

	headers := amqp.Table{eventbus.HeaderTenantID: tid}
	if ev.ComponentBlindIndex != "" {
		headers[eventbus.HeaderBlindIndex] = ev.ComponentBlindIndex
	}
	// Injeta o traceparent (W3C) a partir do contexto — trace contínuo até o worker.
	otel.GetTextMapPropagator().Inject(ctx, eventbus.HeaderCarrier(headers))

	return p.ch.PublishWithContext(ctx, eventbus.ExchangeDe(ev.Tipo), eventbus.RoutingKey(ev.Tipo, tid), false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Headers:      headers,
		Body:         body,
	})
}
