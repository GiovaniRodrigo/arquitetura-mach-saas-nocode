package validation

import "testing"

func f64(v float64) *float64 { return &v }
func i(v int) *int           { return &v }

func schemaExemplo() map[string]CampoDef {
	return map[string]CampoDef{
		"bi-nome":  {BlindIndex: "bi-nome", Tipo: TipoString, Obrigatorio: true, Limites: Limites{MaxLength: i(10)}},
		"bi-idade": {BlindIndex: "bi-idade", Tipo: TipoNumber, Obrigatorio: true, Limites: Limites{Min: f64(0), Max: f64(120)}},
		"bi-nasc":  {BlindIndex: "bi-nasc", Tipo: TipoDate},
		"bi-ativo": {BlindIndex: "bi-ativo", Tipo: TipoBool},
	}
}

func TestValidar_PayloadValido(t *testing.T) {
	erros := Validar(schemaExemplo(), map[string]string{
		"bi-nome":  "Ana",
		"bi-idade": "30",
		"bi-nasc":  "1994-05-01",
		"bi-ativo": "true",
	})
	if len(erros) != 0 {
		t.Fatalf("payload válido gerou erros: %v", erros)
	}
}

func TestValidar_CampoDesconhecido_Rejeitado(t *testing.T) {
	// Submissão maliciosa injeta um blind_index fora do schema (critério 2).
	erros := Validar(schemaExemplo(), map[string]string{
		"bi-nome":     "Ana",
		"bi-idade":    "30",
		"bi-injetado": "'; DROP TABLE",
	})
	if erros["bi-injetado"] != msgDesconhecido {
		t.Fatalf("chave desconhecida deveria ser rejeitada; erros=%v", erros)
	}
}

func TestValidar_ObrigatorioAusente(t *testing.T) {
	erros := Validar(schemaExemplo(), map[string]string{"bi-idade": "30"})
	if erros["bi-nome"] != msgObrigatorio {
		t.Fatalf("esperava erro de obrigatório em bi-nome; erros=%v", erros)
	}
}

func TestValidar_LimitesETipos(t *testing.T) {
	erros := Validar(schemaExemplo(), map[string]string{
		"bi-nome":  "nome-muito-comprido",
		"bi-idade": "999",
		"bi-nasc":  "01/01/2000",
		"bi-ativo": "sim",
	})
	if erros["bi-nome"] != msgMaxLength {
		t.Fatalf("bi-nome deveria falhar por comprimento; got %q", erros["bi-nome"])
	}
	if erros["bi-idade"] != msgMax {
		t.Fatalf("bi-idade deveria falhar por máximo; got %q", erros["bi-idade"])
	}
	if erros["bi-nasc"] != msgTipoDate {
		t.Fatalf("bi-nasc deveria falhar por formato de data; got %q", erros["bi-nasc"])
	}
	if erros["bi-ativo"] != msgTipoBool {
		t.Fatalf("bi-ativo deveria falhar por booleano; got %q", erros["bi-ativo"])
	}
}

// Nenhuma mensagem pode conter nomes reais de coluna/tabela — só termos genéricos.
func TestValidar_MensagensNaoVazamNomes(t *testing.T) {
	erros := Validar(schemaExemplo(), map[string]string{"bi-idade": "-5"})
	for bi, msg := range erros {
		if bi == "" {
			t.Fatal("chave de erro vazia — deveria ser um blind_index")
		}
		if msg == "" {
			t.Fatalf("mensagem vazia para %s", bi)
		}
	}
}
