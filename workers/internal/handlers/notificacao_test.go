package handlers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/machv4/platform/pkg/eventbus"
)

type sinkFake struct {
	canal, destinatario, mensagem string
	chamado                       bool
	err                           error
}

func (s *sinkFake) Enviar(_ context.Context, canal, destinatario, mensagem string) error {
	s.canal, s.destinatario, s.mensagem, s.chamado = canal, destinatario, mensagem, true
	return s.err
}

func notificacao(m map[string]any) eventbus.Corpo {
	p, _ := json.Marshal(m)
	return eventbus.Corpo{Tipo: eventbus.TipoNotificacao, Payload: p}
}

func TestNotificacao_Encaminha(t *testing.T) {
	s := &sinkFake{}
	err := NewNotificacao(s).Handle(context.Background(), notificacao(map[string]any{
		"canal": "email", "destinatario": "a@b.com", "mensagem": "oi",
	}))
	if err != nil {
		t.Fatalf("inesperado: %v", err)
	}
	if !s.chamado || s.canal != "email" || s.destinatario != "a@b.com" {
		t.Fatalf("sink não recebeu os dados: %+v", s)
	}
}

func TestNotificacao_PayloadIncompleto_Falha(t *testing.T) {
	if err := NewNotificacao(&sinkFake{}).Handle(context.Background(), notificacao(map[string]any{"mensagem": "x"})); err == nil {
		t.Fatal("faltando canal/destinatario deveria falhar")
	}
}
