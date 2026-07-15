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
)

// fakeDesign implementa DesignCliente; só os métodos de sistema são exercitados.
type fakeDesign struct {
	sistemas   []*designv1.Sistema
	criarResp  *designv1.Sistema
	criarErr   error
	listarErr  error
	nomeCriado string
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
func (f *fakeDesign) CriarSistema(_ context.Context, in *designv1.CriarSistemaRequest, _ ...grpc.CallOption) (*designv1.Sistema, error) {
	f.nomeCriado = in.GetNome()
	return f.criarResp, f.criarErr
}
func (f *fakeDesign) ListarSistemas(context.Context, *designv1.ListarSistemasRequest, ...grpc.CallOption) (*designv1.ListarSistemasResponse, error) {
	if f.listarErr != nil {
		return nil, f.listarErr
	}
	return &designv1.ListarSistemasResponse{Sistemas: f.sistemas}, nil
}

func TestListarSistemas_OK(t *testing.T) {
	fd := &fakeDesign{sistemas: []*designv1.Sistema{{Id: "a", Nome: "Alfa"}}}
	rec := httptest.NewRecorder()
	ListarSistemas(fd)(rec, httptest.NewRequest(http.MethodGet, "/api/v1/sistemas", nil))

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
