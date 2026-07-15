//go:build integration

// Requer o Postgres do docker-compose com as migrações aplicadas (RLS 0010).
// Executar:
//
//	DATABASE_URL=postgres://mach:mach@localhost:5432/machv4?sslmode=disable \
//	  go test -tags integration ./services/design/internal/store/...
package store

import (
	"testing"
)

func TestCriarEListarSistemas(t *testing.T) {
	pool, tenantA, _, _, _ := setup(t)
	s := New(pool)
	ctx := ctxFor(tenantA)

	id, err := s.CriarSistema(ctx, "Sistema Novo de A")
	if err != nil {
		t.Fatalf("criar sistema: %v", err)
	}
	if id == "" {
		t.Fatal("id vazio ao criar sistema")
	}

	lista, err := s.ListarSistemas(ctx)
	if err != nil {
		t.Fatalf("listar: %v", err)
	}
	// setup já semeia "itg-design-A"; agora há ao menos 2 e o novo deve aparecer.
	var achou bool
	for _, sis := range lista {
		if sis.GetId() == id && sis.GetNome() == "Sistema Novo de A" {
			achou = true
		}
	}
	if !achou {
		t.Fatalf("sistema criado não apareceu na listagem: %+v", lista)
	}
}

func TestListarSistemas_IsolamentoEntreTenants(t *testing.T) {
	pool, tenantA, tenantB, _, sisB := setup(t)
	s := New(pool)

	// O tenant A nunca enxerga o sistema semeado do tenant B (RLS por tenant).
	lista, err := s.ListarSistemas(ctxFor(tenantA))
	if err != nil {
		t.Fatalf("listar A: %v", err)
	}
	for _, sis := range lista {
		if sis.GetId() == sisB {
			t.Fatalf("A não deveria ver o sistema de B (%s)", sisB)
		}
	}
	_ = tenantB
}
