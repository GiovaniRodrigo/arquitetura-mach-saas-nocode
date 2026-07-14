//go:build e2e

// Teste E2E de rastreabilidade (RNF04, critério 6): um único traceparent atravessa
// Gateway → gRPC → AMQP → Worker e forma UM ÚNICO trace no Jaeger.
//
// Encena a cadeia com os transportes reais (gRPC sobre bufconn instrumentado com
// otelgrpc; AMQP contra um RabbitMQ real) exportando via OTLP a um Jaeger real, e
// então consulta a API do Jaeger confirmando que todos os spans partilham o mesmo
// trace. Requer Jaeger e RabbitMQ. Executar:
//
//	OTLP_ENDPOINT=localhost:14317 JAEGER_QUERY=localhost:16687 \
//	RABBITMQ_URL=amqp://mach:mach@localhost:5674/ \
//	  go test -tags e2e ./tests/e2e/...
package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	commonv1 "github.com/machv4/platform/gen/go/construtor/common/v1"
	iamv1 "github.com/machv4/platform/gen/go/construtor/iam/v1"
	"github.com/machv4/platform/pkg/eventbus"
	"github.com/machv4/platform/pkg/telemetry"
	"github.com/machv4/platform/pkg/tenantctx"
)

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// publisherServer reencena o "gRPC" da cadeia: ao ser chamado, publica um evento
// no RabbitMQ — dentro do span do servidor, o traceparent é injetado nos headers
// (mesma lógica do services/logic/internal/events.Publisher, inlined por o pacote
// interno não ser importável a partir de tests/e2e).
type publisherServer struct {
	iamv1.UnimplementedIAMServiceServer
	ch *amqp.Channel
}

func (s *publisherServer) ValidarToken(ctx context.Context, _ *iamv1.ValidarTokenRequest) (*iamv1.ValidarTokenResponse, error) {
	tenant := tenantctx.TenantID(ctx)
	body, _ := json.Marshal(eventbus.Corpo{
		Tipo:    eventbus.TipoWebhook,
		Payload: json.RawMessage(`{"url_destino":"https://exemplo"}`),
	})
	headers := amqp.Table{eventbus.HeaderTenantID: tenant}
	otel.GetTextMapPropagator().Inject(ctx, eventbus.HeaderCarrier(headers))

	err := s.ch.PublishWithContext(ctx, eventbus.Exchange, eventbus.RoutingKey(eventbus.TipoWebhook, tenant), false, false, amqp.Publishing{
		ContentType: "application/json",
		Headers:     headers,
		Body:        body,
	})
	if err != nil {
		return nil, err
	}
	return &iamv1.ValidarTokenResponse{Valido: true}, nil
}

func TestTraceUnicoAtravessaCadeia(t *testing.T) {
	otlp := envOr("OTLP_ENDPOINT", "localhost:14317")
	jaegerQuery := envOr("JAEGER_QUERY", "localhost:16687")
	amqpURL := envOr("RABBITMQ_URL", "amqp://mach:mach@localhost:5674/")

	ctx := context.Background()

	// RabbitMQ real.
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		t.Skipf("RabbitMQ indisponível (%v)", err)
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		t.Fatalf("canal: %v", err)
	}
	defer ch.Close()
	_, _ = ch.QueuePurge(eventbus.FilaWebhooks, false)

	// Telemetria global → Jaeger (via OTLP). Um provider basta: o critério é UM trace.
	shutdown, err := telemetry.Init(ctx, telemetry.Config{ServiceName: "e2e", OTLPEndpoint: otlp})
	if err != nil {
		t.Skipf("OTLP/Jaeger indisponível (%v)", err)
	}

	// --- Gateway: span raiz -------------------------------------------------
	tenant := "t-e2e"
	ctxGw, root := otel.Tracer("gateway").Start(
		tenantctx.NewContext(ctx, &commonv1.TenantContext{TenantId: tenant, Tipo: "dono"}),
		"gateway.request",
	)
	traceID := root.SpanContext().TraceID().String()

	// --- gRPC real (bufconn) instrumentado com otelgrpc --------------------
	lis := bufconn.Listen(1 << 20)
	srv := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(tenantctx.UnaryServerInterceptor()),
	)
	iamv1.RegisterIAMServiceServer(srv, &publisherServer{ch: ch})
	go func() { _ = srv.Serve(lis) }()
	defer srv.Stop()

	cc, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithChainUnaryInterceptor(tenantctx.UnaryClientInterceptor()),
	)
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	defer cc.Close()

	// Gateway → gRPC (→ AMQP publish dentro do handler).
	if _, err := iamv1.NewIAMServiceClient(cc).ValidarToken(ctxGw, &iamv1.ValidarTokenRequest{Jwt: "x"}); err != nil {
		t.Fatalf("chamada gRPC: %v", err)
	}

	// --- Worker: consome, extrai o traceparent e cria o span filho ---------
	msg := getComRetry(t, ch, eventbus.FilaWebhooks)
	ctxMsg := otel.GetTextMapPropagator().Extract(ctx, eventbus.HeaderCarrier(msg.Headers))
	_, workerSpan := otel.Tracer("worker").Start(ctxMsg, "worker.process")
	workerSpan.End()

	root.End()

	// Flush de todos os spans ao Jaeger.
	if err := shutdown(ctx); err != nil {
		t.Fatalf("shutdown/flush: %v", err)
	}

	// --- Verificação no Jaeger --------------------------------------------
	ops := esperarTrace(t, jaegerQuery, traceID)

	// O worker deve estar no MESMO trace que o gateway (critério 6).
	precisa := []string{"gateway.request", "worker.process"}
	for _, op := range precisa {
		if !ops[op] {
			t.Fatalf("operação %q ausente do trace %s; operações=%v", op, traceID, keys(ops))
		}
	}
	// E ao menos um span de gRPC (cliente/servidor) no meio da cadeia.
	temGRPC := false
	for op := range ops {
		if contem(op, "ValidarToken") {
			temGRPC = true
		}
	}
	if !temGRPC {
		t.Fatalf("nenhum span gRPC (ValidarToken) no trace; operações=%v", keys(ops))
	}
}

// esperarTrace consulta a API do Jaeger até o trace aparecer, devolvendo o conjunto
// de operationName dos seus spans.
func esperarTrace(t *testing.T, jaegerQuery, traceID string) map[string]bool {
	t.Helper()
	url := fmt.Sprintf("http://%s/api/traces/%s", jaegerQuery, traceID)

	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url) //nolint:gosec
		if err == nil && resp.StatusCode == http.StatusOK {
			var body struct {
				Data []struct {
					Spans []struct {
						OperationName string `json:"operationName"`
					} `json:"spans"`
				} `json:"data"`
			}
			_ = json.NewDecoder(resp.Body).Decode(&body)
			resp.Body.Close()
			if len(body.Data) > 0 && len(body.Data[0].Spans) > 0 {
				ops := map[string]bool{}
				for _, sp := range body.Data[0].Spans {
					ops[sp.OperationName] = true
				}
				return ops
			}
		} else if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("trace %s não apareceu no Jaeger no prazo", traceID)
	return nil
}

func getComRetry(t *testing.T, ch *amqp.Channel, fila string) amqp.Delivery {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		msg, ok, err := ch.Get(fila, true)
		if err != nil {
			t.Fatalf("get %s: %v", fila, err)
		}
		if ok {
			return msg
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("nenhuma mensagem em %s", fila)
	return amqp.Delivery{}
}

func contem(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
