# Pesquisa: Construtor de Sistemas MACH V4

---

## 1. Padrões Existentes no Projeto

O repositório é greenfield (apenas documentação de arquitetura). Os padrões a seguir vêm dos documentos e são vinculantes para a implementação:

| Arquivo/Padrão | Localização | Relevância |
|----------------|-------------|-----------|
| Pilares MACH e divisão dos 5 microsserviços | `doc/ARCHITECTURE_PILLARS.md` | Define fronteiras de serviço — nenhum serviço pode absorver responsabilidade de outro |
| Gateway híbrido Go/Elixir + debounce write-behind | `doc/GATEWAY_COLLABORATION.md` | Fluxo obrigatório da colaboração (GenServer, Redis snapshot, flush 5s, gRPC batch) |
| Blind Index, JWT→Metadata, mapa booleano de permissões | `doc/DATA_SECURITY.md` | Contratos de segurança — formato do payload de permissões já especificado |
| KEDA `ScaledObject`, filas nomeadas, DLQ, fair queuing | `doc/ASYNC_OBSERVABILITY.md` | Topologia de mensageria e nomes de fila (`webhooks.disparo`, `notificacoes.envio`) |
| Contrato `.proto` oficial `LogicEngineService` | `doc/CONTRACTS_PERFORMANCE.md §5` | Deve ser transcrito literalmente para `proto/construtor/logic/v1/` |
| Batching de render 16ms, deploy por flag, export 4 etapas | `doc/CONTRACTS_PERFORMANCE.md` | Requisitos numéricos verificáveis (16ms, rollback ms, presigned URL) |
| Requisitos numerados RF/RNF/RN | `doc/ANALISE_REQUISITOS.md` | Fonte da rastreabilidade usada em spec/plan/tasks |

---

## 2. Tecnologias e Bibliotecas

| Tecnologia | Versão | Uso | Já instalada? |
|------------|--------|-----|---------------|
| Go | ≥ 1.22 | Gateway, 5 microsserviços, workers | Não |
| Elixir / OTP | ≥ 1.16 / 26 | Motor de colaboração (Phoenix) | Não |
| Phoenix / Phoenix Channels | ~> 1.7 | WebSockets, Channels, Presence | Não |
| grpc-elixir (`grpc`) + `protobuf-elixir` | ~> 0.9 | Cliente gRPC Elixir → Design Engine | Não |
| buf | ≥ 1.30 | Lint, geração e breaking-check dos `.proto` | Não |
| grpc-go + protoc-gen-go | latest | Servidores/clientes gRPC em Go | Não |
| pgx | v5 | Driver PostgreSQL com suporte JSONB | Não |
| PostgreSQL | 16 | Base partilhada multi-tenant (JSONB, RLS, índice parcial) | Não (via Compose) |
| Redis | 7 | Snapshots de colaboração + rate limiting | Não (via Compose) |
| RabbitMQ | 3.13 | Mensageria assíncrona, DLQ | Não (via Compose) |
| KEDA | ≥ 2.14 | Autoscaling por QueueLength, scale-to-zero | Não (cluster k8s) |
| OpenTelemetry SDK (Go/Elixir) | latest | Instrumentação HTTP/gRPC/AMQP | Não |
| Jaeger | latest | Backend de traces | Não (via Compose) |
| MinIO | latest | S3 local para Export Engine em dev | Não (via Compose) |
| React + TypeScript + Vite | React 18 / TS 5 | Headless Player (SPA) | Não |
| golang-jwt/jwt | v5 | Emissão/validação JWT RS256 | Não |

---

## 3. Referências Externas

| Referência | URL | O que resolve |
|------------|-----|--------------|
| KEDA RabbitMQ Scaler | https://keda.sh/docs/latest/scalers/rabbitmq-queue/ | Configuração do trigger `QueueLength` e scale-to-zero (RN10) |
| W3C Trace Context | https://www.w3.org/TR/trace-context/ | Formato do header `traceparent` propagado em HTTP/gRPC/AMQP (RNF04) |
| OpenTelemetry Messaging Semantics | https://opentelemetry.io/docs/specs/semconv/messaging/ | Convenção para spans AMQP produtor/consumidor |
| gRPC Metadata (Go) | https://grpc.io/docs/guides/metadata/ | Propagação de identidade/tenant como metadata binário (RNF02) |
| Phoenix Channels | https://hexdocs.pm/phoenix/channels.html | Canais WebSocket por ecrã (RF06) |
| Phoenix Presence | https://hexdocs.pm/phoenix/Phoenix.Presence.html | Presença/cursores via CRDT (RF06) |
| PostgreSQL Row-Level Security | https://www.postgresql.org/docs/16/ddl-rowsecurity.html | Defesa em profundidade para RN01 |
| RabbitMQ Dead Letter Exchanges | https://www.rabbitmq.com/docs/dlx | DLQ por fila (RN09, RNF06) |
| AWS S3 Presigned URLs | https://docs.aws.amazon.com/AmazonS3/latest/userguide/using-presigned-url.html | Entrega segura da exportação (RF05) |
| Blind Indexing (CipherSweet) | https://ciphersweet.paragonie.com/internals/blind-index-planning | Fundamento criptográfico do Blind Index (RN02) |

---

## 4. Alternativas Consideradas

### Opção A: Monorepo poliglota (Go + Elixir + TS num único repositório)
- **Prós**: contratos `.proto` como fonte única sem publicação de pacotes; refactors atômicos cross-service; CI unificado com `buf breaking`.
- **Contras**: pipelines mais complexos (toolchains distintas); risco de acoplamento indevido entre serviços.
- **Decisão**: **Escolhida** — na fase de fundação, a velocidade de iteração nos contratos domina; a separação futura por repositório continua possível pois os serviços só se conhecem pelos protos.

### Opção B: Multi-repo por serviço desde o início
- **Prós**: isolamento total de ciclo de vida (alinha com o pilar M do MACH).
- **Contras**: exige registry de pacotes proto (BSR), versionamento e sincronização de contratos antes mesmo do primeiro serviço existir.
- **Decisão**: Descartada nesta fase; reavaliada quando os contratos `v1` estabilizarem.

### Opção C: Persistência colaborativa direta (cada mutação → PostgreSQL)
- **Prós**: simplicidade, sem Redis nem GenServer com estado.
- **Contras**: exaustão de escrita com micro-movimentos de UI; contradiz RN06 e `doc/GATEWAY_COLLABORATION.md`.
- **Decisão**: Descartada — write-behind com debounce de 5s é requisito documentado.

### Opção D: Kafka em vez de RabbitMQ
- **Prós**: replay de eventos, throughput superior, particionamento nativo por tenant.
- **Contras**: scaler KEDA de lag é mais complexo; a topologia documentada (exchanges, routing dinâmico, DLQ, fair queuing) é idiomática de RabbitMQ; sem requisito de replay.
- **Decisão**: Descartada — os docs de arquitetura fixam RabbitMQ (`doc/ASYNC_OBSERVABILITY.md`).

### Opção E: Schema-per-tenant em vez de shared database
- **Prós**: isolamento físico mais forte.
- **Contras**: custo de RAM/CPU e de migrações por tenant; contradiz a decisão explícita de Shared Database para eficiência de custos (pilar C).
- **Decisão**: Descartada — shared database + `tenant_id` + RLS é a arquitetura documentada.
