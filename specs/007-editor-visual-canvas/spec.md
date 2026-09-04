# Specification: Visual Editor (Canvas) — FR09

Spec 004 (`specs/004-reestruturacao-ia-navegacao`) introduced the "Screens" tab as a
**navigation shell**: route, 3-column layout (sidebar/canvas/properties), and
empty state — that spec's `plan.md` (§2.3, Risks) explicitly noted that
the full visual editor (real component tree, selection, drag-and-drop,
text formatting, free positioning) was left for a "subsequent dedicated
spec," analogous to what `001 §8` already anticipated.

This spec documents that editor as built: the complete FR09, from the canvas
with a real tree synced via real-time collaboration (spec 001, FR06) to the
component catalog, segment-level text formatting, and free positioning.
Retroactive (as-built) documentation — the implementation already exists in
the code listed in each task in `tasks.md`.

---

## 1. Goal

An authenticated Creator/Collaborator, within a system's "Screens" tab, must
be able to assemble an entire screen via direct manipulation on the canvas —
dragging components from the palette, reordering, resizing, positioning
freely, editing rich text on a selected segment, switching between the
Edit/Preview/Focus modes — with changes synced in real time between
collaborators (spec 001, FR06) and persisted to the Design Engine via
debounced write-behind.

---

## 2. Functional Requirements

| ID | Description | Actor | Priority |
|----|-----------|------|------------|
| FR09.1 | Canvas renders a screen's real component tree (recursive `Componente.componente_filhos`), no longer a placeholder — each node is selectable, shows its accepted children (`aceitaFilhos`), and uses the type catalog from `componentRegistry.ts`. | Creator/Collaborator | High |
| FR09.2 | Add a component via two paths: click in the palette (the "Components" tab of the left panel) or native drag-and-drop from the palette into the canvas, with visual drop-target indicators (`before`/`after`/`inside`). | Creator/Collaborator | High |
| FR09.3 | Reorder/reparent an existing component via drag-and-drop within the canvas (`move` mutation, with index). | Creator/Collaborator | High |
| FR09.4 | Resize a selected component via a drag handle at the bottom-right corner (`update_props` mutation with `largura`/`altura`). | Creator/Collaborator | Medium |
| FR09.5 | Freely position a component (outside the normal flow) when `Estilos.posicao = "absolute"`: numeric X/Y fields in the Inspector and a drag handle at the top-left corner of the selected component (BR09.1). | Creator/Collaborator | Medium |
| FR09.6 | Format text on a **selected segment** (not the whole component) in any type with `temTexto: true`: bold/italic/underline via a floating formatting bar ("Proximity toolbar") that appears near the selection when a segment inside the text being edited is highlighted (double-click enters edit mode). | Creator/Collaborator | Medium |
| FR09.7 | Rich-text HTML is always sanitized (tag/attribute allowlist) before being persisted and before being rendered on the published site (BR09.2). | System | High |
| FR09.8 | Switch between 3 mutually exclusive edit modes: **Edit** (default, full manipulation), **Preview** (opens the published-site preview, `PreviewOverlay`), and **Focus** (hides the side panels, keeping the canvas editable). Keyboard shortcut "F" toggles Edit↔Focus. The chosen mode persists across sub-tab switches (Screens/Rules/Version) via `sessionStorage`. | Creator/Collaborator | Medium |
| FR09.9 | Undo/redo (`Ctrl+Z`/`Ctrl+Shift+Z`) any tree mutation applied in this editor session. | Creator/Collaborator | Medium |
| FR09.10 | Canvas changes are propagated in real time to other collaborators connected to the same screen via the Phoenix channel `screen:<screen_id>` (spec 001, FR06) and persisted to the Design Engine with a 5s debounce (write-behind). | Creator/Collaborator | High |
| FR09.11 | The left panel (component palette) shows the full name of each type without truncation (sufficiently fixed width) and a native tooltip (`title`) with the full label on hover, for items whose name still exceeds the visible width. | Creator/Collaborator | Low |
| FR09.12 | The component catalog covers 32 types grouped into Layout, Text, Form, and Media — including the 8 added in this spec to close gaps against a reference library (PrimeReact): Avatar, Textarea, Radio, ToggleSwitch, Breadcrumb, Alert, Spinner, Skeleton. | Creator/Collaborator | Medium |

---

## 3. Business Rules

| ID | Rule |
|----|-------|
| BR09.1 | A component only gets the free-positioning handle/fields when `Estilos.posicao === "absolute"`; in normal mode (`static`/`relative`, `ESTILOS_BASE`'s default) it follows the parent's block flow. Each node's wrapper on the canvas is `position: relative` by default, so that the children's `position: absolute` is relative to the direct parent node, never to the whole artboard. |
| BR09.2 | HTML sanitization is done via real DOM parsing (a `<template>` element + an allowlist of tags `B/STRONG/I/EM/U/BR/SPAN` and of the `style` properties `font-weight`/`font-style`/`text-decoration`), never by regex — disallowed tags like `SCRIPT`/`STYLE`/`NOSCRIPT` are dropped entirely (content included), the rest are "unwrapped" while keeping their children. |
| BR09.3 | The canvas's root artboard uses normal block flow (not flex) and each node's wrapper reflects that node's own `Estilos.display` (`inline`/`inline-block` become `display: inline-block` on the wrapper, not only on the inner element) — otherwise a flex parent "blockifies" its children and ignores the chosen `display`, and a child with `display: inline-block` whose *wrapper* stays `block` never sits side by side with its neighbor (see `plan.md` §3 for the bug and fix in detail). |
| BR09.4 | The 3 modes in FR09.8 are mutually exclusive — never more than one active at the same time — represented as a segmented control (pill), not as independent toggles. |
| BR09.5 | Editor keyboard shortcuts (`Ctrl+Z`, `F`, `Delete`) are suspended while focus is on a `contentEditable` field (rich-text editing in progress), so they don't conflict with the browser's native typing/shortcuts inside the text. |
| BR09.6 | The Frontend's development port is fixed at `5183` (`vite.config.ts`, `server.port` + `strictPort: true`) — without this pinning, Vite silently falls back to `5173`, diverging from the rest of the documentation/scripts (`USAGE.md`, `build/dev-up.sh`). |

---

## 4. Acceptance Criteria

1. Adding two components with `Display: inline-block` to the root artboard makes them occupy the same line (same Y coordinate), side by side — not stacked.
2. Selecting a text segment inside a component being edited (double-click) and clicking "Bold" in the floating bar applies the formatting **only to the selected segment**, leaving the rest of the text without the formatting.
3. A `<script>` (or any tag outside the allowlist) pasted/typed into the rich text never appears in the persisted HTML nor in the HTML rendered on the published site.
4. Marking a component as `posicao: absolute` makes the X/Y fields appear in the Inspector and the move handle at the top-left corner when selected; dragging the handle updates `x`/`y` via an `update_props` mutation.
5. Switching between the 3 topbar modes never leaves more than one marked as active (`aria-pressed`); the "F" shortcut toggles only between Edit and Focus.
6. `make dev` starts the Frontend on port `5183` — matching the value documented in `USAGE.md` and used by `build/dev-up.sh`.
7. The 8 new components (Avatar, Textarea, Radio, ToggleSwitch, Breadcrumb, Alert, Spinner, Skeleton) appear in the palette, have a functional preview on the canvas (including a spinning Spinner and a pulsing Skeleton), and render correctly in the published-site preview (`PreviewRenderer.tsx`).
8. Hovering over a truncated palette item (e.g., "Contai...") shows the full name via the browser's native tooltip.
