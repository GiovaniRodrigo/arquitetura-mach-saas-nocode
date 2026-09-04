# Tasks: Self Sign-up

<!-- Ordered by execution dependency: migration → store (test before code)
     → proto/gRPC (test before code) → Gateway (test before code) →
     Frontend (test before code) → final verification. -->

## Phase 1 — Migration + Store (IAM)

- [x] 1. Create the `0014_add_senha_users.sql` migration (nullable `ADD COLUMN senha_hash` + partial unique index on `email` for `provedor = 'senha'`), per `data-model.md` §4 (`infra/postgres/migrations/0014_add_senha_users.sql`)
- [x] 2. Write tests covering `CriarTenantEUsuarioComSenha` (success: `dono` tenant + `dono` user created; duplicate email: no tenant/user left behind) and `ObterUsuarioPorEmailSenha` (found / not found) — FR04, FR05, BR01, BR02, BR03, BR06, NFR03. Added to `store_integration_test.go` (the pattern already used by the store, requires Postgres — `-tags integration`), not a new `store_test.go` (`services/iam/internal/store/store_integration_test.go`)
- [x] 3. Implement `ErrEmailJaCadastrado`, `CriarTenantEUsuarioComSenha` (pgx transaction, technical decision 4.3 of `plan.md`), and `ObterUsuarioPorEmailSenha` in `store.go`, adjusting the `DB` interface to expose what's needed to open a transaction (`services/iam/internal/store/store.go`)

## Phase 2 — Proto + IAM gRPC

- [x] 4. Add `RegistrarUsuarioRequest/Response`, `AutenticarSenhaRequest/Response`, and the `RegistrarUsuario`/`AutenticarSenha` RPCs to `IAMService` (`proto/construtor/iam/v1/iam.proto`)
- [x] 5. Run `make proto` and confirm that `gen/go`, `gen/elixir`, `gen/ts` were regenerated without breaking the existing RPCs
- [x] 6. Write `grpc_test.go` for `RegistrarUsuario` (success, duplicate email → `AlreadyExists`, short password → `InvalidArgument`) and `AutenticarSenha` (success, nonexistent email and incorrect password → the **same** `Unauthenticated` error, BR04) (`services/iam/internal/server/grpc_test.go`)
- [x] 7. Implement `RegistrarUsuario` and `AutenticarSenha` in `grpc.go` using `bcrypt.GenerateFromPassword`/`bcrypt.CompareHashAndPassword` (cost ≥ 10, password never in clear text — NFR01) and the new store; extend the local `Store` interface with the 2 methods (`services/iam/internal/server/grpc.go`)

## Phase 3 — Gateway

- [x] 8. Write `routes/auth_test.go` for `POST /api/v1/auth/registro` and `POST /api/v1/auth/login` (success 201/200, duplicate email 409, invalid credentials 401 — identical response for nonexistent email vs. wrong password, Acceptance Criterion 4) (`services/gateway/internal/routes/auth_test.go`)
- [x] 9. Implement `routes/auth.go` (`RegistrarUsuario`, `Login` handlers) translating REST→gRPC in the same style as `tenants.go`, mapping `AlreadyExists`→409, `Unauthenticated`→401, `InvalidArgument`→422, and logging failures without including the password in the log (NFR05) (`services/gateway/internal/routes/auth.go`)
- [x] 10. Register `POST /api/v1/auth/registro` and `POST /api/v1/auth/login` in `router.go`, **outside** `r.Group(Auth)` (public), alongside `oauth.Registrar(r)` (`services/gateway/internal/app/router.go`)

## Phase 4 — Frontend

- [x] 11. Write `Register.test.tsx` covering: rendering the form at `/register` (name/email/password/nome_tenant, FR02), successful submission (token saved + navigation to `/dashboard`), and a duplicate-email error (message shown, fields preserved) (`services/frontend/src/auth/Register.test.tsx`)
- [x] 12. Implement `Register.tsx` (FR02): controlled form, `fetch POST /api/v1/auth/registro`, minimal client-side validation (FR03), saves the token via `session.ts` and navigates to `/dashboard` on success (`services/frontend/src/auth/Register.tsx`)
- [x] 13. Update `Login.test.tsx` to cover the new email/password form (success, invalid credentials) and the "Sign up" link pointing to `/register` (`services/frontend/src/auth/Login.test.tsx`)
- [x] 14. Update `Login.tsx`: add an email/password form (`fetch POST /api/v1/auth/login`) and the "Sign up" link, coexisting with the existing OAuth buttons without changing them (FR01, FR06, BR05) (`services/frontend/src/auth/Login.tsx`)
- [x] 15. Register `<Route path="/register" element={<Register />} />` on the public router (`services/frontend/src/main.tsx`)
- [x] 16. Update `Home.test.tsx` to expect the "Try for Free" CTAs to point to `/register` (FR07) (`services/frontend/src/pages/Home/Home.test.tsx`)
- [x] 17. Update `Home.tsx`: swap `to="/login"` for `to="/register"` on the "Try for Free" CTA links (`services/frontend/src/pages/Home/Home.tsx`)

## Phase 5 — Final Verification

- [x] 18. Run the full suite and confirm nothing broke: `go build ./...`, `go vet ./...`, `go test ./...`; `DATABASE_URL=postgres://mach:mach@localhost:5432/machv4?sslmode=disable go test -tags integration -p 1 ./...` — includes re-running the existing OAuth integration tests unmodified, confirming NFR04; `cd services/frontend && npm test && npm run typecheck`
