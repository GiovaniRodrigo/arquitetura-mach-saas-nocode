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

// --- Dashboard (spec 004, RF04/RF05/RF06) --------------------------------
//
// Os 3 cards do Dashboard (Últimos Acessos, Feedback, Resumo Financeiro)
// agregam dados do tenant do usuário autenticado + seus filhos diretos ao
// mesmo tempo ("tenants vinculados", consistente com RF07/RN05). Isso é
// incompatível com ScopedDB.WithTenant (que fixa exatamente 1 app.tenant_id),
// então eventos_login/feedback/assinaturas_tenant ficam fora da Row-Level
// Security (ver comentário de topo da migração 0010) e o filtro por tenant é
// feito manualmente aqui via WHERE tenant_id = ANY($lista) — mesmo padrão já
// usado para tenants/users nas funções acima.

// EventoLogin é uma linha do histórico de login para o card "Últimos Acessos"
// (RF04).
type EventoLogin struct {
	UsuarioNome string
	TenantNome  string
	CriadoEm    time.Time
}

// Feedback é uma mensagem de feedback/reclamação de um tenant vinculado
// (RF05).
type Feedback struct {
	ID         string
	TenantNome string
	Mensagem   string
	Status     string
	CriadoEm   time.Time
}

// ResumoFinanceiro é a receita agregada do mês corrente dos tenants
// vinculados (RF06).
type ResumoFinanceiro struct {
	ReceitaTotalCentavos int64
	Moeda                string
	Competencia          string
}

// tenantsVinculados devolve o tenant do contexto + seus filhos diretos — o
// conjunto usado pelos 3 cards do Dashboard (Últimos Acessos, Feedback, Resumo
// Financeiro), consistente com RF07/RN05 já implementado (spec 004).
func (s *Store) tenantsVinculados(ctx context.Context, tenantID string) ([]string, error) {
	filhos, err := s.ListarFilhos(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(filhos)+1)
	ids = append(ids, tenantID)
	for _, f := range filhos {
		ids = append(ids, f.ID)
	}
	return ids, nil
}

// RegistrarEventoLogin grava um evento de login para telemetria do card
// "Últimos Acessos" (RF04). Chamado a partir de AutenticarSenha/
// AutenticarThirdParty; falha aqui nunca deve impedir o login (é telemetria
// auxiliar, não o fluxo crítico).
func (s *Store) RegistrarEventoLogin(ctx context.Context, usuarioID, tenantID string) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO eventos_login (usuario_id, tenant_id) VALUES ($1, $2)`,
		usuarioID, tenantID,
	)
	if err != nil {
		return fmt.Errorf("store: registrar evento de login: %w", err)
	}
	return nil
}

// UltimosAcessos devolve os `limite` logins mais recentes dos tenants
// vinculados a tenantID (tenant do contexto + filhos diretos), mais recente
// primeiro (RF04, RN02).
func (s *Store) UltimosAcessos(ctx context.Context, tenantID string, limite int) ([]EventoLogin, error) {
	vinculados, err := s.tenantsVinculados(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("store: últimos acessos: %w", err)
	}
	rows, err := s.db.Query(ctx,
		`SELECT u.nome, t.nome, e.criado_em
		   FROM eventos_login e
		   JOIN users u ON u.id = e.usuario_id
		   JOIN tenants t ON t.id = e.tenant_id
		  WHERE e.tenant_id = ANY($1)
		  ORDER BY e.criado_em DESC
		  LIMIT $2`,
		vinculados, limite,
	)
	if err != nil {
		return nil, fmt.Errorf("store: últimos acessos: %w", err)
	}
	defer rows.Close()

	var out []EventoLogin
	for rows.Next() {
		var ev EventoLogin
		if err := rows.Scan(&ev.UsuarioNome, &ev.TenantNome, &ev.CriadoEm); err != nil {
			return nil, err
		}
		out = append(out, ev)
	}
	return out, rows.Err()
}

// ListarFeedback devolve as mensagens de feedback dos tenants vinculados a
// tenantID, mais recente primeiro, opcionalmente filtradas por status (RF05).
// status == nil devolve todas.
func (s *Store) ListarFeedback(ctx context.Context, tenantID string, status *string) ([]Feedback, error) {
	vinculados, err := s.tenantsVinculados(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("store: listar feedback: %w", err)
	}

	sql := `SELECT f.id, t.nome, f.mensagem, f.status::text, f.criado_em
	          FROM feedback f
	          JOIN tenants t ON t.id = f.tenant_id
	         WHERE f.tenant_id = ANY($1)`
	args := []any{vinculados}
	if status != nil {
		sql += ` AND f.status = $2`
		args = append(args, *status)
	}
	sql += ` ORDER BY f.criado_em DESC`

	rows, err := s.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("store: listar feedback: %w", err)
	}
	defer rows.Close()

	var out []Feedback
	for rows.Next() {
		var f Feedback
		if err := rows.Scan(&f.ID, &f.TenantNome, &f.Mensagem, &f.Status, &f.CriadoEm); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// AtualizarStatusFeedback aplica a transição de status de uma mensagem de
// feedback. RN03: a única transição válida é pendente→respondido — o
// UPDATE só afeta linhas com status='pendente' quando novoStatus='respondido',
// tornando a transição reversa (respondido→pendente) impossível mesmo sob
// corrida. O chamador (grpc.go) já rejeita novoStatus="pendente" antes de
// chegar aqui.
func (s *Store) AtualizarStatusFeedback(ctx context.Context, id, novoStatus string) (Feedback, error) {
	var f Feedback
	err := s.db.QueryRow(ctx,
		`UPDATE feedback f SET status = $2
		   FROM tenants t
		  WHERE f.id = $1 AND f.status = 'pendente' AND t.id = f.tenant_id
		  RETURNING f.id, t.nome, f.mensagem, f.status::text, f.criado_em`,
		id, novoStatus,
	).Scan(&f.ID, &f.TenantNome, &f.Mensagem, &f.Status, &f.CriadoEm)
	if errors.Is(err, pgx.ErrNoRows) {
		return Feedback{}, ErrNaoEncontrado
	}
	if err != nil {
		return Feedback{}, fmt.Errorf("store: atualizar status de feedback: %w", err)
	}
	return f, nil
}

// ResumoFinanceiro soma a receita de assinatura do mês corrente dos tenants
// vinculados a tenantID (RF06, RN04). Sem linhas, devolve 0/BRL (COALESCE
// cobre o SUM de conjunto vazio, que seria NULL). Não existe motor de billing
// real neste repo (fora de escopo — spec.md §8): a tabela é lida como está,
// semeada manualmente nos testes de integração.
func (s *Store) ResumoFinanceiro(ctx context.Context, tenantID string) (ResumoFinanceiro, error) {
	vinculados, err := s.tenantsVinculados(ctx, tenantID)
	if err != nil {
		return ResumoFinanceiro{}, fmt.Errorf("store: resumo financeiro: %w", err)
	}

	var totalCentavos int64
	err = s.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(valor_centavos), 0)
		   FROM assinaturas_tenant
		  WHERE tenant_id = ANY($1) AND to_char(competencia, 'YYYY-MM') = to_char(now(), 'YYYY-MM')`,
		vinculados,
	).Scan(&totalCentavos)
	if err != nil {
		return ResumoFinanceiro{}, fmt.Errorf("store: resumo financeiro: %w", err)
	}

	return ResumoFinanceiro{
		ReceitaTotalCentavos: totalCentavos,
		Moeda:                "BRL",
		Competencia:          time.Now().Format("2006-01"),
	}, nil
}
