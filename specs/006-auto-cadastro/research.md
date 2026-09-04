# Research: Self Sign-up

---

## 1. Existing Patterns in the Project

| File/Pattern | Location | Relevance |
|-----------------|--------------|------------|
| `UpsertUsuarioThirdParty` (find-or-create via `INSERT ... ON CONFLICT ... RETURNING`) | `services/iam/internal/store/store.go:142-155` | Direct model for `CriarTenantEUsuarioComSenha`/`ObterUsuarioPorEmailSenha` — same `users` table, same `Scan` pattern for identity fields. |
| `CriarTenant` (generating `chave_blind_index` with `crypto/rand`, 32 bytes) | `services/iam/internal/server/grpc.go:132-153` | Reused as-is for the tenant created during sign-up — same key-generation approach, same `tipo` (here `'dono'` instead of `'cliente'`). |
| REST→gRPC with a narrow interface + `web.JSON` + gRPC→HTTP error mapping | `services/gateway/internal/routes/tenants.go` | Template for the new `routes/auth.go`: same `Cliente` interface style, `writeTenantError`-like helper, typed JSON body. |
| `auth.Issuer.Issue(userID, tenantID, tipo)` — agnostic of the authentication source | `services/iam/auth/jwt.go:47-63` | Reused directly, with no modification: both `AutenticarThirdParty` and the new `RegistrarUsuario`/`AutenticarSenha` call the same `Issue`. |
| Public routes registered outside `r.Group(Auth)` | `services/gateway/internal/app/router.go:24,33-34` (`oauth.Registrar(r)`) | Same position in the router for the two new routes — they need to be reachable without a JWT. |
| Idempotent migrations (`IF NOT EXISTS` / `ON CONFLICT ... DO NOTHING`) | `infra/postgres/migrations/0012`, `0013` | Style replicated in `0014` (`ADD COLUMN IF NOT EXISTS`, `CREATE UNIQUE INDEX IF NOT EXISTS`). |
| Public router in `main.tsx` (`/login` → `Login`, `*` → `Home`, outside `AppProvider`) | `services/frontend/src/main.tsx:37-46` | Where the `/register` route is added — same sessionless `BrowserRouter`. |
| `/api` proxy already configured in the Vite dev server | `services/frontend/vite.config.ts` (`server.proxy["/api"]`) | Reason for placing the new endpoints under `/api/v1/auth/*` instead of `/auth/*` (technical decision 4.2 of `plan.md`) — avoids CORS/a new proxy in dev. |
| Nginx already proxies `/api/` and `/auth/` to the Gateway | `infra/nginx/*.conf:49-59` | Confirms that in production both prefixes already reach the Gateway; choosing `/api/v1/auth/*` is only for symmetry with the dev server, not a production requirement. |

---

## 2. Technologies and Libraries

| Technology | Version | Use | Already installed? |
|------------|--------|-----|----------------|
| `golang.org/x/crypto/bcrypt` | v0.51.0 (via `golang.org/x/crypto`, `go.mod:47`) | Password hashing (BR03, NFR01) | Yes — currently indirect; becomes a direct import in `services/iam` |
| `pgx/v5` transactions (`pool.Begin`/`tx.Commit`/`tx.Rollback`) | already used in the module (`jackc/pgx/v5`) | Tenant+user atomicity (NFR03, decision 4.3) | Yes — only requires exposing `Begin` on `Store` (see Risk in `plan.md`) |
| `react-router-dom` | already used (`BrowserRouter`, `Route`) | New `/register` route | Yes |

No new dependency to install in any of the three layers.

---

## 3. Architecture/Design Alternatives Considered

### Option A: Unify `RegistrarUsuario` and `AutenticarSenha` into a single RPC with automatic upsert
- **Pros**: one fewer RPC in the proto.
- **Cons**: sign-up and login have incompatible error semantics (409 duplicate
  email vs. 401 invalid credentials, BR04), and sign-up requires fields that
  login doesn't have (`nome`, `nome_tenant`) — mixing the two contracts would break
  the login's generic-error-message guarantee and would make the RPC ambiguous about
  which error to expect.
- **Decision**: Rejected. Two dedicated RPCs, following the convention already used in
  the rest of `iam.proto` (e.g., `ObterTenant`/`AtualizarTenant`/`ExcluirTenant`
  were added as separate RPCs from `ListarTenants`/`CriarTenant`, not
  merged into a generic one).

### Option B: Store the password in a separate `credenciais_senha` table (FK to `users`) instead of a column on `users`
- **Pros**: keeps `users` "clean" for OAuth-only users; isolates the most
  sensitive data (the hash) in a smaller table.
- **Cons**: requires an extra `JOIN` on every password-login flow; the
  unique email index (BR02) would also have to move to that table, duplicating
  the uniqueness logic that already lives in `users`; one more migration and one more
  store method for a still-small domain (a single nullable column).
- **Decision**: Rejected. A nullable column on `users` is simpler, requires no new
  `JOIN`, and the current domain's volume/sensitivity does not justify the split —
  it can be revisited if the authentication model grows (e.g., MFA, spec 004
  FR15, which already reserves space under Settings/Security).

### Option C: Create the tenant from an asynchronous job (RabbitMQ/Workers) instead of synchronously within the sign-up call itself
- **Pros**: keeps the sign-up response fast even if tenant creation
  involves more expensive steps in the future (e.g., provisioning external resources).
- **Cons**: sign-up needs to return the JWT (already carrying `tenant_id`) in the same
  response to authenticate the user immediately (FR04, Acceptance Criterion
  8) — an asynchronous flow would force a polling/callback step just to log in
  the very user who just signed up, a disproportionate amount of complexity
  for an operation that today is just 2 `INSERT`s.
- **Decision**: Rejected. Synchronous, within the same transaction (technical decision
  4.3 of `plan.md`).
