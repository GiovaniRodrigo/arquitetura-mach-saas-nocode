// Package app monta o servidor Design Engine já cabeado com o store JSONB, para
// uso pelo binário e por testes de integração (pacote público, ao contrário de
// internal/*).
package app

import (
	designv1 "github.com/machv4/platform/gen/go/construtor/design/v1"
	"github.com/machv4/platform/pkg/database"
	"github.com/machv4/platform/services/design/internal/server"
	"github.com/machv4/platform/services/design/internal/store"
)

// NewServer devolve o DesignEngineServiceServer persistindo no pool informado
// (satisfaz database.Beginner — ex.: *pgxpool.Pool).
func NewServer(pool database.Beginner) designv1.DesignEngineServiceServer {
	return server.New(store.New(pool))
}
