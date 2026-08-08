// Package server implementa o IAMService gRPC (RF03, RN03).
package server

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonv1 "github.com/machv4/platform/gen/go/construtor/common/v1"
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

	// Área "Conta/Configuração" (spec 004, RF14-RF18).
	AtualizarPerfil(ctx context.Context, userID, nome, fotoURL string) error
	ObterEmailUsuario(ctx context.Context, userID string) (string, error)
	ObterHashSenha(ctx context.Context, userID string) (string, error)
	AtualizarSenha(ctx context.Context, userID, novoHash string) error
	IniciarMfa(ctx context.Context, userID string, segredoCifrado []byte) error
	ObterSegredoMfaPendente(ctx context.Context, userID string) (segredoCifrado []byte, ultimoCodigo string, ultimoCodigoEm *time.Time, err error)
	ConfirmarMfa(ctx context.Context, userID, codigoUsado string, usadoEm time.Time) error
	DesativarMfa(ctx context.Context, userID string) error
	ExcluirConta(ctx context.Context, userID, tenantID string) error
	SolicitarTrocaEmail(ctx context.Context, userID, novoEmail, tokenHash string, expiraEm time.Time) error
	ConfirmarTrocaEmail(ctx context.Context, tokenHash string) error
}

// IAMServer implementa iamv1.IAMServiceServer.
type IAMServer struct {
	iamv1.UnimplementedIAMServiceServer
	validator *auth.Validator
	issuer    *auth.Issuer
	store     Store
	eval      permissions.Evaluator
	// mfaKey cifra/decifra segredos TOTP em repouso (services/iam/auth/mfa.go).
	// Zero-value só é aceitável fora de produção (main.go recusa subir sem
	// IAM_MFA_ENCRYPTION_KEY) — ver ComChaveMfa.
	mfaKey [32]byte
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

	token, err := s.issuer.Issue(userID, tenantID, tipo)
	if err != nil {
		return nil, status.Error(codes.Internal, "falha ao emitir token")
	}
	return &iamv1.AutenticarSenhaResponse{Jwt: token, UserId: userID, TenantId: tenantID, Tipo: tipo}, nil
}

// ─────────────────────────────────────────────────────────────────────────
// Área "Conta/Configuração" (spec 004-reestruturacao-ia-navegacao, RF14-RF18).
//
// Pré-requisito investigado antes de implementar qualquer uma destas rotas:
// user_id do chamador já trafega no TenantContext (pkg/tenantctx, proto
// common/v1/tenant.proto) — o Gateway o preenche em middleware.Auth a partir
// de ValidarTokenResponse.UserId (que vem de claims.Subject, o "sub" do JWT
// emitido por Issuer.Issue). Não foi necessário mudar TenantContext/Claims;
// basta usar tc.GetUserId() depois de tenantctx.Require (ver requireUsuario).
// ─────────────────────────────────────────────────────────────────────────

// erroReautenticacaoNecessaria cobre RNF02: senha_atual ausente ou incorreta
// numa rota que exige reconfirmar a senha (troca de senha, desativar MFA,
// excluir conta). Distinto de erroCredenciaisInvalidas (login) na mensagem —
// o Gateway usa essa distinção para mapear 401 REAUTENTICACAO_NECESSARIA vs
// 401 UNAUTHORIZED genérico (ver writeContaError em routes/conta.go).
var erroReautenticacaoNecessaria = status.Error(codes.Unauthenticated, "reautenticação necessária")

// requireUsuario é o tenantctx.Require desta área: além do tenant, exige que
// o TenantContext carregue user_id — toda rota de Conta/Configuração opera
// sobre "o usuário autenticado", nunca sobre um id recebido no corpo.
func (s *IAMServer) requireUsuario(ctx context.Context) (*commonv1.TenantContext, error) {
	tc, err := tenantctx.Require(ctx)
	if err != nil || tc.GetUserId() == "" {
		return nil, status.Error(codes.Unauthenticated, "contexto de tenant ausente")
	}
	return tc, nil
}

// reautenticar confere senhaAtual contra o hash persistido do usuário (RNF02).
// Contas third-party (senha_hash vazio) nunca reautenticam com sucesso aqui —
// mesmo comportamento de "credenciais inválidas" sem distinguir o motivo.
func (s *IAMServer) reautenticar(ctx context.Context, userID, senhaAtual string) error {
	if senhaAtual == "" {
		return erroReautenticacaoNecessaria
	}
	hash, err := s.store.ObterHashSenha(ctx, userID)
	if err != nil {
		return status.Error(codes.Internal, "falha ao reautenticar")
	}
	if hash == "" || bcrypt.CompareHashAndPassword([]byte(hash), []byte(senhaAtual)) != nil {
		return erroReautenticacaoNecessaria
	}
	return nil
}

// AtualizarPerfil atualiza nome e foto_url do usuário autenticado (RF17).
func (s *IAMServer) AtualizarPerfil(ctx context.Context, req *iamv1.AtualizarPerfilRequest) (*iamv1.AtualizarPerfilResponse, error) {
	tc, err := s.requireUsuario(ctx)
	if err != nil {
		return nil, err
	}
	if req.GetNome() == "" {
		return nil, status.Error(codes.InvalidArgument, "nome obrigatório")
	}
	if err := s.store.AtualizarPerfil(ctx, tc.GetUserId(), req.GetNome(), req.GetFotoUrl()); err != nil {
		return nil, status.Error(codes.Internal, "falha ao atualizar perfil")
	}
	return &iamv1.AtualizarPerfilResponse{}, nil
}

// AtualizarSenha troca a senha da conta (RF14). RNF02: senha_atual reautentica
// antes de qualquer escrita — nunca troca a senha sem confirmar a atual.
func (s *IAMServer) AtualizarSenha(ctx context.Context, req *iamv1.AtualizarSenhaRequest) (*iamv1.AtualizarSenhaResponse, error) {
	tc, err := s.requireUsuario(ctx)
	if err != nil {
		return nil, err
	}
	if len(req.GetSenhaNova()) < 8 {
		return nil, status.Error(codes.InvalidArgument, "senha deve ter no mínimo 8 caracteres")
	}
	if err := s.reautenticar(ctx, tc.GetUserId(), req.GetSenhaAtual()); err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.GetSenhaNova()), bcryptCost)
	if err != nil {
		return nil, status.Error(codes.Internal, "falha ao processar senha")
	}
	if err := s.store.AtualizarSenha(ctx, tc.GetUserId(), string(hash)); err != nil {
		return nil, status.Error(codes.Internal, "falha ao atualizar senha")
	}
	return &iamv1.AtualizarSenhaResponse{}, nil
}

// AtivarMfa gera um novo segredo TOTP, cifra e persiste (sem ligar o MFA
// ainda — só ConfirmarMfa liga) e devolve a otpauth:// URI de exibição única
// (RF15, RNF01): não existe endpoint de releitura, se o usuário perder esta
// resposta precisa reiniciar o fluxo chamando /mfa/ativar de novo.
func (s *IAMServer) AtivarMfa(ctx context.Context, _ *iamv1.AtivarMfaRequest) (*iamv1.AtivarMfaResponse, error) {
	tc, err := s.requireUsuario(ctx)
	if err != nil {
		return nil, err
	}
	email, err := s.store.ObterEmailUsuario(ctx, tc.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, "falha ao carregar usuário")
	}
	chave, err := totp.Generate(totp.GenerateOpts{Issuer: "MACH V4", AccountName: email})
	if err != nil {
		return nil, status.Error(codes.Internal, "falha ao gerar segredo mfa")
	}
	cifrado, err := auth.CifrarSegredo(s.mfaKey, chave.Secret())
	if err != nil {
		return nil, status.Error(codes.Internal, "falha ao cifrar segredo mfa")
	}
	if err := s.store.IniciarMfa(ctx, tc.GetUserId(), cifrado); err != nil {
		if errors.Is(err, store.ErrMfaJaAtivo) {
			return nil, status.Error(codes.FailedPrecondition, "mfa já está ativo nesta conta")
		}
		return nil, status.Error(codes.Internal, "falha ao iniciar mfa")
	}
	return &iamv1.AtivarMfaResponse{SegredoOtpAuthUri: chave.URL()}, nil
}

// ConfirmarMfa valida o código TOTP contra o segredo pendente e, se válido,
// liga o MFA (RF15). Anti-replay: o mesmo código nunca é aceito duas vezes
// seguidas (mfa_ultimo_codigo_usado). totp.Validate já cobre a janela padrão
// de ±1 período de 30s — não reimplementamos essa lógica.
func (s *IAMServer) ConfirmarMfa(ctx context.Context, req *iamv1.ConfirmarMfaRequest) (*iamv1.ConfirmarMfaResponse, error) {
	tc, err := s.requireUsuario(ctx)
	if err != nil {
		return nil, err
	}
	if req.GetCodigo() == "" {
		return nil, status.Error(codes.InvalidArgument, "código obrigatório")
	}
	segredoCifrado, ultimoCodigo, _, err := s.store.ObterSegredoMfaPendente(ctx, tc.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, "falha ao carregar mfa")
	}
	if len(segredoCifrado) == 0 {
		return nil, status.Error(codes.FailedPrecondition, "nenhuma ativação de mfa pendente — chame /conta/mfa/ativar primeiro")
	}
	if ultimoCodigo != "" && req.GetCodigo() == ultimoCodigo {
		return nil, status.Error(codes.Unauthenticated, "código mfa já utilizado")
	}
	segredoClaro, err := auth.DecifrarSegredo(s.mfaKey, segredoCifrado)
	if err != nil {
		return nil, status.Error(codes.Internal, "falha ao decifrar segredo mfa")
	}
	if !totp.Validate(req.GetCodigo(), segredoClaro) {
		return nil, status.Error(codes.Unauthenticated, "código mfa inválido")
	}
	if err := s.store.ConfirmarMfa(ctx, tc.GetUserId(), req.GetCodigo(), time.Now()); err != nil {
		return nil, status.Error(codes.Internal, "falha ao confirmar mfa")
	}
	return &iamv1.ConfirmarMfaResponse{}, nil
}

// DesativarMfa desliga o MFA (RF15). Decisão de segurança que diverge do
// contrato assumido em contracts/api.md (que não pedia reautenticação nesta
// rota): exige senha_atual (RNF02) — desativar 2FA sem confirmar a senha é
// superfície de abuso de sessão sequestrada.
func (s *IAMServer) DesativarMfa(ctx context.Context, req *iamv1.DesativarMfaRequest) (*iamv1.DesativarMfaResponse, error) {
	tc, err := s.requireUsuario(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.reautenticar(ctx, tc.GetUserId(), req.GetSenhaAtual()); err != nil {
		return nil, err
	}
	if err := s.store.DesativarMfa(ctx, tc.GetUserId()); err != nil {
		return nil, status.Error(codes.Internal, "falha ao desativar mfa")
	}
	return &iamv1.DesativarMfaResponse{}, nil
}

// ExcluirConta anonimiza a conta autenticada (RF16, RN07) — nunca um DELETE
// físico (LGPD, preserva o id para referências históricas sem manter PII).
// RN07: bloqueada se o tenant do usuário tiver 1+ filhos vinculados — a
// checagem aqui reaproveita ListarFilhos (fail-fast, mesma consulta de
// ListarTenants); store.ExcluirConta reconta os filhos DENTRO da mesma
// transação da anonimização, para nunca excluir parcialmente numa corrida.
// RNF02: senha_atual reautentica antes de qualquer escrita.
func (s *IAMServer) ExcluirConta(ctx context.Context, req *iamv1.ExcluirContaRequest) (*iamv1.ExcluirContaResponse, error) {
	tc, err := s.requireUsuario(ctx)
	if err != nil {
		return nil, err
	}
	filhos, err := s.store.ListarFilhos(ctx, tc.GetTenantId())
	if err != nil {
		return nil, status.Error(codes.Internal, "falha ao verificar tenants vinculados")
	}
	if len(filhos) > 0 {
		return nil, status.Error(codes.FailedPrecondition, "existem tenants ativos vinculados a esta conta")
	}
	if err := s.reautenticar(ctx, tc.GetUserId(), req.GetSenhaAtual()); err != nil {
		return nil, err
	}
	if err := s.store.ExcluirConta(ctx, tc.GetUserId(), tc.GetTenantId()); err != nil {
		if errors.Is(err, store.ErrTenantAtivoVinculado) {
			return nil, status.Error(codes.FailedPrecondition, "existem tenants ativos vinculados a esta conta")
		}
		return nil, status.Error(codes.Internal, "falha ao excluir conta")
	}
	return &iamv1.ExcluirContaResponse{}, nil
}

// SolicitarTrocaEmail grava o e-mail pendente e o token de confirmação (RF18,
// RN08) — `email` (login) só muda na confirmação. Não há infraestrutura de
// e-mail/SMTP no IAM (dívida técnica já aceita no projeto — mesma
// simplificação de services/workers/internal/handlers/notificacao.go, que só
// loga): loga o link de confirmação com log.Printf em vez de enviar e-mail.
func (s *IAMServer) SolicitarTrocaEmail(ctx context.Context, req *iamv1.SolicitarTrocaEmailRequest) (*iamv1.SolicitarTrocaEmailResponse, error) {
	tc, err := s.requireUsuario(ctx)
	if err != nil {
		return nil, err
	}
	novoEmail := req.GetNovoEmail()
	if novoEmail == "" || !strings.Contains(novoEmail, "@") {
		return nil, status.Error(codes.InvalidArgument, "novo_email inválido")
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, status.Error(codes.Internal, "falha ao gerar token de confirmação")
	}
	token := hex.EncodeToString(tokenBytes)
	soma := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(soma[:])

	if err := s.store.SolicitarTrocaEmail(ctx, tc.GetUserId(), novoEmail, tokenHash, time.Now().Add(time.Hour)); err != nil {
		if errors.Is(err, store.ErrEmailJaCadastrado) {
			return nil, status.Error(codes.AlreadyExists, "e-mail já cadastrado")
		}
		return nil, status.Error(codes.Internal, "falha ao solicitar troca de e-mail")
	}

	log.Printf("iam: confirmação de troca de e-mail para %s: token=%s (válido por 1h)", novoEmail, token)
	return &iamv1.SolicitarTrocaEmailResponse{}, nil
}

// erroTokenEmailInvalido cobre token inexistente e token expirado com a mesma
// mensagem genérica (mesmo racional de erroCredenciaisInvalidas, RN04): não
// distinguir "não existe" de "expirou".
var erroTokenEmailInvalido = status.Error(codes.InvalidArgument, "token de confirmação inválido ou expirado")

// ConfirmarTrocaEmail efetiva a troca de e-mail (RF18, RN08). Não checa o
// tenant/usuário do chamador contra o token de propósito — o token por si só
// já é a prova de posse do novo e-mail (mesmo modelo de link de confirmação
// de e-mail comum a outros produtos).
func (s *IAMServer) ConfirmarTrocaEmail(ctx context.Context, req *iamv1.ConfirmarTrocaEmailRequest) (*iamv1.ConfirmarTrocaEmailResponse, error) {
	if req.GetToken() == "" {
		return nil, erroTokenEmailInvalido
	}
	soma := sha256.Sum256([]byte(req.GetToken()))
	tokenHash := hex.EncodeToString(soma[:])

	err := s.store.ConfirmarTrocaEmail(ctx, tokenHash)
	if errors.Is(err, store.ErrNaoEncontrado) {
		return nil, erroTokenEmailInvalido
	}
	if errors.Is(err, store.ErrEmailJaCadastrado) {
		return nil, status.Error(codes.AlreadyExists, "e-mail já cadastrado")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "falha ao confirmar troca de e-mail")
	}
	return &iamv1.ConfirmarTrocaEmailResponse{}, nil
}

// ComChaveMfa liga a chave de cifra AES-256-GCM usada para os segredos TOTP
// (services/iam/auth/mfa.go) ao servidor — segue o mesmo padrão de
// LogicServer.ComPublicador (services/logic/internal/server/grpc.go): setter
// encadeável chamado depois de New, para não crescer a assinatura de New a
// cada nova dependência opcional.
func (s *IAMServer) ComChaveMfa(chave [32]byte) *IAMServer {
	s.mfaKey = chave
	return s
}
