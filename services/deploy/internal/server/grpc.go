// Package server implementa o DeployEngineService gRPC: publicação por flag ativa
// e rollback instantâneo (RF04, RN04, RN05).
package server

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	deployv1 "github.com/machv4/platform/gen/go/construtor/deploy/v1"
	"github.com/machv4/platform/pkg/database"
	"github.com/machv4/platform/pkg/tenantctx"
	"github.com/machv4/platform/services/deploy/internal/versions"
)

// Gerenciador é o subconjunto do versions.Manager usado pelo servidor (facilita
// testes).
type Gerenciador interface {
	Publicar(ctx context.Context, sistemaID string) (versions.Publicacao, error)
	Rollback(ctx context.Context, sistemaID string, versaoNumero int32) (int32, error)
	ObterAtiva(ctx context.Context, sistemaID string) (versions.VersaoAtiva, error)
}

// DeployServer implementa deployv1.DeployEngineServiceServer.
type DeployServer struct {
	deployv1.UnimplementedDeployEngineServiceServer
	mgr Gerenciador
}

// New cria o servidor sobre o gerenciador de versões.
func New(mgr Gerenciador) *DeployServer {
	return &DeployServer{mgr: mgr}
}

// Publicar cria e ativa uma nova versão do sistema (RN04).
func (s *DeployServer) Publicar(ctx context.Context, req *deployv1.PublicarRequest) (*deployv1.PublicarResponse, error) {
	if err := exigirTenant(ctx); err != nil {
		return nil, err
	}
	if req.GetSistemaId() == "" {
		return nil, status.Error(codes.InvalidArgument, "sistema_id obrigatório")
	}
	pub, err := s.mgr.Publicar(ctx, req.GetSistemaId())
	if err != nil {
		return nil, mapErr(err)
	}
	return &deployv1.PublicarResponse{VersaoId: pub.VersaoID, Numero: pub.Numero}, nil
}

// Rollback reativa a versão anterior (ou a indicada) do sistema (RN05).
func (s *DeployServer) Rollback(ctx context.Context, req *deployv1.RollbackRequest) (*deployv1.RollbackResponse, error) {
	if err := exigirTenant(ctx); err != nil {
		return nil, err
	}
	if req.GetSistemaId() == "" {
		return nil, status.Error(codes.InvalidArgument, "sistema_id obrigatório")
	}
	ativa, err := s.mgr.Rollback(ctx, req.GetSistemaId(), req.GetVersaoNumero())
	if err != nil {
		return nil, mapErr(err)
	}
	return &deployv1.RollbackResponse{VersaoAtiva: ativa}, nil
}

// ObterVersaoAtiva devolve a definição consolidada da versão ativa (RN04).
func (s *DeployServer) ObterVersaoAtiva(ctx context.Context, req *deployv1.ObterVersaoAtivaRequest) (*deployv1.VersaoAtiva, error) {
	if err := exigirTenant(ctx); err != nil {
		return nil, err
	}
	if req.GetSistemaId() == "" {
		return nil, status.Error(codes.InvalidArgument, "sistema_id obrigatório")
	}
	v, err := s.mgr.ObterAtiva(ctx, req.GetSistemaId())
	if err != nil {
		return nil, mapErr(err)
	}
	return &deployv1.VersaoAtiva{VersaoId: v.VersaoID, Numero: v.Numero, DefinicaoJson: v.DefinicaoJSON}, nil
}

func exigirTenant(ctx context.Context) error {
	if _, err := tenantctx.Require(ctx); err != nil {
		return status.Error(codes.Unauthenticated, "contexto de tenant ausente")
	}
	return nil
}

// mapErr traduz os erros do gerenciador em códigos gRPC. Aborted (concorrência)
// vira 409 no Gateway; FailedPrecondition (sem versão anterior) vira 422.
func mapErr(err error) error {
	switch {
	case errors.Is(err, versions.ErrPublicacaoConcorrente):
		return status.Error(codes.Aborted, "publicação concorrente")
	case errors.Is(err, versions.ErrSemVersaoAnterior):
		return status.Error(codes.FailedPrecondition, "sem versão anterior")
	case errors.Is(err, versions.ErrSemVersaoAtiva), errors.Is(err, versions.ErrSistemaInexistente):
		return status.Error(codes.NotFound, "recurso inexistente")
	case errors.Is(err, database.ErrNoTenant):
		return status.Error(codes.Unauthenticated, "contexto de tenant ausente")
	default:
		return status.Error(codes.Internal, "erro interno")
	}
}
