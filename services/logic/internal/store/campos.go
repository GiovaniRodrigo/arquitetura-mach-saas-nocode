// Package store persiste o schema (campos_definicao), as regras de negócio
// (regras_negocio) e as submissões (dados_operacionais), sempre através de
// ScopedDB — que fixa app.tenant_id e recusa acesso sem tenant (RN01, RN02).
package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/machv4/platform/pkg/database"
	"github.com/machv4/platform/pkg/tenantctx"
	"github.com/machv4/platform/services/logic/internal/validation"
)

// ErrNaoEncontrado indica registro inexistente no tenant corrente.
var ErrNaoEncontrado = errors.New("store: registro não encontrado")

// Store agrega o acesso isolado por tenant às tabelas do Logic Engine.
type Store struct {
	db *database.ScopedDB
}

// New cria um Store sobre um pool/conn pgx (satisfaz database.Beginner).
func New(db database.Beginner) *Store {
	return &Store{db: database.New(db)}
}

// DefinirCampo faz upsert da definição de um campo (PK composta blind_index +
// sistema_id). É o C/U do CRUD de campos_definicao indexado por blind_index (RN02).
func (s *Store) DefinirCampo(ctx context.Context, sistemaID string, c validation.CampoDef) error {
	limites, err := json.Marshal(c.Limites)
	if err != nil {
		return fmt.Errorf("store: serializar limites: %w", err)
	}
	tid := tenantctx.TenantID(ctx)

	return s.db.WithTenant(ctx, func(ctx context.Context, tx pgx.Tx) error {
		_, e := tx.Exec(ctx,
			`INSERT INTO campos_definicao (blind_index, sistema_id, tenant_id, tipo, obrigatorio, limites)
			 VALUES ($1, $2, $3, $4, $5, $6)
			 ON CONFLICT (blind_index, sistema_id) DO UPDATE
			   SET tipo = EXCLUDED.tipo, obrigatorio = EXCLUDED.obrigatorio, limites = EXCLUDED.limites`,
			c.BlindIndex, sistemaID, tid, c.Tipo, c.Obrigatorio, limites)
		return e
	})
}

// RemoverCampo apaga uma definição de campo (D do CRUD).
func (s *Store) RemoverCampo(ctx context.Context, sistemaID, blindIndex string) error {
	return s.db.WithTenant(ctx, func(ctx context.Context, tx pgx.Tx) error {
		tag, e := tx.Exec(ctx,
			`DELETE FROM campos_definicao WHERE blind_index = $1 AND sistema_id = $2`, blindIndex, sistemaID)
		if e != nil {
			return e
		}
		if tag.RowsAffected() == 0 {
			return ErrNaoEncontrado
		}
		return nil
	})
}

// SchemaDe carrega todas as definições de campo de um sistema, indexadas por
// blind_index — a fonte da revalidação em SalvarFormulario (RN08).
func (s *Store) SchemaDe(ctx context.Context, sistemaID string) (map[string]validation.CampoDef, error) {
	schema := make(map[string]validation.CampoDef)
	err := s.db.WithTenant(ctx, func(ctx context.Context, tx pgx.Tx) error {
		rows, e := tx.Query(ctx,
			`SELECT blind_index, tipo, obrigatorio, limites FROM campos_definicao WHERE sistema_id = $1`, sistemaID)
		if e != nil {
			return e
		}
		defer rows.Close()

		for rows.Next() {
			var (
				c   validation.CampoDef
				lim []byte
			)
			if e := rows.Scan(&c.BlindIndex, &c.Tipo, &c.Obrigatorio, &lim); e != nil {
				return e
			}
			if len(lim) > 0 {
				if e := json.Unmarshal(lim, &c.Limites); e != nil {
					return fmt.Errorf("store: limites de %q: %w", c.BlindIndex, e)
				}
			}
			schema[c.BlindIndex] = c
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("store: schema: %w", err)
	}
	return schema, nil
}
