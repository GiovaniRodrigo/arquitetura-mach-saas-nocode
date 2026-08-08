// Persistência de sistemas (o contêiner de designs/versões, RN01), sempre através
// de ScopedDB — que fixa app.tenant_id e recusa acesso sem tenant no contexto. A
// RLS (política tenant_isolation, migração 0010) garante o isolamento por tenant
// tanto na leitura quanto na escrita (WITH CHECK).
package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	designv1 "github.com/machv4/platform/gen/go/construtor/design/v1"
	"github.com/machv4/platform/pkg/tenantctx"
)

// CriarSistema insere um sistema no tenant do contexto e devolve o id gerado. O
// tenant_id vem sempre do contexto (nunca do request); a política RLS WITH CHECK
// recusa qualquer tentativa de gravar sob outro tenant.
func (s *Store) CriarSistema(ctx context.Context, nome string) (string, error) {
	tid := tenantctx.TenantID(ctx)

	var id string
	err := s.db.WithTenant(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`INSERT INTO sistemas (tenant_id, nome) VALUES ($1, $2) RETURNING id`,
			tid, nome,
		).Scan(&id)
	})
	if err != nil {
		return "", fmt.Errorf("store: criar sistema: %w", err)
	}
	return id, nil
}

// ListarSistemas devolve os sistemas do tenant corrente, mais recentes primeiro.
// A RLS restringe as linhas ao tenant do contexto (RN01).
func (s *Store) ListarSistemas(ctx context.Context) ([]*designv1.Sistema, error) {
	var sistemas []*designv1.Sistema
	err := s.db.WithTenant(ctx, func(ctx context.Context, tx pgx.Tx) error {
		rows, e := tx.Query(ctx,
			`SELECT id, nome FROM sistemas ORDER BY criado_em DESC`)
		if e != nil {
			return e
		}
		defer rows.Close()
		for rows.Next() {
			var sis designv1.Sistema
			if e := rows.Scan(&sis.Id, &sis.Nome); e != nil {
				return e
			}
			sistemas = append(sistemas, &sis)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("store: listar sistemas: %w", err)
	}
	return sistemas, nil
}

// ListarSistemasDoTenant devolve os sistemas de um tenant explícito (não o do
// contexto), mais recentes primeiro — usado por RF08 quando o Gateway já
// validou (via IAM) que o tenant é filho direto do tenant do contexto (RN05).
// A RLS continua restringindo as linhas ao tenantID fixado na transação.
func (s *Store) ListarSistemasDoTenant(ctx context.Context, tenantID string) ([]*designv1.Sistema, error) {
	var sistemas []*designv1.Sistema
	err := s.db.WithExplicitTenant(ctx, tenantID, func(ctx context.Context, tx pgx.Tx) error {
		rows, e := tx.Query(ctx,
			`SELECT id, nome FROM sistemas ORDER BY criado_em DESC`)
		if e != nil {
			return e
		}
		defer rows.Close()
		for rows.Next() {
			var sis designv1.Sistema
			if e := rows.Scan(&sis.Id, &sis.Nome); e != nil {
				return e
			}
			sistemas = append(sistemas, &sis)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("store: listar sistemas do tenant: %w", err)
	}
	return sistemas, nil
}
