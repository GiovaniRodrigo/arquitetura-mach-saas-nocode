package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	iamv1 "github.com/machv4/platform/gen/go/construtor/iam/v1"
)

// fakeIamTenants implementa TenantsCliente.
type fakeIamTenants struct {
	tenants    []*iamv1.Tenant
	criarResp  *iamv1.Tenant
	criarErr   error
	listarErr  error
	nomeCriado string

	obterResp      *iamv1.Tenant
	obterErr       error
	atualizarResp  *iamv1.Tenant
	atualizarErr   error
	idAtualizado   string
	nomeAtualizado string
	excluirErr     error
	idExcluido     string
}

func (f *fakeIamTenants) ListarTenants(context.Context, *iamv1.ListarTenantsRequest, ...grpc.CallOption) (*iamv1.ListarTenantsResponse, error) {
	if f.listarErr != nil {
		return nil, f.listarErr
	}
	return &iamv1.ListarTenantsResponse{Tenants: f.tenants}, nil
}

func (f *fakeIamTenants) CriarTenant(_ context.Context, in *iamv1.CriarTenantRequest, _ ...grpc.CallOption) (*iamv1.Tenant, error) {
	f.nomeCriado = in.GetNome()
	return f.criarResp, f.criarErr
}

func (f *fakeIamTenants) ObterTenant(context.Context, *iamv1.ObterTenantRequest, ...grpc.CallOption) (*iamv1.Tenant, error) {
	return f.obterResp, f.obterErr
}

func (f *fakeIamTenants) AtualizarTenant(_ context.Context, in *iamv1.AtualizarTenantRequest, _ ...grpc.CallOption) (*iamv1.Tenant, error) {
	f.idAtualizado, f.nomeAtualizado = in.GetId(), in.GetNome()
	return f.atualizarResp, f.atualizarErr
}

func (f *fakeIamTenants) ExcluirTenant(_ context.Context, in *iamv1.ExcluirTenantRequest, _ ...grpc.CallOption) (*iamv1.ExcluirTenantResponse, error) {
	f.idExcluido = in.GetId()
	if f.excluirErr != nil {
		return nil, f.excluirErr
	}
	return &iamv1.ExcluirTenantResponse{}, nil
}

// reqComID monta uma requisição chi com o URL param "id" já resolvido, para
// exercitar handlers que usam chi.URLParam sem precisar de um router real.
func reqComID(method, url, id, corpo string) *http.Request {
	var req *http.Request
	if corpo == "" {
		req = httptest.NewRequest(method, url, nil)
	} else {
		req = httptest.NewRequest(method, url, strings.NewReader(corpo))
	}
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestListarTenants_OK(t *testing.T) {
	fi := &fakeIamTenants{tenants: []*iamv1.Tenant{{Id: "t1", Nome: "Acme", Tipo: "cliente"}}}
	rec := httptest.NewRecorder()
	ListarTenants(fi)(rec, httptest.NewRequest(http.MethodGet, "/api/v1/tenants", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d", rec.Code)
	}
	var corpo struct {
		Tenants []tenantResp `json:"tenants"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &corpo); err != nil {
		t.Fatalf("json: %v", err)
	}
	if len(corpo.Tenants) != 1 || corpo.Tenants[0].Nome != "Acme" {
		t.Fatalf("corpo inesperado: %+v", corpo.Tenants)
	}
}

func TestCriarTenant_Created(t *testing.T) {
	fi := &fakeIamTenants{criarResp: &iamv1.Tenant{Id: "t-1", Nome: "Acme"}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tenants", strings.NewReader(`{"nome":"Acme"}`))
	CriarTenant(fi)(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status: %d (%s)", rec.Code, rec.Body.String())
	}
	if fi.nomeCriado != "Acme" {
		t.Fatalf("nome não propagado: %q", fi.nomeCriado)
	}
}

func TestCriarTenant_NomeVazio_400(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tenants", strings.NewReader(`{"nome":""}`))
	CriarTenant(&fakeIamTenants{})(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperava 400; got %d", rec.Code)
	}
}

func TestCriarTenant_PermissionDenied_403(t *testing.T) {
	fi := &fakeIamTenants{criarErr: status.Error(codes.PermissionDenied, "cliente")}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tenants", strings.NewReader(`{"nome":"Acme"}`))
	CriarTenant(fi)(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("esperava 403; got %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestListarTenants_Erro_500(t *testing.T) {
	fi := &fakeIamTenants{listarErr: status.Error(codes.Internal, "boom")}
	rec := httptest.NewRecorder()
	ListarTenants(fi)(rec, httptest.NewRequest(http.MethodGet, "/api/v1/tenants", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("esperava 500; got %d", rec.Code)
	}
}

func TestObterTenant_OK(t *testing.T) {
	fi := &fakeIamTenants{obterResp: &iamv1.Tenant{Id: "t1", Nome: "Acme"}}
	rec := httptest.NewRecorder()
	ObterTenant(fi)(rec, reqComID(http.MethodGet, "/api/v1/tenants/t1", "t1", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d (%s)", rec.Code, rec.Body.String())
	}
	var corpo tenantResp
	if err := json.Unmarshal(rec.Body.Bytes(), &corpo); err != nil {
		t.Fatalf("json: %v", err)
	}
	if corpo.ID != "t1" || corpo.Nome != "Acme" {
		t.Fatalf("corpo inesperado: %+v", corpo)
	}
}

func TestObterTenant_NotFound_404(t *testing.T) {
	fi := &fakeIamTenants{obterErr: status.Error(codes.NotFound, "cliente não encontrado")}
	rec := httptest.NewRecorder()
	ObterTenant(fi)(rec, reqComID(http.MethodGet, "/api/v1/tenants/t1", "t1", ""))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperava 404; got %d", rec.Code)
	}
}

func TestAtualizarTenant_OK(t *testing.T) {
	fi := &fakeIamTenants{atualizarResp: &iamv1.Tenant{Id: "t1", Nome: "Novo Nome"}}
	rec := httptest.NewRecorder()
	AtualizarTenant(fi)(rec, reqComID(http.MethodPatch, "/api/v1/tenants/t1", "t1", `{"nome":"Novo Nome"}`))
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d (%s)", rec.Code, rec.Body.String())
	}
	if fi.idAtualizado != "t1" || fi.nomeAtualizado != "Novo Nome" {
		t.Fatalf("args não propagados: id=%q nome=%q", fi.idAtualizado, fi.nomeAtualizado)
	}
}

func TestAtualizarTenant_NomeVazio_400(t *testing.T) {
	rec := httptest.NewRecorder()
	AtualizarTenant(&fakeIamTenants{})(rec, reqComID(http.MethodPatch, "/api/v1/tenants/t1", "t1", `{"nome":""}`))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperava 400; got %d", rec.Code)
	}
}

func TestAtualizarTenant_NotFound_404(t *testing.T) {
	fi := &fakeIamTenants{atualizarErr: status.Error(codes.NotFound, "cliente não encontrado")}
	rec := httptest.NewRecorder()
	AtualizarTenant(fi)(rec, reqComID(http.MethodPatch, "/api/v1/tenants/t1", "t1", `{"nome":"Novo"}`))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperava 404; got %d", rec.Code)
	}
}

func TestExcluirTenant_NoContent(t *testing.T) {
	fi := &fakeIamTenants{}
	rec := httptest.NewRecorder()
	ExcluirTenant(fi)(rec, reqComID(http.MethodDelete, "/api/v1/tenants/t1", "t1", ""))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("esperava 204; got %d (%s)", rec.Code, rec.Body.String())
	}
	if fi.idExcluido != "t1" {
		t.Fatalf("id não propagado: %q", fi.idExcluido)
	}
}

func TestExcluirTenant_PermissionDenied_403(t *testing.T) {
	fi := &fakeIamTenants{excluirErr: status.Error(codes.PermissionDenied, "cliente")}
	rec := httptest.NewRecorder()
	ExcluirTenant(fi)(rec, reqComID(http.MethodDelete, "/api/v1/tenants/t1", "t1", ""))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("esperava 403; got %d", rec.Code)
	}
}

func TestExcluirTenant_NotFound_404(t *testing.T) {
	fi := &fakeIamTenants{excluirErr: status.Error(codes.NotFound, "cliente não encontrado")}
	rec := httptest.NewRecorder()
	ExcluirTenant(fi)(rec, reqComID(http.MethodDelete, "/api/v1/tenants/t1", "t1", ""))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperava 404; got %d", rec.Code)
	}
}
