# Research: Resource Monitor

---

## 1. Existing Patterns in the Project

| File/Pattern | Location | Relevance |
|----------------|-------------|-----------|
| `NewServer(pool)` + a public `app.go` | `services/design/app/app.go`, `services/deploy/app/app.go` | Direct model for `services/monitor/app/app.go` (no pool — the monitor doesn't use Postgres). |
| `main.go` with an `env()` helper + `telemetry.Init` + `grpc.NewServer(grpc.StatsHandler(otelgrpc...), grpc.ChainUnaryInterceptor(tenantctx...))` | `services/design/cmd/main.go`, `services/iam/cmd/main.go` | Model for `services/monitor/cmd/main.go`. Note: the Monitor doesn't deal with `TenantContext` (it isn't multi-tenant) — re-evaluate whether the `tenantctx` interceptors make sense here (probably not, see §3). |
| Simple `GET /health` (200 with no body) | `services/gateway/internal/app/router.go:28` | The already existing minimal liveness pattern; the Workers' new `/health` follows the same simplicity, just with a JSON body. |
| `plug :healthz` responding before the router | `services/collab/lib/collab_web/endpoint.ex:48-52` | The exact extension point to include uptime/memory in Collab (FR02). |
| `routes.ResumoFinanceiro(iam)` → `http.HandlerFunc` | `services/gateway/internal/routes/*.go` | Model for `routes.ObterRecursos(monitor)`. |
| `NewRouter(iam, design, logic, deploy, export, rl, oauth)` taking one gRPC client per service | `services/gateway/internal/app/router.go` | Model for adding the 6th client (`monitor`). |
| `useResumoFinanceiro.ts` (`carregando`/`pronto`/`erro` states + `recarregar`) | `services/frontend/src/dashboard/useResumoFinanceiro.ts` | Model for `useRecursos.ts`, plus `setInterval` for auto-refresh (FR07). |
| `CardResumoFinanceiro.tsx` | `services/frontend/src/dashboard/CardResumoFinanceiro.tsx` | Visual/structural model for `CardServicoStatus.tsx`. |
| Nested routes in `App.tsx` inside `/dashboard` + a sidebar item in `DashboardLayout.tsx` | `services/frontend/src/App.tsx:103-114`, `.../DashboardLayout.tsx:76-95` | Where to plug in the Monitor route/nav (FR08). |
| `run_bg <name> go run ./services/<name>/cmd` | `build/dev-up.sh:230-259` | Where to add the Monitor's boot in the local dev flow. |
| `pkg/telemetry`, `pkg/tenantctx`, `pkg/database`, `pkg/eventbus`, `pkg/blindindex` | `pkg/` | Direct precedent for creating `pkg/health` as a new shared package. |
| Services' gRPC ports (project memory, confirmed in code) | IAM `:50051`, Design `:50052`, Logic `:50053`, Deploy `:50054`, Export `:50055`, Gateway HTTP `:8080`, Collab HTTP `:4000` | Baseline for deciding the new ports (§2). |

---

## 2. Ports and Addresses — final decision

| Service | Address/port | Note |
|---------|-----------------|------------|
| Monitor (gRPC, new) | `:50056` (`MONITOR_GRPC_ADDR`) | Next free port in the already-used 50051-50055 gRPC range. |
| Workers (HTTP, new) | `:8081` (`WORKERS_HTTP_ADDR`) | Sits in the HTTP range alongside the Gateway (`:8080`), not the gRPC range — avoids the collision with the Monitor's `:50056` identified in `plan.md` §6, and makes it clear it's an HTTP endpoint, not gRPC. |
| Collab `/healthz` | `:4000` (already existing, `PHX_PORT`/hardcoded in `dev.exs`) | No new port — only the response body changes (FR02). |
| Gateway `/health` | `:8080` (already existing) | No change — the Monitor only does `GET http://localhost:8080/health` and converts "200 OK" into `ServicoStatus{status: "servindo"}` (no uptime/memory for the Gateway in this delivery, since `/health` doesn't return a body — see §3, accepted as a documented limitation). |

All 8 addresses (5 gRPC + Gateway HTTP + Collab HTTP + Workers HTTP) are
configurable via env var in the Monitor's `main.go`, following exactly the same
pattern already used in `services/gateway/cmd/main.go:37-41`.

---

## 3. Architecture/Design Alternatives Considered

### Option A: Standardize everything on HTTP (every service gains a `net/http` with `/health`)
- **Pros**: a single collector type in the Monitor, no need for two protos.
- **Cons**: forces IAM/Design/Logic/Deploy/Export — pure gRPC today — to each stand up a
  second listener just for this; more new surface per service than extending the
  `grpc.Server` each of them already has.
- **Decision**: Discarded. See the full rationale in `plan.md` §1.

### Option B: Prometheus + exporters, Monitor just reads Prometheus
- **Pros**: "real" metrics infrastructure, time series, alerts ready for the future.
- **Cons**: the user's explicit decision for this delivery was not to use Prometheus
  (infra cost + additional learning curve not justified for a first status screen); the
  already existing OTel Collector is for traces, not metrics, and adapting it would cost
  about as much as standing up Prometheus from scratch.
- **Decision**: Discarded for this delivery — documented as a possible future evolution,
  not as "forgotten out of scope."

### Option C: The Gateway's `GET /health` returns a JSON body with the Gateway's own uptime/memory
- **Pros**: the Monitor would have complete data for the Gateway, not just "up/down."
- **Cons**: would change the contract of an already-in-use public endpoint (used by
  liveness/external orchestration, if any); expanding this delivery to edit
  `router.go:28` goes beyond the minimal scope agreed with the user.
- **Decision**: Discarded for this round — accepted as a documented limitation
  (`spec.md` BR02: "each service reports what it can"); the Gateway shows up on the
  screen as up/down with no memory metrics. Could become a one-line extension in a
  future request (`GET /health` starts accepting `Accept: application/json` without
  breaking the current consumer that only looks at the status code).

### Option D: The Monitor uses `tenantctx` interceptors like the other gRPC services
- **Pros**: full consistency with the other `main.go` files' pattern.
- **Cons**: `tenantctx` exists to propagate/validate the tenant of a multi-tenant
  business request; the Monitor's RPC (`ObterRecursos`) has no tenant — it's a
  platform-wide operational query, called by the Gateway with no `TenantContext`.
- **Decision**: Discarded — `services/monitor` registers `otelgrpc` (tracing, NFR03) but
  **does not** chain `tenantctx.UnaryServerInterceptor()`. Document this exception in a
  `main.go` comment so it doesn't look like an accidental omission on the next review.
