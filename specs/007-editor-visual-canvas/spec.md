# Especificação: Editor Visual (Canvas) — RF09

A spec 004 (`specs/004-reestruturacao-ia-navegacao`) introduziu a aba "Telas" como
**casca de navegação**: rota, layout de 3 colunas (sidebar/canvas/propriedades) e
estado vazio — o `plan.md` dessa spec (§2.3, Riscos) registrou explicitamente que
o editor visual completo (árvore real de componentes, seleção, drag-and-drop,
formatação de texto, posicionamento livre) ficava para uma "spec própria
subsequente", análoga ao que `001 §8` já previa.

Esta spec documenta esse editor como construído: o RF09 completo, do canvas com
árvore real sincronizada por colaboração em tempo real (spec 001, RF06) até o
catálogo de componentes, a formatação de texto por trecho e o posicionamento
livre. Documentação retroativa (as-built) — a implementação já existe no
código listado em cada tarefa de `tasks.md`.

---

## 1. Objetivo

Um Criador/Colaborador autenticado, dentro da aba "Telas" de um sistema, deve
conseguir montar uma tela inteira via manipulação direta no canvas — arrastar
componentes da paleta, reordenar, redimensionar, posicionar livremente,
editar texto rico por trecho selecionado, alternar entre os modos
Edição/Visualização/Foco — com as mudanças sincronizadas em tempo real entre
colaboradores (spec 001, RF06) e persistidas no Design Engine via
write-behind debounced.

---

## 2. Requisitos Funcionais

| ID | Descrição | Ator | Prioridade |
|----|-----------|------|------------|
| RF09.1 | Canvas renderiza a árvore real de componentes de uma tela (`Componente.componente_filhos` recursivo), não mais um placeholder — cada nó é selecionável, mostra os filhos aceitos (`aceitaFilhos`) e usa o catálogo de tipos de `componentRegistry.ts`. | Criador/Colaborador | Alta |
| RF09.2 | Adicionar componente por dois caminhos: clique na paleta (aba "Componentes" do painel esquerdo) ou drag-and-drop nativo da paleta para dentro do canvas, com indicadores visuais de destino (`before`/`after`/`inside`). | Criador/Colaborador | Alta |
| RF09.3 | Reordenar/reparentar um componente existente por drag-and-drop dentro do canvas (mutação `move`, com índice). | Criador/Colaborador | Alta |
| RF09.4 | Redimensionar um componente selecionado por alça de arraste no canto inferior direito (mutação `update_props` com `largura`/`altura`). | Criador/Colaborador | Média |
| RF09.5 | Posicionar um componente livremente (fora do fluxo normal) quando `Estilos.posicao = "absolute"`: campos numéricos X/Y no Inspector e alça de arraste no canto superior esquerdo do componente selecionado (RN09.1). | Criador/Colaborador | Média |
| RF09.6 | Formatar texto por **trecho selecionado** (não o componente inteiro) em qualquer tipo com `temTexto: true`: negrito/itálico/sublinhado via barra de formatação flutuante ("Proximity toolbar") que aparece perto da seleção ao marcar um trecho dentro do texto em edição (duplo clique entra em modo de edição). | Criador/Colaborador | Média |
| RF09.7 | O HTML de texto rico é sempre sanitizado (allowlist de tags/atributos) antes de ser persistido e antes de ser renderizado no site publicado (RN09.2). | Sistema | Alta |
| RF09.8 | Alternar entre 3 modos de edição, mutuamente exclusivos: **Edição** (padrão, manipulação completa), **Visualização** (abre o preview do site publicado, `PreviewOverlay`) e **Foco** (esconde os painéis laterais, mantendo o canvas editável). Atalho de teclado "F" alterna Edição↔Foco. O modo escolhido persiste entre trocas de sub-aba (Telas/Regras/Versão) via `sessionStorage`. | Criador/Colaborador | Média |
| RF09.9 | Desfazer/refazer (`Ctrl+Z`/`Ctrl+Shift+Z`) qualquer mutação de árvore aplicada nesta sessão do editor. | Criador/Colaborador | Média |
| RF09.10 | Mudanças no canvas são propagadas em tempo real a outros colaboradores conectados à mesma tela via canal Phoenix `screen:<screen_id>` (spec 001, RF06) e persistidas no Design Engine com debounce de 5s (write-behind). | Criador/Colaborador | Alta |
| RF09.11 | O painel esquerdo (paleta de componentes) exibe o nome completo de cada tipo sem truncar (largura fixa suficiente) e um tooltip nativo (`title`) com o rótulo completo ao passar o mouse, para os itens cujo nome ainda excede a largura visível. | Criador/Colaborador | Baixa |
| RF09.12 | O catálogo de componentes cobre 32 tipos agrupados em Layout, Texto, Formulário e Mídia — incluindo os 8 adicionados nesta spec para fechar lacunas frente a bibliotecas de referência (PrimeReact): Avatar, Textarea, Radio, ToggleSwitch, Breadcrumb, Alerta, Spinner, Skeleton. | Criador/Colaborador | Média |

---

## 3. Regras de Negócio

| ID | Regra |
|----|-------|
| RN09.1 | Um componente só recebe a alça/campos de posicionamento livre quando `Estilos.posicao === "absolute"`; no modo normal (`static`/`relative`, padrão de `ESTILOS_BASE`) ele segue o fluxo de bloco do pai. O wrapper de cada nó no canvas é `position: relative` por padrão, para que `position: absolute` dos filhos seja relativo ao nó pai direto, nunca ao artboard inteiro. |
| RN09.2 | Sanitização de HTML é feita via parsing real de DOM (elemento `<template>` + allowlist de tags `B/STRONG/I/EM/U/BR/SPAN` e das propriedades de `style` `font-weight`/`font-style`/`text-decoration`), nunca por regex — tags não permitidas como `SCRIPT`/`STYLE`/`NOSCRIPT` são descartadas por completo (conteúdo incluído), as demais são "desembrulhadas" mantendo os filhos. |
| RN09.3 | O artboard raiz do canvas usa fluxo de bloco normal (não flex) e cada wrapper de nó reflete o `Estilos.display` do próprio nó (`inline`/`inline-block` viram `display: inline-block` no wrapper, não só no elemento interno) — do contrário um pai flex "blockifica" os filhos e ignora o `display` escolhido, e um filho com `display: inline-block` cujo *wrapper* continua `block` nunca fica lado a lado com o vizinho (ver `plan.md` §3 para o detalhamento do bug e do fix). |
| RN09.4 | Os 3 modos do RF09.8 são mutuamente exclusivos — nunca mais de um ativo ao mesmo tempo — representados como um controle segmentado (pill), não como toggles independentes. |
| RN09.5 | Atalhos de teclado do editor (`Ctrl+Z`, `F`, `Delete`) são suspensos enquanto o foco está num campo `contentEditable` (edição de texto rico em andamento), para não conflitar com a digitação/atalhos nativos do navegador dentro do texto. |
| RN09.6 | A porta de desenvolvimento do Frontend é fixada em `5183` (`vite.config.ts`, `server.port` + `strictPort: true`) — sem essa fixação o Vite cai silenciosamente para `5173`, divergindo do restante da documentação/scripts (`USAGE.md`, `build/dev-up.sh`). |

---

## 4. Critérios de Aceitação

1. Adicionar dois componentes com `Display: inline-block` ao artboard raiz faz com que eles ocupem a mesma linha (mesma coordenada Y), lado a lado — não empilhados.
2. Selecionar um trecho de texto dentro de um componente em edição (duplo clique) e clicar "Negrito" na barra flutuante aplica a formatação **apenas ao trecho selecionado**, preservando o restante do texto sem a formatação.
3. Um `<script>` (ou qualquer tag fora da allowlist) colado/digitado no texto rico nunca aparece no HTML persistido nem no HTML renderizado no site publicado.
4. Marcar um componente como `posicao: absolute` faz aparecer os campos X/Y no Inspector e a alça de mover no canto superior esquerdo quando selecionado; arrastar a alça atualiza `x`/`y` via mutação `update_props`.
5. Trocar entre os 3 modos do topbar nunca deixa mais de um marcado como ativo (`aria-pressed`); o atalho "F" alterna somente entre Edição e Foco.
6. `make dev` inicia o Frontend na porta `5183` — igual ao valor documentado em `USAGE.md` e usado por `build/dev-up.sh`.
7. Os 8 componentes novos (Avatar, Textarea, Radio, ToggleSwitch, Breadcrumb, Alerta, Spinner, Skeleton) aparecem na paleta, têm preview funcional no canvas (inclusive Spinner girando e Skeleton pulsando) e renderizam corretamente no preview do site publicado (`PreviewRenderer.tsx`).
8. Passar o mouse sobre um item truncado da paleta (ex.: "Contai...") mostra o nome completo via tooltip nativo do navegador.
