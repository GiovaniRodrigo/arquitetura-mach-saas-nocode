# Quickstart: Resource Monitor

Guide to run and test this implementation locally.

---

## Prerequisites

- Environment already set up per `specs/001-construtor-sistemas-mach-v4/quickstart.md`
  (Go 1.26 at `$HOME/.local/go/bin`, Elixir 1.17 at `$HOME/.local/elixir1.17/bin`, buf,
  Node 20).
- `make proto` run after creating/changing this request's protos (task 1).

---

## Steps

```bash
# 1. Generate code for the new protos (health.proto, monitor.proto)
make proto

# 2. Bring up the whole platform, including the new monitor service (after task 32)
./build/dev-up.sh --yes

# 3. Manually check the aggregated endpoint (with an authenticated user's token)
curl -s http://localhost:8080/api/v1/monitor/recursos \
  -H "Authorization: Bearer $TOKEN" | jq .

# 4. Simulate a service being down (e.g., logic) and confirm the array still
#    has 8 entries, with logic's entry showing status "indisponivel" (BR01/criterion 2)
kill $(cat .dev-logs/logic.pid 2>/dev/null || pgrep -f 'services/logic/cmd')
curl -s http://localhost:8080/api/v1/monitor/recursos -H "Authorization: Bearer $TOKEN" | jq .

# 5. Open the screen in the browser
#    http://localhost:5173/dashboard/monitor  (or the configured Vite port)
```

---

## Verification

```bash
# Go — pkg/health, services/monitor, changes in iam/design/logic/deploy/export/workers/gateway
go build ./... && go vet ./... && go test ./...

# Elixir — the /healthz extension
cd services/collab && mix test

# Frontend — hook, cards, page, route, sidebar
cd services/frontend && npm test && npm run typecheck
```

Full acceptance criteria are in `spec.md` §6 — the 4 scenarios from `spec.md` §5 must
be verified manually with the platform up (stop an individual service, then the
Monitor itself, to confirm the difference between BR01 and NFR02 described in the
spec).
