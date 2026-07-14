// Package telemetry padroniza a observabilidade distribuída (RNF04) e a proteção
// de dados sensíveis nos sinais (RNF08).
//
// Init configura um TracerProvider com exporter OTLP (→ Collector/Jaeger) e o
// propagador W3C Trace Context, de modo que um único trace atravesse HTTP, gRPC e
// AMQP. Os atributos de span usam SEMPRE identificadores anonimizados
// (platform.tenant_id e o Blind Index do componente) — nunca nomes reais de
// tabelas/campos. O Redactor complementa isso mascarando termos sensíveis que
// eventualmente cheguem a mensagens de log/erro.
package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Nomes canônicos dos atributos de span da plataforma (RNF08: apenas anonimizados).
const (
	AttrTenantID            = "platform.tenant_id"
	AttrComponentBlindIndex = "platform.component.blind_index"
)

// TenantID devolve o atributo de span com o tenant_id.
func TenantID(id string) attribute.KeyValue {
	return attribute.String(AttrTenantID, id)
}

// ComponentBlindIndex devolve o atributo de span com o Blind Index do componente.
// Use SEMPRE este helper em vez do nome real do campo (RNF08).
func ComponentBlindIndex(blindIndex string) attribute.KeyValue {
	return attribute.String(AttrComponentBlindIndex, blindIndex)
}

// Config parametriza o bootstrap de tracing.
type Config struct {
	// ServiceName identifica o serviço no Jaeger (ex.: "iam", "gateway").
	ServiceName string
	// OTLPEndpoint é o alvo gRPC do Collector (ex.: "localhost:4317").
	OTLPEndpoint string
	// Exporter, se fornecido, substitui o OTLP — útil em testes (in-memory).
	Exporter sdktrace.SpanExporter
}

// Init instala o TracerProvider e o propagador globais e devolve a função de
// shutdown, que faz flush dos spans pendentes.
func Init(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	if cfg.ServiceName == "" {
		return nil, fmt.Errorf("telemetry: ServiceName obrigatório")
	}

	exporter := cfg.Exporter
	if exporter == nil {
		var err error
		exporter, err = otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
			otlptracegrpc.WithInsecure(),
		)
		if err != nil {
			return nil, fmt.Errorf("telemetry: exporter OTLP: %w", err)
		}
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName(cfg.ServiceName)),
	)
	if err != nil {
		return nil, fmt.Errorf("telemetry: resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	// Trace Context (W3C) + Baggage: um traceparent único atravessa os protocolos.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}
