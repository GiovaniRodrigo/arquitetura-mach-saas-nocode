# Tarefas: Monitor de Recursos

<!-- Ordenadas por dependência de execução. TDD: teste antes da implementação em cada
camada, seguindo o padrão já confirmado no repo (specs 001/003). -->

> **Nota de pivô de arquitetura (não documentada em spec própria, só em
> comentários de código — "spec 009"):** o desenho original das tasks 14-20
> (microsserviço `services/monitor` fazendo polling gRPC/HTTP de cada serviço)
> foi abandonado em favor de ler CPU/memória do `metrics-server` do Kubernetes
> e RPS/latência/taxa de sucesso do Prometheus do `linkerd-viz` diretamente no
> Gateway (`services/gateway/internal/meshmetrics`). As tasks 1-13
> (proto `health`/`monitor` + `RecursosService` exposto por cada serviço)
> seguem implementadas e úteis como health-check/probe de cada serviço, mas
> **não são mais consumidas pelo Monitor** — nenhum serviço faz mais polling
> nelas. Se este pivô virar definitivo, vale abrir uma spec 009 própria
> documentando a decisão em vez de deixá-la só em comentários.

## Proto e pacote compartilhado (base de tudo)

- [x] 1. Criar `proto/construtor/health/v1/health.proto` e `proto/construtor/monitor/v1/monitor.proto` conforme `contracts/interfaces.md`; rodar `make proto` e confirmar geração em `gen/go`, `gen/ts` sem erro (`buf lint`). (`proto/construtor/health/v1/health.proto`, `proto/construtor/monitor/v1/monitor.proto`)
- [x] 2. Escrever `pkg/health/server_test.go` cobrindo: `ObterStatus` retorna `status="servindo"`, `uptime_segundos >= 0`, `memoria_alocada_bytes > 0`. (`pkg/health/server_test.go`)
- [x] 3. Implementar `pkg/health/server.go` (`RecursosServiceServer`) e `pkg/health/registrar.go` (`Registrar(grpcServer, nome, iniciadoEm)`) até o teste da task 2 passar. (`pkg/health/server.go`, `pkg/health/registrar.go`)

## Serviços Go existentes — expor RecursosService (RF01)

- [x] 4. Adicionar `health.Registrar(grpcServer, "iam", inicio)` em `services/iam/cmd/main.go`; validar com `go build ./services/iam/...`. (`services/iam/cmd/main.go`)
- [x] 5. Idem para `services/design/cmd/main.go` ("design"). (`services/design/cmd/main.go`)
- [x] 6. Idem para `services/logic/cmd/main.go` ("logic"). (`services/logic/cmd/main.go`)
- [x] 7. Idem para `services/deploy/cmd/main.go` ("deploy"). (`services/deploy/cmd/main.go`)
- [x] 8. Idem para `services/export/cmd/main.go` ("export"). (`services/export/cmd/main.go`)

## Workers — endpoint HTTP novo (RF03)

- [x] 9. Escrever `services/workers/internal/health/server_test.go`: `httptest` confirma que `GET /health` retorna 200 e o JSON de `contracts/api.md` (status/uptime/memória/goroutines). (`services/workers/internal/health/server_test.go`)
- [x] 10. Implementar `services/workers/internal/health/server.go` até o teste da task 9 passar. (`services/workers/internal/health/server.go`)
- [x] 11. Subir o servidor da task 10 em `services/workers/cmd/main.go` numa goroutine, endereço `env("WORKERS_HTTP_ADDR", ":8081")` (porta decidida em `research.md` §2), com shutdown gracioso junto ao `signal.Notify` já existente. (`services/workers/cmd/main.go`)

## Collab — estender /healthz (RF02)

- [x] 12. Escrever/estender `services/collab/test/collab_web/endpoint_test.exs`: `GET /healthz` retorna 200 com `status`, `uptime_segundos`, `memoria_alocada_bytes` no corpo. (`services/collab/test/collab_web/endpoint_test.exs`)
- [x] 13. Estender a função `healthz/2` em `services/collab/lib/collab_web/endpoint.ex` até o teste da task 12 passar (`:erlang.memory(:total)` + uptime desde o boot). (`services/collab/lib/collab_web/endpoint.ex`)

## Serviço Monitor — coletores (RF04, depende das tasks 1-13)

> **Abandonado pelo pivô de arquitetura acima — ver nota no topo do arquivo.**
> `services/monitor/` nunca chegou a existir; o Gateway lê os dados direto do
> metrics-server/Prometheus via `internal/meshmetrics`.

- [ ] ~~14. Escrever `services/monitor/internal/poller/coletor_test.go`: `ColetorGRPC` contra um `RecursosService` fake (bufconn) e `ColetorHTTP` contra um `httptest.Server` fake, cobrindo sucesso e timeout/erro de rede.~~ (`services/monitor/internal/poller/coletor_test.go`)
- [ ] ~~15. Implementar `services/monitor/internal/poller/coletor.go` (interface `Coletor`), `coletor_grpc.go`, `coletor_http.go` até os testes da task 14 passarem.~~ (`services/monitor/internal/poller/coletor.go`, `coletor_grpc.go`, `coletor_http.go`)
- [ ] ~~16. Escrever `services/monitor/internal/poller/agregador_test.go`: N coletores fake (alguns ok, um sempre erro, um que trava além do timeout) — confirma RN01 (erro individual não propaga) e RN04/RNF01 (tempo total ≈ maior tempo individual, não soma).~~ (`services/monitor/internal/poller/agregador_test.go`)
- [ ] ~~17. Implementar `services/monitor/internal/poller/agregador.go` (fan-out/fan-in, `plan.md` §4.2) até os testes da task 16 passarem.~~ (`services/monitor/internal/poller/agregador.go`)

## Serviço Monitor — gRPC server e binário

> **Abandonado — mesmo motivo das tasks 14-17.**

- [ ] ~~18. Escrever `services/monitor/internal/server/grpc_test.go`: `ObterRecursos` retorna `ObterRecursosResponse` com 1 entrada por coletor fake injetado, ordem preservada.~~ (`services/monitor/internal/server/grpc_test.go`)
- [ ] ~~19. Implementar `services/monitor/internal/server/grpc.go` (`MonitorServer.ObterRecursos`) até o teste da task 18 passar.~~ (`services/monitor/internal/server/grpc.go`)
- [ ] ~~20. Implementar `services/monitor/app/app.go` (monta `MonitorServer` a partir da lista fixa dos 8 coletores) e `services/monitor/cmd/main.go` (env vars dos 8 endereços, `MONITOR_GRPC_ADDR=:50056`, `telemetry.Init`, sem `tenantctx` — ver `research.md` Opção D).~~ (`services/monitor/app/app.go`, `services/monitor/cmd/main.go`)

## Gateway — fachada REST (RF05)

> Implementado sob o desenho novo: `ObterRecursos(recursosClient)` recebe um
> `*meshmetrics.Client` (não um `monitorv1.MonitorServiceClient` como previsto
> originalmente), que fala com o metrics-server/Prometheus do cluster.

- [x] 21. ~~Escrever `services/gateway/internal/routes/monitor_test.go`: `ObterRecursos(monitor)` serializa a resposta gRPC como JSON 200; erro do client gRPC vira erro HTTP único (RNF02).~~ Testado contra `recursosClient` (interface satisfeita por `*meshmetrics.Client`), mesma garantia de RNF02. (`services/gateway/internal/routes/monitor_test.go`)
- [x] 22. Implementar `services/gateway/internal/routes/monitor.go` até o teste da task 21 passar. (`services/gateway/internal/routes/monitor.go`)
- [x] 23. ~~Adicionar parâmetro `monitor monitorv1.MonitorServiceClient` a `NewRouter`~~ — adicionado `recursos *meshmetrics.Client` a `NewRouter` e a rota `GET /api/v1/monitor/recursos` (grupo autenticado, RNF05) em `services/gateway/internal/app/router.go`; `services/gateway/cmd/main.go` monta o client via `meshmetrics.NewK8sClient`/`NewPrometheusClient` (degrada para 502 se não estiver rodando no cluster, sem MONITOR_GRPC_ADDR). (`services/gateway/internal/app/router.go`, `services/gateway/cmd/main.go`)

## Frontend (RF06, RF07, RF08 — depende da task 23)

- [x] 24. Adicionar tipo `ServicoStatus` em `services/frontend/src/api/types.ts` e método `obterRecursos()` em `services/frontend/src/api/client.ts` (mesmo padrão de `resumoFinanceiro()`). (`services/frontend/src/api/types.ts`, `services/frontend/src/api/client.ts`)
- [x] 25. Escrever `services/frontend/src/dashboard/useRecursos.test.ts`: estados `carregando`/`pronto`/`erro`, `recarregar()` manual, auto-refresh dispara nova chamada após o intervalo configurado. (`services/frontend/src/dashboard/useRecursos.test.ts`)
- [x] 26. Implementar `services/frontend/src/dashboard/useRecursos.ts` até o teste da task 25 passar. (`services/frontend/src/dashboard/useRecursos.ts`)
- [x] 27. Escrever `services/frontend/src/dashboard/CardServicoStatus.test.tsx`: renderiza nome/uptime/memória formatados quando `status="servindo"`; renderiza indicador vermelho + mensagem de erro quando `status="indisponivel"`. (`services/frontend/src/dashboard/CardServicoStatus.test.tsx`)
- [x] 28. Implementar `services/frontend/src/dashboard/CardServicoStatus.tsx` até o teste da task 27 passar. (`services/frontend/src/dashboard/CardServicoStatus.tsx`)
- [x] 29. Escrever `services/frontend/src/pages/Dashboard/Monitor.test.tsx`: renderiza 8 cards a partir do hook; botão "Atualizar" chama `recarregar()`; estado de erro-da-tela-toda (RNF02) não renderiza cards. (`services/frontend/src/pages/Dashboard/Monitor.test.tsx`)
- [x] 30. Implementar `services/frontend/src/pages/Dashboard/Monitor.tsx` até o teste da task 29 passar. (`services/frontend/src/pages/Dashboard/Monitor.tsx`)
- [x] 31. Adicionar rota `monitor` em `services/frontend/src/App.tsx` (dentro de `/dashboard`) e item de sidebar "Monitor" + `case` em `tituloDaPagina` em `services/frontend/src/layout/DashboardLayout.tsx`, cobrindo o novo item em `DashboardLayout.test.tsx`. (`services/frontend/src/App.tsx`, `services/frontend/src/layout/DashboardLayout.tsx`, `services/frontend/src/layout/DashboardLayout.test.tsx`)

## Infra e integração final

- [ ] ~~32. Adicionar `run_bg monitor go run ./services/monitor/cmd` a `build/dev-up.sh`, na ordem correta (depois dos serviços que ele consulta).~~ Não se aplica — não há mais binário `monitor`; `build/dev-up.sh` só ganhou uma nota de documentação apontando `/dashboard/monitor` (CPU/memória via metrics-server, exige rodar dentro de um cluster com o mesh — fora do escopo do dev local). (`build/dev-up.sh`)
- [x] 33 (parcial). Suíte automatizada rodada e verde nesta working tree: `go build ./... && go vet -tags integration ./... && go test ./...` (raiz, 2026-08-21), `mix test` (`services/collab`, 29 testes), `npm run typecheck && npx vitest run` (`services/frontend`, 376 testes). **Pendente**: validação manual dos 4 cenários de `spec.md` §5 com `build/dev-up.sh` de pé — precisa rodar dentro de um cluster Kubernetes com Linkerd/Prometheus (o `meshmetrics.NewK8sClient` só funciona in-cluster), não reproduzível neste ambiente sandbox.
