//go:build integration

// Teste de integração do Deploy Engine + Postgres (RN04, RN05, RNF05, critério 3).
// Requer o Postgres do docker-compose com as migrações aplicadas. Executar:
//
//	DATABASE_URL=postgres://mach:mach@localhost:5432/machv4?sslmode=disable \
//	  go test -tags integration ./services/deploy/tests/...
package tests

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	commonv1 "github.com/machv4/platform/gen/go/construtor/common/v1"
	"github.com/machv4/platform/pkg/tenantctx"
	"github.com/machv4/platform/services/deploy/internal/versions"
)

const appRole = "machapp_deploy_rls_test"

func dsn() string {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v
	}
	return "postgres://mach:mach@localhost:5432/machv4?sslmode=disable"
}

// setup semeia (como superusuário) a role comum, um tenant e um sistema, e devolve
// o Manager sobre um pool que assume a role comum (RLS efetiva), o contexto de
// tenant e o sistema_id.
func setup(t *testing.T) (mgr *versions.Manager, ctx context.Context, sistemaID string) {
	t.Helper()
	bg := context.Background()

	admin, err := pgx.Connect(bg, dsn())
	if err != nil {
		t.Skipf("Postgres indisponível (%v)", err)
	}
	defer admin.Close(bg)

	for _, stmt := range []string{
		`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname='` + appRole + `') THEN CREATE ROLE ` + appRole + ` NOLOGIN NOSUPERUSER; END IF; END $$;`,
		`GRANT USAGE ON SCHEMA public TO ` + appRole + `;`,
		`GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO ` + appRole + `;`,
	} {
		if _, err := admin.Exec(bg, stmt); err != nil {
			t.Fatalf("setup role: %v (%s)", err, stmt)
		}
	}

	var tenantID string
	if err := admin.QueryRow(bg,
		`INSERT INTO tenants (nome, tipo, chave_blind_index) VALUES ('itg-deploy','dono','\x6b') RETURNING id`).Scan(&tenantID); err != nil {
		t.Fatalf("tenant: %v", err)
	}
	t.Cleanup(func() {
		c, err := pgx.Connect(context.Background(), dsn())
		if err != nil {
			return
		}
		defer c.Close(context.Background())
		_, _ = c.Exec(context.Background(), `DELETE FROM tenants WHERE id=$1`, tenantID)
	})

	if err := admin.QueryRow(bg,
		`INSERT INTO sistemas (tenant_id, nome) VALUES ($1,'sistema-deploy') RETURNING id`, tenantID).Scan(&sistemaID); err != nil {
		t.Fatalf("sistema: %v", err)
	}

	cfg, err := pgxpool.ParseConfig(dsn())
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}
	cfg.AfterConnect = func(ctx context.Context, c *pgx.Conn) error {
		_, err := c.Exec(ctx, "SET ROLE "+appRole)
		return err
	}
	pool, err := pgxpool.NewWithConfig(bg, cfg)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)

	ctx = tenantctx.NewContext(bg, &commonv1.TenantContext{TenantId: tenantID, Tipo: "dono"})
	return versions.New(pool), ctx, sistemaID
}

// contarAtivas conta as versões ativas de um sistema via superusuário (bypassa RLS).
func contarAtivas(t *testing.T, sistemaID string) int {
	t.Helper()
	c, err := pgx.Connect(context.Background(), dsn())
	if err != nil {
		t.Fatalf("conexão admin: %v", err)
	}
	defer c.Close(context.Background())
	var n int
	if err := c.QueryRow(context.Background(),
		`SELECT count(*) FROM versoes_sistema WHERE sistema_id=$1 AND ativa=true`, sistemaID).Scan(&n); err != nil {
		t.Fatalf("contagem: %v", err)
	}
	return n
}

// Critério 3 (parte 1): sob N publicações concorrentes, o índice único parcial
// garante exatamente uma versão ativa; as perdedoras falham com concorrência.
func TestUnicidadeFlagAtivaSobConcorrencia(t *testing.T) {
	mgr, ctx, sistemaID := setup(t)

	const n = 8
	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		sucessos int
	)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := mgr.Publicar(ctx, sistemaID)
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				sucessos++
			case errors.Is(err, versions.ErrPublicacaoConcorrente):
				// esperado para as perdedoras da corrida
			default:
				t.Errorf("erro inesperado ao publicar: %v", err)
			}
		}()
	}
	wg.Wait()

	if sucessos < 1 {
		t.Fatal("ao menos uma publicação deveria ter sucesso")
	}
	if ativas := contarAtivas(t, sistemaID); ativas != 1 {
		t.Fatalf("deveria haver exatamente 1 versão ativa; encontrei %d", ativas)
	}
}

// Critério 3 (parte 2): o rollback é uma simples troca de flag e completa em < 100ms.
func TestRollbackRapidoParaAnterior(t *testing.T) {
	mgr, ctx, sistemaID := setup(t)

	if _, err := mgr.Publicar(ctx, sistemaID); err != nil { // numero 1
		t.Fatalf("publicar v1: %v", err)
	}
	if _, err := mgr.Publicar(ctx, sistemaID); err != nil { // numero 2 (ativa)
		t.Fatalf("publicar v2: %v", err)
	}

	inicio := time.Now()
	ativa, err := mgr.Rollback(ctx, sistemaID, 0) // volta à imediatamente anterior
	elapsed := time.Since(inicio)
	if err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if ativa != 1 {
		t.Fatalf("versão ativa após rollback deveria ser 1; got %d", ativa)
	}
	if elapsed > 100*time.Millisecond {
		t.Fatalf("rollback deveria completar em < 100ms; levou %v", elapsed)
	}

	// E continua havendo exatamente uma versão ativa.
	if ativas := contarAtivas(t, sistemaID); ativas != 1 {
		t.Fatalf("deveria haver exatamente 1 versão ativa; encontrei %d", ativas)
	}

	v, err := mgr.ObterAtiva(ctx, sistemaID)
	if err != nil || v.Numero != 1 {
		t.Fatalf("versão ativa consolidada deveria ser 1; got numero=%d err=%v", v.Numero, err)
	}
}

// Sem versão anterior, o rollback é rejeitado (RN05).
func TestRollback_SemAnterior(t *testing.T) {
	mgr, ctx, sistemaID := setup(t)
	if _, err := mgr.Publicar(ctx, sistemaID); err != nil {
		t.Fatalf("publicar: %v", err)
	}
	if _, err := mgr.Rollback(ctx, sistemaID, 0); !errors.Is(err, versions.ErrSemVersaoAnterior) {
		t.Fatalf("esperava ErrSemVersaoAnterior; got %v", err)
	}
}
