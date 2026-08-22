// Package routes: rota do Assistente de Design (chat de IA/RAG, spec
// chat-ia-rag). Ao contrário das demais rotas, não fala gRPC com nenhum
// Engine interno — o "backend" aqui é a base de conhecimento embutida
// (pkg rag) + a API da Anthropic (pkg ia), consultadas direto do Gateway.
package routes

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/machv4/platform/services/gateway/internal/ia"
	"github.com/machv4/platform/services/gateway/internal/rag"
	"github.com/machv4/platform/services/gateway/internal/web"
)

// documentosPorPergunta limita quantos documentos de RAG entram no prompt —
// o suficiente para dar contexto sem inflar o custo/latência de cada turno.
const documentosPorPergunta = 3

// historicoMaximo evita que o cliente envie uma conversa sem limite prático
// no corpo da requisição — o histórico é mantido no frontend (sessionStorage
// do chat), não persistido no servidor.
const historicoMaximo = 40

type mensagemChatDTO struct {
	Papel    string `json:"papel"`
	Conteudo string `json:"conteudo"`
}

type chatIARequest struct {
	SistemaNome string            `json:"sistema_nome"`
	Historico   []mensagemChatDTO `json:"historico"`
}

// ChatIA serve POST /api/v1/ia/chat: recebe o histórico da conversa (a
// última mensagem é sempre do usuário) e devolve a resposta do Assistente
// de Design. RAG: recupera documentos relevantes à última pergunta do
// usuário (pkg rag) e injeta como contexto no prompt de sistema (pkg ia)
// antes de consultar o modelo.
func ChatIA(cliente ia.Cliente) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body chatIARequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			web.Error(w, http.StatusBadRequest, "INVALID_BODY", "corpo JSON inválido")
			return
		}
		if len(body.Historico) == 0 {
			web.Error(w, http.StatusBadRequest, "MISSING_PARAM", "historico obrigatório (ao menos a mensagem do usuário)")
			return
		}
		if len(body.Historico) > historicoMaximo {
			body.Historico = body.Historico[len(body.Historico)-historicoMaximo:]
		}

		ultimaPergunta := body.Historico[len(body.Historico)-1].Conteudo
		if ultimaPergunta == "" {
			web.Error(w, http.StatusBadRequest, "MISSING_PARAM", "última mensagem do histórico não pode ser vazia")
			return
		}

		historico := make([]ia.Mensagem, len(body.Historico))
		for i, m := range body.Historico {
			historico[i] = ia.Mensagem{Papel: m.Papel, Conteudo: m.Conteudo}
		}

		resposta, err := responder(r.Context(), cliente, ia.PedidoResposta{
			SistemaNome: body.SistemaNome,
			Historico:   historico,
			Contexto:    rag.Recuperar(ultimaPergunta, documentosPorPergunta),
		})
		if err != nil {
			writeChatIAError(w, err)
			return
		}
		web.JSON(w, http.StatusOK, map[string]string{"resposta": resposta})
	}
}

// responder isola a checagem de "cliente ausente" (Gateway subiu sem
// ANTHROPIC_API_KEY) do erro de chamada em si, para que writeChatIAError
// possa distingui-los com a mesma mensagem amigável em ambos os casos.
func responder(ctx context.Context, cliente ia.Cliente, pedido ia.PedidoResposta) (string, error) {
	if cliente == nil {
		return "", ia.ErrNaoConfigurado
	}
	return cliente.Responder(ctx, pedido)
}

func writeChatIAError(w http.ResponseWriter, err error) {
	if errors.Is(err, ia.ErrNaoConfigurado) {
		web.Error(w, http.StatusBadGateway, "IA_INDISPONIVEL", "assistente de ia não está configurado neste ambiente")
		return
	}
	web.Error(w, http.StatusBadGateway, "IA_INDISPONIVEL", "não foi possível obter resposta do assistente agora")
}
