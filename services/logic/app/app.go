// Package app monta o servidor Logic Engine já cabeado com o store, para uso pelo
// binário e por testes de integração (pacote público, ao contrário de internal/*).
package app

import (
	logicv1 "github.com/machv4/platform/gen/go/construtor/logic/v1"
	"github.com/machv4/platform/pkg/database"
	"github.com/machv4/platform/services/logic/internal/server"
	"github.com/machv4/platform/services/logic/internal/store"
)

// NewServer devolve o LogicEngineServiceServer persistindo no pool informado
// (satisfaz database.Beginner — ex.: *pgxpool.Pool).
func NewServer(pool database.Beginner) logicv1.LogicEngineServiceServer {
	return server.New(store.New(pool))
}
