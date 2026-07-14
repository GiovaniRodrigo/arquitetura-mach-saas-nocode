package dlq

import (
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/machv4/platform/pkg/eventbus"
)

func TestTentativas(t *testing.T) {
	if n := Tentativas(amqp.Table{}); n != 0 {
		t.Fatalf("sem header deveria ser 0; got %d", n)
	}
	if n := Tentativas(amqp.Table{eventbus.HeaderTentativas: int32(2)}); n != 2 {
		t.Fatalf("int32 deveria ler 2; got %d", n)
	}
}

func TestEsgotouEProximos(t *testing.T) {
	h := amqp.Table{eventbus.HeaderTentativas: 1}
	if Esgotou(h, 3) {
		t.Fatal("1 < 3 não deveria esgotar")
	}
	h = ProximosHeaders(ProximosHeaders(h)) // 1 -> 3
	if Tentativas(h) != 3 {
		t.Fatalf("esperava 3 tentativas; got %d", Tentativas(h))
	}
	if !Esgotou(h, 3) {
		t.Fatal("3 >= 3 deveria esgotar")
	}
}

func TestProximosHeaders_NaoMutaOriginal(t *testing.T) {
	orig := amqp.Table{eventbus.HeaderTenantID: "t-1"}
	novos := ProximosHeaders(orig)
	if _, existe := orig[eventbus.HeaderTentativas]; existe {
		t.Fatal("a cópia não deveria mutar os headers originais")
	}
	if novos[eventbus.HeaderTenantID] != "t-1" {
		t.Fatal("a cópia deveria preservar os demais headers")
	}
}

func TestAlertaDe_IsolaTenant(t *testing.T) {
	h := amqp.Table{eventbus.HeaderTenantID: "t-afetado"}
	a := AlertaDe(h, eventbus.TipoWebhook, "timeout")
	if a.TenantID != "t-afetado" || a.Tipo != eventbus.TipoWebhook || a.Motivo != "timeout" {
		t.Fatalf("alerta inesperado: %+v", a)
	}
}
