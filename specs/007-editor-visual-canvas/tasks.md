# Tarefas: Editor Visual (Canvas)

Documentação retroativa (as-built) — todas as tarefas abaixo já estão
implementadas e com a suíte de testes do Frontend passando (331/331 em
`npm run test`, `npx tsc --noEmit` limpo). Organizadas por fase na ordem em
que foram construídas nesta e nas sessões anteriores que fecharam o RF09.

## Fase 1 — Canvas com árvore real (substitui a casca da spec 004)

- [x] 1. Proto: adicionar `ListarDesigns` ao `DesignEngineService` (`proto/construtor/design/v1/design.proto`)
- [x] 2. Design Engine: `store.Listar` + `grpc.ListarDesigns` (`services/design/internal/store`, `services/design/internal/server/grpc.go`)
- [x] 3. Gateway: rota `GET /api/v1/designs?sistema_id=` (`services/gateway/internal/routes/designs.go`, `services/gateway/internal/app/router.go`)
- [x] 4. Frontend: proxy do Vite para `/socket` (WebSocket do Collab) (`services/frontend/vite.config.ts`)
- [x] 5. Frontend: tipos `Design`/`DesignResumo` + métodos de `ApiClient` (`services/frontend/src/api/types.ts`, `services/frontend/src/api/client.ts`)
- [x] 6. Frontend: `treeOps.ts` — mutações puras da árvore (`add_child`/`update_props`/`move`/`remove`), espelhando `Session.Tree.apply/2` do Collab (`services/frontend/src/collab/treeOps.ts`)
- [x] 7. Frontend: `useTelas.ts` (lista/cria telas do sistema) e `useCanvasDesign.ts` (estado local + `CollabClient` + undo/redo) (`services/frontend/src/systems/useTelas.ts`, `services/frontend/src/systems/useCanvasDesign.ts`)
- [x] 8. Frontend: reescrever `AbaTelas.tsx` para orquestrar canvas real (substitui a casca vazia da spec 004) (`services/frontend/src/pages/Dashboard/abas/AbaTelas.tsx`)
- [x] 9. Frontend: `componentRegistry.ts` — catálogo de tipos com ícone, categoria, `aceitaFilhos`, `temTexto`, `propriedadesPadrao` (`services/frontend/src/systems/componentRegistry.ts`)
- [x] 10. Frontend: `Canvas.tsx` — renderização recursiva, drag&drop nativo (reorder + soltar da paleta), indicadores `before`/`after`/`inside`, alça de resize (`services/frontend/src/pages/Dashboard/editor/Canvas.tsx`)
- [x] 11. Frontend: `Inspector.tsx` — painel de propriedades seccionado por tipo (`services/frontend/src/pages/Dashboard/editor/Inspector.tsx`)
- [x] 12. Frontend: `EditorTopbar.tsx`, `PainelComponentes.tsx`/`PainelLayers.tsx` (`services/frontend/src/pages/Dashboard/editor/`)
- [x] 13. Frontend: undo/redo (`Ctrl+Z`/`Ctrl+Shift+Z`) sobre as mutações locais (`services/frontend/src/pages/Dashboard/abas/AbaTelas.tsx`)
- [x] 14. Collab: mutação `move` com índice em `add_child`/`move` (`services/collab/lib/collab/session/tree.ex`, `services/collab/test/collab/session/tree_test.exs`)

## Fase 2 — Rich text por trecho (Proximity toolbar)

- [x] 15. `sanitizeHtml.ts` — sanitizador baseado em DOM real (`<template>` + allowlist de tags/estilos), com `TAGS_CONTEUDO_DESCARTADO` para `SCRIPT`/`STYLE`/`NOSCRIPT` (não "desembrulhar", descartar por completo) (`services/frontend/src/systems/sanitizeHtml.ts`, `sanitizeHtml.test.ts` — 11 testes)
- [x] 16. `FloatingToolbar.tsx` — toolbar flutuante via portal, posicionada relativa à seleção de texto, `onMouseDown` com `preventDefault` para não perder a seleção ao clicar (`services/frontend/src/pages/Dashboard/editor/FloatingToolbar.tsx`)
- [x] 17. `TextoEditavel` em `Canvas.tsx` — `contentEditable`, duplo clique entra em edição, `execCommand('bold'|'italic'|'underline')` no trecho selecionado, commit sanitizado no blur/Escape, `e.stopPropagation()` no keydown para não disparar atalhos globais durante a digitação (`services/frontend/src/pages/Dashboard/editor/Canvas.tsx`)
- [x] 18. `Inspector.tsx` — campo "Texto" usa `htmlParaTextoPlano` para exibir o conteúdo sem markup, com dica apontando para o duplo clique no canvas (`services/frontend/src/pages/Dashboard/editor/Inspector.tsx`)
- [x] 19. `PreviewRenderer.tsx` — `sanitizarHtml` + `dangerouslySetInnerHTML` nos tipos com texto (heading/paragrafo/botao/link/badge/checkbox/radio/avatar) (`services/frontend/src/pages/Dashboard/editor/PreviewRenderer.tsx`)
- [x] 20. Testes: edição inline entra em modo de edição no duplo clique; blur salva HTML sanitizado via `update_props`, incluindo caso de `<script>` sendo removido no commit (`services/frontend/src/pages/Dashboard/abas/AbaTelas.test.tsx`)

## Fase 3 — Posicionamento livre

- [x] 21. `Estilos.posicao`/`x`/`y` + `ESTILOS_BASE = { posicao: 'relative', x: 0, y: 0 }` (posição base já necessária para `absolute` dos filhos funcionar relativo ao pai direto) (`services/frontend/src/systems/componentRegistry.ts`)
- [x] 22. `Canvas.tsx` — `iniciarMove`/`movePreview` (mesmo racional de `iniciarResize`), alça no canto superior esquerdo visível só quando `ativo && ehAbsoluto`, `draggable={!editandoTexto && !ehAbsoluto}` (`services/frontend/src/pages/Dashboard/editor/Canvas.tsx`)
- [x] 23. `Inspector.tsx` — campos numéricos X/Y, renderizados só quando `estilos.posicao === 'absolute'` (`services/frontend/src/pages/Dashboard/editor/Inspector.tsx`)
- [x] 24. `AbaTelas.tsx` — `handleMoverAbsoluto` (mutação `update_props` com `x`/`y`/`posicao: 'absolute'`) (`services/frontend/src/pages/Dashboard/abas/AbaTelas.tsx`)
- [x] 25. Testes: alça de mover só aparece em componente selecionado com `posicao: absolute`; campos X/Y aparecem/escondem conforme `posicao` (`services/frontend/src/pages/Dashboard/abas/AbaTelas.test.tsx`, `services/frontend/src/pages/Dashboard/editor/Inspector.test.tsx`)

## Fase 4 — Layout inline (`Display: inline`/`inline-block`)

- [x] 26. `componentRegistry.ts` — `Estilos.display` estendido de `'block' | 'flex' | 'none'` para incluir `'inline' | 'inline-block'` (`services/frontend/src/systems/componentRegistry.ts`)
- [x] 27. `Inspector.tsx` — dropdown "Display" passa a oferecer as 5 opções (`services/frontend/src/pages/Dashboard/editor/Inspector.tsx`)
- [x] 28. `Canvas.tsx` — artboard raiz deixa de ser `flex flex-col` (blockificava todo filho, ignorando o `display` escolhido) e passa a fluxo de bloco normal com espaçamento via `[&>*+*]:mt-2` (`services/frontend/src/pages/Dashboard/editor/Canvas.tsx`)
- [x] 29. `Canvas.tsx` — wrapper de cada nó (`data-testid="layer-*"`) passa a espelhar `Estilos.display` (`exibicaoExterna`), não só o elemento interno — sem isso o wrapper (sempre `block` por padrão) continuava forçando quebra de linha mesmo com o conteúdo interno `inline-block` (bug descoberto em verificação ao vivo, ver `plan.md` §2.3) (`services/frontend/src/pages/Dashboard/editor/Canvas.tsx`)
- [x] 30. Teste de regressão: dois componentes com `Display: inline-block` ficam lado a lado — asserção sobre `style.display` do **wrapper** (`layer-b1`/`layer-b2`), não só do elemento interno (`services/frontend/src/pages/Dashboard/abas/AbaTelas.test.tsx`)
- [x] 31. Verificação ao vivo (Chrome, `javascript_tool` + `getBoundingClientRect`): dois botões com `Display: inline-block` passam de empilhados (mesma X, Y diferente) para lado a lado (mesma Y, X adjacente) após o fix

## Fase 5 — Scrollbars e paleta de componentes

- [x] 32. Classe utilitária `.scrollbar-app` (Firefox via `scrollbar-width`/`scrollbar-color`, Chrome/Safari via `::-webkit-scrollbar*`) seguindo os tokens de cor do design system (`services/frontend/src/index.css`)
- [x] 33. Aplicar `.scrollbar-app` aos containers `overflow-y-auto` do painel esquerdo (`PainelComponentes.tsx`, `PainelLayers.tsx`)
- [x] 34. Grid do editor (`220px_1fr_260px` → `280px_1fr_260px`) para o painel esquerdo mostrar o nome completo dos componentes sem truncar (`services/frontend/src/pages/Dashboard/abas/AbaTelas.tsx`)
- [x] 35. `title={item.rotulo}` em cada botão da paleta — tooltip nativo para os rótulos que ainda truncam (`services/frontend/src/pages/Dashboard/editor/PainelComponentes.tsx`, `PainelComponentes.test.tsx`)

## Fase 6 — Modos Edição/Visualização/Foco

- [x] 36. `ModoEditor = 'edicao' | 'visualizacao' | 'foco'` substitui os estados independentes `focoAtivo`/`previewAberto` — exclusividade por construção (RN09.4) (`services/frontend/src/pages/Dashboard/editor/EditorTopbar.tsx`, `services/frontend/src/pages/Dashboard/abas/AbaTelas.tsx`)
- [x] 37. Persistência do modo via `useSessionStorageState('mach:sistema:<id>:modo', 'edicao')`, sobrevivendo à troca entre sub-abas Telas/Regras/Versão (`services/frontend/src/pages/Dashboard/abas/AbaTelas.tsx`)
- [x] 38. `SeletorModo` — controle segmentado (pill), `role="group"`, botões `aria-pressed`, atalho "F" alternando só Edição↔Foco (redesenhado de toggles individuais para pill após validação visual com o usuário) (`services/frontend/src/pages/Dashboard/editor/EditorTopbar.tsx`)
- [x] 39. Guarda nos atalhos globais (`Ctrl+Z`, "F", `Delete`) para não disparar quando o foco está num campo `contentEditable` (RN09.5) (`services/frontend/src/pages/Dashboard/abas/AbaTelas.tsx`)
- [x] 40. Testes do seletor de modo (`role="button"`/`aria-pressed`, exclusividade) e do preview (`modo === 'visualizacao'`) (`services/frontend/src/pages/Dashboard/abas/AbaTelas.test.tsx`)

## Fase 7 — Catálogo: 8 componentes novos (paridade PrimeReact)

- [x] 41. `componentRegistry.ts` — entradas de Avatar, Textarea, Radio, ToggleSwitch, Breadcrumb, Alerta, Spinner, Skeleton (`services/frontend/src/systems/componentRegistry.ts`, `componentRegistry.test.ts`)
- [x] 42. `Canvas.tsx` — preview de cada tipo (`BreadcrumbPreview`, `SpinnerPreview` com `animate-spin`, `ToggleSwitchPreview` somente leitura, `AlertaPreview` com esquema de cor por variante; Avatar/Skeleton reaproveitam os caminhos de imagem/`animate-pulse` já existentes) (`services/frontend/src/pages/Dashboard/editor/Canvas.tsx`)
- [x] 43. `Inspector.tsx` — `SecaoBreadcrumb`, `SecaoToggle`, `SecaoAlerta`, `SecaoAvatar` (Radio/Textarea/Spinner/Skeleton reaproveitam seções genéricas existentes) (`services/frontend/src/pages/Dashboard/editor/Inspector.tsx`, `Inspector.test.tsx`)
- [x] 44. `PreviewRenderer.tsx` — 8 `case`s novos para o site publicado, incluindo Toggle clicável (estado local) e Breadcrumb/Spinner/Skeleton com o mesmo comportamento visual do canvas (`services/frontend/src/pages/Dashboard/editor/PreviewRenderer.tsx`)

## Fase 8 — Operacional: porta de desenvolvimento

- [x] 45. `vite.config.ts` — `server.port: 5183` + `server.strictPort: true` (sem isso o Vite cai silenciosamente para 5173) (`services/frontend/vite.config.ts`)
- [x] 46. Corrigir as 3 referências a `5173` divergentes: `build/dev-up.sh`, `USAGE.md`, `services/frontend/playwright.config.ts`

## Encerramento

- [x] 47. `npx tsc --noEmit` limpo e `npx vitest run` — 331/331 testes passando (`services/frontend/`)
- [x] 48. Verificação manual via Chrome de cada item desta spec (rich text por trecho persistindo via debounce do Collab, posicionamento livre, layout inline, scrollbars, seletor de modo, porta 5183) — sem tarefa aberta de UI pendente de checagem ao vivo
