# Quickstart: Construtor de Sistemas MACH V4

Guia para rodar e testar a plataforma localmente.

---

## Pré-requisitos

- Docker + Docker Compose v2
- Go ≥ 1.22
- Elixir ≥ 1.16 (OTP 26) — apenas para trabalhar no motor de colaboração
- Node ≥ 20 — apenas para trabalhar no Headless Player
- `buf` CLI ≥ 1.30 (`go install github.com/bufbuild/buf/cmd/buf@latest`)

---

## Passos

```bash
# 1. Gerar chaves RS256 para o IAM (uma vez)
mkdir -p .keys && openssl genrsa -out .keys/jwt_private.pem 2048 \
  && openssl rsa -in .keys/jwt_private.pem -pubout -out .keys/jwt_public.pem

# 2. Subir a infraestrutura (PostgreSQL, Redis, RabbitMQ, Jaeger, MinIO)
docker compose up -d postgres redis rabbitmq jaeger minio

# 3. Rodar as migrações
make migrate

# 4. Gerar os stubs a partir dos contratos .proto
make proto

# 5. Subir todos os serviços (gateway, engines, collab, workers)
docker compose up -d

# 6. Criar tenant e utilizador de desenvolvimento (seed)
make seed

# 7. Obter um token de teste
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -d '{"email":"dev@local","password":"dev"}' | jq -r .token
```

---

## Verificação

1. **Gateway**: `curl http://localhost:8080/healthz` → `200`.
2. **Fluxo completo**: com o token, crie um design (`POST /api/v1/designs`), publique (`POST /api/v1/sistemas/{id}/publicar`) e submeta um formulário (`POST /api/v1/formularios`) — ver `contracts/api.md`.
3. **Colaboração**: abra `ws://localhost:4000/socket` em dois clientes no mesmo ecrã; mutações de um aparecem no outro; após 5s de silêncio, o log do collab mostra 1 chamada `SalvarDesign`.
4. **Traces**: acesse o Jaeger em `http://localhost:16686` e busque pelo serviço `gateway` — um trace deve atravessar gateway → logic-engine → rabbitmq → worker.
5. **Filas**: management do RabbitMQ em `http://localhost:15672` (guest/guest) — filas `webhooks.disparo` e `notificacoes.envio` com DLQs.

```bash
# Rodar testes específicos desta demanda
make test                # tudo (Go + ExUnit + player)
go test ./services/...   # apenas serviços Go
(cd collab && mix test)  # apenas motor de colaboração
(cd player && npm test)  # apenas Headless Player
```

---

## Variáveis de Ambiente

| Variável | Valor de Exemplo | Descrição |
|----------|-----------------|-----------|
| `DATABASE_URL` | `postgres://app:app@localhost:5432/construtor` | Conexão PostgreSQL |
| `REDIS_URL` | `redis://localhost:6379/0` | Snapshots de colaboração e rate limiting |
| `AMQP_URL` | `amqp://guest:guest@localhost:5672/` | RabbitMQ |
| `JWT_PRIVATE_KEY_PATH` | `.keys/jwt_private.pem` | Assinatura de tokens (apenas IAM) |
| `JWT_PUBLIC_KEY_PATH` | `.keys/jwt_public.pem` | Validação de tokens (gateway/serviços) |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `http://localhost:4317` | Export de traces (Collector/Jaeger) |
| `S3_ENDPOINT` | `http://localhost:9000` | MinIO local (Export Engine) |
| `S3_BUCKET` | `exports` | Bucket dos pacotes de exportação |
| `PRESIGNED_URL_TTL` | `10m` | Expiração dos links de download |
| `COLLAB_DEBOUNCE_MS` | `5000` | Janela de inatividade antes do flush gRPC (RN06) |
| `GATEWAY_RATE_LIMIT_RPS` | `50` | Limite de requisições por tenant |
