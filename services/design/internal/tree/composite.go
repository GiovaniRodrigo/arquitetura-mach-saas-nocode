// Package tree modela a árvore recursiva de UI (padrão Composite) do Design
// Engine, com validação estrutural e (de)serialização para a coluna JSONB
// `designs.arvore` (RF01).
package tree

import (
	"encoding/json"
	"errors"
	"fmt"

	designv1 "github.com/machv4/platform/gen/go/construtor/design/v1"
)

// ErrArvoreInvalida indica que a árvore Composite viola uma invariante estrutural.
// O servidor gRPC a traduz em InvalidArgument e o Gateway em 422 INVALID_TREE.
var ErrArvoreInvalida = errors.New("tree: árvore estruturalmente inválida (RF01)")

// MaxProfundidade limita a recursão para barrar árvores patológicas (proteção
// contra stack overflow e payloads maliciosos).
const MaxProfundidade = 64

// Validar percorre a árvore garantindo que: existe raiz; todo nó tem blind_index
// e tipo; nenhum blind_index se repete (o índice é a identidade do componente); as
// propriedades, quando presentes, são JSON válido; e a profundidade é limitada.
func Validar(raiz *designv1.Componente) error {
	if raiz == nil {
		return fmt.Errorf("%w: árvore vazia", ErrArvoreInvalida)
	}
	vistos := make(map[string]struct{})
	return validarNo(raiz, vistos, 1)
}

func validarNo(n *designv1.Componente, vistos map[string]struct{}, prof int) error {
	if prof > MaxProfundidade {
		return fmt.Errorf("%w: profundidade excede %d níveis", ErrArvoreInvalida, MaxProfundidade)
	}
	if n.GetBlindIndex() == "" {
		return fmt.Errorf("%w: componente sem blind_index", ErrArvoreInvalida)
	}
	if n.GetTipo() == "" {
		return fmt.Errorf("%w: componente %q sem tipo", ErrArvoreInvalida, n.GetBlindIndex())
	}
	if _, dup := vistos[n.GetBlindIndex()]; dup {
		return fmt.Errorf("%w: blind_index duplicado %q", ErrArvoreInvalida, n.GetBlindIndex())
	}
	vistos[n.GetBlindIndex()] = struct{}{}

	if p := n.GetPropriedades(); len(p) > 0 && !json.Valid(p) {
		return fmt.Errorf("%w: propriedades de %q não são JSON válido", ErrArvoreInvalida, n.GetBlindIndex())
	}

	for _, f := range n.GetComponenteFilhos() {
		if err := validarNo(f, vistos, prof+1); err != nil {
			return err
		}
	}
	return nil
}

// no espelha Componente para (de)serialização legível na coluna arvore (jsonb).
// Diferente de protojson, mantém `propriedades` como objeto JSON aninhado em vez
// de string base64, deixando a coluna consultável por operadores jsonb.
type no struct {
	BlindIndex   string          `json:"blind_index"`
	Tipo         string          `json:"tipo"`
	Propriedades json.RawMessage `json:"propriedades,omitempty"`
	Filhos       []no            `json:"componente_filhos,omitempty"`
}

// Serializar converte a árvore Composite no JSON gravado em designs.arvore.
func Serializar(raiz *designv1.Componente) ([]byte, error) {
	return json.Marshal(paraNo(raiz))
}

func paraNo(c *designv1.Componente) no {
	n := no{BlindIndex: c.GetBlindIndex(), Tipo: c.GetTipo()}
	if p := c.GetPropriedades(); len(p) > 0 {
		n.Propriedades = json.RawMessage(p)
	}
	for _, f := range c.GetComponenteFilhos() {
		n.Filhos = append(n.Filhos, paraNo(f))
	}
	return n
}

// Desserializar reconstrói a árvore Composite a partir do JSONB persistido.
func Desserializar(raw []byte) (*designv1.Componente, error) {
	var n no
	if err := json.Unmarshal(raw, &n); err != nil {
		return nil, fmt.Errorf("tree: desserializar arvore: %w", err)
	}
	return deNo(n), nil
}

func deNo(n no) *designv1.Componente {
	c := &designv1.Componente{BlindIndex: n.BlindIndex, Tipo: n.Tipo}
	if len(n.Propriedades) > 0 {
		c.Propriedades = []byte(n.Propriedades)
	}
	for _, f := range n.Filhos {
		c.ComponenteFilhos = append(c.ComponenteFilhos, deNo(f))
	}
	return c
}
