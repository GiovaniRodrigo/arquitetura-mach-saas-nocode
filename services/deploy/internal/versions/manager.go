// Package versions gere o ciclo de publicação por flag ativa e o rollback
// instantâneo (RF04, RN04, RN05). Toda mutação corre numa única transação via
// ScopedDB; a unicidade de "no máximo uma versão ativa por sistema" é imposta pelo
// índice único parcial uq_versoes_sistema_ativa (migração 0003), não por lógica de
// aplicação — publicações concorrentes colidem no banco e uma falha limpa.
package versions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/machv4/platform/pkg/database"
	"github.com/machv4/platform/pkg/tenantctx"
)

var (
	// ErrPublicacaoConcorrente sinaliza corrida de publicação para o mesmo sistema
	// (violação de índice único) — o Gateway responde 409 PUBLISH_IN_PROGRESS.
	ErrPublicacaoConcorrente = errors.New("versions: publicação concorrente para o sistema (RN04)")
	// ErrSemVersaoAnterior indica ausência de versão-alvo para o rollback.
	ErrSemVersaoAnterior = errors.New("versions: não há versão anterior para reverter (RN05)")
	// ErrSemVersaoAtiva indica que o sistema não tem versão ativa.
	ErrSemVersaoAtiva = errors.New("versions: nenhuma versão ativa")
	// ErrSistemaInexistente indica sistema fora do tenant corrente (ou inexistente).
	ErrSistemaInexistente = errors.New("versions: sistema inexistente no tenant")
)

// Manager opera o versionamento isolado por tenant.
type Manager struct {
	db *database.ScopedDB
}

// New cria um Manager sobre um pool/conn pgx (satisfaz database.Beginner).
func New(db database.Beginner) *Manager {
	return &Manager{db: database.New(db)}
}

// Publicacao é o resultado de publicar.
type Publicacao struct {
	VersaoID string
	Numero   int32
}

// VersaoAtiva é a versão ativa consolidada (consumida pelo Frontend).
type VersaoAtiva struct {
	VersaoID      string
	Numero        int32
	DefinicaoJSON []byte
}

// Publicar cria uma nova versão ativa a partir do estado atual dos designs e
// campos do sistema, desativando a anterior — tudo numa transação (RN04). Uma
// corrida entre duas publicações resolve-se no índice único: a perdedora recebe
// ErrPublicacaoConcorrente.
func (m *Manager) Publicar(ctx context.Context, sistemaID string) (Publicacao, error) {
	tid := tenantctx.TenantID(ctx)
	var pub Publicacao

	err := m.db.WithTenant(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if err := exigirSistema(ctx, tx, sistemaID); err != nil {
			return err
		}
		def, err := consolidar(ctx, tx, sistemaID)
		if err != nil {
			return err
		}

		var numero int32
		if err := tx.QueryRow(ctx,
			`SELECT COALESCE(MAX(numero),0)+1 FROM versoes_sistema WHERE sistema_id=$1`, sistemaID,
		).Scan(&numero); err != nil {
			return err
		}

		// Desativa a atual e insere a nova como ativa. O índice único parcial
		// garante a invariante mesmo sob concorrência.
		if _, err := tx.Exec(ctx,
			`UPDATE versoes_sistema SET ativa=false WHERE sistema_id=$1 AND ativa=true`, sistemaID); err != nil {
			return err
		}
		if err := tx.QueryRow(ctx,
			`INSERT INTO versoes_sistema (sistema_id, tenant_id, definicao_json, ativa, numero)
			 VALUES ($1,$2,$3,true,$4) RETURNING id`,
			sistemaID, tid, def, numero,
		).Scan(&pub.VersaoID); err != nil {
			return err
		}
		pub.Numero = numero
		return nil
	})
	if err != nil {
		if isUniqueViolation(err) {
			return Publicacao{}, ErrPublicacaoConcorrente
		}
		if errors.Is(err, ErrSistemaInexistente) {
			return Publicacao{}, err
		}
		return Publicacao{}, fmt.Errorf("versions: publicar: %w", err)
	}
	return pub, nil
}

// Rollback reativa a versão indicada (ou a imediatamente anterior à ativa, quando
// versaoNumero=0) apenas alternando a flag — sem downtime (RN05, RNF05).
func (m *Manager) Rollback(ctx context.Context, sistemaID string, versaoNumero int32) (int32, error) {
	var ativa int32

	err := m.db.WithTenant(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var alvo int32
		if versaoNumero > 0 {
			var existe bool
			if err := tx.QueryRow(ctx,
				`SELECT EXISTS(SELECT 1 FROM versoes_sistema WHERE sistema_id=$1 AND numero=$2)`,
				sistemaID, versaoNumero).Scan(&existe); err != nil {
				return err
			}
			if !existe {
				return ErrSemVersaoAnterior
			}
			alvo = versaoNumero
		} else {
			// Maior numero estritamente menor que o da versão ativa atual.
			err := tx.QueryRow(ctx,
				`SELECT numero FROM versoes_sistema
				  WHERE sistema_id=$1
				    AND numero < (SELECT numero FROM versoes_sistema WHERE sistema_id=$1 AND ativa=true)
				  ORDER BY numero DESC LIMIT 1`, sistemaID).Scan(&alvo)
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrSemVersaoAnterior
			}
			if err != nil {
				return err
			}
		}

		if _, err := tx.Exec(ctx,
			`UPDATE versoes_sistema SET ativa=false WHERE sistema_id=$1 AND ativa=true`, sistemaID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx,
			`UPDATE versoes_sistema SET ativa=true WHERE sistema_id=$1 AND numero=$2`, sistemaID, alvo); err != nil {
			return err
		}
		ativa = alvo
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrSemVersaoAnterior) {
			return 0, err
		}
		if isUniqueViolation(err) {
			return 0, ErrPublicacaoConcorrente
		}
		return 0, fmt.Errorf("versions: rollback: %w", err)
	}
	return ativa, nil
}

// ObterAtiva devolve a versão ativa consolidada do sistema.
func (m *Manager) ObterAtiva(ctx context.Context, sistemaID string) (VersaoAtiva, error) {
	var v VersaoAtiva
	err := m.db.WithTenant(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`SELECT id, numero, definicao_json FROM versoes_sistema WHERE sistema_id=$1 AND ativa=true`,
			sistemaID).Scan(&v.VersaoID, &v.Numero, &v.DefinicaoJSON)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return VersaoAtiva{}, ErrSemVersaoAtiva
	}
	if err != nil {
		return VersaoAtiva{}, fmt.Errorf("versions: versão ativa: %w", err)
	}
	return v, nil
}

// VersaoResumo é uma versão listada (sem a definição consolidada completa).
type VersaoResumo struct {
	ID       string
	Numero   int32
	Ativa    bool
	CriadoEm time.Time
}

// ListarVersoes devolve todas as versões do sistema, mais recente primeiro (RF12).
func (m *Manager) ListarVersoes(ctx context.Context, sistemaID string) ([]VersaoResumo, error) {
	var versoes []VersaoResumo
	err := m.db.WithTenant(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if err := exigirSistema(ctx, tx, sistemaID); err != nil {
			return err
		}
		rows, err := tx.Query(ctx,
			`SELECT id, numero, ativa, criado_em FROM versoes_sistema WHERE sistema_id=$1 ORDER BY numero DESC`,
			sistemaID)
		if err != nil {
			return err
		}
		defer rows.Close()
		versoes = []VersaoResumo{}
		for rows.Next() {
			var v VersaoResumo
			if err := rows.Scan(&v.ID, &v.Numero, &v.Ativa, &v.CriadoEm); err != nil {
				return err
			}
			versoes = append(versoes, v)
		}
		return rows.Err()
	})
	if err != nil {
		if errors.Is(err, ErrSistemaInexistente) {
			return nil, err
		}
		return nil, fmt.Errorf("versions: listar versões: %w", err)
	}
	return versoes, nil
}

// RollbackPorID resolve um id de versão para o número correspondente e delega a
// Rollback (mesma semântica/garantias de RN05) — usado pela fachada REST de
// versoes/{versaoId}/reverter (spec 004, RF12), que referencia versão por id em
// vez de número.
func (m *Manager) RollbackPorID(ctx context.Context, sistemaID, versaoID string) (int32, error) {
	var numero int32
	err := m.db.WithTenant(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if err := exigirSistema(ctx, tx, sistemaID); err != nil {
			return err
		}
		err := tx.QueryRow(ctx,
			`SELECT numero FROM versoes_sistema WHERE id=$1 AND sistema_id=$2`,
			versaoID, sistemaID).Scan(&numero)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrSemVersaoAnterior
		}
		return err
	})
	if err != nil {
		if errors.Is(err, ErrSistemaInexistente) {
			return 0, err
		}
		if errors.Is(err, ErrSemVersaoAnterior) {
			return 0, err
		}
		return 0, fmt.Errorf("versions: resolver versão por id: %w", err)
	}
	return m.Rollback(ctx, sistemaID, numero)
}

// exigirSistema confirma que o sistema pertence ao tenant corrente (RLS). Sem isso
// publicar-se-ia contra um sistema de outro tenant.
func exigirSistema(ctx context.Context, tx pgx.Tx, sistemaID string) error {
	var existe bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM sistemas WHERE id=$1)`, sistemaID).Scan(&existe); err != nil {
		return err
	}
	if !existe {
		return ErrSistemaInexistente
	}
	return nil
}

// consolidar monta a definicao_json a partir do estado atual: os designs (árvore
// de UI) e os campos (schema) do sistema — o que o Frontend renderiza.
func consolidar(ctx context.Context, tx pgx.Tx, sistemaID string) ([]byte, error) {
	type designOut struct {
		ID     string          `json:"id"`
		Nome   string          `json:"nome"`
		Arvore json.RawMessage `json:"arvore"`
	}

	drows, err := tx.Query(ctx,
		`SELECT id, nome, arvore FROM designs WHERE sistema_id=$1 ORDER BY criado_em`, sistemaID)
	if err != nil {
		return nil, err
	}
	designs := []designOut{}
	for drows.Next() {
		var (
			d      designOut
			arvore []byte
		)
		if err := drows.Scan(&d.ID, &d.Nome, &arvore); err != nil {
			drows.Close()
			return nil, err
		}
		d.Arvore = arvore
		designs = append(designs, d)
	}
	drows.Close()
	if err := drows.Err(); err != nil {
		return nil, err
	}

	campos := map[string]json.RawMessage{}
	crows, err := tx.Query(ctx,
		`SELECT blind_index,
		        jsonb_build_object('tipo', tipo, 'obrigatorio', obrigatorio, 'limites', limites)
		   FROM campos_definicao WHERE sistema_id=$1`, sistemaID)
	if err != nil {
		return nil, err
	}
	for crows.Next() {
		var (
			bi  string
			def []byte
		)
		if err := crows.Scan(&bi, &def); err != nil {
			crows.Close()
			return nil, err
		}
		campos[bi] = def
	}
	crows.Close()
	if err := crows.Err(); err != nil {
		return nil, err
	}

	return json.Marshal(struct {
		Designs []designOut                `json:"designs"`
		Campos  map[string]json.RawMessage `json:"campos"`
	}{Designs: designs, Campos: campos})
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
