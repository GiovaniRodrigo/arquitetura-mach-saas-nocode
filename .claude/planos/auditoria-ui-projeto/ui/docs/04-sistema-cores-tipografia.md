# Color and Typography System — Extension Proposal

> Reaffirms the system already documented in `008-monitor-recursos/ui/docs/04-sistema-cores-tipografia.md`
> (HSL tokens from `index.css`, Inter/Outfit/JetBrains Mono). This document formalizes the two tokens
> that the whole-project scan proved necessary — `--success` and `--warning` — with concrete
> evidence of repeated use, and proposes migrating the files that currently use loose color values.

## New tokens proposed in `services/frontend/src/index.css`

```css
:root {
  /* ...existing tokens... */
  --success: 152 69% 31%;          /* same hue as the emerald-600 already used across the app */
  --success-foreground: 0 0% 100%;
  --warning: 38 92% 50%;           /* amber-500, already used in CardFeedback.tsx */
  --warning-foreground: 24 10% 10%;
}

.dark {
  /* ...existing tokens... */
  --success: 152 55% 42%;          /* approximates the emerald-400 used in current dark mode */
  --success-foreground: 0 0% 100%;
  --warning: 38 92% 55%;           /* approximates amber-400 */
  --warning-foreground: 24 10% 10%;
}
```

Values calibrated to match what's already in production (Tailwind's `emerald-600`/`emerald-400` and
`amber-500`/`amber-400`) — the migration is **purely structural** (tokenizing what already exists
visually), not a repaint.

## Recommended migration (file → replacement)

| File | Before | After |
|---|---|---|
| `pages/Dashboard/Perfil.tsx:94` | `text-emerald-600 dark:text-emerald-400` | `text-success` |
| `pages/Dashboard/ClienteSistemas.tsx:121` | `text-emerald-600 dark:text-emerald-400` | `text-success` |
| `pages/Dashboard/abas/AbaVersao.tsx:84` | `text-emerald-600 dark:text-emerald-400` | `text-success` |
| `configuracao/SegurancaForm.tsx:127` | `text-emerald-600 dark:text-emerald-400` | `text-success` |
| `configuracao/WhiteLabelForm.tsx:89` | `text-emerald-600 dark:text-emerald-400` | `text-success` |
| `dashboard/CardServicoStatus.tsx:44` | `bg-green-500` / `bg-red-500` | `bg-success` / `bg-destructive` (already recommended in `008-monitor-recursos`) |
| `dashboard/CardFeedback.tsx:28-29` | `bg-amber-500/15 text-amber-600 dark:text-amber-400` | `bg-warning/15 text-warning` |

Requires adding `success` and `warning` (with `-foreground`) to the color mapping in
`tailwind.config.js` (same pattern already configured for `destructive`/`accent`), so that
`bg-success`/`text-success`/`bg-warning`/`text-warning` work as Tailwind utilities.

## Typography and spacing

No changes — `Inter`/`Outfit`/`JetBrains Mono` and the `text-2xl`/`text-md`/`text-sm` scale are
already used consistently across the 12 screens reviewed in this audit. The only point of attention
(non-typographic): standardize the input component per `03-principios-aplicados.md` §3.
