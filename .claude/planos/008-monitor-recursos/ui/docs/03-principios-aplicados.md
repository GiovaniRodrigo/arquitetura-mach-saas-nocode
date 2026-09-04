# Applied Principles — Resource Monitor Screen

## 1. Obvious Starting Point

The visual starting point is the **summary card** at the top (`data-principle="inicio-obvio"` in
the wireframe): "7 of 8 services operational" with a large overall-status badge, before any
individual card. The user's eye should answer "is everything OK?" in under 1
second, without scanning all 8 cards. The "Refresh" button stays in the same card, on the right —
a secondary action, not competing with the summary.

## 2. Clear Reversal

There is no destructive action on this screen (it's read-only). What exists is the error state (NFR02): the
global error state's "Try again" button is always visible and requires no navigation —
it keeps the user in the same place. Auto-refresh (FR07) never replaces valid data with a
paralyzing error state: if an update fails, the last valid data remains
visible with a discreet warning, instead of the screen "disappearing" and forcing the user to
start over.

## 3. Consistent Logic

- Reuses `ElevatedCard`/`TonalCard` already used in `Dashboard` and the current `CardServicoStatus` — no
  new visual component is invented.
- Reuses the `Skeleton` from `StateViews.tsx` (already used on other dashboard screens) instead of the
  "Loading…" text the current screen uses in isolation — the same loading pattern across the
  whole application.
- The status indicator (colored dot) keeps the same color semantics (`--destructive` = down)
  used elsewhere in the app (e.g., the user menu badge, `ErrorState`).
- Card hover/focus follow the same `transition-all duration-200` already used in `ElevatedCard`.

## 4. Follow Conventions

- Green = operational, red = unavailable: universal status-page convention (Uptime
  Kuma, GitHub Status, every *status.io*-like) — no new color semantics invented.
- Icons with universal meaning from the `lucide-react` package already in use: `RefreshCw` (refresh),
  `Cpu`, `MemoryStick`/`HardDrive`, `Activity` (RPS/latency) — the same icon library already
  used throughout the sidebar, without introducing a second icon package.
- Domain terminology the technical team already knows: "CPU", "Memory", "Requests/s",
  "Success rate", "p99 latency" — terms already used in the current `CardServicoStatus.tsx`, kept as is.

## 5. Feedback and Milestones

- **Loading**: an 8-rectangle animated skeleton (reusing `Skeleton`), not a generic spinner.
- **Last updated**: "Updated Xs ago" text next to the "Refresh" button, updated on
  every tick — visually confirms auto-refresh (FR07) is active without requiring F5.
- **"Refreshing now" indicator**: the `RefreshCw` icon spins (`animate-spin`) during the
  fetch, including on automatic refreshes — today the screen doesn't distinguish "loading for the
  first time" from "refreshing in the background", which can make the user think it's frozen.
- **Per-card error vs. screen error** (BR01 vs. NFR02): already correctly implemented in the current
  base — kept and visually reinforced with a subtle red border on the individual unavailable card
  (today only the internal text changes), to stay scannable even in a large grid.

## 6. Proximity and Adaptation

- The 5 metrics on each card are grouped into two visual rows: CPU+Memory (process
  resources) and RPS+Success rate+Latency (traffic/mesh) — today they appear as a single list
  with no grouping, making it hard to scan "is this about the process or about the network?".
  See `docs/01-contexto.md`.
- Responsive grid kept (`grid-cols-1 sm:grid-cols-2 lg:grid-cols-3`, already implemented) — 1
  column on mobile, 3 on desktop.
- "Refresh" button and timestamp stay attached to the summary card (same control region),
  physically distant from the data cards (proximity rule: control ≠ data).

## 7. Interface Is Content

- No purely decorative element is added — every color, icon, or bar carries
  information (status, usage trend).
- The (existing) header and sidebar keep their current minimum height — not
  altered by this demand.
- The CPU/memory progress bar replaces part of the loose text without increasing the card's
  height — information per occupied area increases.

## 8. General Visual Design Principles

- **Make the subject obvious**: each card has the service name (`font-heading font-bold`) and a
  category icon (e.g., `Server`) at the top — the name already exists; the icon is added for quick
  visual reinforcement in a dense grid.
- **Appropriate data visualization**: CPU and memory use a **progress bar** (proportion of a
  total, even if the "total" is a reference limit rather than a hard quota) — not a
  line chart, because there is no time series (out of scope, spec §8). RPS/latency
  remain as numbers (they're instantaneous rates, not proportions — a pie/bar doesn't apply).
- **Integrated form and content**: red = alert/unavailable, green = healthy — already the
  adopted convention; extended to the card's left border (4px) instead of just the dot, a
  visual reinforcement that doesn't require looking at the card's corner.
- **Metaphors for new concepts**: there is no new concept on this screen — the terms are
  already familiar to the target technical audience (BR03), so no analogy device is needed.

## 9. Design Decision Matrix

| Decision | Obvious Start | Clear Reversal | Consistency | Convention | Feedback | Proximity | Content > Deco. |
|---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| "X of Y operational" summary card at the top | ✓ | — | ✓ | ✓ | ✓ | ✓ | ✓ |
| Skeleton (reusing `StateViews.tsx`) on loading | — | — | ✓ | ✓ | ✓ | — | ✓ |
| "Updated Xs ago" timestamp | — | — | ✗ (new) | ✓ | ✓ | ✓ | ✓ |
| Spinning icon (`animate-spin`) during refresh | — | — | ✓ | ✓ | ✓ | — | ✓ |
| Status-colored left border on the card | — | — | ✓ | ✓ | ✓ | — | ✓ |
| Progress bar for CPU/memory | — | — | ✓ | ✓ | ✓ | ✓ | ✓ |
| Process vs. traffic grouping within the card | — | — | ✓ | — | — | ✓ | ✓ |
| Global error state with "Try again" (already existing) | — | ✓ | ✓ | ✓ | ✓ | — | — |
