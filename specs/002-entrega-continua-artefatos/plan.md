# Implementation Plan: CI/CD Pipeline for Compiled Artifact Delivery

The strategy reuses the Phase 11 validation pipeline (`.github/workflows/ci.yml`) as a *gate* and adds a release pipeline (`cd.yml`) that compiles the artifacts on the runner, packages only the executable content, and delivers it to the host via rsync/SSH with atomic activation via symlink. Production now runs the services under `systemd` (Go binaries and the OTP release of `collab`) and the player under Nginx, from release directories versioned by the git sha. All compilation and delivery logic lives in idempotent scripts (`build/`) invoked both by CI and locally, guaranteeing parity.

---

## 1. Files to Create/Edit

### 1.1. Pipeline (GitHub Actions)

* **`.github/workflows/ci.yml`** *(edit)*: keep the validation jobs (Phase 11) and expose them as a reusable *gate* (`workflow_call`) for the release pipeline to consume. [FR01, FR10]
* **`.github/workflows/cd.yml`** *(create)*: triggers on push to `main` (→ staging) and on `v*` tags (→ production); invokes CI as a *gate*, compiles the artifacts, packages, publishes, and delivers. Uses the `staging` and `production` *environments* (the latter with protection/approval). [FR02–FR08, FR11, BR02, BR03]

### 1.2. Artifact compilation

* **`build/build-artifacts.sh`** *(create)*: generates `gen/` (`buf generate`), compiles the 7 Go binaries (`CGO_ENABLED=0 go build`), the OTP release (`mix release`), and the `player/dist` (`vite build`); packages each into `dist/artifacts/<unit>-<sha>.tar.gz` containing only the executable. [FR02, FR03, BR01, BR05, BR06]
* **`collab/mix.exs`** *(edit)*: add the `releases:` configuration (name `collab`, `include_executables_for: [:unix]`) so `mix release` produces a self-contained OTP package with ERTS. [FR02]
* **`collab/rel/env.sh.eex`** *(create)*: release runtime environment variables (port, OTel endpoint, Redis, gRPC DSN). [FR02]

### 1.3. Delivery (CD)

* **`build/deploy.sh`** *(create)*: receives host/user/environment and the artifacts directory; runs `rsync -a --delete` to `releases/<sha>`, extracts the tarballs, atomically swaps the `current` symlink (`ln -sfn`), restarts the `systemd` units, and serves the player. Idempotent. [FR07, FR08, BR01, BR04]
* **`build/smoke-test.sh`** *(create)*: queries the healthchecks of each service after activation; a non-zero exit code signals failure. [FR11]
* **`build/rollback.sh`** *(create)*: repoints `current` to the immediately previous release (or a given sha) and restarts the services, without recompiling. [FR09, BR07, BR08]

### 1.4. Production runtime

* **`infra/systemd/machv4-gateway.service`** and the rest (`iam`, `design`, `logic`, `deploy`, `export`, `workers`, `collab`) *(create)*: `systemd` units pointing to `/opt/machv4/current/bin/<unit>`, running as a non-root user, with `Restart=on-failure` and environment variables via `EnvironmentFile`. [FR08, NFR01]
* **`infra/nginx/machv4.conf`** *(create)*: serves the static `player` from `/opt/machv4/current/player` and reverse-proxies to the `gateway` (`:8080`). [FR08]
* **`infra/deploy/README.md`** *(create)*: host layout (`/opt/machv4/{releases,current}`), service user, prerequisites. [NFR01, NFR04]

### 1.5. Secrets and environments

* **GitHub Environments `staging` and `production`** *(configure)*: secrets `SSH_PRIVATE_KEY`, `SSH_HOST`, `SSH_USER`, `SSH_KNOWN_HOSTS`. The production gate is the **manual trigger** of `cd.yml` (`workflow_dispatch`), since *required reviewers* require a paid plan/public repo. [FR06, BR03, NFR01]

---

## 2. Technical Strategy

### 2.1. Self-contained artifacts, toolchain-free production

Each unit becomes a package that **runs without the build environment**:

```bash
# Go — static binary, no host libc, no toolchain in production (BR06)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -trimpath -ldflags "-s -w -X main.version=$SHA" \
  -o dist/bin/gateway ./gateway/cmd

# Elixir — self-contained OTP release (bundles ERTS + compiled BEAM, no mix/deps)
MIX_ENV=prod mix release collab   # -> _build/prod/rel/collab

# Player — minified static bundle (no node_modules)
npm ci && npm run build           # -> player/dist
```

Packaging includes **only** each build's output directory; `build-artifacts.sh` assembles the tarballs from `dist/`, never from the repo root, guaranteeing BR01.

### 2.2. Atomic activation via symlink and rollback without recompilation

The host keeps `/opt/machv4/releases/<sha>/` and a `current` symlink. Activation swaps the alias in a single atomic operation; rollback is the same swap pointed at a different sha:

```bash
# activation (deploy.sh)
ln -sfn "/opt/machv4/releases/$SHA" /opt/machv4/current.tmp
mv -Tf /opt/machv4/current.tmp /opt/machv4/current     # atomic rename
systemctl restart 'machv4-*.service'

# rollback (rollback.sh) — points to the previous release, no build (BR07)
ln -sfn "/opt/machv4/releases/$PREV" /opt/machv4/current.tmp
mv -Tf /opt/machv4/current.tmp /opt/machv4/current
systemctl restart 'machv4-*.service'
```

`releases/` retains the N most recent versions (cleanup at the end of deploy), enabling immediate rollback.

### 2.3. Strict separation CI → artifact → CD

`ci.yml` only **validates** (gate). `cd.yml` only enters the build/delivery phase if the gate passed (`needs`/`workflow_call`), materializing the flow `[Git] → [Runner compiles] → [Production receives artifacts]`. The runner is the only point with a toolchain, dev dependencies, and build secrets; the production host only receives tarballs via rsync and never clones the repository.

### 2.4. Smoke test with automatic rollback

After restarting the services, `smoke-test.sh` validates the healthchecks; any failure triggers `rollback.sh` in the same job, restoring the previous release (BR08) before marking the deploy as failed.

---

## 3. Dependencies and Prerequisites

- [ ] Staging and production host(s) provisioned: non-root service user, `systemd`, `rsync`, Nginx, and the `/opt/machv4` directory with the right permissions. (Out of scope — prerequisite.)
- [ ] Database migration strategy defined: migrations (`infra/postgres/migrations/`) must be applied before activating a release that depends on a new schema. (Prerequisite; not automated in this effort.)
- [ ] OTel Collector reachable from the host (endpoint per environment). [NFR06]
- [ ] SSH secrets configured in the GitHub Environments `staging` and `production`. [NFR01]
- [ ] Toolchains on the runner: Go 1.26, OTP 26.2/Elixir 1.17.3, Node 20, `buf` 1.42.0 (already used in Phase 11's `ci.yml`).

---

## 4. Risks and Points of Attention

| Risk | Impact | Mitigation |
|-------|---------|-----------|
| The per-artifact/systemd model loses KEDA's scale-to-zero for the `workers` (existing k8s manifest). | High | Run `workers` as an always-on `systemd` service with a fixed replica **or** keep `workers` on the container/KEDA substrate and apply the per-artifact model only to the rest. Decision documented in `research.md`. |
| Coupling between deploy and schema migration can produce a release incompatible with the database. | High | Adopt *backward-compatible* migrations (expand/contract); apply the migration before deploy in the `quickstart.md` runbook. |
| A single production host is a single point of failure (SPOF). | Medium | Document as a limitation; the rsync/symlink approach is replicable to multiple hosts in a future iteration. |
| Leakage of a dev artifact into production due to a misconfigured script. | High | `build-artifacts.sh` assembles tarballs only from `dist/`; the acceptance test inspects the host (criterion 1). |
| A compromised SSH key grants access to the production host. | High | Dedicated key per environment, minimal scope, stored in a protected GitHub Environment; `production` requires approval (BR03). |
