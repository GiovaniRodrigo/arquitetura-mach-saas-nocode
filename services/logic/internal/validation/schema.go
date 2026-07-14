// Package validation revalida, no servidor, o payload de um formulário contra o
// schema salvo (campos_definicao), independentemente do que o cliente aceitou
// (RN08). O mapa de erros é sempre indexado pelo blind_index do componente e as
// mensagens jamais expõem nomes reais de colunas ou tabelas (RNF08, critério 2).
package validation

import (
	"strconv"
	"time"
)

// Tipos de campo suportados.
const (
	TipoString = "string"
	TipoNumber = "number"
	TipoDate   = "date"
	TipoBool   = "bool"
)

// Limites são as restrições opcionais de um campo (coluna limites, jsonb).
type Limites struct {
	Min       *float64 `json:"min,omitempty"`
	Max       *float64 `json:"max,omitempty"`
	MinLength *int     `json:"min_length,omitempty"`
	MaxLength *int     `json:"max_length,omitempty"`
}

// CampoDef é a definição de um campo, identificado apenas pelo blind_index.
type CampoDef struct {
	BlindIndex  string
	Tipo        string
	Obrigatorio bool
	Limites     Limites
}

// Mensagens genéricas — nunca revelam nome real de campo/coluna (RNF08).
const (
	msgDesconhecido = "campo não pertence ao schema"
	msgObrigatorio  = "campo obrigatório ausente"
	msgTipoNumber   = "valor não é um número válido"
	msgTipoDate     = "valor não é uma data válida (AAAA-MM-DD)"
	msgTipoBool     = "valor não é booleano (true/false)"
	msgMin          = "valor abaixo do mínimo permitido"
	msgMax          = "valor acima do máximo permitido"
	msgMinLength    = "texto mais curto que o mínimo permitido"
	msgMaxLength    = "texto mais longo que o máximo permitido"
)

// Validar confronta os dados submetidos com o schema e devolve os erros
// indexados por blind_index. Mapa vazio ⇒ payload válido.
//
// Regras: (1) todo blind_index submetido deve existir no schema — chaves
// desconhecidas são rejeitadas, barrando submissões que contornam o cliente
// (critério 2); (2) campos obrigatórios ausentes/vazios falham; (3) valores
// presentes são validados por tipo e limites.
func Validar(schema map[string]CampoDef, dados map[string]string) map[string]string {
	erros := make(map[string]string)

	// (1) Rejeita qualquer chave fora do schema.
	for bi := range dados {
		if _, ok := schema[bi]; !ok {
			erros[bi] = msgDesconhecido
		}
	}

	for bi, campo := range schema {
		valor, presente := dados[bi]
		if !presente || valor == "" {
			if campo.Obrigatorio {
				erros[bi] = msgObrigatorio
			}
			continue // sem valor a validar
		}
		if msg := validarValor(campo, valor); msg != "" {
			erros[bi] = msg
		}
	}

	return erros
}

func validarValor(campo CampoDef, valor string) string {
	switch campo.Tipo {
	case TipoNumber:
		n, err := strconv.ParseFloat(valor, 64)
		if err != nil {
			return msgTipoNumber
		}
		if campo.Limites.Min != nil && n < *campo.Limites.Min {
			return msgMin
		}
		if campo.Limites.Max != nil && n > *campo.Limites.Max {
			return msgMax
		}
	case TipoDate:
		if _, err := time.Parse("2006-01-02", valor); err != nil {
			return msgTipoDate
		}
	case TipoBool:
		if valor != "true" && valor != "false" {
			return msgTipoBool
		}
	default: // string e tipos textuais
		if campo.Limites.MinLength != nil && len(valor) < *campo.Limites.MinLength {
			return msgMinLength
		}
		if campo.Limites.MaxLength != nil && len(valor) > *campo.Limites.MaxLength {
			return msgMaxLength
		}
	}
	return ""
}
