//go:build integration

// Teste ponta-a-ponta do Export Engine (RF05, RNF01, critério 7): cria um Job,
// aguarda concluir e baixa o Presigned URL, verificando que o NDJSON traz as três
// fontes (design, regras, operacional) do sistema. Requer Postgres + MinIO:
//
//	DATABASE_URL=postgres://mach:mach@localhost:5432/machv4?sslmode=disable \
//	S3_ENDPOINT=localhost:9010 \
//	  go test -tags integration ./services/export/tests/...
package tests

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	commonv1 "github.com/machv4/platform/gen/go/construtor/common/v1"
	exportv1 "github.com/machv4/platform/gen/go/construtor/export/v1"
	"github.com/machv4/platform/pkg/tenantctx"
	exportapp "github.com/machv4/platform/services/export/app"
	"github.com/machv4/platform/services/export/internal/jobs"
	"github.com/machv4/platform/services/export/internal/storage"
)

const appRole = "machapp_export_rls_test"

func dsn() string {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v
	}
	return "postgres://mach:mach@localhost:5432/machv4?sslmode=disable"
}

func endpoint() string {
	if v := os.Getenv("S3_ENDPOINT"); v != "" {
		return v
	}
	return "localhost:9010"
}

func TestExportacaoCompleta(t *testing.T) {
	ctx := context.Background()

	store, err := storage.New(ctx, storage.Config{
		Endpoint: endpoint(), AccessKey: "mach", SecretKey: "machsecret",
		Bucket: "exports-e2e", PresignTTL: 5 * time.Minute,
	})
	if err != nil {
		t.Skipf("MinIO indisponível (%v)", err)
	}

	admin, err := pgx.Connect(ctx, dsn())
	if err != nil {
		t.Skipf("Postgres indisponível (%v)", err)
	}
	defer admin.Close(ctx)

	for _, stmt := range []string{
		`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname='` + appRole + `') THEN CREATE ROLE ` + appRole + ` NOLOGIN NOSUPERUSER; END IF; END $$;`,
		`GRANT USAGE ON SCHEMA public TO ` + appRole + `;`,
		`GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO ` + appRole + `;`,
	} {
		if _, err := admin.Exec(ctx, stmt); err != nil {
			t.Fatalf("setup role: %v", err)
		}
	}

	var tenantID, sistemaID string
	if err := admin.QueryRow(ctx, `INSERT INTO tenants (nome, tipo, chave_blind_index) VALUES ('itg-export','dono','\x6b') RETURNING id`).Scan(&tenantID); err != nil {
		t.Fatalf("tenant: %v", err)
	}
	t.Cleanup(func() {
		c, _ := pgx.Connect(context.Background(), dsn())
		if c != nil {
			defer c.Close(context.Background())
			_, _ = c.Exec(context.Background(), `DELETE FROM tenants WHERE id=$1`, tenantID)
		}
	})
	if err := admin.QueryRow(ctx, `INSERT INTO sistemas (tenant_id, nome) VALUES ($1,'sistema-export') RETURNING id`, tenantID).Scan(&sistemaID); err != nil {
		t.Fatalf("sistema: %v", err)
	}

	// Semeia as três fontes.
	if _, err := admin.Exec(ctx, `INSERT INTO designs (tenant_id, sistema_id, nome, arvore) VALUES ($1,$2,'Tela',$3)`,
		tenantID, sistemaID, `{"blind_index":"bi-root","tipo":"container"}`); err != nil {
		t.Fatalf("design: %v", err)
	}
	if _, err := admin.Exec(ctx, `INSERT INTO regras_negocio (tenant_id, sistema_id, arvore_decisao) VALUES ($1,$2,$3)`,
		tenantID, sistemaID, `{"no":"acao","tipo":"webhook.disparo"}`); err != nil {
		t.Fatalf("regra: %v", err)
	}
	if _, err := admin.Exec(ctx, `INSERT INTO dados_operacionais (tenant_id, sistema_id, valores) VALUES ($1,$2,$3)`,
		tenantID, sistemaID, `{"bi-nome":"Ana"}`); err != nil {
		t.Fatalf("dados: %v", err)
	}

	// Pool com a role comum (RLS efetiva).
	cfg, _ := pgxpool.ParseConfig(dsn())
	cfg.AfterConnect = func(ctx context.Context, c *pgx.Conn) error {
		_, err := c.Exec(ctx, "SET ROLE "+appRole)
		return err
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)

	srv := exportapp.NewServer(pool, store)
	tctx := tenantctx.NewContext(ctx, &commonv1.TenantContext{TenantId: tenantID, Tipo: "dono"})

	criado, err := srv.CriarJob(tctx, &exportv1.CriarJobRequest{SistemaId: sistemaID})
	if err != nil {
		t.Fatalf("criar job: %v", err)
	}

	// Aguarda o pipeline concluir (pronto) com o Presigned URL.
	var job *exportv1.Job
	prazo := time.Now().Add(10 * time.Second)
	for time.Now().Before(prazo) {
		job, err = srv.ObterJob(tctx, &exportv1.ObterJobRequest{Id: criado.GetId()})
		if err != nil {
			t.Fatalf("obter job: %v", err)
		}
		if job.GetStatus() == jobs.StatusPronto || job.GetStatus() == jobs.StatusErro {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if job.GetStatus() != jobs.StatusPronto {
		t.Fatalf("job deveria estar pronto; status=%s", job.GetStatus())
	}
	if job.GetArquivoUrl() == "" || job.GetExpiraEm() == "" {
		t.Fatalf("pronto deveria ter URL e expiração: %+v", job)
	}

	// Baixa o artefato e confirma as três fontes.
	resp, err := http.Get(job.GetArquivoUrl()) //nolint:gosec // URL assinado de teste
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	defer resp.Body.Close()

	origens := map[string]bool{}
	sc := bufio.NewScanner(resp.Body)
	for sc.Scan() {
		linha := strings.TrimSpace(sc.Text())
		if linha == "" {
			continue
		}
		var rec struct {
			Origem string `json:"origem"`
		}
		if err := json.Unmarshal([]byte(linha), &rec); err != nil {
			t.Fatalf("linha NDJSON inválida %q: %v", linha, err)
		}
		origens[rec.Origem] = true
	}
	for _, o := range []string{"design", "regras", "operacional"} {
		if !origens[o] {
			t.Fatalf("exportação deveria conter a origem %q; origens=%v", o, origens)
		}
	}
}
