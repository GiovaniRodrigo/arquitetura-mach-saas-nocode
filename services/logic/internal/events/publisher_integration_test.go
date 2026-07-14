//go:build integration

// Verifica o publisher contra um RabbitMQ real com as definitions.json aplicadas:
// a routing key `webhook.disparo.<tenant>` casa com o binding `webhook.disparo.*`
// da fila e os cabeçalhos (tenant, blind_index, traceparent) chegam intactos
// (RF08, RN09, RNF04). Executar:
//
//	RABBITMQ_URL=amqp://mach:mach@localhost:5673/ \
//	  go test -tags integration ./services/logic/internal/events/...
package events

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	commonv1 "github.com/machv4/platform/gen/go/construtor/common/v1"
	"github.com/machv4/platform/pkg/eventbus"
	"github.com/machv4/platform/pkg/tenantctx"
)

func amqpURL() string {
	if v := os.Getenv("RABBITMQ_URL"); v != "" {
		return v
	}
	return "amqp://mach:mach@localhost:5673/"
}

func TestPublicar_ChegaNaFilaComHeaders(t *testing.T) {
	conn, err := amqp.Dial(amqpURL())
	if err != nil {
		t.Skipf("RabbitMQ indisponível (%v)", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		t.Fatalf("canal: %v", err)
	}
	defer ch.Close()

	// Limpa a fila antes de publicar (isolamento entre execuções).
	_, _ = ch.QueuePurge(eventbus.FilaWebhooks, false)

	otel.SetTextMapPropagator(propagation.TraceContext{})
	traceID, _ := trace.TraceIDFromHex("0123456789abcdef0123456789abcdef")
	spanID, _ := trace.SpanIDFromHex("0123456789abcdef")
	sc := trace.NewSpanContext(trace.SpanContextConfig{TraceID: traceID, SpanID: spanID, TraceFlags: trace.FlagsSampled})

	ctx := trace.ContextWithSpanContext(
		tenantctx.NewContext(context.Background(), &commonv1.TenantContext{TenantId: "t-int", Tipo: "dono"}),
		sc,
	)

	err = NewPublisher(ch).Publicar(ctx, Evento{
		Tipo:                eventbus.TipoWebhook,
		ComponentBlindIndex: "bi-int",
		Payload:             json.RawMessage(`{"url_destino":"https://exemplo"}`),
	})
	if err != nil {
		t.Fatalf("publicar: %v", err)
	}

	// A mensagem deve ter sido roteada para webhooks.disparo pelo binding.
	msg := getComRetry(t, ch, eventbus.FilaWebhooks)

	if msg.Headers[eventbus.HeaderTenantID] != "t-int" {
		t.Fatalf("header tenant ausente/errado: %v", msg.Headers)
	}
	if msg.Headers[eventbus.HeaderBlindIndex] != "bi-int" {
		t.Fatalf("header blind_index ausente: %v", msg.Headers)
	}
	tp, _ := msg.Headers[eventbus.HeaderTraceparent].(string)
	if tp == "" {
		t.Fatalf("traceparent ausente (RNF04): %v", msg.Headers)
	}

	var corpo eventbus.Corpo
	if err := json.Unmarshal(msg.Body, &corpo); err != nil || corpo.Tipo != eventbus.TipoWebhook {
		t.Fatalf("corpo inesperado: %s (err=%v)", msg.Body, err)
	}
}

func getComRetry(t *testing.T, ch *amqp.Channel, fila string) amqp.Delivery {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		msg, ok, err := ch.Get(fila, true)
		if err != nil {
			t.Fatalf("get %s: %v", fila, err)
		}
		if ok {
			return msg
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("nenhuma mensagem em %s no prazo", fila)
	return amqp.Delivery{}
}
