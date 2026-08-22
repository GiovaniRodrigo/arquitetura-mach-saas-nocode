# Interfaces: Monitor de Recursos

## `proto/construtor/health/v1/health.proto`

Implementado por IAM, Design, Logic, Deploy e Export via `pkg/health` (RF01). Contrato de
"como estou eu".

```protobuf
syntax = "proto3";

package construtor.health.v1;

option go_package = "github.com/machv4/platform/gen/go/construtor/health/v1;healthv1";

message ObterStatusRequest {}

message ServicoStatus {
  string nome = 1;                     // ex.: "iam", "design"
  string tipo = 2;                     // "grpc" | "http"
  string status = 3;                   // "servindo" | "indisponivel"
  int64 uptime_segundos = 4;
  int64 memoria_alocada_bytes = 5;      // heap em uso (Go: MemStats.Alloc / Elixir: erlang.memory(:total))
  int64 memoria_sistema_bytes = 6;      // memória total do SO (Go: MemStats.Sys); 0/ausente quando o runtime não expõe (RN02)
  int32 goroutines = 7;                 // 0/ausente para runtimes sem goroutines (Elixir usa process_count, ver nota no server Collab)
  string mensagem_erro = 8;             // preenchido só quando status = "indisponivel"
}

service RecursosService {
  rpc ObterStatus(ObterStatusRequest) returns (ServicoStatus);
}
```

**Implementações esperadas**: `pkg/health.Server` (Go, único, reaproveitado pelos 5
serviços via `Registrar`). Collab e Workers **não** implementam este proto — respondem via
HTTP simples com o mesmo shape de campos em JSON (ver `contracts/api.md`), consumidos pelo
`ColetorHTTP` do Monitor.

---

## `proto/construtor/monitor/v1/monitor.proto`

Implementado só pelo serviço `services/monitor`. Contrato de "como estão todos" (RF04).

```protobuf
syntax = "proto3";

package construtor.monitor.v1;

import "construtor/health/v1/health.proto";

option go_package = "github.com/machv4/platform/gen/go/construtor/monitor/v1;monitorv1";

message ObterRecursosRequest {}

message ObterRecursosResponse {
  repeated construtor.health.v1.ServicoStatus servicos = 1;
  int64 coletado_em_unix = 2; // timestamp da coleta, para o Frontend saber "há quanto tempo" (RF06/RF07)
}

service MonitorService {
  rpc ObterRecursos(ObterRecursosRequest) returns (ObterRecursosResponse);
}
```

**Implementações esperadas**: `services/monitor/internal/server.MonitorServer`.

---

## `services/monitor/internal/poller.Coletor` (Go, interno — não gerado de proto)

```go
package poller

import (
	"context"

	healthv1 "github.com/machv4/platform/gen/go/construtor/health/v1"
)

// Coletor abstrai como obter o ServicoStatus de um serviço monitorado,
// independente do transporte (gRPC ou HTTP) — ver plan.md §2 (Strategy).
type Coletor interface {
	Nome() string
	Coletar(ctx context.Context) (*healthv1.ServicoStatus, error)
}
```

**Implementações esperadas**:
- `ColetorGRPC` (`coletor_grpc.go`) — usado para IAM, Design, Logic, Deploy, Export.
- `ColetorHTTP` (`coletor_http.go`) — usado para Gateway (`/health`), Collab (`/healthz`),
  Workers (`/health`); faz o parse do JSON de cada endpoint (formatos ligeiramente
  diferentes — ver `contracts/api.md`) e monta o mesmo `*healthv1.ServicoStatus` de saída.
