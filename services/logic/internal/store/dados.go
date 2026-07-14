package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/machv4/platform/pkg/tenantctx"
)

// InserirDados grava uma submissão de formulário já validada em
// dados_operacionais (mapa blind_index → valor na coluna JSONB) — RF07.
func (s *Store) InserirDados(ctx context.Context, sistemaID string, valores map[string]string) (string, error) {
	raw, err := json.Marshal(valores)
	if err != nil {
		return "", fmt.Errorf("store: serializar valores: %w", err)
	}
	tid := tenantctx.TenantID(ctx)

	var id string
	err = s.db.WithTenant(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`INSERT INTO dados_operacionais (tenant_id, sistema_id, valores)
			 VALUES ($1, $2, $3) RETURNING id`,
			tid, sistemaID, raw,
		).Scan(&id)
	})
	if err != nil {
		return "", fmt.Errorf("store: inserir dados: %w", err)
	}
	return id, nil
}
