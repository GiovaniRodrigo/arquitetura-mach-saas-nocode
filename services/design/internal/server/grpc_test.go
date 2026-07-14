package server

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonv1 "github.com/machv4/platform/gen/go/construtor/common/v1"
	designv1 "github.com/machv4/platform/gen/go/construtor/design/v1"
	"github.com/machv4/platform/services/design/internal/store"
	"github.com/machv4/platform/pkg/tenantctx"
)

// fakeStore registra chamadas e devolve respostas controladas.
type fakeStore struct {
	criarID   string
	criarErr  error
	obterD    *designv1.Design
	obterErr  error
	salvarErr error
	salvou    *designv1.Design
}

func (f *fakeStore) Criar(_ context.Context, _ *designv1.Design) (string, error) {
	return f.criarID, f.criarErr
}
func (f *fakeStore) Obter(_ context.Context, _ string) (*designv1.Design, error) {
	return f.obterD, f.obterErr
}
func (f *fakeStore) Atualizar(_ context.Context, _ *designv1.Design) error { return nil }
func (f *fakeStore) Remover(_ context.Context, _ string) error            { return nil }
func (f *fakeStore) Salvar(_ context.Context, d *designv1.Design) error {
	f.salvou = d
	return f.salvarErr
}

func comTenant() context.Context {
	return tenantctx.NewContext(context.Background(), &commonv1.TenantContext{TenantId: "t-1", Tipo: "dono"})
}

func arvoreValida() *designv1.Componente {
	return &designv1.Componente{BlindIndex: "bi-root", Tipo: "container"}
}

func code(err error) codes.Code { return status.Code(err) }

func TestCriarDesign_SemTenant_Unauthenticated(t *testing.T) {
	s := New(&fakeStore{})
	_, err := s.CriarDesign(context.Background(), &designv1.Design{SistemaId: "s1", Arvore: arvoreValida()})
	if code(err) != codes.Unauthenticated {
		t.Fatalf("esperava Unauthenticated; got %v", err)
	}
}

func TestCriarDesign_ArvoreInvalida_InvalidArgument(t *testing.T) {
	s := New(&fakeStore{})
	// Árvore com filho sem tipo → inválida.
	arvore := &designv1.Componente{BlindIndex: "r", Tipo: "container",
		ComponenteFilhos: []*designv1.Componente{{BlindIndex: "f", Tipo: ""}}}
	_, err := s.CriarDesign(comTenant(), &designv1.Design{SistemaId: "s1", Arvore: arvore})
	if code(err) != codes.InvalidArgument {
		t.Fatalf("esperava InvalidArgument; got %v", err)
	}
}

func TestCriarDesign_OK(t *testing.T) {
	s := New(&fakeStore{criarID: "d-123"})
	out, err := s.CriarDesign(comTenant(), &designv1.Design{SistemaId: "s1", Nome: "Tela", Arvore: arvoreValida()})
	if err != nil {
		t.Fatalf("inesperado: %v", err)
	}
	if out.GetId() != "d-123" || out.GetSistemaId() != "s1" {
		t.Fatalf("resposta inesperada: %+v", out)
	}
}

func TestObterDesign_NaoEncontrado_NotFound(t *testing.T) {
	s := New(&fakeStore{obterErr: store.ErrNaoEncontrado})
	_, err := s.ObterDesign(comTenant(), &designv1.ObterDesignRequest{Id: "x"})
	if code(err) != codes.NotFound {
		t.Fatalf("esperava NotFound; got %v", err)
	}
}

func TestSalvarDesign_SemID_InvalidArgument(t *testing.T) {
	s := New(&fakeStore{})
	_, err := s.SalvarDesign(comTenant(), &designv1.SalvarDesignRequest{
		Design: &designv1.Design{SistemaId: "s1", Arvore: arvoreValida()}})
	if code(err) != codes.InvalidArgument {
		t.Fatalf("esperava InvalidArgument sem id; got %v", err)
	}
}

func TestSalvarDesign_OK(t *testing.T) {
	fs := &fakeStore{}
	s := New(fs)
	resp, err := s.SalvarDesign(comTenant(), &designv1.SalvarDesignRequest{
		Design: &designv1.Design{Id: "d-9", SistemaId: "s1", Arvore: arvoreValida()}})
	if err != nil {
		t.Fatalf("inesperado: %v", err)
	}
	if !resp.GetSucesso() || fs.salvou.GetId() != "d-9" {
		t.Fatalf("salvar não propagou o design: %+v", fs.salvou)
	}
}
