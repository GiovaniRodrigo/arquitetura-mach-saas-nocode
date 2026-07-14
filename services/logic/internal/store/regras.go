package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/machv4/platform/pkg/tenantctx"
)

// Regra é uma regra de negócio persistida (árvore de decisão em JSONB).
type Regra struct {
	ID            string
	SistemaID     string
	ArvoreDecisao []byte
}

// CriarRegra insere uma nova regra e devolve o id gerado.
func (s *Store) CriarRegra(ctx context.Context, r Regra) (string, error) {
	tid := tenantctx.TenantID(ctx)
	var id string
	err := s.db.WithTenant(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`INSERT INTO regras_negocio (sistema_id, tenant_id, arvore_decisao)
			 VALUES ($1, $2, $3) RETURNING id`,
			r.SistemaID, tid, r.ArvoreDecisao,
		).Scan(&id)
	})
	if err != nil {
		return "", fmt.Errorf("store: criar regra: %w", err)
	}
	return id, nil
}

// ObterRegra devolve uma regra por id, restrita ao tenant do contexto.
func (s *Store) ObterRegra(ctx context.Context, id string) (Regra, error) {
	var r Regra
	err := s.db.WithTenant(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`SELECT id, sistema_id, arvore_decisao FROM regras_negocio WHERE id = $1`, id,
		).Scan(&r.ID, &r.SistemaID, &r.ArvoreDecisao)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Regra{}, ErrNaoEncontrado
	}
	if err != nil {
		return Regra{}, fmt.Errorf("store: obter regra: %w", err)
	}
	return r, nil
}

// AtualizarRegra substitui a árvore de decisão de uma regra existente.
func (s *Store) AtualizarRegra(ctx context.Context, r Regra) error {
	return s.db.WithTenant(ctx, func(ctx context.Context, tx pgx.Tx) error {
		tag, e := tx.Exec(ctx,
			`UPDATE regras_negocio SET arvore_decisao = $2 WHERE id = $1`, r.ID, r.ArvoreDecisao)
		if e != nil {
			return e
		}
		if tag.RowsAffected() == 0 {
			return ErrNaoEncontrado
		}
		return nil
	})
}

// RemoverRegra apaga uma regra do tenant corrente.
func (s *Store) RemoverRegra(ctx context.Context, id string) error {
	return s.db.WithTenant(ctx, func(ctx context.Context, tx pgx.Tx) error {
		tag, e := tx.Exec(ctx, `DELETE FROM regras_negocio WHERE id = $1`, id)
		if e != nil {
			return e
		}
		if tag.RowsAffected() == 0 {
			return ErrNaoEncontrado
		}
		return nil
	})
}
