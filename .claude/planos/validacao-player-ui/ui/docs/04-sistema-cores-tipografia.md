# Color and Typography System (recommended)

Anchored in popular references (Material Design 3; no-code builder patterns like
Appsmith/Retool; Google/GitHub IDP guidelines). A single system for **every** screen of the
player, eliminating the current inconsistency (only Login has styling).

## Color Palette

### Primary (indigo — common in productivity tools/builders)
- 50  `#EEF2FF`
- 100 `#E0E7FF`
- 200 `#C7D2FE`
- 300 `#A5B4FC`
- 400 `#818CF8`
- **500 `#6366F1`** ← main (CTA)
- 600 `#4F46E5` ← CTA hover
- 700 `#4338CA`
- 800 `#3730A3`
- 900 `#312E81`

> Reference: an indigo/violet tone is recurrent in internal-tools builders (Retool/Appsmith) and
> in Material 3 (the "Indigo" scheme). High legibility for white-on-color buttons.

### Neutrals (surfaces and text)
- App background `#F8FAFC` · Surface/card `#FFFFFF` · Border `#E2E8F0`
- Strong text `#0F172A` · Secondary text `#475569` · Subtle text `#94A3B8`

### Semantic (Material 3 / universal convention)
- Success `#16A34A` (background `#DCFCE7`)
- Warning `#D97706` (background `#FEF3C7`)
- Error   `#DC2626` (background `#FEE2E2`)
- Info    `#2563EB` (background `#DBEAFE`)

### IDP branding (guideline must be followed)
- Google: **white** button with `#DADCE0` border, 4-color "G" logo, `#3C4043` text.
- GitHub: **black** button `#24292F` (or white with a border), white text, official GitHub logo.

## Typography

### Families (popular trend, free on Google Fonts)
- **Headings**: `Inter` — one of the most widely adopted UI fonts in SaaS/dashboards
  (tens of millions of downloads/month on Google Fonts).
- **Body**: `Inter` (same family, weights 400/500/600) — consistency and great UI readability.
- **Code/Mono**: `JetBrains Mono` (to display technical ids/tokens, e.g. `sistemaId`).

> No-CDN fallback: `system-ui, -apple-system, "Segoe UI", Roboto, sans-serif` (the player
> already uses `system-ui` today; keep it as a fallback and optionally load Inter).

### Scale (16px base, ~1.25 ratio)
- Display / h1: `2rem` / 32px — 700
- h2: `1.5rem` / 24px — 600
- h3: `1.25rem` / 20px — 600
- Body: `1rem` / 16px — 400
- Body-sm: `0.875rem` / 14px — 400
- Caption: `0.75rem` / 12px — 500 (labels, help text)

## Spacing (Grid)
- **Base unit: 8px** (scale 4/8/12/16/24/32/48/64).
- Columns: 12 (desktop) · 4 (mobile). Content container `max-width: 1120px`.
- Login card: `max-width: 400px`, padding 32px, `radius: 12px`.

## Tokens (for reuse — CSS custom properties)
```css
:root{
  --color-primary:#6366F1; --color-primary-600:#4F46E5;
  --color-bg:#F8FAFC; --color-surface:#FFFFFF; --color-border:#E2E8F0;
  --color-text:#0F172A; --color-text-2:#475569; --color-text-3:#94A3B8;
  --color-success:#16A34A; --color-warning:#D97706; --color-error:#DC2626; --color-info:#2563EB;
  --space:8px; --radius:12px; --radius-sm:8px;
  --font-ui:"Inter",system-ui,-apple-system,"Segoe UI",Roboto,sans-serif;
  --tap-min:44px;
}
```

## Sources
- Material Design 3 — https://m3.material.io/
- Sign in with Google best practices — https://developers.google.com/identity/siwg/best-practices
- SaaS login page design 2025 — https://lollypop.design/blog/2025/october/saas-login-page-design/
