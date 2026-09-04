# Color and Typography System — reused from the project

The project already has a Tailwind design system + HSL tokens (`src/index.css`,
`tailwind.config.js`) used by market Copilot-Chat/Notion-AI-like panels
(neutral panel, brand accent only on the highlighted element). The AI chat
**does not introduce a new palette** — it reuses the existing tokens 1:1.

## Color Palette (already-existing tokens)

### Primary
- `--primary`: `239 84% 67%` (indigo `#6366f1`) — used on the FAB and the
  user bubble (`bg-primary/10`), the same accent color as the rest of the
  product (buttons, active sidebar).

### Semantic (reused, none created)
- AI response error → `--destructive`
- Streaming/loading → `--muted` (skeleton)
- System context pill → `--accent` (teal `173 80% 40%`), the same color
  already used for secondary emphasis in the project.

### Surfaces
- Panel (`Sheet`): `--popover` / `--popover-foreground` (same token as the
  user menu in the header).
- Assistant bubble: `--secondary`.
- User bubble: `--primary` at 10% opacity.

## Typography (families already configured in `tailwind.config.js`)

- Panel title ("Design Assistant"): `font-heading` (Outfit) — the same
  font as every Dashboard section title.
- Message body: `font-sans` (Inter).
- No new font is added — keeps the project's current download/CDN as is.

## Spacing

- Base unit: `4px`/`8px` (grid already used in the project's Tailwind).
- Panel: `24rem`–`28rem` width on desktop (default breakpoint from
  `components/ui/sheet.tsx`), full-width on mobile.
- Radius: `var(--radius)` (0.75rem) on bubbles and the FAB button, consistent
  with the rest of the components (`rounded-2xl` on the user menu,
  `rounded-full` on header action buttons).
