// Package server implementa o IAMService gRPC (RF03, RN03).
package server

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	iamv1 "github.com/machv4/platform/gen/go/construtor/iam/v1"
	"github.com/machv4/platform/pkg/tenantctx"
	"github.com/machv4/platform/services/iam/auth"
	"github.com/machv4/platform/services/iam/internal/permissions"
)

// Store carrega permissões e materializa identidades (satisfeito por store.Store).
type Store interface {
	PermissoesDe(ctx context.Context, blindIndexes []string) ([]permissions.Permissao, error)
	UpsertUsuarioThirdParty(ctx context.Context, provedor, externalID, email, nome string) (userID, tenantID, tipo string, err error)
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
