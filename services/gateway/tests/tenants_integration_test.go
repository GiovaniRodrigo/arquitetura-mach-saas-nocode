//go:build integration

// Teste ponta-a-ponta das rotas de tenants: Gateway → IAM (gRPC) → Postgres,
// fechando o fluxo "criar um negócio para um cliente" (spec 004, RF07). Requer
// o Postgres do docker-compose com as migrações aplicadas. Executar:
//
//	DATABASE_URL=postgres://mach:mach@localhost:5432/machv4?sslmode=disable \
//	  go test -tags integration ./services/gateway/tests/...
package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTenants_DonoCriaEListaNaCadeiaReal(t *testing.T) {
	handler, iss, tenantA, _ := harness(t)
	tok := tokenDe(t, iss, tenantA, "dono")

	// POST cria o tenant filho (cliente).
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tenants", strings.NewReader(`{"nome":"Cliente E2E"}`))
	req.Header.Set("Authorization", "Bearer "+tok)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST esperava 201; got=%d body=%s", rec.Code, rec.Body.String())
	}
	var criado struct{ ID, Nome string }
	if err := json.Unmarshal(rec.Body.Bytes(), &criado); err != nil {
		t.Fatalf("json criado: %v", err)
	}
	if criado.ID == "" || criado.Nome != "Cliente E2E" {
		t.Fatalf("resposta de criação inesperada: %+v", criado)
	}

	// GET lista e o tenant criado aparece.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/tenants", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET esperava 200; got=%d", rec.Code)
	}
	var lista struct {
		Tenants []struct{ ID, Nome string } `json:"tenants"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &lista); err != nil {
		t.Fatalf("json lista: %v", err)
	}
	var achou bool
	for _, tt := range lista.Tenants {
		if tt.ID == criado.ID {
			achou = true
		}
	}
	if !achou {
		t.Fatalf("tenant criado não apareceu na listagem: %+v", lista.Tenants)
	}
}

func TestTenants_ClienteNaoCria_403(t *testing.T) {
	handler, iss, tenantA, _ := harness(t)
	tok := tokenDe(t, iss, tenantA, "cliente")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tenants", strings.NewReader(`{"nome":"Proibido"}`))
	req.Header.Set("Authorization", "Bearer "+tok)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("cliente final deveria receber 403; got=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestTenants_IsolamentoEntreTenants(t *testing.T) {
	handler, iss, tenantA, tenantB := harness(t)

	// A cria um tenant filho.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tenants", strings.NewReader(`{"nome":"So de A"}`))
	req.Header.Set("Authorization", "Bearer "+tokenDe(t, iss, tenantA, "dono"))
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST A esperava 201; got=%d body=%s", rec.Code, rec.Body.String())
	}
	var criado struct{ ID string }
	_ = json.Unmarshal(rec.Body.Bytes(), &criado)

	// B lista e nunca enxerga o tenant filho de A (RN01/RN05: hierarquia por parent_id).
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/tenants", nil)
	req.Header.Set("Authorization", "Bearer "+tokenDe(t, iss, tenantB, "dono"))
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET B esperava 200; got=%d", rec.Code)
	}
	var lista struct {
		Tenants []struct{ ID string } `json:"tenants"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &lista)
	for _, tt := range lista.Tenants {
		if tt.ID == criado.ID {
			t.Fatalf("B não deveria ver o tenant de A (%s)", criado.ID)
		}
	}
}

func TestTenants_CicloCompletoDeVisualizarAtualizarExcluir(t *testing.T) {
	handler, iss, tenantA, _ := harness(t)
	tok := tokenDe(t, iss, tenantA, "dono")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tenants", strings.NewReader(`{"nome":"Ciclo E2E"}`))
	req.Header.Set("Authorization", "Bearer "+tok)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST esperava 201; got=%d body=%s", rec.Code, rec.Body.String())
	}
	var criado struct{ ID, Nome string }
	_ = json.Unmarshal(rec.Body.Bytes(), &criado)

	// GET /{id} devolve o tenant recém-criado.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/tenants/"+criado.ID, nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /{id} esperava 200; got=%d body=%s", rec.Code, rec.Body.String())
	}

	// PATCH renomeia.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/tenants/"+criado.ID, strings.NewReader(`{"nome":"Ciclo Renomeado"}`))
	req.Header.Set("Authorization", "Bearer "+tok)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("PATCH esperava 200; got=%d body=%s", rec.Code, rec.Body.String())
	}
	var atualizado struct{ Nome string }
	_ = json.Unmarshal(rec.Body.Bytes(), &atualizado)
	if atualizado.Nome != "Ciclo Renomeado" {
		t.Fatalf("nome não atualizado: %+v", atualizado)
	}

	// DELETE remove.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/tenants/"+criado.ID, nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE esperava 204; got=%d body=%s", rec.Code, rec.Body.String())
	}

	// GET depois do DELETE vira 404.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/tenants/"+criado.ID, nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET pós-exclusão esperava 404; got=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestTenants_IsolamentoNoVisualizarAtualizarExcluir(t *testing.T) {
	handler, iss, tenantA, tenantB := harness(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tenants", strings.NewReader(`{"nome":"So de A 2"}`))
	req.Header.Set("Authorization", "Bearer "+tokenDe(t, iss, tenantA, "dono"))
	handler.ServeHTTP(rec, req)
	var criado struct{ ID string }
	_ = json.Unmarshal(rec.Body.Bytes(), &criado)

	tokB := tokenDe(t, iss, tenantB, "dono")
	for _, tc := range []struct {
		nome   string
		method string
		corpo  string
	}{
		{"GET", http.MethodGet, ""},
		{"PATCH", http.MethodPatch, `{"nome":"Invasão"}`},
		{"DELETE", http.MethodDelete, ""},
	} {
		rec := httptest.NewRecorder()
		var req *http.Request
		if tc.corpo == "" {
			req = httptest.NewRequest(tc.method, "/api/v1/tenants/"+criado.ID, nil)
		} else {
			req = httptest.NewRequest(tc.method, "/api/v1/tenants/"+criado.ID, strings.NewReader(tc.corpo))
		}
		req.Header.Set("Authorization", "Bearer "+tokB)
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s de B sobre tenant de A deveria ser 404; got=%d body=%s", tc.nome, rec.Code, rec.Body.String())
		}
	}
}
