package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
