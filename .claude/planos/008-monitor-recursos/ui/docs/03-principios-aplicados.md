# Princípios Aplicados — Tela Monitor de Recursos

## 1. Início Óbvio

O ponto de partida visual é o **card-resumo** no topo (`data-principle="inicio-obvio"` no
wireframe): "7 de 8 serviços operacionais" com uma badge grande de status geral, antes de
qualquer card individual. O olho do usuário deve responder "está tudo bem?" em menos de 1
segundo, sem escanear os 8 cards. O botão "Atualizar" continua no mesmo card, à direita —
ação secundária, não compete com o resumo.

## 2. Reversão Clara

Não há ação destrutiva nesta tela (é só leitura). O que existe é o estado de erro (RNF02): o
botão "Tentar novamente" do estado de erro global é sempre visível e não exige navegação —
mantém o usuário no mesmo lugar. O auto-refresh (RF07) nunca substitui dados válidos por um
estado de erro paralisante: se uma atualização falha, os últimos dados válidos continuam
visíveis com um aviso discreto, em vez de a tela "sumir" e forçar o usuário a recomeçar.

## 3. Lógica Consistente

- Reusa `ElevatedCard`/`TonalCard` já usados em `Dashboard` e `CardServicoStatus` atual — nenhum
  componente visual novo é inventado.
- Reusa o `Skeleton` de `StateViews.tsx` (já usado em outras telas do dashboard) em vez do texto
  "Carregando…" que a tela atual usa isoladamente — mesmo padrão de loading em toda a aplicação.
- O indicador de status (dot colorido) mantém a mesma semântica de cor (`--destructive` = fora
  do ar) usada em outros pontos do app (ex.: badge do menu do usuário, `ErrorState`).
- Hover/foco dos cards seguem o mesmo `transition-all duration-200` já usado em `ElevatedCard`.

## 4. Observar Convenções

- Verde = operacional, vermelho = indisponível: convenção universal de status pages (Uptime
  Kuma, GitHub Status, todos os *status.io*-likes) — não inventar semântica de cor nova.
- Ícones com significado universal do próprio `lucide-react` já em uso: `RefreshCw` (atualizar),
  `Cpu`, `MemoryStick`/`HardDrive`, `Activity` (RPS/latência) — mesma biblioteca de ícones já
  usada em toda a sidebar, sem introduzir um segundo pacote de ícones.
- Terminologia do domínio que a equipe técnica já conhece: "CPU", "Memória", "Requisições/s",
  "Taxa de sucesso", "Latência p99" — termos já usados em `CardServicoStatus.tsx` atual, mantidos.

## 5. Feedback e Marcos

- **Loading**: skeleton de 8 retângulos animados (reuso de `Skeleton`), não spinner genérico.
- **Última atualização**: texto "Atualizado há Xs" ao lado do botão "Atualizar", atualizado a
  cada tick — confirma visualmente que o auto-refresh (RF07) está ativo sem exigir F5.
- **Indicador de "atualizando agora"**: o ícone `RefreshCw` gira (`animate-spin`) durante o
  fetch, inclusive nos refreshes automáticos — hoje a tela não distingue "carregando pela
  primeira vez" de "atualizando em background", o que pode fazer o usuário achar que travou.
- **Erro por card vs erro de tela** (RN01 vs RNF02): já implementado corretamente na base atual
  — mantido e reforçado visualmente com borda vermelha sutil no card individual indisponível
  (hoje só o texto interno muda), para ficar escaneável mesmo em grid grande.

## 6. Proximidade e Adaptação

- As 5 métricas de cada card ficam agrupadas em duas linhas visuais: CPU+Memória (recursos do
  processo) e RPS+Taxa de sucesso+Latência (tráfego/mesh) — hoje aparecem como uma lista única
  sem agrupamento, dificultando escanear "isso é sobre o processo ou sobre a rede?".
  Ver `docs/01-contexto.md`.
- Grid responsivo mantido (`grid-cols-1 sm:grid-cols-2 lg:grid-cols-3`, já implementado) — 1
  coluna no mobile, 3 no desktop.
- Botão "Atualizar" e timestamp ficam colados ao card-resumo (mesma região de controle),
  fisicamente distantes dos cards de dado (regra de proximidade: controle ≠ dado).

## 7. Interface é Conteúdo

- Nenhum elemento puramente decorativo é adicionado — toda cor, ícone ou barra carrega
  informação (status, tendência de uso).
- Header (já existente) e sidebar (já existente) permanecem com a altura mínima atual — não
  alterados por esta demanda.
- Barra de progresso de CPU/memória substitui parte do texto solto sem aumentar a altura do
  card — informação por área ocupada aumenta.

## 8. Princípios Gerais de Design Visual

- **Torne o assunto óbvio**: cada card tem o nome do serviço (`font-heading font-bold`) e um
  ícone de categoria (ex. `Server`) no topo — já existe o nome, adiciona-se o ícone para reforço
  visual rápido em grid denso.
- **Visualização de dados adequada**: CPU e memória usam **barra de progresso** (proporção de um
  total, mesmo que o "total" seja um limite de referência, não uma quota rígida) — não um
  gráfico de linha, porque não há série temporal (fora de escopo, spec §8). RPS/latência
  continuam como número (são taxas instantâneas, não proporções — pizza/barra não se aplica).
- **Forma e conteúdo integrados**: vermelho = alerta/indisponível, verde = saudável — já é a
  convenção adotada; estendida para a borda esquerda do card (4px) em vez de só o dot, reforço
  visual sem exigir olhar para o canto do card.
- **Metáforas para conceitos novos**: não há conceito novo nesta tela — os termos já são
  familiares ao público técnico-alvo (RN03), não é necessário nenhum recurso de analogia.

## 9. Matriz de Decisão de Design

| Decisão | Início Óbvio | Reversão Clara | Consistência | Convenção | Feedback | Proximidade | Conteúdo > Deco. |
|---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| Card-resumo "X de Y operacionais" no topo | ✓ | — | ✓ | ✓ | ✓ | ✓ | ✓ |
| Skeleton (reuso `StateViews.tsx`) no loading | — | — | ✓ | ✓ | ✓ | — | ✓ |
| Timestamp "Atualizado há Xs" | — | — | ✗ (novo) | ✓ | ✓ | ✓ | ✓ |
| Ícone girando (`animate-spin`) durante refresh | — | — | ✓ | ✓ | ✓ | — | ✓ |
| Borda esquerda colorida por status no card | — | — | ✓ | ✓ | ✓ | — | ✓ |
| Barra de progresso para CPU/memória | — | — | ✓ | ✓ | ✓ | ✓ | ✓ |
| Agrupamento processo vs tráfego dentro do card | — | — | ✓ | — | — | ✓ | ✓ |
| Estado de erro global com "Tentar novamente" (já existente) | — | ✓ | ✓ | ✓ | ✓ | — | — |
