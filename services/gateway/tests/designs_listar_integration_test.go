//go:build integration

// Teste ponta-a-ponta de GET /api/v1/designs?sistema_id={id} (RF09): lista as
// telas (Designs) de um sistema. Reusa sistemasHarness (mesma cadeia real
// Gateway → Design Engine → Postgres com RLS efetiva). Executar:
//
//	DATABASE_URL=postgres://mach:mach@localhost:5432/machv4?sslmode=disable \
//	  go test -tags integration -p 1 ./services/gateway/tests/...
package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// criarSistema cria um sistema via API e devolve seu id.
func criarSistema(t *testing.T, handler http.Handler, token, nome string) string {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sistemas", strings.NewReader(`{"nome":"`+nome+`"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("criar sistema: esperava 201; got=%d body=%s", rec.Code, rec.Body.String())
	}
	var criado struct{ ID string }
	if err := json.Unmarshal(rec.Body.Bytes(), &criado); err != nil {
		t.Fatalf("json sistema criado: %v", err)
	}
	return criado.ID
}

// criarDesign cria uma tela (Design) via API e devolve seu id.
func criarDesign(t *testing.T, handler http.Handler, token, sistemaID, nome string) string {
	t.Helper()
	corpo := `{"sistema_id":"` + sistemaID + `","nome":"` + nome + `","arvore":{"blind_index":"root-` + nome + `","tipo":"tela"}}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/designs", strings.NewReader(corpo))
	req.Header.Set("Authorization", "Bearer "+token)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("criar design: esperava 201; got=%d body=%s", rec.Code, rec.Body.String())
	}
	var criado struct {
		DesignID string `json:"design_id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &criado); err != nil {
		t.Fatalf("json design criado: %v", err)
	}
	return criado.DesignID
}

func listarTelas(t *testing.T, handler http.Handler, token, sistemaID string) []struct{ ID, Nome string } {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/designs?sistema_id="+sistemaID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("listar telas: esperava 200; got=%d body=%s", rec.Code, rec.Body.String())
	}
	var lista struct {
		Telas []struct{ ID, Nome string } `json:"telas"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &lista); err != nil {
		t.Fatalf("json lista: %v", err)
	}
	return lista.Telas
}

func TestListarDesigns_SistemaSemTelas_ListaVazia(t *testing.T) {
	handler, iss, tenantA, _ := sistemasHarness(t)
	tokA := tokenDe(t, iss, tenantA, "dono")
	sisID := criarSistema(t, handler, tokA, "Sistema Sem Telas")

	telas := listarTelas(t, handler, tokA, sisID)
	if len(telas) != 0 {
		t.Fatalf("esperava lista vazia; got=%+v", telas)
	}
}

func TestListarDesigns_ComTelas_ListaTodas(t *testing.T) {
	handler, iss, tenantA, _ := sistemasHarness(t)
	tokA := tokenDe(t, iss, tenantA, "dono")
	sisID := criarSistema(t, handler, tokA, "CRM")
	d1 := criarDesign(t, handler, tokA, sisID, "Home")
	d2 := criarDesign(t, handler, tokA, sisID, "Cadastro")

	telas := listarTelas(t, handler, tokA, sisID)
	if len(telas) != 2 {
		t.Fatalf("esperava 2 telas; got=%+v", telas)
	}
	ids := map[string]bool{telas[0].ID: true, telas[1].ID: true}
	if !ids[d1] || !ids[d2] {
		t.Fatalf("telas criadas não apareceram na listagem: got=%+v want=[%s %s]", telas, d1, d2)
	}
}

// O teste mais importante deste arquivo: B não pode enxergar as telas do
// sistema de A mesmo sabendo o sistema_id exato (RLS por tenant no Design
// Engine, não confiança em quem chama).
func TestListarDesigns_IsolamentoEntreTenants(t *testing.T) {
	handler, iss, tenantA, tenantB := sistemasHarness(t)
	tokA := tokenDe(t, iss, tenantA, "dono")
	tokB := tokenDe(t, iss, tenantB, "dono")

	sisID := criarSistema(t, handler, tokA, "So de A")
	criarDesign(t, handler, tokA, sisID, "Tela de A")

	telas := listarTelas(t, handler, tokB, sisID)
	if len(telas) != 0 {
		t.Fatalf("B não deveria ver telas do sistema de A: %+v", telas)
	}
}
