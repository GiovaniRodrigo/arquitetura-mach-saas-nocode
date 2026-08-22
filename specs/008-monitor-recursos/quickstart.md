# Quickstart: Monitor de Recursos

Guia para rodar e testar esta implementação localmente.

---

## Pré-requisitos

- Ambiente já configurado conforme `specs/001-construtor-sistemas-mach-v4/quickstart.md`
  (Go 1.26 em `$HOME/.local/go/bin`, Elixir 1.17 em `$HOME/.local/elixir1.17/bin`, buf,
  Node 20).
- `make proto` executado após criar/alterar os protos desta demanda (task 1).

---

## Passos

```bash
# 1. Gerar código dos novos protos (health.proto, monitor.proto)
make proto

# 2. Subir toda a plataforma, incluindo o novo serviço monitor (após task 32)
./build/dev-up.sh --yes

# 3. Conferir manualmente o endpoint agregado (com token de um usuário autenticado)
curl -s http://localhost:8080/api/v1/monitor/recursos \
  -H "Authorization: Bearer $TOKEN" | jq .

# 4. Simular um serviço fora do ar (ex.: logic) e confirmar que o array continua
#    com 8 entradas, a do logic com status "indisponivel" (RN01/critério 2)
kill $(cat .dev-logs/logic.pid 2>/dev/null || pgrep -f 'services/logic/cmd')
curl -s http://localhost:8080/api/v1/monitor/recursos -H "Authorization: Bearer $TOKEN" | jq .

# 5. Abrir a tela no navegador
#    http://localhost:5173/dashboard/monitor  (ou a porta configurada do Vite)
```

---

## Verificação

```bash
# Go — pkg/health, services/monitor, alterações em iam/design/logic/deploy/export/workers/gateway
go build ./... && go vet ./... && go test ./...

# Elixir — extensão do /healthz
cd services/collab && mix test

# Frontend — hook, cards, página, rota, sidebar
cd services/frontend && npm test && npm run typecheck
```

Critérios de aceitação completos em `spec.md` §6 — os 4 cenários de `spec.md` §5 devem
ser verificados manualmente com a plataforma de pé (parar um serviço individual, depois o
próprio Monitor, para confirmar a diferença entre RN01 e RNF02 descrita na spec).
