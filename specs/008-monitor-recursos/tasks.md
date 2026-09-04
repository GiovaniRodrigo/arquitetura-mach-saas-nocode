# Tasks: Resource Monitor

<!-- Ordered by execution dependency. TDD: test before implementation at each
layer, following the pattern already confirmed in the repo (specs 001/003). -->

> **Architecture-pivot note (not documented in its own spec, only in
> code comments — "spec 009"):** the original design for tasks 14-20
> (the `services/monitor` microservice doing gRPC/HTTP polling of each service)
> was abandoned in favor of reading CPU/memory from Kubernetes'
> `metrics-server` and RPS/latency/success-rate from `linkerd-viz`'s
> Prometheus directly in the Gateway (`services/gateway/internal/meshmetrics`).
> Tasks 1-13 (the `health`/`monitor` protos + `RecursosService` exposed by
> each service) remain implemented and useful as a health-check/probe for
> each service, but **are no longer consumed by the Monitor** — no service
> polls them anymore. If this pivot becomes permanent, it's worth opening a
> dedicated spec 009 documenting the decision instead of leaving it only in
> comments.

## Proto and shared package (foundation for everything)

- [x] 1. Create `proto/construtor/health/v1/health.proto` and `proto/construtor/monitor/v1/monitor.proto` per `contracts/interfaces.md`; run `make proto` and confirm generation in `gen/go`, `gen/ts` with no error (`buf lint`). (`proto/construtor/health/v1/health.proto`, `proto/construtor/monitor/v1/monitor.proto`)
- [x] 2. Write `pkg/health/server_test.go` covering: `ObterStatus` returns `status="servindo"`, `uptime_segundos >= 0`, `memoria_alocada_bytes > 0`. (`pkg/health/server_test.go`)
- [x] 3. Implement `pkg/health/server.go` (`RecursosServiceServer`) and `pkg/health/registrar.go` (`Registrar(grpcServer, nome, iniciadoEm)`) until task 2's test passes. (`pkg/health/server.go`, `pkg/health/registrar.go`)

## Existing Go services — expose RecursosService (FR01)

- [x] 4. Add `health.Registrar(grpcServer, "iam", inicio)` in `services/iam/cmd/main.go`; validate with `go build ./services/iam/...`. (`services/iam/cmd/main.go`)
- [x] 5. Same for `services/design/cmd/main.go` ("design"). (`services/design/cmd/main.go`)
- [x] 6. Same for `services/logic/cmd/main.go` ("logic"). (`services/logic/cmd/main.go`)
- [x] 7. Same for `services/deploy/cmd/main.go` ("deploy"). (`services/deploy/cmd/main.go`)
- [x] 8. Same for `services/export/cmd/main.go` ("export"). (`services/export/cmd/main.go`)

## Workers — new HTTP endpoint (FR03)

- [x] 9. Write `services/workers/internal/health/server_test.go`: `httptest` confirms that `GET /health` returns 200 and the JSON from `contracts/api.md` (status/uptime/memory/goroutines). (`services/workers/internal/health/server_test.go`)
- [x] 10. Implement `services/workers/internal/health/server.go` until task 9's test passes. (`services/workers/internal/health/server.go`)
- [x] 11. Start task 10's server in `services/workers/cmd/main.go` in a goroutine, address `env("WORKERS_HTTP_ADDR", ":8081")` (port decided in `research.md` §2), with graceful shutdown alongside the already existing `signal.Notify`. (`services/workers/cmd/main.go`)

## Collab — extend /healthz (FR02)

- [x] 12. Write/extend `services/collab/test/collab_web/endpoint_test.exs`: `GET /healthz` returns 200 with `status`, `uptime_segundos`, `memoria_alocada_bytes` in the body. (`services/collab/test/collab_web/endpoint_test.exs`)
- [x] 13. Extend the `healthz/2` function in `services/collab/lib/collab_web/endpoint.ex` until task 12's test passes (`:erlang.memory(:total)` + uptime since boot). (`services/collab/lib/collab_web/endpoint.ex`)

## Monitor service — collectors (FR04, depends on tasks 1-13)

> **Abandoned by the architecture pivot above — see the note at the top of the file.**
> `services/monitor/` never came to exist; the Gateway reads the data directly from
> the metrics-server/Prometheus via `internal/meshmetrics`.

- [ ] ~~14. Write `services/monitor/internal/poller/coletor_test.go`: `ColetorGRPC` against a fake `RecursosService` (bufconn) and `ColetorHTTP` against a fake `httptest.Server`, covering success and timeout/network error.~~ (`services/monitor/internal/poller/coletor_test.go`)
- [ ] ~~15. Implement `services/monitor/internal/poller/coletor.go` (the `Coletor` interface), `coletor_grpc.go`, `coletor_http.go` until task 14's tests pass.~~ (`services/monitor/internal/poller/coletor.go`, `coletor_grpc.go`, `coletor_http.go`)
- [ ] ~~16. Write `services/monitor/internal/poller/agregador_test.go`: N fake collectors (some ok, one always erroring, one that hangs past the timeout) — confirms BR01 (an individual error doesn't propagate) and BR04/NFR01 (total time ≈ the largest individual time, not the sum).~~ (`services/monitor/internal/poller/agregador_test.go`)
- [ ] ~~17. Implement `services/monitor/internal/poller/agregador.go` (fan-out/fan-in, `plan.md` §4.2) until task 16's tests pass.~~ (`services/monitor/internal/poller/agregador.go`)

## Monitor service — gRPC server and binary

> **Abandoned — same reason as tasks 14-17.**

- [ ] ~~18. Write `services/monitor/internal/server/grpc_test.go`: `ObterRecursos` returns an `ObterRecursosResponse` with 1 entry per injected fake collector, order preserved.~~ (`services/monitor/internal/server/grpc_test.go`)
- [ ] ~~19. Implement `services/monitor/internal/server/grpc.go` (`MonitorServer.ObterRecursos`) until task 18's test passes.~~ (`services/monitor/internal/server/grpc.go`)
- [ ] ~~20. Implement `services/monitor/app/app.go` (assembles `MonitorServer` from the fixed list of 8 collectors) and `services/monitor/cmd/main.go` (env vars for the 8 addresses, `MONITOR_GRPC_ADDR=:50056`, `telemetry.Init`, no `tenantctx` — see `research.md` Option D).~~ (`services/monitor/app/app.go`, `services/monitor/cmd/main.go`)

## Gateway — REST facade (FR05)

> Implemented under the new design: `ObterRecursos(recursosClient)` takes a
> `*meshmetrics.Client` (not a `monitorv1.MonitorServiceClient` as originally
> planned), which talks to the cluster's metrics-server/Prometheus.

- [x] 21. ~~Write `services/gateway/internal/routes/monitor_test.go`: `ObterRecursos(monitor)` serializes the gRPC response as JSON 200; a gRPC client error becomes a single HTTP error (NFR02).~~ Tested against `recursosClient` (an interface satisfied by `*meshmetrics.Client`), the same NFR02 guarantee. (`services/gateway/internal/routes/monitor_test.go`)
- [x] 22. Implement `services/gateway/internal/routes/monitor.go` until task 21's test passes. (`services/gateway/internal/routes/monitor.go`)
- [x] 23. ~~Add a `monitor monitorv1.MonitorServiceClient` parameter to `NewRouter`~~ — added `recursos *meshmetrics.Client` to `NewRouter` and the route `GET /api/v1/monitor/recursos` (authenticated group, NFR05) in `services/gateway/internal/app/router.go`; `services/gateway/cmd/main.go` assembles the client via `meshmetrics.NewK8sClient`/`NewPrometheusClient` (degrades to 502 if not running in the cluster, no MONITOR_GRPC_ADDR). (`services/gateway/internal/app/router.go`, `services/gateway/cmd/main.go`)

## Frontend (FR06, FR07, FR08 — depends on task 23)

- [x] 24. Add the `ServicoStatus` type in `services/frontend/src/api/types.ts` and the `obterRecursos()` method in `services/frontend/src/api/client.ts` (same pattern as `resumoFinanceiro()`). (`services/frontend/src/api/types.ts`, `services/frontend/src/api/client.ts`)
- [x] 25. Write `services/frontend/src/dashboard/useRecursos.test.ts`: `carregando`/`pronto`/`erro` states, manual `recarregar()`, auto-refresh triggers a new call after the configured interval. (`services/frontend/src/dashboard/useRecursos.test.ts`)
- [x] 26. Implement `services/frontend/src/dashboard/useRecursos.ts` until task 25's test passes. (`services/frontend/src/dashboard/useRecursos.ts`)
- [x] 27. Write `services/frontend/src/dashboard/CardServicoStatus.test.tsx`: renders formatted name/uptime/memory when `status="servindo"`; renders a red indicator + error message when `status="indisponivel"`. (`services/frontend/src/dashboard/CardServicoStatus.test.tsx`)
- [x] 28. Implement `services/frontend/src/dashboard/CardServicoStatus.tsx` until task 27's test passes. (`services/frontend/src/dashboard/CardServicoStatus.tsx`)
- [x] 29. Write `services/frontend/src/pages/Dashboard/Monitor.test.tsx`: renders 8 cards from the hook; the "Update" button calls `recarregar()`; the whole-screen error state (NFR02) doesn't render cards. (`services/frontend/src/pages/Dashboard/Monitor.test.tsx`)
- [x] 30. Implement `services/frontend/src/pages/Dashboard/Monitor.tsx` until task 29's test passes. (`services/frontend/src/pages/Dashboard/Monitor.tsx`)
- [x] 31. Add the `monitor` route in `services/frontend/src/App.tsx` (inside `/dashboard`) and a "Monitor" sidebar item + `case` in `tituloDaPagina` in `services/frontend/src/layout/DashboardLayout.tsx`, covering the new item in `DashboardLayout.test.tsx`. (`services/frontend/src/App.tsx`, `services/frontend/src/layout/DashboardLayout.tsx`, `services/frontend/src/layout/DashboardLayout.test.tsx`)

## Infra and final integration

- [ ] ~~32. Add `run_bg monitor go run ./services/monitor/cmd` to `build/dev-up.sh`, in the right order (after the services it queries).~~ Not applicable — there is no more `monitor` binary; `build/dev-up.sh` only gained a documentation note pointing to `/dashboard/monitor` (CPU/memory via metrics-server, requires running inside a cluster with the mesh — out of scope for local dev). (`build/dev-up.sh`)
- [x] 33 (partial). Automated suite run and green in this working tree: `go build ./... && go vet -tags integration ./... && go test ./...` (root, 2026-08-21), `mix test` (`services/collab`, 29 tests), `npm run typecheck && npx vitest run` (`services/frontend`, 376 tests). **Pending**: manual validation of the 4 scenarios from `spec.md` §5 with `build/dev-up.sh` up — needs to run inside a Kubernetes cluster with Linkerd/Prometheus (`meshmetrics.NewK8sClient` only works in-cluster), not reproducible in this sandbox environment.
