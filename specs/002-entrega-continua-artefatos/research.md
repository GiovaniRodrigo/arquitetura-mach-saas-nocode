# Research: CI/CD Pipeline for Compiled Artifact Delivery

---

## 1. Existing Patterns in the Project

| File/Pattern | Location | Relevance |
|----------------|-------------|-----------|
| Validation pipeline (Phase 11) | `.github/workflows/ci.yml` | Reusable base of the CI *gate*: `proto`/`go`/`elixir`/`player`/`integration` jobs. `cd.yml` consumes it via `workflow_call`. |
| `.proto` stub generation | `Makefile` (`make proto`), `buf.gen.yaml` | The artifact build needs to generate `gen/` before compiling Go/Elixir. `gen/` is gitignored (BR05). |
| Go entrypoints | `gateway/cmd`, `services/{iam,design,logic,deploy,export}/cmd`, `workers/cmd` | The 7 binary units to compile (`go build ./<path>/cmd`). |
| Elixir service | `collab/mix.exs` | Needs a `releases:` config for `mix release` to produce a self-contained OTP package. |
| Player | `player/package.json` (`build`), `player/vite.config.ts` | `vite build` → `player/dist` (minified static bundle). |
| KEDA/k8s manifest | `infra/k8s/keda/scaledobject-workers.yaml` | **Alternative** substrate (container/Kubernetes) — reference for the architecture decision (section 4). Uses `ghcr.io/machv4/workers:latest`, namespace `mach`. |
| OTel instrumentation | Go services + `collab` (Phase 9) | The binaries already export OTLP; the production runtime just needs to point `OTEL_EXPORTER_OTLP_ENDPOINT` at the environment's Collector. |
| Runner toolchains | project memories | Go 1.26 (`$HOME/.local/go`), OTP 26.2/Elixir 1.17.3, `buf` 1.42.0, `protoc-gen-elixir` via escript. |

---

## 2. Technologies and Libraries

| Technology | Version | Use | Already installed? |
|------------|--------|-----|---------------|
| GitHub Actions | — | CI/CD orchestration, *environments*, and manual approval | Yes (Phase 11) |
| `go build` (`CGO_ENABLED=0`) | Go 1.26 | Static binaries, no toolchain in production | Yes (runner) |
| `mix release` | Elixir 1.17.3 / OTP 26.2 | Self-contained OTP release of `collab` | Yes (runner) |
| Vite | 5.x | `vite build` → static `dist/` | Yes (player) |
| rsync over SSH | — | Incremental (delta) transfer of artifacts only | Standard on the Ubuntu runner |
| `webfactory/ssh-agent` (or native `ssh-agent`) | — | Injecting the SSH key into the deploy job | No (to add) |
| systemd | — | Service supervision on the host (restart, EnvironmentFile) | Assumed on the host |
| Nginx | — | Serve the static player + proxy to the gateway | Assumed on the host |

---

## 3. External References

| Reference | URL | What it resolves |
|------------|-----|--------------|
| Deploying Elixir releases | https://hexdocs.pm/mix/Mix.Tasks.Release.html | Configuring a self-contained `mix release` |
| Go — building static binaries | https://pkg.go.dev/cmd/go#hdr-Compile_packages_and_dependencies | `-trimpath`, `-ldflags`, `CGO_ENABLED=0` flags |
| GitHub Environments / workflow_dispatch | https://docs.github.com/actions/deployment/targeting-different-environments | Per-environment secret scoping; production gate via manual trigger (BR03). *Required reviewers* need a paid plan/public repo. |
| rsync deploy pattern | https://rsync.samba.org/documentation.html | Incremental delivery and `--delete` |
| Zero-downtime symlink swap | https://12factor.net (build/release/run) | Build → release → run separation; atomic activation |

---

## 4. Alternatives Considered

### Option A: Compiled artifact delivery (rsync/SSH + systemd) — **Chosen**
- **Pros**: production without a toolchain, source, or `.git` (NFR01); static Go binaries and the OTP release are naturally self-contained; atomic activation/rollback via symlink; simple, no orchestrator. Aligns exactly with the requested flow `[Git] → [Runner] → [Production receives only artifacts]`.
- **Cons**: no native scale-to-zero for the `workers`; a single host tends toward SPOF; schema migration remains decoupled from deploy.
- **Decision**: **Chosen** — meets the core requirement of sending only artifacts.

### Option B: Container images + Kubernetes/GHCR + KEDA
- **Pros**: a KEDA manifest already exists (`scaledobject-workers.yaml`) and `ghcr.io/machv4/*` images; scale-to-zero for the workers; portability.
- **Cons**: the image is the artifact, but the model departs from the "only compiled files via rsync" requirement; requires a cluster, registry, and credentials; higher operational complexity.
- **Decision**: **Rejected** for this effort; remains a valid path specifically for the `workers` (queue-depth-based autoscaling). See the corresponding risk in `plan.md`.

### Option C: Clone the repository on the host and build in production (`git pull` + `go build`/`npm install`)
- **Pros**: trivial to set up.
- **Cons**: violates BR01/NFR01 (source, `.git`, toolchain, and dev-deps in production); non-reproducible builds; larger attack surface.
- **Decision**: **Rejected** — this is exactly the anti-pattern this effort eliminates.

### Option D: Publish the player to object storage/CDN (S3/MinIO) instead of Nginx on the host
- **Pros**: offloads static traffic from the host; caching/CDN; MinIO is already in the stack (Phase 8).
- **Cons**: introduces a second deploy target and CDN/domain configuration.
- **Decision**: **Rejected** as the standard; recorded as a future evolution. The adopted standard serves `dist/` via Nginx on the host.
