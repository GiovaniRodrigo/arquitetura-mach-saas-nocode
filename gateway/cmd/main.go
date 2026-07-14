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

	deployv1 "github.com/machv4/platform/gen/go/construtor/deploy/v1"
	designv1 "github.com/machv4/platform/gen/go/construtor/design/v1"
	iamv1 "github.com/machv4/platform/gen/go/construtor/iam/v1"
	logicv1 "github.com/machv4/platform/gen/go/construtor/logic/v1"
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
	logicAddr := env("LOGIC_GRPC_ADDR", "localhost:50053")
	deployAddr := env("DEPLOY_GRPC_ADDR", "localhost:50054")
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

	logicConn, err := dial(logicAddr)
	if err != nil {
		log.Fatalf("logic client: %v", err)
	}
	defer logicConn.Close()

	deployConn, err := dial(deployAddr)
	if err != nil {
		log.Fatalf("deploy client: %v", err)
	}
	defer deployConn.Close()

	iam := iamv1.NewIAMServiceClient(iamConn)
	design := designv1.NewDesignEngineServiceClient(designConn)
	logic := logicv1.NewLogicEngineServiceClient(logicConn)
	deploy := deployv1.NewDeployEngineServiceClient(deployConn)
	rl := middleware.NewRateLimiter(50, 100) // 50 req/s, burst 100 por tenant

	handler := app.NewRouter(iam, design, logic, deploy, rl)

	log.Printf("Gateway ouvindo em %s (IAM %s, Design %s, Logic %s, Deploy %s)", httpAddr, iamAddr, designAddr, logicAddr, deployAddr)
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
