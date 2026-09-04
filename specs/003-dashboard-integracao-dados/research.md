# Research: Dashboard — Data Integration and Functionality

---

## 1. Existing Patterns in the Project

The project already contains almost everything Phase 1 needs — the dashboard just doesn't reuse it.
The research confirms that most of the work is **reuse and wiring**, not building from scratch.

| File/Pattern | Location | Relevance |
|----------------|-------------|-----------|
| `SeletorSistemas.tsx` | `player/src/systems/` | Canonical reference: `listarSistemas()`/`criarSistema()` with a skeleton (`animate-pulse`), empty state, error + "Try again", and `abrirSistema()` via query string. Source to extract into `useSistemas`. |
| `ApiClient` | `player/src/api/client.ts` | Already exposes `listarSistemas`, `criarSistema`, `versaoAtiva`, `permissoes`, `criarExportacao`; injectable `fetch` for tests. No new network call is needed in Phase 1. |
| `types.ts` | `player/src/api/types.ts` | `Sistema` only has `id`/`nome` → confirms FR07/FR09 require a contract extension (Phase 2). |
| `session.ts` | `player/src/auth/session.ts` | JWT persisted in `localStorage` (`mach_token`); basis for reading identity claims (FR03). |
| M3 components | `player/src/components/m3/` | `TonalCard`, `ElevatedCard`, `FabButton`, `NavPill` — reuse to preserve NFR01. Note: they use fixed colors (`bg-white`, `bg-slate-100`, `bg-blue-200`) that **do not respond to `dark:`** — will need theme tokens for dark mode (FR05). |
| `sidebar.tsx` / `SidebarProvider` | `player/src/components/ui/` | Responsiveness and sidebar toggle already solved (uses `use-mobile`); preserve (NFR02). |
| `phoenixSocket.ts` | `player/src/collab/` | Existing Phoenix channel; basis for real-time presence (FR08, Phase 2), not currently used by the dashboard. |
| Test pattern | `*.test.tsx` | Vitest + Testing Library + mocked `fetch`/`matchMedia`; follow the same style in the new tests. |

---

## 2. Technologies and Libraries

| Technology | Version | Use | Already installed? |
|------------|--------|-----|---------------|
| React + react-router-dom | 6+ | Dashboard SPA and routes | Yes |
| Tailwind CSS | 4 (`@tailwindcss/postcss`) | M3 styles and `dark:` for theme | Yes |
| lucide-react | — | Icons (Home, Folder, Settings, LogOut) | Yes |
| Vitest + Testing Library | — | Unit/behavior tests | Yes |
| Playwright | — | E2E (configured; optional for dashboard flows) | Yes |
| phoenix (`@types/phoenix`) | — | Collaboration WebSocket (Phase 2) | Yes |

No new dependency is needed for Phase 1. Dark mode uses Tailwind's
class strategy (`dark:`), with no additional library.

---

## 3. External References

| Reference | URL | What it resolves |
|------------|-----|--------------|
| Dashboard wireframe | `.claude/planos/001-construtor-sistemas-mach-v4/ui/wireframes/dashboard.html` | Rich target vision (tenant, Cmd+K, status/version, presence, DLQ, filters, empty/skeleton) — basis for FR07–FR13 |
| M3 rules | `.agents/Demandas/dashboard-m3-regras-negocio.md` | Original FR/NFR for the M3 aesthetic (hero, metrics, FAB, responsive) |
| UI docs | `.claude/planos/001-.../ui/docs/04-sistema-cores-tipografia.md` | Color/typography system for theme tokens |
| Tailwind dark mode | https://tailwindcss.com/docs/dark-mode | `dark` class strategy on `<html>` (NFR04) |

---

## 4. Alternatives Considered

### Option A: Rewrite `Projects` from scratch with its own API call
- **Pros**: Fast in the very short term.
- **Cons**: Duplicates `SeletorSistemas` (violates NFR05/DRY); two maintenance
  points for loading/error states.
- **Decision**: Rejected.

### Option B: Extract `useSistemas` and reuse it in `Projects` and `SeletorSistemas`
- **Pros**: Single source of truth; independently testable; removes duplication.
- **Cons**: Requires refactoring `SeletorSistemas` (with test coverage beforehand).
- **Decision**: **Chosen.**

### Option C (theme): a theme library (e.g., next-themes-like)
- **Pros**: Ready-made.
- **Cons**: Extra dependency; a simple Tailwind SPA doesn't justify it.
- **Decision**: Rejected — a custom `ThemeContext` + synchronous `initTheme()`.

### Option D (identity): fetch the user profile from a dedicated endpoint
- **Pros**: Always up-to-date data.
- **Cons**: The endpoint may not exist; an extra round-trip just to display name/avatar.
- **Decision**: Read JWT claims for display (FR03/NFR06); a profile endpoint remains
  a future evolution.

### Option E (metrics): aggregated metrics endpoint
- **Pros**: Precise numbers (pending tasks, members).
- **Cons**: Doesn't exist today.
- **Decision**: Derive from `listarSistemas()` in Phase 1; specify the contract for the future.
