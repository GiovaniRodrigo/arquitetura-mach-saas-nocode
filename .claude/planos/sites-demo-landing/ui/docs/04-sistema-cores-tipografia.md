# Color and Typography System

> The current palette and body font (defined in `build/seed-demo-site.sh:39-49` and inherited from
> the app's `tailwind.config.js`) already match what market research points to as correct for
> this domain — see `02-referencias.md`. This document **does not propose changing** the hue or
> the body font; it formalizes the derived scale (today only partially present, some ad-hoc
> values) and proposes a single platform gain (a display font on headings).

## Color Palette

### Primary (validated — Tailwind `indigo-600`, reference: Figma "Indigo" color page)
- Main: `#4f46e5` (already in use, `COR_PRIMARIA`)
- Dark (hover/emphasis states, stats strip): `#4338ca` (already in use, `COR_PRIMARIA_ESCURA`)
- Light (badge/icon backgrounds): `#eef2ff` (already in use, `COR_PRIMARIA_CLARA`)

Full scale recommended for consistency (only the 50/600/700 tones are used today — the
rest exist in Tailwind indigo and can be used if the palette needs more variation, e.g.
form focus states):

| Tone | Hex | Suggested use |
|---|---|---|
| 50 | `#eef2ff` | badge background, icon (already in use as `COR_PRIMARIA_CLARA`) |
| 100 | `#e0e7ff` | light background hover (not used today) |
| 500 | `#6366f1` | default color in `componentRegistry.ts` (button/badge with no custom style) |
| 600 | `#4f46e5` | primary (already in use) |
| 700 | `#4338ca` | dark/emphasis (already in use as `COR_PRIMARIA_ESCURA`) |

### Neutrals (already in use, no change)
- Text: `#111827` (`COR_TEXTO`)
- Muted text: `#6b7280` (`COR_TEXTO_MUTED`)
- Border: `#e5e7eb` (`COR_BORDA`)
- Light background (alternating sections): `#f9fafb` (`COR_FUNDO_CLARO`)
- Dark (footer/stats strip): `#111827` (`COR_ESCURO`)
- Muted on dark: `#9ca3af` (`COR_MUTED_NO_ESCURO`)

### Semantic (absent from the Home today — only relevant if the demo gains form/alert states)
These tokens already exist in the component catalog via `alerta` (`PreviewRenderer.tsx:171-180`,
`ESQUEMAS_ALERTA`) — reuse them, don't reinvent:
- Success: `#15803d` text / `#f0fdf4` background
- Warning: `#b45309` text / `#fffbeb` background
- Error: `#b91c1c` text / `#fef2f2` background
- Info: `#1d4ed8` text / `#eff6ff` background

## Typography

### Families
- **Body (validated, no change)**: Inter — loaded in `index.css:1` and `tailwind.config.js:18`
  as the default font. Reference: Inter appears in 182 of the SaaS sites analyzed by the
  saaslandingpage.com ranking (see `02-referencias.md`) — it's the market's #1 choice, the
  project already got this right.
- **Display/headings (new recommendation)**: Outfit — **already loaded** in `index.css:1`
  (`Outfit:wght@400;500;600;700;800`) but **never used** in any component, because `Estilos`
  (`componentRegistry.ts:45-68`) has no font-family field — all text inherits the app's font
  (Inter). Reference: Outfit described as a "geometric sans-serif with a friendlier personality
  than Inter, rounded terminals" (saaslandingpage.com) — the Inter (body) + Outfit (headings)
  combo is exactly the "two typographic voices" pattern from the 2025 trend references
  (`01-contexto.md`), at **zero network cost** (the font is already downloaded by the app, it
  just needs to be wired up). Requires: adding `fonteFamilia?: 'inter' | 'outfit'` to `Estilos`
  and an `if (estilos.fonteFamilia) css.fontFamily = ...` in `estilosCss.ts` — a small, localized
  change, out of scope for this content audit (recorded here for a future iteration, if the user
  agrees).

### Scale (consolidating the sizes already used on the Home, without introducing new ones)
| Role | Size | Weight | Where it already appears |
|---|---|---|---|
| Hero heading | 44px | bold (700) | `hero-heading` |
| Section H2 | 28-30px | bold (700) | `features-heading`, `planos-heading`, `depoimentos-heading`, `faq-heading` |
| Card H3 | 16-20px | bold (700) | `feat-N-titulo`, `plano-N-nome` |
| Body | 14-17px | normal (400) | general paragraphs |
| Caption/label | 12-13px | medium (500) | badges, stat labels, footer copy |

No change recommended to the scale — it's already consistent and follows a reasonable progression
(44 → 28-30 → 16-20 → 14-17 → 12-13). The gain is in varying the **family**, not the **size**.

## Spacing and Radius (consistency adjustment, see `03-principios-aplicados.md` item 3)

Recommended `border-radius` scale (today it varies freely between 8-16px per section with no
criteria):

| Role | Radius | Where to apply |
|---|---|---|
| Small (badge, button, input) | `8px` | already the predominant pattern — keep |
| Medium (card) | `16px` | unify `feature_card`/`testimonial_card` (currently 12px) and `pricing_card` (already 16px) at this value |
| Large (hero image, showcase) | `20px` | raise from 16px — reinforces hierarchy (images "float" more than cards) |

## Shadow (new — layers instead of a single value)

Recommended pattern for any "elevated" element (hero image, highlighted pricing card,
testimonial cards), replacing the current single shadow with two layers (ambient +
contact) — a technique described in the 2025 trend references as common in products like
Linear/Stripe, and fully supported today by the `sombra` field (free-form string → `box-shadow`):

```
sombra: "0 1px 2px rgba(17,24,39,.06), 0 24px 48px -12px rgba(17,24,39,.18)"
```

For elements with the brand color (primary CTA, highlighted pricing), the contact layer uses the
primary color instead of pure neutral:

```
sombra: "0 1px 2px rgba(79,70,229,.15), 0 20px 40px -8px rgba(79,70,229,.35)"
```
