# Quickstart: Self Sign-up

Guide to running and testing this implementation locally.

---

## Prerequisites

- The local stack running (`make dev` or `make dev-no-frontend` — see `build/dev-up.sh`).
- Migration `0014_add_senha_users.sql` applied (`make migrate` already runs all pending migrations in order).
- Contracts regenerated after changing `iam.proto` (`make proto`).

---

## Steps

```bash
# 1. Apply the new migration (if the stack was already up before it existed)
make migrate

# 2. Regenerate the gRPC contracts (gen/go, gen/elixir, gen/ts) after editing iam.proto
make proto

# 3. Bring up the stack (infra + services + gateway + collab [+ frontend])
make dev

# 4. In the browser: http://localhost:5183 (or whichever port Vite reports)
#    → click "Try for Free" → fill in name/email/password/business name
```

---

## Verification

```bash
# Smoke test via cURL (Gateway on :8080)
curl -i -X POST http://localhost:8080/api/v1/auth/registro \
  -H 'Content-Type: application/json' \
  -d '{"nome":"Ana","email":"ana@example.com","senha":"12345678","nome_tenant":"Ana LTDA"}'

curl -i -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"ana@example.com","senha":"12345678"}'

# Repeating the sign-up with the same email should return 409:
curl -i -X POST http://localhost:8080/api/v1/auth/registro \
  -H 'Content-Type: application/json' \
  -d '{"nome":"Ana","email":"ana@example.com","senha":"12345678","nome_tenant":"Outra"}'

# Tests for this initiative
go test ./services/iam/... ./services/gateway/...

DATABASE_URL=postgres://mach:mach@localhost:5432/machv4?sslmode=disable \
  go test -tags integration -p 1 ./services/gateway/... ./services/iam/...

cd services/frontend && npm test -- Register Login Home && npm run typecheck
```

Quick sanity criterion: the second call to `/api/v1/auth/registro` (same
email) should return `409` **and** `SELECT count(*) FROM tenants` should not have
increased relative to the first call — this confirms atomicity (NFR03).
