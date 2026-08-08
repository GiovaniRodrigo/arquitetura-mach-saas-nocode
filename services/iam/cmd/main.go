// Comando iam sobe o IAM Service gRPC (RF03).
package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"log"
	"net"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"

	iamv1 "github.com/machv4/platform/gen/go/construtor/iam/v1"
	"github.com/machv4/platform/pkg/telemetry"
	"github.com/machv4/platform/pkg/tenantctx"
	"github.com/machv4/platform/services/iam/app"
	"github.com/machv4/platform/services/iam/auth"
)

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	ctx := context.Background()

	addr := env("IAM_GRPC_ADDR", ":50051")
	dsn := env("DATABASE_URL", "postgres://mach:mach@localhost:5432/machv4?sslmode=disable")
	otlp := env("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317")

	shutdown, err := telemetry.Init(ctx, telemetry.Config{ServiceName: "iam", OTLPEndpoint: otlp})
	if err != nil {
		log.Fatalf("telemetry: %v", err)
	}
	defer func() { _ = shutdown(ctx) }()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pool.Close()

	priv, err := loadPrivateKey(env("JWT_PRIVATE_KEY_PATH", ""))
	if err != nil {
		log.Fatalf("chave JWT: %v", err)
	}
	ttl, err := time.ParseDuration(env("JWT_TTL", "12h"))
	if err != nil {
		log.Fatalf("JWT_TTL inválido: %v", err)
	}

	mfaKey := loadMfaKey()

	iam := app.NewServerComMfa(priv, ttl, pool, mfaKey)

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(tenantctx.UnaryServerInterceptor()),
		grpc.ChainStreamInterceptor(tenantctx.StreamServerInterceptor()),
	)
	iamv1.RegisterIAMServiceServer(grpcServer, iam)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen %s: %v", addr, err)
	}
	log.Printf("IAM Service ouvindo em %s", addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

// loadPrivateKey carrega a chave privada RSA (PEM PKCS#1 ou PKCS#8) usada para
// emitir e validar os JWT. Sem caminho, gera um par efêmero (apenas dev) — nesse
// modo, tokens não sobrevivem a reinícios do processo.
func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	if path == "" {
		log.Println("AVISO: JWT_PRIVATE_KEY_PATH não definido — gerando chave efêmera (dev)")
		return auth.GenerateKey()
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, errors.New("PEM inválido")
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaPriv, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("chave não é RSA")
	}
	return rsaPriv, nil
}

// loadMfaKey carrega a chave AES-256-GCM (32 bytes, base64) usada para cifrar
// os segredos TOTP do MFA (spec 004, RF15) em repouso. Diferente de
// loadPrivateKey (chave JWT, que TEM fallback efêmero de dev): aqui NÃO há
// fallback silencioso — a variável é obrigatória e o processo derruba o boot
// (log.Fatalf) se estiver ausente ou não decodificar para exatos 32 bytes.
// Um segredo TOTP "efêmero" recriado a cada restart do processo invalidaria
// os apps autenticadores de todos os usuários com MFA ativo — muito pior do
// que simplesmente travar a subida do serviço.
//
// Gerar um valor válido: openssl rand -base64 32
func loadMfaKey() [32]byte {
	raw := env("IAM_MFA_ENCRYPTION_KEY", "")
	if raw == "" {
		log.Fatalf("IAM_MFA_ENCRYPTION_KEY ausente: obrigatória, sem fallback efêmero (um segredo TOTP recriado a cada restart invalidaria o MFA de todos os usuários). Gere um valor com: openssl rand -base64 32")
	}
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		log.Fatalf("IAM_MFA_ENCRYPTION_KEY inválida: não é base64 válido: %v", err)
	}
	if len(decoded) != 32 {
		log.Fatalf("IAM_MFA_ENCRYPTION_KEY inválida: precisa decodificar para exatos 32 bytes (AES-256); got=%d bytes", len(decoded))
	}
	var chave [32]byte
	copy(chave[:], decoded)
	return chave
}
