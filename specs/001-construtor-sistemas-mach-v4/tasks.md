# Tarefas: Construtor de Sistemas MACH V4 — Fundação da Plataforma

<!-- Ordenadas por dependência de execução. Cada tarefa é atômica (≤ 1 dia). -->

## Fase 0 — Fundação e Contratos

- [x] 1. Inicializar monorepo: estrutura de pastas, `Makefile`, `.gitignore`, workspace Go (`go.work`) (`Makefile`, `go.work`)
- [x] 2. Escrever `docker-compose.yml` com PostgreSQL 16, Redis 7, RabbitMQ 3.13 (management), Jaeger e MinIO (`docker-compose.yml`)
- [x] 3. Criar contrato comum de tenant/identidade (`proto/construtor/common/v1/tenant.proto`) [RNF02]
- [x] 4. Transcrever contrato oficial do Logic Engine de `doc/CONTRACTS_PERFORMANCE.md §5` e ampliar com CRUD de regras (`proto/construtor/logic/v1/logic.proto`) [RF02, RF07]
- [x] 5. Criar contratos Design, IAM, Deploy e Export (`proto/construtor/{design,iam,deploy,export}/v1/*.proto`) [RF01, RF03, RF04, RF05]
- [x] 6. Configurar buf (lint + breaking + geração Go/Elixir/TS) e alvo `make proto` (`buf.yaml`, `buf.gen.yaml`)
- [x] 7. Escrever migrações 0001–0009 conforme `data-model.md §4` (`infra/postgres/migrations/`) [RN01, RN02, RN04]
- [x] 8. Escrever migração 0010 de Row-Level Security por `tenant_id` (`infra/postgres/migrations/0010_enable_row_level_security.sql`) [RN01]

## Fase 1 — Bibliotecas Partilhadas Go

- [ ] 9. Implementar `pkg/tenantctx` com testes: extração/injeção de tenant em gRPC Metadata + interceptores (`pkg/tenantctx/`) [RN01, RNF02]
- [ ] 10. Implementar `pkg/blindindex` com testes: HMAC-SHA256 com chave por tenant (`pkg/blindindex/`) [RN02, RNF08]
- [ ] 11. Implementar `pkg/database` com testes: `TenantScopedQuerier` que injeta filtro de tenant e rejeita queries sem contexto (`pkg/database/`) [RN01]
- [ ] 12. Implementar `pkg/telemetry`: bootstrap OTel, propagador W3C, atributos `platform.tenant_id`/`platform.component.blind_index`, redator de dados sensíveis (`pkg/telemetry/`) [RNF04, RNF08]

## Fase 2 — IAM Service e Gateway

- [ ] 13. IAM: store de tenants hierárquicos e papéis com testes (`services/iam/internal/store/`) [RF03]
- [ ] 14. IAM: emissão e validação de JWT RS256 com claims `tenant_id`/`sub`/`tipo` (`services/iam/internal/auth/jwt.go`) [RF03]
- [ ] 15. IAM: avaliador de permissões server-side retornando mapa `blind_index → {view, click}` com testes (`services/iam/internal/permissions/evaluator.go`) [RN03]
- [ ] 16. IAM: servidor gRPC + bootstrap com telemetry (`services/iam/internal/server/grpc.go`, `services/iam/cmd/main.go`)
- [ ] 17. Gateway: middleware de autenticação JWT → Metadata gRPC (`gateway/internal/middleware/auth.go`) [RF03, RNF02]
- [ ] 18. Gateway: middlewares de rate limiting por tenant e tracing raiz (`gateway/internal/middleware/{ratelimit,tracing}.go`) [RNF04]
- [ ] 19. Gateway: bootstrap HTTP + clientes gRPC + rota de permissões (`gateway/cmd/main.go`, `gateway/internal/routes/permissions.go`) [RF03]
- [ ] 20. Teste de integração: request sem JWT → 401; JWT tenant A nunca acessa dados do tenant B (`gateway/tests/auth_integration_test.go`) [RN01, critério 1]

## Fase 3 — Design Engine

- [ ] 21. Modelo Composite da árvore recursiva com validação estrutural e testes (`services/design/internal/tree/composite.go`) [RF01]
- [ ] 22. Persistência JSONB com `TenantScopedQuerier` (`services/design/internal/store/jsonb.go`) [RF01, RN01]
- [ ] 23. Servidor gRPC do Design Engine incluindo `SalvarDesign` em lote (`services/design/internal/server/grpc.go`, `services/design/cmd/main.go`) [RF01, RN06]
- [ ] 24. Gateway: rotas REST→gRPC de designs (`gateway/internal/routes/designs.go`) [RF01]

## Fase 4 — Logic Engine

- [ ] 25. Árvore de decisão (nós lógicos) e interpretador com testes (`services/logic/internal/rules/tree.go`) [RF02]
- [ ] 26. CRUD de `campos_definicao` por blind_index (`services/logic/internal/store/campos.go`) [RN02]
- [ ] 27. Revalidação de payload contra schema com mapa de erros por blind_index e testes (`services/logic/internal/validation/schema.go`) [RN08, RNF08, critério 2]
- [ ] 28. Servidor gRPC `SalvarFormulario` + persistência em `dados_operacionais` (`services/logic/internal/server/grpc.go`, `services/logic/cmd/main.go`) [RF07]
- [ ] 29. Gateway: rotas de regras e formulários (`gateway/internal/routes/{regras,formularios}.go`) [RF02, RF07]
- [ ] 30. Teste de integração: submissão maliciosa direto na API é rejeitada sem expor nomes reais (`services/logic/tests/validation_integration_test.go`) [RN08, RNF08, critério 2]

## Fase 5 — Deploy Engine

- [ ] 31. Gerenciador de versões: publicar/rollback em transação atômica com índice único parcial (`services/deploy/internal/versions/manager.go`) [RN04, RN05]
- [ ] 32. Servidor gRPC `Publicar`/`Rollback`/`ObterVersaoAtiva` + rotas no Gateway (`services/deploy/internal/server/grpc.go`, `gateway/internal/routes/deploy.go`) [RF04]
- [ ] 33. Teste de integração: rollback < 100ms e unicidade da flag ativa sob concorrência (`services/deploy/tests/rollback_test.go`) [RNF05, critério 3]

## Fase 6 — Motor de Colaboração (Elixir)

- [ ] 34. Bootstrap do projeto Phoenix (sem HTML), socket e `ScreenChannel` com autorização por JWT (`collab/lib/collab_web/channels/screen_channel.ex`) [RF06]
- [ ] 35. `ScreenServer` (GenServer por ecrã via Registry): estado da árvore + aplicação de mutações (`collab/lib/collab/session/screen_server.ex`) [RF06]
- [ ] 36. Snapshots incrementais no Redis + reidratação na inicialização (`collab/lib/collab/session/redis_snapshot.ex`) [RN06, risco de perda de edições]
- [ ] 37. Debounce de 5s e flush único via cliente gRPC `SalvarDesign` (`collab/lib/collab/session/screen_server.ex`, `collab/lib/collab/grpc/design_client.ex`) [RN06, critério 4]
- [ ] 38. Phoenix Presence para cursores/utilizadores online (`collab/lib/collab_web/presence.ex`) [RF06]
- [ ] 39. Bloqueio otimista por blind_index com liberação por timeout (`collab/lib/collab/session/locks.ex`) [RN07]
- [ ] 40. Teste ExUnit: mutação de A chega a B; 5s de silêncio → exatamente 1 chamada gRPC (`collab/test/collab/session/screen_server_test.exs`) [RN06, critério 4]

## Fase 7 — Mensageria Assíncrona e Workers

- [ ] 41. Definições RabbitMQ: exchanges, filas `webhooks.disparo`/`notificacoes.envio`, DLQs e políticas de fair queuing (`infra/rabbitmq/definitions.json`) [RN09]
- [ ] 42. Publisher AMQP no Logic Engine com `tenant_id`, `component_blind_index` e `traceparent` nos headers (`services/logic/internal/events/publisher.go`) [RF08, RNF04]
- [ ] 43. Worker consumidor genérico com extração de trace e handlers de webhook/notificação (`workers/cmd/main.go`, `workers/internal/handlers/`) [RF08]
- [ ] 44. Roteamento de falhas contínuas para DLQ + alerta ao tenant (`workers/internal/dlq/`) [RN09, RNF06]
- [ ] 45. Manifests k8s dos workers + `ScaledObject` KEDA (`minReplicaCount: 0`, `maxReplicaCount: 50`, trigger QueueLength) (`infra/k8s/keda/scaledobject-workers.yaml`) [RN10, RNF03, critério 5]

## Fase 8 — Export Engine

- [ ] 46. Gerenciador do ciclo de vida do Job com estados e testes (`services/export/internal/jobs/manager.go`) [RF05]
- [ ] 47. Coletor via gRPC Server Streaming em chunks dos 3 serviços fonte (`services/export/internal/collector/streaming.go`) [RF05, RNF01]
- [ ] 48. Upload S3/MinIO + Presigned URL com expiração curta + rotas no Gateway (`services/export/internal/storage/s3.go`, `gateway/internal/routes/export.go`) [RF05, critério 7]

## Fase 9 — Observabilidade Fim-a-Fim

- [ ] 49. OTel Collector config + wiring Jaeger no Compose (`infra/otel/collector-config.yaml`, `docker-compose.yml`) [RNF04]
- [ ] 50. Instrumentar o motor Elixir com OpenTelemetry (spans de channel e flush gRPC) (`collab/lib/collab/telemetry.ex`) [RNF04]
- [ ] 51. Teste E2E de trace: um `traceparent` atravessa Gateway → gRPC → AMQP → Worker num único trace no Jaeger (`tests/e2e/tracing_test.go`) [RNF04, critério 6]

## Fase 10 — Headless Player

- [ ] 52. Bootstrap Vite + React + TS e cliente HTTP autenticado (`player/package.json`, `player/src/api/client.ts`)
- [ ] 53. `CompositeRenderer`: renderização recursiva por `componente_filhos` a partir da versão ativa (`player/src/renderer/CompositeRenderer.tsx`) [RF01, RN04]
- [ ] 54. Batcher de 16ms com diffing único por lote + teste de timing (`player/src/renderer/batcher.ts`) [RNF07]
- [ ] 55. Validador local por mapa de blind_index (`player/src/validation/blindIndexValidator.ts`) [RN08]
- [ ] 56. Aplicação do mapa de permissões `{view, click}` e rotas dinâmicas SPA (`player/src/permissions/permissionMap.ts`, `player/src/router/dynamicRoutes.ts`) [RN03]
- [ ] 57. Cliente WebSocket Phoenix para colaboração no modo builder-preview (`player/src/collab/phoenixSocket.ts`) [RF06]

## Fase 11 — Encerramento

- [ ] 58. Pipeline CI: buf lint/breaking, testes Go, ExUnit, testes player, subida do Compose para integração (`.github/workflows/ci.yml`)
- [ ] 59. Executar a suíte de testes completa (Go + ExUnit + player + integração + E2E) e corrigir regressões (`make test`)
