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

	designv1 "github.com/machv4/platform/gen/go/construtor/design/v1"
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
	designAddr := env("DESIGN_GRPC_ADDR", "localhost:50052")
	otlp := env("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317")

	shutdown, err := telemetry.Init(ctx, telemetry.Config{ServiceName: "gateway", OTLPEndpoint: otlp})
	if err != nil {
		log.Fatalf("telemetry: %v", err)
	}
	defer func() { _ = shutdown(ctx) }()

	// Clientes gRPC internos: propagam TenantContext (Metadata) e o trace.
	iamConn, err := dial(iamAddr)
	if err != nil {
		log.Fatalf("iam client: %v", err)
	}
	defer iamConn.Close()

	designConn, err := dial(designAddr)
	if err != nil {
		log.Fatalf("design client: %v", err)
	}
	defer designConn.Close()

	iam := iamv1.NewIAMServiceClient(iamConn)
	design := designv1.NewDesignEngineServiceClient(designConn)
	rl := middleware.NewRateLimiter(50, 100) // 50 req/s, burst 100 por tenant

	handler := app.NewRouter(iam, design, rl)

	log.Printf("Gateway ouvindo em %s (IAM em %s, Design em %s)", httpAddr, iamAddr, designAddr)
	if err := http.ListenAndServe(httpAddr, handler); err != nil {
		log.Fatalf("http: %v", err)
	}
}

// dial abre um cliente gRPC interno com propagação de TenantContext e trace.
func dial(addr string) (*grpc.ClientConn, error) {
	return grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithChainUnaryInterceptor(tenantctx.UnaryClientInterceptor()),
		grpc.WithChainStreamInterceptor(tenantctx.StreamClientInterceptor()),
	)
}
