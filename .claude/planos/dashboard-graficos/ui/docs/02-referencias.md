# Referências — Gráficos na Dashboard

| Referência | URL | Popularidade | Aplicabilidade |
|---|---|---|---|
| shadcn/ui Charts (área/linha/donut) | ui.shadcn.com/charts | Extensão oficial do design system já usado no projeto (`shadcn` no `package.json`); base é Recharts, ~53M downloads/semana no npm, 27k+ estrelas no GitHub | Fonte primária de padrão visual e de markup/tokens para todos os gráficos desta demanda — evita reinventar espaçamento, tooltip, legenda |
| shadcn/ui (design system) | ui.shadcn.com | 116k+ estrelas no GitHub | Confirma que o projeto já está no ecossistema certo para adotar charts sem trocar de design system |
| "Dashboard for a Finance SaaS ✦ Twisty" — HALO LAB | dribbble.com (tag saas-analytics) | 1.9k likes / 715k views | Padrão "número grande + area chart com gradiente logo abaixo" para KPI de receita — aplicado em `CardResumoFinanceiro` |
| "Charts and Tables for Financial SaaS Dashboard" — Extej UI/UX | dribbble.com/shots/25399613 | Destaque na tag saas-analytics-dashboard | Confirma combinação de card numérico + gráfico + tabela no mesmo domínio financeiro do `CardResumoFinanceiro` |
| NN/g — "Data Visualizations for Dashboards" | nngroup.com/videos/data-visualizations-dashboards | Nielsen Norman Group — referência de pesquisa de UX com evidência empírica | Base do princípio "cada gráfico responde uma pergunta"; usado para decidir o que **não** virar gráfico (ex.: lista de últimos acessos permanece lista) |
| NN/g — "Clutter-Free Charts" | nngroup.com/videos/chartjunk | Nielsen Norman Group | Base para remover grid lines, eixo Y numerado e legendas redundantes dos sparklines/donut propostos |
