# Contexto do Projeto — Gráficos na Dashboard (`/dashboard`)

## Domínio

MAYS — Make Your SaaS (arquitetura MACH V4): plataforma SaaS no-code que permite a um
usuário ("dono"/"parceiro") criar sistemas sem programar, geri-los por cliente/tenant, e
acompanhar a operação do próprio negócio. A rota pedida (`/ui/dashboard` → equivalente a
`/dashboard` no app real, ver `pages/Dashboard/Dashboard.tsx`) é a **home operacional**: o
usuário cai nela a cada login e precisa responder rápido a "como está meu negócio hoje?".

## Público-Alvo

Donos/parceiros semi-técnicos (não programadores, mas confortáveis com métricas de
negócio: contagem, receita, tendência). Não é um perfil de analista de dados — os gráficos
precisam comunicar em segundos, sem exigir leitura de eixo ou legenda complexa
(alinhado ao público já mapeado em `001-construtor-sistemas-mach-v4` e reafirmado em
`auditoria-ui-projeto`).

## Diagnóstico — por que "faltam gráficos"

Leitura de `src/pages/Dashboard/Dashboard.tsx` e de todo `src/dashboard/*` confirma o
relato do usuário: **nenhum componente do dashboard usa visualização de dados** — só
números isolados e listas:

| Componente | Dado exibido hoje | Formato |
|---|---|---|
| Métrica "Sistemas" | `sistemas_total` (int) | Número grande, sem contexto de variação |
| Métrica "Publicados" / "Rascunhos" | `—` (placeholder, não implementado) | Texto estático |
| `CardResumoFinanceiro` | `receita_total_centavos` + `competencia` (1 mês) | Número grande + legenda |
| `CardFeedback` | Lista de reclamações com badge `pendente`/`respondido` | `<ul>` de texto |
| `CardUltimosAcessos` | Lista de acessos recentes | `<ul>` de texto |

Nenhum hook (`useMetricas`, `useResumoFinanceiro`, `useFeedback`, `useUltimosAcessos`)
retorna série temporal hoje — todos trazem um valor/lista pontual do instante atual. Isso
não impede o design: a tela `Monitor.tsx` (`008-monitor-recursos`) já resolve o mesmo
problema (métrica instantânea sem histórico) com indicadores pontuais bem desenhados, e
este documento propõe o próximo passo natural — adicionar a camada de tendência/proporção
onde o dado já permite (contagens, proporção pendente/respondido) e sinalizar, sem
implementar, onde falta série histórica no backend (receita mês a mês, sistemas
criados ao longo do tempo).

## Referências Visuais Encontradas

- **shadcn/ui Charts** (`ui.shadcn.com/charts`) — biblioteca oficial de gráficos do mesmo
  design system que o projeto já usa (`shadcn` no `package.json`, `components/ui/*`).
  Construída sobre **Recharts**, que sozinho tem **~53M downloads/semana no npm e 27k+
  estrelas no GitHub** — é hoje a combinação mais adotada para dashboards SaaS em
  React + Tailwind, exatamente a stack deste projeto.
- **shadcn/ui** como design system de origem tem **116k+ estrelas no GitHub**, confirmando
  que seguir os padrões visuais oficiais de chart do mesmo sistema (não uma lib genérica)
  é a escolha de menor risco de inconsistência visual.
- **Dribbble — "Dashboard for a Finance SaaS ✦ Twisty"** (HALO LAB, 1.9k likes / 715k
  views) e **"Charts and Tables for Financial SaaS Dashboard"** (Extej UI/UX) — referências
  de dashboards financeiros SaaS com o mesmo tipo de dado que `CardResumoFinanceiro`
  (receita ao longo do tempo): confirmam o padrão "número grande + mini gráfico de área
  logo abaixo" como convenção de mercado para KPI de receita.
- **Nielsen Norman Group — "Data Visualizations for Dashboards"** e **"Clutter-Free
  Charts"**: recomenda eliminar "chartjunk" e funcionalismo mínimo — cada gráfico deve
  responder uma pergunta específica do usuário, não decorar a tela. Também documenta o
  padrão de escaneamento em F (varredura horizontal no topo, depois vertical à esquerda),
  o que reforça manter os KPIs com maior densidade visual na primeira dobra.

## Tendências Identificadas

1. **Sparklines em cards de métrica** — mini gráfico de linha/área sem eixos, encaixado
   sob o número grande, com badge de variação percentual (↑/↓) colorido por
   sucesso/alerta. Praticamente onipresente em dashboards SaaS financeiros (Stripe,
   Linear, os exemplos do shadcn/ui Charts).
2. **Donut/ring chart para proporção binária** — pendente vs. respondido, ativo vs.
   inativo. Substitui contagem mental item a item por leitura instantânea de proporção.
3. **Area chart para série temporal de receita/crescimento** — preenchimento com
   gradiente na cor primária, mantendo o número absoluto como texto de destaque acima do
   gráfico (não como legenda dentro dele).
4. **Funcionalismo mínimo (NN/g)** — sem grid lines pesadas, sem eixo Y numerado quando o
   valor exato já está escrito por extenso ao lado; o gráfico existe para comunicar
   *tendência*, a leitura exata fica no texto.
5. **Estados de carregamento consistentes com o resto do app** — este projeto já tem um
   padrão maduro de `Skeleton`/`EmptyState`/`ErrorState` (`components/ui/StateViews`,
   usado em todos os cards do dashboard). Os novos gráficos devem seguir o mesmo
   contrato, não inventar um padrão de loading próprio.
