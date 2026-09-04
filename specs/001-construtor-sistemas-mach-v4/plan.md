# Implementation Plan: MACH V4 System Builder — Platform Foundation

Strategy: a polyglot monorepo with Protocol Buffers contracts as the single source of truth (`proto/`), Go services sharing internal libraries (`pkg/`), an isolated Elixir collaboration engine, and declarative infrastructure (Docker Compose for dev, Kubernetes/KEDA manifests for production). Implementation proceeds in vertical phases: first the foundation (contracts + local infra + security), then each engine, and finally the cross-cutting layers (messaging, observability, player).

---

## 1. Files to Create/Edit

### 1.1. Contracts and Monorepo Foundation

* **`proto/construtor/common/v1/tenant.proto`**: common messages (`TenantContext`, shared types) — the basis for BR01/NFR02.
* **`proto/construtor/design/v1/design.proto`**: the `DesignEngineService` service (recursive tree CRUD, batched `SalvarDesign`) — FR01, FR06.
* **`proto/construtor/logic/v1/logic.proto`**: the `LogicEngineService` service (`SalvarFormulario` Unary, rules CRUD) — FR02, FR07 (official contract from `doc/CONTRACTS_PERFORMANCE.md §5`).
* **`proto/construtor/iam/v1/iam.proto`**: the `IAMService` service (`AvaliarPermissoes`, `ValidarToken`) — FR03.
* **`proto/construtor/deploy/v1/deploy.proto`**: the `DeployEngineService` service (`Publicar`, `Rollback`) — FR04.
* **`proto/construtor/export/v1/export.proto`**: the `ExportEngineService` service (`CriarJob`, server-streaming `ColetarDados`) — FR05.
* **`buf.yaml` / `buf.gen.yaml`**: lint, breaking-change detection, and Go/Elixir/TypeScript stub generation.
* **`Makefile`**: `proto`, `test`, `up`, `migrate` targets.
* **`docker-compose.yml`**: PostgreSQL, Redis, RabbitMQ, Jaeger, local services.

### 1.2. Shared Go Libraries (`pkg/`)

* **`pkg/tenantctx/`**: extraction/injection of `tenant_id` + identity in gRPC Metadata (BR01, NFR02); server/client interceptors.
* **`pkg/blindindex/`**: generation and verification of the cryptographic hash (HMAC-SHA256 with a per-tenant key) — BR02, NFR08.
* **`pkg/telemetry/`**: OpenTelemetry bootstrap (tracer, W3C propagators, OTLP→Jaeger exporter), `platform.tenant_id` and `platform.component.blind_index` attributes — NFR04.
* **`pkg/database/`**: pgx pool with a mandatory per-tenant filter hook (a guard that rejects queries without `tenant_id`) — BR01.

### 1.3. IAM Service (`services/iam/`)

* **`services/iam/cmd/main.go`**: gRPC server + telemetry bootstrap.
* **`services/iam/internal/auth/jwt.go`**: JWT issuance/validation (RS256), `tenant_id`, `sub`, `tipo` claims.
* **`services/iam/internal/permissions/evaluator.go`**: server-side evaluation of dynamic conditions; returns a `blind_index → {view, click}` map (BR03).
* **`services/iam/internal/store/`**: persistence of hierarchical tenants and permissions.

### 1.4. API Gateway in Go (`gateway/`)

* **`gateway/cmd/main.go`**: HTTP server (chi/echo) + gRPC clients.
* **`gateway/internal/middleware/auth.go`**: JWT validation from the `Authorization` header, injection into Metadata (FR03, NFR02).
* **`gateway/internal/middleware/ratelimit.go`**: per-tenant rate limiting.
* **`gateway/internal/middleware/tracing.go`**: root Trace ID generation (NFR04).
* **`gateway/internal/routes/`**: REST→gRPC translation per resource (`designs.go`, `regras.go`, `deploy.go`, `export.go`, `formularios.go`) — contracts in `contracts/api.md`.

### 1.5. Design Engine (`services/design/`)

* **`services/design/internal/tree/composite.go`**: recursive tree model (`componente_filhos`) with structural validation (FR01).
* **`services/design/internal/store/jsonb.go`**: definition persistence in a JSONB column with `tenant_id` (BR01).
* **`services/design/internal/server/grpc.go`**: implementation of `DesignEngineService`, including the batched `SalvarDesign` consumed by the Elixir engine (BR06).

### 1.6. Logic Engine (`services/logic/`)

* **`services/logic/internal/rules/tree.go`**: decision tree (logic nodes) and interpreter (FR02).
* **`services/logic/internal/validation/schema.go`**: payload revalidation against `CampoDefinicao` by blind_index; anonymized error map (BR02, BR08, NFR08).
* **`services/logic/internal/events/publisher.go`**: publishing of AMQP events with `tenant_id` + `component_blind_index` and trace propagation in the headers (FR08, NFR04).
* **`services/logic/internal/server/grpc.go`**: `SalvarFormulario` and rules CRUD.

### 1.7. Deploy Engine (`services/deploy/`)

* **`services/deploy/internal/versions/manager.go`**: atomic publish/rollback transaction via active flag (BR04, BR05, NFR05).
* **`services/deploy/internal/server/grpc.go`**: `Publicar`, `Rollback`, `ObterVersaoAtiva`.

### 1.8. Export Engine (`services/export/`)

* **`services/export/internal/jobs/manager.go`**: Job lifecycle (criado → coletando → pronto → expirado) (FR05).
* **`services/export/internal/collector/streaming.go`**: consumption via gRPC Server Streaming in chunks (NFR01).
* **`services/export/internal/storage/s3.go`**: bucket upload and short-lived Presigned URL generation.

### 1.9. Elixir Collaboration Engine (`collab/`)

* **`collab/lib/collab_web/channels/screen_channel.ex`**: Phoenix Channel per screen being edited (FR06).
* **`collab/lib/collab/session/screen_server.ex`**: GenServer per screen; in-memory tree; 5s debounce and a single flush via gRPC (BR06).
* **`collab/lib/collab/session/redis_snapshot.ex`**: snapshot replication in Redis.
* **`collab/lib/collab_web/presence.ex`**: Phoenix Presence (CRDT) for cursors and online users.
* **`collab/lib/collab/session/locks.ex`**: optimistic locking by blind_index (BR07).
* **`collab/lib/collab/grpc/design_client.ex`**: gRPC client for the Design Engine.

### 1.10. Asynchronous Workers (`workers/`)

* **`workers/cmd/main.go`**: generic AMQP consumer with `traceparent` extraction from the headers (NFR04).
* **`workers/internal/handlers/webhook.go`** and **`notification.go`**: task execution (FR08).
* **`workers/internal/dlq/`**: routing of persistent failures to the DLQ + tenant alert (BR09, NFR06).

### 1.11. Infrastructure (`infra/`)

* **`infra/rabbitmq/definitions.json`**: exchanges, queues (`webhooks.disparo`, `notificacoes.envio`), DLQs, and fair-queuing policies (BR09).
* **`infra/k8s/keda/scaledobject-workers.yaml`**: `ScaledObject` with `QueueLength` trigger, `minReplicaCount: 0`, `maxReplicaCount: 50` (BR10, NFR03).
* **`infra/k8s/`**: deployments/services for each component.
* **`infra/otel/collector-config.yaml`**: OTLP → Jaeger pipeline (NFR04).
* **`infra/postgres/migrations/`**: versioned migrations (see `data-model.md §4`).

### 1.12. Headless Player (`player/`)

* **`player/src/renderer/CompositeRenderer.tsx`**: recursive rendering by `componente_filhos` (FR01/H of MACH).
* **`player/src/renderer/batcher.ts`**: accumulation of mutations in 16ms windows with a single diff per batch (NFR07).
* **`player/src/validation/blindIndexValidator.ts`**: local validation via the definitions map (BR08).
* **`player/src/permissions/permissionMap.ts`**: application of the `blind_index → {view, click}` boolean map (BR03).
* **`player/src/router/dynamicRoutes.ts`**: SPA navigation via `redirect` actions.

---

## 2. Technical Strategy

### 2.1. Contracts First (True API-first)

All development starts from the `.proto` files compiled with **buf**. No service defines its own boundary types: the generated stubs (Go, Elixir via `grpc-elixir`, TypeScript for the player to consume via the Gateway) are the only interface. CI runs `buf breaking` against `main` to prevent contract breakage — alternative discarded: OpenAPI-first (internal REST), rejected because internal communication is exclusively gRPC (NFR01).

```protobuf
// proto/construtor/logic/v1/logic.proto — official contract (doc/CONTRACTS_PERFORMANCE.md §5)
service LogicEngineService {
  rpc SalvarFormulario (SalvarFormularioRequest) returns (SalvarFormularioResponse);
}
```

### 2.2. Tenant Isolation Enforced by Construction, Not by Discipline

The `tenant_id` filter (BR01) doesn't rely on every developer remembering the `WHERE` clause: `pkg/database` exposes only a `TenantScopedQuerier` that requires `tenantctx.TenantID(ctx)` and injects the predicate automatically. Queries without tenant context fail at runtime and are blocked by lint in CI. Alternative considered: PostgreSQL's native Row-Level Security (`SET app.tenant_id`) — kept as an additional reinforcement at the migration layer (defense in depth), not as the sole mechanism.

### 2.3. Write-Behind Debounce in the GenServer

Every screen being edited lives in a `GenServer` registered via `Registry` by `{sistema_id, screen_id}`. Every mutation resets a 5s timer (`Process.send_after`); on timeout, the consolidated tree is sent in a single `SalvarDesign` call (BR06). Incremental snapshots go to Redis on every mutation for recovery in case the BEAM node goes down. Alternative discarded: persisting every mutation directly to PostgreSQL — rejected due to the write-exhaustion risk documented in `doc/GATEWAY_COLLABORATION.md`.

### 2.4. Trace Propagation Across Three Protocols

A single trace spans HTTP → gRPC → AMQP (NFR04): the Gateway creates the root span; OTel interceptors propagate it via gRPC Metadata; the AMQP publisher serializes `traceparent` into the message headers; the worker extracts it and opens a child span. The `platform.tenant_id` and `platform.component.blind_index` attributes are added to every span — never real field names (NFR08).

### 2.5. Scale-to-Zero with KEDA

Workers are Deployments with `replicas` managed by KEDA via a `ScaledObject` (`rabbitmq` trigger, `QueueLength` metric). Empty queue → 0 pods; buildup → scales up to 50 (BR10, NFR03). Fair queuing (BR09) is implemented with per-tenant routing keys + an `x-max-priority` policy/low consumer prefetch, and a per-queue DLQ with a tenant-targeted alert.

---

## 3. Dependencies and Prerequisites

- [ ] Go ≥ 1.22, Elixir ≥ 1.16/OTP 26, Node ≥ 20 installed
- [ ] `buf` CLI installed for contract generation/validation
- [ ] Docker + Docker Compose for local infrastructure (PostgreSQL 16, Redis 7, RabbitMQ 3.13 with the management plugin, Jaeger)
- [ ] Kubernetes cluster with KEDA installed (only for validating production manifests; dev uses Compose)
- [ ] S3/GCS bucket (or local MinIO) for the Export Engine
- [ ] RS256 key pair for JWT signing in the IAM Service

---

## 4. Risks and Points of Attention

| Risk | Impact | Mitigation |
|-------|---------|-----------|
| Cross-tenant leakage from an unfiltered query (BR01) | High | Mandatory `TenantScopedQuerier` + PostgreSQL RLS as a second layer + multi-tenant integration test in CI |
| Loss of edits if the BEAM node crashes before the 5s flush (BR06) | High | Redis snapshot on every mutation; on restart the GenServer rehydrates from Redis before accepting connections |
| Exposure of real names in logs/traces (NFR08) | High | Central redactor in `pkg/telemetry`; review of every log call site in code review; a test that greps for schema names in error payloads |
| Broken gRPC contract between polyglot services | Medium | `buf breaking` in CI + `v1` versioning on proto packages |
| Noisy neighbor exhausting workers (BR09) | Medium | Fair queuing via tenant routing key, low prefetch, isolated DLQ, and per-tenant alerts |
| Elixir + Go in the same monorepo complicates CI | Medium | Separate pipelines per directory with independent caching; a unified Makefile only for local dev |
| Leaked Presigned URL (FR05) | Medium | Short expiration (minutes), optional IP-bound URL, private bucket with public access blocked |
| Ordering of concurrent mutations on the same component (BR07) | Medium | Optimistic locking by blind_index in the GenServer (single source of ordering per screen) |
