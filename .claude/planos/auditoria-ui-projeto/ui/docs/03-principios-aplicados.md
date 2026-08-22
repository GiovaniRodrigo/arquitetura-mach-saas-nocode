# Princípios Aplicados — Auditoria de Escopo Global

## 1. Início Óbvio

Cada tela do dashboard já segue o mesmo padrão: `TonalCard` de cabeçalho com título + subtítulo,
seguido da ação/dado principal (Clientes → lista; Configurações → cards de seção; Perfil → dados
de conta). Consistente em todas as 7 telas lidas — nenhuma mudança necessária aqui.

## 2. Reversão Clara

- Ações destrutivas (excluir cliente, excluir conta) já pedem confirmação por senha
  (`SegurancaForm.tsx`) ou estão isoladas com `title` explicando a irreversibilidade
  (`ClienteSistemas.tsx:142`). Padrão correto, mantido.
- **Gap real**: nenhuma das duas usa um diálogo de confirmação (`components/ui/dialog.tsx` já
  existe no projeto, mas não é usado nesses fluxos) — hoje "Excluir cliente"/"Excluir conta" agem
  no primeiro clique. Reautenticação por senha (exclusão de conta) mitiga cliques acidentais, mas
  "Excluir cliente" em `ClienteSistemas.tsx:140` não pede senha nem confirmação — só um `title`
  de tooltip, que passa despercebido. Recomenda-se reusar `dialog.tsx` aqui.

## 3. Lógica Consistente

Achado principal desta auditoria (detalhado em `01-contexto.md`):

- **Cor de sucesso**: `text-emerald-600 dark:text-emerald-400` repetida literalmente em 5 arquivos
  (`Perfil.tsx:94`, `ClienteSistemas.tsx:121`, `abas/AbaVersao.tsx:84`, `SegurancaForm.tsx:127`,
  `WhiteLabelForm.tsx:89`). É consistente *entre si* (mesmo valor sempre) — o problema não é
  divergência, é que não está no sistema de tokens (`index.css`), então não pode ser ajustada num
  só lugar nem geriada por tema como `--destructive` já é. Resolvido no doc `04`.
- **Cor de alerta/pendência**: `bg-amber-500/15 text-amber-600 dark:text-amber-400` em
  `CardFeedback.tsx:28-29` — mesmo caso, único uso de âmbar no projeto, também sem token.
- **Tamanho de input**: dois padrões coexistem para o mesmo componente semântico (campo de texto
  simples) — `px-4 py-3 rounded-xl` (Clientes/ClienteSistemas, formulários "primários" de uma
  tela) vs `px-3 py-2 rounded-lg` (Perfil/SegurancaForm, formulários de conta). Não há um
  componente `<Input>` compartilhado sendo reusado (existe `components/ui/input.tsx`, mas nenhuma
  dessas 4 telas o importa — todas reescrevem a `className` do `<input>` na mão). Recomendação:
  migrar os 4 formulários para `components/ui/input.tsx`, unificando o tamanho por decisão
  consciente (não por acidente de cópia-e-cola).

## 4. Observar Convenções

- **Idioma**: `Dashboard.tsx` (a primeira tela que qualquer usuário vê após login) está em inglês
  ("Build your Next Flow", "Get Started", "Create") enquanto todo o resto do dashboard é PT-BR.
  Isso quebra a convenção mais básica de um produto: idioma único e prometido ao usuário. É a
  correção de mais alto impacto desta auditoria — visível a 100% dos usuários, na primeira tela.
- Ícones (`lucide-react`), tipografia (Outfit/Inter/JetBrains Mono) e componentes `m3/`/`ui/` são
  reusados de forma consistente em todas as 12 telas lidas — nenhum outro desvio de convenção
  encontrado além dos já listados.

## 5. Feedback e Marcos

- Padrão de sucesso inline (`role="status"`, texto verde junto ao botão) é usado de forma
  consistente em Perfil, ClienteSistemas, SegurancaForm, WhiteLabelForm, AbaVersao — bom padrão,
  só falta o token (item 3).
- Padrão de erro (`role="alert"`, `text-destructive`) também consistente em todas as telas com
  formulário — já usa o token correto, ao contrário do sucesso. Nenhuma mudança necessária.
- `Skeleton`/`EmptyState`/`ErrorState` de `StateViews.tsx` são reusados por Clientes,
  ClienteSistemas e SeletorSistemas — o único lugar que ainda não reusava era `Monitor.tsx`
  (já endereçado na auditoria `008-monitor-recursos`).

## 6. Proximidade e Adaptação

Grids responsivos (`grid-cols-1 md:grid-cols-2`/`lg:grid-cols-3`) e `max-w-5xl mx-auto`/`max-w-2xl
mx-auto` são usados de forma consistente para limitar a largura de leitura em telas largas — sem
achados novos aqui.

## 7. Interface é Conteúdo

Login/Register dedicam metade da tela (`md:w-1/2`) a um painel puramente de marca — desperdício de
espaço em telas menores que `md`? Não: em mobile o painel de marca desaparece
(`flex-col md:flex-row`, o painel unicamente decorativo não tem `md:hidden` mas fica empilhado
acima do formulário) — comportamento aceitável para uma tela de portão (reforço de marca antes da
ação é uma escolha válida aqui, diferente do dashboard interno onde isso seria desperdício).

## 8. Princípios Gerais de Design Visual

- **Forma e conteúdo integrados**: uma vez que `--success`/`--warning` existam como tokens (ver
  doc `04`), toda mensagem de sucesso do app herda automaticamente qualquer ajuste de tema futuro
  — hoje um ajuste de contraste exigiria editar 5 arquivos manualmente.
- **Visualização de dados adequada**: a única tabela do dashboard (`ClienteSistemas.tsx`) usa
  texto simples para "Nome" — adequado, não é um dado que peça outra representação.

## 9. Matriz de Decisão de Design

| Decisão | Início Óbvio | Reversão Clara | Consistência | Convenção | Feedback | Proximidade | Conteúdo > Deco. |
|---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| Promover `--success`/`--warning` a tokens em `index.css` | — | — | ✓ | ✓ | ✓ | — | ✓ |
| Traduzir `Dashboard.tsx` para PT-BR | — | — | ✓ | ✓ | — | — | — |
| Unificar tamanho de input via `components/ui/input.tsx` | — | — | ✓ | ✓ | — | — | — |
| Confirmação (dialog) antes de "Excluir cliente" | — | ✓ | ✓ | ✓ | ✓ | — | — |
| Manter identidade visual própria de Login/Register (zinc split-screen) | ✓ | — | ✓ (interna) | ✓ | — | — | ✓ |
