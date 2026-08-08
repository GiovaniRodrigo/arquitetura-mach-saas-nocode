// Package dlq decide o roteamento de falhas contínuas para a Dead-Letter Queue e
// produz o alerta ao tenant afetado, mantendo a falha de um tenant isolada dos
// demais (RN09, RNF06).
package dlq

import (
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/machv4/platform/pkg/eventbus"
)

// Tentativas lê o número de falhas já acumuladas no cabeçalho x-tentativas.
func Tentativas(headers amqp.Table) int {
	switch v := headers[eventbus.HeaderTentativas].(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	default:
		return 0
	}
}

// Esgotou indica que já se atingiu o teto de tentativas (vai para a DLQ).
func Esgotou(headers amqp.Table, maxTentativas int) bool {
	return Tentativas(headers) >= maxTentativas
}

// ProximosHeaders devolve uma cópia dos cabeçalhos com x-tentativas incrementado,
// para a re-publicação da mensagem numa nova tentativa.
func ProximosHeaders(headers amqp.Table) amqp.Table {
	novos := amqp.Table{}
	for k, v := range headers {
		novos[k] = v
	}
	novos[eventbus.HeaderTentativas] = Tentativas(headers) + 1
	return novos
}

// Alerta é a notificação de esgotamento enviada ao tenant afetado (RNF06).
type Alerta struct {
	TenantID string `json:"tenant_id"`
	Tipo     string `json:"tipo"`
	Motivo   string `json:"motivo"`
}

// AlertaDe monta o alerta a partir da entrega falhada. O tenant vem do cabeçalho,
// garantindo que o alerta atinge apenas o tenant dono da mensagem (isolamento).
func AlertaDe(headers amqp.Table, tipo, motivo string) Alerta {
	tenant, _ := headers[eventbus.HeaderTenantID].(string)
	return Alerta{TenantID: tenant, Tipo: tipo, Motivo: motivo}
}
