# Quickstart: Pipeline CI/CD por Entrega de Artefatos Compilados

Guia para compilar os artefatos localmente, simular a entrega e verificar o resultado.

---

## Pré-requisitos

- Go 1.26 (`$HOME/.local/go/bin` no PATH), OTP 26.2 + Elixir 1.17.3, Node 20, `buf` 1.42.0 e `protoc-gen-elixir` (escript) — mesmos toolchains da Fase 11.
- `rsync` e um cliente SSH.
- Para o ensaio de deploy: um host alvo com `systemd`, Nginx, usuário de serviço non-root e o diretório `/opt/machv4` gravável; ou um container/VM local equivalente.

---

## Passos

```bash
# 1. Compilar todos os artefatos (gera gen/, binários Go, release OTP, dist do player)
#    Saída: dist/artifacts/<unidade>-<sha>.tar.gz (somente executáveis)
SHA=$(git rev-parse --short HEAD) scripts/build-artifacts.sh

# 2. Inspecionar um artefato — não deve conter fonte, .git, node_modules nem deps
tar -tzf dist/artifacts/gateway-"$SHA".tar.gz | head

# 3. (Ensaio) Entregar ao host de staging: rsync -> releases/<sha> + symlink + restart
scripts/deploy.sh --env staging --host "$STAGING_HOST" --user deploy --sha "$SHA"

# 4. Smoke test pós-deploy (healthchecks dos serviços)
scripts/smoke-test.sh --host "$STAGING_HOST"

# 5. (Se necessário) Rollback para o release anterior, sem recompilar
scripts/rollback.sh --env staging --host "$STAGING_HOST"
```

No CI, o mesmo caminho é executado por `.github/workflows/cd.yml`: push em `main` entrega a staging automaticamente; uma tag `vX.Y.Z` entrega a produção após aprovação manual no *environment* `production`.

---

## Verificação

- **Só artefatos em produção**: no host, `ls -la /opt/machv4/current/` mostra apenas binários/`release`/`player`; `find /opt/machv4/current -name '*.go' -o -name '.git'` retorna vazio (Critério 1).
- **Ativação atômica**: `readlink /opt/machv4/current` aponta ao `releases/<sha>` recém-entregue.
- **Serviços saudáveis**: `systemctl is-active 'machv4-*'` e os healthchecks do smoke test retornam OK.
- **Rollback**: após `rollback.sh`, `readlink /opt/machv4/current` aponta ao sha anterior e os serviços reiniciam sem novo build.

```bash
# Suíte completa do repositório (garante que o gate de CI segue verde)
make test                                   # Go
cd collab && mix test                       # Elixir
cd player && npm test                       # Player
# Integração/E2E: ver specs/001-.../quickstart.md (Compose up + tags integration/e2e)
```

---

## Variáveis de Ambiente

| Variável | Valor de Exemplo | Descrição |
|----------|-----------------|-----------|
| `SHA` | `24a4fce` | Git sha curto que nomeia o release (RN04) |
| `STAGING_HOST` / `PROD_HOST` | `10.0.1.20` | Host alvo do deploy (via secret no CI) |
| `SSH_USER` | `deploy` | Usuário SSH non-root no host (RNF01) |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `otel-collector.internal:4317` | Collector do ambiente para os binários (RNF06) |
| `DATABASE_URL` | `postgres://mach:***@db.internal:5432/machv4` | DSN aplicado via `EnvironmentFile` no host |
| `RELEASES_KEEP` | `5` | Nº de releases retidos para rollback |
