# Color and Typography System

> This screen **does not introduce** a new palette or typography — it reuses the tokens already defined in
> `services/frontend/src/index.css` (Consistent Logic/Follow Conventions). Documented here
> for reference on this demand, with the specific addition of semantic status colors the screen
> needs and that don't yet have a dedicated token.

## Color Palette (existing tokens, HSL via CSS custom properties)

### Primary
- `--primary`: `hsl(239 84% 67%)` — indigo (market reference: the same blue-violet family
  used by technical SaaS dashboards like Linear/Railway)
- `--primary-foreground`: `hsl(0 0% 100%)`

### Neutrals (dashboard base)
- `--background`: `hsl(210 40% 98%)` (light) / `hsl(240 10% 4%)` (dark)
- `--card`: `hsl(0 0% 100%)` (light) / `hsl(240 7% 8%)` (dark)
- `--secondary` / `--muted`: `hsl(210 40% 96%)` (light) / `hsl(240 5% 26%)` (dark)
- `--border`: `hsl(214 32% 91%)` (light) / `hsl(240 5% 26%)` (dark)
- `--muted-foreground`: `hsl(215 16% 47%)` (light) / `hsl(240 5% 65%)` (dark)

### Semantic (usage on this screen)
- **Success / operational**: `--accent` (`hsl(173 80% 40%)`, teal) already exists in the design system,
  but the screen currently uses `bg-green-500` (raw Tailwind, outside the token system) — **recommendation**:
  use `--accent` for the "serving" indicator, aligning with the existing token instead of a loose
  Tailwind color that doesn't respond to the theme.
- **Error / unavailable**: `--destructive` (`hsl(0 84.2% 60.2%)` light / `hsl(0 62.8% 30.6%)`
  dark) — already the correct token, but the screen uses raw `bg-red-500` on the dot; swap for
  `bg-destructive` for consistency with `ErrorState`/`StateViews.tsx`.
- **Warning (high resource usage, e.g. CPU bar > 80%)**: there is no `--warning` token in the
  current design system. Recommendation: use Tailwind's `amber-500`/`amber-600` (the same market
  convention — Grafana, Datadog, and Vercel use amber for "warning" between green and red) until
  a `--warning` token is formalized in `index.css`; changing the global design system is out of
  this demand's scope.
- **Info**: `--primary` (indigo) already covers informational states (e.g., the "refreshing" badge).

## Typography (existing families, `index.css` line 1)

- **Titles** (`font-heading`): Outfit — loaded via Google Fonts across the whole project, weights
  400–800. Used in the cards' `<h2>`/`<h3>`, kept unchanged.
- **Body**: Inter — weights 400–700, used in metrics and running text.
- **Code/Mono**: JetBrains Mono — reserved for monospaced snippets (e.g., `Ctrl K` in the header);
  on this screen, optionally applicable to numeric metric values (e.g., "0.25 cores",
  "12.4 ms") for more legible tabular alignment in a dense grid — a common convention in
  technical dashboards (Grafana, Datadog use mono for numeric values).

### Scale (already in use in the project, kept)
- Summary-card title: `text-2xl font-heading font-bold`
- Service-card title: `text-md font-heading font-bold`
- Metric (label): `text-sm text-muted-foreground`
- Metric (value): `text-sm font-medium` (or `font-mono` if the above recommendation is adopted)
- Timestamp/caption: `text-xs text-muted-foreground`

## Spacing (Grid)

- Base unit: `4px` (Tailwind default, already in use across the whole project — `gap-4`, `p-6`, etc.)
- Border radius: `--radius: 0.75rem` as the base, but the design system's cards use
  `rounded-3xl` (24px) — visual convention already established for content cards, kept as is.
- Card grid: `grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4` (already implemented, kept).
