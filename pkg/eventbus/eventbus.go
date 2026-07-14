// Package eventbus define o contrato partilhado da mensageria assíncrona
// (RabbitMQ): nomes de exchanges/filas, cabeçalhos, tipos de evento e o
// roteamento por tenant (fair queuing, RN09). É consumido tanto pelo publisher
// no Logic Engine quanto pelos workers, mantendo o contrato numa fonte única.
package eventbus

import (
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Topologia AMQP (ver infra/rabbitmq/definitions.json).
const (
	Exchange = "eventos" // topic — recebe os eventos publicados
)

// Cabeçalhos de propagação. O traceparent (W3C) é injetado/extraído pelo
// propagador OTel; os demais carregam a identidade do tenant (RNF02, RNF04).
const (
	HeaderTenantID    = "x-tenant-id"
	HeaderBlindIndex  = "x-component-blind-index"
	HeaderTentativas  = "x-tentativas"
	HeaderTraceparent = "traceparent"
)

// Tipos de evento — também o prefixo da routing key.
const (
	TipoWebhook     = "webhook.disparo"
	TipoNotificacao = "notificacao.envio"
)

// Filas principais e respetivas DLQs.
const (
	FilaWebhooks     = "webhooks.disparo"
	FilaNotificacoes = "notificacoes.envio"
	SufixoDLQ        = ".dlq"
)

// RoutingKey monta `<tipo>.<tenant_id>` — a fila liga com `<tipo>.*`, e o tenant
// no fim permite distribuição justa entre tenants (RN09).
func RoutingKey(tipo, tenantID string) string {
	return tipo + "." + tenantID
}

// FilaDe mapeia o tipo de evento à fila principal, ou "" se desconhecido.
func FilaDe(tipo string) string {
	switch tipo {
	case TipoWebhook:
		return FilaWebhooks
	case TipoNotificacao:
		return FilaNotificacoes
	default:
		return ""
	}
}

// FilaRetry é a fila de re-tentativa (TTL + dead-letter de volta à principal).
func FilaRetry(fila string) string { return fila + ".retry" }

// FilaDLQ é a fila morta terminal da fila principal.
func FilaDLQ(fila string) string { return fila + SufixoDLQ }

// Corpo é a mensagem de negócio publicada (a identidade viaja nos headers).
type Corpo struct {
	Tipo    string          `json:"tipo"`
	Payload json.RawMessage `json:"payload"`
}

// HeaderCarrier adapta um amqp.Table ao propagador OTel (traceparent W3C).
type HeaderCarrier amqp.Table

// Get implementa propagation.TextMapCarrier.
func (c HeaderCarrier) Get(key string) string {
	if v, ok := c[key].(string); ok {
		return v
	}
	return ""
}

// Set implementa propagation.TextMapCarrier.
func (c HeaderCarrier) Set(key, value string) { c[key] = value }

// Keys implementa propagation.TextMapCarrier.
func (c HeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}
