# Interfaces: Resource Monitor

## `proto/construtor/health/v1/health.proto`

Implemented by IAM, Design, Logic, Deploy, and Export via `pkg/health` (FR01). The
"how am I doing" contract.

```protobuf
syntax = "proto3";

package construtor.health.v1;

option go_package = "github.com/machv4/platform/gen/go/construtor/health/v1;healthv1";

message ObterStatusRequest {}

message ServicoStatus {
  string nome = 1;                     // e.g.: "iam", "design"
  string tipo = 2;                     // "grpc" | "http"
  string status = 3;                   // "servindo" | "indisponivel"
  int64 uptime_segundos = 4;
  int64 memoria_alocada_bytes = 5;      // heap in use (Go: MemStats.Alloc / Elixir: erlang.memory(:total))
  int64 memoria_sistema_bytes = 6;      // total OS memory (Go: MemStats.Sys); 0/absent when the runtime doesn't expose it (BR02)
  int32 goroutines = 7;                 // 0/absent for runtimes without goroutines (Elixir uses process_count, see the note in the Collab server)
  string mensagem_erro = 8;             // filled in only when status = "indisponivel"
}

service RecursosService {
  rpc ObterStatus(ObterStatusRequest) returns (ServicoStatus);
}
```

**Expected implementations**: `pkg/health.Server` (Go, a single one, reused by the 5
services via `Registrar`). Collab and Workers **do not** implement this proto — they
respond via plain HTTP with the same field shape in JSON (see `contracts/api.md`),
consumed by the Monitor's `ColetorHTTP`.

---

## `proto/construtor/monitor/v1/monitor.proto`

Implemented only by the `services/monitor` service. The "how is everyone doing"
contract (FR04).

```protobuf
syntax = "proto3";

package construtor.monitor.v1;

import "construtor/health/v1/health.proto";

option go_package = "github.com/machv4/platform/gen/go/construtor/monitor/v1;monitorv1";

message ObterRecursosRequest {}

message ObterRecursosResponse {
  repeated construtor.health.v1.ServicoStatus servicos = 1;
  int64 coletado_em_unix = 2; // collection timestamp, so the Frontend knows "how long ago" (FR06/FR07)
}

service MonitorService {
  rpc ObterRecursos(ObterRecursosRequest) returns (ObterRecursosResponse);
}
```

**Expected implementations**: `services/monitor/internal/server.MonitorServer`.

---

## `services/monitor/internal/poller.Coletor` (Go, internal — not proto-generated)

```go
package poller

import (
	"context"

	healthv1 "github.com/machv4/platform/gen/go/construtor/health/v1"
)

// Coletor abstracts how to obtain the ServicoStatus of a monitored service,
// independent of transport (gRPC or HTTP) — see plan.md §2 (Strategy).
type Coletor interface {
	Nome() string
	Coletar(ctx context.Context) (*healthv1.ServicoStatus, error)
}
```

**Expected implementations**:
- `ColetorGRPC` (`coletor_grpc.go`) — used for IAM, Design, Logic, Deploy, Export.
- `ColetorHTTP` (`coletor_http.go`) — used for Gateway (`/health`), Collab (`/healthz`),
  Workers (`/health`); parses the JSON from each endpoint (slightly different formats —
  see `contracts/api.md`) and builds the same output `*healthv1.ServicoStatus`.
