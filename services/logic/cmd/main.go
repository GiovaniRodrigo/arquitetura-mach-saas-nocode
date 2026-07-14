// Comando logic sobe o Logic Engine gRPC (RF02, RF07).
package main

import (
	"context"
	"log"
	"net"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"

	logicv1 "github.com/machv4/platform/gen/go/construtor/logic/v1"
	"github.com/machv4/platform/pkg/telemetry"
	"github.com/machv4/platform/pkg/tenantctx"
	"github.com/machv4/platform/services/logic/app"
)

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	ctx := context.Background()

	addr := env("LOGIC_GRPC_ADDR", ":50053")
	dsn := env("DATABASE_URL", "postgres://mach:mach@localhost:5432/machv4?sslmode=disable")
	otlp := env("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317")

	shutdown, err := telemetry.Init(ctx, telemetry.Config{ServiceName: "logic", OTLPEndpoint: otlp})
	if err != nil {
		log.Fatalf("telemetry: %v", err)
	}
	defer func() { _ = shutdown(ctx) }()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pool.Close()

	logic := app.NewServer(pool)

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(tenantctx.UnaryServerInterceptor()),
		grpc.ChainStreamInterceptor(tenantctx.StreamServerInterceptor()),
	)
	logicv1.RegisterLogicEngineServiceServer(grpcServer, logic)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen %s: %v", addr, err)
	}
	log.Printf("Logic Engine ouvindo em %s", addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
