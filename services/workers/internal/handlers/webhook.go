package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/machv4/platform/pkg/eventbus"
)

// webhookPayload é o corpo esperado de um evento webhook.disparo.
type webhookPayload struct {
	URLDestino string          `json:"url_destino"`
	Corpo      json.RawMessage `json:"corpo"`
}

// Webhook entrega webhooks a integrações de terceiros via HTTP POST.
type Webhook struct {
	client *http.Client
}

// NewWebhook cria o handler com o cliente HTTP informado (injetável em teste).
func NewWebhook(client *http.Client) *Webhook {
	if client == nil {
		client = http.DefaultClient
	}
	return &Webhook{client: client}
}

// Handle faz POST do corpo ao destino. Respostas fora de 2xx são falha (retry/DLQ).
func (h *Webhook) Handle(ctx context.Context, corpo eventbus.Corpo) error {
	var p webhookPayload
	if err := json.Unmarshal(corpo.Payload, &p); err != nil {
		return fmt.Errorf("webhook: payload inválido: %w", err)
	}
	if p.URLDestino == "" {
		return fmt.Errorf("webhook: url_destino ausente")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.URLDestino, bytes.NewReader(p.Corpo))
	if err != nil {
		return fmt.Errorf("webhook: montar requisição: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook: entrega falhou: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: destino respondeu %d", resp.StatusCode)
	}
	return nil
}
