// Package store persiste designs (árvore Composite) na coluna JSONB
// `designs.arvore`, sempre através de ScopedDB — que fixa app.tenant_id e recusa
// qualquer acesso sem tenant no contexto (RF01, RN01).
package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	designv1 "github.com/machv4/platform/gen/go/construtor/design/v1"
	"github.com/machv4/platform/pkg/database"
	"github.com/machv4/platform/pkg/tenantctx"
	"github.com/machv4/platform/services/design/internal/tree"
)

// ErrNaoEncontrado indica design inexistente no tenant corrente.
var ErrNaoEncontrado = errors.New("store: design não encontrado")

// Store grava e lê designs isolados por tenant.
type Store struct {
	db *database.ScopedDB
}

// New cria um Store sobre um pool/conn pgx (satisfaz database.Beginner).
func New(db database.Beginner) *Store {
	return &Store{db: database.New(db)}
}

// Criar insere um novo design e devolve o id gerado. A árvore vem serializada
// para a coluna arvore (jsonb); tenant_id é o do contexto, nunca do request.
func (s *Store) Criar(ctx context.Context, d *designv1.Design) (string, error) {
	raw, err := tree.Serializar(d.GetArvore())
	if err != nil {
		return "", err
	}
	tid := tenantctx.TenantID(ctx)

	var id string
	err = s.db.WithTenant(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`INSERT INTO designs (tenant_id, sistema_id, nome, arvore)
			 VALUES ($1, $2, $3, $4)
			 RETURNING id`,
			tid, d.GetSistemaId(), d.GetNome(), raw,
		).Scan(&id)
	})
	if err != nil {
		return "", fmt.Errorf("store: criar design: %w", err)
	}
	return id, nil
}

// Obter devolve o design por id. A RLS restringe ao tenant corrente, então um id
// de outro tenant retorna ErrNaoEncontrado (RN01).
func (s *Store) Obter(ctx context.Context, id string) (*designv1.Design, error) {
	var (
		sistemaID, nome string
		raw             []byte
	)
	err := s.db.WithTenant(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`SELECT sistema_id, nome, arvore FROM designs WHERE id = $1`, id,
		).Scan(&sistemaID, &nome, &raw)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNaoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("store: obter design: %w", err)
	}

	arvore, err := tree.Desserializar(raw)
	if err != nil {
		return nil, err
	}
	return &designv1.Design{Id: id, SistemaId: sistemaID, Nome: nome, Arvore: arvore}, nil
}

// Atualizar substitui nome e árvore de um design existente. sistema_id é imutável.
func (s *Store) Atualizar(ctx context.Context, d *designv1.Design) error {
	raw, err := tree.Serializar(d.GetArvore())
	if err != nil {
		return err
	}
	return s.db.WithTenant(ctx, func(ctx context.Context, tx pgx.Tx) error {
		tag, e := tx.Exec(ctx,
			`UPDATE designs SET nome = $2, arvore = $3, atualizado_em = now() WHERE id = $1`,
			d.GetId(), d.GetNome(), raw)
		if e != nil {
			return e
		}
		if tag.RowsAffected() == 0 {
			return ErrNaoEncontrado
		}
		return nil
	})
}

// Remover apaga um design do tenant corrente.
func (s *Store) Remover(ctx context.Context, id string) error {
	return s.db.WithTenant(ctx, func(ctx context.Context, tx pgx.Tx) error {
		tag, e := tx.Exec(ctx, `DELETE FROM designs WHERE id = $1`, id)
		if e != nil {
			return e
		}
		if tag.RowsAffected() == 0 {
			return ErrNaoEncontrado
		}
		return nil
	})
}

// Salvar é o upsert da colaboração write-behind: o motor Elixir envia, após o
// debounce de 5s, o estado consolidado do design (id já conhecido) numa só
// chamada (RN06). Cria se ausente; caso contrário sobrescreve nome e árvore.
func (s *Store) Salvar(ctx context.Context, d *designv1.Design) error {
	raw, err := tree.Serializar(d.GetArvore())
	if err != nil {
		return err
	}
	tid := tenantctx.TenantID(ctx)

	return s.db.WithTenant(ctx, func(ctx context.Context, tx pgx.Tx) error {
		_, e := tx.Exec(ctx,
			`INSERT INTO designs (id, tenant_id, sistema_id, nome, arvore)
			 VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (id) DO UPDATE
			   SET nome = EXCLUDED.nome, arvore = EXCLUDED.arvore, atualizado_em = now()`,
			d.GetId(), tid, d.GetSistemaId(), d.GetNome(), raw)
		return e
	})
}
