# Sistema de Cores e Tipografia

Fundamentado nas referências populares (Vercel Geist, Linear, Retool) e nas tendências dark-first 2025–26. Estratégia dual: **dark-first** para Dashboard/Builder (público técnico, sessões longas) e **light-first** para o Headless Player (Cliente Final leigo). Os tokens abaixo alimentam diretamente os wireframes.

## Paleta de Cores

### Neutros — escala base (referência: Vercel Geist, escala de cinza deliberada)

Superfície canônica **dark** (Dashboard/Builder):

| Token | Hex | Uso |
|---|---|---|
| `--bg-base` | `#0A0A0A` | Fundo raiz da aplicação |
| `--bg-surface` | `#141414` | Cards, painéis, barras laterais |
| `--bg-surface-2` | `#1C1C1C` | Superfície elevada (menus, popovers) |
| `--bg-canvas` | `#101012` | Canvas do Builder |
| `--border` | `#262626` | Bordas e dividers |
| `--border-strong` | `#333333` | Bordas de foco/hover |
| `--text-primary` | `#FAFAFA` | Título e corpo principal |
| `--text-secondary` | `#A1A1A1` | Texto secundário/labels |
| `--text-muted` | `#6B6B6B` | Placeholders, disabled |

Superfície **light** (Headless Player):

| Token | Hex | Uso |
|---|---|---|
| `--bg-base` | `#FFFFFF` | Fundo |
| `--bg-surface` | `#FAFAFA` | Card do formulário |
| `--border` | `#E5E5E5` | Bordas de campo |
| `--text-primary` | `#171717` | Ink (Geist) |
| `--text-secondary` | `#525252` | Labels/ajuda |

### Primária / Acento (referência: contenção Linear — um único acento)

Violeta-índigo, distinto do azul genérico, reservado a CTA primário e estado ativo:

| Passo | Hex | Uso |
|---|---|---|
| 50  | `#EEF0FF` | Fundo de destaque suave (light) |
| 100 | `#DCE0FF` | Chips/badges de acento |
| 200 | `#B8C0FF` | Bordas de acento suaves |
| 300 | `#8E9BFF` | Hover em superfícies escuras |
| 400 | `#6E7BFF` | Ícones ativos |
| **500** | **`#5B63F5`** | **Cor primária — CTA, foco, seleção** |
| 600 | `#4A50DB` | Hover do CTA |
| 700 | `#3B40B0` | Pressed |
| 800 | `#2E3288` | — |
| 900 | `#242766` | Texto sobre fundo de acento claro |

### Cores Semânticas (forma reforça significado — princípio 8)

| Papel | Hex (dark) | Hex (light) | Uso |
|---|---|---|---|
| **Sucesso** (publicado/salvo) | `#3FB950` | `#16A34A` | Versão ativa, "salvo", submissão OK |
| **Alerta** (rascunho/pendente) | `#D29922` | `#CA8A04` | Rascunho não publicado, job coletando |
| **Erro** (falha/destrutivo) | `#F85149` | `#DC2626` | Erro de validação, ação destrutiva |
| **Info** (neutro) | `#58A6FF` | `#2563EB` | Dicas, links, traces |

### Cores de colaboração (referência: Figma multiplayer — cores por utilizador)

Paleta de cursores/avatares, atribuídas ciclicamente por utilizador presente (RF06):

`#F97316` · `#EC4899` · `#8B5CF6` · `#06B6D4` · `#22C55E` · `#EAB308` · `#EF4444` · `#3B82F6`

> Cada cor acompanha o cursor nomeado, a borda do componente em edição e o avatar de presença — a mesma cor liga as três representações do utilizador (consistência, princípio 3).

## Tipografia

### Famílias (baseado em tendência dev-tools popular)

- **Títulos/Display**: **Geist** (Vercel) ou fallback **Inter Display** — padrão dos design systems dev-tools mais copiados de 2024–25.
- **Corpo/UI**: **Inter** — a fonte de interface densa mais usada em dev-tools (Retool, Linear); ~milhões de downloads/mês no Google Fonts, top-3 histórico.
- **Mono/Código**: **JetBrains Mono** ou **Geist Mono** — para `blind_index`, `trace_id`, valores técnicos e badges de versão.

```css
--font-display: "Geist", "Inter", system-ui, sans-serif;
--font-body: "Inter", system-ui, -apple-system, sans-serif;
--font-mono: "JetBrains Mono", "Geist Mono", ui-monospace, monospace;
```

### Escala tipográfica (base 16px · razão ~1.25)

| Nível | rem | px | Peso | Uso |
|---|---|---|---|---|
| Display | 2.25rem | 36px | 600 | Título de tela vazia / hero do Player |
| h1 | 1.75rem | 28px | 600 | Título de página (Dashboard) |
| h2 | 1.375rem | 22px | 600 | Título de seção / nome do sistema |
| h3 | 1.125rem | 18px | 500 | Subtítulos, cabeçalho de card |
| body-lg | 1rem | 16px | 400 | Corpo do Player (leigo — maior) |
| body | 0.875rem | 14px | 400 | Corpo/UI padrão (Builder/Dashboard) |
| caption | 0.75rem | 12px | 500 | Labels, metadados, ajuda |
| mono | 0.8125rem | 13px | 500 | `blind_index`, versões, traces |

Altura de linha: 1.5 para corpo, 1.2 para títulos. Labels top-aligned no Player (NN/g: preenchimento mais rápido).

## Espaçamento (Grid)

- **Base unit: 4px** (escala: 4 · 8 · 12 · 16 · 24 · 32 · 48 · 64).
- **Colunas**: 12 (desktop) · 8 (tablet) · 4 (mobile).
- **Raio de borda**: `--radius-sm: 6px` (campos/botões), `--radius-md: 10px` (cards/painéis), `--radius-lg: 14px` (modais), `--radius-full: 9999px` (avatares/chips).
- **Altura de linhas de UI**: header 48px · linha de tabela 44px · alvo de toque mínimo 44px (Player).
- **Elevação** (dark): sombras discretas + borda de 1px — a borda faz o trabalho de separação, não a sombra (estética Geist).

## Tokens CSS de referência (usados nos wireframes)

```css
:root {
  /* dark-first — Dashboard/Builder */
  --bg-base: #0A0A0A;
  --bg-surface: #141414;
  --bg-surface-2: #1C1C1C;
  --bg-canvas: #101012;
  --border: #262626;
  --border-strong: #333333;
  --text-primary: #FAFAFA;
  --text-secondary: #A1A1A1;
  --text-muted: #6B6B6B;

  --accent: #5B63F5;
  --accent-hover: #4A50DB;
  --accent-soft: rgba(91,99,245,0.12);

  --success: #3FB950;
  --warning: #D29922;
  --danger:  #F85149;
  --info:    #58A6FF;

  --space-unit: 4px;
  --radius-sm: 6px;
  --radius-md: 10px;
  --radius-lg: 14px;

  --font-body: "Inter", system-ui, sans-serif;
  --font-mono: "JetBrains Mono", ui-monospace, monospace;
}
```

---

**Fontes**:
- [Vercel Geist — Colors](https://vercel.com/geist/colors)
- [Geist Design System Breakdown (DesignSystems.one)](https://www.designsystems.one/design-systems/vercel-geist)
- [19 Best Dark Mode Dashboard Templates 2026 (AdminLTE)](https://adminlte.io/blog/dark-dashboard-templates/)
- [Multiplayer Editing in Figma](https://www.figma.com/blog/multiplayer-editing-in-figma/)
- [Web Form Design (Nielsen Norman Group)](https://www.nngroup.com/articles/web-form-design/)
