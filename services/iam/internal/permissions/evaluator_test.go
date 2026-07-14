package permissions

import "testing"

func TestAvaliar_FailClosed_PorPadrao(t *testing.T) {
	got := Evaluator{}.Avaliar(Sujeito{Papel: "editor"}, nil, []string{"bi-1", "bi-2"})

	if len(got) != 2 {
		t.Fatalf("esperava 2 entradas; got=%d", len(got))
	}
	for bi, d := range got {
		if d.View || d.Click {
			t.Fatalf("%s deveria começar negado; got=%+v", bi, d)
		}
	}
}

func TestAvaliar_ConcedePorPapel(t *testing.T) {
	perms := []Permissao{
		{BlindIndex: "bi-1", Papel: "editor", View: true, Click: true},
		{BlindIndex: "bi-1", Papel: "leitor", View: true, Click: false},
	}
	got := Evaluator{}.Avaliar(Sujeito{Papel: "leitor"}, perms, []string{"bi-1"})

	if !got["bi-1"].View || got["bi-1"].Click {
		t.Fatalf("leitor deveria ver mas não clicar; got=%+v", got["bi-1"])
	}
}

func TestAvaliar_AcumulaPorOR(t *testing.T) {
	perms := []Permissao{
		{BlindIndex: "bi-1", Papel: "editor", View: true},
		{BlindIndex: "bi-1", Papel: "editor", Click: true},
	}
	got := Evaluator{}.Avaliar(Sujeito{Papel: "editor"}, perms, []string{"bi-1"})
	if !got["bi-1"].View || !got["bi-1"].Click {
		t.Fatalf("permissões deveriam acumular; got=%+v", got["bi-1"])
	}
}

func TestAvaliar_PapelCuringa(t *testing.T) {
	perms := []Permissao{{BlindIndex: "bi-1", Papel: "*", View: true}}
	got := Evaluator{}.Avaliar(Sujeito{Papel: "qualquer"}, perms, []string{"bi-1"})
	if !got["bi-1"].View {
		t.Fatal("papel curinga deveria conceder a qualquer sujeito")
	}
}

func TestAvaliar_CondicaoDinamica(t *testing.T) {
	perms := []Permissao{{
		BlindIndex: "bi-secreto",
		Papel:      "gestor",
		Condicao:   map[string]any{"atributo": "tipo", "igual": "dono"},
		View:       true,
		Click:      true,
	}}

	// tipo=dono satisfaz a condição.
	dono := Evaluator{}.Avaliar(Sujeito{Papel: "gestor", Tipo: "dono"}, perms, []string{"bi-secreto"})
	if !dono["bi-secreto"].View {
		t.Fatal("dono deveria receber acesso")
	}

	// tipo=cliente não satisfaz → negado, mesmo com view=true na linha.
	cliente := Evaluator{}.Avaliar(Sujeito{Papel: "gestor", Tipo: "cliente"}, perms, []string{"bi-secreto"})
	if cliente["bi-secreto"].View {
		t.Fatal("condição não satisfeita deveria negar")
	}
}

func TestAvaliar_IgnoraPermissoesNaoPedidas(t *testing.T) {
	perms := []Permissao{{BlindIndex: "bi-outro", Papel: "editor", View: true}}
	got := Evaluator{}.Avaliar(Sujeito{Papel: "editor"}, perms, []string{"bi-1"})
	if _, existe := got["bi-outro"]; existe {
		t.Fatal("não deveria retornar componentes fora do pedido")
	}
}

func TestCondicao_MalformadaNega(t *testing.T) {
	if condicaoSatisfeita(map[string]any{"operador": "regex"}, Sujeito{}) {
		t.Fatal("condição malformada deveria negar (fail-closed)")
	}
}
