# Sistema de Cores e Tipografia — reaproveitado do projeto

O projeto já tem um design system Tailwind + tokens HSL (`src/index.css`,
`tailwind.config.js`) usado por Copilot Chat/Notion-AI-like panels do mercado
(painel neutro, acento de marca só no elemento de destaque). O chat de IA
**não introduz nova paleta** — reaproveita 1:1 os tokens existentes.

## Paleta de Cores (tokens já existentes)

### Primária
- `--primary`: `239 84% 67%` (indigo `#6366f1`) — usado no FAB e na bolha do
  usuário (`bg-primary/10`), mesma cor de destaque do resto do produto
  (botões, sidebar ativa).

### Semântica (reaproveitada, sem criar novas)
- Erro de resposta da IA → `--destructive`
- Streaming/carregando → `--muted` (skeleton)
- Pill de contexto do sistema → `--accent` (teal `173 80% 40%`), mesma cor já
  usada para destaque secundário no projeto.

### Superfícies
- Painel (`Sheet`): `--popover` / `--popover-foreground` (mesmo token do menu
  de usuário no header).
- Bolha do assistente: `--secondary`.
- Bolha do usuário: `--primary` em 10% de opacidade.

## Tipografia (famílias já configuradas em `tailwind.config.js`)

- Título do painel ("Assistente de Design"): `font-heading` (Outfit) — mesma
  fonte de todos os títulos de seção do Dashboard.
- Corpo das mensagens: `font-sans` (Inter).
- Nenhuma fonte nova é adicionada — mantém o download/CDN atual do projeto.

## Espaçamento

- Base unit: `4px`/`8px` (grid já usado no Tailwind do projeto).
- Painel: largura `24rem`–`28rem` em desktop (breakpoint padrão do
  `components/ui/sheet.tsx`), full-width em mobile.
- Raio: `var(--radius)` (0.75rem) em bolhas e no botão FAB, consistente com o
  resto dos componentes (`rounded-2xl` no menu de usuário, `rounded-full` em
  botões de ação do header).
