// Comando gateway sobe o API Gateway HTTP (:8080), traduzindo REST para gRPC.
package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	iamv1 "github.com/machv4/platform/gen/go/construtor/iam/v1"
	"github.com/machv4/platform/gateway/internal/app"
	"github.com/machv4/platform/gateway/internal/middleware"
	"github.com/machv4/platform/pkg/telemetry"
	"github.com/machv4/platform/pkg/tenantctx"
)

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	ctx := context.Background()

	httpAddr := env("GATEWAY_HTTP_ADDR", ":8080")
	iamAddr := env("IAM_GRPC_ADDR", "localhost:50051")
	otlp := env("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317")

	shutdown, err := telemetry.Init(ctx, telemetry.Config{ServiceName: "gateway", OTLPEndpoint: otlp})
	if err != nil {
		log.Fatalf("telemetry: %v", err)
	}
	defer func() { _ = shutdown(ctx) }()

	// Cliente gRPC do IAM: propaga TenantContext (Metadata) e o trace.
	conn, err := grpc.NewClient(iamAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithChainUnaryInterceptor(tenantctx.UnaryClientInterceptor()),
		grpc.WithChainStreamInterceptor(tenantctx.StreamClientInterceptor()),
	)
	if err != nil {
		log.Fatalf("iam client: %v", err)
	}
	defer conn.Close()

	iam := iamv1.NewIAMServiceClient(conn)
	rl := middleware.NewRateLimiter(50, 100) // 50 req/s, burst 100 por tenant

	handler := app.NewRouter(iam, rl)

	log.Printf("Gateway ouvindo em %s (IAM em %s)", httpAddr, iamAddr)
	if err := http.ListenAndServe(httpAddr, handler); err != nil {
		log.Fatalf("http: %v", err)
	}
}
