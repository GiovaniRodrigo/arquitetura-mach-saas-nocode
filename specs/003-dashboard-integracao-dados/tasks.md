# Tasks: Dashboard — Data Integration and Functionality

Ordered by execution dependency. Phase 1 (FR01–FR06) delivers the functional dashboard;
Phase 2 (FR07–FR13) depends on contract/backend enrichment. Each task is atomic
(≤ 1 day) and references the affected files.

## Phase 1 — Integration, actions, theme, states, identity

- [ ] 1. Create reusable UI state components `Skeleton`, `EmptyState`, `ErrorState` in M3 style, with `aria-busy`/`role="alert"` (FR06, NFR03) (`player/src/components/ui/StateViews.tsx`)
- [ ] 2. Write the systems hook test covering all four states (loading/ready/empty/error) with mocked `fetch` (FR02, FR06) (`player/src/systems/useSistemas.test.ts`)
- [ ] 3. Implement the `useSistemas(client)` hook wrapping `listarSistemas`/`criarSistema` + states + `recarregar()` (FR02, FR06, NFR05) (`player/src/systems/useSistemas.ts`)
- [ ] 4. Refactor `SeletorSistemas` to consume `useSistemas` and `StateViews`, removing state duplication (NFR05) (`player/src/systems/SeletorSistemas.tsx`)
- [ ] 5. Write the JWT decoder test (extraction of `name`/`email`/initials; invalid token → `null`) (FR03) (`player/src/auth/jwt.test.ts`)
- [ ] 6. Implement `lerClaims(token)` to extract identity from the JWT (read-only, no signature validation) (FR03, NFR06) (`player/src/auth/jwt.ts`)
- [ ] 7. Create `AppContext` providing `ApiClient` and `UsuarioAutenticado` to the dashboard tree (FR01, FR02, FR03) (`player/src/app/AppContext.tsx`)
- [ ] 8. Wrap the `/dashboard` routes in `AppContext` and remove the duplicated generic `<nav>` at lines 67–75 (FR03, navigation) (`player/src/App.tsx`)
- [ ] 9. Write the `ThemeProvider` test (toggle switches the `dark` class; persists and re-reads `mach_theme`) (FR05, NFR04) (`player/src/theme/ThemeProvider.test.tsx`)
- [ ] 10. Implement `ThemeProvider`/`ThemeContext` (state, toggle, `localStorage`, `dark` class on `<html>`) (FR05) (`player/src/theme/ThemeProvider.tsx`)
- [ ] 11. Implement synchronous `initTheme()` and call it at boot before render; wrap App in `ThemeProvider` (NFR04) (`player/src/theme/initTheme.ts`, `player/src/main.tsx`)
- [ ] 12. Adjust the M3 components (`TonalCard`, `ElevatedCard`, `FabButton`) and `DashboardLayout` to use theme tokens that respond to `dark:` (FR05, NFR01) (`player/src/components/m3/*.tsx`, `player/src/layout/DashboardLayout.tsx`)
- [ ] 13. Display the user's real name/initials in the header via `AppContext`, replacing "Welcome, User"/"U" (FR03) (`player/src/layout/DashboardLayout.tsx`)
- [ ] 14. Turn the inert avatar (C7) into a user menu — profile/settings/sign out, reusing `encerrarSessao` (C8) (FR14) (`player/src/layout/DashboardLayout.tsx`)
- [ ] 15. Rewrite `Projects` to render the system grid via `useSistemas` + `StateViews`; "Open project" (C4) navigates and the "Create new project" card (C3) starts creation (FR02, FR04, FR06) (`player/src/pages/Dashboard/Projects.tsx`)
- [ ] 16. Implement `useMetricas` deriving counters from `listarSistemas()` (FR01) (`player/src/dashboard/useMetricas.ts`)
- [ ] 17. Rewrite `Overview` with real metrics via `useMetricas`; "Get Started" (C1) and the "Create" FAB (C2) trigger system creation; loading/empty/error states; remove `alert()` (FR01, FR04, FR06) (`player/src/pages/Dashboard/Overview.tsx`)
- [ ] 18. Wire `Settings`: "Toggle Theme" (C6) to `ThemeContext` and "Edit Profile" (C5) to a route/placeholder (FR04, FR05) (`player/src/pages/Dashboard/Settings.tsx`)
- [ ] 19. Audit the Control Inventory (spec §2.1): confirm C1–C7 have a real handler (no `alert()`/clickable `<div>` without a handler) and C8–C10 remain functional (FR04) (`player/src/pages/Dashboard/*.tsx`, `player/src/layout/DashboardLayout.tsx`)
- [ ] 20. Update the dashboard screen tests for behavior — each control C1–C7 triggers its action (spy/mock) and the data states are covered — instead of static text (FR01, FR02, FR04, FR06, FR14) (`player/src/pages/Dashboard/Overview.test.tsx`, `Projects.test.tsx`, `Settings.test.tsx`, `player/src/layout/DashboardLayout.test.tsx`)

## Phase 2 — Advanced wireframe features (depend on contract/backend)

- [ ] 21. Extend the `Sistema` consumption to display status (Published/Draft/Failed) and active version in the cards, degrading gracefully when absent (FR07, BR04) (`player/src/pages/Dashboard/Projects.tsx`, `player/src/api/types.ts`)
- [ ] 22. Add status filters (All/Published/Drafts) and a grid/list toggle in `Projects` (FR12) (`player/src/pages/Dashboard/Projects.tsx`)
- [ ] 23. Implement per-system collaborator presence (stacked avatars) via `phoenixSocket` (FR08) (`player/src/dashboard/PresencaColaboradores.tsx`, `player/src/collab/phoenixSocket.ts`)
- [ ] 24. Display a failure alert/DLQ count per system (FR09, BR09) (`player/src/pages/Dashboard/Projects.tsx`)
- [ ] 25. Implement a Cmd/Ctrl+K command palette for system search and actions (FR10) (`player/src/dashboard/CommandPalette.tsx`, `player/src/layout/DashboardLayout.tsx`)
- [ ] 26. Implement the hierarchical tenant selector and notification indicator in the top bar (FR11, FR13, BR01) (`player/src/dashboard/TenantSwitcher.tsx`, `player/src/layout/DashboardLayout.tsx`)

## Wrap-up

- [ ] 27. Run the full suite and the build: `npm run test` and `npm run build` (`tsc --noEmit` + `vite build`) must pass without errors (`player/`)
