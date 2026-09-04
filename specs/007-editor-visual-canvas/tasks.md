# Tasks: Visual Editor (Canvas)

Retroactive (as-built) documentation — all the tasks below are already
implemented and the Frontend's test suite passes (331/331 in
`npm run test`, `npx tsc --noEmit` clean). Organized by phase, in the order
in which they were built in this and previous sessions that closed FR09.

## Phase 1 — Canvas with a real tree (replaces spec 004's shell)

- [x] 1. Proto: add `ListarDesigns` to `DesignEngineService` (`proto/construtor/design/v1/design.proto`)
- [x] 2. Design Engine: `store.Listar` + `grpc.ListarDesigns` (`services/design/internal/store`, `services/design/internal/server/grpc.go`)
- [x] 3. Gateway: route `GET /api/v1/designs?sistema_id=` (`services/gateway/internal/routes/designs.go`, `services/gateway/internal/app/router.go`)
- [x] 4. Frontend: Vite proxy for `/socket` (Collab's WebSocket) (`services/frontend/vite.config.ts`)
- [x] 5. Frontend: `Design`/`DesignResumo` types + `ApiClient` methods (`services/frontend/src/api/types.ts`, `services/frontend/src/api/client.ts`)
- [x] 6. Frontend: `treeOps.ts` — pure tree mutations (`add_child`/`update_props`/`move`/`remove`), mirroring Collab's `Session.Tree.apply/2` (`services/frontend/src/collab/treeOps.ts`)
- [x] 7. Frontend: `useTelas.ts` (lists/creates a system's screens) and `useCanvasDesign.ts` (local state + `CollabClient` + undo/redo) (`services/frontend/src/systems/useTelas.ts`, `services/frontend/src/systems/useCanvasDesign.ts`)
- [x] 8. Frontend: rewrite `AbaTelas.tsx` to orchestrate the real canvas (replaces spec 004's empty shell) (`services/frontend/src/pages/Dashboard/abas/AbaTelas.tsx`)
- [x] 9. Frontend: `componentRegistry.ts` — type catalog with icon, category, `aceitaFilhos`, `temTexto`, `propriedadesPadrao` (`services/frontend/src/systems/componentRegistry.ts`)
- [x] 10. Frontend: `Canvas.tsx` — recursive rendering, native drag&drop (reorder + drop from the palette), `before`/`after`/`inside` indicators, resize handle (`services/frontend/src/pages/Dashboard/editor/Canvas.tsx`)
- [x] 11. Frontend: `Inspector.tsx` — properties panel sectioned by type (`services/frontend/src/pages/Dashboard/editor/Inspector.tsx`)
- [x] 12. Frontend: `EditorTopbar.tsx`, `PainelComponentes.tsx`/`PainelLayers.tsx` (`services/frontend/src/pages/Dashboard/editor/`)
- [x] 13. Frontend: undo/redo (`Ctrl+Z`/`Ctrl+Shift+Z`) over the local mutations (`services/frontend/src/pages/Dashboard/abas/AbaTelas.tsx`)
- [x] 14. Collab: `move` mutation with index in `add_child`/`move` (`services/collab/lib/collab/session/tree.ex`, `services/collab/test/collab/session/tree_test.exs`)

## Phase 2 — Segment-level rich text (Proximity toolbar)

- [x] 15. `sanitizeHtml.ts` — sanitizer based on a real DOM (`<template>` + tag/style allowlist), with `TAGS_CONTEUDO_DESCARTADO` for `SCRIPT`/`STYLE`/`NOSCRIPT` (drop entirely, don't "unwrap") (`services/frontend/src/systems/sanitizeHtml.ts`, `sanitizeHtml.test.ts` — 11 tests)
- [x] 16. `FloatingToolbar.tsx` — floating toolbar via portal, positioned relative to the text selection, `onMouseDown` with `preventDefault` so the selection isn't lost on click (`services/frontend/src/pages/Dashboard/editor/FloatingToolbar.tsx`)
- [x] 17. `TextoEditavel` in `Canvas.tsx` — `contentEditable`, double-click enters edit mode, `execCommand('bold'|'italic'|'underline')` on the selected segment, sanitized commit on blur/Escape, `e.stopPropagation()` on keydown so global shortcuts don't fire while typing (`services/frontend/src/pages/Dashboard/editor/Canvas.tsx`)
- [x] 18. `Inspector.tsx` — the "Text" field uses `htmlParaTextoPlano` to show the content without markup, with a hint pointing to double-click on the canvas (`services/frontend/src/pages/Dashboard/editor/Inspector.tsx`)
- [x] 19. `PreviewRenderer.tsx` — `sanitizarHtml` + `dangerouslySetInnerHTML` on the types with text (heading/paragraph/button/link/badge/checkbox/radio/avatar) (`services/frontend/src/pages/Dashboard/editor/PreviewRenderer.tsx`)
- [x] 20. Tests: inline editing enters edit mode on double-click; blur saves sanitized HTML via `update_props`, including the case of a `<script>` being stripped on commit (`services/frontend/src/pages/Dashboard/abas/AbaTelas.test.tsx`)

## Phase 3 — Free positioning

- [x] 21. `Estilos.posicao`/`x`/`y` + `ESTILOS_BASE = { posicao: 'relative', x: 0, y: 0 }` (base position already needed for children's `absolute` to work relative to the direct parent) (`services/frontend/src/systems/componentRegistry.ts`)
- [x] 22. `Canvas.tsx` — `iniciarMove`/`movePreview` (same rationale as `iniciarResize`), handle at the top-left corner visible only when `ativo && ehAbsoluto`, `draggable={!editandoTexto && !ehAbsoluto}` (`services/frontend/src/pages/Dashboard/editor/Canvas.tsx`)
- [x] 23. `Inspector.tsx` — numeric X/Y fields, rendered only when `estilos.posicao === 'absolute'` (`services/frontend/src/pages/Dashboard/editor/Inspector.tsx`)
- [x] 24. `AbaTelas.tsx` — `handleMoverAbsoluto` (`update_props` mutation with `x`/`y`/`posicao: 'absolute'`) (`services/frontend/src/pages/Dashboard/abas/AbaTelas.tsx`)
- [x] 25. Tests: move handle only appears on a selected component with `posicao: absolute`; X/Y fields show/hide depending on `posicao` (`services/frontend/src/pages/Dashboard/abas/AbaTelas.test.tsx`, `services/frontend/src/pages/Dashboard/editor/Inspector.test.tsx`)

## Phase 4 — Inline layout (`Display: inline`/`inline-block`)

- [x] 26. `componentRegistry.ts` — `Estilos.display` extended from `'block' | 'flex' | 'none'` to include `'inline' | 'inline-block'` (`services/frontend/src/systems/componentRegistry.ts`)
- [x] 27. `Inspector.tsx` — the "Display" dropdown now offers the 5 options (`services/frontend/src/pages/Dashboard/editor/Inspector.tsx`)
- [x] 28. `Canvas.tsx` — the root artboard stops being `flex flex-col` (which blockified every child, ignoring the chosen `display`) and moves to normal block flow with spacing via `[&>*+*]:mt-2` (`services/frontend/src/pages/Dashboard/editor/Canvas.tsx`)
- [x] 29. `Canvas.tsx` — each node's wrapper (`data-testid="layer-*"`) now mirrors `Estilos.display` (`exibicaoExterna`), not only the inner element — without this the wrapper (always `block` by default) kept forcing a line break even with `inline-block` inner content (bug found during live verification, see `plan.md` §2.3) (`services/frontend/src/pages/Dashboard/editor/Canvas.tsx`)
- [x] 30. Regression test: two components with `Display: inline-block` sit side by side — assertion on the **wrapper's** `style.display` (`layer-b1`/`layer-b2`), not only the inner element's (`services/frontend/src/pages/Dashboard/abas/AbaTelas.test.tsx`)
- [x] 31. Live verification (Chrome, `javascript_tool` + `getBoundingClientRect`): two buttons with `Display: inline-block` go from stacked (same X, different Y) to side by side (same Y, adjacent X) after the fix

## Phase 5 — Scrollbars and component palette

- [x] 32. `.scrollbar-app` utility class (Firefox via `scrollbar-width`/`scrollbar-color`, Chrome/Safari via `::-webkit-scrollbar*`) following the design system's color tokens (`services/frontend/src/index.css`)
- [x] 33. Apply `.scrollbar-app` to the `overflow-y-auto` containers of the left panel (`PainelComponentes.tsx`, `PainelLayers.tsx`)
- [x] 34. Editor grid (`220px_1fr_260px` → `280px_1fr_260px`) so the left panel shows the full component names without truncation (`services/frontend/src/pages/Dashboard/abas/AbaTelas.tsx`)
- [x] 35. `title={item.rotulo}` on each palette button — native tooltip for labels that still get truncated (`services/frontend/src/pages/Dashboard/editor/PainelComponentes.tsx`, `PainelComponentes.test.tsx`)

## Phase 6 — Edit/Preview/Focus modes

- [x] 36. `ModoEditor = 'edicao' | 'visualizacao' | 'foco'` replaces the independent states `focoAtivo`/`previewAberto` — exclusivity by construction (BR09.4) (`services/frontend/src/pages/Dashboard/editor/EditorTopbar.tsx`, `services/frontend/src/pages/Dashboard/abas/AbaTelas.tsx`)
- [x] 37. Mode persistence via `useSessionStorageState('mach:sistema:<id>:modo', 'edicao')`, surviving the switch between the Screens/Rules/Version sub-tabs (`services/frontend/src/pages/Dashboard/abas/AbaTelas.tsx`)
- [x] 38. `SeletorModo` — segmented control (pill), `role="group"`, `aria-pressed` buttons, "F" shortcut toggling only Edit↔Focus (redesigned from individual toggles to a pill after visual validation with the user) (`services/frontend/src/pages/Dashboard/editor/EditorTopbar.tsx`)
- [x] 39. Guard on the global shortcuts (`Ctrl+Z`, "F", `Delete`) so they don't fire when focus is on a `contentEditable` field (BR09.5) (`services/frontend/src/pages/Dashboard/abas/AbaTelas.tsx`)
- [x] 40. Tests for the mode selector (`role="button"`/`aria-pressed`, exclusivity) and the preview (`modo === 'visualizacao'`) (`services/frontend/src/pages/Dashboard/abas/AbaTelas.test.tsx`)

## Phase 7 — Catalog: 8 new components (PrimeReact parity)

- [x] 41. `componentRegistry.ts` — entries for Avatar, Textarea, Radio, ToggleSwitch, Breadcrumb, Alert, Spinner, Skeleton (`services/frontend/src/systems/componentRegistry.ts`, `componentRegistry.test.ts`)
- [x] 42. `Canvas.tsx` — preview for each type (`BreadcrumbPreview`, `SpinnerPreview` with `animate-spin`, read-only `ToggleSwitchPreview`, `AlertaPreview` with a per-variant color scheme; Avatar/Skeleton reuse the existing image/`animate-pulse` paths) (`services/frontend/src/pages/Dashboard/editor/Canvas.tsx`)
- [x] 43. `Inspector.tsx` — `SecaoBreadcrumb`, `SecaoToggle`, `SecaoAlerta`, `SecaoAvatar` (Radio/Textarea/Spinner/Skeleton reuse existing generic sections) (`services/frontend/src/pages/Dashboard/editor/Inspector.tsx`, `Inspector.test.tsx`)
- [x] 44. `PreviewRenderer.tsx` — 8 new `case`s for the published site, including a clickable Toggle (local state) and Breadcrumb/Spinner/Skeleton with the same visual behavior as the canvas (`services/frontend/src/pages/Dashboard/editor/PreviewRenderer.tsx`)

## Phase 8 — Operational: development port

- [x] 45. `vite.config.ts` — `server.port: 5183` + `server.strictPort: true` (without this Vite silently falls back to 5173) (`services/frontend/vite.config.ts`)
- [x] 46. Fix the 3 diverging references to `5173`: `build/dev-up.sh`, `USAGE.md`, `services/frontend/playwright.config.ts`

## Closeout

- [x] 47. `npx tsc --noEmit` clean and `npx vitest run` — 331/331 tests passing (`services/frontend/`)
- [x] 48. Manual verification via Chrome of every item in this spec (segment-level rich text persisting via Collab's debounce, free positioning, inline layout, scrollbars, mode selector, port 5183) — no open UI task pending live checking
