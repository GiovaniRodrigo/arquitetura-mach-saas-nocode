// Package server implementa o LogicEngineService gRPC: revalidação e persistência
// de formulários (RF07, RN08) e o CRUD de regras de negócio (RF02).
package server

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	logicv1 "github.com/machv4/platform/gen/go/construtor/logic/v1"
	"github.com/machv4/platform/pkg/database"
	"github.com/machv4/platform/pkg/tenantctx"
	"github.com/machv4/platform/services/logic/internal/rules"
	"github.com/machv4/platform/services/logic/internal/store"
	"github.com/machv4/platform/services/logic/internal/validation"
)

// Persistencia é o subconjunto do store usado pelo servidor (facilita testes).
type Persistencia interface {
	SchemaDe(ctx context.Context, sistemaID string) (map[string]validation.CampoDef, error)
	InserirDados(ctx context.Context, sistemaID string, valores map[string]string) (string, error)
	CriarRegra(ctx context.Context, r store.Regra) (string, error)
	ObterRegra(ctx context.Context, id string) (store.Regra, error)
	AtualizarRegra(ctx context.Context, r store.Regra) error
	RemoverRegra(ctx context.Context, id string) error
}

// LogicServer implementa logicv1.LogicEngineServiceServer.
type LogicServer struct {
	logicv1.UnimplementedLogicEngineServiceServer
	store Persistencia
}

// New cria o servidor sobre a camada de persistência.
func New(store Persistencia) *LogicServer {
	return &LogicServer{store: store}
}

// SalvarFormulario revalida o payload contra o schema salvo (RN08) e só persiste
// se não houver erros. A validação que falha não é erro gRPC: devolve
// sucesso=false + mapa de erros por blind_index, para o Gateway responder 422 sem
// jamais expor nomes reais de campos (RNF08, critério 2).
func (s *LogicServer) SalvarFormulario(ctx context.Context, req *logicv1.SalvarFormularioRequest) (*logicv1.SalvarFormularioResponse, error) {
	if err := exigirTenant(ctx); err != nil {
		return nil, err
	}
	if req.GetSistemaId() == "" {
		return nil, status.Error(codes.InvalidArgument, "sistema_id obrigatório")
	}

	schema, err := s.store.SchemaDe(ctx, req.GetSistemaId())
	if err != nil {
		return nil, mapErr(err)
	}

	if erros := validation.Validar(schema, req.GetDadosFormulario()); len(erros) > 0 {
		return &logicv1.SalvarFormularioResponse{
			Sucesso:        false,
			ErrosValidacao: erros,
			MensagemStatus: "Falha de validação",
		}, nil
	}

	if _, err := s.store.InserirDados(ctx, req.GetSistemaId(), req.GetDadosFormulario()); err != nil {
		return nil, mapErr(err)
	}
	return &logicv1.SalvarFormularioResponse{Sucesso: true, MensagemStatus: "Registo gravado"}, nil
}

// CriarRegra valida a árvore de decisão e persiste uma nova regra.
func (s *LogicServer) CriarRegra(ctx context.Context, req *logicv1.Regra) (*logicv1.Regra, error) {
	if err := exigirTenant(ctx); err != nil {
		return nil, err
	}
	if req.GetSistemaId() == "" {
		return nil, status.Error(codes.InvalidArgument, "sistema_id obrigatório")
	}
	if err := validarArvore(req.GetArvoreDecisao()); err != nil {
		return nil, err
	}
	id, err := s.store.CriarRegra(ctx, store.Regra{SistemaID: req.GetSistemaId(), ArvoreDecisao: req.GetArvoreDecisao()})
	if err != nil {
		return nil, mapErr(err)
	}
	return &logicv1.Regra{Id: id, SistemaId: req.GetSistemaId(), ArvoreDecisao: req.GetArvoreDecisao()}, nil
}

// ObterRegra devolve uma regra por id, restrita ao tenant do contexto.
func (s *LogicServer) ObterRegra(ctx context.Context, req *logicv1.ObterRegraRequest) (*logicv1.Regra, error) {
	if err := exigirTenant(ctx); err != nil {
		return nil, err
	}
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id obrigatório")
	}
	r, err := s.store.ObterRegra(ctx, req.GetId())
	if err != nil {
		return nil, mapErr(err)
	}
	return &logicv1.Regra{Id: r.ID, SistemaId: r.SistemaID, ArvoreDecisao: r.ArvoreDecisao}, nil
}

// AtualizarRegra revalida a árvore e sobrescreve uma regra existente.
func (s *LogicServer) AtualizarRegra(ctx context.Context, req *logicv1.Regra) (*logicv1.Regra, error) {
	if err := exigirTenant(ctx); err != nil {
		return nil, err
	}
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id obrigatório")
	}
	if err := validarArvore(req.GetArvoreDecisao()); err != nil {
		return nil, err
	}
	if err := s.store.AtualizarRegra(ctx, store.Regra{ID: req.GetId(), ArvoreDecisao: req.GetArvoreDecisao()}); err != nil {
		return nil, mapErr(err)
	}
	return req, nil
}

// RemoverRegra apaga uma regra do tenant corrente.
func (s *LogicServer) RemoverRegra(ctx context.Context, req *logicv1.ObterRegraRequest) (*logicv1.RemoverResponse, error) {
	if err := exigirTenant(ctx); err != nil {
		return nil, err
	}
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id obrigatório")
	}
	if err := s.store.RemoverRegra(ctx, req.GetId()); err != nil {
		return nil, mapErr(err)
	}
	return &logicv1.RemoverResponse{Sucesso: true}, nil
}

// validarArvore garante que a árvore de decisão é bem formada antes de persistir.
func validarArvore(raw []byte) error {
	arvore, err := rules.Parse(raw)
	if err != nil {
		return status.Error(codes.InvalidArgument, "árvore de decisão inválida")
	}
	if err := rules.Validar(arvore); err != nil {
		return status.Error(codes.InvalidArgument, "árvore de decisão inválida")
	}
	return nil
}

func exigirTenant(ctx context.Context) error {
	if _, err := tenantctx.Require(ctx); err != nil {
		return status.Error(codes.Unauthenticated, "contexto de tenant ausente")
	}
	return nil
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, store.ErrNaoEncontrado):
		return status.Error(codes.NotFound, "registro não encontrado")
	case errors.Is(err, database.ErrNoTenant):
		return status.Error(codes.Unauthenticated, "contexto de tenant ausente")
	default:
		return status.Error(codes.Internal, "erro interno")
	}
}
