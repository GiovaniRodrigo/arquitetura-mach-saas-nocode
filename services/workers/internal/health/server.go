// Package health implementa o endpoint HTTP mínimo de recursos do Workers
// (spec 008-monitor-recursos, RF03). Diferente dos demais serviços Go (IAM,
// Design, Logic, Deploy, Export), Workers não tem um grpc.Server — só consome
// RabbitMQ — então aqui a exposição é HTTP simples com encoding/json, em vez
// de reaproveitar pkg/health (gRPC-specific). A lógica de leitura de
// runtime.MemStats/NumGoroutine é a mesma inspiração de pkg/health/server.go.
package health

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"
)

// resposta é o shape JSON documentado em contracts/api.md para GET /health.
type resposta struct {
	Status              string `json:"status"`
	UptimeSegundos      int64  `json:"uptime_segundos"`
	MemoriaAlocadaBytes int64  `json:"memoria_alocada_bytes"`
	MemoriaSistemaBytes int64  `json:"memoria_sistema_bytes"`
	Goroutines          int32  `json:"goroutines"`
}

// NovoHandler cria o http.Handler que responde GET /health com status, uptime
// e memória do processo Workers. inicio é o instante em que o processo
// começou a consumir filas, usado para calcular o uptime.
func NovoHandler(inicio time.Time) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		corpo := resposta{
			Status:              "servindo",
			UptimeSegundos:      int64(time.Since(inicio).Seconds()),
			MemoriaAlocadaBytes: int64(m.Alloc),
			MemoriaSistemaBytes: int64(m.Sys),
			Goroutines:          int32(runtime.NumGoroutine()),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(corpo)
	})
}
