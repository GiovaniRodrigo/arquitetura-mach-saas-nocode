# Color and Typography System

Grounded in the popular references (Vercel Geist, Linear, Retool) and the 2025–26 dark-first trends. Dual strategy: **dark-first** for Dashboard/Builder (technical audience, long sessions) and **light-first** for the Headless Player (layperson End Customer). The tokens below feed directly into the wireframes.

## Color Palette

### Neutrals — base scale (reference: Vercel Geist, deliberate gray scale)

Canonical **dark** surface (Dashboard/Builder):

| Token | Hex | Use |
|---|---|---|
| `--bg-base` | `#0A0A0A` | Application root background |
| `--bg-surface` | `#141414` | Cards, panels, sidebars |
| `--bg-surface-2` | `#1C1C1C` | Elevated surface (menus, popovers) |
| `--bg-canvas` | `#101012` | Builder canvas |
| `--border` | `#262626` | Borders and dividers |
| `--border-strong` | `#333333` | Focus/hover borders |
| `--text-primary` | `#FAFAFA` | Title and main body text |
| `--text-secondary` | `#A1A1A1` | Secondary text/labels |
| `--text-muted` | `#6B6B6B` | Placeholders, disabled |

**Light** surface (Headless Player):

| Token | Hex | Use |
|---|---|---|
| `--bg-base` | `#FFFFFF` | Background |
| `--bg-surface` | `#FAFAFA` | Form card |
| `--border` | `#E5E5E5` | Field borders |
| `--text-primary` | `#171717` | Ink (Geist) |
| `--text-secondary` | `#525252` | Labels/help text |

### Primary / Accent (reference: Linear restraint — a single accent)

Violet-indigo, distinct from generic blue, reserved for the primary CTA and active state:

| Step | Hex | Use |
|---|---|---|
| 50  | `#EEF0FF` | Soft highlight background (light) |
| 100 | `#DCE0FF` | Accent chips/badges |
| 200 | `#B8C0FF` | Soft accent borders |
| 300 | `#8E9BFF` | Hover on dark surfaces |
| 400 | `#6E7BFF` | Active icons |
| **500** | **`#5B63F5`** | **Primary color — CTA, focus, selection** |
| 600 | `#4A50DB` | CTA hover |
| 700 | `#3B40B0` | Pressed |
| 800 | `#2E3288` | — |
| 900 | `#242766` | Text over light accent background |

### Semantic Colors (form reinforces meaning — principle 8)

| Role | Hex (dark) | Hex (light) | Use |
|---|---|---|---|
| **Success** (published/saved) | `#3FB950` | `#16A34A` | Active version, "saved", successful submission |
| **Warning** (draft/pending) | `#D29922` | `#CA8A04` | Unpublished draft, job collecting |
| **Error** (failure/destructive) | `#F85149` | `#DC2626` | Validation error, destructive action |
| **Info** (neutral) | `#58A6FF` | `#2563EB` | Hints, links, traces |

### Collaboration colors (reference: Figma multiplayer — per-user colors)

Cursor/avatar palette, assigned cyclically per present user (FR06):

`#F97316` · `#EC4899` · `#8B5CF6` · `#06B6D4` · `#22C55E` · `#EAB308` · `#EF4444` · `#3B82F6`

> Each color travels with the named cursor, the border of the component being edited, and the presence avatar — the same color links the three representations of the user (consistency, principle 3).

## Typography

### Families (based on popular dev-tools trends)

- **Titles/Display**: **Geist** (Vercel) or fallback **Inter Display** — the standard for the most-copied dev-tools design systems of 2024–25.
- **Body/UI**: **Inter** — the most widely used dense interface font in dev-tools (Retool, Linear); ~millions of downloads/month on Google Fonts, historical top-3.
- **Mono/Code**: **JetBrains Mono** or **Geist Mono** — for `blind_index`, `trace_id`, technical values, and version badges.

```css
--font-display: "Geist", "Inter", system-ui, sans-serif;
--font-body: "Inter", system-ui, -apple-system, sans-serif;
--font-mono: "JetBrains Mono", "Geist Mono", ui-monospace, monospace;
```

### Type Scale (base 16px · ratio ~1.25)

| Level | rem | px | Weight | Use |
|---|---|---|---|---|
| Display | 2.25rem | 36px | 600 | Empty-state screen title / Player hero |
| h1 | 1.75rem | 28px | 600 | Page title (Dashboard) |
| h2 | 1.375rem | 22px | 600 | Section title / system name |
| h3 | 1.125rem | 18px | 500 | Subtitles, card header |
| body-lg | 1rem | 16px | 400 | Player body (layperson — larger) |
| body | 0.875rem | 14px | 400 | Standard body/UI (Builder/Dashboard) |
| caption | 0.75rem | 12px | 500 | Labels, metadata, help text |
| mono | 0.8125rem | 13px | 500 | `blind_index`, versions, traces |

Line height: 1.5 for body, 1.2 for titles. Top-aligned labels in the Player (NN/g: faster completion).

## Spacing (Grid)

- **Base unit: 4px** (scale: 4 · 8 · 12 · 16 · 24 · 32 · 48 · 64).
- **Columns**: 12 (desktop) · 8 (tablet) · 4 (mobile).
- **Border radius**: `--radius-sm: 6px` (fields/buttons), `--radius-md: 10px` (cards/panels), `--radius-lg: 14px` (modals), `--radius-full: 9999px` (avatars/chips).
- **UI row heights**: header 48px · table row 44px · minimum touch target 44px (Player).
- **Elevation** (dark): subtle shadows + 1px border — the border does the separation work, not the shadow (Geist aesthetic).

## Reference CSS Tokens (used in the wireframes)

```css
:root {
  /* dark-first — Dashboard/Builder */
  --bg-base: #0A0A0A;
  --bg-surface: #141414;
  --bg-surface-2: #1C1C1C;
  --bg-canvas: #101012;
  --border: #262626;
  --border-strong: #333333;
  --text-primary: #FAFAFA;
  --text-secondary: #A1A1A1;
  --text-muted: #6B6B6B;

  --accent: #5B63F5;
  --accent-hover: #4A50DB;
  --accent-soft: rgba(91,99,245,0.12);

  --success: #3FB950;
  --warning: #D29922;
  --danger:  #F85149;
  --info:    #58A6FF;

  --space-unit: 4px;
  --radius-sm: 6px;
  --radius-md: 10px;
  --radius-lg: 14px;

  --font-body: "Inter", system-ui, sans-serif;
  --font-mono: "JetBrains Mono", ui-monospace, monospace;
}
```

---

**Sources**:
- [Vercel Geist — Colors](https://vercel.com/geist/colors)
- [Geist Design System Breakdown (DesignSystems.one)](https://www.designsystems.one/design-systems/vercel-geist)
- [19 Best Dark Mode Dashboard Templates 2026 (AdminLTE)](https://adminlte.io/blog/dark-dashboard-templates/)
- [Multiplayer Editing in Figma](https://www.figma.com/blog/multiplayer-editing-in-figma/)
- [Web Form Design (Nielsen Norman Group)](https://www.nngroup.com/articles/web-form-design/)
