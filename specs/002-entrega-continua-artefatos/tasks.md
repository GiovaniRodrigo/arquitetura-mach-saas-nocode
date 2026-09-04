# Tasks: CI/CD Pipeline for Compiled Artifact Delivery

<!-- Ordered by execution dependency. Each task is atomic (≤ 1 day). -->

- [x] 1. Refactor the validation pipeline to be reusable as a *gate* via `workflow_call`, preserving the proto/go/elixir/player/integration jobs (`.github/workflows/ci.yml`) [FR01, FR10]
- [x] 2. Add the `releases:` configuration (name `collab`, `include_executables_for: [:unix]`) for `mix release` (`collab/mix.exs`) [FR02]
- [x] 3. Create the Elixir release runtime variables (port, OTLP, Redis, gRPC DSN) (`collab/rel/env.sh.eex`) [FR02, NFR06]
- [x] 4. Write the artifact compilation script: `buf generate`, 7 `CGO_ENABLED=0` Go binaries, `mix release`, `vite build`, packaging only `dist/` into tarballs by sha (`build/build-artifacts.sh`) [FR02, FR03, BR01, BR05, BR06]
- [x] 5. Create the non-root `systemd` units for the 8 services, pointing to `/opt/machv4/current` with `EnvironmentFile` (`infra/systemd/machv4-*.service`) [FR08, NFR01]
- [x] 6. Create the Nginx configuration that serves the static player and proxies to the gateway (`infra/nginx/machv4.conf`) [FR08]
- [x] 7. Document the host layout and prerequisites (`/opt/machv4/{releases,current}`, service user) (`infra/deploy/README.md`) [NFR01, NFR04]
- [x] 8. Write the delivery script: rsync the artifacts to `releases/<sha>`, extraction, atomic symlink swap, and service restart (`build/deploy.sh`) [FR07, FR08, BR01, BR04]
- [x] 9. Write the post-deploy smoke test (healthcheck for each service, exit ≠ 0 on failure) (`build/smoke-test.sh`) [FR11]
- [x] 10. Write the rollback script (repoint `current` to the previous release/a given sha + restart, no build) (`build/rollback.sh`) [FR09, BR07, BR08]
- [x] 11. Create the release pipeline: CI gate, build/packaging, artifact publishing, and deploy to staging (push to `main`) and production (`v*` tag, with approval) (`.github/workflows/cd.yml`) [FR04, FR05, FR06, BR02, BR03]
- [x] 12. Chain the smoke test with automatic rollback in the deploy job (`.github/workflows/cd.yml`, `build/smoke-test.sh`, `build/rollback.sh`) [FR11, BR08]
- [x] 13. Configure the GitHub Environments `staging` and `production` (SSH secrets; production gate via manual `workflow_dispatch` trigger, since *required reviewers* require a paid plan) — document in (`infra/deploy/README.md`) [FR06, BR03, NFR01]
- [x] 14. Validate the build locally: `build/build-artifacts.sh` produces tarballs with executables only; inspect for the absence of source/`.git`/`node_modules`/`deps` (`build/build-artifacts.sh`) [Criterion 1, BR01]
- [x] 15. Rehearse the end-to-end flow on a staging host (deploy → smoke → rollback) and fix regressions (`build/deploy.sh`, `build/rollback.sh`, `.github/workflows/cd.yml`) [Criteria 2–7]
- [x] 16. Run the repository's full test suite (`make test` + Elixir + player + integration/E2E) ensuring the `ci.yml` refactor caused no regressions
