package health

import (
	"time"

	"google.golang.org/grpc"

	healthv1 "github.com/machv4/platform/gen/go/construtor/health/v1"
)

// Registrar cadastra o RecursosService no grpc.Server do serviço chamador.
// Chamada única no main.go, logo após criar o *grpc.Server — mesmo ponto onde
// cada serviço já registra seu próprio serviço de negócio.
func Registrar(grpcServer *grpc.Server, nome string, inicio time.Time) {
	healthv1.RegisterRecursosServiceServer(grpcServer, NewServer(nome, inicio))
}
