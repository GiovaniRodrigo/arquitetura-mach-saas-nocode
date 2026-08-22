# Pesquisa: Monitor de Recursos

---

## 1. Padrões Existentes no Projeto

| Arquivo/Padrão | Localização | Relevância |
|----------------|-------------|-----------|
| `NewServer(pool)` + `app.go` público | `services/design/app/app.go`, `services/deploy/app/app.go` | Modelo direto para `services/monitor/app/app.go` (sem pool — monitor não usa Postgres). |
| `main.go` com `env()` helper + `telemetry.Init` + `grpc.NewServer(grpc.StatsHandler(otelgrpc...), grpc.ChainUnaryInterceptor(tenantctx...))` | `services/design/cmd/main.go`, `services/iam/cmd/main.go` | Modelo para `services/monitor/cmd/main.go`. Nota: Monitor não lida com `TenantContext` (não é multi-tenant) — reavaliar se `tenantctx` interceptors fazem sentido aqui (provavelmente não, ver §3). |
| `GET /health` simples (200 sem corpo) | `services/gateway/internal/app/router.go:28` | Padrão mínimo de liveness já existente; o novo `/health` do Workers segue a mesma simplicidade, só que com corpo JSON. |
| `plug :healthz` com resposta antes do router | `services/collab/lib/collab_web/endpoint.ex:48-52` | Ponto exato de extensão para incluir uptime/memória no Collab (RF02). |
| `routes.ResumoFinanceiro(iam)` → `http.HandlerFunc` | `services/gateway/internal/routes/*.go` | Modelo para `routes.ObterRecursos(monitor)`. |
| `NewRouter(iam, design, logic, deploy, export, rl, oauth)` recebendo um client gRPC por serviço | `services/gateway/internal/app/router.go` | Modelo para adicionar o 6º client (`monitor`). |
| `useResumoFinanceiro.ts` (estados `carregando`/`pronto`/`erro` + `recarregar`) | `services/frontend/src/dashboard/useResumoFinanceiro.ts` | Modelo para `useRecursos.ts`, acrescido de `setInterval` para auto-refresh (RF07). |
| `CardResumoFinanceiro.tsx` | `services/frontend/src/dashboard/CardResumoFinanceiro.tsx` | Modelo visual/estrutural para `CardServicoStatus.tsx`. |
| Rotas aninhadas em `App.tsx` dentro de `/dashboard` + item de sidebar em `DashboardLayout.tsx` | `services/frontend/src/App.tsx:103-114`, `.../DashboardLayout.tsx:76-95` | Onde plugar a rota/nav de Monitor (RF08). |
| `run_bg <nome> go run ./services/<nome>/cmd` | `build/dev-up.sh:230-259` | Onde adicionar o boot do Monitor no fluxo de dev local. |
| `pkg/telemetry`, `pkg/tenantctx`, `pkg/database`, `pkg/eventbus`, `pkg/blindindex` | `pkg/` | Precedente direto para criar `pkg/health` como novo pacote compartilhado. |
| Portas gRPC dos serviços (memória do projeto, confirmado em código) | IAM `:50051`, Design `:50052`, Logic `:50053`, Deploy `:50054`, Export `:50055`, Gateway HTTP `:8080`, Collab HTTP `:4000` | Base para decidir as portas novas (§2). |

---

## 2. Portas e Endereços — decisão final

| Serviço | Endereço/porta | Observação |
|---------|-----------------|------------|
| Monitor (gRPC, novo) | `:50056` (`MONITOR_GRPC_ADDR`) | Próxima porta livre na faixa gRPC 50051-50055 já em uso. |
| Workers (HTTP, novo) | `:8081` (`WORKERS_HTTP_ADDR`) | Fica na faixa HTTP junto do Gateway (`:8080`), não na faixa gRPC — evita a colisão com `:50056` do Monitor identificada em `plan.md` §6, e deixa claro que é um endpoint HTTP, não gRPC. |
| Collab `/healthz` | `:4000` (já existente, `PHX_PORT`/hardcoded em `dev.exs`) | Sem porta nova — só o corpo da resposta muda (RF02). |
| Gateway `/health` | `:8080` (já existente) | Sem mudança — o Monitor só faz `GET http://localhost:8080/health` e converte "200 OK" em `ServicoStatus{status: "servindo"}` (sem uptime/memória do Gateway nesta entrega, já que `/health` não retorna corpo — ver §3, aceito como limitação documentada). |

Todos os 8 endereços (5 gRPC + Gateway HTTP + Collab HTTP + Workers HTTP) são
configuráveis via env var no `main.go` do Monitor, seguindo exatamente o padrão já usado
em `services/gateway/cmd/main.go:37-41`.

---

## 3. Alternativas de Arquitetura/Design Consideradas

### Opção A: Padronizar tudo em HTTP (todos os serviços ganham um `net/http` com `/health`)
- **Prós**: um único tipo de coletor no Monitor, sem precisar de dois protos.
- **Contras**: obriga IAM/Design/Logic/Deploy/Export — hoje gRPC puro — a subir um segundo
  listener cada, só para isso; mais superfície nova por serviço do que estender o
  `grpc.Server` que cada um já tem.
- **Decisão**: Descartada. Ver justificativa completa em `plan.md` §1.

### Opção B: Prometheus + exporters, Monitor só lê o Prometheus
- **Prós**: infraestrutura de métricas "de verdade", série temporal, alertas prontos no
  futuro.
- **Contras**: decisão explícita do usuário nesta entrega foi não usar Prometheus (custo
  de infra + aprendizado adicional não justificado para uma primeira tela de status); o
  OTel Collector já existente é para traces, não métricas, e adaptá-lo teria custo
  similar a montar o Prometheus do zero.
- **Decisão**: Descartada nesta entrega — documentado como possível evolução futura,
  não como "fora de escopo esquecido".

### Opção C: `GET /health` do Gateway devolver corpo JSON com uptime/memória do próprio Gateway
- **Prós**: o Monitor teria dado completo do Gateway, não só "up/down".
- **Contras**: mudaria o contrato de um endpoint público já em uso (usado por
  liveness/orquestração externa, se houver); ampliar essa entrega para editar
  `router.go:28` sai do escopo mínimo definido com o usuário.
- **Decisão**: Descartada nesta rodada — aceito como limitação documentada
  (`spec.md` RN02: "cada serviço reporta o que consegue"); o Gateway aparece na tela como
  up/down sem métricas de memória. Pode virar uma extensão de uma linha numa demanda
  futura (`GET /health` passar a aceitar `Accept: application/json` sem quebrar o
  consumidor atual que só olha o status code).

### Opção D: Monitor usar `tenantctx` interceptors como os demais serviços gRPC
- **Prós**: consistência total com o padrão dos outros `main.go`.
- **Contras**: `tenantctx` existe para propagar/validar o tenant de uma requisição de
  negócio multi-tenant; a RPC do Monitor (`ObterRecursos`) não tem tenant — é uma consulta
  operacional da plataforma inteira, chamada pelo Gateway sem `TenantContext`.
- **Decisão**: Descartada — `services/monitor` registra `otelgrpc` (tracing, RNF03) mas
  **não** encadeia `tenantctx.UnaryServerInterceptor()`. Documentar essa exceção no
  comentário do `main.go` para não parecer omissão acidental na próxima revisão.
