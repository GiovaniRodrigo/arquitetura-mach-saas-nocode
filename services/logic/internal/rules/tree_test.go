package rules

import (
	"errors"
	"testing"
)

func TestValidar_Invalidas(t *testing.T) {
	casos := map[string]*No{
		"raiz nil":            nil,
		"nó desconhecido":     {No: "loop"},
		"ação sem tipo":       {No: NoAcao},
		"operador inválido":   {No: NoCondicao, Operador: "xor", BlindIndex: "bi", Entao: &No{No: NoAcao, Tipo: "x"}},
		"condição sem bi":     {No: NoCondicao, Operador: OpIgual, Entao: &No{No: NoAcao, Tipo: "x"}},
		"condição sem entao":  {No: NoCondicao, Operador: OpIgual, BlindIndex: "bi"},
		"filho inválido":      {No: NoCondicao, Operador: OpIgual, BlindIndex: "bi", Entao: &No{No: "??"}},
	}
	for nome, arvore := range casos {
		t.Run(nome, func(t *testing.T) {
			if err := Validar(arvore); !errors.Is(err, ErrArvoreInvalida) {
				t.Fatalf("esperava ErrArvoreInvalida para %q; got %v", nome, err)
			}
		})
	}
}

func TestParse_JSONMalformado(t *testing.T) {
	if _, err := Parse([]byte(`{nao-json`)); !errors.Is(err, ErrArvoreInvalida) {
		t.Fatalf("esperava ErrArvoreInvalida; got %v", err)
	}
}

func TestAvaliar_RamosEAcao(t *testing.T) {
	// SE idade >= 18 ENTÃO acao(aprovar) SENÃO acao(recusar)
	arvore := &No{
		No: NoCondicao, Operador: OpMaiorIgual, BlindIndex: "bi-idade", Valor: "18",
		Entao: &No{No: NoAcao, Tipo: "aprovar"},
		Senao: &No{No: NoAcao, Tipo: "recusar"},
	}

	ac, err := Avaliar(arvore, map[string]string{"bi-idade": "21"})
	if err != nil {
		t.Fatalf("avaliar: %v", err)
	}
	if ac == nil || ac.Tipo != "aprovar" {
		t.Fatalf("esperava aprovar; got %+v", ac)
	}

	ac, err = Avaliar(arvore, map[string]string{"bi-idade": "16"})
	if err != nil {
		t.Fatalf("avaliar: %v", err)
	}
	if ac == nil || ac.Tipo != "recusar" {
		t.Fatalf("esperava recusar; got %+v", ac)
	}
}

func TestAvaliar_SenaoAusente_SemAcao(t *testing.T) {
	arvore := &No{
		No: NoCondicao, Operador: OpIgual, BlindIndex: "bi-uf", Valor: "BA",
		Entao: &No{No: NoAcao, Tipo: "webhook.disparo"},
	}
	ac, err := Avaliar(arvore, map[string]string{"bi-uf": "SP"})
	if err != nil {
		t.Fatalf("avaliar: %v", err)
	}
	if ac != nil {
		t.Fatalf("condição falsa sem 'senao' deveria não disparar ação; got %+v", ac)
	}
}

func TestComparar_NaoNumericoEmRelacional(t *testing.T) {
	// Comparação relacional com texto não numérico → falso, sem erro.
	ok, err := comparar("abc", OpMaior, "10")
	if err != nil || ok {
		t.Fatalf("esperava (false,nil); got (%v,%v)", ok, err)
	}
}
