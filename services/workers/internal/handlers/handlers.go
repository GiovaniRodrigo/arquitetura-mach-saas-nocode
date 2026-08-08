// Package handlers implementa o processamento de cada tipo de evento assíncrono
// (webhooks, notificações) — a lógica desacoplada do fluxo síncrono (RF08).
package handlers

import (
	"context"

	"github.com/machv4/platform/pkg/eventbus"
)

// Handler processa um evento. Um erro sinaliza falha de entrega (o consumidor
// decide retry/DLQ); nil = sucesso (ack).
type Handler interface {
	Handle(ctx context.Context, corpo eventbus.Corpo) error
}

// HandlerFunc adapta uma função a Handler.
type HandlerFunc func(ctx context.Context, corpo eventbus.Corpo) error

// Handle implementa Handler.
func (f HandlerFunc) Handle(ctx context.Context, corpo eventbus.Corpo) error {
	return f(ctx, corpo)
}
