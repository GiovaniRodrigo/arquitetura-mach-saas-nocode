# Contexto do Projeto

## Domínio

MACH V4 é a plataforma SaaS no-code "MAYS — Make Your SaaS": um construtor visual de sistemas
(telas, componentes, fluxos) composto por 8 microsserviços (IAM, Design, Logic, Deploy, Export,
Workers, Collab, Gateway) rodando em Kubernetes com service mesh Linkerd.

A tela **Monitor de Recursos** (`/dashboard/monitor`, spec `008-monitor-recursos`/`009`) é uma
tela de **observabilidade interna de infraestrutura**: mostra, para cada um dos 8 serviços, se
está no ar, CPU, memória, requisições/s, taxa de sucesso e latência p99 — dados vindos do
metrics-server do Kubernetes e do Prometheus do Linkerd-viz via `services/gateway/internal/meshmetrics`.
Não é uma tela de produto voltada ao cliente final do SaaS — é uma tela operacional, equivalente
a um status board de DevOps/SRE embutido no próprio dashboard da plataforma.

## Público-Alvo

Qualquer usuário autenticado do dashboard (RN03 — não existe hoje papel de "administrador de
plataforma" separado), mas o *uso real* esperado é de perfil técnico: quem opera/mantém a
plataforma MACH V4, olhando a tela para diagnosticar "algo está fora do ar?" ou "algum serviço
está sob carga?". A linguagem e a densidade de informação devem seguir o padrão de dashboards de
observabilidade (Grafana, Datadog, Vercel, Railway) — não o padrão de telas de produto para
usuário leigo.

## Stack Frontend (já implementada)

- **React + TypeScript**, Tailwind CSS com tokens HSL via CSS custom properties (`--primary`,
  `--secondary`, `--destructive`, `--muted-foreground`, etc.), suporte a tema claro/escuro
  (`.dark` no root, alternância já implementada em `DashboardLayout.tsx`).
- **Sistema de componentes "M3"** (`src/components/m3/`): `ElevatedCard` (fundo `--card`,
  `rounded-3xl`, sombra), `TonalCard` (fundo `--secondary`, sem sombra, para destaque de seção),
  `FabButton`, `NavPill` — nomenclatura inspirada em Material Design 3 (elevated/tonal/filled),
  mas aplicada sobre um visual mais próximo de dashboards SaaS modernos (cantos muito
  arredondados, `shadow-sm`, paleta indigo/teal) do que ao Material puro do Android.
  `src/components/ui/`: primitives shadcn-like (`button`, `dialog`, `sheet`, `sidebar`,
  `switch`, `tooltip`) + `StateViews.tsx` já padroniza `Skeleton`/`EmptyState`/`ErrorState`
  reutilizados pelas telas do dashboard.
- **Ícones**: `lucide-react` (mesmo pacote usado em toda a sidebar/header).
- **Tipografia**: Inter (corpo), Outfit (`font-heading`, títulos), JetBrains Mono (código/atalhos
  como `Ctrl K`).
- Tela atual (`Monitor.tsx` + `CardServicoStatus.tsx` + `useRecursos.ts`) já implementa: card de
  cabeçalho tonal com botão "Atualizar", grid de cards por serviço (indicador verde/vermelho +
  lista de métricas), estado de erro único de página (RNF02) e auto-refresh a cada 10s (RF07).
  Este documento propõe refinamentos visuais sobre essa base, não uma reconstrução.

## Referências Visuais Encontradas

| Referência | Métrica de popularidade | Por que é relevante |
|---|---|---|
| [Uptime Kuma](https://github.com/louislam/uptime-kuma) | 90,1k stars no GitHub | Referência dominante em dashboards de status de serviço self-hosted; grid de cards por serviço com indicador de cor, mesma forma de agregação "N serviços, 1 fora do ar" que a tela do Monitor precisa comunicar. |
| [Vercel Dashboard / Geist Design System](https://vercel.com/blog/dashboard-redesign) | Design system oficial de uma das plataformas dev mais usadas do mercado (referência de facto para dashboards técnicos) | Mostra que, para público técnico, a cor deve ser reservada ao status (verde/vermelho/âmbar) — o resto da UI é neutro (escala de cinza), evitando "ruído" visual competindo com o dado. |
| [Railway Observability Dashboard](https://docs.railway.com/observability) | Plataforma PaaS popular entre devs (referência comparativa recorrente com Fly.io/Vercel em benchmarks de mercado) | Cards de métrica com gráfico/indicador visual leve (não só número) para CPU/memória/rede — reforça uso de barra de progresso ou sparkline em vez de número solto. |
| [Grafana](https://grafana.com) | Ferramenta de observabilidade mais usada no mercado (base de comparação de todo dashboard de métricas) | Painel de "single stat" com cor semântica de fundo/borda conforme threshold — inspira o uso de cor no *card inteiro*, não só no indicador, quando um serviço está indisponível. |
| [Nielsen Norman Group — Dashboard Design](https://www.nngroup.com/articles/dashboard-design/) | Pesquisa de UX com evidência (referência acadêmica, não estética) | Fundamenta a hierarquia: status geral > exceções > detalhes — a tela deve deixar óbvio primeiro "quantos serviços saudáveis" antes de exigir leitura card a card. |

## Tendências Identificadas

1. **Cor reservada ao significado**: em dashboards técnicos populares (Vercel, Grafana), a UI é
   majoritariamente neutra/monocromática — cor é usada *só* para status (verde/vermelho/âmbar),
   nunca decorativa. A tela atual já segue isso parcialmente (dot verde/vermelho); pode reforçar
   estendendo a cor para a borda/fundo sutil do card quando indisponível.
2. **Resumo agregado no topo**: dashboards de status (Uptime Kuma, status pages) sempre mostram
   primeiro um resumo ("7/8 operacional") antes da grade detalhada — atende ao princípio de
   Início Óbvio e à pesquisa da NN/g sobre hierarquia "visão geral → exceção → detalhe".
3. **Métricas como barra/indicador visual, não só número**: Railway e Grafana usam barra de
   progresso ou sparkline para CPU/memória, permitindo escanear "está alto?" sem fazer conta
   mental — mais rápido que ler "0.25 núcleos" isoladamente.
4. **Skeleton loading em vez de texto "Carregando…"**: já existe `Skeleton` em
   `StateViews.tsx`, usado por outras telas do dashboard — a tela Monitor deveria reusá-lo
   (consistência, Lógica Consistente) em vez do parágrafo de texto atual.
5. **Timestamp de última atualização**: dashboards com auto-refresh (Grafana, Vercel, Railway)
   sempre mostram "atualizado há Xs" perto do botão de refresh — comunica que o auto-refresh
   (RF07) está de fato funcionando, sem o usuário precisar adivinhar.
