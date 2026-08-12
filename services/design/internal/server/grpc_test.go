package server

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonv1 "github.com/machv4/platform/gen/go/construtor/common/v1"
	designv1 "github.com/machv4/platform/gen/go/construtor/design/v1"
	"github.com/machv4/platform/pkg/tenantctx"
	"github.com/machv4/platform/services/design/internal/store"
)

// fakeStore registra chamadas e devolve respostas controladas.
type fakeStore struct {
	criarID     string
	criarErr    error
	obterD      *designv1.Design
	obterErr    error
	salvarErr   error
	salvou      *designv1.Design
	sistemaID   string
	sistemaNome string
	listaSis    []*designv1.Sistema

	listaDesigns           []*designv1.DesignResumo
	listarDesignsErr       error
	sistemaIDDesignsPedido string

	listaSisTenant   []*designv1.Sistema
	tenantIDPedido   string
	listarTenantErr  error
	whiteLabelOut    store.WhiteLabel
	whiteLabelErr    error
	whiteLabelPedido store.WhiteLabel
}

func (f *fakeStore) Criar(_ context.Context, _ *designv1.Design) (string, error) {
	return f.criarID, f.criarErr
}
func (f *fakeStore) Obter(_ context.Context, _ string) (*designv1.Design, error) {
	return f.obterD, f.obterErr
}
func (f *fakeStore) Listar(_ context.Context, sistemaID string) ([]*designv1.DesignResumo, error) {
	f.sistemaIDDesignsPedido = sistemaID
	return f.listaDesigns, f.listarDesignsErr
}
func (f *fakeStore) Atualizar(_ context.Context, _ *designv1.Design) error { return nil }
func (f *fakeStore) Remover(_ context.Context, _ string) error             { return nil }
func (f *fakeStore) Salvar(_ context.Context, d *designv1.Design) error {
	f.salvou = d
	return f.salvarErr
}
func (f *fakeStore) CriarSistema(_ context.Context, nome string) (string, error) {
	f.sistemaNome = nome
	return f.sistemaID, nil
}
func (f *fakeStore) ListarSistemas(_ context.Context) ([]*designv1.Sistema, error) {
	return f.listaSis, nil
}
func (f *fakeStore) ListarSistemasDoTenant(_ context.Context, tenantID string) ([]*designv1.Sistema, error) {
	f.tenantIDPedido = tenantID
	if f.listarTenantErr != nil {
		return nil, f.listarTenantErr
	}
	return f.listaSisTenant, nil
}
func (f *fakeStore) AtualizarWhiteLabel(_ context.Context, dados store.WhiteLabel) (store.WhiteLabel, error) {
	f.whiteLabelPedido = dados
	if f.whiteLabelErr != nil {
		return store.WhiteLabel{}, f.whiteLabelErr
	}
	return f.whiteLabelOut, nil
}

func comTenant() context.Context { return comTenantTipo("dono") }

func comTenantTipo(tipo string) context.Context {
	return tenantctx.NewContext(context.Background(), &commonv1.TenantContext{TenantId: "t-1", Tipo: tipo})
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

func TestListarDesigns_SemTenant_Unauthenticated(t *testing.T) {
	s := New(&fakeStore{})
	_, err := s.ListarDesigns(context.Background(), &designv1.ListarDesignsRequest{SistemaId: "s1"})
	if code(err) != codes.Unauthenticated {
		t.Fatalf("esperava Unauthenticated; got %v", err)
	}
}

func TestListarDesigns_SemSistemaID_InvalidArgument(t *testing.T) {
	s := New(&fakeStore{})
	_, err := s.ListarDesigns(comTenant(), &designv1.ListarDesignsRequest{})
	if code(err) != codes.InvalidArgument {
		t.Fatalf("esperava InvalidArgument; got %v", err)
	}
}

func TestListarDesigns_OK(t *testing.T) {
	fs := &fakeStore{listaDesigns: []*designv1.DesignResumo{{Id: "d1", Nome: "Home"}, {Id: "d2", Nome: "Cadastro"}}}
	s := New(fs)
	out, err := s.ListarDesigns(comTenant(), &designv1.ListarDesignsRequest{SistemaId: "s1"})
	if err != nil {
		t.Fatalf("inesperado: %v", err)
	}
	if fs.sistemaIDDesignsPedido != "s1" {
		t.Fatalf("sistema_id não propagado: %q", fs.sistemaIDDesignsPedido)
	}
	if len(out.GetDesigns()) != 2 || out.GetDesigns()[0].GetNome() != "Home" {
		t.Fatalf("lista inesperada: %+v", out.GetDesigns())
	}
}

func TestCriarSistema_SemTenant_Unauthenticated(t *testing.T) {
	s := New(&fakeStore{})
	_, err := s.CriarSistema(context.Background(), &designv1.CriarSistemaRequest{Nome: "Demo"})
	if code(err) != codes.Unauthenticated {
		t.Fatalf("esperava Unauthenticated; got %v", err)
	}
}

func TestCriarSistema_Cliente_PermissionDenied(t *testing.T) {
	s := New(&fakeStore{sistemaID: "sis-1"})
	_, err := s.CriarSistema(comTenantTipo("cliente"), &designv1.CriarSistemaRequest{Nome: "Demo"})
	if code(err) != codes.PermissionDenied {
		t.Fatalf("cliente não pode criar; esperava PermissionDenied, got %v", err)
	}
}

func TestCriarSistema_NomeVazio_InvalidArgument(t *testing.T) {
	s := New(&fakeStore{})
	_, err := s.CriarSistema(comTenantTipo("parceiro"), &designv1.CriarSistemaRequest{Nome: ""})
	if code(err) != codes.InvalidArgument {
		t.Fatalf("esperava InvalidArgument; got %v", err)
	}
}

func TestCriarSistema_DonoEParceiro_OK(t *testing.T) {
	for _, tipo := range []string{"dono", "parceiro"} {
		fs := &fakeStore{sistemaID: "sis-9"}
		s := New(fs)
		out, err := s.CriarSistema(comTenantTipo(tipo), &designv1.CriarSistemaRequest{Nome: "Demo"})
		if err != nil {
			t.Fatalf("tipo %s: inesperado: %v", tipo, err)
		}
		if out.GetId() != "sis-9" || out.GetNome() != "Demo" || fs.sistemaNome != "Demo" {
			t.Fatalf("tipo %s: resposta inesperada: %+v", tipo, out)
		}
	}
}

func TestListarSistemas_SemTenant_Unauthenticated(t *testing.T) {
	s := New(&fakeStore{})
	_, err := s.ListarSistemas(context.Background(), &designv1.ListarSistemasRequest{})
	if code(err) != codes.Unauthenticated {
		t.Fatalf("esperava Unauthenticated; got %v", err)
	}
}

func TestListarSistemas_OK(t *testing.T) {
	fs := &fakeStore{listaSis: []*designv1.Sistema{{Id: "a", Nome: "Alfa"}, {Id: "b", Nome: "Beta"}}}
	s := New(fs)
	// Cliente final pode listar (só não pode criar).
	out, err := s.ListarSistemas(comTenantTipo("cliente"), &designv1.ListarSistemasRequest{})
	if err != nil {
		t.Fatalf("inesperado: %v", err)
	}
	if len(out.GetSistemas()) != 2 || out.GetSistemas()[0].GetNome() != "Alfa" {
		t.Fatalf("lista inesperada: %+v", out.GetSistemas())
	}
}

// Com tenant_id preenchido, o Design Engine delega ao store explícito (RF08) —
// SEM revalidar hierarquia aqui (essa é responsabilidade do Gateway via IAM,
// ver comentário do método ListarSistemas).
func TestListarSistemas_ComTenantId_UsaStoreExplicito(t *testing.T) {
	fs := &fakeStore{
		listaSis:       []*designv1.Sistema{{Id: "contexto", Nome: "DoContexto"}},
		listaSisTenant: []*designv1.Sistema{{Id: "filho", Nome: "DoFilho"}},
	}
	s := New(fs)
	out, err := s.ListarSistemas(comTenant(), &designv1.ListarSistemasRequest{TenantId: "tenant-filho"})
	if err != nil {
		t.Fatalf("inesperado: %v", err)
	}
	if fs.tenantIDPedido != "tenant-filho" {
		t.Fatalf("tenant_id não propagado ao store: %q", fs.tenantIDPedido)
	}
	if len(out.GetSistemas()) != 1 || out.GetSistemas()[0].GetId() != "filho" {
		t.Fatalf("deveria usar a lista do tenant explícito, não a do contexto: %+v", out.GetSistemas())
	}
}

func TestAtualizarWhiteLabel_OK(t *testing.T) {
	fs := &fakeStore{whiteLabelOut: store.WhiteLabel{LogoURL: "https://x/logo.png", CorPrimaria: "#111111", DominioValidado: false}}
	s := New(fs)
	out, err := s.AtualizarWhiteLabel(comTenant(), &designv1.AtualizarWhiteLabelRequest{LogoUrl: "https://x/logo.png", CorPrimaria: "#111111"})
	if err != nil {
		t.Fatalf("inesperado: %v", err)
	}
	if out.GetLogoUrl() != "https://x/logo.png" || out.GetDominioValidado() {
		t.Fatalf("resposta inesperada: %+v", out)
	}
}

func TestAtualizarWhiteLabel_SemTenant_Unauthenticated(t *testing.T) {
	s := New(&fakeStore{})
	_, err := s.AtualizarWhiteLabel(context.Background(), &designv1.AtualizarWhiteLabelRequest{})
	if code(err) != codes.Unauthenticated {
		t.Fatalf("esperava Unauthenticated; got %v", err)
	}
}
