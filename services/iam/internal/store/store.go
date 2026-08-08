// Package store persiste e consulta tenants hierárquicos e permissões (RF03).
//
// A tabela tenants não é isolada por RLS (ela É o tenant) e é consultada
// diretamente. Já permissoes é multi-tenant: além da RLS (migração 0010, segunda
// camada), o carregamento aplica um filtro explícito por tenant_id derivado do
// TenantContext — defesa em profundidade, correto mesmo que a conexão tenha
// BYPASSRLS.
package store

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/machv4/platform/pkg/tenantctx"
	"github.com/machv4/platform/services/iam/internal/permissions"
)

// ErrNaoEncontrado indica tenant inexistente.
var ErrNaoEncontrado = errors.New("store: tenant não encontrado")

// ErrSemTenant indica ausência de tenant no contexto ao consultar dados isolados.
var ErrSemTenant = errors.New("store: contexto sem tenant (RN01)")

// ErrEmailJaCadastrado indica e-mail já usado por outra conta de senha (RN02,
// índice único parcial da migração 0014).
var ErrEmailJaCadastrado = errors.New("store: e-mail já cadastrado")

// TenantPadraoID é o tenant fixo onde entram os usuários autenticados via
// provedor third-party (migração 0013). Todo login OAuth vira 'cliente' aqui.
const TenantPadraoID = "00000000-0000-0000-0000-000000000001"

// Tenant é um nó da hierarquia Dono → Parceiro → Cliente Final.
type Tenant struct {
	ID       string
	ParentID *string
	Nome     string
	Tipo     string
}

// DB é o subconjunto de pgxpool.Pool usado pelo store (facilita testes). Begin
// é usado só por CriarTenantEUsuarioComSenha (RNF03 — atomicidade tenant+usuário).
type DB interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

// Store agrega o acesso a tenants e permissões.
type Store struct {
	db DB
}

// New cria um Store sobre um pool/conn pgx.
func New(db DB) *Store {
	return &Store{db: db}
}

// CriarTenant insere um tenant. parentID nil = raiz (Dono). chaveBlindIndex é a
// chave HMAC por tenant usada em pkg/blindindex.
func (s *Store) CriarTenant(ctx context.Context, nome, tipo string, parentID *string, chaveBlindIndex []byte) (Tenant, error) {
	var t Tenant
	err := s.db.QueryRow(ctx,
		`INSERT INTO tenants (parent_id, nome, tipo, chave_blind_index)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, parent_id, nome, tipo::text`,
		parentID, nome, tipo, chaveBlindIndex,
	).Scan(&t.ID, &t.ParentID, &t.Nome, &t.Tipo)
	if err != nil {
		return Tenant{}, fmt.Errorf("store: criar tenant: %w", err)
	}
	return t, nil
}

// ObterTenant devolve um tenant por id.
func (s *Store) ObterTenant(ctx context.Context, id string) (Tenant, error) {
	var t Tenant
	err := s.db.QueryRow(ctx,
		`SELECT id, parent_id, nome, tipo::text FROM tenants WHERE id = $1`, id,
	).Scan(&t.ID, &t.ParentID, &t.Nome, &t.Tipo)
	if errors.Is(err, pgx.ErrNoRows) {
		return Tenant{}, ErrNaoEncontrado
	}
	if err != nil {
		return Tenant{}, fmt.Errorf("store: obter tenant: %w", err)
	}
	return t, nil
}

// AtualizarTenant renomeia um tenant existente.
func (s *Store) AtualizarTenant(ctx context.Context, id, nome string) (Tenant, error) {
	var t Tenant
	err := s.db.QueryRow(ctx,
		`UPDATE tenants SET nome = $2 WHERE id = $1
		 RETURNING id, parent_id, nome, tipo::text`,
		id, nome,
	).Scan(&t.ID, &t.ParentID, &t.Nome, &t.Tipo)
	if errors.Is(err, pgx.ErrNoRows) {
		return Tenant{}, ErrNaoEncontrado
	}
	if err != nil {
		return Tenant{}, fmt.Errorf("store: atualizar tenant: %w", err)
	}
	return t, nil
}

// ExcluirTenant remove um tenant por id. A exclusão é em cascata sobre
// sistemas/designs/versões/dados vinculados (ver ON DELETE CASCADE nas
// migrações que referenciam tenants).
func (s *Store) ExcluirTenant(ctx context.Context, id string) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM tenants WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("store: excluir tenant: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNaoEncontrado
	}
	return nil
}

// ListarFilhos devolve os tenants filhos diretos de parentID (hierarquia).
func (s *Store) ListarFilhos(ctx context.Context, parentID string) ([]Tenant, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, parent_id, nome, tipo::text FROM tenants WHERE parent_id = $1 ORDER BY nome`, parentID)
	if err != nil {
		return nil, fmt.Errorf("store: listar filhos: %w", err)
	}
	defer rows.Close()

	var out []Tenant
	for rows.Next() {
		var t Tenant
		if err := rows.Scan(&t.ID, &t.ParentID, &t.Nome, &t.Tipo); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// UpsertUsuarioThirdParty faz find-or-create do usuário por (provedor, external_id)
// e devolve a identidade para emissão do JWT. Novos usuários entram no tenant
// padrão como 'cliente' (migração 0013); logins seguintes atualizam email/nome.
func (s *Store) UpsertUsuarioThirdParty(ctx context.Context, provedor, externalID, email, nome string) (userID, tenantID, tipo string, err error) {
	err = s.db.QueryRow(ctx,
		`INSERT INTO users (provedor, external_id, email, nome, tenant_id, tipo)
		 VALUES ($1, $2, $3, $4, $5, 'cliente')
		 ON CONFLICT (provedor, external_id)
		 DO UPDATE SET email = EXCLUDED.email, nome = EXCLUDED.nome, atualizado_em = now()
		 RETURNING id, tenant_id, tipo::text`,
		provedor, externalID, email, nome, TenantPadraoID,
	).Scan(&userID, &tenantID, &tipo)
	if err != nil {
		return "", "", "", fmt.Errorf("store: upsert usuário third-party: %w", err)
	}
	return userID, tenantID, tipo, nil
}

// CriarTenantEUsuarioComSenha materializa o auto cadastro (spec 006, RF04): cria
// um tenant raiz (tipo 'dono') e, na mesma transação, o usuário registrante
// (provedor='senha', tipo 'dono') com a senha já em hash. Atômico por RNF03 — se
// o e-mail já existir para uma conta de senha (índice único parcial da migração
// 0014), a transação é revertida e nenhum tenant órfão sobra.
func (s *Store) CriarTenantEUsuarioComSenha(ctx context.Context, nomeUsuario, email, senhaHash, nomeTenant string) (userID, tenantID string, err error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return "", "", fmt.Errorf("store: iniciar transação: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op após commit bem-sucedido

	chave := make([]byte, 32)
	if _, err := rand.Read(chave); err != nil {
		return "", "", fmt.Errorf("store: gerar chave do tenant: %w", err)
	}

	if err := tx.QueryRow(ctx,
		`INSERT INTO tenants (parent_id, nome, tipo, chave_blind_index)
		 VALUES (NULL, $1, 'dono', $2) RETURNING id`,
		nomeTenant, chave,
	).Scan(&tenantID); err != nil {
		return "", "", fmt.Errorf("store: criar tenant do cadastro: %w", err)
	}

	err = tx.QueryRow(ctx,
		`INSERT INTO users (provedor, external_id, email, nome, senha_hash, tenant_id, tipo)
		 VALUES ('senha', $1, $1, $2, $3, $4, 'dono')
		 RETURNING id`,
		email, nomeUsuario, senhaHash, tenantID,
	).Scan(&userID)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return "", "", ErrEmailJaCadastrado
	}
	if err != nil {
		return "", "", fmt.Errorf("store: criar usuário do cadastro: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", "", fmt.Errorf("store: commit do cadastro: %w", err)
	}
	return userID, tenantID, nil
}

// ObterUsuarioPorEmailSenha busca uma conta de senha (provedor='senha') pelo
// e-mail, para o fluxo de login (spec 006, RF06). ErrNaoEncontrado cobre e-mail
// inexistente — o chamador (grpc.go) devolve a mesma mensagem genérica de erro
// para isso e para senha incorreta (RN04).
func (s *Store) ObterUsuarioPorEmailSenha(ctx context.Context, email string) (userID, tenantID, tipo, senhaHash string, err error) {
	err = s.db.QueryRow(ctx,
		`SELECT id, tenant_id, tipo::text, coalesce(senha_hash, '')
		   FROM users
		  WHERE provedor = 'senha' AND email = $1`,
		email,
	).Scan(&userID, &tenantID, &tipo, &senhaHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", "", "", ErrNaoEncontrado
	}
	if err != nil {
		return "", "", "", "", fmt.Errorf("store: obter usuário por e-mail/senha: %w", err)
	}
	return userID, tenantID, tipo, senhaHash, nil
}

// PermissoesDe carrega as permissões do tenant corrente para os componentes
// informados. Rejeita se não houver tenant no contexto.
func (s *Store) PermissoesDe(ctx context.Context, blindIndexes []string) ([]permissions.Permissao, error) {
	tid := tenantctx.TenantID(ctx)
	if tid == "" {
		return nil, ErrSemTenant
	}
	rows, err := s.db.Query(ctx,
		`SELECT blind_index, papel, condicao, view, click
		   FROM permissoes
		  WHERE tenant_id = $1 AND blind_index = ANY($2)`,
		tid, blindIndexes)
	if err != nil {
		return nil, fmt.Errorf("store: permissões: %w", err)
	}
	defer rows.Close()

	var out []permissions.Permissao
	for rows.Next() {
		var p permissions.Permissao
		if err := rows.Scan(&p.BlindIndex, &p.Papel, &p.Condicao, &p.View, &p.Click); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ─────────────────────────────────────────────────────────────────────────
// Área "Conta/Configuração" (spec 004-reestruturacao-ia-navegacao, RF14-RF18):
// perfil, senha, MFA TOTP, exclusão (anonimização) e troca de e-mail. Todos os
// métodos abaixo operam por userID (não por tenant), diferente do restante do
// arquivo — a identidade do usuário chamador vem do TenantContext (user_id).
// ─────────────────────────────────────────────────────────────────────────

// ErrMfaJaAtivo indica que a conta já tem MFA ativo — AtivarMfa não permite
// gerar um novo segredo por cima de um MFA já confirmado (o usuário precisa
// desativar antes de reconfigurar).
var ErrMfaJaAtivo = errors.New("store: mfa já ativo")

// ErrTenantAtivoVinculado indica que o tenant do usuário ainda tem 1+ tenants
// filhos vinculados — bloqueia a exclusão de conta (RN07).
var ErrTenantAtivoVinculado = errors.New("store: tenant ativo vinculado")

// AtualizarPerfil atualiza nome e foto_url do usuário (RF17).
func (s *Store) AtualizarPerfil(ctx context.Context, userID, nome, fotoURL string) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE users SET nome = $2, foto_url = $3, atualizado_em = now() WHERE id = $1`,
		userID, nome, fotoURL)
	if err != nil {
		return fmt.Errorf("store: atualizar perfil: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNaoEncontrado
	}
	return nil
}

// ObterEmailUsuario devolve o e-mail de login do usuário — usado para nomear a
// conta no otpauth:// URI (RF15) e para as mensagens de erro de troca de e-mail.
func (s *Store) ObterEmailUsuario(ctx context.Context, userID string) (string, error) {
	var email string
	err := s.db.QueryRow(ctx, `SELECT email FROM users WHERE id = $1`, userID).Scan(&email)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNaoEncontrado
	}
	if err != nil {
		return "", fmt.Errorf("store: obter email do usuário: %w", err)
	}
	return email, nil
}

// ObterHashSenha busca o hash de senha do usuário por id — para reautenticação
// (RNF02) em troca de senha, MFA e exclusão de conta. Diferente de
// ObterUsuarioPorEmailSenha (usado só no login), esta consulta é por userID.
func (s *Store) ObterHashSenha(ctx context.Context, userID string) (string, error) {
	var hash string
	err := s.db.QueryRow(ctx,
		`SELECT coalesce(senha_hash, '') FROM users WHERE id = $1`, userID,
	).Scan(&hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNaoEncontrado
	}
	if err != nil {
		return "", fmt.Errorf("store: obter hash de senha: %w", err)
	}
	return hash, nil
}

// AtualizarSenha persiste o novo hash de senha (RF14). O chamador (grpc.go) já
// validou senha_atual via bcrypt antes de chegar aqui.
func (s *Store) AtualizarSenha(ctx context.Context, userID, novoHash string) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE users SET senha_hash = $2, atualizado_em = now() WHERE id = $1`,
		userID, novoHash)
	if err != nil {
		return fmt.Errorf("store: atualizar senha: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNaoEncontrado
	}
	return nil
}

// IniciarMfa persiste um novo segredo TOTP cifrado (etapa "ativar", RF15) sem
// marcar mfa_ativo — só ConfirmarMfa liga o MFA. Zera qualquer estado de
// anti-replay de uma ativação anterior abandonada. Recusa com ErrMfaJaAtivo se
// a conta já tiver MFA confirmado (é preciso desativar antes de reconfigurar).
func (s *Store) IniciarMfa(ctx context.Context, userID string, segredoCifrado []byte) error {
	var ativo bool
	err := s.db.QueryRow(ctx, `SELECT mfa_ativo FROM users WHERE id = $1`, userID).Scan(&ativo)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNaoEncontrado
	}
	if err != nil {
		return fmt.Errorf("store: consultar mfa antes de iniciar: %w", err)
	}
	if ativo {
		return ErrMfaJaAtivo
	}
	if _, err := s.db.Exec(ctx,
		`UPDATE users
		    SET mfa_segredo_cifrado = $2, mfa_ultimo_codigo_usado = NULL, mfa_ultimo_codigo_em = NULL, atualizado_em = now()
		  WHERE id = $1`,
		userID, segredoCifrado,
	); err != nil {
		return fmt.Errorf("store: salvar segredo mfa: %w", err)
	}
	return nil
}

// ObterSegredoMfaPendente devolve o segredo cifrado e o estado de anti-replay
// para a etapa "confirmar" (RF15). segredoCifrado nil (ou vazio) indica que
// nenhuma ativação foi iniciada — o chamador trata isso como token inválido.
func (s *Store) ObterSegredoMfaPendente(ctx context.Context, userID string) (segredoCifrado []byte, ultimoCodigo string, ultimoCodigoEm *time.Time, err error) {
	var codigo *string
	err = s.db.QueryRow(ctx,
		`SELECT mfa_segredo_cifrado, mfa_ultimo_codigo_usado, mfa_ultimo_codigo_em FROM users WHERE id = $1`,
		userID,
	).Scan(&segredoCifrado, &codigo, &ultimoCodigoEm)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", nil, ErrNaoEncontrado
	}
	if err != nil {
		return nil, "", nil, fmt.Errorf("store: obter segredo mfa pendente: %w", err)
	}
	if codigo != nil {
		ultimoCodigo = *codigo
	}
	return segredoCifrado, ultimoCodigo, ultimoCodigoEm, nil
}

// ConfirmarMfa liga o MFA (mfa_ativo = true) e registra o código usado para o
// anti-replay (RF15): a próxima confirmação/uso não pode repetir codigoUsado.
func (s *Store) ConfirmarMfa(ctx context.Context, userID, codigoUsado string, usadoEm time.Time) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE users
		    SET mfa_ativo = true, mfa_ultimo_codigo_usado = $2, mfa_ultimo_codigo_em = $3, atualizado_em = now()
		  WHERE id = $1`,
		userID, codigoUsado, usadoEm)
	if err != nil {
		return fmt.Errorf("store: confirmar mfa: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNaoEncontrado
	}
	return nil
}

// DesativarMfa desliga o MFA e descarta o segredo cifrado (RF15) — reativar
// depois exige um novo enrollment completo (ativar → confirmar), nunca reusa
// um segredo antigo.
func (s *Store) DesativarMfa(ctx context.Context, userID string) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE users
		    SET mfa_ativo = false, mfa_segredo_cifrado = NULL, mfa_ultimo_codigo_usado = NULL, mfa_ultimo_codigo_em = NULL, atualizado_em = now()
		  WHERE id = $1`,
		userID)
	if err != nil {
		return fmt.Errorf("store: desativar mfa: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNaoEncontrado
	}
	return nil
}

// ExcluirConta anonimiza o usuário (LGPD — não é DELETE físico, preserva o id
// para qualquer referência histórica sem manter PII, RF16/RN07). O check de
// bloqueio (tenant do usuário sem filhos vinculados) e a anonimização
// acontecem na MESMA transação: o gRPC já consultou ListarFilhos antes de
// chegar aqui (fail-fast), mas essa recontagem dentro da transação evita que
// uma corrida (um tenant filho criado entre o check e o commit) deixe a conta
// anonimizada com um cliente órfão vinculado.
func (s *Store) ExcluirConta(ctx context.Context, userID, tenantID string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("store: iniciar transação de exclusão de conta: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op após commit bem-sucedido

	var filhos int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM tenants WHERE parent_id = $1`, tenantID).Scan(&filhos); err != nil {
		return fmt.Errorf("store: contar filhos do tenant na exclusão: %w", err)
	}
	if filhos > 0 {
		return ErrTenantAtivoVinculado
	}

	tag, err := tx.Exec(ctx,
		`UPDATE users
		    SET nome = '',
		        email = gen_random_uuid()::text || '@removido.local',
		        external_id = gen_random_uuid()::text,
		        foto_url = '',
		        senha_hash = NULL,
		        provedor = 'removido',
		        mfa_segredo_cifrado = NULL,
		        mfa_ativo = false,
		        mfa_ultimo_codigo_usado = NULL,
		        mfa_ultimo_codigo_em = NULL,
		        email_pendente = NULL,
		        email_token_hash = NULL,
		        email_token_expira_em = NULL,
		        atualizado_em = now()
		  WHERE id = $1`,
		userID)
	if err != nil {
		return fmt.Errorf("store: anonimizar usuário: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNaoEncontrado
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("store: commit da exclusão de conta: %w", err)
	}
	return nil
}

// SolicitarTrocaEmail grava o e-mail pendente e o hash (sha256, hex) do token
// de confirmação (RF18, RN08) — `email` (login) só muda em ConfirmarTrocaEmail.
// Recusa com ErrEmailJaCadastrado se novoEmail já pertencer a outra conta de
// senha (mesma checagem lógica da constraint parcial da migração 0014, que só
// se aplica à coluna `email`, não a `email_pendente`).
func (s *Store) SolicitarTrocaEmail(ctx context.Context, userID, novoEmail, tokenHash string, expiraEm time.Time) error {
	var emailEmUso bool
	if err := s.db.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM users WHERE provedor = 'senha' AND email = $1)`, novoEmail,
	).Scan(&emailEmUso); err != nil {
		return fmt.Errorf("store: verificar e-mail em uso: %w", err)
	}
	if emailEmUso {
		return ErrEmailJaCadastrado
	}

	tag, err := s.db.Exec(ctx,
		`UPDATE users
		    SET email_pendente = $2, email_token_hash = $3, email_token_expira_em = $4, atualizado_em = now()
		  WHERE id = $1`,
		userID, novoEmail, tokenHash, expiraEm)
	if err != nil {
		return fmt.Errorf("store: solicitar troca de e-mail: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNaoEncontrado
	}
	return nil
}

// ConfirmarTrocaEmail efetiva email = email_pendente (RF18, RN08) quando o
// hash do token bate e ainda não expirou. Token inexistente e token expirado
// devolvem o mesmo ErrNaoEncontrado — mesmo racional de erroCredenciaisInvalidas
// em grpc.go (RN04): não distinguir "não existe" de "expirou".
func (s *Store) ConfirmarTrocaEmail(ctx context.Context, tokenHash string) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE users
		    SET email = email_pendente, email_pendente = NULL, email_token_hash = NULL, email_token_expira_em = NULL, atualizado_em = now()
		  WHERE email_token_hash = $1 AND email_token_expira_em > now()`,
		tokenHash)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		// Corrida rara: outra conta confirmou o mesmo e-mail pendente entre a
		// solicitação e esta confirmação (SolicitarTrocaEmail só checa unicidade
		// no momento do POST, não há lock entre as duas etapas).
		return ErrEmailJaCadastrado
	}
	if err != nil {
		return fmt.Errorf("store: confirmar troca de e-mail: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNaoEncontrado
	}
	return nil
}
