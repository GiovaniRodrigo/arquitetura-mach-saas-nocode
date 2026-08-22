// Comando workers consome eventos assíncronos do RabbitMQ e os processa fora do
// fluxo síncrono (RF08). Cada mensagem tem o trace extraído do cabeçalho
// (traceparent, RNF04), é roteada ao handler do seu tipo e, em falha, segue para
// re-tentativa com backoff ou para a DLQ com alerta ao tenant (RN09, RNF06).
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"

	"github.com/machv4/platform/pkg/eventbus"
	"github.com/machv4/platform/pkg/telemetry"
	"github.com/machv4/platform/services/workers/internal/consumer"
	"github.com/machv4/platform/services/workers/internal/dlq"
	"github.com/machv4/platform/services/workers/internal/handlers"
	workershealth "github.com/machv4/platform/services/workers/internal/health"
)

const (
	maxTentativas = 5
	prefetch      = 1 // fair dispatch: um worker não açambarca a fila (RN09)
)

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	inicio := time.Now()
	ctx := context.Background()

	amqpURL := env("RABBITMQ_URL", "amqp://mach:mach@localhost:5672/")
	otlp := env("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317")
	httpAddr := env("WORKERS_HTTP_ADDR", ":8081")

	shutdown, err := telemetry.Init(ctx, telemetry.Config{ServiceName: "workers", OTLPEndpoint: otlp})
	if err != nil {
		log.Fatalf("telemetry: %v", err)
	}
	defer func() { _ = shutdown(ctx) }()

	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		log.Fatalf("rabbitmq: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("canal: %v", err)
	}
	defer ch.Close()

	if err := ch.Qos(prefetch, 0, false); err != nil {
		log.Fatalf("qos: %v", err)
	}

	roteador := consumer.NewRoteador(map[string]handlers.Handler{
		eventbus.TipoWebhook:     handlers.NewWebhook(nil),
		eventbus.TipoNotificacao: handlers.NewNotificacao(nil),
	}, maxTentativas)

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	httpServer := &http.Server{Addr: httpAddr, Handler: workershealth.NovoHandler(inicio)}
	go func() {
		log.Printf("workers health HTTP ouvindo em %s", httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("health HTTP encerrado: %v", err)
		}
	}()

	filas := append(eventbus.FilasDe(eventbus.TipoWebhook), eventbus.FilasDe(eventbus.TipoNotificacao)...)

	var wg sync.WaitGroup
	for _, fila := range filas {
		wg.Add(1)
		go func(fila string) {
			defer wg.Done()
			if err := consumir(ctx, ch, roteador, fila); err != nil {
				log.Printf("consumo de %s encerrado: %v", fila, err)
			}
		}(fila)
	}

	log.Printf("workers consumindo %d filas-shard: %v", len(filas), filas)
	<-ctx.Done()
	wg.Wait()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown do health HTTP: %v", err)
	}
}

func consumir(ctx context.Context, ch *amqp.Channel, roteador *consumer.Roteador, fila string) error {
	entregas, err := ch.Consume(fila, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume %s: %w", fila, err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case d, ok := <-entregas:
			if !ok {
				return nil
			}
			tratar(ctx, ch, roteador, fila, d)
		}
	}
}

func tratar(ctx context.Context, ch *amqp.Channel, roteador *consumer.Roteador, fila string, d amqp.Delivery) {
	// Continua o trace do publisher (RNF04).
	msgCtx := otel.GetTextMapPropagator().Extract(ctx, eventbus.HeaderCarrier(d.Headers))

	var corpo eventbus.Corpo
	if err := json.Unmarshal(d.Body, &corpo); err != nil {
		// Mensagem ilegível: direto à DLQ (não re-tentável).
		_ = publicar(ch, eventbus.FilaDLQ(fila), d.Headers, d.Body, "")
		_ = d.Ack(false)
		return
	}

	tentativas := dlq.Tentativas(d.Headers)
	desfecho, procErr := roteador.Processar(msgCtx, corpo, tentativas)

	switch desfecho {
	case consumer.Ack:
		_ = d.Ack(false)

	case consumer.Retry:
		// Re-publica na fila de retry com backoff exponencial e tentativa++.
		headers := dlq.ProximosHeaders(d.Headers)
		if err := publicarComTTL(ch, eventbus.FilaRetry(fila), headers, d.Body, backoff(tentativas)); err != nil {
			log.Printf("retry de %s falhou: %v", fila, err)
			_ = d.Nack(false, true)
			return
		}
		_ = d.Ack(false)

	case consumer.DeadLetter:
		alerta := dlq.AlertaDe(d.Headers, corpo.Tipo, motivo(procErr))
		log.Printf("DLQ tenant=%s tipo=%s motivo=%s", alerta.TenantID, alerta.Tipo, alerta.Motivo)
		_ = publicar(ch, eventbus.FilaDLQ(fila), d.Headers, d.Body, "")
		_ = d.Ack(false)
	}
}

// publicar envia à fila `dest` pelo exchange default (routing key = nome da fila).
func publicar(ch *amqp.Channel, dest string, headers amqp.Table, body []byte, expiration string) error {
	return ch.PublishWithContext(context.Background(), "", dest, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Headers:      headers,
		Body:         body,
		Expiration:   expiration,
	})
}

func publicarComTTL(ch *amqp.Channel, dest string, headers amqp.Table, body []byte, ttl time.Duration) error {
	return publicar(ch, dest, headers, body, strconv.FormatInt(ttl.Milliseconds(), 10))
}

// backoff exponencial limitado, a partir do número de tentativas já feitas.
func backoff(tentativas int) time.Duration {
	d := time.Duration(math.Pow(2, float64(tentativas))) * time.Second
	if d > 60*time.Second {
		return 60 * time.Second
	}
	return d
}

func motivo(err error) string {
	if err == nil {
		return "tipo desconhecido"
	}
	return err.Error()
}
