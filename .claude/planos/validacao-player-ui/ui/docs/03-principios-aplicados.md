# Princípios de Design Aplicados às Telas do Player

Avaliação das telas reais do Headless Player (`Login.tsx`, estados de `App.tsx`,
`CompositeRenderer.tsx`) contra 9 princípios. Cada item traz **estado atual → gap → ação**.

## Telas avaliadas
- **T1 — Login** (sem sessão): título "Plataforma MACH", subtítulo, 2 links sociais em texto.
- **T2 — Loading**: `<div>Carregando…</div>`.
- **T3 — Empty / sem sistema**: "Autenticado ✓. Nenhum sistema selecionado…".
- **T4 — Erro**: `<div role="alert">{erro}</div>` (string crua do erro).
- **T5 — Tela dinâmica**: `nav` de links + formulário renderizado + botão (sem estilo próprio).

---

### 1. Início Óbvio
- **T1**: há um ponto de partida (2 botões), mas **sem CTA primário destacado** — Google e
  GitHub têm peso visual idêntico. Ação: promover 1 primário ("Continuar com Google" cheio,
  colorido) e GitHub como secundário (outline).
- **T5**: `nav` de links sem indicação de item ativo nem do "primeiro passo". Ação: destacar
  rota ativa e o CTA principal do formulário.

### 2. Reversão Clara
- **T4**: erro é um beco sem saída — texto cru, **sem botão "Tentar novamente"** nem "Voltar
  ao login". Ação: adicionar retry e link de saída.
- **T5**: submissão do formulário sem "Cancelar"/limpar visível. Ação: par Cancelar/Enviar.

### 3. Lógica Consistente
- **Global**: só o Login tem estilo; T2–T5 saem com aparência default do browser →
  **inconsistência visual severa** entre telas do mesmo produto. Ação: sistema de design único
  (tokens de cor/tipografia/spacing) aplicado a todas as telas e ao CompositeRenderer.
- Estados de `:hover`/`:focus` não definidos nos links de login. Ação: padronizar.

### 4. Observar Convenções
- **T1**: botões sociais **sem o logo dos IDPs** — desvio das guidelines do Google (G de 4
  cores) e do GitHub (mark oficial). Ação: incluir os logos oficiais.
- Ícones semânticos ausentes em T4/T5 (erro sem ícone de alerta). Ação: usar ícones universais.

### 5. Feedback e Marcos
- **T2**: "Carregando…" é texto genérico. Ação: **skeleton screen** com a silhueta da tela.
- **T5**: submissão sem estado de loading/sucesso claro além de `status` textual. Ação: toast de
  sucesso + estado de botão "enviando…".
- Erros de validação (`data-bi`) aparecem como `<p role="alert">` sem âncora visual ao campo.
  Ação: mensagem inline sob o campo, com cor semântica.

### 6. Proximidade e Adaptação
- **T1**: `max-width:360px` centrado é ok, mas **não há media queries** nem alvos de toque
  garantidos (padding 12px → alvo < 44px em alguns casos). Ação: mobile-first, alvos ≥44px.
- **T5**: erros de validação ficam agrupados **depois** do formulário, longe dos campos. Ação:
  aproximar mensagem do campo que ela afeta.

### 7. Interface é Conteúdo
- **T1**: enxuta (bom), mas a economia vira **falta de identidade** (sem logo/marca). Ação:
  adicionar só o essencial de marca (logo + nome), sem decoração supérflua.
- **T3**: empty state instrui via `<code>?sistema=…</code>` — jargão técnico exposto ao
  usuário final. Ação: transformar em ação guiada ("Selecionar sistema") em vez de query param.

### 8. Princípios Gerais de Design Visual
- **Assunto óbvio**: falta logo/identidade MACH no topo do Login e um título de contexto nas
  telas dinâmicas. Ação: cabeçalho com marca.
- **Dados adequados**: T5 renderiza formulários; ok. Onde houver listas/CRUD, usar tabela.
- **Forma e conteúdo**: erro sem cor semântica (vermelho) e sucesso sem verde. Ação: aplicar
  paleta semântica (doc 04).
- **Metáforas**: "sistema" e "versão ativa" são conceitos novos ao leigo — apoiar com microcopy.

### 9. Matriz de Decisão de Design (estado atual)

| Decisão / Tela | Início Óbvio | Reversão Clara | Consistência | Convenção | Feedback | Proximidade | Conteúdo > Deco. |
|---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| T1 — Login (botões sociais) | ✗ | ✓ | ✗ | ✗ | ✗ | ✗ | ✓ |
| T2 — Loading | — | — | ✗ | ✗ | ✗ | — | ✓ |
| T3 — Empty / sem sistema | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ |
| T4 — Erro | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✓ |
| T5 — Tela dinâmica (form) | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✓ |

Legenda: ✓ atende · ✗ não atende · — não aplicável.

## Prioridades (maior impacto primeiro)
1. **Sistema de design único** aplicado a todas as telas (resolve Consistência em T2–T5).
2. **Redesenho do Login**: card + logo, CTA primário, logos dos IDPs, foco/hover, responsivo.
3. **Estados de verdade**: skeleton (T2), empty com ação (T3), erro com retry (T4).
4. **Feedback de formulário**: validação inline + toast de sucesso + estado de envio (T5).
