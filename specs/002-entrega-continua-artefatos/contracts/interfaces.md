# Interfaces: CI/CD Pipeline for Compiled Artifact Delivery

This effort does not expose its own HTTP API; the contracts are the **artifact format**, the **host layout**, the **script interface**, and the **secrets** consumed by the pipeline.

---

## 1. Artifact contract (`dist/artifacts/<unit>-<sha>.tar.gz`)

Each tarball contains **only** executable content. Including source, tests, `.git`, `node_modules`, `deps`, or versioned stubs is forbidden (BR01, BR05).

| Unit | Build command | Tarball contents |
|---------|------------------|---------------------|
| `gateway`, `iam`, `design`, `logic`, `deploy`, `export`, `workers` | `CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=<sha>" -o bin/<unit> ./<path>/cmd` | `bin/<unit>` (single static binary) |
| `collab` | `MIX_ENV=prod mix release collab` | OTP release tree (`bin/`, `lib/`, `releases/`, ERTS) |
| `player` | `npm ci && npm run build` | Contents of `dist/` (minified HTML/JS/CSS) |

**Naming**: `<unit>-<sha>.tar.gz`, where `<sha>` is the short git sha (immutable, BR04).

---

## 2. Host layout contract

```
/opt/machv4/
├── releases/
│   ├── <current-sha>/
│   │   ├── bin/{gateway,iam,design,logic,deploy,export,workers}
│   │   ├── collab/           # self-contained OTP release
│   │   └── player/           # static bundle (Nginx docroot)
│   └── <previous-sha>/       # retained for rollback (RELEASES_KEEP)
└── current -> releases/<current-sha>   # symlink swapped atomically (BR04, BR07)
```

- Activation is an atomic `rename` of the `current` symlink; it never points to a partially transferred directory (Criterion 4).
- `releases/` retains the `RELEASES_KEEP` most recent versions.

---

## 3. `systemd` unit contract

Name: `machv4-<unit>.service` for `gateway, iam, design, logic, deploy, export, workers, collab`.

| Field | Value |
|-------|-------|
| `ExecStart` | `/opt/machv4/current/bin/<unit>` (or `/opt/machv4/current/collab/bin/collab start`) |
| `User` | non-root service user (NFR01) |
| `EnvironmentFile` | `/etc/machv4/<unit>.env` (DSN, OTLP, ports — never in the artifact) |
| `Restart` | `on-failure` |

---

## 4. Script interface

```bash
# Compiles and packages all artifacts from dist/ (never from the repo root)
build/build-artifacts.sh
#   input:  SHA (env, default: git rev-parse --short HEAD)
#   output: dist/artifacts/<unit>-<SHA>.tar.gz ; exit 0 = success

# Delivers the artifacts to an environment and activates the release
build/deploy.sh --env <staging|production> --host <host> --user <user> --sha <sha>
#   effect:   rsync -> releases/<sha> ; atomic ln -sfn ; systemctl restart 'machv4-*'

# Checks service health after activation
build/smoke-test.sh --host <host>
#   contract: exit 0 = all healthy ; exit ≠ 0 = failure (triggers rollback, BR08)

# Reverts to the previous release (or a given sha), without recompiling
build/rollback.sh --env <staging|production> --host <host> [--sha <sha>]
```

---

## 5. Secrets and environments contract (GitHub)

| Environment | Secret | Use |
|-------------|--------|-----|
| `staging`, `production` | `SSH_PRIVATE_KEY` | Dedicated per-environment key for rsync/SSH (NFR01) |
| `staging`, `production` | `SSH_HOST` | Deploy target host |
| `staging`, `production` | `SSH_USER` | Non-root service user |
| `staging`, `production` | `SSH_KNOWN_HOSTS` | Host fingerprint (avoids TOFU on the runner) |

- `staging`: triggered automatically on push to `main`.
- `production`: triggered by a **manual trigger** of `cd.yml` (`workflow_dispatch`,
  typically with `--ref vX.Y.Z`) — the trigger is the human gate (BR02, BR03).
  Environment *required reviewers* require a paid plan/public repo; the manual
  trigger fulfills the same role on the free plan.

---

## 6. Smoke test contract (healthchecks)

| Service | Check |
|---------|-------------|
| `gateway` | `GET http://<host>:8080/healthz` → `200` |
| `iam`, `design`, `logic`, `deploy`, `export` | gRPC *health check* on the service port (`50051`–`50055`) |
| `collab` | `GET http://<host>:4000/healthz` (Phoenix) → `200` |
| `player` | `GET http://<host>/` (Nginx serves `index.html`) → `200` |
| `workers` | active `systemd` unit (`systemctl is-active machv4-workers`) |

Any failed check ⇒ `smoke-test.sh` returns ≠ 0 ⇒ automatic rollback (BR08).
