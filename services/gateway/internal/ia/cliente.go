// Package ia encapsula a chamada ao modelo de linguagem (Anthropic Claude)
// que alimenta o Assistente de Design — a especificidade de "expert em
// design de sistemas" vem do prompt de sistema + do contexto recuperado via
// RAG (pacote rag), nunca de fine-tuning.
package ia

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/machv4/platform/services/gateway/internal/rag"
)

// ErrNaoConfigurado é devolvido quando não há ANTHROPIC_API_KEY configurada
// — o Gateway sobe mesmo assim (mesmo padrão de degradação graciosa já
// usado para meshmetrics), e a rota HTTP traduz isso para 502.
var ErrNaoConfigurado = errors.New("assistente de ia não configurado (ANTHROPIC_API_KEY ausente)")

// Mensagem é um turno da conversa trocado com o assistente.
type Mensagem struct {
	Papel    string // "usuario" ou "assistente"
	Conteudo string
}

// PedidoResposta agrupa o que o Assistente de Design precisa para responder:
// o histórico da conversa, o nome do sistema em foco (quando houver) e os
// documentos de RAG já recuperados para a última pergunta do usuário.
type PedidoResposta struct {
	SistemaNome string
	Historico   []Mensagem
	Contexto    []rag.Documento
}

// Cliente é o subconjunto usado pela rota HTTP — permite injetar um dublê
// nos testes sem chamar a API da Anthropic de verdade.
type Cliente interface {
	Responder(ctx context.Context, pedido PedidoResposta) (string, error)
}

const modeloPadrao = "claude-opus-5"

const promptSistema = `Você é o Assistente de Design do MAYS (Make Your SaaS), um construtor no-code de sistemas SaaS multi-tenant. Seu papel é atuar como um especialista sênior em design e arquitetura de sistemas, ajudando o usuário a estruturar o sistema que está montando no builder: modelagem de entidades, estrutura de telas, regras de negócio, multi-tenancy, segurança e versionamento.

Responda em português do Brasil, de forma direta e prática, sempre relacionando a recomendação ao foco/descrição que o usuário deu do sistema. Quando útil, aponte em qual aba do builder (Sistemas, Telas, Regras de Negócio, Versão) a recomendação se aplica. Use o CONTEXTO DE REFERÊNCIA abaixo quando for relevante à pergunta, mas não cite os documentos como fontes — incorpore o conhecimento na resposta como se fosse seu. Se a pergunta não tiver relação com design/arquitetura de sistemas, redirecione educadamente o usuário para esse tema.`

// AnthropicCliente implementa Cliente usando a Messages API da Anthropic.
type AnthropicCliente struct {
	sdk    anthropic.Client
	modelo string
}

// NovoAnthropicCliente cria o cliente. `apiKey` vazio ainda constrói o
// cliente (a chamada subsequente falhará com erro de autenticação) — quem
// decide se a IA fica desabilitada por completo é o cmd/main.go, que nesse
// caso passa `nil` como ia.Cliente para o router (ver ChatIA).
func NovoAnthropicCliente(apiKey string) *AnthropicCliente {
	var opts []option.RequestOption
	if apiKey != "" {
		opts = append(opts, option.WithAPIKey(apiKey))
	}
	return &AnthropicCliente{
		sdk:    anthropic.NewClient(opts...),
		modelo: modeloPadrao,
	}
}

// Responder monta o prompt (system + contexto RAG + histórico) e chama a
// Messages API.
func (c *AnthropicCliente) Responder(ctx context.Context, pedido PedidoResposta) (string, error) {
	if c == nil {
		return "", ErrNaoConfigurado
	}

	system := montarPromptSistema(pedido)

	mensagens := make([]anthropic.MessageParam, 0, len(pedido.Historico))
	for _, m := range pedido.Historico {
		bloco := anthropic.NewTextBlock(m.Conteudo)
		if m.Papel == "assistente" {
			mensagens = append(mensagens, anthropic.NewAssistantMessage(bloco))
		} else {
			mensagens = append(mensagens, anthropic.NewUserMessage(bloco))
		}
	}
	if len(mensagens) == 0 {
		return "", errors.New("histórico vazio: nenhuma mensagem para responder")
	}

	resp, err := c.sdk.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(c.modelo),
		MaxTokens: 1536,
		System:    []anthropic.TextBlockParam{{Text: system}},
		Messages:  mensagens,
	})
	if err != nil {
		return "", fmt.Errorf("anthropic: %w", err)
	}

	var sb strings.Builder
	for _, bloco := range resp.Content {
		if txt, ok := bloco.AsAny().(anthropic.TextBlock); ok {
			sb.WriteString(txt.Text)
		}
	}
	resposta := strings.TrimSpace(sb.String())
	if resposta == "" {
		return "", errors.New("anthropic: resposta vazia")
	}
	return resposta, nil
}

func montarPromptSistema(pedido PedidoResposta) string {
	var sb strings.Builder
	sb.WriteString(promptSistema)
	if pedido.SistemaNome != "" {
		fmt.Fprintf(&sb, "\n\nSistema atualmente em foco pelo usuário: %q.", pedido.SistemaNome)
	}
	if len(pedido.Contexto) > 0 {
		sb.WriteString("\n\nCONTEXTO DE REFERÊNCIA (boas práticas de design de sistemas):")
		for _, doc := range pedido.Contexto {
			sb.WriteString("\n---\n")
			sb.WriteString(doc.Conteudo)
		}
	}
	return sb.String()
}
