// Package server implementa o DesignEngineService gRPC: CRUD da árvore Composite
// e a persistência em lote da colaboração (RF01, RN06).
package server

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	designv1 "github.com/machv4/platform/gen/go/construtor/design/v1"
	"github.com/machv4/platform/pkg/database"
	"github.com/machv4/platform/pkg/tenantctx"
	"github.com/machv4/platform/services/design/internal/store"
	"github.com/machv4/platform/services/design/internal/tree"
)

// Persistencia é o subconjunto do store usado pelo servidor (facilita testes).
type Persistencia interface {
	Criar(ctx context.Context, d *designv1.Design) (string, error)
	Obter(ctx context.Context, id string) (*designv1.Design, error)
	Atualizar(ctx context.Context, d *designv1.Design) error
	Remover(ctx context.Context, id string) error
	Salvar(ctx context.Context, d *designv1.Design) error
	CriarSistema(ctx context.Context, nome string) (string, error)
	ListarSistemas(ctx context.Context) ([]*designv1.Sistema, error)
	ListarSistemasDoTenant(ctx context.Context, tenantID string) ([]*designv1.Sistema, error)
	AtualizarWhiteLabel(ctx context.Context, dados store.WhiteLabel) (store.WhiteLabel, error)
}

// DesignServer implementa designv1.DesignEngineServiceServer.
type DesignServer struct {
	designv1.UnimplementedDesignEngineServiceServer
	store Persistencia
}

// New cria o servidor sobre a camada de persistência.
func New(store Persistencia) *DesignServer {
	return &DesignServer{store: store}
}

// CriarDesign valida a árvore e persiste um novo design no tenant do contexto.
func (s *DesignServer) CriarDesign(ctx context.Context, d *designv1.Design) (*designv1.Design, error) {
	if err := exigirTenant(ctx); err != nil {
		return nil, err
	}
	if d.GetSistemaId() == "" {
		return nil, status.Error(codes.InvalidArgument, "sistema_id obrigatório")
	}
	if err := tree.Validar(d.GetArvore()); err != nil {
		return nil, status.Error(codes.InvalidArgument, "árvore recursiva inválida")
	}
	id, err := s.store.Criar(ctx, d)
	if err != nil {
		return nil, mapErr(err)
	}
	return &designv1.Design{Id: id, SistemaId: d.GetSistemaId(), Nome: d.GetNome(), Arvore: d.GetArvore()}, nil
}

// ObterDesign devolve um design por id, restrito ao tenant do contexto.
func (s *DesignServer) ObterDesign(ctx context.Context, req *designv1.ObterDesignRequest) (*designv1.Design, error) {
	if err := exigirTenant(ctx); err != nil {
		return nil, err
	}
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id obrigatório")
	}
	d, err := s.store.Obter(ctx, req.GetId())
	if err != nil {
		return nil, mapErr(err)
	}
	return d, nil
}

// AtualizarDesign revalida a árvore e sobrescreve um design existente.
func (s *DesignServer) AtualizarDesign(ctx context.Context, d *designv1.Design) (*designv1.Design, error) {
	if err := exigirTenant(ctx); err != nil {
		return nil, err
	}
	if d.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id obrigatório")
	}
	if err := tree.Validar(d.GetArvore()); err != nil {
		return nil, status.Error(codes.InvalidArgument, "árvore recursiva inválida")
	}
	if err := s.store.Atualizar(ctx, d); err != nil {
		return nil, mapErr(err)
	}
	return d, nil
}

// RemoverDesign apaga um design do tenant corrente.
func (s *DesignServer) RemoverDesign(ctx context.Context, req *designv1.ObterDesignRequest) (*designv1.RemoverResponse, error) {
	if err := exigirTenant(ctx); err != nil {
		return nil, err
	}
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id obrigatório")
	}
	if err := s.store.Remover(ctx, req.GetId()); err != nil {
		return nil, mapErr(err)
	}
	return &designv1.RemoverResponse{Sucesso: true}, nil
}

// SalvarDesign é o flush único da colaboração write-behind (RN06): recebe o
// estado consolidado de um design (id conhecido) e faz upsert.
func (s *DesignServer) SalvarDesign(ctx context.Context, req *designv1.SalvarDesignRequest) (*designv1.SalvarDesignResponse, error) {
	if err := exigirTenant(ctx); err != nil {
		return nil, err
	}
	d := req.GetDesign()
	if d.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id obrigatório para persistência em lote")
	}
	if err := tree.Validar(d.GetArvore()); err != nil {
		return nil, status.Error(codes.InvalidArgument, "árvore recursiva inválida")
	}
	if err := s.store.Salvar(ctx, d); err != nil {
		return nil, mapErr(err)
	}
	return &designv1.SalvarDesignResponse{Sucesso: true}, nil
}

// CriarSistema cria um sistema no tenant do contexto. A criação é restrita a
// dono/parceiro; cliente final apenas lista (RN01, RN02).
func (s *DesignServer) CriarSistema(ctx context.Context, req *designv1.CriarSistemaRequest) (*designv1.Sistema, error) {
	tc, err := tenantctx.Require(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "contexto de tenant ausente")
	}
	if tc.GetTipo() != "dono" && tc.GetTipo() != "parceiro" {
		return nil, status.Error(codes.PermissionDenied, "apenas dono ou parceiro podem criar sistema")
	}
	if req.GetNome() == "" {
		return nil, status.Error(codes.InvalidArgument, "nome obrigatório")
	}
	id, err := s.store.CriarSistema(ctx, req.GetNome())
	if err != nil {
		return nil, mapErr(err)
	}
	return &designv1.Sistema{Id: id, Nome: req.GetNome()}, nil
}

// ListarSistemas devolve os sistemas do tenant do contexto (aberto a qualquer
// autenticado; a RLS isola por tenant).
//
// Fronteira de confiança (RF08/RN05): quando req.tenant_id vem preenchido, o
// Design Engine lista os sistemas DESSE tenant em vez do tenant do contexto,
// SEM validar aqui se ele é filho do tenant do contexto — essa validação de
// hierarquia é responsabilidade exclusiva do Gateway, que a faz via
// IAMServiceClient.ObterTenant (RN05) ANTES de chamar esta RPC. O Design
// Engine confia que qualquer chamador desta RPC com tenant_id preenchido já
// autorizou o acesso cross-tenant. Nunca exponha esta RPC diretamente a um
// cliente não confiável sem essa validação prévia.
func (s *DesignServer) ListarSistemas(ctx context.Context, req *designv1.ListarSistemasRequest) (*designv1.ListarSistemasResponse, error) {
	if err := exigirTenant(ctx); err != nil {
		return nil, err
	}
	var (
		sistemas []*designv1.Sistema
		err      error
	)
	if tid := req.GetTenantId(); tid != "" {
		sistemas, err = s.store.ListarSistemasDoTenant(ctx, tid)
	} else {
		sistemas, err = s.store.ListarSistemas(ctx)
	}
	if err != nil {
		return nil, mapErr(err)
	}
	return &designv1.ListarSistemasResponse{Sistemas: sistemas}, nil
}

// AtualizarWhiteLabel faz upsert da personalização de marca (logo/cores/domínio
// próprio) do tenant do contexto (RF13, RNF03). A validação do domínio é
// assíncrona e fora de escopo: dominio_validado volta sempre false.
func (s *DesignServer) AtualizarWhiteLabel(ctx context.Context, req *designv1.AtualizarWhiteLabelRequest) (*designv1.WhiteLabel, error) {
	if err := exigirTenant(ctx); err != nil {
		return nil, err
	}
	out, err := s.store.AtualizarWhiteLabel(ctx, store.WhiteLabel{
		LogoURL:        req.GetLogoUrl(),
		CorPrimaria:    req.GetCorPrimaria(),
		CorSecundaria:  req.GetCorSecundaria(),
		DominioProprio: req.GetDominioProprio(),
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return &designv1.WhiteLabel{
		LogoUrl:         out.LogoURL,
		CorPrimaria:     out.CorPrimaria,
		CorSecundaria:   out.CorSecundaria,
		DominioProprio:  out.DominioProprio,
		DominioValidado: out.DominioValidado,
	}, nil
}

// exigirTenant rejeita chamadas sem TenantContext no Metadata (RN01, RNF02).
func exigirTenant(ctx context.Context) error {
	if _, err := tenantctx.Require(ctx); err != nil {
		return status.Error(codes.Unauthenticated, "contexto de tenant ausente")
	}
	return nil
}

// mapErr traduz erros da persistência em códigos gRPC.
func mapErr(err error) error {
	switch {
	case errors.Is(err, store.ErrNaoEncontrado):
		return status.Error(codes.NotFound, "design não encontrado")
	case errors.Is(err, database.ErrNoTenant):
		return status.Error(codes.Unauthenticated, "contexto de tenant ausente")
	default:
		return status.Error(codes.Internal, "erro interno")
	}
}
