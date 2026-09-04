# Quickstart: CI/CD Pipeline for Compiled Artifact Delivery

Guide to compile the artifacts locally, simulate the delivery, and verify the result.

---

## Prerequisites

- Go 1.26 (`$HOME/.local/go/bin` on PATH), OTP 26.2 + Elixir 1.17.3, Node 20, `buf` 1.42.0, and `protoc-gen-elixir` (escript) — the same toolchains as Phase 11.
- `rsync` and an SSH client.
- For the deploy rehearsal: a target host with `systemd`, Nginx, a non-root service user, and a writable `/opt/machv4` directory; or an equivalent local container/VM.

---

## Steps

```bash
# 1. Compile all artifacts (generates gen/, Go binaries, OTP release, player dist)
#    Output: dist/artifacts/<unit>-<sha>.tar.gz (executables only)
SHA=$(git rev-parse --short HEAD) build/build-artifacts.sh

# 2. Inspect an artifact — it must not contain source, .git, node_modules, or deps
tar -tzf dist/artifacts/gateway-"$SHA".tar.gz | head

# 3. (Rehearsal) Deliver to the staging host: rsync -> releases/<sha> + symlink + restart
build/deploy.sh --env staging --host "$STAGING_HOST" --user deploy --sha "$SHA"

# 4. Post-deploy smoke test (service healthchecks)
build/smoke-test.sh --host "$STAGING_HOST"

# 5. (If needed) Rollback to the previous release, without recompiling
build/rollback.sh --env staging --host "$STAGING_HOST"
```

In CI, the same path is executed by `.github/workflows/cd.yml`: a push to `main` delivers to staging automatically; a `vX.Y.Z` tag compiles/publishes the artifacts, and production is promoted via a manual trigger (`gh workflow run cd.yml --ref vX.Y.Z`).

---

## Verification

- **Only artifacts in production**: on the host, `ls -la /opt/machv4/current/` shows only binaries/`release`/`player`; `find /opt/machv4/current -name '*.go' -o -name '.git'` returns empty (Criterion 1).
- **Atomic activation**: `readlink /opt/machv4/current` points to the just-delivered `releases/<sha>`.
- **Healthy services**: `systemctl is-active 'machv4-*'` and the smoke test's healthchecks return OK.
- **Rollback**: after `rollback.sh`, `readlink /opt/machv4/current` points to the previous sha and the services restart without a new build.

```bash
# Full repository suite (ensures the CI gate stays green)
make test                                   # Go
cd collab && mix test                       # Elixir
cd player && npm test                       # Player
# Integration/E2E: see specs/001-.../quickstart.md (Compose up + integration/e2e tags)
```

---

## Environment Variables

| Variable | Example Value | Description |
|----------|-----------------|-----------|
| `SHA` | `24a4fce` | Short git sha that names the release (BR04) |
| `STAGING_HOST` / `PROD_HOST` | `10.0.1.20` | Deploy target host (via secret in CI) |
| `SSH_USER` | `deploy` | Non-root SSH user on the host (NFR01) |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `otel-collector.internal:4317` | Environment's Collector for the binaries (NFR06) |
| `DATABASE_URL` | `postgres://mach:***@db.internal:5432/machv4` | DSN applied via `EnvironmentFile` on the host |
| `RELEASES_KEEP` | `5` | Number of releases retained for rollback |
