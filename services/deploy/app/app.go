// Package app monta o servidor Deploy Engine já cabeado com o gerenciador de
// versões, para uso pelo binário e por testes de integração (pacote público, ao
// contrário de internal/*).
package app

import (
	deployv1 "github.com/machv4/platform/gen/go/construtor/deploy/v1"
	"github.com/machv4/platform/pkg/database"
	"github.com/machv4/platform/services/deploy/internal/server"
	"github.com/machv4/platform/services/deploy/internal/versions"
)

// NewServer devolve o DeployEngineServiceServer operando sobre o pool informado
// (satisfaz database.Beginner — ex.: *pgxpool.Pool).
func NewServer(pool database.Beginner) deployv1.DeployEngineServiceServer {
	return server.New(versions.New(pool))
}
