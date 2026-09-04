# Tasks: AI and Business Rules Restructuring

Ordered by execution dependency, spec-kit + TDD pattern (test before the
corresponding implementation, as in `specs/001` and `specs/003`). Each task is
atomic (≤1 day) and references the affected files. Phase 1 does not depend on a new
API contract; Phases 2–5 implement against `contracts/api.md` (assumed — see `plan.md §3`)
with mocked `fetch` in tests until the backend exposes the real endpoints.

## Phase 1 — Renaming the existing IA + Home and Help (no new backend)

- [x] 1. Update `DashboardLayout.test.tsx` to expect the new labels/routes (Dashboard `/dashboard`, Clients `/dashboard/clientes`, Settings `/dashboard/configuracao`, + Registration/Profile `/dashboard/perfil` and Help `/dashboard/ajuda` items) (FR03, FR07, FR13, FR17, FR20) (`player/src/layout/DashboardLayout.test.tsx`)
- [x] 2. Rename the sidebar items and the two avatar-menu links in `DashboardLayout.tsx` to the new labels/routes, adding the Registration/Profile and Help items (FR03, FR07, FR13, FR17, FR20) (`player/src/layout/DashboardLayout.tsx`)
- [x] 3. Rename `Overview.tsx`/`Overview.test.tsx` to `Dashboard.tsx`/`Dashboard.test.tsx` and `Projects.tsx`/`Projects.test.tsx` to `Clientes.tsx`/`Clientes.test.tsx` (rename + import adjustment only; existing behavior preserved) (`player/src/pages/Dashboard/Dashboard.tsx`, `Dashboard.test.tsx`, `Clientes.tsx`, `Clientes.test.tsx`)
- [x] 4. Rename `Settings.tsx`/`Settings.test.tsx` to `Configuracao.tsx`/`Configuracao.test.tsx`, removing the "User Profile" card (moves to Phase 2) (`player/src/pages/Dashboard/Configuracao.tsx`, `Configuracao.test.tsx`)
- [x] 5. Update `App.tsx`: swap imports/routes for `Dashboard`/`Clientes`/`Configuracao`, move `settings/perfil` to the top-level `perfil` route, add empty routes `clientes/:tenantId`, `clientes/:tenantId/sistemas/:sistemaId/*`, `configuracao`, `ajuda` (FR07-FR21) (`player/src/App.tsx`)
- [x] 6. Write a `Home.test.tsx` test covering public rendering (without `AppProvider`) and the presence of the "Sign In"/"Sign Up" CTAs with the correct `href`/route (FR01, FR02) (`player/src/pages/Home/Home.test.tsx`)
- [x] 7. Implement `Home.tsx` (public product-presentation landing page, "Sign In" CTA → login, "Sign Up/Try for Free" CTA → trial flow) and register the public route in `App.tsx` outside `AppProvider` (FR01, FR02) (`player/src/pages/Home/Home.tsx`, `player/src/App.tsx`)
- [x] 8. Write an `Ajuda.test.tsx` test covering article listing by category and filtering by search term (FR20, FR21) (`player/src/pages/Dashboard/Ajuda.test.tsx`)
- [x] 9. Implement `artigos.ts` (initial static content) and `Ajuda.tsx` (search + list by category, `StateViews` for empty) (FR20, FR21, BR09) (`player/src/ajuda/artigos.ts`, `player/src/pages/Dashboard/Ajuda.tsx`)

## Phase 2 — Registration/Profile (name/photo editing + email change with confirmation)

- [x] 10. Write a `client.test.ts` test for `atualizarPerfil`, `solicitarTrocaEmail`, `confirmarTrocaEmail` (payload, headers, `ApiError` handling) (FR17, FR18) (`player/src/api/client.test.ts`)
- [x] 11. Implement `atualizarPerfil`, `solicitarTrocaEmail`, `confirmarTrocaEmail` in `ApiClient` (FR17, FR18) (`player/src/api/client.ts`, `player/src/api/types.ts`)
- [x] 12. Write a `Perfil.test.tsx` test covering: name/photo editing saves directly; changing the email triggers `solicitarTrocaEmail` and shows a "confirm at the new email" notice without changing the displayed email; the "Change password" link navigates to `/dashboard/configuracao#seguranca` (FR17-FR19, BR08) (`player/src/pages/Dashboard/Perfil.test.tsx`)
- [x] 13. Update `Perfil.tsx` (move to the top-level route, name/photo/email fields, confirmation flow, shortcut link to Security) (FR17-FR19, BR08) (`player/src/pages/Dashboard/Perfil.tsx`)

## Phase 3 — Settings: White Label and Security

- [x] 14. Write a `client.test.ts` test for `atualizarWhiteLabel`, `atualizarSenha`, `ativarMfa`, `confirmarMfa`, `desativarMfa`, `excluirConta` — including the `409 TENANT_ATIVO_VINCULADO` case (FR13-FR16, BR07) (`player/src/api/client.test.ts`)
- [x] 15. Implement the above methods in `ApiClient` (FR13-FR16) (`player/src/api/client.ts`, `player/src/api/types.ts`)
- [x] 16. Write a `WhiteLabelForm.test.tsx` test (save logo/colors/domain; show a "validating domain" state when the API responds 202) (FR13, NFR03) (`player/src/configuracao/WhiteLabelForm.test.tsx`)
- [x] 17. Implement `WhiteLabelForm.tsx` (FR13, NFR03) (`player/src/configuracao/WhiteLabelForm.tsx`)
- [x] 18. Write a `SegurancaForm.test.tsx` test covering the 3 flows: password change; two-step MFA activation (QR code shown once, then removed from the DOM); account deletion blocked when the API returns `TENANT_ATIVO_VINCULADO` (FR14-FR16, BR07, NFR01, NFR02) (`player/src/configuracao/SegurancaForm.test.tsx`)
- [x] 19. Implement `SegurancaForm.tsx` (FR14-FR16, BR07, NFR01, NFR02) (`player/src/configuracao/SegurancaForm.tsx`)
- [x] 20. Compose `Configuracao.tsx` with the Appearance (existing) + White Label + Security sections, with a `#seguranca` anchor (FR13-FR16) (`player/src/pages/Dashboard/Configuracao.tsx`)

## Phase 4 — Dashboard: Recent Access, Feedback, and Financial Summary cards

- [x] 21. Write a `client.test.ts` test for `listarUltimosAcessos`, `listarFeedback` (with status filter), `atualizarStatusFeedback`, `resumoFinanceiro` (FR04-FR06) (`player/src/api/client.test.ts`)
- [x] 22. Implement the above methods in `ApiClient` (FR04-FR06) (`player/src/api/client.ts`, `player/src/api/types.ts`)
- [x] 23. Write `useUltimosAcessos.test.ts`, `useFeedback.test.ts`, `useResumoFinanceiro.test.ts` tests covering the 4 states (loading/ready/empty/error) with mocked `fetch`, following the same pattern as `useSistemas.test.ts` (FR04-FR06, NFR05) (`player/src/dashboard/useUltimosAcessos.test.ts`, `useFeedback.test.ts`, `useResumoFinanceiro.test.ts`)
- [x] 24. Implement `useUltimosAcessos.ts`, `useFeedback.ts`, `useResumoFinanceiro.ts` (FR04-FR06, BR02-BR04) (`player/src/dashboard/useUltimosAcessos.ts`, `useFeedback.ts`, `useResumoFinanceiro.ts`)
- [x] 25. Write `CardUltimosAcessos.test.tsx`, `CardFeedback.test.tsx` (including the mark-as-answered action), `CardResumoFinanceiro.test.tsx` tests using `StateViews` (FR04-FR06, NFR05) (`player/src/dashboard/CardUltimosAcessos.test.tsx`, `CardFeedback.test.tsx`, `CardResumoFinanceiro.test.tsx`)
- [x] 26. Implement the 3 card components above (FR04-FR06) (`player/src/dashboard/CardUltimosAcessos.tsx`, `CardFeedback.tsx`, `CardResumoFinanceiro.tsx`)
- [x] 27. Compose `Dashboard.tsx` with the 3 new cards alongside the existing metrics (FR03-FR06, BR01) (`player/src/pages/Dashboard/Dashboard.tsx`)

## Phase 5 — Clients: Tenant → System → tabs navigation

- [x] 28. Write a `client.test.ts` test for `listarTenants`, `listarSistemas` with `tenant_id` filter, `listarRegrasNegocio`/`criarRegraNegocio`, `listarVersoes`/`publicarVersao`/`reverterVersao` (FR07, FR08, FR10, FR12) (`player/src/api/client.test.ts`)
- [x] 29. Implement the above methods in `ApiClient` (FR07, FR08, FR10, FR12) (`player/src/api/client.ts`, `player/src/api/types.ts`)
- [x] 30. Write a `useTenants.test.ts` test (4 states, following the `useSistemas.test.ts` pattern) (FR07) (`player/src/clientes/useTenants.test.ts`)
- [x] 31. Implement `useTenants.ts` and rewrite `Clientes.tsx` to list tenants via `useTenants` + `StateViews`, "Open client" navigating to `clientes/:tenantId` (FR07, BR01) (`player/src/clientes/useTenants.ts`, `player/src/pages/Dashboard/Clientes.tsx`)
- [x] 32. Write a `ClienteSistemas.test.tsx` test (lists the tenant's systems via filtered `useSistemas`; "Open system" navigates to the tabs) (FR08, BR05) (`player/src/pages/Dashboard/ClienteSistemas.test.tsx`)
- [x] 33. Implement `ClienteSistemas.tsx` (FR08, BR05) (`player/src/pages/Dashboard/ClienteSistemas.tsx`)
- [x] 34. Write a `SistemaAbas.test.tsx` test (navigation between the 3 tabs via `<Outlet/>`, active tab highlighted) (FR09-FR12) (`player/src/pages/Dashboard/SistemaAbas.test.tsx`)
- [x] 35. Implement `SistemaAbas.tsx` and the nested `telas`/`regras`/`versao` routes in `App.tsx` (FR09-FR12) (`player/src/pages/Dashboard/SistemaAbas.tsx`, `player/src/App.tsx`)
- [x] 36. Write an `AbaVersao.test.tsx` test (lists versions, publishes, rolls back, reusing the `abrirSistema.ts` pattern) (FR12) (`player/src/pages/Dashboard/abas/AbaVersao.test.tsx`)
- [x] 37. Implement `AbaVersao.tsx` (FR12) (`player/src/pages/Dashboard/abas/AbaVersao.tsx`)
- [x] 38. Write an `AbaRegrasNegocio.test.tsx` test covering creation of a single-component rule (numeric CPF/11 characters as an example) and the empty/placeholder state for a multi-component rule (FR10, BR06) (`player/src/pages/Dashboard/abas/AbaRegrasNegocio.test.tsx`)
- [x] 39. Implement `AbaRegrasNegocio.tsx` — single-component rule CRUD; FR11 (multi-component) as an explicit placeholder ("coming soon"), not as real functionality (FR10, BR06) (`player/src/pages/Dashboard/abas/AbaRegrasNegocio.tsx`)
- [x] 40. Write an `AbaTelas.test.tsx` test covering only the shell (3-column layout renders, "no screens created yet" empty state) — without simulating drag-and-drop, which does not exist at this stage (FR09) (`player/src/pages/Dashboard/abas/AbaTelas.test.tsx`)
- [x] 41. Implement `AbaTelas.tsx` as a navigation shell (empty screens sidebar, central area with a canvas placeholder, empty properties panel) — the functional editor is left for a dedicated spec (see `plan.md §2.3`/Risks) (FR09) (`player/src/pages/Dashboard/abas/AbaTelas.tsx`)
  - Functional editor implemented in `specs/007-editor-visual-canvas` (fully closes FR09).

## Wrap-up

- [x] 42. Run the full suite and the build: `npm run test`, `npm run typecheck`, and `npm run build` must all pass with no errors (`player/`)
