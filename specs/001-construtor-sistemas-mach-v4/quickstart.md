# Quickstart: MACH V4 System Builder

Guide to run and test the platform locally.

---

## Prerequisites

- Docker + Docker Compose v2
- Go ≥ 1.22
- Elixir ≥ 1.16 (OTP 26) — only needed to work on the collaboration engine
- Node ≥ 20 — only needed to work on the Headless Player
- `buf` CLI ≥ 1.30 (`go install github.com/bufbuild/buf/cmd/buf@latest`)

---

## Steps

```bash
# 1. Generate RS256 keys for the IAM (once)
mkdir -p .keys && openssl genrsa -out .keys/jwt_private.pem 2048 \
  && openssl rsa -in .keys/jwt_private.pem -pubout -out .keys/jwt_public.pem

# 2. Start the infrastructure (PostgreSQL, Redis, RabbitMQ, Jaeger, MinIO)
docker compose up -d postgres redis rabbitmq jaeger minio

# 3. Run the migrations
make migrate

# 4. Generate the stubs from the .proto contracts
make proto

# 5. Start all services (gateway, engines, collab, workers)
docker compose up -d

# 6. Create a development tenant and user (seed)
make seed

# 7. Get a test token
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -d '{"email":"dev@local","password":"dev"}' | jq -r .token
```

---

## Verification

1. **Gateway**: `curl http://localhost:8080/healthz` → `200`.
2. **Full flow**: with the token, create a design (`POST /api/v1/designs`), publish it (`POST /api/v1/sistemas/{id}/publicar`), and submit a form (`POST /api/v1/formularios`) — see `contracts/api.md`.
3. **Collaboration**: open `ws://localhost:4000/socket` in two clients on the same screen; mutations from one appear on the other; after 5s of silence, the collab log shows a single `SalvarDesign` call.
4. **Traces**: open Jaeger at `http://localhost:16686` and search for the `gateway` service — a trace should span gateway → logic-engine → rabbitmq → worker.
5. **Queues**: RabbitMQ management at `http://localhost:15672` (guest/guest) — `webhooks.disparo` and `notificacoes.envio` queues with DLQs.

```bash
# Run tests specific to this feature
make test                # everything (Go + ExUnit + player)
go test ./services/...   # Go services only
(cd collab && mix test)  # collaboration engine only
(cd player && npm test)  # Headless Player only
```

---

## Environment Variables

| Variable | Example Value | Description |
|----------|-----------------|-----------|
| `DATABASE_URL` | `postgres://app:app@localhost:5432/construtor` | PostgreSQL connection |
| `REDIS_URL` | `redis://localhost:6379/0` | Collaboration snapshots and rate limiting |
| `AMQP_URL` | `amqp://guest:guest@localhost:5672/` | RabbitMQ |
| `JWT_PRIVATE_KEY_PATH` | `.keys/jwt_private.pem` | Token signing (IAM only) |
| `JWT_PUBLIC_KEY_PATH` | `.keys/jwt_public.pem` | Token validation (gateway/services) |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `http://localhost:4317` | Trace export (Collector/Jaeger) |
| `S3_ENDPOINT` | `http://localhost:9000` | Local MinIO (Export Engine) |
| `S3_BUCKET` | `exports` | Export package bucket |
| `PRESIGNED_URL_TTL` | `10m` | Download link expiration |
| `COLLAB_DEBOUNCE_MS` | `5000` | Inactivity window before the gRPC flush (BR06) |
| `GATEWAY_RATE_LIMIT_RPS` | `50` | Requests-per-tenant limit |
