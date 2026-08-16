# Sistema de Cores e Tipografia — Gráficos na Dashboard

> Não propõe tokens novos. `services/frontend/src/index.css` já define `--primary`,
> `--success`, `--warning`, `--destructive` e `--muted` (light + dark) — os únicos
> tokens que um sparkline/donut/bar chart minimalista precisa. Esta demanda só documenta
> **como mapear esses tokens para papéis de gráfico**, seguindo a convenção do
> shadcn/ui Charts (referência: `02-referencias.md`).

## Paleta de Cores (reaproveitada de `index.css`)

### Séries neutras (contagem sem "bom/ruim", ex.: sistemas criados)
- Linha/preenchimento: `hsl(var(--primary))` — mesma cor do CTA e do hero card
- Gradiente de área: `hsl(var(--primary) / 0.25)` → `hsl(var(--primary) / 0)`

### Séries com polaridade (variação percentual, receita)
- Positiva: `hsl(var(--success))`
- Negativa: `hsl(var(--destructive))`
- Neutra/pendente: `hsl(var(--warning))`

### Donut (proporção binária)
- Fatia "pendente": `hsl(var(--warning))`
- Fatia "respondido": `hsl(var(--success))`
- Trilho de fundo (parte vazia): `hsl(var(--muted))`

### Estrutura do gráfico (linhas de apoio, quando houver)
- Grid/eixo (só quando estritamente necessário — ver princípio "Interface é Conteúdo"):
  `hsl(var(--border))`
- Texto de valor/tooltip: `hsl(var(--foreground))`
- Texto de legenda secundária: `hsl(var(--muted-foreground))`

## Tipografia

Sem alterações — `Inter` (corpo) e `Outfit` (`font-heading`, já usada nos `<h3>` de cada
card) continuam sendo as únicas famílias envolvidas. Números grandes dos KPIs mantêm
`text-3xl font-heading font-bold`, já usado hoje em `Dashboard.tsx`. Rótulos de eixo/tick
(quando existirem) usam `text-[11px] text-muted-foreground` — mesmo tamanho do `<kbd>`
do atalho Ctrl+K no header, já a menor escala tipográfica em uso no app.

## Espaçamento

- Sparkline: altura fixa de `40px`, largura 100% do card, sem padding interno próprio
  (herda o `p-*` do `ElevatedCard`).
- Donut: `64px` de diâmetro, centralizado à esquerda do bloco de legenda/total.
- Mini bar chart (Resumo Financeiro): `48px` de altura, barras com `4px` de gap,
  `radius` de topo `2px` (ecoa `--radius: 0.75rem` do sistema, escalado para o tamanho
  miniatura do elemento).
