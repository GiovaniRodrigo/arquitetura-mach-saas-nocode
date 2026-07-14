package tree

import (
	"errors"
	"strconv"
	"testing"

	designv1 "github.com/machv4/platform/gen/go/construtor/design/v1"
)

func comp(bi, tipo string, filhos ...*designv1.Componente) *designv1.Componente {
	return &designv1.Componente{BlindIndex: bi, Tipo: tipo, ComponenteFilhos: filhos}
}

func TestValidar_ArvoreValida(t *testing.T) {
	raiz := comp("bi-root", "container",
		comp("bi-a", "input_texto"),
		comp("bi-b", "botao"),
	)
	raiz.ComponenteFilhos[0].Propriedades = []byte(`{"label":"Nome"}`)

	if err := Validar(raiz); err != nil {
		t.Fatalf("árvore válida rejeitada: %v", err)
	}
}

func TestValidar_Invalidas(t *testing.T) {
	casos := map[string]*designv1.Componente{
		"raiz nil":            nil,
		"sem blind_index":     comp("", "container"),
		"sem tipo":            comp("bi-x", ""),
		"filho sem tipo":      comp("bi-root", "container", comp("bi-a", "")),
		"propriedades quebrad": func() *designv1.Componente {
			c := comp("bi-root", "container")
			c.Propriedades = []byte(`{nao-e-json`)
			return c
		}(),
		"blind_index duplicado": comp("bi-root", "container",
			comp("bi-dup", "input_texto"),
			comp("bi-dup", "botao"),
		),
	}

	for nome, arvore := range casos {
		t.Run(nome, func(t *testing.T) {
			err := Validar(arvore)
			if err == nil {
				t.Fatalf("esperava erro para %q", nome)
			}
			if !errors.Is(err, ErrArvoreInvalida) {
				t.Fatalf("erro deveria envolver ErrArvoreInvalida; got %v", err)
			}
		})
	}
}

func TestValidar_ProfundidadeExcedida(t *testing.T) {
	// Cadeia linear com MaxProfundidade+1 níveis.
	raiz := comp("bi-0", "container")
	atual := raiz
	for i := 1; i <= MaxProfundidade; i++ {
		filho := comp("bi-"+strconv.Itoa(i), "container")
		atual.ComponenteFilhos = []*designv1.Componente{filho}
		atual = filho
	}
	if err := Validar(raiz); !errors.Is(err, ErrArvoreInvalida) {
		t.Fatalf("esperava ErrArvoreInvalida por profundidade; got %v", err)
	}
}

func TestSerializar_RoundTrip(t *testing.T) {
	orig := comp("bi-root", "container",
		comp("bi-a", "input_texto"),
		comp("bi-b", "grupo", comp("bi-c", "botao")),
	)
	orig.ComponenteFilhos[0].Propriedades = []byte(`{"placeholder":"email"}`)

	raw, err := Serializar(orig)
	if err != nil {
		t.Fatalf("serializar: %v", err)
	}

	got, err := Desserializar(raw)
	if err != nil {
		t.Fatalf("desserializar: %v", err)
	}

	if got.GetBlindIndex() != "bi-root" || len(got.GetComponenteFilhos()) != 2 {
		t.Fatalf("estrutura raiz não preservada: %+v", got)
	}
	neto := got.GetComponenteFilhos()[1].GetComponenteFilhos()
	if len(neto) != 1 || neto[0].GetBlindIndex() != "bi-c" || neto[0].GetTipo() != "botao" {
		t.Fatalf("recursão não preservada: %+v", neto)
	}
	if string(got.GetComponenteFilhos()[0].GetPropriedades()) != `{"placeholder":"email"}` {
		t.Fatalf("propriedades não preservadas: %s", got.GetComponenteFilhos()[0].GetPropriedades())
	}
}
