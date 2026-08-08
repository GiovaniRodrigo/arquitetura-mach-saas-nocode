package server

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonv1 "github.com/machv4/platform/gen/go/construtor/common/v1"
	deployv1 "github.com/machv4/platform/gen/go/construtor/deploy/v1"
	"github.com/machv4/platform/pkg/tenantctx"
	"github.com/machv4/platform/services/deploy/internal/versions"
)

type fakeMgr struct {
	pub          versions.Publicacao
	pubErr       error
	rbAtiva      int32
	rbErr        error
	ativa        versions.VersaoAtiva
	ativaErr     error
	listVersoes  []versions.VersaoResumo
	listErr      error
	rbPorIDAtiva int32
	rbPorIDErr   error
}

func (f *fakeMgr) Publicar(context.Context, string) (versions.Publicacao, error) {
	return f.pub, f.pubErr
}
func (f *fakeMgr) Rollback(context.Context, string, int32) (int32, error) {
	return f.rbAtiva, f.rbErr
}
func (f *fakeMgr) ObterAtiva(context.Context, string) (versions.VersaoAtiva, error) {
	return f.ativa, f.ativaErr
}
func (f *fakeMgr) ListarVersoes(context.Context, string) ([]versions.VersaoResumo, error) {
	return f.listVersoes, f.listErr
}
func (f *fakeMgr) RollbackPorID(context.Context, string, string) (int32, error) {
	return f.rbPorIDAtiva, f.rbPorIDErr
}

func comTenant() context.Context {
	return tenantctx.NewContext(context.Background(), &commonv1.TenantContext{TenantId: "t-1", Tipo: "dono"})
}

func code(err error) codes.Code { return status.Code(err) }

func TestPublicar_SemTenant_Unauthenticated(t *testing.T) {
	s := New(&fakeMgr{})
	_, err := s.Publicar(context.Background(), &deployv1.PublicarRequest{SistemaId: "s1"})
	if code(err) != codes.Unauthenticated {
		t.Fatalf("esperava Unauthenticated; got %v", err)
	}
}

func TestPublicar_Concorrente_Aborted(t *testing.T) {
	s := New(&fakeMgr{pubErr: versions.ErrPublicacaoConcorrente})
	_, err := s.Publicar(comTenant(), &deployv1.PublicarRequest{SistemaId: "s1"})
	if code(err) != codes.Aborted {
		t.Fatalf("esperava Aborted; got %v", err)
	}
}

func TestPublicar_OK(t *testing.T) {
	s := New(&fakeMgr{pub: versions.Publicacao{VersaoID: "v-1", Numero: 7}})
	out, err := s.Publicar(comTenant(), &deployv1.PublicarRequest{SistemaId: "s1"})
	if err != nil {
		t.Fatalf("inesperado: %v", err)
	}
	if out.GetVersaoId() != "v-1" || out.GetNumero() != 7 {
		t.Fatalf("resposta inesperada: %+v", out)
	}
}

func TestRollback_SemAnterior_FailedPrecondition(t *testing.T) {
	s := New(&fakeMgr{rbErr: versions.ErrSemVersaoAnterior})
	_, err := s.Rollback(comTenant(), &deployv1.RollbackRequest{SistemaId: "s1"})
	if code(err) != codes.FailedPrecondition {
		t.Fatalf("esperava FailedPrecondition; got %v", err)
	}
}

func TestObterVersaoAtiva_Inexistente_NotFound(t *testing.T) {
	s := New(&fakeMgr{ativaErr: versions.ErrSemVersaoAtiva})
	_, err := s.ObterVersaoAtiva(comTenant(), &deployv1.ObterVersaoAtivaRequest{SistemaId: "s1"})
	if code(err) != codes.NotFound {
		t.Fatalf("esperava NotFound; got %v", err)
	}
}

func TestListarVersoes_SemTenant_Unauthenticated(t *testing.T) {
	s := New(&fakeMgr{})
	_, err := s.ListarVersoes(context.Background(), &deployv1.ListarVersoesRequest{SistemaId: "s1"})
	if code(err) != codes.Unauthenticated {
		t.Fatalf("esperava Unauthenticated; got %v", err)
	}
}

func TestListarVersoes_OK(t *testing.T) {
	criado := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	s := New(&fakeMgr{listVersoes: []versions.VersaoResumo{
		{ID: "v-2", Numero: 2, Ativa: true, CriadoEm: criado},
		{ID: "v-1", Numero: 1, Ativa: false, CriadoEm: criado},
	}})
	out, err := s.ListarVersoes(comTenant(), &deployv1.ListarVersoesRequest{SistemaId: "s1"})
	if err != nil {
		t.Fatalf("inesperado: %v", err)
	}
	if len(out.GetVersoes()) != 2 {
		t.Fatalf("esperava 2 versões; got %d", len(out.GetVersoes()))
	}
	if out.GetVersoes()[0].GetId() != "v-2" || out.GetVersoes()[0].GetAtiva() != true {
		t.Fatalf("resposta inesperada: %+v", out.GetVersoes()[0])
	}
	if out.GetVersoes()[0].GetCriadoEm() != criado.Format(time.RFC3339) {
		t.Fatalf("criado_em não formatado como RFC3339: %v", out.GetVersoes()[0].GetCriadoEm())
	}
}

func TestPublicarVersao_OK(t *testing.T) {
	s := New(&fakeMgr{pub: versions.Publicacao{VersaoID: "v-3", Numero: 3}})
	out, err := s.PublicarVersao(comTenant(), &deployv1.PublicarVersaoRequest{SistemaId: "s1"})
	if err != nil {
		t.Fatalf("inesperado: %v", err)
	}
	if out.GetVersaoId() != "v-3" || out.GetNumero() != 3 {
		t.Fatalf("resposta inesperada: %+v", out)
	}
}

func TestReverterVersao_SemVersao_FailedPrecondition(t *testing.T) {
	s := New(&fakeMgr{rbPorIDErr: versions.ErrSemVersaoAnterior})
	_, err := s.ReverterVersao(comTenant(), &deployv1.ReverterVersaoRequest{SistemaId: "s1", VersaoId: "v-x"})
	if code(err) != codes.FailedPrecondition {
		t.Fatalf("esperava FailedPrecondition; got %v", err)
	}
}

func TestReverterVersao_OK(t *testing.T) {
	s := New(&fakeMgr{rbPorIDAtiva: 4})
	out, err := s.ReverterVersao(comTenant(), &deployv1.ReverterVersaoRequest{SistemaId: "s1", VersaoId: "v-4"})
	if err != nil {
		t.Fatalf("inesperado: %v", err)
	}
	if out.GetVersaoAtiva() != 4 {
		t.Fatalf("resposta inesperada: %+v", out)
	}
}
