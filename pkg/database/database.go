// Package database impõe o isolamento multi-tenant por construção (RN01).
//
// Nenhum serviço deve emitir SQL diretamente contra o pool: o acesso passa por
// ScopedDB.WithTenant, que (1) recusa qualquer operação sem tenant no contexto e
// (2) fixa o GUC app.tenant_id na transação, ativando a Row-Level Security da
// migração 0010 como segunda camada de defesa. Assim o filtro por tenant deixa de
// depender de cada desenvolvedor lembrar do WHERE (plan §2.2).
//
// Importante: a RLS é ignorada por superusuários e por roles com BYPASSRLS. Em
// runtime a aplicação DEVE conectar com uma role comum (sem SUPERUSER), caso
// contrário o app.tenant_id não terá efeito.
package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/machv4/platform/pkg/tenantctx"
)

// ErrNoTenant é retornado quando se tenta acessar dados sem tenant no contexto.
var ErrNoTenant = errors.New("database: operação sem contexto de tenant (RN01)")

// Beginner abre transações. É satisfeito por *pgxpool.Pool e por *pgx.Conn.
type Beginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// ScopedDB é o único ponto de entrada para dados isolados por tenant.
type ScopedDB struct {
	db Beginner
}

// New cria um ScopedDB sobre um pool/conn pgx.
func New(db Beginner) *ScopedDB {
	return &ScopedDB{db: db}
}

// WithTenant executa fn dentro de uma transação com app.tenant_id fixado no tenant
// corrente. Faz rollback automático se fn retornar erro; commit caso contrário.
// Retorna ErrNoTenant — sem tocar no banco — quando o contexto não tem tenant.
func (s *ScopedDB) WithTenant(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	tid := tenantctx.TenantID(ctx)
	if tid == "" {
		return ErrNoTenant
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	// Rollback pós-commit é no-op em pgx; garante liberação em qualquer saída.
	defer func() { _ = tx.Rollback(ctx) }()

	// set_config(nome, valor, is_local=true) tem escopo de transação (como
	// SET LOCAL) e aceita bind param — evita interpolar o tenant_id no SQL.
	if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", tid); err != nil {
		return err
	}

	if err := fn(ctx, tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// WithExplicitTenant executa fn com app.tenant_id fixado no tenantID informado
// diretamente (não o do contexto) — usado quando a autorização cross-tenant já
// foi validada por outra camada (ex.: Gateway validando hierarquia via IAM antes
// de repassar ao Design, RF08/RN05). Nunca chame isto sem essa validação prévia.
func (s *ScopedDB) WithExplicitTenant(ctx context.Context, tenantID string, fn func(ctx context.Context, tx pgx.Tx) error) error {
	if tenantID == "" {
		return ErrNoTenant
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, "SELECT set_config('app.tenant_id', $1, true)", tenantID); err != nil {
		return err
	}
	if err := fn(ctx, tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
