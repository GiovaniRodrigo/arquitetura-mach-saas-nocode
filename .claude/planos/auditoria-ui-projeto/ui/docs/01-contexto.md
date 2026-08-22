# Contexto do Projeto — Auditoria UI de Escopo Global

> Esta auditoria amplia o escopo do `/ui` de uma tela isolada (última execução: `008-monitor-recursos`)
> para **todo o frontend** do projeto. Não repete o que já foi documentado — inventaria a cobertura
> existente e foca no que ainda não tinha passado por uma análise de UI, além de registrar
> **inconsistências reais encontradas ao ler o código de ponta a ponta** (só visíveis numa varredura
> de projeto inteiro, não tela a tela).

## Domínio

MAYS — Make Your SaaS (arquitetura MACH V4): plataforma SaaS no-code que permite a um usuário
("dono"/"parceiro") criar sistemas (telas, regras de negócio, dados) sem programar, geri-los por
cliente/tenant, e operar a própria plataforma (monitoramento de infraestrutura). O Frontend cobre
três jornadas distintas:

1. **Portão público** (não autenticado): Login, Cadastro (Register).
2. **Dashboard operacional** (autenticado): Dashboard/Home, Clientes → Sistemas do Cliente →
   abas (Telas/Regras de Negócio/Versão), Configurações, Cadastro/Perfil, Ajuda, Monitor de Recursos.
3. **Construtor visual** (editor de telas dentro da aba "Telas"): Canvas, Inspector, Painel de
   Componentes/Layers — um editor tipo Figma/Webflow.

## Público-Alvo

Reafirma o levantamento de `001-construtor-sistemas-mach-v4/ui/docs/01-contexto.md`: donos/parceiros
(semi-técnicos, criam e publicam sistemas) e, para a tela Monitor, perfil mais técnico (operação da
própria plataforma). Login/Register atendem quem ainda não tem conta — a página de portão pode (e
deve) ter uma identidade visual própria, mais "marketing", diferente do dashboard interno.

## Inventário de Telas e Cobertura de UI

| Tela / Componente | Arquivo | Padrão de UI | Já auditado? |
|---|---|---|---|
| Login | `auth/Login.tsx` | Split-screen (zinc-900 + formulário) | ✓ `validacao-player-ui/ui/wireframes/login.html` |
| Cadastro | `auth/Register.tsx` | Mesmo split-screen do Login (4 campos a mais) | Espelha o Login, não repetido aqui |
| Dashboard (Home) | `pages/Dashboard/Dashboard.tsx` | Hero + métricas + FAB | ✓ `001-construtor-sistemas-mach-v4/ui/wireframes/dashboard.html` |
| Clientes (lista) | `pages/Dashboard/Clientes.tsx` | Grid de cards + formulário de criação | Parcial — grid de card já coberto por outras telas; formulário não |
| Sistemas do Cliente | `pages/Dashboard/ClienteSistemas.tsx` | Formulário de edição + **tabela de dados** | **Novo nesta auditoria** — único uso de `<table>` no dashboard |
| Configurações | `pages/Dashboard/Configuracao.tsx` + `SegurancaForm`/`WhiteLabelForm` | Cards empilhados com formulários (senha, MFA, exclusão de conta, white label) | **Novo nesta auditoria** |
| Cadastro/Perfil | `pages/Dashboard/Perfil.tsx` | Formulário de conta (nome, foto, e-mail) | **Novo nesta auditoria** (mesmo padrão de formulário de Configurações) |
| Ajuda | `pages/Dashboard/Ajuda.tsx` | Busca + lista agrupada por categoria | Baixa complexidade, padrão já coberto por outros grids de card |
| Monitor de Recursos | `pages/Dashboard/Monitor.tsx` | Cards de status | ✓ `008-monitor-recursos/ui/` (auditoria anterior) |
| Seletor de Sistemas | `systems/SeletorSistemas.tsx` | Grid de cards + criação | Mesmo padrão de Clientes.tsx |
| Construtor (abas Telas/Regras/Versão) | `pages/Dashboard/abas/*`, `pages/Dashboard/editor/*` | Editor visual complexo | ✓ `001-construtor-sistemas-mach-v4/ui/wireframes/builder.html` |
| Player/tela publicada | (serviço `player`) | Renderização headless | ✓ `validacao-player-ui/ui/wireframes/tela-dinamica.html` + `estados.html` |

## Achados de Consistência (só visíveis em varredura de projeto inteiro)

Uma leitura tela a tela não pega isto — só apareceu ao comparar o código de 12+ arquivos lado a lado:

1. **Cor de status fora do sistema de tokens, em 7 arquivos diferentes.** `text-emerald-600
   dark:text-emerald-400` aparece hardcoded em `Perfil.tsx:94`, `ClienteSistemas.tsx:121`,
   `abas/AbaVersao.tsx:84`, `SegurancaForm.tsx:127`, `WhiteLabelForm.tsx:89`; `bg-amber-500/15
   text-amber-600` em `CardFeedback.tsx:28-29`; `bg-green-500`/`bg-red-500` em
   `CardServicoStatus.tsx:44` (já sinalizado na auditoria `008-monitor-recursos`). Ou seja: **todo
   mensagem de sucesso do app usa a mesma cor emerald, sempre com o mesmo par claro/escuro — só
   nunca foi promovida a token** `--success` em `index.css`. Isso não é uma preferência de uma tela,
   é um padrão real e consistente do projeto que falta formalizar (ver `04-sistema-cores-tipografia.md`).
2. **`Dashboard.tsx` é a única tela do dashboard autenticado com texto em inglês.** "Build your Next
   Flow", "Start creating projects and designing your business architecture with our intuitive
   node-based editor", "Get Started", "Create" — enquanto Clientes, Configurações, Perfil, Ajuda,
   Monitor, ClienteSistemas são 100% PT-BR. Ver `03-principios-aplicados.md` (Observar Convenções).
3. **Dois tamanhos de input para o mesmo tipo de campo de texto**, sem motivo funcional aparente:
   formulários "de destaque" (criar cliente em `Clientes.tsx`, editar nome em `ClienteSistemas.tsx`)
   usam `px-4 py-3 rounded-xl`; formulários de conta (`Perfil.tsx`, `SegurancaForm.tsx`) usam
   `px-3 py-2 rounded-lg`. Mesmo componente semântico (campo de texto simples), duas dimensões.
4. **Login e Register compartilham a mesma identidade visual "portão"** (split-screen zinc-900 +
   formulário), consistente entre si — não é uma inconsistência, é uma decisão de design deliberada
   e correta (jornada pública ≠ jornada autenticada). Registrado aqui só para não ser confundido com
   os achados 1–3 numa leitura futura.

## Referências Visuais Encontradas

| Referência | Popularidade | Por que é relevante |
|---|---|---|
| SaaSUI (Notion/Linear/Figma/Stripe screens) | Biblioteca de referência dedicada a padrões de SaaS reais (login, settings, dashboards) | Confirma que "settings page com cards empilhados por seção" (já usado em `Configuracao.tsx`) é o padrão vigente em SaaS B2B de referência — não precisa reinventar, só refinar consistência. |
| shadcn/ui — data table (`tasks` example) | Base de um dos design systems mais adotados no ecossistema React/Tailwind atual (a própria stack do projeto é shadcn-like) | É a referência de tabela mais próxima da stack já usada (`components/ui/*` já segue convenções shadcn) — aplicável diretamente à tabela de sistemas em `ClienteSistemas.tsx`. |
| Stripe Dashboard (billing/settings) | Referência de mercado recorrente em comparações de SaaS B2B | Confirma o padrão de mensagem de sucesso inline junto ao botão de ação (usado em Perfil/SegurancaForm) em vez de toast — adequado para formulários de configuração de baixa frequência. |
