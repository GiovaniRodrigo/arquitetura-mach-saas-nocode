# Princípios de Design Aplicados

Aplicação dos 9 princípios às três telas-alvo: **Dashboard**, **Construtor Visual (Builder)** e **Headless Player**. Cada princípio traz decisões concretas ancoradas nos RF/RN da spec.

---

## 1. Início Óbvio

- **Dashboard**: o ponto de partida é o botão **"+ Novo Sistema"** (CTA primário, acento único, canto superior direito) e, abaixo, o card do último sistema editado em destaque. O olho vai do título → busca `Cmd+K` → grelha de sistemas.
- **Builder**: o início é o **canvas central** — a maior superfície, sempre iluminada; a biblioteca de componentes (esquerda) convida ao primeiro arrasto com um *empty state* "Arraste um componente para começar".
- **Player**: o início é o **primeiro campo do formulário já em foco** (autofocus) com o CTA "Enviar" fixo ao fim; em multi-etapa, o passo 1 aberto e o indicador "Passo 1 de N" no topo.

## 2. Reversão Clara

- **Rollback (RF04/RN05)** é a reversão-mãe do produto: acessível em 1 clique no painel de versões, com **diálogo de confirmação** ("Reverter para a versão 6? O sistema publicado passa a servir a v6 imediatamente."), mostrando o que muda. Executa em < 100 ms (critério de aceitação 3) e confirma via toast.
- **Ações destrutivas** (excluir sistema, excluir componente, excluir regra) exigem confirmação explícita; o botão destrutivo é vermelho e nunca é o foco padrão do diálogo.
- **Builder**: toda mutação é reversível via **Undo/Redo** (`Cmd+Z` / `Cmd+Shift+Z`); como a persistência é write-behind com debounce de 5 s (RN06), o undo local é instantâneo antes do flush.
- **Formulário (Player)**: "Voltar" sempre visível em fluxos multi-etapa; nenhum dado é perdido ao voltar.

## 3. Lógica Consistente

- Um mesmo tipo de componente (ex.: `input_texto`) renderiza igual no canvas do Builder e no Player — **paridade de renderização** garantida pelo mesmo contrato de árvore Composite (RF01).
- Estados de interação uniformes em todo o produto: `hover` (elevação sutil + borda de acento), `focus` (anel de foco de 2 px acessível), `disabled` (opacidade 0.4, cursor `not-allowed`), `selected` (borda de acento sólida).
- Posições fixas: barra superior (contexto do tenant + ações globais) sempre no topo; ações de linha de tabela sempre à direita; breadcrumb sempre abaixo da barra superior.

## 4. Observar Convenções

- **Ícones semânticos universais**: 🗑 lixeira = excluir, ✏️ lápis = editar, ＋ = adicionar, ⎘ = duplicar, ⤺ = reverter/rollback, ↑ = publicar, ⭳ = exportar. (Set: Lucide — padrão de dev-tools.)
- **Material 3 / HIG** para campos de formulário, chips, switches e diálogos.
- **Terminologia do domínio** que o utilizador já conhece: "Sistema", "Versão", "Publicar", "Rollback", "Colaboradores", "Regras" — exatamente os termos da spec, sem jargão de infra (o utilizador nunca vê "gRPC", "GenServer" ou "blind_index" cru; este último aparece apenas em ferramentas de admin/debug como *badge* mono).
- **Command palette `Cmd+K`** (convenção Linear/Figma) para navegação e ações rápidas.

## 5. Feedback e Marcos

- **Estados explícitos** em toda superfície: `loading` (skeleton screens, não spinners), `success` (toast verde não intrusivo), `error` (toast/campo vermelho com mensagem acionável), `empty` (ilustração + CTA).
- **Colaboração (RF06)**: presença sempre visível (avatares empilhados no topo), cursores nomeados coloridos no canvas, badge "🔒 em edição por Ana" no componente bloqueado (RN07), e indicador de estado de persistência: **"Todas as alterações salvas"** ↔ **"Salvando…"** (reflete o debounce de 5 s / `flush_ok`, RN06).
- **Publicação (RF04)**: barra de progresso do publish; ao concluir, marco visível "Versão 7 publicada · ativa agora".
- **Exportação (RF05)**: como é assíncrona (202 imediato), mostra job com estado `criado → coletando → pronto`; quando pronto, botão de download com aviso de expiração ("link expira em 10 min").
- **Player**: validação inline no *blur* (NN/g); ao submeter, spinner no botão → confirmação; erros do servidor (422) mapeados ao campo exato via `blind_index` (RN08).

## 6. Proximidade e Adaptação

- **Agrupamento**: no painel de propriedades do Builder, campos relacionados ficam em seções colapsáveis (Layout · Estilo · Dados · Regras) com espaçamento interno menor que o entre-seções.
- **Ações contextuais próximas aos dados**: toolbar flutuante aparece sobre o componente selecionado no canvas (duplicar/excluir/bloquear), não numa barra distante.
- **Responsividade**:
  - Dashboard: grelha fluida (4 → 2 → 1 colunas).
  - Builder: **desktop-only** honestamente assumido (canvas exige espaço); em telas menores, aviso "Abra num ecrã maior para editar".
  - Player: **mobile-first**, um campo por linha, alvos de toque ≥ 44 px.

## 7. Interface é Conteúdo

- **Máximo de dados úteis sem scroll**: o canvas do Builder ocupa a maior área; barras laterais colapsáveis; header de 48 px de altura.
- **Zero decoração vazia**: sem gradientes ornamentais, sombras pesadas ou ilustrações que não comuniquem estado. A cor de acento é reservada — quando tudo é colorido, nada tem prioridade (contenção Linear).
- **Densidade calibrada por público**: densa no Builder/Dashboard (técnico); arejada no Player (leigo).

## 8. Princípios Gerais de Design Visual

- **Torne o assunto óbvio**: cada tela abre com título + ícone de contexto (nome do sistema no Builder; nome do formulário no Player; "Meus Sistemas" no Dashboard).
- **Visualização de dados adequada**: histórico de versões como **linha do tempo vertical** (sequência temporal); estados de job de exportação como **stepper**; profundidade da fila/uso em cards numéricos — não gráficos decorativos.
- **Forma e conteúdo integrados**: semântica de cor consistente — verde = sucesso/publicado, âmbar = rascunho/pendente, vermelho = erro/destrutivo, azul-info = neutro informativo, acento = ação primária/ativo.
- **Metáforas para conceitos novos**: **árvore Composite** representada como uma *tree view* de arquivos (metáfora familiar); **Blind Index** como um cadeado/hash (privacidade); **versão ativa** como um interruptor de "no ar".

## 9. Matriz de Decisão de Design

| Decisão | Início Óbvio | Reversão Clara | Consistência | Convenção | Feedback | Proximidade | Conteúdo > Deco. |
|---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| CTA "+ Novo Sistema" (Dashboard) | ✓ | ✗ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Card de Sistema c/ status de versão | ✓ | ✗ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Layout 3-colunas do Builder | ✓ | ✗ | ✓ | ✓ | ✗ | ✓ | ✓ |
| Cursores/presença de colaboração (RF06) | ✗ | ✗ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Badge "🔒 em edição" (RN07) | ✗ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Indicador "Salvando…/Salvo" (RN06) | ✗ | ✓ | ✓ | ✓ | ✓ | ✗ | ✓ |
| Painel de versões + Rollback (RF04) | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Diálogo de confirmação de Rollback | ✗ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Formulário do Player + validação inline (RF07) | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Erro por campo via blind_index (RN08) | ✗ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Job de Exportação (stepper + expiração) | ✓ | ✗ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Command palette `Cmd+K` | ✓ | ✗ | ✓ | ✓ | ✓ | ✗ | ✓ |

Legenda: ✓ = princípio aplicado nesta decisão; ✗ = não é o foco desta decisão (aplicado por outra).
