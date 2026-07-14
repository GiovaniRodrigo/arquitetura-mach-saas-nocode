// Package collector reúne, em chunks, os dados exportáveis de um sistema a partir
// das três fontes (design, regras, dados operacionais), sem carregar tudo em
// memória (RNF01). A emissão por linha mantém a RAM constante para grandes volumes.
package collector

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/machv4/platform/pkg/database"
)

// Origens dos chunks (também o valor do campo ChunkDados.origem).
const (
	OrigemDesign      = "design"
	OrigemRegras      = "regras"
	OrigemOperacional = "operacional"
)

// Emit recebe cada chunk coletado. Um erro aborta a coleta.
type Emit func(origem string, conteudo []byte) error

// Collector lê as fontes isoladas por tenant via ScopedDB.
type Collector struct {
	db *database.ScopedDB
}

// New cria um Collector sobre um pool/conn pgx.
func New(db database.Beginner) *Collector {
	return &Collector{db: database.New(db)}
}

// Coletar percorre as três fontes do sistema e chama emit por linha. Tudo dentro
// de uma transação com o tenant fixado (RN01); as consultas são sequenciais e
// cada cursor é fechado antes da próxima (exigência do pgx sobre a mesma conn).
func (c *Collector) Coletar(ctx context.Context, sistemaID string, emit Emit) error {
	return c.db.WithTenant(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if err := streamRows(ctx, tx, emit, OrigemDesign,
			`SELECT jsonb_build_object('id', id, 'nome', nome, 'arvore', arvore)
			   FROM designs WHERE sistema_id=$1 ORDER BY criado_em`, sistemaID); err != nil {
			return err
		}
		if err := streamRows(ctx, tx, emit, OrigemRegras,
			`SELECT jsonb_build_object('id', id, 'arvore_decisao', arvore_decisao)
			   FROM regras_negocio WHERE sistema_id=$1 ORDER BY criado_em`, sistemaID); err != nil {
			return err
		}
		return streamRows(ctx, tx, emit, OrigemOperacional,
			`SELECT jsonb_build_object('id', id, 'valores', valores, 'criado_em', criado_em)
			   FROM dados_operacionais WHERE sistema_id=$1 ORDER BY criado_em`, sistemaID)
	})
}

func streamRows(ctx context.Context, tx pgx.Tx, emit Emit, origem, sql, sistemaID string) error {
	rows, err := tx.Query(ctx, sql, sistemaID)
	if err != nil {
		return fmt.Errorf("collector: consultar %s: %w", origem, err)
	}
	defer rows.Close()

	for rows.Next() {
		var conteudo []byte
		if err := rows.Scan(&conteudo); err != nil {
			return fmt.Errorf("collector: ler %s: %w", origem, err)
		}
		if err := emit(origem, conteudo); err != nil {
			return err
		}
	}
	return rows.Err()
}
