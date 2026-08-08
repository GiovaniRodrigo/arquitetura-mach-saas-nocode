package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/machv4/platform/pkg/tenantctx"
)

// RegraValidacao é uma regra de validação de estado de componente
// (regex/tamanho/obrigatório) persistida — entidade distinta de Regra (árvore
// de decisão do motor de ações), sem integração com o disparo de ações de
// SalvarFormulario nesta fase (RF10/RF11, RN06).
type RegraValidacao struct {
	ID           string
	SistemaID    string
	BlindIndexes []string
	Tipo         string
	Parametros   []byte // jsonb bruto
}

// ListarRegrasValidacao devolve todas as regras de validação de um sistema,
// restritas ao tenant do contexto.
func (s *Store) ListarRegrasValidacao(ctx context.Context, sistemaID string) ([]RegraValidacao, error) {
	var regras []RegraValidacao
	err := s.db.WithTenant(ctx, func(ctx context.Context, tx pgx.Tx) error {
		rows, e := tx.Query(ctx,
			`SELECT id, sistema_id, blind_indexes, tipo, parametros
			 FROM regras_validacao_componente WHERE sistema_id = $1`, sistemaID)
		if e != nil {
			return e
		}
		defer rows.Close()

		for rows.Next() {
			var r RegraValidacao
			if e := rows.Scan(&r.ID, &r.SistemaID, &r.BlindIndexes, &r.Tipo, &r.Parametros); e != nil {
				return e
			}
			regras = append(regras, r)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("store: listar regras de validação: %w", err)
	}
	return regras, nil
}

// CriarRegraValidacao insere uma nova regra de validação e devolve o registro
// completo, com o id gerado.
func (s *Store) CriarRegraValidacao(ctx context.Context, r RegraValidacao) (RegraValidacao, error) {
	tid := tenantctx.TenantID(ctx)
	err := s.db.WithTenant(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`INSERT INTO regras_validacao_componente (sistema_id, tenant_id, blind_indexes, tipo, parametros)
			 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			r.SistemaID, tid, r.BlindIndexes, r.Tipo, r.Parametros,
		).Scan(&r.ID)
	})
	if err != nil {
		return RegraValidacao{}, fmt.Errorf("store: criar regra de validação: %w", err)
	}
	return r, nil
}
