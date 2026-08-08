# Quickstart: Auto Cadastro (Self Sign-up)

Guia para rodar e testar esta implementação localmente.

---

## Pré-requisitos

- Stack local no ar (`make dev` ou `make dev-no-frontend` — ver `build/dev-up.sh`).
- Migração `0014_add_senha_users.sql` aplicada (`make migrate` já roda todas as migrações pendentes em ordem).
- Contratos regenerados após alterar `iam.proto` (`make proto`).

---

## Passos

```bash
# 1. Aplicar a migração nova (se a stack já estava no ar antes dela existir)
make migrate

# 2. Regenerar os contratos gRPC (gen/go, gen/elixir, gen/ts) após editar iam.proto
make proto

# 3. Subir o stack (infra + services + gateway + collab [+ frontend])
make dev

# 4. No navegador: http://localhost:5183 (ou a porta que o Vite reportar)
#    → clicar em "Testar grátis" → preencher nome/e-mail/senha/nome do negócio
```

---

## Verificação

```bash
# Fumaça via cURL (Gateway em :8080)
curl -i -X POST http://localhost:8080/api/v1/auth/registro \
  -H 'Content-Type: application/json' \
  -d '{"nome":"Ana","email":"ana@example.com","senha":"12345678","nome_tenant":"Ana LTDA"}'

curl -i -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"ana@example.com","senha":"12345678"}'

# Repetir o registro com o mesmo e-mail deve devolver 409:
curl -i -X POST http://localhost:8080/api/v1/auth/registro \
  -H 'Content-Type: application/json' \
  -d '{"nome":"Ana","email":"ana@example.com","senha":"12345678","nome_tenant":"Outra"}'

# Testes desta demanda
go test ./services/iam/... ./services/gateway/...

DATABASE_URL=postgres://mach:mach@localhost:5432/machv4?sslmode=disable \
  go test -tags integration -p 1 ./services/gateway/... ./services/iam/...

cd services/frontend && npm test -- Register Login Home && npm run typecheck
```

Critério de sanidade rápida: a segunda chamada de `/api/v1/auth/registro` (mesmo
e-mail) deve devolver `409` **e** `SELECT count(*) FROM tenants` não deve ter
aumentado em relação à primeira chamada — confirma a atomicidade (RNF03).
