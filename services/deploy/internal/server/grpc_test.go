package server

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonv1 "github.com/machv4/platform/gen/go/construtor/common/v1"
	deployv1 "github.com/machv4/platform/gen/go/construtor/deploy/v1"
	"github.com/machv4/platform/pkg/tenantctx"
	"github.com/machv4/platform/services/deploy/internal/versions"
)

type fakeMgr struct {
	pub      versions.Publicacao
	pubErr   error
	rbAtiva  int32
	rbErr    error
	ativa    versions.VersaoAtiva
	ativaErr error
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
