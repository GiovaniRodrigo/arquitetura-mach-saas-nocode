# Contexto do Projeto

## Domínio

**Construtor de Sistemas MACH V4** é uma plataforma **Low-Code/No-Code multi-tenant** (SaaS) que permite a utilizadores construir aplicações digitais através de uma interface visual, sem escrever código. A arquitetura segue os pilares **MACH** (Microservices, API-first, Cloud-native, Headless), com 5 microsserviços core, colaboração em tempo real (Elixir/Phoenix), publicação/rollback instantâneos e exportação assíncrona de dados.

O produto pertence à categoria de **visual app builders / internal tooling builders** — o mesmo espaço competitivo de Retool, Webflow, Framer, Builder.io, Plasmic, Appsmith e Budibase. O diferencial técnico está na segurança por design (Blind Index / anonimização LGPD, isolamento multi-tenant) e na colaboração em tempo real estilo Figma.

### Superfícies de UI identificadas na spec

| Superfície | Atores | RF associados | Estado na spec |
|---|---|---|---|
| **Dashboard de Sistemas** — lista de sistemas do tenant, versões, status de publicação | Criador, Administrador | RF01, RF04 | Implícita (frente do produto) |
| **Construtor Visual (Builder)** — canvas drag-and-drop, árvore de componentes, painel de propriedades, colaboração em tempo real com cursores/presença/bloqueio | Criador, Colaborador | RF01, RF02, RF06 | Editor visual fora do escopo *back-end*, mas é a face do produto |
| **Headless Player** — renderizador de formulários dinâmicos publicados, com validação por Blind Index | Cliente Final | RF07 | **Em escopo** (SPA web) |
| Painéis de apoio — Publicar/Rollback, Exportação, Permissões | Criador, Administrador | RF04, RF05, RF03 | Em escopo (contratos REST) |

> **Nota de escopo**: a spec `001` cobre o back-end, contratos gRPC e o Headless Player. O editor visual *drag-and-drop* é declarado como "demanda própria" (spec §8). Este trabalho de UI antecipa o design dessas telas para orientar a demanda futura, mantendo o Headless Player (em escopo) como entrega prioritária.

## Público-Alvo

Três perfis distintos, com necessidades de UI opostas — o design precisa servir os três sem comprometer nenhum:

1. **Criador/Colaborador** (técnico-intermédio): monta sistemas no Builder. Espera densidade de informação, atalhos de teclado, feedback imediato, colaboração fluida. Referência mental: Figma, Retool, Linear.
2. **Administrador (Dono/Parceiro)**: gere tenants hierárquicos, permissões por componente, exportações. Espera clareza, auditabilidade, controlo. Referência mental: painéis de governança enterprise.
3. **Cliente Final** (leigo): apenas preenche formulários publicados via Headless Player. Espera simplicidade, zero fricção, validação clara. Referência mental: Typeform, Google Forms.

Contexto de uso: **desktop-first** para Builder/Dashboard (trabalho de construção em telas grandes), **mobile-first** para o Headless Player (formulários consumidos em qualquer dispositivo).

## Referências Visuais Encontradas

| Referência | Por que é relevante | Métrica de popularidade |
|---|---|---|
| **Figma — Multiplayer / Live Cursors** | Padrão-ouro de colaboração em tempo real (cursores nomeados coloridos, presença, follow-mode) — diretamente aplicável ao RF06 | Avaliação ~US$ 20 bi; co-edição aumenta velocidade de projeto em ~35% (Product Brief/Medium) |
| **Vercel Geist Design System** | Sistema *dark-first* de referência: preto puro `oklch(0 0 0)`, escala de cinza de 200 passos, fonte Geist | Design system público muito copiado de 2024–25; base de milhares de projetos shadcn |
| **Linear** | Estudo de contenção: superfícies quase-pretas + 1 cor de acento; densidade sem ruído | Benchmark de "dark-first" citado como padrão de mercado 2025 |
| **Retool** | Builder de ferramentas internas: canvas + painel de propriedades + Inter em tamanhos densos; acento laranja `#EF5350` + azul `#3D5AFE` | Líder de mercado em internal tooling; referência direta de layout de builder |
| **Framer** | Canvas com animação/microinterações no núcleo; preview de transições em tempo real | Plataforma premiada (Awwwards) para prototipagem interativa |
| **WeWeb / Budibase / Appsmith** | Padrões de builder no-code multi-tenant (lista de apps, ambientes, publish) | Reviews "best drag-and-drop app builders 2025/2026" (WeWeb, UI Bakery, Zapier) |

## Tendências Identificadas (2025–2026)

1. **Dark-first como padrão, não como toggle** — Linear e Vercel tratam o dark como superfície canônica; o light é a alternativa. Ferramentas de construção passam longas horas em tela; o dark reduz fadiga. *(Aplicar: Dashboard e Builder dark-first; Player claro por padrão para o Cliente Final leigo.)*
2. **Colaboração como comunicação** — o cursor deixou de ser presença e virou intenção: seguir cursor, comentar in-context, bloqueio visual de componentes. *(Aplicar direto ao RF06/RN07.)*
3. **AI-assisted building** — geração de UI por prompt/imagem (Mendix Maia, Framer) e sugestões contextuais de layout. *(Reservar espaço no Builder para um "AI panel" futuro.)*
4. **Visual + code (governança enterprise)** — auditabilidade, SSO, versionamento visível, importação de componentes custom. *(Aplicar ao painel de versões/rollback — RF04 — e à hierarquia de tenants.)*
5. **Performance-focused defaults** — apps visuais precisam carregar rápido; batching de render (RNF07: lotes de 16ms) e skeleton screens em vez de spinners.

---

**Fontes**:
- [Figma's Collaborative Canvas (Medium/Product Brief)](https://medium.com/@productbrief/figmas-collaborative-canvas-how-real-time-design-built-a-20-billion-creative-empire-efefc6126a93)
- [Figma's Live Cursor UI (Designilo, 2025)](https://designilo.com/2025/07/20/figmas-live-cursor-ui-a-new-era-of-design-dev-collaboration/)
- [Vercel Geist — Colors](https://vercel.com/geist/colors)
- [Geist Design System Breakdown (DesignSystems.one)](https://www.designsystems.one/design-systems/vercel-geist)
- [19 Best Dark Mode Dashboard Templates 2026 (AdminLTE)](https://adminlte.io/blog/dark-dashboard-templates/)
- [25 Best Drag-and-Drop App Builders (WeWeb)](https://www.weweb.io/blog/drag-and-drop-app-builder-tools)
- [10 Top Drag and Drop app builders 2025 (UI Bakery)](https://uibakery.io/blog/drag-and-drop-app-builders)
