package telemetry

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestAtributos_UsamNomesAnonimizados(t *testing.T) {
	if got := TenantID("abc").Key; string(got) != AttrTenantID {
		t.Fatalf("chave do tenant=%q", got)
	}
	if got := ComponentBlindIndex("8f3b").Key; string(got) != AttrComponentBlindIndex {
		t.Fatalf("chave do componente=%q", got)
	}
	// Garante que a convenção não regrediu para um nome real de campo.
	if AttrComponentBlindIndex != "platform.component.blind_index" {
		t.Fatal("atributo do componente deve referenciar blind_index, nunca o nome real")
	}
}

func TestInit_InstalaPropagadorW3CeShutdown(t *testing.T) {
	exp := tracetest.NewInMemoryExporter()
	shutdown, err := Init(context.Background(), Config{
		ServiceName: "iam",
		Exporter:    exp,
	})
	if err != nil {
		t.Fatalf("Init falhou: %v", err)
	}
	t.Cleanup(func() { _ = shutdown(context.Background()) })

	// O propagador global precisa expor o traceparent (W3C) para atravessar
	// HTTP → gRPC → AMQP (RNF04).
	fields := otel.GetTextMapPropagator().Fields()
	if !contains(fields, "traceparent") {
		t.Fatalf("propagador sem traceparent; fields=%v", fields)
	}
}

func TestInit_ExigeServiceName(t *testing.T) {
	if _, err := Init(context.Background(), Config{Exporter: tracetest.NewInMemoryExporter()}); err == nil {
		t.Fatal("Init deveria exigir ServiceName")
	}
}

func TestRedactor(t *testing.T) {
	r := NewRedactor("email", "cpf")

	got := r.Redact("falha ao gravar Email do usuario e CPF")
	if got != "falha ao gravar [REDACTED] do usuario e [REDACTED]" {
		t.Fatalf("redação incorreta: %q", got)
	}

	// Blind Index não contém nome real → não deve ser tocado.
	bi := "8f3b2a1c9d"
	if r.Redact(bi) != bi {
		t.Fatal("Redactor alterou um Blind Index")
	}
}

func TestRedactor_Attrs(t *testing.T) {
	r := NewRedactor("telefone")
	in := []attribute.KeyValue{
		attribute.String("erro", "telefone invalido"),
		ComponentBlindIndex("abc123"),
		attribute.Int("tentativas", 3),
	}
	out := r.RedactAttrs(in)

	if out[0].Value.AsString() != "[REDACTED] invalido" {
		t.Fatalf("valor string não redigido: %q", out[0].Value.AsString())
	}
	if out[1].Value.AsString() != "abc123" {
		t.Fatal("Blind Index redigido indevidamente")
	}
	if out[2].Value.AsInt64() != 3 {
		t.Fatal("atributo não-string alterado")
	}
}

func TestRedactor_SemTermos_NoOp(t *testing.T) {
	r := NewRedactor()
	if r.Redact("qualquer coisa") != "qualquer coisa" {
		t.Fatal("Redactor vazio deveria ser no-op")
	}
}

func contains(ss []string, target string) bool {
	for _, s := range ss {
		if s == target {
			return true
		}
	}
	return false
}
