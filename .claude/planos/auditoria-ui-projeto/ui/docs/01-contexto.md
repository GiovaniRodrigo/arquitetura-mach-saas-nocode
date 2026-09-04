# Project Context — Global-Scope UI Audit

> This audit widens the `/ui` scope from a single isolated screen (last run: `008-monitor-recursos`)
> to **the entire frontend** of the project. It does not repeat what has already been documented — it
> inventories existing coverage and focuses on what had not yet gone through a UI review, in addition
> to recording **real inconsistencies found while reading the code end to end** (only visible in a
> whole-project scan, not screen by screen).

## Domain

MAYS — Make Your SaaS (MACH V4 architecture): a no-code SaaS platform that lets a user
("owner"/"partner") create systems (screens, business rules, data) without coding, manage them per
client/tenant, and operate the platform itself (infrastructure monitoring). The Frontend covers
three distinct journeys:

1. **Public gateway** (unauthenticated): Login, Sign-up (Register).
2. **Operational dashboard** (authenticated): Dashboard/Home, Clients → Client Systems →
   tabs (Screens/Business Rules/Version), Settings, Account/Profile, Help, Resource Monitor.
3. **Visual builder** (screen editor inside the "Screens" tab): Canvas, Inspector, Component/Layer
   Panel — a Figma/Webflow-style editor.

## Target Audience

Reaffirms the survey from `001-construtor-sistemas-mach-v4/ui/docs/01-contexto.md`: owners/partners
(semi-technical, create and publish systems) and, for the Monitor screen, a more technical profile
(operating the platform itself). Login/Register serve people who don't yet have an account — the
gateway page can (and should) have its own visual identity, more "marketing"-oriented, distinct from
the internal dashboard.

## Screen Inventory and UI Coverage

| Screen / Component | File | UI Pattern | Already audited? |
|---|---|---|---|
| Login | `auth/Login.tsx` | Split-screen (zinc-900 + form) | ✓ `validacao-player-ui/ui/wireframes/login.html` |
| Sign-up | `auth/Register.tsx` | Same split-screen as Login (4 extra fields) | Mirrors Login, not repeated here |
| Dashboard (Home) | `pages/Dashboard/Dashboard.tsx` | Hero + metrics + FAB | ✓ `001-construtor-sistemas-mach-v4/ui/wireframes/dashboard.html` |
| Clients (list) | `pages/Dashboard/Clientes.tsx` | Card grid + creation form | Partial — card grid already covered by other screens; form is not |
| Client Systems | `pages/Dashboard/ClienteSistemas.tsx` | Edit form + **data table** | **New in this audit** — the only use of `<table>` in the dashboard |
| Settings | `pages/Dashboard/Configuracao.tsx` + `SegurancaForm`/`WhiteLabelForm` | Stacked cards with forms (password, MFA, account deletion, white label) | **New in this audit** |
| Account/Profile | `pages/Dashboard/Perfil.tsx` | Account form (name, photo, email) | **New in this audit** (same form pattern as Settings) |
| Help | `pages/Dashboard/Ajuda.tsx` | Search + list grouped by category | Low complexity, pattern already covered by other card grids |
| Resource Monitor | `pages/Dashboard/Monitor.tsx` | Status cards | ✓ `008-monitor-recursos/ui/` (previous audit) |
| System Selector | `systems/SeletorSistemas.tsx` | Card grid + creation | Same pattern as Clientes.tsx |
| Builder (Screens/Rules/Version tabs) | `pages/Dashboard/abas/*`, `pages/Dashboard/editor/*` | Complex visual editor | ✓ `001-construtor-sistemas-mach-v4/ui/wireframes/builder.html` |
| Player/published screen | (`player` service) | Headless rendering | ✓ `validacao-player-ui/ui/wireframes/tela-dinamica.html` + `estados.html` |

## Consistency Findings (only visible in a whole-project scan)

A screen-by-screen read does not catch this — it only surfaced when comparing the code of 12+ files
side by side:

1. **Status color outside the token system, in 7 different files.** `text-emerald-600
   dark:text-emerald-400` appears hardcoded in `Perfil.tsx:94`, `ClienteSistemas.tsx:121`,
   `abas/AbaVersao.tsx:84`, `SegurancaForm.tsx:127`, `WhiteLabelForm.tsx:89`; `bg-amber-500/15
   text-amber-600` in `CardFeedback.tsx:28-29`; `bg-green-500`/`bg-red-500` in
   `CardServicoStatus.tsx:44` (already flagged in the `008-monitor-recursos` audit). In other words:
   **every success message in the app uses the same emerald color, always with the same light/dark
   pair — it was just never promoted to a token** `--success` in `index.css`. This isn't a single
   screen's preference, it's a real, consistent project pattern that still needs to be formalized
   (see `04-sistema-cores-tipografia.md`).
2. **`Dashboard.tsx` is the only screen in the authenticated dashboard with English text.** "Build
   your Next Flow", "Start creating projects and designing your business architecture with our
   intuitive node-based editor", "Get Started", "Create" — while Clientes, Configurações, Perfil,
   Ajuda, Monitor, ClienteSistemas are 100% PT-BR. See `03-principios-aplicados.md` (Follow
   Conventions).
3. **Two input sizes for the same text-field type**, with no apparent functional reason:
   "highlighted" forms (create client in `Clientes.tsx`, edit name in `ClienteSistemas.tsx`)
   use `px-4 py-3 rounded-xl`; account forms (`Perfil.tsx`, `SegurancaForm.tsx`) use
   `px-3 py-2 rounded-lg`. Same semantic component (plain text field), two dimensions.
4. **Login and Register share the same "gateway" visual identity** (split-screen zinc-900 +
   form), consistent with each other — this is not an inconsistency, it's a deliberate and correct
   design decision (public journey ≠ authenticated journey). Recorded here only so it isn't confused
   with findings 1–3 in a future read.

## Visual References Found

| Reference | Popularity | Why it's relevant |
|---|---|---|
| SaaSUI (Notion/Linear/Figma/Stripe screens) | Reference library dedicated to real SaaS patterns (login, settings, dashboards) | Confirms that "settings page with cards stacked by section" (already used in `Configuracao.tsx`) is the prevailing pattern in reference B2B SaaS — no need to reinvent it, just refine consistency. |
| shadcn/ui — data table (`tasks` example) | Base of one of the most widely adopted design systems in the current React/Tailwind ecosystem (the project's own stack is shadcn-like) | The closest table reference to the stack already in use (`components/ui/*` already follows shadcn conventions) — directly applicable to the systems table in `ClienteSistemas.tsx`. |
| Stripe Dashboard (billing/settings) | Recurring market reference in B2B SaaS comparisons | Confirms the pattern of an inline success message next to the action button (used in Perfil/SegurancaForm) instead of a toast — appropriate for low-frequency settings forms. |
