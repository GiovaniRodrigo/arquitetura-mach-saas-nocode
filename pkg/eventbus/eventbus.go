// Package eventbus define o contrato partilhado da mensageria assíncrona
// (RabbitMQ): nomes de exchanges/filas, cabeçalhos, tipos de evento e o
// roteamento por tenant. O fair queuing (RN09) é feito por exchanges
// x-consistent-hash que shardeiam cada tipo de evento em NumShards filas fixas
// pelo hash da routing key (`<tipo>.<tenant_id>`) — um tenant cai sempre no mesmo
// shard, então um burst satura no máximo 1/NumShards da capacidade de consumo.
// É consumido tanto pelo publisher no Logic Engine quanto pelos workers, mantendo
// o contrato numa fonte única.
package eventbus

import (
	"encoding/json"
	"strconv"

	amqp "github.com/rabbitmq/amqp091-go"
)

// NumShards é o número de filas-shard por tipo de evento. Cada tenant cai sempre
// no mesmo shard (hash determinístico da routing key pelo RabbitMQ), limitando o
// impacto de um tenant barulhento a no máximo 1/NumShards da capacidade de
// consumo — fair queuing sem proliferação de filas (RN09).
const NumShards = 4

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

// Prefixos das filas principais (cada uma sharded em NumShards filas — ver
// FilasDe) e sufixo comum das DLQs.
const (
	FilaWebhooks     = "webhooks.disparo"
	FilaNotificacoes = "notificacoes.envio"
	SufixoDLQ        = ".dlq"
)

// ExchangeDe mapeia o tipo de evento ao exchange x-consistent-hash que o roteia
// entre as NumShards filas por hash da routing key — cada tenant cai sempre no
// mesmo shard, limitando o impacto de um tenant barulhento (fair queuing, RN09).
// Devolve "" para um tipo desconhecido.
func ExchangeDe(tipo string) string {
	switch tipo {
	case TipoWebhook:
		return "eventos.webhook"
	case TipoNotificacao:
		return "eventos.notificacao"
	default:
		return ""
	}
}

// RoutingKey monta `<tipo>.<tenant_id>` — string hasheada pelo exchange
// x-consistent-hash para decidir o shard (RN09): determinística por tenant.
func RoutingKey(tipo, tenantID string) string {
	return tipo + "." + tenantID
}

// FilaDe mapeia o tipo de evento ao prefixo da fila principal, ou "" se
// desconhecido. Ver FilasDe para os nomes de fila-shard reais a consumir.
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

// FilasDe devolve os NumShards nomes de fila-shard do tipo de evento (ex.:
// "webhooks.disparo.0".."webhooks.disparo.3"), ou nil se o tipo for desconhecido.
func FilasDe(tipo string) []string {
	base := FilaDe(tipo)
	if base == "" {
		return nil
	}
	filas := make([]string, NumShards)
	for i := range filas {
		filas[i] = base + "." + strconv.Itoa(i)
	}
	return filas
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
