// Package app monta o servidor IAM já cabeado com validador de JWT e store,
// para uso pelo binário e por testes de integração de outros módulos (o pacote é
// público, ao contrário de internal/*).
package app

import (
	"crypto/rsa"
	"time"

	iamv1 "github.com/machv4/platform/gen/go/construtor/iam/v1"
	"github.com/machv4/platform/services/iam/auth"
	"github.com/machv4/platform/services/iam/internal/server"
	"github.com/machv4/platform/services/iam/internal/store"
)

// NewServer devolve o IAMServiceServer que valida e emite JWT com a chave privada
// priv (a pública é derivada dela) e persiste identidades/permissões no pool
// informado (satisfaz store.DB — ex.: *pgxpool.Pool). ttl é o tempo de vida do JWT.
func NewServer(priv *rsa.PrivateKey, ttl time.Duration, pool store.DB) iamv1.IAMServiceServer {
	st := store.New(pool)
	return server.New(auth.NewValidator(&priv.PublicKey), auth.NewIssuer(priv, ttl), st)
}
