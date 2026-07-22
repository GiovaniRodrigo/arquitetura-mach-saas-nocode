//go:build integration

// Verifica o publisher contra um RabbitMQ real com as definitions.json aplicadas:
// a mensagem é roteada pelo exchange x-consistent-hash a um dos shards de
// webhooks.disparo (RN09) e os cabeçalhos (tenant, blind_index, traceparent)
// chegam intactos (RF08, RNF04). Executar:
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

	// Limpa todos os shards antes de publicar (isolamento entre execuções).
	for _, fila := range eventbus.FilasDe(eventbus.TipoWebhook) {
		_, _ = ch.QueuePurge(fila, false)
	}

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

	// A mensagem deve ter sido roteada a algum shard de webhooks.disparo pelo
	// exchange x-consistent-hash.
	msg := getDeAlgumaFila(t, ch, eventbus.FilasDe(eventbus.TipoWebhook))

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

// getDeAlgumaFila poll a lista de filas-shard até achar uma mensagem em alguma
// delas — o shard exato depende do hash da routing key, decidido pelo broker.
func getDeAlgumaFila(t *testing.T, ch *amqp.Channel, filas []string) amqp.Delivery {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		for _, fila := range filas {
			msg, ok, err := ch.Get(fila, true)
			if err != nil {
				t.Fatalf("get %s: %v", fila, err)
			}
			if ok {
				return msg
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("nenhuma mensagem em %v no prazo", filas)
	return amqp.Delivery{}
}

// TestPublicar_MesmoTenantSempreNoMesmoShard confirma o determinismo do hash: N
// mensagens do mesmo tenant caem sempre no mesmo shard (RN09) — a base do fair
// queuing, que a topologia anterior (fila única) não garantia.
func TestPublicar_MesmoTenantSempreNoMesmoShard(t *testing.T) {
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

	filas := eventbus.FilasDe(eventbus.TipoWebhook)
	for _, fila := range filas {
		_, _ = ch.QueuePurge(fila, false)
	}

	ctx := tenantctx.NewContext(context.Background(), &commonv1.TenantContext{TenantId: "t-shard-fixo", Tipo: "dono"})
	p := NewPublisher(ch)
	const n = 10
	for i := 0; i < n; i++ {
		if err := p.Publicar(ctx, Evento{Tipo: eventbus.TipoWebhook, Payload: json.RawMessage(`{}`)}); err != nil {
			t.Fatalf("publicar #%d: %v", i, err)
		}
	}

	shardsComMensagens := map[string]bool{}
	total := 0
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) && total < n {
		for _, fila := range filas {
			_, ok, err := ch.Get(fila, true)
			if err != nil {
				t.Fatalf("get %s: %v", fila, err)
			}
			if ok {
				total++
				shardsComMensagens[fila] = true
			}
		}
		if total < n {
			time.Sleep(50 * time.Millisecond)
		}
	}
	if total != n {
		t.Fatalf("esperava %d mensagens no total; achei %d", n, total)
	}
	if len(shardsComMensagens) != 1 {
		t.Fatalf("mesmo tenant deveria cair sempre no mesmo shard; mensagens espalhadas em %v", shardsComMensagens)
	}
}
