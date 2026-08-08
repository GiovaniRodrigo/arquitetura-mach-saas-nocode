//go:build integration

// Teste de integração da fachada REST de versoes/{id} (spec 004, RF12): listagem
// e reversão por id sobre a mesma lógica de Publicar/Rollback já coberta em
// rollback_test.go. Requer o Postgres do docker-compose com as migrações
// aplicadas. Executar:
//
//	DATABASE_URL=postgres://mach:mach@localhost:5432/machv4?sslmode=disable \
//	  go test -tags integration -p 1 ./services/deploy/tests/...
package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/machv4/platform/services/deploy/internal/versions"
)

// segundoSistemaMesmoTenant insere (como superusuário) um segundo sistema no
// mesmo tenant de sistemaID, para testar isolamento entre sistemas de um mesmo
// tenant (em contraste com isolamento entre tenants, já coberto por setup()).
func segundoSistemaMesmoTenant(t *testing.T, sistemaID string) string {
	t.Helper()
	bg := context.Background()
	admin, err := pgx.Connect(bg, dsn())
	if err != nil {
		t.Skipf("Postgres indisponível (%v)", err)
	}
	defer admin.Close(bg)

	var tenantID string
	if err := admin.QueryRow(bg, `SELECT tenant_id FROM sistemas WHERE id=$1`, sistemaID).Scan(&tenantID); err != nil {
		t.Fatalf("obter tenant_id: %v", err)
	}
	var outroSistemaID string
	if err := admin.QueryRow(bg,
		`INSERT INTO sistemas (tenant_id, nome) VALUES ($1,'sistema-deploy-2') RETURNING id`, tenantID,
	).Scan(&outroSistemaID); err != nil {
		t.Fatalf("inserir segundo sistema: %v", err)
	}
	return outroSistemaID
}

// RF12: ListarVersoes devolve todas as versões do sistema, mais recente
// primeiro, com a flag ativa correta.
func TestListarVersoes_OrdemDecrescente(t *testing.T) {
	mgr, ctx, sistemaID := setup(t)

	if _, err := mgr.Publicar(ctx, sistemaID); err != nil { // numero 1
		t.Fatalf("publicar v1: %v", err)
	}
	if _, err := mgr.Publicar(ctx, sistemaID); err != nil { // numero 2
		t.Fatalf("publicar v2: %v", err)
	}
	if _, err := mgr.Publicar(ctx, sistemaID); err != nil { // numero 3 (ativa)
		t.Fatalf("publicar v3: %v", err)
	}

	versoes, err := mgr.ListarVersoes(ctx, sistemaID)
	if err != nil {
		t.Fatalf("listar versões: %v", err)
	}
	if len(versoes) != 3 {
		t.Fatalf("esperava 3 versões; got %d", len(versoes))
	}
	// Mais recente primeiro.
	if versoes[0].Numero != 3 || versoes[1].Numero != 2 || versoes[2].Numero != 1 {
		t.Fatalf("ordem inesperada: %+v", versoes)
	}
	if !versoes[0].Ativa {
		t.Fatalf("a versão mais recente deveria estar ativa: %+v", versoes[0])
	}
	if versoes[1].Ativa || versoes[2].Ativa {
		t.Fatalf("apenas a versão 3 deveria estar ativa: %+v", versoes)
	}
	for _, v := range versoes {
		if v.ID == "" {
			t.Fatalf("versão sem id: %+v", v)
		}
		if v.CriadoEm.IsZero() {
			t.Fatalf("versão sem criado_em: %+v", v)
		}
	}
}

// RLS: ListarVersoes não pode enxergar (nem confirmar a existência de) um
// sistema de outro tenant.
func TestListarVersoes_IsolaPorTenant(t *testing.T) {
	_, ctxA, sistemaA := setup(t)
	mgrB, ctxB, _ := setup(t)

	if _, err := mgrB.ListarVersoes(ctxB, sistemaA); !errors.Is(err, versions.ErrSistemaInexistente) {
		t.Fatalf("esperava ErrSistemaInexistente ao listar sistema de outro tenant; got %v", err)
	}
	_ = ctxA
}

// RF12: RollbackPorID resolve o id de uma versão para o número correspondente e
// reativa-a, com as mesmas garantias de RN05 do Rollback por número.
func TestRollbackPorID_ReverteParaVersaoCorreta(t *testing.T) {
	mgr, ctx, sistemaID := setup(t)

	if _, err := mgr.Publicar(ctx, sistemaID); err != nil { // numero 1
		t.Fatalf("publicar v1: %v", err)
	}
	if _, err := mgr.Publicar(ctx, sistemaID); err != nil { // numero 2 (ativa)
		t.Fatalf("publicar v2: %v", err)
	}

	versoes, err := mgr.ListarVersoes(ctx, sistemaID)
	if err != nil {
		t.Fatalf("listar versões: %v", err)
	}
	var idV1 string
	for _, v := range versoes {
		if v.Numero == 1 {
			idV1 = v.ID
		}
	}
	if idV1 == "" {
		t.Fatalf("não encontrei o id da versão 1: %+v", versoes)
	}

	ativa, err := mgr.RollbackPorID(ctx, sistemaID, idV1)
	if err != nil {
		t.Fatalf("rollback por id: %v", err)
	}
	if ativa != 1 {
		t.Fatalf("versão ativa após reverter deveria ser 1; got %d", ativa)
	}

	v, err := mgr.ObterAtiva(ctx, sistemaID)
	if err != nil || v.Numero != 1 {
		t.Fatalf("versão ativa consolidada deveria ser 1; got numero=%d err=%v", v.Numero, err)
	}
	if ativas := contarAtivas(t, sistemaID); ativas != 1 {
		t.Fatalf("deveria haver exatamente 1 versão ativa; encontrei %d", ativas)
	}
}

// Defesa equivalente a exigirSistema: um id de versão válido, mas de OUTRO
// sistema do MESMO tenant, não pode reverter o sistema-alvo (a query de
// RollbackPorID filtra por id E sistema_id).
func TestRollbackPorID_VersaoDeOutroSistema_Rejeitada(t *testing.T) {
	mgr, ctx, sistemaA := setup(t)
	sistemaB := segundoSistemaMesmoTenant(t, sistemaA)

	if _, err := mgr.Publicar(ctx, sistemaA); err != nil {
		t.Fatalf("publicar em sistemaA: %v", err)
	}
	versoesA, err := mgr.ListarVersoes(ctx, sistemaA)
	if err != nil || len(versoesA) == 0 {
		t.Fatalf("listar versões de sistemaA: %v (%+v)", err, versoesA)
	}
	idVersaoA := versoesA[0].ID

	if _, err := mgr.RollbackPorID(ctx, sistemaB, idVersaoA); !errors.Is(err, versions.ErrSemVersaoAnterior) {
		t.Fatalf("esperava ErrSemVersaoAnterior ao reverter sistemaB com versão de sistemaA; got %v", err)
	}
}

// Defesa equivalente a exigirSistema/RLS: um id de versão de um sistema de OUTRO
// tenant não pode reverter, mesmo citando o sistema_id correto dentro do próprio
// tenant.
func TestRollbackPorID_VersaoDeOutroTenant_Rejeitada(t *testing.T) {
	mgr, ctx, sistemaA := setup(t)
	mgrB, ctxB, sistemaB := setup(t)

	if _, err := mgr.Publicar(ctx, sistemaA); err != nil {
		t.Fatalf("publicar em sistemaA (tenant A): %v", err)
	}
	versoesA, err := mgr.ListarVersoes(ctx, sistemaA)
	if err != nil || len(versoesA) == 0 {
		t.Fatalf("listar versões de sistemaA: %v (%+v)", err, versoesA)
	}
	idVersaoA := versoesA[0].ID

	if _, err := mgrB.Publicar(ctxB, sistemaB); err != nil {
		t.Fatalf("publicar em sistemaB (tenant B): %v", err)
	}

	if _, err := mgrB.RollbackPorID(ctxB, sistemaB, idVersaoA); !errors.Is(err, versions.ErrSemVersaoAnterior) {
		t.Fatalf("esperava ErrSemVersaoAnterior ao reverter com id de versão de outro tenant; got %v", err)
	}
}
