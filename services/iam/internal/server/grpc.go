// Package server implementa o IAMService gRPC (RF03, RN03).
package server

import (
	"context"
	"crypto/rand"
	"errors"
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	iamv1 "github.com/machv4/platform/gen/go/construtor/iam/v1"
	"github.com/machv4/platform/pkg/tenantctx"
	"github.com/machv4/platform/services/iam/auth"
	"github.com/machv4/platform/services/iam/internal/permissions"
	"github.com/machv4/platform/services/iam/internal/store"
)

// bcryptCost isolado numa constante (RNF01): fácil de subir depois sem
// migração de dados, já que o custo fica embutido no próprio hash.
const bcryptCost = 12

// erroCredenciaisInvalidas é devolvido tanto para e-mail inexistente quanto
// para senha incorreta (RN04) — nunca revela qual dos dois falhou.
var erroCredenciaisInvalidas = status.Error(codes.Unauthenticated, "credenciais inválidas")

// Store carrega permissões e materializa identidades e tenants (satisfeito por
// store.Store).
type Store interface {
	PermissoesDe(ctx context.Context, blindIndexes []string) ([]permissions.Permissao, error)
	UpsertUsuarioThirdParty(ctx context.Context, provedor, externalID, email, nome string) (userID, tenantID, tipo string, err error)
	ListarFilhos(ctx context.Context, parentID string) ([]store.Tenant, error)
	CriarTenant(ctx context.Context, nome, tipo string, parentID *string, chaveBlindIndex []byte) (store.Tenant, error)
	ObterTenant(ctx context.Context, id string) (store.Tenant, error)
	AtualizarTenant(ctx context.Context, id, nome string) (store.Tenant, error)
	ExcluirTenant(ctx context.Context, id string) error
	CriarTenantEUsuarioComSenha(ctx context.Context, nomeUsuario, email, senhaHash, nomeTenant string) (userID, tenantID string, err error)
	ObterUsuarioPorEmailSenha(ctx context.Context, email string) (userID, tenantID, tipo, senhaHash string, err error)
	RegistrarEventoLogin(ctx context.Context, usuarioID, tenantID string) error
	UltimosAcessos(ctx context.Context, tenantID string, limite int) ([]store.EventoLogin, error)
	ListarFeedback(ctx context.Context, tenantID string, status *string) ([]store.Feedback, error)
	AtualizarStatusFeedback(ctx context.Context, id, novoStatus string) (store.Feedback, error)
	ResumoFinanceiro(ctx context.Context, tenantID string) (store.ResumoFinanceiro, error)
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
	if err := s.store.RegistrarEventoLogin(ctx, userID, tenantID); err != nil {
		log.Printf("iam: falha ao registrar evento de login (third-party): %v", err)
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

// filhoDoContexto busca um tenant por id e garante que é filho direto do
// tenant do contexto (RN05) — usado por Obter/Atualizar/Excluir para nunca
// operar sobre (nem revelar a existência de) um tenant fora da hierarquia do
// chamador. Um id inexistente ou fora da hierarquia devolve o mesmo NotFound.
func (s *IAMServer) filhoDoContexto(ctx context.Context, parentID, id string) (store.Tenant, error) {
	t, err := s.store.ObterTenant(ctx, id)
	if errors.Is(err, store.ErrNaoEncontrado) {
		return store.Tenant{}, status.Error(codes.NotFound, "cliente não encontrado")
	}
	if err != nil {
		return store.Tenant{}, status.Error(codes.Internal, "falha ao obter tenant")
	}
	if t.ParentID == nil || *t.ParentID != parentID {
		return store.Tenant{}, status.Error(codes.NotFound, "cliente não encontrado")
	}
	return t, nil
}

// ObterTenant devolve um cliente específico sob o tenant do contexto (spec
// 004, RF07). Mesma restrição de CriarTenant (dono/parceiro).
func (s *IAMServer) ObterTenant(ctx context.Context, req *iamv1.ObterTenantRequest) (*iamv1.Tenant, error) {
	tc, err := tenantctx.Require(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "contexto de tenant ausente")
	}
	if tc.GetTipo() != "dono" && tc.GetTipo() != "parceiro" {
		return nil, status.Error(codes.PermissionDenied, "apenas dono ou parceiro podem visualizar clientes")
	}
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id obrigatório")
	}
	t, err := s.filhoDoContexto(ctx, tc.GetTenantId(), req.GetId())
	if err != nil {
		return nil, err
	}
	return &iamv1.Tenant{Id: t.ID, Nome: t.Nome, Tipo: t.Tipo}, nil
}

// AtualizarTenant renomeia um cliente sob o tenant do contexto (spec 004,
// RF07). Mesma restrição de CriarTenant (dono/parceiro).
func (s *IAMServer) AtualizarTenant(ctx context.Context, req *iamv1.AtualizarTenantRequest) (*iamv1.Tenant, error) {
	tc, err := tenantctx.Require(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "contexto de tenant ausente")
	}
	if tc.GetTipo() != "dono" && tc.GetTipo() != "parceiro" {
		return nil, status.Error(codes.PermissionDenied, "apenas dono ou parceiro podem atualizar clientes")
	}
	if req.GetNome() == "" {
		return nil, status.Error(codes.InvalidArgument, "nome obrigatório")
	}
	if _, err := s.filhoDoContexto(ctx, tc.GetTenantId(), req.GetId()); err != nil {
		return nil, err
	}
	t, err := s.store.AtualizarTenant(ctx, req.GetId(), req.GetNome())
	if err != nil {
		return nil, status.Error(codes.Internal, "falha ao atualizar tenant")
	}
	return &iamv1.Tenant{Id: t.ID, Nome: t.Nome, Tipo: t.Tipo}, nil
}

// ExcluirTenant remove um cliente sob o tenant do contexto (spec 004, RF07).
// Mesma restrição de CriarTenant (dono/parceiro). A exclusão é em cascata
// sobre sistemas/designs/versões/dados do cliente (ver store.ExcluirTenant).
func (s *IAMServer) ExcluirTenant(ctx context.Context, req *iamv1.ExcluirTenantRequest) (*iamv1.ExcluirTenantResponse, error) {
	tc, err := tenantctx.Require(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "contexto de tenant ausente")
	}
	if tc.GetTipo() != "dono" && tc.GetTipo() != "parceiro" {
		return nil, status.Error(codes.PermissionDenied, "apenas dono ou parceiro podem excluir clientes")
	}
	if _, err := s.filhoDoContexto(ctx, tc.GetTenantId(), req.GetId()); err != nil {
		return nil, err
	}
	if err := s.store.ExcluirTenant(ctx, req.GetId()); err != nil {
		return nil, status.Error(codes.Internal, "falha ao excluir tenant")
	}
	return &iamv1.ExcluirTenantResponse{}, nil
}

// RegistrarUsuario materializa o auto cadastro (spec 006, RF04): cria um
// tenant próprio (tipo dono) e a conta de senha do usuário registrante,
// devolvendo o JWT MACH já autenticado — mesmo formato de token emitido por
// AutenticarThirdParty (RN05).
func (s *IAMServer) RegistrarUsuario(ctx context.Context, req *iamv1.RegistrarUsuarioRequest) (*iamv1.RegistrarUsuarioResponse, error) {
	if req.GetNome() == "" || req.GetEmail() == "" || req.GetSenha() == "" || req.GetNomeTenant() == "" {
		return nil, status.Error(codes.InvalidArgument, "nome, email, senha e nome_tenant são obrigatórios")
	}
	if !strings.Contains(req.GetEmail(), "@") {
		return nil, status.Error(codes.InvalidArgument, "email inválido")
	}
	if len(req.GetSenha()) < 8 {
		return nil, status.Error(codes.InvalidArgument, "senha deve ter no mínimo 8 caracteres")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.GetSenha()), bcryptCost)
	if err != nil {
		return nil, status.Error(codes.Internal, "falha ao processar senha")
	}

	userID, tenantID, err := s.store.CriarTenantEUsuarioComSenha(ctx, req.GetNome(), req.GetEmail(), string(hash), req.GetNomeTenant())
	if errors.Is(err, store.ErrEmailJaCadastrado) {
		return nil, status.Error(codes.AlreadyExists, "e-mail já cadastrado")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "falha ao registrar usuário")
	}

	token, err := s.issuer.Issue(userID, tenantID, "dono")
	if err != nil {
		return nil, status.Error(codes.Internal, "falha ao emitir token")
	}
	return &iamv1.RegistrarUsuarioResponse{Jwt: token, UserId: userID, TenantId: tenantID, Tipo: "dono"}, nil
}

// AutenticarSenha materializa o login por e-mail/senha (spec 006, RF06) de uma
// conta criada via RegistrarUsuario. E-mail inexistente e senha incorreta
// devolvem exatamente o mesmo erro (RN04) — nunca revelam qual dos dois falhou.
func (s *IAMServer) AutenticarSenha(ctx context.Context, req *iamv1.AutenticarSenhaRequest) (*iamv1.AutenticarSenhaResponse, error) {
	if req.GetEmail() == "" || req.GetSenha() == "" {
		return nil, status.Error(codes.InvalidArgument, "email e senha são obrigatórios")
	}

	userID, tenantID, tipo, hash, err := s.store.ObterUsuarioPorEmailSenha(ctx, req.GetEmail())
	if errors.Is(err, store.ErrNaoEncontrado) {
		return nil, erroCredenciaisInvalidas
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "falha ao autenticar")
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.GetSenha())) != nil {
		return nil, erroCredenciaisInvalidas
	}
	if err := s.store.RegistrarEventoLogin(ctx, userID, tenantID); err != nil {
		log.Printf("iam: falha ao registrar evento de login (senha): %v", err)
	}

	token, err := s.issuer.Issue(userID, tenantID, tipo)
	if err != nil {
		return nil, status.Error(codes.Internal, "falha ao emitir token")
	}
	return &iamv1.AutenticarSenhaResponse{Jwt: token, UserId: userID, TenantId: tenantID, Tipo: tipo}, nil
}

// ListarUltimosAcessos devolve os 10 logins mais recentes dos tenants
// vinculados ao tenant do contexto (tenant + filhos diretos), para o card
// "Últimos Acessos" do Dashboard (spec 004, RF04, RN02).
func (s *IAMServer) ListarUltimosAcessos(ctx context.Context, _ *iamv1.ListarUltimosAcessosRequest) (*iamv1.ListarUltimosAcessosResponse, error) {
	tc, err := tenantctx.Require(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "contexto de tenant ausente")
	}
	eventos, err := s.store.UltimosAcessos(ctx, tc.GetTenantId(), 10)
	if err != nil {
		return nil, status.Error(codes.Internal, "falha ao listar últimos acessos")
	}
	out := make([]*iamv1.EventoLogin, 0, len(eventos))
	for _, e := range eventos {
		out = append(out, &iamv1.EventoLogin{
			UsuarioNome: e.UsuarioNome,
			TenantNome:  e.TenantNome,
			CriadoEm:    e.CriadoEm.Format(time.RFC3339),
		})
	}
	return &iamv1.ListarUltimosAcessosResponse{Eventos: out}, nil
}

// ListarFeedback devolve as mensagens de feedback dos tenants vinculados ao
// tenant do contexto, opcionalmente filtradas por status (spec 004, RF05).
func (s *IAMServer) ListarFeedback(ctx context.Context, req *iamv1.ListarFeedbackRequest) (*iamv1.ListarFeedbackResponse, error) {
	tc, err := tenantctx.Require(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "contexto de tenant ausente")
	}
	var filtroStatus *string
	if valor := req.GetStatus(); valor != "" {
		if valor != "pendente" && valor != "respondido" {
			return nil, status.Error(codes.InvalidArgument, "status deve ser 'pendente' ou 'respondido'")
		}
		filtroStatus = &valor
	}
	itens, err := s.store.ListarFeedback(ctx, tc.GetTenantId(), filtroStatus)
	if err != nil {
		return nil, status.Error(codes.Internal, "falha ao listar feedback")
	}
	out := make([]*iamv1.Feedback, 0, len(itens))
	for _, f := range itens {
		out = append(out, &iamv1.Feedback{
			Id:         f.ID,
			TenantNome: f.TenantNome,
			Mensagem:   f.Mensagem,
			Status:     f.Status,
			CriadoEm:   f.CriadoEm.Format(time.RFC3339),
		})
	}
	return &iamv1.ListarFeedbackResponse{Itens: out}, nil
}

// AtualizarStatusFeedback aplica a transição de status de uma mensagem de
// feedback (spec 004, RF05, RN03). A única transição aceita é
// pendente→respondido — o inverso (respondido→pendente) é sempre rejeitado,
// mesmo antes de tocar o banco.
func (s *IAMServer) AtualizarStatusFeedback(ctx context.Context, req *iamv1.AtualizarStatusFeedbackRequest) (*iamv1.Feedback, error) {
	if _, err := tenantctx.Require(ctx); err != nil {
		return nil, status.Error(codes.Unauthenticated, "contexto de tenant ausente")
	}
	if req.GetStatus() != "respondido" {
		return nil, status.Error(codes.InvalidArgument, "único status aceito é 'respondido' — a transição respondido→pendente não é permitida (RN03)")
	}
	f, err := s.store.AtualizarStatusFeedback(ctx, req.GetId(), req.GetStatus())
	if errors.Is(err, store.ErrNaoEncontrado) {
		return nil, status.Error(codes.NotFound, "feedback não encontrado")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "falha ao atualizar status de feedback")
	}
	return &iamv1.Feedback{
		Id:         f.ID,
		TenantNome: f.TenantNome,
		Mensagem:   f.Mensagem,
		Status:     f.Status,
		CriadoEm:   f.CriadoEm.Format(time.RFC3339),
	}, nil
}

// ResumoFinanceiro devolve a receita agregada do mês corrente dos tenants
// vinculados ao tenant do contexto, para o card "Resumo Financeiro" do
// Dashboard (spec 004, RF06, RN04).
func (s *IAMServer) ResumoFinanceiro(ctx context.Context, _ *iamv1.ResumoFinanceiroRequest) (*iamv1.ResumoFinanceiroResponse, error) {
	tc, err := tenantctx.Require(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "contexto de tenant ausente")
	}
	r, err := s.store.ResumoFinanceiro(ctx, tc.GetTenantId())
	if err != nil {
		return nil, status.Error(codes.Internal, "falha ao calcular resumo financeiro")
	}
	return &iamv1.ResumoFinanceiroResponse{
		ReceitaTotalCentavos: r.ReceitaTotalCentavos,
		Moeda:                r.Moeda,
		Competencia:          r.Competencia,
	}, nil
}
