# Sistema de Cores e Tipografia (recomendado)

Ancorado nas referências populares (Material Design 3; padrões de builders no-code como
Appsmith/Retool; guidelines dos IDPs Google/GitHub). Um único sistema para **todas** as telas
do player, eliminando a inconsistência atual (só o Login tem estilo).

## Paleta de Cores

### Primária (índigo — comum em ferramentas de produtividade/builders)
- 50  `#EEF2FF`
- 100 `#E0E7FF`
- 200 `#C7D2FE`
- 300 `#A5B4FC`
- 400 `#818CF8`
- **500 `#6366F1`** ← principal (CTA)
- 600 `#4F46E5` ← hover do CTA
- 700 `#4338CA`
- 800 `#3730A3`
- 900 `#312E81`

> Referência: tom índigo/violeta é recorrente em internal-tools builders (Retool/Appsmith) e
> no Material 3 (esquema "Indigo"). Alta legibilidade em botão branco-sobre-cor.

### Neutros (superfícies e texto)
- Fundo app `#F8FAFC` · Superfície/card `#FFFFFF` · Borda `#E2E8F0`
- Texto forte `#0F172A` · Texto secundário `#475569` · Texto sutil `#94A3B8`

### Semântica (Material 3 / convenção universal)
- Sucesso `#16A34A` (fundo `#DCFCE7`)
- Alerta  `#D97706` (fundo `#FEF3C7`)
- Erro    `#DC2626` (fundo `#FEE2E2`)
- Info    `#2563EB` (fundo `#DBEAFE`)

### Marcas dos IDPs (obrigatório seguir guideline)
- Google: botão **branco** com borda `#DADCE0`, logo "G" 4 cores, texto `#3C4043`.
- GitHub: botão **preto** `#24292F` (ou branco com borda), texto branco, logo GitHub oficial.

## Tipografia

### Famílias (tendência popular, gratuitas no Google Fonts)
- **Títulos**: `Inter` — uma das fontes UI mais adotadas em SaaS/dashboards
  (dezenas de milhões de downloads/mês no Google Fonts).
- **Corpo**: `Inter` (mesma família, pesos 400/500/600) — consistência e ótima leitura em UI.
- **Código/Mono**: `JetBrains Mono` (para exibir ids/tokens técnicos, ex.: `sistemaId`).

> Fallback sem CDN: `system-ui, -apple-system, "Segoe UI", Roboto, sans-serif` (o player hoje
> já usa `system-ui`; manter como fallback e opcionalmente carregar Inter).

### Escala (base 16px, ratio ~1.25)
- Display / h1: `2rem` / 32px — 700
- h2: `1.5rem` / 24px — 600
- h3: `1.25rem` / 20px — 600
- Body: `1rem` / 16px — 400
- Body-sm: `0.875rem` / 14px — 400
- Caption: `0.75rem` / 12px — 500 (labels, ajuda)

## Espaçamento (Grid)
- **Base unit: 8px** (escala 4/8/12/16/24/32/48/64).
- Colunas: 12 (desktop) · 4 (mobile). Container de conteúdo `max-width: 1120px`.
- Card de login: `max-width: 400px`, padding 32px, `radius: 12px`.

## Tokens (para reuso — CSS custom properties)
```css
:root{
  --color-primary:#6366F1; --color-primary-600:#4F46E5;
  --color-bg:#F8FAFC; --color-surface:#FFFFFF; --color-border:#E2E8F0;
  --color-text:#0F172A; --color-text-2:#475569; --color-text-3:#94A3B8;
  --color-success:#16A34A; --color-warning:#D97706; --color-error:#DC2626; --color-info:#2563EB;
  --space:8px; --radius:12px; --radius-sm:8px;
  --font-ui:"Inter",system-ui,-apple-system,"Segoe UI",Roboto,sans-serif;
  --tap-min:44px;
}
```

## Fontes
- Material Design 3 — https://m3.material.io/
- Sign in with Google best practices — https://developers.google.com/identity/siwg/best-practices
- SaaS login page design 2025 — https://lollypop.design/blog/2025/october/saas-login-page-design/
