# USAGE — How to run MACH V4 locally

Monorepo startup guide. Order: **infra → proto → Go services → Gateway → Collab → Frontend**.

All build/startup/deploy scripts for the repo live in **`build/`**.

---

## Guided startup (recommended)

```bash
./build/dev-up.sh              # brings everything up, with checks and confirmation prompts
./build/dev-up.sh --no-frontend  # brings everything up except the frontend (e.g., you already run Vite in another terminal)
./build/dev-up.sh --yes        # doesn't ask anything, assumes "yes" for every prompt
```

What it does, in order, with visual feedback (✓/✗/!) at each step:

1. **Pre-checks** — confirms `docker`, `go`, `node`, `npm`, `mix`, `buf` are on the PATH and checks the Go version; automatically adjusts the PATH for the local toolchains (`$HOME/.local/go`, `$HOME/.local/elixir1.17`).
2. **Infra** (`make up` + `make migrate`) — warns and asks for confirmation if any port is already in use (e.g., MinIO 9000 used by another project) before proceeding.
3. **Proto** (`make proto`) — regenerates `gen/go`, `gen/elixir`, `gen/ts`.
4. **gRPC Services** (iam, design, logic, deploy, export) — start up in background, with an active wait until each port responds.
5. **Workers** (RabbitMQ).
6. **Gateway** (`:8080`).
7. **Collab** (Phoenix, `:4000`).
8. **Frontend** — installs deps if missing and asks whether to open now (`npm run dev`, foreground).

At the end, it shows a summary with the URLs of each service. **Ctrl+C stops all processes started by the script.**

**Logs**: everything — including `make up`, `make proto`, `npm install`, `mix deps.get` — is recorded to `.dev-logs/<name>.log` (a single, gitignored folder), in addition to appearing on screen.

---

## Manual step-by-step

Use this if you prefer running each piece by hand, or to debug a specific step reported as failed by `dev-up.sh`.

### 0. Toolchain prerequisites

The system's apt is too old for the repo's deps — use the locally installed versions:

```bash
# Go 1.26 (apt has 1.22, insufficient for minio-go/x-net/protobuf)
export PATH="$HOME/.local/go/bin:$PATH"

# Elixir 1.17.3 / OTP 25 (apt has 1.14, insufficient for Phoenix/Plug/Bandit)
export PATH="$HOME/.local/elixir1.17/bin:$PATH"
export MIX_HOME="$HOME/.mix"
export HEX_HOME="$HOME/.hex"
```

Other dependencies: Docker + Docker Compose, Node 20, `buf` (`make tools` installs it into `$(go env GOPATH)/bin`).

### 1. Bring up the infrastructure (Docker Compose)

```bash
make up        # postgres:5432, redis:6379, rabbitmq:5672/15672, jaeger:16686, otel-collector:4317/4318, minio:9000/9001
make migrate   # applies infra/postgres/migrations/*.sql
```

> **Port gotcha**: compose's `minio` uses host port **9000/9001**. If another project already occupies that port, spin up a standalone MinIO (`docker run -p 9010:9000 ...` with creds `mach`/`machsecret`) and point `S3_ENDPOINT=localhost:9010` in the envs below.

### 2. Generate the proto stubs

Required before compiling Go and Elixir (`collab` compiles `gen/elixir`, which is gitignored):

```bash
make proto     # buf lint + buf generate → gen/go, gen/elixir, gen/ts
```

Requires `buf` (`make tools`) and, for the Elixir target, `protoc-gen-elixir` on the PATH:

```bash
mix escript.install hex protobuf   # once, with Elixir 1.17 already on the PATH
export PATH="$HOME/.mix/escripts:$PATH"
```

### 3. Bring up the gRPC services (Go)

Each service in its own terminal, from the repo root. All of them read `DATABASE_URL` and `OTEL_EXPORTER_OTLP_ENDPOINT` with defaults already pointing at the compose infra — you normally don't need to set anything:

```bash
go run ./services/iam/cmd      # IAM      :50051
go run ./services/design/cmd   # Design   :50052
go run ./services/logic/cmd    # Logic    :50053 (uses RABBITMQ_URL)
go run ./services/deploy/cmd   # Deploy   :50054
go run ./services/export/cmd   # Export   :50055 (uses S3_ENDPOINT/S3_ACCESS_KEY/S3_SECRET_KEY/S3_BUCKET)
```

#### Asynchronous workers (RabbitMQ)

```bash
go run ./services/workers/cmd  # consumes queues via RABBITMQ_URL (default amqp://mach:mach@localhost:5672/)
```

### 4. Bring up the HTTP Gateway

```bash
go run ./services/gateway/cmd  # :8080
```

Reads the addresses of the services above via env (defaults are already correct for local):
`GATEWAY_HTTP_ADDR`, `IAM_GRPC_ADDR`, `DESIGN_GRPC_ADDR`, `LOGIC_GRPC_ADDR`, `DEPLOY_GRPC_ADDR`, `EXPORT_GRPC_ADDR`, `OTEL_EXPORTER_OTLP_ENDPOINT`.

For social login in dev, optionally: `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`, `OAUTH_ALLOWED_REDIRECT_URIS`.

### 5. Bring up Collab (Elixir/Phoenix — real-time collaboration)

```bash
cd services/collab
mix deps.get
mix phx.server     # http://localhost:4000
```

> `jose` is pinned to `1.11.5` in `mix.exs` (compat with OTP 25).

### 6. Bring up the Frontend (Vite/React)

```bash
cd services/frontend
npm install         # if node_modules doesn't exist yet
npm run dev
```

Config in `services/frontend/.env.local` (`VITE_BYPASS_AUTH=true` skips auth in dev). Runtime config injected via `window.__FRONTEND_CONFIG__` (host's baseUrl/token/sistemaId).

Port fixed at `5183` via `server.port` + `server.strictPort: true` in `vite.config.ts` — without `strictPort`, Vite silently falls back to `5173` if the port is busy. Visual editor (Screens tab — canvas, rich text, free positioning, component catalog) documented in `specs/007-editor-visual-canvas/`.

---

## Port quick reference

| Service          | Port   |
|-------------------|---------|
| Postgres           | 5432    |
| Redis              | 6379    |
| RabbitMQ (AMQP)    | 5672    |
| RabbitMQ (mgmt UI) | 15672   |
| Jaeger UI          | 16686   |
| OTel Collector     | 4317 (gRPC) / 4318 (HTTP) |
| MinIO (S3 API)     | 9000    |
| MinIO (console)    | 9001    |
| IAM Service        | 50051   |
| Design Service     | 50052   |
| Logic Service      | 50053   |
| Deploy Service     | 50054   |
| Export Service     | 50055   |
| Gateway (HTTP)     | 8080    |
| Collab (Phoenix)   | 4000    |
| Frontend (Vite dev)| 5183    |

## Useful commands (Makefile)

```bash
make help            # lists all targets
make down             # tears down the compose infra
make test             # go test ./...
make tidy             # go mod tidy
make proto-breaking   # buf breaking --against main
```

## Build and deploy (CI/CD, spec 002)

Also in `build/`, used by the `.github/workflows/cd.yml` pipeline (and reusable locally to rehearse a release):

```bash
SHA=$(git rev-parse --short HEAD) build/build-artifacts.sh          # packages the 7 binaries + Elixir release + services/frontend/dist into dist/artifacts/
build/deploy.sh --env staging --host <host> --user <user> --sha <sha>   # rsync + atomic symlink swap + restart
build/smoke-test.sh --host <host>                                    # post-deploy healthcheck
build/rollback.sh --env staging --host <host> [--sha <sha>]          # point current back to the previous release, without a rebuild
```

Full details in `specs/002-entrega-continua-artefatos/` and `infra/deploy/README.md`.

## Integration / E2E tests

Require the compose infra to be up (`make up` + `make migrate`) and run serially (`-p 1`) to avoid a race on `GRANT ... ON SCHEMA`:

```bash
DATABASE_URL=postgres://mach:mach@localhost:5432/machv4?sslmode=disable \
  go test -tags integration -p 1 ./...

# Tracing E2E needs rabbitmq/jaeger/otel-collector running (via compose)
OTLP_ENDPOINT=localhost:4317 JAEGER_QUERY=localhost:16686 \
  RABBITMQ_URL=amqp://mach:mach@localhost:5672/ \
  go test -tags e2e ./tests/e2e/...
```

Frontend tests: `cd services/frontend && npm test` (vitest) and `npm run typecheck` (`tsc --noEmit`).
