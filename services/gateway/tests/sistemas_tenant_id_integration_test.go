//go:build integration

// Teste ponta-a-ponta de GET /api/v1/sistemas?tenant_id={id} (spec 004, RF08):
// o endpoint de maior risco de vazamento cross-tenant da spec. Cobre os três
// cenários exigidos: tenant_id ausente (comportamento atual preservado),
// tenant_id de um filho direto (lista os sistemas do filho) e tenant_id fora
// da hierarquia — um tenant alheio ou um "neto" — que deve dar 404 e nunca
// vazar os sistemas do tenant alheio na resposta. Requer o Postgres do
// docker-compose com as migrações aplicadas. Executar:
//
//	DATABASE_URL=postgres://mach:mach@localhost:5432/machv4?sslmode=disable \
//	  go test -tags integration -p 1 ./services/gateway/tests/...
package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5"
)

// tenantIDHarness estende sistemasHarness com uma hierarquia de tenants: A
// (raiz, dono) → filho (cliente, parent_id=A) → neto (cliente,
// parent_id=filho), mais um tenant totalmente alheio (B, raiz). Semeia um
// sistema em cada um dos três (filho/neto/alheio) para detectar vazamento.
func tenantIDHarness(t *testing.T) (handler http.Handler, tokenA string, filhoID, netoID, alheioID, sisFilhoID, sisNetoID, sisAlheioID string) {
	t.Helper()
	ctx := context.Background()

	h, iss, tenantA, tenantB := sistemasHarness(t)
	tokenA = tokenDe(t, iss, tenantA, "dono")

	admin, err := pgx.Connect(ctx, dsn())
	if err != nil {
		t.Skipf("Postgres indisponível (%v)", err)
	}
	t.Cleanup(func() { _ = admin.Close(context.Background()) })

	if err := admin.QueryRow(ctx,
		`INSERT INTO tenants (parent_id, nome, tipo, chave_blind_index) VALUES ($1,'itg-sis-filho','cliente','\x6b44') RETURNING id`,
		tenantA,
	).Scan(&filhoID); err != nil {
		t.Fatalf("tenant filho: %v", err)
	}
	if err := admin.QueryRow(ctx,
		`INSERT INTO tenants (parent_id, nome, tipo, chave_blind_index) VALUES ($1,'itg-sis-neto','cliente','\x6b45') RETURNING id`,
		filhoID,
	).Scan(&netoID); err != nil {
		t.Fatalf("tenant neto: %v", err)
	}
	alheioID = tenantB // já semeado por sistemasHarness, sem relação de parentesco com A

	if err := admin.QueryRow(ctx,
		`INSERT INTO sistemas (tenant_id, nome) VALUES ($1,'Sistema do Filho') RETURNING id`, filhoID,
	).Scan(&sisFilhoID); err != nil {
		t.Fatalf("sistema filho: %v", err)
	}
	if err := admin.QueryRow(ctx,
		`INSERT INTO sistemas (tenant_id, nome) VALUES ($1,'Sistema do Neto') RETURNING id`, netoID,
	).Scan(&sisNetoID); err != nil {
		t.Fatalf("sistema neto: %v", err)
	}
	if err := admin.QueryRow(ctx,
		`INSERT INTO sistemas (tenant_id, nome) VALUES ($1,'Sistema Alheio') RETURNING id`, alheioID,
	).Scan(&sisAlheioID); err != nil {
		t.Fatalf("sistema alheio: %v", err)
	}

	t.Cleanup(func() {
		c, err := pgx.Connect(context.Background(), dsn())
		if err != nil {
			return
		}
		defer c.Close(context.Background())
		_, _ = c.Exec(context.Background(), `DELETE FROM tenants WHERE id = ANY($1)`, []string{filhoID, netoID})
	})

	return h, tokenA, filhoID, netoID, alheioID, sisFilhoID, sisNetoID, sisAlheioID
}

func listarSistemasIDs(t *testing.T, rec *httptest.ResponseRecorder) []string {
	t.Helper()
	var lista struct {
		Sistemas []struct{ ID string } `json:"sistemas"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &lista); err != nil {
		t.Fatalf("json lista: %v", err)
	}
	ids := make([]string, 0, len(lista.Sistemas))
	for _, s := range lista.Sistemas {
		ids = append(ids, s.ID)
	}
	return ids
}

func contains(ids []string, id string) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}

// (a) Sem tenant_id na query, o comportamento atual é preservado: lista os
// sistemas do tenant do JWT (A), e os sistemas dos outros tenants da
// hierarquia (filho/neto/alheio) nunca aparecem.
func TestListarSistemas_SemTenantID_ComportamentoAtualPreservado(t *testing.T) {
	handler, tokenA, _, _, _, sisFilho, sisNeto, sisAlheio := tenantIDHarness(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/sistemas", nil)
	req.Header.Set("Authorization", "Bearer "+tokenA)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200; got=%d body=%s", rec.Code, rec.Body.String())
	}
	ids := listarSistemasIDs(t, rec)
	for _, alheio := range []string{sisFilho, sisNeto, sisAlheio} {
		if contains(ids, alheio) {
			t.Fatalf("sem tenant_id, A não deveria ver sistema de outro tenant (%s): %+v", alheio, ids)
		}
	}
}

// (b) tenant_id de um filho direto do tenant do JWT → 200, lista os sistemas
// desse filho.
func TestListarSistemas_ComTenantID_FilhoDireto_ListaSistemasDoFilho(t *testing.T) {
	handler, tokenA, filhoID, _, _, sisFilho, _, _ := tenantIDHarness(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/sistemas?tenant_id="+filhoID, nil)
	req.Header.Set("Authorization", "Bearer "+tokenA)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200; got=%d body=%s", rec.Code, rec.Body.String())
	}
	ids := listarSistemasIDs(t, rec)
	if !contains(ids, sisFilho) {
		t.Fatalf("sistema do filho não apareceu na listagem: %+v", ids)
	}
}

// (c) O TESTE MAIS IMPORTANTE DO PR: tenant_id que NÃO é filho direto do
// tenant do JWT — seja um "neto" (dois níveis abaixo) ou um tenant totalmente
// alheio — deve dar 404, e os sistemas desse tenant nunca podem vazar na
// resposta.
func TestListarSistemas_ComTenantID_ForaDaHierarquia_404_SemVazamento(t *testing.T) {
	handler, tokenA, _, netoID, alheioID, _, sisNeto, sisAlheio := tenantIDHarness(t)

	casos := []struct {
		nome        string
		tenantID    string
		sistemaFuga string
	}{
		{"neto (dois níveis abaixo, não filho direto)", netoID, sisNeto},
		{"tenant totalmente alheio", alheioID, sisAlheio},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/sistemas?tenant_id="+c.tenantID, nil)
			req.Header.Set("Authorization", "Bearer "+tokenA)
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusNotFound {
				t.Fatalf("%s: esperava 404; got=%d body=%s", c.nome, rec.Code, rec.Body.String())
			}
			// Corpo de erro nunca deve conter o id do sistema alheio (RNF08:
			// erros não expõem dados de outro tenant).
			if rec.Body.Len() > 0 {
				body := rec.Body.String()
				if len(body) > 0 && contains([]string{body}, c.sistemaFuga) {
					t.Fatalf("%s: resposta de erro vazou o id do sistema alheio: %s", c.nome, body)
				}
			}
		})
	}
}
