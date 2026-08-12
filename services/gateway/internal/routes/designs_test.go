package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	designv1 "github.com/machv4/platform/gen/go/construtor/design/v1"
)

func TestListarDesigns_SemSistemaID_BadRequest(t *testing.T) {
	h := ListarDesigns(&fakeDesign{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/designs", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperava 400; got=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestListarDesigns_OK(t *testing.T) {
	fd := &fakeDesign{designsResp: []*designv1.DesignResumo{{Id: "d1", Nome: "Home"}, {Id: "d2", Nome: "Cadastro"}}}
	h := ListarDesigns(fd)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/designs?sistema_id=s1", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200; got=%d body=%s", rec.Code, rec.Body.String())
	}
	if fd.sistemaIDDesignsPedido != "s1" {
		t.Fatalf("sistema_id não propagado: %q", fd.sistemaIDDesignsPedido)
	}
	var body struct {
		Telas []struct {
			ID   string `json:"id"`
			Nome string `json:"nome"`
		} `json:"telas"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	if len(body.Telas) != 2 || body.Telas[0].Nome != "Home" {
		t.Fatalf("lista inesperada: %+v", body.Telas)
	}
}
