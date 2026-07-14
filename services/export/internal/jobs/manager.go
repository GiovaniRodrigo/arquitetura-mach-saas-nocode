// Package jobs gere o ciclo de vida dos Jobs de exportação (RF05): criação,
// transições de estado e o resultado (Presigned URL). Persistência isolada por
// tenant via ScopedDB.
package jobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/machv4/platform/pkg/database"
	"github.com/machv4/platform/pkg/tenantctx"
)

// Estados do Job (enum job_status na migração 0008).
const (
	StatusCriado    = "criado"
	StatusColetando = "coletando"
	StatusPronto    = "pronto"
	StatusErro      = "erro"
	StatusExpirado  = "expirado"
)

var (
	// ErrNaoEncontrado indica Job inexistente no tenant corrente.
	ErrNaoEncontrado = errors.New("jobs: job não encontrado")
	// ErrTransicaoInvalida indica mudança de estado não permitida.
	ErrTransicaoInvalida = errors.New("jobs: transição de estado inválida")
)

// transicoes define o autómato de estados permitido.
var transicoes = map[string][]string{
	StatusCriado:    {StatusColetando, StatusErro, StatusExpirado},
	StatusColetando: {StatusPronto, StatusErro, StatusExpirado},
	StatusPronto:    {StatusExpirado},
	StatusErro:      {},
	StatusExpirado:  {},
}

// TransicaoValida indica se `de → para` é permitida.
func TransicaoValida(de, para string) bool {
	for _, p := range transicoes[de] {
		if p == para {
			return true
		}
	}
	return false
}

// Job é o estado persistido de uma exportação.
type Job struct {
	ID         string
	SistemaID  string
	Status     string
	ArquivoURL string
	ExpiraEm   *time.Time
}

// Manager persiste e transiciona Jobs isolados por tenant.
type Manager struct {
	db *database.ScopedDB
}

// New cria um Manager sobre um pool/conn pgx (satisfaz database.Beginner).
func New(db database.Beginner) *Manager {
	return &Manager{db: database.New(db)}
}

// Criar insere um Job no estado inicial `criado` e devolve o id.
func (m *Manager) Criar(ctx context.Context, sistemaID string) (string, error) {
	tid := tenantctx.TenantID(ctx)
	var id string
	err := m.db.WithTenant(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`INSERT INTO jobs_exportacao (tenant_id, sistema_id, status)
			 VALUES ($1, $2, $3) RETURNING id`,
			tid, sistemaID, StatusCriado,
		).Scan(&id)
	})
	if err != nil {
		return "", fmt.Errorf("jobs: criar: %w", err)
	}
	return id, nil
}

// Obter devolve o Job por id, restrito ao tenant do contexto.
func (m *Manager) Obter(ctx context.Context, id string) (Job, error) {
	var j Job
	err := m.db.WithTenant(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`SELECT id, sistema_id, status::text, coalesce(arquivo_url,''), expira_em
			   FROM jobs_exportacao WHERE id=$1`, id,
		).Scan(&j.ID, &j.SistemaID, &j.Status, &j.ArquivoURL, &j.ExpiraEm)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Job{}, ErrNaoEncontrado
	}
	if err != nil {
		return Job{}, fmt.Errorf("jobs: obter: %w", err)
	}
	return j, nil
}

// Transicionar muda o estado do Job validando o autómato (RF05).
func (m *Manager) Transicionar(ctx context.Context, id, novo string) error {
	return m.db.WithTenant(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var atual string
		if err := tx.QueryRow(ctx, `SELECT status::text FROM jobs_exportacao WHERE id=$1`, id).Scan(&atual); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNaoEncontrado
			}
			return err
		}
		if !TransicaoValida(atual, novo) {
			return fmt.Errorf("%w: %s → %s", ErrTransicaoInvalida, atual, novo)
		}
		_, err := tx.Exec(ctx, `UPDATE jobs_exportacao SET status=$2 WHERE id=$1`, id, novo)
		return err
	})
}

// MarcarPronto conclui o Job com a URL e a expiração, validando a transição.
func (m *Manager) MarcarPronto(ctx context.Context, id, url string, expiraEm time.Time) error {
	return m.db.WithTenant(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var atual string
		if err := tx.QueryRow(ctx, `SELECT status::text FROM jobs_exportacao WHERE id=$1`, id).Scan(&atual); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNaoEncontrado
			}
			return err
		}
		if !TransicaoValida(atual, StatusPronto) {
			return fmt.Errorf("%w: %s → %s", ErrTransicaoInvalida, atual, StatusPronto)
		}
		_, err := tx.Exec(ctx,
			`UPDATE jobs_exportacao SET status=$2, arquivo_url=$3, expira_em=$4 WHERE id=$1`,
			id, StatusPronto, url, expiraEm)
		return err
	})
}
