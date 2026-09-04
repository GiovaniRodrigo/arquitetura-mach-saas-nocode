# Implementation Plan: AI and Business Rules Restructuring

Work is mostly in `player/` (Vite/React/TS). The strategy separates what is pure
navigation/renaming (does not depend on a new backend) from what needs an API
contract that does not exist yet (mocked at this stage, following the same graceful-degradation
pattern already used in `specs/003-dashboard-integracao-dados`). The **Screens** tab (FR09) and part of
**Business Rules** (FR10/FR11) on the Clients screen are the visual editor (drag-and-drop
canvas) that `001-construtor-sistemas-mach-v4 §8` already marked as its **own
initiative**: today there is no canvas/editor code at all in `player/src` (confirmed
by search) — this plan delivers only the navigation shell up to those tabs, not the
editor itself (see §4 Risks).

---

## 1. Files to Create/Edit

### 1.1. Navigation and renaming (`layout/`, `App.tsx`)

* **`player/src/layout/DashboardLayout.tsx`**: rename sidebar items — "Home" (index `/dashboard`) → **Dashboard** label; "Projects" → **Clients** (`/dashboard/clientes`); "Settings" → **Settings** (`/dashboard/configuracao`); add top-level items **Registration/Profile** (`/dashboard/perfil`, currently nested under `settings/perfil`) and **Help** (`/dashboard/ajuda`). Update the two avatar-menu links (lines 141–153) that currently point to `/dashboard/settings`.
* **`player/src/App.tsx`**: add `clientes`, `clientes/:tenantId`, `clientes/:tenantId/sistemas/:sistemaId` routes (with `telas`/`regras`/`versao` sub-routes), `configuracao`, `ajuda`; move `settings/perfil` to `perfil` (top-level item); add a public `/` or `/home` route for the new Home (outside `AppProvider`/`DashboardLayout`, requiring no session).
* **`player/src/pages/Home/Home.tsx`** (new): public landing page (FR01/FR02), without `AppProvider`.

### 1.2. Dashboard (renaming `Overview` + 3 new cards)

* **`player/src/pages/Dashboard/Overview.tsx`** → rename to **`player/src/pages/Dashboard/Dashboard.tsx`** (update import in `App.tsx`); keep the existing metrics (FR03) and add the 3 new cards.
* **`player/src/dashboard/useUltimosAcessos.ts`** (new): hook modeled on `useSistemas.ts`/`useMetricas.ts` (loading/ready/empty/error states) consuming `client.listarUltimosAcessos()` (FR04, BR02).
* **`player/src/dashboard/useFeedback.ts`** (new): same pattern, consuming `client.listarFeedback()`, with status filter (FR05, BR03).
* **`player/src/dashboard/useResumoFinanceiro.ts`** (new): same pattern, consuming `client.resumoFinanceiro()` (FR06, BR04).
* **`player/src/dashboard/CardUltimosAcessos.tsx`**, **`CardFeedback.tsx`**, **`CardResumoFinanceiro.tsx`** (new): presentational components using the existing `StateViews` (`Skeleton`/`EmptyState`/`ErrorState`).
* **`player/src/api/client.ts`**: add `listarUltimosAcessos()`, `listarFeedback()`, `resumoFinanceiro()` (see `contracts/api.md`).
* **`player/src/api/types.ts`**: add `EventoLogin`, `Feedback`, `ResumoFinanceiro`.

### 1.3. Clients (renaming `Projects` + tenant → system → tabs navigation)

* **`player/src/pages/Dashboard/Projects.tsx`** → rename to **`player/src/pages/Dashboard/Clientes.tsx`**: lists tenants instead of systems directly (FR07); reuses `StateViews` and the `useSistemas` pattern.
* **`player/src/clientes/useTenants.ts`** (new): hook analogous to `useSistemas.ts` for `client.listarTenants()` (FR07).
* **`player/src/pages/Dashboard/ClienteSistemas.tsx`** (new, route `clientes/:tenantId`): lists the systems of the selected tenant, reusing `useSistemas` filtered by `tenantId` (FR08, BR05).
* **`player/src/pages/Dashboard/SistemaAbas.tsx`** (new, route `clientes/:tenantId/sistemas/:sistemaId`): shell with the 3 tabs (Screens/Business Rules/Version) via a nested `react-router-dom` `<Outlet/>`.
* **`player/src/pages/Dashboard/abas/AbaTelas.tsx`** (new): **navigation shell and empty state only** — the infinite canvas itself is outside the atomic scope of this `tasks.md` (see §4).
* **`player/src/pages/Dashboard/abas/AbaRegrasNegocio.tsx`** (new): simple CRUD for single-component validation rules (FR10) — list + form (`blind_index` field, validation type, parameters). Multi-component rules (FR11) remain an empty state/placeholder at this stage (see §4).
* **`player/src/pages/Dashboard/abas/AbaVersao.tsx`** (new): lists versions (reuses `client.versaoAtiva`) and publishes/rolls back (FR12) — reuses the action pattern already used in `abrirSistema.ts`.
* **`player/src/api/client.ts`**: add `listarTenants()`, `listarRegrasNegocio(sistemaId)`, `criarRegraNegocio(...)`, `listarVersoes(sistemaId)`, `publicarVersao(sistemaId)`, `reverterVersao(sistemaId, versaoId)`.

### 1.4. Settings (renaming `Settings` + White Label + Security)

* **`player/src/pages/Dashboard/Settings.tsx`** → rename to **`player/src/pages/Dashboard/Configuracao.tsx`**: keeps "Appearance" (theme, already existing); removes the "User Profile" card (FR17-19 move to the new top-level Registration/Profile screen); adds White Label and Security sections.
* **`player/src/configuracao/WhiteLabelForm.tsx`** (new): logo (upload), colors (color picker), custom domain + validation state (FR13, NFR03).
* **`player/src/configuracao/SegurancaForm.tsx`** (new): three actions — update password, enable/disable MFA (TOTP, with QR code shown once), delete account (blocked if an active tenant exists — FR14-FR16, BR07, NFR01, NFR02).
* **`player/src/api/client.ts`**: add `atualizarWhiteLabel(...)`, `atualizarSenha(...)`, `ativarMfa()`, `confirmarMfa(codigo)`, `desativarMfa()`, `excluirConta()`.

### 1.5. Registration/Profile (promoted to a top-level item)

* **`player/src/pages/Dashboard/Perfil.tsx`**: move from `settings/perfil` to the top-level `perfil` route; add name/photo fields (direct edit) and email (with confirmation flow, FR18/BR08); add a "Change password" link pointing to `/dashboard/configuracao#seguranca`.
* **`player/src/api/client.ts`**: add `atualizarPerfil({nome, foto})`, `solicitarTrocaEmail(novoEmail)`, `confirmarTrocaEmail(token)`.

### 1.6. Help (new)

* **`player/src/pages/Dashboard/Ajuda.tsx`** (new): search (controlled `<input>`) + list of articles by category (FR20/FR21).
* **`player/src/ajuda/artigos.ts`** (new): initial static content (local array), with a signature already prepared to be swapped for `client.buscarArtigos(termo)` once the CMS exists (see Out of Scope in `spec.md`).

---

## 2. Technical Strategy

### 2.1. Phases by backend dependency (same pattern as `specs/003`)

`spec.md` already splits the FRs between those that depend on an API contract that does not
yet exist and those that don't. This plan replicates the **Phase 1 / Phase 2** split from
`specs/003/tasks.md`: Phase 1 delivers everything that only depends on renaming/reorganizing
already-existing routes and screens with static content (Home, Help); Phase 2 delivers the
cards/forms that need new endpoints — implemented **against an assumed contract**
(`contracts/api.md`), with `ApiClient` already isolating that boundary (the same rationale as
the `ApiError`/`parseJsonSeguro` pair that already tolerates unexpected Gateway responses).

### 2.2. Reusing state hooks (NFR05 from 003, applied here)

Every new hook (`useUltimosAcessos`, `useFeedback`, `useResumoFinanceiro`, `useTenants`)
follows the `useSistemas.ts` signature (`carregando | pronto | vazio | erro` state +
`recarregar()`), so that `CardUltimosAcessos`/`CardFeedback`/`CardResumoFinanceiro`
use the same existing `StateViews` without duplicating loading/empty/error logic.

### 2.3. Screens-tab canvas: shell, not editor

The Screens tab (FR09) is described in `spec.md` as an infinite canvas with a screens sidebar
and a properties panel — that is a full visual editor (drag-and-drop, selection,
component-tree manipulation), which does not exist today in `player/src`. Implementing it
as atomic ≤1-day tasks would not be honest: this plan delivers the route, the 3-column
layout (sidebar/canvas/properties), and the empty state, and treats the functional editor as
groundwork for a subsequent dedicated spec (FR09 is only partially covered — see
Risks).

> **Update:** the functional editor was implemented in
> `specs/007-editor-visual-canvas` (real tree, drag & drop, per-fragment
> rich text, free positioning, catalog of 32 components) — FR09 is
> now fully covered there.

---

## 3. Dependencies and Prerequisites

- [ ] API contract defined in `contracts/api.md` (assumed endpoints) reviewed/approved before implementing the Phase 2 hooks.
- [ ] Backend (Gateway/IAM/Design/Logic) exposing the `contracts/api.md` endpoints — nonexistent today; until then, Phase 2 can be developed with mocked `fetch` in tests (same pattern as the current `Overview.test.tsx`/`Projects.test.tsx`).
- [ ] Product decision on which visual component to use for the White Label color picker / logo upload (FR13) — not specified in this initiative.

---

## 4. Risks and Points of Attention

| Risk | Impact | Mitigation |
|-------|---------|-----------|
| FR09 (Screens tab) is a full visual editor, not an ordinary CRUD screen | High — weeks of effort, not 1-day atomic tasks | This plan delivers only the navigation shell (§2.3); recommend opening a dedicated spec for the canvas, analogous to what `001 §8` already anticipated |
| FR11 (multi-component rules) has non-trivial UI modeling (selecting N components + an expression) | Medium | Implement FR10 (single component) first; FR11 remains a placeholder at this stage |
| The `contracts/api.md` endpoints do not exist in the Gateway today | High — Phase 2 has no real data until the backend exposes the endpoints | Follow the `specs/003` pattern: build the UI already prepared for the real contract, with tests using mocked `fetch`; do not block Phase 1 |
| Account deletion (FR16/BR07) and email change (FR18/BR08) touch IAM/authentication, a sensitive security area | High | Require reauthentication (NFR02) on both actions; cover with a block test (active tenant) and a test that the email is not applied before confirmation |
| Renaming existing routes (`/dashboard/projects`, `/dashboard/settings`) breaks links/bookmarks and current tests | Medium | Update `DashboardLayout.test.tsx` and the `Overview`/`Projects`/`Settings` tests in the same renaming task (see `tasks.md` Phase 1) |
