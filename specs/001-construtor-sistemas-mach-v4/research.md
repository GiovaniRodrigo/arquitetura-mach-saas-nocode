# Research: MACH V4 System Builder

---

## 1. Existing Patterns in the Project

The repository is greenfield (architecture documentation only). The following patterns come from the documents and are binding for the implementation:

| File/Pattern | Location | Relevance |
|----------------|-------------|-----------|
| MACH pillars and the 5-microservice split | `doc/ARCHITECTURE_PILLARS.md` | Defines service boundaries — no service may absorb another's responsibility |
| Hybrid Go/Elixir Gateway + write-behind debounce | `doc/GATEWAY_COLLABORATION.md` | Mandatory collaboration flow (GenServer, Redis snapshot, 5s flush, gRPC batch) |
| Blind Index, JWT→Metadata, boolean permission map | `doc/DATA_SECURITY.md` | Security contracts — the permission payload format is already specified |
| KEDA `ScaledObject`, named queues, DLQ, fair queuing | `doc/ASYNC_OBSERVABILITY.md` | Messaging topology and queue names (`webhooks.disparo`, `notificacoes.envio`) |
| Official `LogicEngineService` `.proto` contract | `doc/CONTRACTS_PERFORMANCE.md §5` | Must be transcribed verbatim into `proto/construtor/logic/v1/` |
| 16ms render batching, flag-based deploy, 4-step export | `doc/CONTRACTS_PERFORMANCE.md` | Verifiable numeric requirements (16ms, rollback in ms, presigned URL) |
| Numbered RF/RNF/RN requirements | `doc/ANALISE_REQUISITOS.md` | Source of the traceability used in spec/plan/tasks |

---

## 2. Technologies and Libraries

| Technology | Version | Use | Already installed? |
|------------|--------|-----|---------------|
| Go | ≥ 1.22 | Gateway, 5 microservices, workers | No |
| Elixir / OTP | ≥ 1.16 / 26 | Collaboration engine (Phoenix) | No |
| Phoenix / Phoenix Channels | ~> 1.7 | WebSockets, Channels, Presence | No |
| grpc-elixir (`grpc`) + `protobuf-elixir` | ~> 0.9 | Elixir gRPC client → Design Engine | No |
| buf | ≥ 1.30 | Lint, generation, and breaking-check for `.proto` | No |
| grpc-go + protoc-gen-go | latest | gRPC servers/clients in Go | No |
| pgx | v5 | PostgreSQL driver with JSONB support | No |
| PostgreSQL | 16 | Shared multi-tenant database (JSONB, RLS, partial index) | No (via Compose) |
| Redis | 7 | Collaboration snapshots + rate limiting | No (via Compose) |
| RabbitMQ | 3.13 | Asynchronous messaging, DLQ | No (via Compose) |
| KEDA | ≥ 2.14 | Autoscaling by QueueLength, scale-to-zero | No (k8s cluster) |
| OpenTelemetry SDK (Go/Elixir) | latest | HTTP/gRPC/AMQP instrumentation | No |
| Jaeger | latest | Trace backend | No (via Compose) |
| MinIO | latest | Local S3 for the Export Engine in dev | No (via Compose) |
| React + TypeScript + Vite | React 18 / TS 5 | Headless Player (SPA) | No |
| golang-jwt/jwt | v5 | RS256 JWT issuance/validation | No |

---

## 3. External References

| Reference | URL | What it addresses |
|------------|-----|--------------|
| KEDA RabbitMQ Scaler | https://keda.sh/docs/latest/scalers/rabbitmq-queue/ | Configuring the `QueueLength` trigger and scale-to-zero (BR10) |
| W3C Trace Context | https://www.w3.org/TR/trace-context/ | Format of the `traceparent` header propagated over HTTP/gRPC/AMQP (NFR04) |
| OpenTelemetry Messaging Semantics | https://opentelemetry.io/docs/specs/semconv/messaging/ | Convention for AMQP producer/consumer spans |
| gRPC Metadata (Go) | https://grpc.io/docs/guides/metadata/ | Propagating identity/tenant as binary metadata (NFR02) |
| Phoenix Channels | https://hexdocs.pm/phoenix/channels.html | Per-screen WebSocket channels (FR06) |
| Phoenix Presence | https://hexdocs.pm/phoenix/Phoenix.Presence.html | Presence/cursors via CRDT (FR06) |
| PostgreSQL Row-Level Security | https://www.postgresql.org/docs/16/ddl-rowsecurity.html | Defense in depth for BR01 |
| RabbitMQ Dead Letter Exchanges | https://www.rabbitmq.com/docs/dlx | Per-queue DLQ (BR09, NFR06) |
| AWS S3 Presigned URLs | https://docs.aws.amazon.com/AmazonS3/latest/userguide/using-presigned-url.html | Secure export delivery (FR05) |
| Blind Indexing (CipherSweet) | https://ciphersweet.paragonie.com/internals/blind-index-planning | Cryptographic basis of the Blind Index (BR02) |

---

## 4. Alternatives Considered

### Option A: Polyglot monorepo (Go + Elixir + TS in a single repository)
- **Pros**: `.proto` contracts as a single source of truth with no package publishing; atomic cross-service refactors; unified CI with `buf breaking`.
- **Cons**: more complex pipelines (distinct toolchains); risk of undue coupling between services.
- **Decision**: **Chosen** — in the foundation phase, iteration speed on the contracts dominates; a future split by repository remains possible since services only know each other through the protos.

### Option B: Multi-repo per service from the start
- **Pros**: full lifecycle isolation (aligns with the M pillar of MACH).
- **Cons**: requires a proto package registry (BSR), versioning, and contract synchronization before the first service even exists.
- **Decision**: Discarded at this stage; to be reassessed once the `v1` contracts stabilize.

### Option C: Direct collaborative persistence (every mutation → PostgreSQL)
- **Pros**: simplicity, no Redis or stateful GenServer.
- **Cons**: write exhaustion from UI micro-movements; contradicts BR06 and `doc/GATEWAY_COLLABORATION.md`.
- **Decision**: Discarded — write-behind with a 5s debounce is a documented requirement.

### Option D: Kafka instead of RabbitMQ
- **Pros**: event replay, higher throughput, native per-tenant partitioning.
- **Cons**: the lag-based KEDA scaler is more complex; the documented topology (exchanges, dynamic routing, DLQ, fair queuing) is idiomatic to RabbitMQ; no replay requirement.
- **Decision**: Discarded — the architecture docs pin down RabbitMQ (`doc/ASYNC_OBSERVABILITY.md`).

### Option E: Schema-per-tenant instead of a shared database
- **Pros**: stronger physical isolation.
- **Cons**: RAM/CPU cost and per-tenant migrations; contradicts the explicit decision for Shared Database for cost efficiency (pillar C).
- **Decision**: Discarded — shared database + `tenant_id` + RLS is the documented architecture.
