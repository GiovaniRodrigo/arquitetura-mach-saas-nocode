# Tasks: MACH V4 System Builder — Platform Foundation

<!-- Ordered by execution dependency. Each task is atomic (≤ 1 day). -->

## Phase 0 — Foundation and Contracts

- [x] 1. Initialize the monorepo: folder structure, `Makefile`, `.gitignore`, Go workspace (`go.work`) (`Makefile`, `go.work`)
- [x] 2. Write `docker-compose.yml` with PostgreSQL 16, Redis 7, RabbitMQ 3.13 (management), Jaeger, and MinIO (`docker-compose.yml`)
- [x] 3. Create the common tenant/identity contract (`proto/construtor/common/v1/tenant.proto`) [NFR02]
- [x] 4. Transcribe the official Logic Engine contract from `doc/CONTRACTS_PERFORMANCE.md §5` and extend it with the rules CRUD (`proto/construtor/logic/v1/logic.proto`) [FR02, FR07]
- [x] 5. Create the Design, IAM, Deploy, and Export contracts (`proto/construtor/{design,iam,deploy,export}/v1/*.proto`) [FR01, FR03, FR04, FR05]
- [x] 6. Configure buf (lint + breaking + Go/Elixir/TS generation) and the `make proto` target (`buf.yaml`, `buf.gen.yaml`)
- [x] 7. Write migrations 0001–0009 per `data-model.md §4` (`infra/postgres/migrations/`) [BR01, BR02, BR04]
- [x] 8. Write migration 0010 for Row-Level Security by `tenant_id` (`infra/postgres/migrations/0010_enable_row_level_security.sql`) [BR01]

## Phase 1 — Shared Go Libraries

- [x] 9. Implement `pkg/tenantctx` with tests: tenant extraction/injection in gRPC Metadata + interceptors (`pkg/tenantctx/`) [BR01, NFR02]
- [x] 10. Implement `pkg/blindindex` with tests: HMAC-SHA256 with a per-tenant key (`pkg/blindindex/`) [BR02, NFR08]
- [x] 11. Implement `pkg/database` with tests: `TenantScopedQuerier` that injects the tenant filter and rejects queries without context (`pkg/database/`) [BR01]
- [x] 12. Implement `pkg/telemetry`: OTel bootstrap, W3C propagator, `platform.tenant_id`/`platform.component.blind_index` attributes, sensitive-data redactor (`pkg/telemetry/`) [NFR04, NFR08]

## Phase 2 — IAM Service and Gateway

- [x] 13. IAM: hierarchical tenant and role store with tests (`services/iam/internal/store/`) [FR03]
- [x] 14. IAM: RS256 JWT issuance and validation with `tenant_id`/`sub`/`tipo` claims (`services/iam/internal/auth/jwt.go`) [FR03]
- [x] 15. IAM: server-side permission evaluator returning a `blind_index → {view, click}` map, with tests (`services/iam/internal/permissions/evaluator.go`) [BR03]
- [x] 16. IAM: gRPC server + telemetry bootstrap (`services/iam/internal/server/grpc.go`, `services/iam/cmd/main.go`)
- [x] 17. Gateway: JWT authentication middleware → gRPC Metadata (`gateway/internal/middleware/auth.go`) [FR03, NFR02]
- [x] 18. Gateway: per-tenant rate-limiting and root-tracing middlewares (`gateway/internal/middleware/{ratelimit,tracing}.go`) [NFR04]
- [x] 19. Gateway: HTTP bootstrap + gRPC clients + permissions route (`gateway/cmd/main.go`, `gateway/internal/routes/permissions.go`) [FR03]
- [x] 20. Integration test: request without JWT → 401; tenant A's JWT never accesses tenant B's data (`gateway/tests/auth_integration_test.go`) [BR01, criterion 1]

## Phase 3 — Design Engine

- [x] 21. Composite model of the recursive tree with structural validation and tests (`services/design/internal/tree/composite.go`) [FR01]
- [x] 22. JSONB persistence with `TenantScopedQuerier` (`services/design/internal/store/jsonb.go`) [FR01, BR01]
- [x] 23. Design Engine gRPC server including batched `SalvarDesign` (`services/design/internal/server/grpc.go`, `services/design/cmd/main.go`) [FR01, BR06]
- [x] 24. Gateway: REST→gRPC design routes (`gateway/internal/routes/designs.go`) [FR01]

## Phase 4 — Logic Engine

- [x] 25. Decision tree (logic nodes) and interpreter with tests (`services/logic/internal/rules/tree.go`) [FR02]
- [x] 26. CRUD of `campos_definicao` by blind_index (`services/logic/internal/store/campos.go`) [BR02]
- [x] 27. Payload revalidation against the schema with a blind_index error map, and tests (`services/logic/internal/validation/schema.go`) [BR08, NFR08, criterion 2]
- [x] 28. gRPC server `SalvarFormulario` + persistence in `dados_operacionais` (`services/logic/internal/server/grpc.go`, `services/logic/cmd/main.go`) [FR07]
- [x] 29. Gateway: rules and forms routes (`gateway/internal/routes/{regras,formularios}.go`) [FR02, FR07]
- [x] 30. Integration test: a malicious submission directly to the API is rejected without exposing real names (`services/logic/tests/validation_integration_test.go`) [BR08, NFR08, criterion 2]

## Phase 5 — Deploy Engine

- [x] 31. Version manager: publish/rollback in an atomic transaction with a partial unique index (`services/deploy/internal/versions/manager.go`) [BR04, BR05]
- [x] 32. gRPC server `Publicar`/`Rollback`/`ObterVersaoAtiva` + Gateway routes (`services/deploy/internal/server/grpc.go`, `gateway/internal/routes/deploy.go`) [FR04]
- [x] 33. Integration test: rollback < 100ms and uniqueness of the active flag under concurrency (`services/deploy/tests/rollback_test.go`) [NFR05, criterion 3]

## Phase 6 — Collaboration Engine (Elixir)

- [x] 34. Phoenix project bootstrap (no HTML), socket, and `ScreenChannel` with JWT authorization (`collab/lib/collab_web/channels/screen_channel.ex`) [FR06]
- [x] 35. `ScreenServer` (GenServer per screen via Registry): tree state + mutation application (`collab/lib/collab/session/screen_server.ex`) [FR06]
- [x] 36. Incremental snapshots in Redis + rehydration on startup (`collab/lib/collab/session/redis_snapshot.ex`) [BR06, risk of losing edits]
- [x] 37. 5s debounce and a single flush via the `SalvarDesign` gRPC client (`collab/lib/collab/session/screen_server.ex`, `collab/lib/collab/grpc/design_client.ex`) [BR06, criterion 4]
- [x] 38. Phoenix Presence for cursors/online users (`collab/lib/collab_web/presence.ex`) [FR06]
- [x] 39. Optimistic locking by blind_index with timeout release (`collab/lib/collab/session/locks.ex`) [BR07]
- [x] 40. ExUnit test: A's mutation reaches B; 5s of silence → exactly 1 gRPC call (`collab/test/collab/session/screen_server_test.exs`) [BR06, criterion 4]

## Phase 7 — Asynchronous Messaging and Workers

- [x] 41. RabbitMQ definitions: exchanges, `webhooks.disparo`/`notificacoes.envio` queues, DLQs, and fair-queuing policies (`infra/rabbitmq/definitions.json`) [BR09]
- [x] 42. AMQP publisher in the Logic Engine with `tenant_id`, `component_blind_index`, and `traceparent` in the headers (`services/logic/internal/events/publisher.go`) [FR08, NFR04]
- [x] 43. Generic consumer worker with trace extraction and webhook/notification handlers (`workers/cmd/main.go`, `workers/internal/handlers/`) [FR08]
- [x] 44. Routing of persistent failures to the DLQ + tenant alert (`workers/internal/dlq/`) [BR09, NFR06]
- [x] 45. Worker k8s manifests + KEDA `ScaledObject` (`minReplicaCount: 0`, `maxReplicaCount: 50`, QueueLength trigger) (`infra/k8s/keda/scaledobject-workers.yaml`) [BR10, NFR03, criterion 5]

## Phase 8 — Export Engine

- [x] 46. Job lifecycle manager with states and tests (`services/export/internal/jobs/manager.go`) [FR05]
- [x] 47. Collector via gRPC Server Streaming in chunks from the 3 source services (`services/export/internal/collector/streaming.go`) [FR05, NFR01]
- [x] 48. S3/MinIO upload + short-lived Presigned URL + Gateway routes (`services/export/internal/storage/s3.go`, `gateway/internal/routes/export.go`) [FR05, criterion 7]

## Phase 9 — End-to-End Observability

- [x] 49. OTel Collector config + wiring Jaeger into Compose (`infra/otel/collector-config.yaml`, `docker-compose.yml`) [NFR04]
- [x] 50. Instrument the Elixir engine with OpenTelemetry (channel spans and gRPC flush) (`collab/lib/collab/telemetry.ex`) [NFR04]
- [x] 51. E2E trace test: a `traceparent` spans Gateway → gRPC → AMQP → Worker in a single trace in Jaeger (`tests/e2e/tracing_test.go`) [NFR04, criterion 6]

## Phase 10 — Headless Player

- [x] 52. Vite + React + TS bootstrap and an authenticated HTTP client (`player/package.json`, `player/src/api/client.ts`)
- [x] 53. `CompositeRenderer`: recursive rendering by `componente_filhos` from the active version (`player/src/renderer/CompositeRenderer.tsx`) [FR01, BR04]
- [x] 54. 16ms batcher with a single diff per batch + timing test (`player/src/renderer/batcher.ts`) [NFR07]
- [x] 55. Local validator via the blind_index map (`player/src/validation/blindIndexValidator.ts`) [BR08]
- [x] 56. Application of the `{view, click}` permission map and dynamic SPA routes (`player/src/permissions/permissionMap.ts`, `player/src/router/dynamicRoutes.ts`) [BR03]
- [x] 57. Phoenix WebSocket client for collaboration in builder-preview mode (`player/src/collab/phoenixSocket.ts`) [FR06]

## Phase 11 — Closeout

- [x] 58. CI pipeline: buf lint/breaking, Go tests, ExUnit, player tests, bringing up Compose for integration (`.github/workflows/ci.yml`)
- [x] 59. Run the full test suite (Go + ExUnit + player + integration + E2E) and fix regressions (`make test`)
