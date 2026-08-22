# Sistema de Cores e Tipografia

> Esta tela **não introduz** paleta ou tipografia nova — reaproveita os tokens já definidos em
> `services/frontend/src/index.css` (Lógica Consistente/Observar Convenções). Documentado aqui
> para referência da demanda, com a adição pontual de cores semânticas de status que a tela
> precisa e que ainda não têm token dedicado.

## Paleta de Cores (tokens existentes, HSL via CSS custom properties)

### Primária
- `--primary`: `hsl(239 84% 67%)` — indigo (referência de mercado: mesma família de azul-violeta
  usada por dashboards SaaS técnicos como Linear/Railway)
- `--primary-foreground`: `hsl(0 0% 100%)`

### Neutras (base do dashboard)
- `--background`: `hsl(210 40% 98%)` (claro) / `hsl(240 10% 4%)` (escuro)
- `--card`: `hsl(0 0% 100%)` (claro) / `hsl(240 7% 8%)` (escuro)
- `--secondary` / `--muted`: `hsl(210 40% 96%)` (claro) / `hsl(240 5% 26%)` (escuro)
- `--border`: `hsl(214 32% 91%)` (claro) / `hsl(240 5% 26%)` (escuro)
- `--muted-foreground`: `hsl(215 16% 47%)` (claro) / `hsl(240 5% 65%)` (escuro)

### Semântica (uso nesta tela)
- **Sucesso / operacional**: `--accent` (`hsl(173 80% 40%)`, teal) já existe no design system
  mas hoje a tela usa `bg-green-500` (Tailwind cru, fora do token system) — **recomendação**:
  usar `--accent` para o indicador "servindo", alinhando ao token existente em vez de uma cor
  Tailwind solta que não responde ao tema.
- **Erro / indisponível**: `--destructive` (`hsl(0 84.2% 60.2%)` claro / `hsl(0 62.8% 30.6%)`
  escuro) — já é o token correto, mas a tela usa `bg-red-500` cru no dot; trocar por
  `bg-destructive` para consistência com `ErrorState`/`StateViews.tsx`.
- **Atenção (uso alto de recurso, ex. barra de CPU > 80%)**: não existe token `--warning` no
  design system atual. Recomendação: usar `amber-500`/`amber-600` do Tailwind (mesma convenção
  de mercado — Grafana, Datadog e Vercel usam âmbar para "warning" entre verde e vermelho) até
  que um token `--warning` seja formalizado no `index.css`; escopo desta demanda não inclui
  alterar o design system global.
- **Info**: `--primary` (indigo) já cobre estados informativos (ex. badge "atualizando").

## Tipografia (famílias existentes, `index.css` linha 1)

- **Títulos** (`font-heading`): Outfit — carregada via Google Fonts no projeto todo, pesos
  400–800. Usada em `<h2>`/`<h3>` dos cards, mantida sem alteração.
- **Corpo**: Inter — pesos 400–700, usada em métricas e texto corrido.
- **Código/Mono**: JetBrains Mono — reservada a trechos monoespaçados (ex. `Ctrl K` no header);
  nesta tela, aplicável opcionalmente aos valores numéricos de métrica (ex. "0.25 núcleos",
  "12.4 ms") para alinhamento tabular mais legível em grid denso — convenção comum em
  dashboards técnicos (Grafana, Datadog usam mono para valores numéricos).

### Escala (já em uso no projeto, mantida)
- Título de card-resumo: `text-2xl font-heading font-bold`
- Título de card de serviço: `text-md font-heading font-bold`
- Métrica (label): `text-sm text-muted-foreground`
- Métrica (valor): `text-sm font-medium` (ou `font-mono` se adotada a recomendação acima)
- Timestamp/legenda: `text-xs text-muted-foreground`

## Espaçamento (Grid)

- Base unit: `4px` (padrão Tailwind, já em uso em todo o projeto — `gap-4`, `p-6`, etc.)
- Raio de borda: `--radius: 0.75rem` como base, mas os cards do design system usam
  `rounded-3xl` (24px) — convenção visual já estabelecida para cards de conteúdo, mantida.
- Grid de cards: `grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4` (já implementado, mantido).
