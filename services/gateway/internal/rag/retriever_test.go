package rag

import "testing"

func TestRecuperar_PriorizaTermoDaConsulta(t *testing.T) {
	docs := Recuperar("como isolar dados entre tenants em multi-tenancy?", 2)
	if len(docs) == 0 {
		t.Fatal("esperava ao menos um documento")
	}
	if docs[0].ID != "01-multi-tenancy" {
		t.Errorf("esperava 01-multi-tenancy no topo, veio %q", docs[0].ID)
	}
}

func TestRecuperar_SemCorrespondenciaCaiParaFallback(t *testing.T) {
	docs := Recuperar("xyzabc termo que nao existe em lugar nenhum qwerty", 3)
	if len(docs) != 3 {
		t.Fatalf("esperava fallback com 3 documentos, veio %d", len(docs))
	}
}

func TestRecuperar_RespeitaLimite(t *testing.T) {
	docs := Recuperar("regras de negócio e versionamento", 1)
	if len(docs) != 1 {
		t.Fatalf("esperava 1 documento, veio %d", len(docs))
	}
}

func TestRecuperar_LimiteZeroDevolveNil(t *testing.T) {
	if docs := Recuperar("qualquer coisa", 0); docs != nil {
		t.Errorf("esperava nil, veio %v", docs)
	}
}
