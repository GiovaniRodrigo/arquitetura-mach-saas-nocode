// Package server implementa o IAMService gRPC (RF03, RN03).
package server

import (
	"context"
	"crypto/rand"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	iamv1 "github.com/machv4/platform/gen/go/construtor/iam/v1"
	"github.com/machv4/platform/pkg/tenantctx"
	"github.com/machv4/platform/services/iam/auth"
	"github.com/machv4/platform/services/iam/internal/permissions"
	"github.com/machv4/platform/services/iam/internal/store"
)

// Store carrega permissões e materializa identidades e tenants (satisfeito por
// store.Store).
type Store interface {
	PermissoesDe(ctx context.Context, blindIndexes []string) ([]permissions.Permissao, error)
	UpsertUsuarioThirdParty(ctx context.Context, provedor, externalID, email, nome string) (userID, tenantID, tipo string, err error)
	ListarFilhos(ctx context.Context, parentID string) ([]store.Tenant, error)
	CriarTenant(ctx context.Context, nome, tipo string, parentID *string, chaveBlindIndex []byte) (store.Tenant, error)
}

// IAMServer implementa iamv1.IAMServiceServer.
type IAMServer struct {
	iamv1.UnimplementedIAMServiceServer
	validator *auth.Validator
	issuer    *auth.Issuer
	store     Store
	eval      permissions.Evaluator
}

// New cria o servidor com o validador/emissor de JWT e o store de identidades e
// permissões. O issuer detém a chave privada — só o IAM emite tokens (RF03).
func New(validator *auth.Validator, issuer *auth.Issuer, store Store) *IAMServer {
	return &IAMServer{validator: validator, issuer: issuer, store: store}
}

// ValidarToken verifica o JWT e devolve a identidade. Token inválido não é erro
// gRPC: responde valido=false para o Gateway decidir o 401 (RF03).
func (s *IAMServer) ValidarToken(_ context.Context, req *iamv1.ValidarTokenRequest) (*iamv1.ValidarTokenResponse, error) {
	claims, err := s.validator.Validate(req.GetJwt())
	if err != nil {
		return &iamv1.ValidarTokenResponse{Valido: false}, nil
	}
	return &iamv1.ValidarTokenResponse{
		Valido:   true,
		TenantId: claims.TenantID,
		UserId:   claims.Subject,
		Tipo:     claims.Tipo,
	}, nil
}

// AutenticarThirdParty materializa a identidade de um login social já validado
// pelo Gateway (que conduziu o fluxo OAuth com o provedor) e emite o JWT MACH.
// Faz find-or-create do usuário e assina o token com a identidade resultante.
func (s *IAMServer) AutenticarThirdParty(ctx context.Context, req *iamv1.AutenticarThirdPartyRequest) (*iamv1.AutenticarThirdPartyResponse, error) {
	if req.GetProvedor() == "" || req.GetExternalId() == "" {
		return nil, status.Error(codes.InvalidArgument, "provedor e external_id são obrigatórios")
	}
	userID, tenantID, tipo, err := s.store.UpsertUsuarioThirdParty(ctx, req.GetProvedor(), req.GetExternalId(), req.GetEmail(), req.GetNome())
	if err != nil {
		return nil, status.Error(codes.Internal, "falha ao registrar usuário")
	}
	token, err := s.issuer.Issue(userID, tenantID, tipo)
	if err != nil {
		return nil, status.Error(codes.Internal, "falha ao emitir token")
	}
	return &iamv1.AutenticarThirdPartyResponse{
		Jwt:      token,
		UserId:   userID,
		TenantId: tenantID,
		Tipo:     tipo,
	}, nil
}

// AvaliarPermissoes computa o mapa booleano por componente para o tenant do
// contexto. A lógica das regras nunca sai daqui (RN03); o cliente recebe só o
// resultado. O tenant vem do TenantContext (Metadata gRPC), nunca do request.
func (s *IAMServer) AvaliarPermissoes(ctx context.Context, req *iamv1.AvaliarPermissoesRequest) (*iamv1.AvaliarPermissoesResponse, error) {
	tc, err := tenantctx.Require(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "contexto de tenant ausente")
	}

	perms, err := s.store.PermissoesDe(ctx, req.GetBlindIndexes())
	if err != nil {
		return nil, status.Error(codes.Internal, "falha ao carregar permissões")
	}

	// Nesta fase o papel do sujeito equivale ao tipo do token; um modelo de
	// papéis por usuário virá quando existir tabela de usuários.
	suj := permissions.Sujeito{Papel: tc.GetTipo(), Tipo: tc.GetTipo()}
	decisoes := s.eval.Avaliar(suj, perms, req.GetBlindIndexes())

	out := make(map[string]*iamv1.PermissaoComponente, len(decisoes))
	for bi, d := range decisoes {
		out[bi] = &iamv1.PermissaoComponente{View: d.View, Click: d.Click}
	}
	return &iamv1.AvaliarPermissoesResponse{Permissions: out}, nil
}

// ListarTenants devolve os tenants filhos diretos do tenant do contexto — os
// clientes/negócios que o Dono/Parceiro autenticado gerencia (spec 004, RF07,
// RN05: a hierarquia é sempre Tenant → filhos, nunca lateral).
func (s *IAMServer) ListarTenants(ctx context.Context, _ *iamv1.ListarTenantsRequest) (*iamv1.ListarTenantsResponse, error) {
	tc, err := tenantctx.Require(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "contexto de tenant ausente")
	}
	filhos, err := s.store.ListarFilhos(ctx, tc.GetTenantId())
	if err != nil {
		return nil, status.Error(codes.Internal, "falha ao listar tenants")
	}
	out := make([]*iamv1.Tenant, 0, len(filhos))
	for _, t := range filhos {
		out = append(out, &iamv1.Tenant{Id: t.ID, Nome: t.Nome, Tipo: t.Tipo})
	}
	return &iamv1.ListarTenantsResponse{Tenants: out}, nil
}

// CriarTenant cria um novo tenant "cliente" sob o tenant do contexto (spec 004,
// RF07). Restrito a dono/parceiro, mesma regra de CriarSistema (001, RN01) —
// um cliente final não gerencia outros tenants.
func (s *IAMServer) CriarTenant(ctx context.Context, req *iamv1.CriarTenantRequest) (*iamv1.Tenant, error) {
	tc, err := tenantctx.Require(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "contexto de tenant ausente")
	}
	if tc.GetTipo() != "dono" && tc.GetTipo() != "parceiro" {
		return nil, status.Error(codes.PermissionDenied, "apenas dono ou parceiro podem criar clientes")
	}
	if req.GetNome() == "" {
		return nil, status.Error(codes.InvalidArgument, "nome obrigatório")
	}
	chave := make([]byte, 32)
	if _, err := rand.Read(chave); err != nil {
		return nil, status.Error(codes.Internal, "falha ao gerar chave do tenant")
	}
	parentID := tc.GetTenantId()
	t, err := s.store.CriarTenant(ctx, req.GetNome(), "cliente", &parentID, chave)
	if err != nil {
		return nil, status.Error(codes.Internal, "falha ao criar tenant")
	}
	return &iamv1.Tenant{Id: t.ID, Nome: t.Nome, Tipo: t.Tipo}, nil
}
