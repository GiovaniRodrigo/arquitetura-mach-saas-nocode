// Package app monta o servidor Export Engine já cabeado com jobs, coletor e
// storage, para uso pelo binário e por testes de integração (pacote público, ao
// contrário de internal/*).
package app

import (
	exportv1 "github.com/machv4/platform/gen/go/construtor/export/v1"
	"github.com/machv4/platform/pkg/database"
	"github.com/machv4/platform/services/export/internal/collector"
	"github.com/machv4/platform/services/export/internal/jobs"
	"github.com/machv4/platform/services/export/internal/server"
	"github.com/machv4/platform/services/export/internal/storage"
)

// NewServer devolve o ExportEngineServiceServer persistindo no pool informado e
// carregando artefatos no armazenamento fornecido.
func NewServer(pool database.Beginner, store storage.Armazenamento) exportv1.ExportEngineServiceServer {
	return server.New(jobs.New(pool), collector.New(pool), store)
}
