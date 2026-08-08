// Persistência do White Label (RF13, RNF03), sempre através de ScopedDB — que
// fixa app.tenant_id e recusa acesso sem tenant no contexto. A RLS (política
// tenant_isolation, migração 0015) garante o isolamento por tenant tanto na
// leitura quanto na escrita (WITH CHECK).
package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/machv4/platform/pkg/tenantctx"
)

// WhiteLabel espelha a personalização de marca de um tenant.
type WhiteLabel struct {
	LogoURL         string
	CorPrimaria     string
	CorSecundaria   string
	DominioProprio  string
	DominioValidado bool
}

// AtualizarWhiteLabel faz upsert da personalização de marca do tenant do
// contexto. A validação do domínio próprio é assíncrona e fora de escopo desta
// entrega: dominio_validado é sempre gravado como false (RNF03).
func (s *Store) AtualizarWhiteLabel(ctx context.Context, dados WhiteLabel) (WhiteLabel, error) {
	tid := tenantctx.TenantID(ctx)

	var out WhiteLabel
	err := s.db.WithTenant(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`INSERT INTO config_white_label (tenant_id, logo_url, cor_primaria, cor_secundaria, dominio_proprio, dominio_validado)
			 VALUES ($1, $2, $3, $4, $5, false)
			 ON CONFLICT (tenant_id) DO UPDATE
			   SET logo_url = EXCLUDED.logo_url,
			       cor_primaria = EXCLUDED.cor_primaria,
			       cor_secundaria = EXCLUDED.cor_secundaria,
			       dominio_proprio = EXCLUDED.dominio_proprio,
			       dominio_validado = false,
			       atualizado_em = now()
			 RETURNING logo_url, cor_primaria, cor_secundaria, dominio_proprio, dominio_validado`,
			tid, dados.LogoURL, dados.CorPrimaria, dados.CorSecundaria, dados.DominioProprio,
		).Scan(&out.LogoURL, &out.CorPrimaria, &out.CorSecundaria, &out.DominioProprio, &out.DominioValidado)
	})
	if err != nil {
		return WhiteLabel{}, fmt.Errorf("store: atualizar white label: %w", err)
	}
	return out, nil
}
