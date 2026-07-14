// Comando export sobe o Export Engine gRPC (RF05).
package main

import (
	"context"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"

	exportv1 "github.com/machv4/platform/gen/go/construtor/export/v1"
	"github.com/machv4/platform/pkg/telemetry"
	"github.com/machv4/platform/pkg/tenantctx"
	"github.com/machv4/platform/services/export/app"
	"github.com/machv4/platform/services/export/internal/storage"
)

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	ctx := context.Background()

	addr := env("EXPORT_GRPC_ADDR", ":50055")
	dsn := env("DATABASE_URL", "postgres://mach:mach@localhost:5432/machv4?sslmode=disable")
	otlp := env("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317")

	shutdown, err := telemetry.Init(ctx, telemetry.Config{ServiceName: "export", OTLPEndpoint: otlp})
	if err != nil {
		log.Fatalf("telemetry: %v", err)
	}
	defer func() { _ = shutdown(ctx) }()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pool.Close()

	ttlMin, _ := strconv.Atoi(env("EXPORT_PRESIGN_TTL_MIN", "10"))
	store, err := storage.New(ctx, storage.Config{
		Endpoint:   env("S3_ENDPOINT", "localhost:9000"),
		AccessKey:  env("S3_ACCESS_KEY", "mach"),
		SecretKey:  env("S3_SECRET_KEY", "machsecret"),
		Bucket:     env("S3_BUCKET", "exports"),
		UseSSL:     env("S3_USE_SSL", "false") == "true",
		PresignTTL: time.Duration(ttlMin) * time.Minute,
	})
	if err != nil {
		log.Fatalf("storage: %v", err)
	}

	exportSrv := app.NewServer(pool, store)

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(tenantctx.UnaryServerInterceptor()),
		grpc.ChainStreamInterceptor(tenantctx.StreamServerInterceptor()),
	)
	exportv1.RegisterExportEngineServiceServer(grpcServer, exportSrv)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen %s: %v", addr, err)
	}
	log.Printf("Export Engine ouvindo em %s", addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
