package database

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	commonv1 "github.com/machv4/platform/gen/go/construtor/common/v1"
	"github.com/machv4/platform/pkg/tenantctx"
)

// spyBeginner falha se Begin for chamado — usado para provar que o guard de
// tenant curto-circuita ANTES de qualquer contato com o banco.
type spyBeginner struct{ called bool }

func (s *spyBeginner) Begin(context.Context) (pgx.Tx, error) {
	s.called = true
	return nil, errors.New("Begin não deveria ter sido chamado")
}

func TestWithTenant_SemTenant_Rejeita(t *testing.T) {
	spy := &spyBeginner{}
	db := New(spy)

	err := db.WithTenant(context.Background(), func(context.Context, pgx.Tx) error {
		t.Fatal("fn não deveria executar sem tenant")
		return nil
	})

	if !errors.Is(err, ErrNoTenant) {
		t.Fatalf("esperava ErrNoTenant; got=%v", err)
	}
	if spy.called {
		t.Fatal("abriu transação sem contexto de tenant — vazamento de RN01")
	}
}

func TestWithTenant_TenantVazio_Rejeita(t *testing.T) {
	spy := &spyBeginner{}
	// Contexto com TenantContext presente mas tenant_id vazio ainda é inválido.
	ctx := tenantctx.NewContext(context.Background(), &commonv1.TenantContext{UserId: "x"})

	err := New(spy).WithTenant(ctx, func(context.Context, pgx.Tx) error { return nil })
	if !errors.Is(err, ErrNoTenant) {
		t.Fatalf("esperava ErrNoTenant; got=%v", err)
	}
	if spy.called {
		t.Fatal("abriu transação com tenant_id vazio")
	}
}
