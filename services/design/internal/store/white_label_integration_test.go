//go:build integration

// Requer o Postgres do docker-compose com as migrações aplicadas (incl. 0015).
// Executar:
//
//	DATABASE_URL=postgres://mach:mach@localhost:5432/machv4?sslmode=disable \
//	  go test -tags integration ./services/design/internal/store/...
package store

import (
	"testing"
)

func TestAtualizarWhiteLabel_CriaESobrescreve(t *testing.T) {
	pool, tenantA, _, _, _ := setup(t)
	s := New(pool)
	ctx := ctxFor(tenantA)

	out, err := s.AtualizarWhiteLabel(ctx, WhiteLabel{
		LogoURL:        "https://cdn.example.com/logo.png",
		CorPrimaria:    "#112233",
		CorSecundaria:  "#445566",
		DominioProprio: "cliente.example.com",
	})
	if err != nil {
		t.Fatalf("criar: %v", err)
	}
	if out.LogoURL != "https://cdn.example.com/logo.png" || out.CorPrimaria != "#112233" ||
		out.CorSecundaria != "#445566" || out.DominioProprio != "cliente.example.com" {
		t.Fatalf("valores não preservados: %+v", out)
	}
	if out.DominioValidado {
		t.Fatalf("dominio_validado deveria ser sempre false (validação é assíncrona, fora de escopo); got %+v", out)
	}

	// Upsert: chamar de novo com valores diferentes sobrescreve a mesma linha
	// (tenant_id é PK), não cria uma segunda.
	out2, err := s.AtualizarWhiteLabel(ctx, WhiteLabel{
		LogoURL:        "https://cdn.example.com/logo-v2.png",
		CorPrimaria:    "#000000",
		CorSecundaria:  "#ffffff",
		DominioProprio: "",
	})
	if err != nil {
		t.Fatalf("sobrescrever: %v", err)
	}
	if out2.LogoURL != "https://cdn.example.com/logo-v2.png" || out2.CorPrimaria != "#000000" ||
		out2.CorSecundaria != "#ffffff" || out2.DominioProprio != "" {
		t.Fatalf("upsert não aplicado: %+v", out2)
	}
	if out2.DominioValidado {
		t.Fatalf("dominio_validado deveria continuar false após upsert; got %+v", out2)
	}
}

func TestAtualizarWhiteLabel_IsolamentoEntreTenants(t *testing.T) {
	pool, tenantA, tenantB, _, _ := setup(t)
	s := New(pool)

	if _, err := s.AtualizarWhiteLabel(ctxFor(tenantA), WhiteLabel{LogoURL: "https://a/logo.png"}); err != nil {
		t.Fatalf("criar A: %v", err)
	}
	if _, err := s.AtualizarWhiteLabel(ctxFor(tenantB), WhiteLabel{LogoURL: "https://b/logo.png"}); err != nil {
		t.Fatalf("criar B: %v", err)
	}

	outA, err := s.AtualizarWhiteLabel(ctxFor(tenantA), WhiteLabel{LogoURL: "https://a/logo.png", CorPrimaria: "#aaaaaa"})
	if err != nil {
		t.Fatalf("reler A: %v", err)
	}
	if outA.LogoURL != "https://a/logo.png" || outA.CorPrimaria != "#aaaaaa" {
		t.Fatalf("A não deveria ver/sobrescrever o white label de B: %+v", outA)
	}
}
