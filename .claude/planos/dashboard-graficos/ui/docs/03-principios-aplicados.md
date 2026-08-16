# Princípios Aplicados — Gráficos na Dashboard

## Proposta concreta (o que muda em cada card existente)

| Card | Hoje | Proposta | Gráfico |
|---|---|---|---|
| Métricas topo (Sistemas / Publicados / Rascunhos) | Número isolado, sem contexto | Número + variação percentual (badge ↑/↓) + sparkline dos últimos 30 dias | Sparkline de área, sem eixo |
| `CardResumoFinanceiro` | Número da competência atual + texto | Número em destaque + mini gráfico de barras dos últimos 6 meses abaixo | Barras compactas (mini bar chart) |
| `CardFeedback` | Só lista de itens | Donut pendente/respondido acima da lista (lista continua sendo a ação, o donut é o resumo) | Donut/ring chart |
| `CardUltimosAcessos` | Lista de eventos | **Mantém lista, sem gráfico** — é um feed de eventos discretos, não uma métrica agregável; forçar gráfico aqui violaria "Interface é Conteúdo" (dado de log não vira decoração) | — |

> Nota de dado: sparklines de métrica e o gráfico de receita mensal exigem série
> histórica que os hooks atuais (`useMetricas`, `useResumoFinanceiro`) não retornam hoje.
> O wireframe já modela o **estado vazio** desse cenário (ver `dashboard.html`,
> `data-principle="feedback-marcos"` na seção de sparkline) para o caso do backend ainda
> não expor histórico — o card degrada para o formato atual (só número), nunca quebra.

## 1. Início Óbvio
O hero card "Construa seu próximo sistema" continua sendo o ponto de partida visual
(maior contraste de cor, CTA primário) — os gráficos entram *depois* dele na hierarquia,
como contexto, não competindo pela primeira ação do usuário.

## 2. Reversão Clara
Gráficos são somente leitura — não há ação destrutiva neles. O único controle interativo
(toggle de período do sparkline de receita, ex.: 6M/12M) não precisa de confirmação, é
uma troca de visualização, não uma mutação de dado.

## 3. Lógica Consistente
Todo sparkline usa a mesma altura (40px), a mesma ausência de eixo, e a mesma regra de
cor: verde/`--success` quando a variação é positiva, `--destructive` quando negativa,
`--primary` quando é uma série neutra (ex.: contagem de sistemas, que não tem "bom" ou
"ruim" categórico). O donut do `CardFeedback` reaproveita as mesmas cores já usadas nos
badges da lista abaixo dele (`--warning` para pendente, `--success` para respondido) —
mesmo dado, duas representações, mesma paleta.

## 4. Observar Convenções
Segue o padrão de mercado documentado em `02-referencias.md`: sparkline sob o número
(não ao lado), donut com o total no centro, area chart com gradiente suave na cor
primária. Ícones de tendência usam as setas já disponíveis via `lucide-react`
(`TrendingUp`/`TrendingDown`), biblioteca já usada no projeto (`DashboardLayout.tsx`).

## 5. Feedback e Marcos
Os gráficos herdam o contrato de estados já usado em todo o dashboard
(`components/ui/StateViews`: `Skeleton`, `EmptyState`, `ErrorState`):
- **Carregando**: skeleton retangular no lugar do gráfico (mesma altura final, evita
  layout shift).
- **Vazio/sem histórico**: card degrada para o formato atual (só número), sem espaço em
  branco quebrado.
- **Erro**: mesmo `ErrorState` com botão "Tentar novamente" já usado nos outros cards.

## 6. Proximidade e Adaptação
Sparkline fica imediatamente abaixo do número que ele explica (não em outra coluna).
Em mobile, os 3 cards de métrica empilham (`grid-cols-1`, já é o comportamento atual) e o
sparkline mantém largura total do card — nunca é cortado ou exige scroll horizontal.

## 7. Interface é Conteúdo
Nenhum gráfico é puramente decorativo: cada um substitui uma pergunta que o usuário faria
mentalmente ("subiu ou desceu?", "quantos estão pendentes?") por uma resposta visual
imediata. `CardUltimosAcessos` foi deliberadamente **excluído** da proposta por não ter
essa pergunta agregada por trás.

## 8. Princípios Gerais de Design Visual
- **Assunto óbvio**: cada gráfico fica sob um `<h3>` que já nomeia a métrica — não precisa
  de título próprio.
- **Visualização adequada ao dado**: linha/área para tendência temporal (sparkline,
  receita), donut para proporção de duas partes de um todo (pendente/respondido) — nunca
  pizza para mais de 2-3 categorias, nunca barra para série temporal contínua curta.
- **Forma reforça significado**: verde = crescimento/sucesso, vermelho = queda,
  âmbar = pendência — mesmo vocabulário de cor que o resto do app já usa
  (`--success`/`--warning`/`--destructive` de `index.css`).

## 9. Matriz de Decisão de Design

| Decisão | Início Óbvio | Reversão Clara | Consistência | Convenção | Feedback | Proximidade | Conteúdo > Deco. |
|---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| Sparkline nos cards de métrica | ✓ | ✓ (read-only) | ✓ | ✓ | ✓ | ✓ | ✓ |
| Mini bar chart em Resumo Financeiro | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Donut em Feedback | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Gráfico em Últimos Acessos (rejeitado) | — | — | — | — | — | — | ✗ (sem pergunta agregada) |
