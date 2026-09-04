# Applied Principles — Global-Scope Audit

## 1. Obvious Start

Every dashboard screen already follows the same pattern: a header `TonalCard` with title + subtitle,
followed by the main action/data (Clientes → list; Configurações → section cards; Perfil → account
data). Consistent across all 7 screens reviewed — no change needed here.

## 2. Clear Reversal

- Destructive actions (delete client, delete account) already require password confirmation
  (`SegurancaForm.tsx`) or are isolated with a `title` explaining the irreversibility
  (`ClienteSistemas.tsx:142`). Correct pattern, kept as is.
- **Real gap**: neither one uses a confirmation dialog (`components/ui/dialog.tsx` already
  exists in the project, but isn't used in these flows) — today "Delete client"/"Delete account" act
  on the first click. Password re-authentication (account deletion) mitigates accidental clicks, but
  "Delete client" in `ClienteSistemas.tsx:140` requires neither a password nor confirmation — just a
  tooltip `title`, which goes unnoticed. Recommendation: reuse `dialog.tsx` here.

## 3. Consistent Logic

Main finding of this audit (detailed in `01-contexto.md`):

- **Success color**: `text-emerald-600 dark:text-emerald-400` repeated literally in 5 files
  (`Perfil.tsx:94`, `ClienteSistemas.tsx:121`, `abas/AbaVersao.tsx:84`, `SegurancaForm.tsx:127`,
  `WhiteLabelForm.tsx:89`). It's consistent *with itself* (always the same value) — the problem
  isn't divergence, it's that it isn't in the token system (`index.css`), so it can't be adjusted in
  a single place or managed per theme the way `--destructive` already is. Resolved in doc `04`.
- **Alert/pending color**: `bg-amber-500/15 text-amber-600 dark:text-amber-400` in
  `CardFeedback.tsx:28-29` — same case, the only use of amber in the project, also without a token.
- **Input size**: two patterns coexist for the same semantic component (plain text field) —
  `px-4 py-3 rounded-xl` (Clientes/ClienteSistemas, a screen's "primary" forms) vs
  `px-3 py-2 rounded-lg` (Perfil/SegurancaForm, account forms). There's no shared `<Input>`
  component being reused (`components/ui/input.tsx` exists, but none of these 4 screens imports
  it — all of them hand-write the `className` of the `<input>`). Recommendation: migrate the 4 forms
  to `components/ui/input.tsx`, unifying the size as a conscious decision (not a copy-paste
  accident).

## 4. Follow Conventions

- **Language**: `Dashboard.tsx` (the first screen any user sees after login) is in English
  ("Build your Next Flow", "Get Started", "Create") while the rest of the dashboard is PT-BR.
  This breaks a product's most basic convention: a single language promised to the user. It's the
  highest-impact fix in this audit — visible to 100% of users, on the first screen.
- Icons (`lucide-react`), typography (Outfit/Inter/JetBrains Mono), and `m3/`/`ui/` components are
  reused consistently across all 12 screens reviewed — no other convention deviation found beyond
  the ones already listed.

## 5. Feedback and Milestones

- The inline success pattern (`role="status"`, green text next to the button) is used consistently
  in Perfil, ClienteSistemas, SegurancaForm, WhiteLabelForm, AbaVersao — a good pattern, just
  missing the token (item 3).
- The error pattern (`role="alert"`, `text-destructive`) is also consistent across every screen with
  a form — it already uses the correct token, unlike success. No change needed.
- `Skeleton`/`EmptyState`/`ErrorState` from `StateViews.tsx` are reused by Clientes,
  ClienteSistemas, and SeletorSistemas — the only place that didn't reuse them yet was
  `Monitor.tsx` (already addressed in the `008-monitor-recursos` audit).

## 6. Proximity and Adaptation

Responsive grids (`grid-cols-1 md:grid-cols-2`/`lg:grid-cols-3`) and `max-w-5xl mx-auto`/`max-w-2xl
mx-auto` are used consistently to limit reading width on wide screens — no new findings here.

## 7. Interface Is Content

Login/Register dedicate half the screen (`md:w-1/2`) to a purely branding panel — is that wasted
space on screens narrower than `md`? No: on mobile the brand panel disappears
(`flex-col md:flex-row`; the purely decorative panel doesn't have `md:hidden` but stacks above the
form instead) — acceptable behavior for a gateway screen (brand reinforcement before the action is a
valid choice here, unlike the internal dashboard where it would be wasteful).

## 8. General Visual Design Principles

- **Integrated form and content**: once `--success`/`--warning` exist as tokens (see
  doc `04`), every success message in the app automatically inherits any future theme adjustment —
  today, a contrast tweak would require manually editing 5 files.
- **Appropriate data visualization**: the dashboard's only table (`ClienteSistemas.tsx`) uses plain
  text for "Name" — appropriate, since this isn't data that calls for a different representation.

## 9. Design Decision Matrix

| Decision | Obvious Start | Clear Reversal | Consistency | Convention | Feedback | Proximity | Content > Deco. |
|---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| Promote `--success`/`--warning` to tokens in `index.css` | — | — | ✓ | ✓ | ✓ | — | ✓ |
| Translate `Dashboard.tsx` to PT-BR | — | — | ✓ | ✓ | — | — | — |
| Unify input size via `components/ui/input.tsx` | — | — | ✓ | ✓ | — | — | — |
| Confirmation (dialog) before "Delete client" | — | ✓ | ✓ | ✓ | ✓ | — | — |
| Keep Login/Register's own visual identity (zinc split-screen) | ✓ | — | ✓ (internal) | ✓ | — | — | ✓ |
