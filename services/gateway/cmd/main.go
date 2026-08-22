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
	exportv1 "github.com/machv4/platform/gen/go/construtor/export/v1"
	iamv1 "github.com/machv4/platform/gen/go/construtor/iam/v1"
	logicv1 "github.com/machv4/platform/gen/go/construtor/logic/v1"
	"github.com/machv4/platform/pkg/telemetry"
	"github.com/machv4/platform/pkg/tenantctx"
	"github.com/machv4/platform/services/gateway/internal/app"
	"github.com/machv4/platform/services/gateway/internal/meshmetrics"
	"github.com/machv4/platform/services/gateway/internal/middleware"
	"github.com/machv4/platform/services/gateway/internal/routes"
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
	exportAddr := env("EXPORT_GRPC_ADDR", "localhost:50055")
	// spec 009: recursos (CPU/memória/RPS/latência/sucesso) vêm direto do
	// Kubernetes (metrics-server) e do Prometheus do linkerd-viz — não há
	// mais um serviço "monitor" gRPC próprio nem MONITOR_GRPC_ADDR.
	k8sNamespace := env("K8S_NAMESPACE", "machv4")
	prometheusURL := env("LINKERD_PROMETHEUS_URL", "http://prometheus.linkerd-viz.svc.cluster.local:9090")
	otlp := env("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317")
	appURL := env("APP_URL", "http://localhost:8080")

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

	exportConn, err := dial(exportAddr)
	if err != nil {
		log.Fatalf("export client: %v", err)
	}
	defer exportConn.Close()

	iam := iamv1.NewIAMServiceClient(iamConn)
	design := designv1.NewDesignEngineServiceClient(designConn)
	logic := logicv1.NewLogicEngineServiceClient(logicConn)
	deploy := deployv1.NewDeployEngineServiceClient(deployConn)
	export := exportv1.NewExportEngineServiceClient(exportConn)
	rl := middleware.NewRateLimiter(50, 100) // 50 req/s, burst 100 por tenant

	// K8sClient exige rodar dentro do cluster (lê o ServiceAccount montado no
	// pod) — sem isso, o Gateway ainda sobe, mas /api/v1/monitor/recursos
	// responde 502 (RNF02: erro de tela única, não derruba o resto do Gateway).
	var recursos *meshmetrics.Client
	if k8s, err := meshmetrics.NewK8sClient(k8sNamespace); err != nil {
		log.Printf("AVISO: meshmetrics desabilitado (%v) — /api/v1/monitor/recursos responderá 502", err)
	} else {
		recursos = meshmetrics.NewClient(k8s, meshmetrics.NewPrometheusClient(prometheusURL, k8sNamespace))
	}

	// Login social (Google/GitHub) — credenciais e allowlist por ambiente. Sem
	// GOOGLE_/GITHUB_CLIENT_* configurados, oauth é nil e as rotas ficam desligadas.
	oauth := routes.NewOAuthHandler(
		iam, appURL,
		os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET"),
		os.Getenv("GITHUB_CLIENT_ID"), os.Getenv("GITHUB_CLIENT_SECRET"),
		os.Getenv("OAUTH_ALLOWED_REDIRECT_URIS"),
	)
	if oauth != nil {
		log.Println("login social habilitado (/auth/{provedor})")
	}

	handler := app.NewRouter(iam, design, logic, deploy, export, recursos, rl, oauth)

	log.Printf("Gateway ouvindo em %s (IAM %s, Design %s, Logic %s, Deploy %s, Export %s, recursos via k8s/prometheus ns=%s)", httpAddr, iamAddr, designAddr, logicAddr, deployAddr, exportAddr, k8sNamespace)
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
