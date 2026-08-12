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

	designv1 "github.com/machv4/platform/gen/go/construtor/design/v1"
	iamv1 "github.com/machv4/platform/gen/go/construtor/iam/v1"
)

// fakeDesign implementa DesignCliente; só os métodos de sistema são exercitados.
type fakeDesign struct {
	sistemas       []*designv1.Sistema
	criarResp      *designv1.Sistema
	criarErr       error
	listarErr      error
	nomeCriado     string
	tenantIDPedido string

	designsResp            []*designv1.DesignResumo
	listarDesignsErr       error
	sistemaIDDesignsPedido string
}

func (f *fakeDesign) CriarDesign(context.Context, *designv1.Design, ...grpc.CallOption) (*designv1.Design, error) {
	return nil, nil
}
func (f *fakeDesign) ObterDesign(context.Context, *designv1.ObterDesignRequest, ...grpc.CallOption) (*designv1.Design, error) {
	return nil, nil
}
func (f *fakeDesign) AtualizarDesign(context.Context, *designv1.Design, ...grpc.CallOption) (*designv1.Design, error) {
	return nil, nil
}
func (f *fakeDesign) RemoverDesign(context.Context, *designv1.ObterDesignRequest, ...grpc.CallOption) (*designv1.RemoverResponse, error) {
	return nil, nil
}
func (f *fakeDesign) ListarDesigns(_ context.Context, in *designv1.ListarDesignsRequest, _ ...grpc.CallOption) (*designv1.ListarDesignsResponse, error) {
	f.sistemaIDDesignsPedido = in.GetSistemaId()
	if f.listarDesignsErr != nil {
		return nil, f.listarDesignsErr
	}
	return &designv1.ListarDesignsResponse{Designs: f.designsResp}, nil
}
func (f *fakeDesign) CriarSistema(_ context.Context, in *designv1.CriarSistemaRequest, _ ...grpc.CallOption) (*designv1.Sistema, error) {
	f.nomeCriado = in.GetNome()
	return f.criarResp, f.criarErr
}
func (f *fakeDesign) ListarSistemas(_ context.Context, in *designv1.ListarSistemasRequest, _ ...grpc.CallOption) (*designv1.ListarSistemasResponse, error) {
	f.tenantIDPedido = in.GetTenantId()
	if f.listarErr != nil {
		return nil, f.listarErr
	}
	return &designv1.ListarSistemasResponse{Sistemas: f.sistemas}, nil
}
func (f *fakeDesign) AtualizarWhiteLabel(context.Context, *designv1.AtualizarWhiteLabelRequest, ...grpc.CallOption) (*designv1.WhiteLabel, error) {
	return nil, nil
}

func TestListarSistemas_OK(t *testing.T) {
	fd := &fakeDesign{sistemas: []*designv1.Sistema{{Id: "a", Nome: "Alfa"}}}
	rec := httptest.NewRecorder()
	ListarSistemas(fd, &fakeIamTenants{})(rec, httptest.NewRequest(http.MethodGet, "/api/v1/sistemas", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d", rec.Code)
	}
	var corpo struct {
		Sistemas []sistemaResp `json:"sistemas"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &corpo); err != nil {
		t.Fatalf("json: %v", err)
	}
	if len(corpo.Sistemas) != 1 || corpo.Sistemas[0].Nome != "Alfa" {
		t.Fatalf("corpo inesperado: %+v", corpo.Sistemas)
	}
	if fd.tenantIDPedido != "" {
		t.Fatalf("sem tenant_id na query, não deveria propagar tenant_id ao Design: %q", fd.tenantIDPedido)
	}
}

// Com ?tenant_id, o handler PRIMEIRO valida a hierarquia via IAM.ObterTenant
// (RN05) antes de repassar ao Design Engine (RF08).
func TestListarSistemas_ComTenantID_ValidaHierarquiaAntesDeChamarDesign(t *testing.T) {
	fd := &fakeDesign{sistemas: []*designv1.Sistema{{Id: "a", Nome: "Alfa"}}}
	fi := &fakeIamTenants{obterResp: &iamv1.Tenant{Id: "filho-1", Nome: "Cliente"}}
	rec := httptest.NewRecorder()
	ListarSistemas(fd, fi)(rec, httptest.NewRequest(http.MethodGet, "/api/v1/sistemas?tenant_id=filho-1", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d (%s)", rec.Code, rec.Body.String())
	}
	if fd.tenantIDPedido != "filho-1" {
		t.Fatalf("tenant_id não propagado ao Design: %q", fd.tenantIDPedido)
	}
}

// Quando o IAM nega (tenant fora da hierarquia — RN05), o Design Engine NUNCA
// é chamado: este é o gate de segurança de RF08.
func TestListarSistemas_ComTenantID_ForaDaHierarquia_404_NaoChamaDesign(t *testing.T) {
	fd := &fakeDesign{sistemas: []*designv1.Sistema{{Id: "vazando", Nome: "NaoDeveriaAparecer"}}}
	fi := &fakeIamTenants{obterErr: status.Error(codes.NotFound, "cliente não encontrado")}
	rec := httptest.NewRecorder()
	ListarSistemas(fd, fi)(rec, httptest.NewRequest(http.MethodGet, "/api/v1/sistemas?tenant_id=alheio", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperava 404; got %d (%s)", rec.Code, rec.Body.String())
	}
	if fd.tenantIDPedido != "" {
		t.Fatalf("Design não deveria ter sido chamado após IAM negar; tenant_id recebido=%q", fd.tenantIDPedido)
	}
}

func TestCriarSistema_Created(t *testing.T) {
	fd := &fakeDesign{criarResp: &designv1.Sistema{Id: "sis-1", Nome: "Demo"}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sistemas", strings.NewReader(`{"nome":"Demo"}`))
	CriarSistema(fd)(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status: %d (%s)", rec.Code, rec.Body.String())
	}
	if fd.nomeCriado != "Demo" {
		t.Fatalf("nome não propagado: %q", fd.nomeCriado)
	}
}

func TestCriarSistema_NomeVazio_400(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sistemas", strings.NewReader(`{"nome":""}`))
	CriarSistema(&fakeDesign{})(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperava 400; got %d", rec.Code)
	}
}

func TestCriarSistema_PermissionDenied_403(t *testing.T) {
	fd := &fakeDesign{criarErr: status.Error(codes.PermissionDenied, "cliente")}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sistemas", strings.NewReader(`{"nome":"Demo"}`))
	CriarSistema(fd)(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("esperava 403; got %d (%s)", rec.Code, rec.Body.String())
	}
}
