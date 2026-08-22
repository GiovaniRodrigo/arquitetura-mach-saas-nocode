// Package health implementa o RecursosService (construtor.health.v1) — o
// contrato de "como estou eu" que IAM, Design, Logic, Deploy e Export expõem
// de forma idêntica, para que o serviço Monitor (spec 008-monitor-recursos,
// RF01) possa consultar status/uptime/memória de forma uniforme.
package health

import (
	"context"
	"runtime"
	"time"

	healthv1 "github.com/machv4/platform/gen/go/construtor/health/v1"
)

// Server implementa healthv1.RecursosServiceServer para um único serviço,
// identificado por nome e pelo instante em que subiu.
type Server struct {
	healthv1.UnimplementedRecursosServiceServer

	nome   string
	inicio time.Time
}

// NewServer cria o Server de recursos de um serviço gRPC. inicio é o instante
// em que o processo começou a servir, usado para calcular o uptime.
func NewServer(nome string, inicio time.Time) *Server {
	return &Server{nome: nome, inicio: inicio}
}

// ObterStatus lê runtime.MemStats/NumGoroutine do processo atual — nunca
// retorna erro (um serviço vivo sempre consegue relatar a si mesmo).
func (s *Server) ObterStatus(context.Context, *healthv1.ObterStatusRequest) (*healthv1.ServicoStatus, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &healthv1.ServicoStatus{
		Nome:                s.nome,
		Tipo:                "grpc",
		Status:              "servindo",
		UptimeSegundos:      int64(time.Since(s.inicio).Seconds()),
		MemoriaAlocadaBytes: int64(m.Alloc),
		MemoriaSistemaBytes: int64(m.Sys),
		Goroutines:          int32(runtime.NumGoroutine()),
	}, nil
}
