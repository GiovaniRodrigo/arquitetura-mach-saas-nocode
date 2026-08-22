package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/machv4/platform/services/gateway/internal/ia"
)

// clienteIAFalso é um dublê de ia.Cliente para testar a rota sem chamar a
// API da Anthropic de verdade.
type clienteIAFalso struct {
	resposta string
	err      error
	recebido ia.PedidoResposta
}

func (c *clienteIAFalso) Responder(_ context.Context, pedido ia.PedidoResposta) (string, error) {
	c.recebido = pedido
	if c.err != nil {
		return "", c.err
	}
	return c.resposta, nil
}

func doChatIA(t *testing.T, handler http.HandlerFunc, body any) *httptest.ResponseRecorder {
	t.Helper()
	buf, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ia/chat", bytes.NewReader(buf))
	rec := httptest.NewRecorder()
	handler(rec, req)
	return rec
}

func TestChatIA_RespondeComSucesso(t *testing.T) {
	falso := &clienteIAFalso{resposta: "recomendação de arquitetura"}
	rec := doChatIA(t, ChatIA(falso), chatIARequest{
		SistemaNome: "Clínica Fácil",
		Historico: []mensagemChatDTO{
			{Papel: "usuario", Conteudo: "como isolar dados entre clínicas?"},
		},
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, corpo = %s", rec.Code, rec.Body.String())
	}
	var out map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out["resposta"] != "recomendação de arquitetura" {
		t.Errorf("resposta = %q", out["resposta"])
	}
	if falso.recebido.SistemaNome != "Clínica Fácil" {
		t.Errorf("sistema_nome não propagado: %q", falso.recebido.SistemaNome)
	}
	if len(falso.recebido.Contexto) == 0 {
		t.Error("esperava contexto de RAG recuperado para a pergunta")
	}
}

func TestChatIA_HistoricoVazio(t *testing.T) {
	rec := doChatIA(t, ChatIA(&clienteIAFalso{}), chatIARequest{Historico: nil})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, esperava 400", rec.Code)
	}
}

func TestChatIA_UltimaMensagemVazia(t *testing.T) {
	rec := doChatIA(t, ChatIA(&clienteIAFalso{}), chatIARequest{
		Historico: []mensagemChatDTO{{Papel: "usuario", Conteudo: ""}},
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, esperava 400", rec.Code)
	}
}

func TestChatIA_ClienteNaoConfigurado(t *testing.T) {
	rec := doChatIA(t, ChatIA(nil), chatIARequest{
		Historico: []mensagemChatDTO{{Papel: "usuario", Conteudo: "oi"}},
	})
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, esperava 502", rec.Code)
	}
}

func TestChatIA_ErroDoModelo(t *testing.T) {
	rec := doChatIA(t, ChatIA(&clienteIAFalso{err: errors.New("timeout")}), chatIARequest{
		Historico: []mensagemChatDTO{{Papel: "usuario", Conteudo: "oi"}},
	})
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, esperava 502", rec.Code)
	}
}
