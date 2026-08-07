# USAGE — Como rodar o MACH V4 localmente

Guia de startup do monorepo. Ordem: **infra → proto → Go services → Gateway → Collab → Player**.

Todos os scripts de build/startup/deploy do repo vivem em **`build/`**.

---

## Startup guiado (recomendado)

```bash
./build/dev-up.sh              # sobe tudo, com checagens e prompts de confirmação
./build/dev-up.sh --no-player  # sobe tudo menos o player (ex.: você já roda o Vite noutro terminal)
./build/dev-up.sh --yes        # não pergunta nada, assume "sim" em todos os prompts
```

O que ele faz, na ordem, com feedback visual (✓/✗/!) a cada etapa:

1. **Pré-checagens** — confirma `docker`, `go`, `node`, `npm`, `mix`, `buf` no PATH e a versão do Go; ajusta o PATH automaticamente para as toolchains locais (`$HOME/.local/go`, `$HOME/.local/elixir1.17`).
2. **Infra** (`make up` + `make migrate`) — avisa e pede confirmação se alguma porta já estiver ocupada (ex.: MinIO 9000 usado por outro projeto) antes de prosseguir.
3. **Proto** (`make proto`) — regenera `gen/go`, `gen/elixir`, `gen/ts`.
4. **Services gRPC** (iam, design, logic, deploy, export) — sobem em background, com espera ativa até cada porta responder.
5. **Workers** (RabbitMQ).
6. **Gateway** (`:8080`).
7. **Collab** (Phoenix, `:4000`).
8. **Player** — instala deps se faltarem e pergunta se quer abrir agora (`npm run dev`, foreground).

Ao final, mostra um resumo com as URLs de cada serviço. **Ctrl+C encerra todos os processos que o script iniciou.**

**Logs**: tudo — inclusive `make up`, `make proto`, `npm install`, `mix deps.get` — é gravado em `.dev-logs/<nome>.log` (pasta única, gitignored), além de aparecer na tela.

---

## Passo a passo manual

Use isto se preferir rodar cada peça na mão, ou para debugar uma etapa específica que o `dev-up.sh` reportou com falha.

### 0. Pré-requisitos de toolchain

O apt do sistema é velho demais para as deps do repo — use as versões locais instaladas:

```bash
# Go 1.26 (apt tem 1.22, insuficiente para minio-go/x-net/protobuf)
export PATH="$HOME/.local/go/bin:$PATH"

# Elixir 1.17.3 / OTP 25 (apt tem 1.14, insuficiente para Phoenix/Plug/Bandit)
export PATH="$HOME/.local/elixir1.17/bin:$PATH"
export MIX_HOME="$HOME/.mix"
export HEX_HOME="$HOME/.hex"
```

Outras dependências: Docker + Docker Compose, Node 20, `buf` (`make tools` instala em `$(go env GOPATH)/bin`).

### 1. Sobe a infraestrutura (Docker Compose)

```bash
make up        # postgres:5432, redis:6379, rabbitmq:5672/15672, jaeger:16686, otel-collector:4317/4318, minio:9000/9001
make migrate   # aplica infra/postgres/migrations/*.sql
```

> **Gotcha de porta**: o `minio` do compose usa a porta host **9000/9001**. Se outro projeto já ocupar essa porta, suba um MinIO avulso (`docker run -p 9010:9000 ...` com creds `mach`/`machsecret`) e aponte `S3_ENDPOINT=localhost:9010` nas envs abaixo.

### 2. Gera os stubs de proto

Necessário antes de compilar Go e Elixir (o `collab` compila `gen/elixir`, que é gitignored):

```bash
make proto     # buf lint + buf generate → gen/go, gen/elixir, gen/ts
```

Requer `buf` (`make tools`) e, para o alvo Elixir, `protoc-gen-elixir` no PATH:

```bash
mix escript.install hex protobuf   # uma vez, com o Elixir 1.17 já no PATH
export PATH="$HOME/.mix/escripts:$PATH"
```

### 3. Sobe os serviços gRPC (Go)

Cada serviço em um terminal, a partir da raiz do repo. Todos leem `DATABASE_URL` e `OTEL_EXPORTER_OTLP_ENDPOINT` com defaults já apontando para a infra do compose — normalmente não precisa setar nada:

```bash
go run ./services/iam/cmd      # IAM      :50051
go run ./services/design/cmd   # Design   :50052
go run ./services/logic/cmd    # Logic    :50053 (usa RABBITMQ_URL)
go run ./services/deploy/cmd   # Deploy   :50054
go run ./services/export/cmd   # Export   :50055 (usa S3_ENDPOINT/S3_ACCESS_KEY/S3_SECRET_KEY/S3_BUCKET)
```

#### Workers assíncronos (RabbitMQ)

```bash
go run ./workers/cmd           # consome filas via RABBITMQ_URL (default amqp://mach:mach@localhost:5672/)
```

### 4. Sobe o Gateway HTTP

```bash
go run ./gateway/cmd           # :8080
```

Lê os endereços dos services acima via env (defaults já corretos para local):
`GATEWAY_HTTP_ADDR`, `IAM_GRPC_ADDR`, `DESIGN_GRPC_ADDR`, `LOGIC_GRPC_ADDR`, `DEPLOY_GRPC_ADDR`, `EXPORT_GRPC_ADDR`, `OTEL_EXPORTER_OTLP_ENDPOINT`.

Para login social em dev, opcionalmente: `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`, `OAUTH_ALLOWED_REDIRECT_URIS`.

### 5. Sobe o Collab (Elixir/Phoenix — colaboração em tempo real)

```bash
cd collab
mix deps.get
mix phx.server     # http://localhost:4000
```

> `jose` está pinado em `1.11.5` no `mix.exs` (compat com OTP 25).

### 6. Sobe o Player (frontend Vite/React)

```bash
cd player
npm install         # se node_modules ainda não existir
npm run dev
```

Config em `player/.env.local` (`VITE_BYPASS_AUTH=true` pula auth em dev). Config runtime injetada via `window.__PLAYER_CONFIG__` (baseUrl/token/sistemaId do host).

---

## Referência rápida de portas

| Serviço          | Porta   |
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
| Player (Vite dev)  | 5173 (default do Vite) |

## Comandos úteis (Makefile)

```bash
make help            # lista todos os alvos
make down             # derruba a infra do compose
make test             # go test ./...
make tidy             # go mod tidy
make proto-breaking   # buf breaking --against main
```

## Build e deploy (CI/CD, spec 002)

Também em `build/`, usados pelo pipeline `.github/workflows/cd.yml` (e reaproveitáveis localmente para ensaiar um release):

```bash
SHA=$(git rev-parse --short HEAD) build/build-artifacts.sh          # empacota os 7 binários + release Elixir + player/dist em dist/artifacts/
build/deploy.sh --env staging --host <host> --user <user> --sha <sha>   # rsync + troca atômica de symlink + restart
build/smoke-test.sh --host <host>                                    # healthcheck pós-deploy
build/rollback.sh --env staging --host <host> [--sha <sha>]          # repontar current ao release anterior, sem rebuild
```

Detalhes completos em `specs/002-entrega-continua-artefatos/` e `infra/deploy/README.md`.

## Testes de integração / E2E

Exigem a infra do compose de pé (`make up` + `make migrate`) e rodam serial (`-p 1`) para evitar race em `GRANT ... ON SCHEMA`:

```bash
DATABASE_URL=postgres://mach:mach@localhost:5432/machv4?sslmode=disable \
  go test -tags integration -p 1 ./...

# E2E de tracing precisa de rabbitmq/jaeger/otel-collector rodando (via compose)
OTLP_ENDPOINT=localhost:4317 JAEGER_QUERY=localhost:16686 \
  RABBITMQ_URL=amqp://mach:mach@localhost:5672/ \
  go test -tags e2e ./tests/e2e/...
```

Testes do player: `cd player && npm test` (vitest) e `npm run typecheck` (`tsc --noEmit`).
