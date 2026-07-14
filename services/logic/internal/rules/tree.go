// Package rules modela a árvore de decisão (nós lógicos) das regras de negócio e
// oferece um interpretador que a avalia contra os dados submetidos (RF02).
//
// A árvore é serializada em JSON na coluna regras_negocio.arvore_decisao. Cada nó
// é uma condição (compara o valor de um componente, identificado por blind_index,
// com um literal e ramifica em `entao`/`senao`) ou uma ação (folha) a disparar.
package rules

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ErrArvoreInvalida indica que a árvore de decisão viola uma invariante
// estrutural. O servidor a traduz em InvalidArgument e o Gateway em 422.
var ErrArvoreInvalida = errors.New("rules: árvore de decisão inválida (RF02)")

// MaxProfundidade limita a recursão (proteção contra árvores patológicas).
const MaxProfundidade = 64

// Tipos de nó.
const (
	NoCondicao = "condicao"
	NoAcao     = "acao"
)

// Operadores de condição suportados.
const (
	OpIgual      = "igual"
	OpDiferente  = "diferente"
	OpMaior      = "maior"
	OpMenor      = "menor"
	OpMaiorIgual = "maior_igual"
	OpMenorIgual = "menor_igual"
	OpContem     = "contem"
)

var operadoresValidos = map[string]struct{}{
	OpIgual: {}, OpDiferente: {}, OpMaior: {}, OpMenor: {},
	OpMaiorIgual: {}, OpMenorIgual: {}, OpContem: {},
}

// No é um nó da árvore de decisão. `No` discrimina o tipo; os demais campos
// aplicam-se a condição (Operador/BlindIndex/Valor/Entao/Senao) ou a ação
// (Tipo/Config).
type No struct {
	No string `json:"no"`

	// Condição.
	Operador   string `json:"operador,omitempty"`
	BlindIndex string `json:"blind_index,omitempty"`
	Valor      string `json:"valor,omitempty"`
	Entao      *No    `json:"entao,omitempty"`
	Senao      *No    `json:"senao,omitempty"`

	// Ação (folha).
	Tipo   string          `json:"tipo,omitempty"`
	Config json.RawMessage `json:"config,omitempty"`
}

// Acao é o resultado de uma avaliação que chegou a uma folha.
type Acao struct {
	Tipo   string
	Config json.RawMessage
}

// Parse desserializa a árvore a partir do JSON de arvore_decisao.
func Parse(raw []byte) (*No, error) {
	var n No
	if err := json.Unmarshal(raw, &n); err != nil {
		return nil, fmt.Errorf("%w: JSON malformado: %v", ErrArvoreInvalida, err)
	}
	return &n, nil
}

// Validar garante que a árvore é bem formada: nós reconhecidos, condições com
// operador válido e ramo `entao`, ações com `tipo`, e profundidade limitada.
func Validar(raiz *No) error {
	if raiz == nil {
		return fmt.Errorf("%w: árvore vazia", ErrArvoreInvalida)
	}
	return validarNo(raiz, 1)
}

func validarNo(n *No, prof int) error {
	if prof > MaxProfundidade {
		return fmt.Errorf("%w: profundidade excede %d níveis", ErrArvoreInvalida, MaxProfundidade)
	}
	switch n.No {
	case NoAcao:
		if n.Tipo == "" {
			return fmt.Errorf("%w: ação sem tipo", ErrArvoreInvalida)
		}
		return nil
	case NoCondicao:
		if _, ok := operadoresValidos[n.Operador]; !ok {
			return fmt.Errorf("%w: operador desconhecido %q", ErrArvoreInvalida, n.Operador)
		}
		if n.BlindIndex == "" {
			return fmt.Errorf("%w: condição sem blind_index", ErrArvoreInvalida)
		}
		if n.Entao == nil {
			return fmt.Errorf("%w: condição sem ramo 'entao'", ErrArvoreInvalida)
		}
		if err := validarNo(n.Entao, prof+1); err != nil {
			return err
		}
		if n.Senao != nil {
			return validarNo(n.Senao, prof+1)
		}
		return nil
	default:
		return fmt.Errorf("%w: nó desconhecido %q", ErrArvoreInvalida, n.No)
	}
}

// Avaliar percorre a árvore com os valores submetidos (blind_index → valor) e
// devolve a ação da folha alcançada, ou nil se nenhum ramo levou a uma ação.
func Avaliar(raiz *No, contexto map[string]string) (*Acao, error) {
	if err := Validar(raiz); err != nil {
		return nil, err
	}
	return avaliarNo(raiz, contexto)
}

func avaliarNo(n *No, contexto map[string]string) (*Acao, error) {
	if n == nil {
		return nil, nil
	}
	if n.No == NoAcao {
		return &Acao{Tipo: n.Tipo, Config: n.Config}, nil
	}
	ok, err := comparar(contexto[n.BlindIndex], n.Operador, n.Valor)
	if err != nil {
		return nil, err
	}
	if ok {
		return avaliarNo(n.Entao, contexto)
	}
	return avaliarNo(n.Senao, contexto)
}

// comparar avalia `atual <op> esperado`. Operadores relacionais numéricos exigem
// que ambos os lados sejam numéricos; igualdade e contém operam sobre texto.
func comparar(atual, op, esperado string) (bool, error) {
	switch op {
	case OpIgual:
		return atual == esperado, nil
	case OpDiferente:
		return atual != esperado, nil
	case OpContem:
		return strings.Contains(atual, esperado), nil
	case OpMaior, OpMenor, OpMaiorIgual, OpMenorIgual:
		a, err1 := strconv.ParseFloat(atual, 64)
		e, err2 := strconv.ParseFloat(esperado, 64)
		if err1 != nil || err2 != nil {
			// Valor não numérico numa comparação relacional: condição falsa, não erro
			// — dados de formulário são texto livre e não devem quebrar a regra.
			return false, nil
		}
		switch op {
		case OpMaior:
			return a > e, nil
		case OpMenor:
			return a < e, nil
		case OpMaiorIgual:
			return a >= e, nil
		default:
			return a <= e, nil
		}
	default:
		return false, fmt.Errorf("%w: operador desconhecido %q", ErrArvoreInvalida, op)
	}
}
