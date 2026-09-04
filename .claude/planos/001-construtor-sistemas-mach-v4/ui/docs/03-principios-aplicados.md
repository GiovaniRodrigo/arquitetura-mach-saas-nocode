# Applied Design Principles

Application of the 9 principles to the three target screens: **Dashboard**, **Visual Builder**, and **Headless Player**. Each principle brings concrete decisions anchored in the spec's FRs/BRs.

---

## 1. Obvious Starting Point

- **Dashboard**: the starting point is the **"+ New System"** button (primary CTA, single accent, top-right corner) and, below it, the featured card of the last edited system. The eye goes from title → `Cmd+K` search → systems grid.
- **Builder**: the starting point is the **central canvas** — the largest surface, always lit; the component library (left) invites the first drag with an *empty state* "Drag a component to start".
- **Player**: the starting point is the **form's first field already in focus** (autofocus) with the "Submit" CTA fixed at the end; in multi-step, step 1 is open with the "Step 1 of N" indicator at the top.

## 2. Clear Reversal

- **Rollback (FR04/BR05)** is the product's mother reversal action: accessible in 1 click from the versions panel, with a **confirmation dialog** ("Revert to version 6? The published system will immediately start serving v6."), showing what changes. Executes in < 100ms (acceptance criterion 3) and confirms via toast.
- **Destructive actions** (delete system, delete component, delete rule) require explicit confirmation; the destructive button is red and never the dialog's default focus.
- **Builder**: every mutation is reversible via **Undo/Redo** (`Cmd+Z` / `Cmd+Shift+Z`); since persistence is write-behind with a 5s debounce (BR06), local undo is instantaneous before the flush.
- **Form (Player)**: "Back" always visible in multi-step flows; no data is lost when going back.

## 3. Consistent Logic

- The same component type (e.g., `text_input`) renders identically on the Builder canvas and in the Player — **rendering parity** guaranteed by the same Composite tree contract (FR01).
- Uniform interaction states across the whole product: `hover` (subtle elevation + accent border), `focus` (accessible 2px focus ring), `disabled` (0.4 opacity, `not-allowed` cursor), `selected` (solid accent border).
- Fixed positions: top bar (tenant context + global actions) always at the top; table row actions always on the right; breadcrumb always below the top bar.

## 4. Follow Conventions

- **Universal semantic icons**: 🗑 trash = delete, ✏️ pencil = edit, ＋ = add, ⎘ = duplicate, ⤺ = revert/rollback, ↑ = publish, ⭳ = export. (Set: Lucide — dev-tools standard.)
- **Material 3 / HIG** for form fields, chips, switches, and dialogs.
- **Domain terminology** the user already knows: "System", "Version", "Publish", "Rollback", "Collaborators", "Rules" — exactly the spec's terms, with no infra jargon (the user never sees raw "gRPC", "GenServer", or "blind_index"; the latter appears only in admin/debug tools as a mono *badge*).
- **`Cmd+K` command palette** (Linear/Figma convention) for navigation and quick actions.

## 5. Feedback and Milestones

- **Explicit states** across every surface: `loading` (skeleton screens, not spinners), `success` (unobtrusive green toast), `error` (toast/red field with actionable message), `empty` (illustration + CTA).
- **Collaboration (FR06)**: presence always visible (stacked avatars at the top), named colored cursors on the canvas, "🔒 being edited by Ana" badge on the locked component (BR07), and a persistence state indicator: **"All changes saved"** ↔ **"Saving…"** (reflects the 5s debounce / `flush_ok`, BR06).
- **Publishing (FR04)**: publish progress bar; on completion, a visible milestone "Version 7 published · live now".
- **Export (FR05)**: since it's asynchronous (immediate 202), shows a job with state `created → collecting → ready`; when ready, a download button with an expiration notice ("link expires in 10 min").
- **Player**: inline validation on *blur* (NN/g); on submit, button spinner → confirmation; server errors (422) mapped to the exact field via `blind_index` (BR08).

## 6. Proximity and Adaptation

- **Grouping**: in the Builder's properties panel, related fields sit in collapsible sections (Layout · Style · Data · Rules) with tighter internal spacing than between sections.
- **Contextual actions near the data**: a floating toolbar appears over the selected component on the canvas (duplicate/delete/lock), not in a distant bar.
- **Responsiveness**:
  - Dashboard: fluid grid (4 → 2 → 1 columns).
  - Builder: honestly **desktop-only** (canvas needs space); on smaller screens, a "Open on a larger screen to edit" notice.
  - Player: **mobile-first**, one field per row, touch targets ≥ 44px.

## 7. Interface Is Content

- **Maximum useful data without scrolling**: the Builder canvas occupies the largest area; collapsible sidebars; 48px-tall header.
- **Zero empty decoration**: no ornamental gradients, heavy shadows, or illustrations that don't communicate state. The accent color is reserved — when everything is colored, nothing has priority (Linear restraint).
- **Density calibrated by audience**: dense in Builder/Dashboard (technical); airy in the Player (layperson).

## 8. General Visual Design Principles

- **Make the subject obvious**: every screen opens with a title + context icon (system name in the Builder; form name in the Player; "My Systems" in the Dashboard).
- **Appropriate data visualization**: version history as a **vertical timeline** (temporal sequence); export job states as a **stepper**; queue depth/usage as numeric cards — not decorative charts.
- **Integrated form and content**: consistent color semantics — green = success/published, amber = draft/pending, red = error/destructive, info-blue = neutral informational, accent = primary action/active.
- **Metaphors for new concepts**: **Composite tree** represented as a file *tree view* (familiar metaphor); **Blind Index** as a lock/hash (privacy); **active version** as an "on air" switch.

## 9. Design Decision Matrix

| Decision | Obvious Start | Clear Reversal | Consistency | Convention | Feedback | Proximity | Content > Deco. |
|---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| "+ New System" CTA (Dashboard) | ✓ | ✗ | ✓ | ✓ | ✓ | ✓ | ✓ |
| System card w/ version status | ✓ | ✗ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Builder's 3-column layout | ✓ | ✗ | ✓ | ✓ | ✗ | ✓ | ✓ |
| Collaboration cursors/presence (FR06) | ✗ | ✗ | ✓ | ✓ | ✓ | ✓ | ✓ |
| "🔒 being edited" badge (BR07) | ✗ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| "Saving…/Saved" indicator (BR06) | ✗ | ✓ | ✓ | ✓ | ✓ | ✗ | ✓ |
| Versions panel + Rollback (FR04) | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Rollback confirmation dialog | ✗ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Player form + inline validation (FR07) | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Per-field error via blind_index (BR08) | ✗ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Export job (stepper + expiration) | ✓ | ✗ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `Cmd+K` command palette | ✓ | ✗ | ✓ | ✓ | ✓ | ✗ | ✓ |

Legend: ✓ = principle applied in this decision; ✗ = not the focus of this decision (applied by another).
