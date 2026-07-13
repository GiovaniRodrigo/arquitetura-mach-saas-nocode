# Referências Populares

Todas as referências abaixo pertencem ao **mesmo domínio** do produto (visual/no-code builders, ferramentas de colaboração e design systems de dev-tools) e trazem métrica de popularidade — conforme a regra de escopo do skill (só recomendar o que tem evidência de adoção).

## Produtos e Design Systems

| Referência | URL | Popularidade | Aplicabilidade |
|---|---|---|---|
| **Figma — Multiplayer Editing** | https://www.figma.com/blog/multiplayer-editing-in-figma/ | Avaliação ~US$ 20 bi; co-edição eleva velocidade de projeto ~35%; cursores fluidos com 10+ editores simultâneos | Modelo dos **cursores nomeados coloridos, presença e follow-mode** do Builder (RF06). Diffs otimizados por operação = base conceitual do batching (RNF07) |
| **Vercel Geist Design System** | https://vercel.com/geist/colors | Design system público de referência 2024–25; adotado por milhares de projetos shadcn | **Paleta dark-first**: ink `#171717`, body `#0A0A0A`/`#fafafa`, escala de cinza de 200 passos para bordas/dividers/disabled. Base da paleta neutra do Dashboard/Builder |
| **Linear** | https://linear.app | Benchmark de mercado citado como padrão "dark-first" 2025 | Contenção: 1 cor de acento sobre superfícies quase-pretas; densidade sem ruído; comando `Cmd+K`. Modelo da barra de comando e da hierarquia visual do Dashboard |
| **Retool** | https://retool.com | Líder de mercado em internal tooling | **Layout de builder de 3 colunas** (biblioteca · canvas · propriedades); Inter em tamanhos densos; acento laranja `#EF5350` + azul `#3D5AFE`. Molde direto do Construtor Visual |
| **Framer** | https://www.framer.com | Premiado (Awwwards); prototipagem interativa com animação no núcleo | Preview de transições/microinterações em tempo real no canvas. Referência para o **modo Preview** do Builder e as animações de 60 Hz do Player (RNF07) |
| **WeWeb** | https://www.weweb.io/blog/drag-and-drop-app-builder-tools | Listado entre os "25 melhores drag-and-drop app builders 2026" | Padrões de **publish/ambientes** e lista de projetos multi-tenant. Base do Dashboard e do fluxo Publicar/Rollback (RF04) |
| **Budibase / Appsmith** | https://uibakery.io/blog/drag-and-drop-app-builders | Reviews "top drag-and-drop app builders 2025" (UI Bakery) | Padrões open-source de builder no-code multi-tenant: painel de dados, binding por configuração. Referência do painel de **Regras de Negócio** (RF02) |
| **Typeform / Google Forms** | https://www.typeform.com | Referência de UX de formulários de massa | Modelo do **Headless Player**: um campo em foco, validação inline, indicador de progresso, mobile-first (RF07) |

## Guidelines e Pesquisa de UX

| Fonte | URL | Autoridade | O que aplicar |
|---|---|---|---|
| **Nielsen Norman Group — Form Design** | https://www.nngroup.com/articles/web-form-design/ | Pesquisa de UX com evidência | Labels top-aligned (preenchimento mais rápido); validação inline no *blur*; menos campos = maior taxa de conclusão. Aplicado ao Player (RF07) |
| **Inline Validation — SubUX** | https://subux.pro/guides/article/inline-validation | Guia consolidado | Validar ao **sair do campo (blur)**, não durante a digitação; mensagem específica próxima ao campo. Aplicado ao mapa de erros por `blind_index` (RN08) |
| **Material Design 3** | https://m3.material.io | Google HIG | Estados de componente (hover/focus/disabled), elevação, campos de texto, chips. Convenções do Player e dos controlos do Builder |
| **Apple HIG** | https://developer.apple.com/design/human-interface-guidelines | Apple HIG | Semântica de ícones, hierarquia de ações destrutivas/confirmação (Rollback, Delete) |
| **Multi-Step Form UX (Growform)** | https://www.growform.co/must-follow-ux-best-practices-when-designing-a-multi-step-form/ | Guia de conversão | Indicador "Passo 2 de 4", ordem de campos, remover campos supérfluos. Aplicado a formulários multi-etapa no Player |

## Síntese de decisões extraídas das referências

- **Tema**: dark-first (Geist/Linear) para Dashboard/Builder; light-first para o Player (público leigo).
- **Layout do Builder**: 3 colunas Retool (biblioteca · canvas · propriedades) + camada de colaboração Figma por cima.
- **Tipografia**: Inter (corpo/UI densa, padrão dev-tools) + Geist/Inter Display para títulos; JetBrains Mono para `blind_index` e valores técnicos.
- **Cor de acento**: um único acento (contenção Linear) — violeta/índigo, distinto do azul-genérico, reservado para CTA primário e estado ativo.
- **Formulários**: validação inline no blur (NN/g + SubUX), erros por campo, progresso visível.

---

**Fontes**:
- [Multiplayer Editing in Figma](https://www.figma.com/blog/multiplayer-editing-in-figma/)
- [Vercel Geist — Colors](https://vercel.com/geist/colors)
- [Geist Design System Breakdown (DesignSystems.one)](https://www.designsystems.one/design-systems/vercel-geist)
- [Drag-and-Drop App Builders (WeWeb)](https://www.weweb.io/blog/drag-and-drop-app-builder-tools)
- [Top Drag and Drop app builders 2025 (UI Bakery)](https://uibakery.io/blog/drag-and-drop-app-builders)
- [Inline Validation UX (SubUX)](https://subux.pro/guides/article/inline-validation)
- [Multi-Step Form UX Best Practices (Growform)](https://www.growform.co/must-follow-ux-best-practices-when-designing-a-multi-step-form/)
