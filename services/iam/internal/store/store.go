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
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/machv4/platform/pkg/tenantctx"
	"github.com/machv4/platform/services/iam/internal/permissions"
)

// ErrNaoEncontrado indica tenant inexistente.
var ErrNaoEncontrado = errors.New("store: tenant não encontrado")

// ErrSemTenant indica ausência de tenant no contexto ao consultar dados isolados.
var ErrSemTenant = errors.New("store: contexto sem tenant (RN01)")

// Tenant é um nó da hierarquia Dono → Parceiro → Cliente Final.
type Tenant struct {
	ID       string
	ParentID *string
	Nome     string
	Tipo     string
}

// DB é o subconjunto de pgxpool.Pool usado pelo store (facilita testes).
type DB interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
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
