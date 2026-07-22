package events

import (
	"context"
	"encoding/json"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	commonv1 "github.com/machv4/platform/gen/go/construtor/common/v1"
	"github.com/machv4/platform/pkg/eventbus"
	"github.com/machv4/platform/pkg/tenantctx"
)

type canalFake struct {
	exchange string
	key      string
	msg      amqp.Publishing
	chamado  bool
}

func (c *canalFake) PublishWithContext(_ context.Context, exchange, key string, _, _ bool, msg amqp.Publishing) error {
	c.exchange, c.key, c.msg, c.chamado = exchange, key, msg, true
	return nil
}

func comTenant() context.Context {
	return tenantctx.NewContext(context.Background(), &commonv1.TenantContext{TenantId: "t-abc", Tipo: "dono"})
}

func TestPublicar_SemTenant(t *testing.T) {
	p := NewPublisher(&canalFake{})
	err := p.Publicar(context.Background(), Evento{Tipo: eventbus.TipoWebhook})
	if err == nil {
		t.Fatal("esperava erro sem tenant no contexto")
	}
}

func TestPublicar_RoutingKeyEHeaders(t *testing.T) {
	fake := &canalFake{}
	p := NewPublisher(fake)

	err := p.Publicar(comTenant(), Evento{
		Tipo:                eventbus.TipoWebhook,
		ComponentBlindIndex: "bi-123",
		Payload:             json.RawMessage(`{"url_destino":"https://x"}`),
	})
	if err != nil {
		t.Fatalf("publicar: %v", err)
	}

	if !fake.chamado {
		t.Fatal("deveria ter publicado")
	}
	if fake.exchange != eventbus.ExchangeDe(eventbus.TipoWebhook) {
		t.Fatalf("exchange inesperado: %q", fake.exchange)
	}
	if fake.key != "webhook.disparo.t-abc" {
		t.Fatalf("routing key deveria conter o tenant (RN09); got %q", fake.key)
	}
	if fake.msg.Headers[eventbus.HeaderTenantID] != "t-abc" {
		t.Fatalf("header tenant ausente: %v", fake.msg.Headers)
	}
	if fake.msg.Headers[eventbus.HeaderBlindIndex] != "bi-123" {
		t.Fatalf("header blind_index ausente: %v", fake.msg.Headers)
	}

	var corpo eventbus.Corpo
	if err := json.Unmarshal(fake.msg.Body, &corpo); err != nil {
		t.Fatalf("corpo inválido: %v", err)
	}
	if corpo.Tipo != eventbus.TipoWebhook {
		t.Fatalf("tipo no corpo inesperado: %q", corpo.Tipo)
	}
}

func TestPublicar_InjetaTraceparent(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})

	traceID, _ := trace.TraceIDFromHex("0123456789abcdef0123456789abcdef")
	spanID, _ := trace.SpanIDFromHex("0123456789abcdef")
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceID, SpanID: spanID, TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithSpanContext(comTenant(), sc)

	fake := &canalFake{}
	if err := NewPublisher(fake).Publicar(ctx, Evento{Tipo: eventbus.TipoWebhook}); err != nil {
		t.Fatalf("publicar: %v", err)
	}
	tp, _ := fake.msg.Headers[eventbus.HeaderTraceparent].(string)
	if tp == "" {
		t.Fatalf("traceparent deveria estar nos headers (RNF04); got %v", fake.msg.Headers)
	}
}
