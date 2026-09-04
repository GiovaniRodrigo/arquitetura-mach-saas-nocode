# Research: AI and Business Rules Restructuring

---

## 1. Existing Patterns in the Project (reuse, don't duplicate)

| Pattern | Location | Relevance |
|--------|-------------|-----------|
| Data hook with `carregando/pronto/vazio/erro` states + `recarregar()` | `player/src/systems/useSistemas.ts`, `player/src/dashboard/useMetricas.ts` | Direct model for `useUltimosAcessos`, `useFeedback`, `useResumoFinanceiro`, `useTenants` (§2.2 of `plan.md`) |
| UI state components (`Skeleton`/`EmptyState`/`ErrorState`, `aria-busy`/`role="alert"`) | `player/src/components/ui/StateViews.tsx` | Reuse across all new Dashboard cards and the Clients/Help lists (NFR05) |
| Injected identity context + client | `player/src/app/AppContext.tsx` | Already exposes `usuario`/`client` to any screen under `DashboardLayout`; the new screens (Clients, Settings, Profile, Help) consume the same context without creating a new one |
| Reading JWT claims for display only | `player/src/auth/jwt.ts` (`usuarioDe`, `podeCriarSistema`) | BR10 from 003 (hide the creation CTA without permission) is the same rationale to apply to Clients/White Label — authorization is never decided on the frontend |
| Persistent theme without a flash (FOUC) | `player/src/theme/ThemeProvider.tsx`, `theme/initTheme.ts` | Preserve as-is; Settings keeps the "Appearance" section unchanged |
| `ApiClient` tolerant of non-JSON responses (`ApiError`, `parseJsonSeguro`) | `player/src/api/client.ts` | Every new client method (§1 of `plan.md`) inherits this handling — no new `try/catch` around `JSON.parse` should be hand-written |
| Simple navigation/mutation action (open system) | `player/src/systems/abrirSistema.ts` | Model for `publicarVersao`/`reverterVersao` (FR12) |
| Avatar menu with click-outside close | `player/src/layout/DashboardLayout.tsx` (lines 34–42) | Reuse the same `useRef`/`mousedown` pattern if Settings or Profile need similar menus (e.g., status selector in Feedback) |
| Behavior test with mocked `fetch` | `player/src/pages/Dashboard/Projects.test.tsx`, `Overview.test.tsx` (cited in `specs/003/tasks.md`) | Model for the new `Dashboard.test.tsx`, `Clientes.test.tsx` tests |
| Multi-tenant isolation in the backend (Go) | `pkg/database.ScopedDB.WithTenant`, `pkg/tenantctx` ([[machv4-verified-workflow]]) | Any new `contracts/api.md` endpoint implemented in the Gateway/services must go through this same mechanism — BR01 in this spec is the same BR01 from 001 |

---

## 2. Technologies and Libraries

| Technology | Version | Use | Already installed? |
|------------|--------|-----|---------------|
| React Router | (already in use, `react-router-dom`) | Nested routes `clientes/:tenantId/sistemas/:sistemaId/{telas,regras,versao}` | Yes |
| vitest + jsdom | (already in use) | Behavior tests for the new screens | Yes |
| lucide-react | (already in use) | Icons for the new sidebar items (Help, Registration/Profile) | Yes |
| QR code library (e.g., `qrcode` or generation via a `data:` URI on the backend) | — | Displaying the TOTP QR code when enabling MFA (FR15) | **No** — decision pending (product/backend may return a ready-made image) |
| Color-picker component / file upload for the logo | — | White Label (FR13) | **No** — no library chosen yet; see `plan.md §3` |

---

## 3. External References

| Reference | What it addresses |
|------------|---------------|
| RFC 6238 (TOTP) | Format of the MFA secret/code (NFR01) |
| Custom-domain convention via a DNS TXT record | Assumed mechanism for the White Label's `dominio_validado` (NFR03) — actual implementation out of scope for this spec |

---

## 4. Alternatives Considered

### Option A: Implement the Screens-tab canvas (FR09) as part of this same initiative
- **Pros**: delivers the full Clients flow in one shot.
- **Cons**: it is a full visual editor (drag-and-drop, selection, component tree) — several weeks of effort, incompatible with ≤1-day atomic tasks; `001-construtor-sistemas-mach-v4 §8` had already isolated this item as its own initiative.
- **Decision**: Rejected for this spec — only the navigation shell is delivered (see `plan.md §2.3` and Risks).

### Option B: Block all of Phase 2 (Dashboard/Settings) until the backend exposes the `contracts/api.md` endpoints
- **Pros**: avoids rework if the contract changes.
- **Cons**: repeats the same blocker that `specs/003` deliberately avoided (it assumed fields "to be provided by the Gateway" and proceeded with graceful degradation).
- **Decision**: Rejected — follow the precedent set by 003, building the UI against the assumed contract with tests using mocked `fetch`.
