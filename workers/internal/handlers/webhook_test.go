package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/machv4/platform/pkg/eventbus"
)

func evento(url string) eventbus.Corpo {
	payload, _ := json.Marshal(map[string]any{
		"url_destino": url,
		"corpo":       map[string]string{"ola": "mundo"},
	})
	return eventbus.Corpo{Tipo: eventbus.TipoWebhook, Payload: payload}
}

func TestWebhook_2xx_Sucesso(t *testing.T) {
	var recebido map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&recebido)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	if err := NewWebhook(srv.Client()).Handle(context.Background(), evento(srv.URL)); err != nil {
		t.Fatalf("2xx deveria ter sucesso: %v", err)
	}
	if recebido["ola"] != "mundo" {
		t.Fatalf("corpo não entregue: %v", recebido)
	}
}

func TestWebhook_5xx_Falha(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	if err := NewWebhook(srv.Client()).Handle(context.Background(), evento(srv.URL)); err == nil {
		t.Fatal("5xx deveria falhar (retry/DLQ)")
	}
}

func TestWebhook_SemURL_Falha(t *testing.T) {
	if err := NewWebhook(nil).Handle(context.Background(), evento("")); err == nil {
		t.Fatal("url ausente deveria falhar")
	}
}
