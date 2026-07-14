package tenantctx

import (
	"context"
	"testing"

	"google.golang.org/grpc/metadata"

	commonv1 "github.com/machv4/platform/gen/go/construtor/common/v1"
)

func sample() *commonv1.TenantContext {
	return &commonv1.TenantContext{
		TenantId:   "11111111-1111-1111-1111-111111111111",
		UserId:     "user-42",
		Tipo:       "parceiro",
		TraceState: "congo=t61rcWkgMzE",
	}
}

func TestContextRoundTrip(t *testing.T) {
	ctx := NewContext(context.Background(), sample())

	if got := TenantID(ctx); got != sample().TenantId {
		t.Fatalf("TenantID=%q", got)
	}
	if got := UserID(ctx); got != "user-42" {
		t.Fatalf("UserID=%q", got)
	}
	if _, err := Require(ctx); err != nil {
		t.Fatalf("Require devolveu erro: %v", err)
	}
}

func TestRequire_Ausente(t *testing.T) {
	if _, err := Require(context.Background()); err != ErrNoTenantContext {
		t.Fatalf("esperava ErrNoTenantContext; got=%v", err)
	}
	// Contexto presente mas sem tenant_id também é insuficiente.
	ctx := NewContext(context.Background(), &commonv1.TenantContext{UserId: "x"})
	if _, err := Require(ctx); err != ErrNoTenantContext {
		t.Fatalf("esperava ErrNoTenantContext para tenant_id vazio; got=%v", err)
	}
}

// Simula a travessia de rede: inject (cliente, Metadata de saída) → transporte →
// extract (servidor, Metadata de entrada). O TenantContext deve sobreviver intacto.
func TestInjectExtract_TravessiaGRPC(t *testing.T) {
	clientCtx := inject(NewContext(context.Background(), sample()))

	md, ok := metadata.FromOutgoingContext(clientCtx)
	if !ok {
		t.Fatal("Metadata de saída não foi populado pelo inject")
	}
	// A identidade nunca deve trafegar em texto legível fora do Metadata binário.
	if vals := md.Get(MetadataKey); len(vals) == 0 {
		t.Fatalf("chave %q ausente no Metadata", MetadataKey)
	}

	serverCtx := extract(metadata.NewIncomingContext(context.Background(), md))

	got, err := Require(serverCtx)
	if err != nil {
		t.Fatalf("servidor não recuperou o TenantContext: %v", err)
	}
	if got.GetTenantId() != sample().TenantId || got.GetTipo() != "parceiro" || got.GetUserId() != "user-42" {
		t.Fatalf("TenantContext corrompido na travessia: %+v", got)
	}
}

func TestExtract_SemMetadata(t *testing.T) {
	// Sem Metadata de entrada, o extract não deve inventar contexto.
	if _, ok := FromContext(extract(context.Background())); ok {
		t.Fatal("extract criou contexto de tenant sem Metadata")
	}
}
