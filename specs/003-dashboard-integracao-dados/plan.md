# Implementation Plan: Dashboard — Data Integration and Functionality

The strategy is to evolve the mocked dashboard in vertical increments, starting with
the highest-value data integration (Projects/Overview) and reusing what is already
validated in `SeletorSistemas.tsx`. The system listing/creation logic is extracted
into a shared hook (`useSistemas`), the `ApiClient`/identity is injected via a
React context (today the `client` only exists in `App.tsx`), a `ThemeContext`
is added for dark mode, and the UI states (loading/empty/error) are standardized into
reusable components. Phase 2 features (tenant, Cmd+K, presence, DLQ) are kept isolated
behind flags/optional fields so as not to block Phase 1.

---

## 1. Files to Create/Edit

### 1.1. Application context (client + user)

* **`player/src/app/AppContext.tsx`** (new): provides the `ApiClient` and the authenticated
  user (derived from the JWT) to the whole dashboard tree, avoiding prop drilling. (FR01, FR02, FR03)
* **`player/src/auth/jwt.ts`** (new): decodes the JWT payload (base64url, without
  signature validation — claims read only) to extract name/email/initials. (FR03, NFR06)
* **`player/src/App.tsx`** (edit): wrap the `/dashboard` routes with the `AppContext`
  provider; remove the duplicated generic `<nav>` (lines 67–75) that coexists with the
  sidebar. (FR03, FR-nav)

### 1.2. Data hook and service

* **`player/src/systems/useSistemas.ts`** (new): hook that wraps
  `listarSistemas()`/`criarSistema()` with `loading | ready | empty | error` states
  and a `reload()` action. Single source of truth for `Projects` and `SeletorSistemas`. (FR02, FR06, NFR05)
* **`player/src/systems/SeletorSistemas.tsx`** (edit): refactor to consume
  `useSistemas`, eliminating state duplication. (NFR05)
* **`player/src/dashboard/useMetricas.ts`** (new): hook that derives the
  Overview metrics from the systems (and, once available, from the metrics endpoint). (FR01)

### 1.3. Theme (dark mode)

* **`player/src/theme/ThemeProvider.tsx`** (new): `ThemeContext` + provider that reads/writes
  `mach_theme` in `localStorage` and toggles the `dark` class on `<html>`. (FR05, NFR04)
* **`player/src/theme/initTheme.ts`** (new): synchronous script applied at boot (imported
  early in `main.tsx`) to avoid a theme flash. (NFR04)
* **`player/src/main.tsx`** (edit): call `initTheme()` before render and wrap
  the App in `ThemeProvider`. (FR05, NFR04)

### 1.4. UI state components

* **`player/src/components/ui/StateViews.tsx`** (new): reusable `Skeleton`, `EmptyState`, and
  `ErrorState` (with a retry button), in M3 style. (FR06, NFR03)

### 1.5. Dashboard screens

* **`player/src/pages/Dashboard/Overview.tsx`** (edit): real metrics via
  `useMetricas`; "Get Started" and the FAB trigger creation; loading/empty/error states. (FR01, FR04, FR06)
* **`player/src/pages/Dashboard/Projects.tsx`** (edit): system grid via
  `useSistemas`; functional "Create new project" card and "Open project"; status/version
  per card (FR07); filters/view toggle (FR12, Phase 2). (FR02, FR04, FR06, FR07)
* **`player/src/pages/Dashboard/Settings.tsx`** (edit): "Toggle Theme" wired to the
  `ThemeContext`; "Edit Profile" navigates to a placeholder. (FR05)
* **`player/src/layout/DashboardLayout.tsx`** (edit): real user name/avatar via
  `AppContext`; turn the inert avatar (C7) into a user menu (profile/settings/sign out,
  reusing `encerrarSessao`); (Phase 2) tenant selector, notifications, Cmd+K. (FR03, FR14, FR11, FR13)

> **Control coverage (FR04):** the edits to `Overview`, `Projects`, `Settings`, and
> `DashboardLayout` above must, together, give a real handler to **all**
> controls C1–C7 of the Inventory (spec §2.1) and preserve C8–C10. The system
> creation/opening handlers reuse `criarSistema`/`abrirSistema` (via `useSistemas`), without duplicating logic.

### 1.6. Phase 2 (isolated)

* **`player/src/dashboard/CommandPalette.tsx`** (new, Phase 2): Cmd+K search. (FR10)
* **`player/src/dashboard/PresencaColaboradores.tsx`** (new, Phase 2): stacked avatars
  via `collab/phoenixSocket.ts`. (FR08)
* **`player/src/dashboard/TenantSwitcher.tsx`** (new, Phase 2): tenant selector. (FR11)

### 1.7. Tests

* **`player/src/systems/useSistemas.test.ts`** (new): states via mocked `fetch`. (FR02, FR06)
* **`player/src/theme/ThemeProvider.test.tsx`** (new): toggle + persistence. (FR05)
* **`player/src/auth/jwt.test.ts`** (new): claims extraction. (FR03)
* **`player/src/pages/Dashboard/*.test.tsx`** (edit): replace static-text
  assertions with behavior (states, actions, data). (FR01, FR02, FR04, FR06)

---

## 2. Technical Strategy

### 2.1. Single source of truth for system data (NFR05)

Today `SeletorSistemas` manually implements `listarSistemas` + skeleton + empty + error
+ retry. `Projects` would recreate all of that as a mock. Instead, a hook is extracted:

```ts
// useSistemas.ts (sketch)
type State =
  | { phase: "loading" }
  | { phase: "ready"; sistemas: Sistema[] }
  | { phase: "empty" }
  | { phase: "error"; message: string };

export function useSistemas(client: ApiClient) {
  const [state, setState] = useState<State>({ phase: "loading" });
  const [attempt, setAttempt] = useState(0);
  useEffect(() => { /* listarSistemas → ready | empty | error */ }, [client, attempt]);
  return { state, reload: () => setAttempt(t => t + 1) };
}
```

`SeletorSistemas` and `Projects` now consume the same hook and the same state
components (`StateViews`).

### 2.2. Identity via JWT claims (FR03/NFR06)

The JWT already lives in `localStorage` (`auth/session.ts`). A payload decoder reads
`name`/`email`/`sub` **for display only** — authorization remains on the Gateway. No
signature validation is done on the client.

```ts
// jwt.ts (sketch)
export function readClaims(token: string): { name?: string; email?: string } | null {
  const [, payload] = token.split(".");
  if (!payload) return null;
  try { return JSON.parse(atob(payload.replace(/-/g,'+').replace(/_/g,'/'))); }
  catch { return null; }
}
```

### 2.3. Flash-free theme (FR05/NFR04)

`initTheme()` runs synchronously at boot and applies the `dark` class on `<html>`
from `localStorage` before the first paint; `ThemeProvider` exposes the toggle for
Settings. Tailwind already supports `dark:` via class.

### 2.4. Overview metrics (FR01)

Until a dedicated metrics endpoint exists, `useMetricas` derives counters
from `listarSistemas()` (total, and — once `Sistema` is enriched — active vs.
draft). The metrics contract is specified in `contracts/api.md` for when
the backend exposes it.

### 2.5. Phasing

- **Phase 1 (High)**: FR01–FR06 — integration, actions, theme, states, identity.
- **Phase 2 (Medium/Low)**: FR07–FR13 — status/version per card, presence, DLQ, Cmd+K,
  tenant switcher, filters, notifications. Depend on contract/backend enrichment.

---

## 3. Dependencies and Prerequisites

- [ ] Gateway reachable with a valid JWT (OAuth flow from `auth/session.ts`) or
      `VITE_BYPASS_AUTH=true` for development.
- [ ] `GET /api/v1/sistemas` endpoint operational (already consumed by `SeletorSistemas`).
- [ ] (Phase 2 — FR07/FR09) Enrichment of the `Sistema` payload on the Gateway with
      `status`, `versao_ativa`, and DLQ metrics — see `data-model.md` and `contracts/api.md`.
- [ ] (Phase 2 — FR08) Per-system presence topics available on the
      collaboration server (Phoenix).

---

## 4. Risks and Points of Attention

| Risk | Impact | Mitigation |
|-------|---------|-----------|
| The current `Sistema` only has `id`/`nome`; FR07/FR09 require fields that don't exist. | High | Phasing: Phase 1 does not depend on them; Phase 2 is treated as optional/degradable until the contract is extended. |
| There is no aggregated metrics endpoint for the Overview. | Medium | Derive metrics from `listarSistemas()` in `useMetricas`; specify a future contract. |
| Refactoring `SeletorSistemas` may regress already-validated behavior. | Medium | Extract the hook with tests covering all four states before swapping the screen. |
| Theme flash (FOUC) if the provider applies the theme only after mount. | Medium | Synchronous `initTheme()` at boot, before `createRoot().render`. |
| Reading JWT claims on the client could be mistaken for authorization. | High (security) | Claims used for display only; authorization remains on the Gateway (NFR06). |
| Navigation duplication (`App.tsx`'s `<nav>` + sidebar). | Low | Remove the generic `<nav>` when wrapping the dashboard in the layout. |
