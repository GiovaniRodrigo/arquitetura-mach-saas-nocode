// Package app monta o servidor IAM já cabeado com validador de JWT e store,
// para uso pelo binário e por testes de integração de outros módulos (o pacote é
// público, ao contrário de internal/*).
package app

import (
	"crypto/rsa"

	iamv1 "github.com/machv4/platform/gen/go/construtor/iam/v1"
	"github.com/machv4/platform/services/iam/auth"
	"github.com/machv4/platform/services/iam/internal/server"
	"github.com/machv4/platform/services/iam/internal/store"
)

// NewServer devolve o IAMServiceServer validando tokens com pub e carregando
// permissões do pool informado (satisfaz store.DB — ex.: *pgxpool.Pool).
func NewServer(pub *rsa.PublicKey, pool store.DB) iamv1.IAMServiceServer {
	return server.New(auth.NewValidator(pub), store.New(pool))
}
