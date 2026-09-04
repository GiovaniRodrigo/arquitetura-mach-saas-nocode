# Artifact-based deploy — host layout and prerequisites (spec 002)

This document describes the target of continuous delivery: a Linux host with `systemd` and
Nginx that receives **only compiled artifacts** (Go binaries, `collab`'s OTP
release, `frontend`'s static bundle) via `rsync`/SSH. The host never clones the
repository and has no build toolchain (NFR01, BR01).

## 1. Filesystem layout

```
/opt/machv4/
├── releases/
│   ├── <sha>/                # one directory per release, immutable (BR04)
│   │   ├── bin/{gateway,iam,design,logic,deploy,export,workers}
│   │   ├── collab/           # self-contained OTP release (bin/, lib/, releases/, erts-*)
│   │   └── frontend/         # static dist (Nginx docroot)
│   └── <previous-sha>/       # kept for rollback (RELEASES_KEEP, default 5)
└── current -> releases/<sha> # symlink swapped atomically on activation (BR04, BR07)

/etc/machv4/                  # per-service EnvironmentFiles (secrets; outside the artifact)
├── gateway.env  iam.env  design.env  logic.env  deploy.env  export.env
├── workers.env
└── collab.env
```

## 2. Host prerequisites (provisioning — out of scope for this task)

The idempotent [`provision-host.sh`](./provision-host.sh) script creates and configures
everything below. Run it as root **on each host** (staging and production), passing
the CI public key corresponding to the environment:

```bash
sudo ./infra/deploy/provision-host.sh --pubkey ~/ci_machv4_staging.pub
```

It performs, in a re-runnable way:

- Non-root service user `machv4` (owner of `/opt/machv4`).
- SSH deploy user (e.g., `deploy`) member of the `machv4` group, with write access to
  `/opt/machv4` (dir `2775`/setgid) and authorization to `systemctl restart 'machv4-*'`
  via `sudo` without a password (restricted `sudoers` rule — `/etc/sudoers.d/machv4-deploy`,
  validated with `visudo -cf`).
- CI public key authorized in `~deploy/.ssh/authorized_keys` (via `--pubkey`).
- `/etc/machv4/` (0750) with a `<service>.env` stub per unit (to be filled in — §3).
- `infra/systemd/*.service` units installed in `/etc/systemd/system/` and
  enabled (without `start` — only after the 1st deploy creates `current`).
- Nginx pointing to `infra/nginx/machv4.conf` (`sites-enabled/` or `conf.d/`),
  with `nginx -t` + reload.

Flags: `--service-user`, `--deploy-user`, `--base`, `--env-dir`, `--no-systemd`,
`--no-nginx`, `--repo-dir`, `--help`. OS prerequisites that are **not** created
by the script: `systemd`, `rsync`, `tar`, Nginx installed, and the OTel Collector
reachable (endpoint in each `*.env`).

## 3. EnvironmentFiles (`/etc/machv4/<service>.env`)

Each file loads secrets/endpoints at runtime — **never** versioned or
included in the artifact. Minimal examples:

```ini
# /etc/machv4/gateway.env
GATEWAY_ADDR=:8080
IAM_GRPC_ADDR=127.0.0.1:50051
DESIGN_GRPC_ADDR=127.0.0.1:50052
LOGIC_GRPC_ADDR=127.0.0.1:50053
OTEL_EXPORTER_OTLP_ENDPOINT=otel-collector.internal:4317

# /etc/machv4/collab.env
SECRET_KEY_BASE=<generate with: mix phx.gen.secret>
PHX_HOST=app.exemplo.com
PORT=4000
REDIS_URL=redis://127.0.0.1:6379
DESIGN_GRPC_ADDR=127.0.0.1:50052
OTEL_EXPORTER_OTLP_ENDPOINT=otel-collector.internal:4317
```

## 4. GitHub Environments and secrets (task 13)

Configure two *environments* in the repository (**Settings → Environments**):

| Environment | Triggered by | Gate |
|-------------|--------------|------|
| `staging` | push to `main` | — (automatic deploy) |
| `production` | **manual trigger** (`workflow_dispatch`) | the act of triggering itself is the human approval — BR03 |

> Note: environment *required reviewers* require a public repo or a paid plan
> (Pro/Team/Enterprise). On this private repo on the free plan, the production
> gate is the **manual trigger** of `cd.yml`. To promote a tag:
>
> ```bash
> gh workflow run cd.yml --ref vX.Y.Z
> ```

Secrets per environment (same names; different values):

| Secret | Description |
|--------|-----------|
| `SSH_PRIVATE_KEY` | Private key dedicated to the environment (pair registered in the deploy user's `authorized_keys`) |
| `SSH_HOST` | Target host/IP |
| `SSH_USER` | Deploy user (e.g., `deploy`) |
| `SSH_KNOWN_HOSTS` | Output of `ssh-keyscan <host>` (avoids TOFU on the runner) |

Principles: one key per environment, minimal scope, `production` always behind
approval. The runner is the only place with build toolchain and secrets; the host
only receives the tarballs.

## 5. Manual operation

```bash
# Deploy of a specific sha (normally done by cd.yml)
build/deploy.sh --env staging --host "$SSH_HOST" --user deploy --sha <sha>

# Rollback to the previous release
build/rollback.sh --env production --host "$SSH_HOST" --user deploy
```
