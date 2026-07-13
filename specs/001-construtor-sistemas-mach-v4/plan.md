# Plano de Implementação: Construtor de Sistemas MACH V4 — Fundação da Plataforma

Estratégia: monorepo poliglota com contratos Protocol Buffers como fonte única de verdade (`proto/`), serviços Go compartilhando bibliotecas internas (`pkg/`), motor de colaboração em Elixir isolado, e infraestrutura declarativa (Docker Compose para dev, manifests Kubernetes/KEDA para produção). A implementação avança em fases verticais: primeiro a fundação (contratos + infra local + segurança), depois cada engine, por fim as camadas transversais (mensageria, observabilidade, player).

---

## 1. Arquivos a Criar/Editar

### 1.1. Contratos e Fundação do Monorepo

* **`proto/construtor/common/v1/tenant.proto`**: mensagens comuns (`TenantContext`, tipos partilhados) — base de RN01/RNF02.
* **`proto/construtor/design/v1/design.proto`**: serviço `DesignEngineService` (CRUD de árvore recursiva, `SalvarDesign` em lote) — RF01, RF06.
* **`proto/construtor/logic/v1/logic.proto`**: serviço `LogicEngineService` (`SalvarFormulario` Unary, CRUD de regras) — RF02, RF07 (contrato oficial de `doc/CONTRACTS_PERFORMANCE.md §5`).
* **`proto/construtor/iam/v1/iam.proto`**: serviço `IAMService` (`AvaliarPermissoes`, `ValidarToken`) — RF03.
* **`proto/construtor/deploy/v1/deploy.proto`**: serviço `DeployEngineService` (`Publicar`, `Rollback`) — RF04.
* **`proto/construtor/export/v1/export.proto`**: serviço `ExportEngineService` (`CriarJob`, `ColetarDados` server-streaming) — RF05.
* **`buf.yaml` / `buf.gen.yaml`**: lint, breaking-change detection e geração de stubs Go/Elixir/TypeScript.
* **`Makefile`**: alvos `proto`, `test`, `up`, `migrate`.
* **`docker-compose.yml`**: PostgreSQL, Redis, RabbitMQ, Jaeger, serviços locais.

### 1.2. Bibliotecas Go Partilhadas (`pkg/`)

* **`pkg/tenantctx/`**: extração/injeção de `tenant_id` + identidade em gRPC Metadata (RN01, RNF02); interceptores server/client.
* **`pkg/blindindex/`**: geração e verificação do hash criptográfico (HMAC-SHA256 com chave por tenant) — RN02, RNF08.
* **`pkg/telemetry/`**: bootstrap OpenTelemetry (tracer, propagadores W3C, exporter OTLP→Jaeger), atributos `platform.tenant_id` e `platform.component.blind_index` — RNF04.
* **`pkg/database/`**: pool pgx com hook obrigatório de filtro por tenant (guard que rejeita queries sem `tenant_id`) — RN01.

### 1.3. IAM Service (`services/iam/`)

* **`services/iam/cmd/main.go`**: bootstrap gRPC server + telemetry.
* **`services/iam/internal/auth/jwt.go`**: emissão/validação de JWT (RS256), claims `tenant_id`, `sub`, `tipo`.
* **`services/iam/internal/permissions/evaluator.go`**: avaliação server-side das condições dinâmicas; retorna mapa `blind_index → {view, click}` (RN03).
* **`services/iam/internal/store/`**: persistência de tenants hierárquicos e permissões.

### 1.4. API Gateway em Go (`gateway/`)

* **`gateway/cmd/main.go`**: servidor HTTP (chi/echo) + clientes gRPC.
* **`gateway/internal/middleware/auth.go`**: validação JWT do header `Authorization`, injeção em Metadata (RF03, RNF02).
* **`gateway/internal/middleware/ratelimit.go`**: rate limiting por tenant.
* **`gateway/internal/middleware/tracing.go`**: geração do Trace ID raiz (RNF04).
* **`gateway/internal/routes/`**: tradução REST→gRPC por recurso (`designs.go`, `regras.go`, `deploy.go`, `export.go`, `formularios.go`) — contratos em `contracts/api.md`.

### 1.5. Design Engine (`services/design/`)

* **`services/design/internal/tree/composite.go`**: modelo da árvore recursiva (`componente_filhos`) com validação estrutural (RF01).
* **`services/design/internal/store/jsonb.go`**: persistência da definição em coluna JSONB com `tenant_id` (RN01).
* **`services/design/internal/server/grpc.go`**: implementação do `DesignEngineService`, incluindo `SalvarDesign` em lote consumido pelo motor Elixir (RN06).

### 1.6. Logic Engine (`services/logic/`)

* **`services/logic/internal/rules/tree.go`**: árvore de decisão (nós lógicos) e interpretador (RF02).
* **`services/logic/internal/validation/schema.go`**: revalidação de payload contra `CampoDefinicao` por blind_index; mapa de erros anónimo (RN02, RN08, RNF08).
* **`services/logic/internal/events/publisher.go`**: publicação de eventos AMQP com `tenant_id` + `component_blind_index` e propagação de trace nos headers (RF08, RNF04).
* **`services/logic/internal/server/grpc.go`**: `SalvarFormulario` e CRUD de regras.

### 1.7. Deploy Engine (`services/deploy/`)

* **`services/deploy/internal/versions/manager.go`**: transação atômica publicar/rollback por flag ativa (RN04, RN05, RNF05).
* **`services/deploy/internal/server/grpc.go`**: `Publicar`, `Rollback`, `ObterVersaoAtiva`.

### 1.8. Export Engine (`services/export/`)

* **`services/export/internal/jobs/manager.go`**: ciclo de vida do Job (criado → coletando → pronto → expirado) (RF05).
* **`services/export/internal/collector/streaming.go`**: consumo via gRPC Server Streaming em chunks (RNF01).
* **`services/export/internal/storage/s3.go`**: upload para bucket e geração de Presigned URL com expiração curta.

### 1.9. Motor de Colaboração Elixir (`collab/`)

* **`collab/lib/collab_web/channels/screen_channel.ex`**: Phoenix Channel por ecrã em edição (RF06).
* **`collab/lib/collab/session/screen_server.ex`**: GenServer por ecrã; árvore em memória; debounce de 5s e flush único via gRPC (RN06).
* **`collab/lib/collab/session/redis_snapshot.ex`**: replicação de snapshots no Redis.
* **`collab/lib/collab_web/presence.ex`**: Phoenix Presence (CRDT) para cursores e utilizadores online.
* **`collab/lib/collab/session/locks.ex`**: bloqueio otimista por blind_index (RN07).
* **`collab/lib/collab/grpc/design_client.ex`**: cliente gRPC para o Design Engine.

### 1.10. Workers Assíncronos (`workers/`)

* **`workers/cmd/main.go`**: consumidor AMQP genérico com extração de `traceparent` dos headers (RNF04).
* **`workers/internal/handlers/webhook.go`** e **`notification.go`**: execução das tarefas (RF08).
* **`workers/internal/dlq/`**: roteamento de falhas contínuas para DLQ + alerta ao tenant (RN09, RNF06).

### 1.11. Infraestrutura (`infra/`)

* **`infra/rabbitmq/definitions.json`**: exchanges, filas (`webhooks.disparo`, `notificacoes.envio`), DLQs e políticas de fair queuing (RN09).
* **`infra/k8s/keda/scaledobject-workers.yaml`**: `ScaledObject` com trigger `QueueLength`, `minReplicaCount: 0`, `maxReplicaCount: 50` (RN10, RNF03).
* **`infra/k8s/`**: deployments/services de cada componente.
* **`infra/otel/collector-config.yaml`**: pipeline OTLP → Jaeger (RNF04).
* **`infra/postgres/migrations/`**: migrações versionadas (ver `data-model.md §4`).

### 1.12. Headless Player (`player/`)

* **`player/src/renderer/CompositeRenderer.tsx`**: renderização recursiva por `componente_filhos` (RF01/H do MACH).
* **`player/src/renderer/batcher.ts`**: acumulação de mutações em janelas de 16ms com diffing único por lote (RNF07).
* **`player/src/validation/blindIndexValidator.ts`**: validação local pelo mapa de definições (RN08).
* **`player/src/permissions/permissionMap.ts`**: aplicação do mapa booleano `blind_index → {view, click}` (RN03).
* **`player/src/router/dynamicRoutes.ts`**: navegação SPA via ações `redirect`.

---

## 2. Estratégia Técnica

### 2.1. Contratos primeiro (API-first real)

Todo o desenvolvimento parte dos `.proto` compilados com **buf**. Nenhum serviço define tipos de fronteira próprios: os stubs gerados (Go, Elixir via `grpc-elixir`, TypeScript para o player consumir via Gateway) são a única interface. O CI roda `buf breaking` contra a `main` para impedir quebra de contrato — alternativa descartada: OpenAPI-first (REST interno), rejeitada porque a comunicação interna é exclusivamente gRPC (RNF01).

```protobuf
// proto/construtor/logic/v1/logic.proto — contrato oficial (doc/CONTRACTS_PERFORMANCE.md §5)
service LogicEngineService {
  rpc SalvarFormulario (SalvarFormularioRequest) returns (SalvarFormularioResponse);
}
```

### 2.2. Isolamento de tenant imposto por construção, não por disciplina

O filtro `tenant_id` (RN01) não depende de cada desenvolvedor lembrar do `WHERE`: o `pkg/database` expõe apenas um `TenantScopedQuerier` que exige `tenantctx.TenantID(ctx)` e injeta o predicado automaticamente. Queries sem contexto de tenant falham em runtime e são bloqueadas por lint no CI. Alternativa considerada: Row-Level Security nativa do PostgreSQL (`SET app.tenant_id`) — mantida como reforço adicional na camada de migração (defesa em profundidade), não como único mecanismo.

### 2.3. Debounce write-behind no GenServer

Cada ecrã em edição vive num `GenServer` registado via `Registry` por `{sistema_id, screen_id}`. Toda mutação reinicia um timer de 5s (`Process.send_after`); no timeout, a árvore consolidada segue numa única chamada `SalvarDesign` (RN06). Snapshots incrementais vão ao Redis a cada mutação para recuperação em caso de queda do nó BEAM. Alternativa descartada: persistir cada mutação diretamente no PostgreSQL — rejeitada pelo risco de exaustão de escrita documentado em `doc/GATEWAY_COLLABORATION.md`.

### 2.4. Propagação de trace através de três protocolos

Um único trace atravessa HTTP → gRPC → AMQP (RNF04): o Gateway cria o span raiz; interceptores OTel propagam via Metadata gRPC; o publisher AMQP serializa `traceparent` nos headers da mensagem; o worker extrai e abre span filho. Os atributos `platform.tenant_id` e `platform.component.blind_index` são adicionados em todos os spans — nunca nomes reais de campos (RNF08).

### 2.5. Scale-to-zero com KEDA

Workers são Deployments com `replicas` geridas pelo KEDA via `ScaledObject` (trigger `rabbitmq`, métrica `QueueLength`). Fila vazia → 0 pods; acúmulo → escala até 50 (RN10, RNF03). Fair queuing (RN09) implementado com routing keys por tenant + política de `x-max-priority`/consumer prefetch baixo, e DLQ por fila com alarme direcionado ao tenant.

---

## 3. Dependências e Pré-requisitos

- [ ] Go ≥ 1.22, Elixir ≥ 1.16/OTP 26, Node ≥ 20 instalados
- [ ] `buf` CLI instalado para geração/validação dos contratos
- [ ] Docker + Docker Compose para infraestrutura local (PostgreSQL 16, Redis 7, RabbitMQ 3.13 com plugin de management, Jaeger)
- [ ] Cluster Kubernetes com KEDA instalado (apenas para validação dos manifests de produção; dev usa Compose)
- [ ] Bucket S3/GCS (ou MinIO local) para o Export Engine
- [ ] Par de chaves RS256 para assinatura JWT do IAM Service

---

## 4. Riscos e Pontos de Atenção

| Risco | Impacto | Mitigação |
|-------|---------|-----------|
| Vazamento cross-tenant por query sem filtro (RN01) | Alto | `TenantScopedQuerier` obrigatório + RLS PostgreSQL como segunda camada + teste de integração multi-tenant no CI |
| Perda de edições se o nó BEAM cair antes do flush de 5s (RN06) | Alto | Snapshot em Redis a cada mutação; na reinicialização o GenServer reidrata do Redis antes de aceitar conexões |
| Exposição de nomes reais em logs/traces (RNF08) | Alto | Redator central no `pkg/telemetry`; revisão de todos os pontos de log no code review; teste que faz grep de nomes de schema nos payloads de erro |
| Contrato gRPC quebrado entre serviços poliglotas | Médio | `buf breaking` no CI + versionamento `v1` nos packages proto |
| Noisy neighbor esgotando workers (RN09) | Médio | Fair queuing por routing key de tenant, prefetch baixo, DLQ isolada e alertas por tenant |
| Elixir + Go no mesmo monorepo complica CI | Médio | Pipelines separados por diretório com cache independente; Makefile unificado apenas para dev local |
| Presigned URL vazada (RF05) | Médio | Expiração curta (minutos), URL vinculada a IP opcional, bucket privado com bloqueio de acesso público |
| Ordem de mutações concorrentes no mesmo componente (RN07) | Médio | Bloqueio otimista por blind_index no GenServer (fonte única de ordem por ecrã) |
