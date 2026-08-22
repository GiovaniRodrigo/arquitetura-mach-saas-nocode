# Princípios Aplicados — Assistente de IA (Chat RAG)

## 1. Início Óbvio
Um único botão flutuante (FAB) circular, ícone de "sparkles", fixo no canto
inferior direito, em `z-50`, visível em **todas** as páginas do Dashboard
(incluído o editor Canvas de viewport fixo). É o único ponto de entrada — sem
ambiguidade sobre onde "falar com a IA".

## 2. Reversão Clara
O painel é um `Sheet` (drawer) não-modal-bloqueante: `Esc`, clique fora ou o
próprio FAB (agora em estado "fechar") o fecham a qualquer momento sem
descartar o histórico da conversa (mantido em memória/sessionStorage). Nenhuma
ação da IA aplica algo automaticamente no Canvas — é só recomendação em texto.

## 3. Lógica Consistente
O `Sheet` reaproveita o componente `components/ui/sheet.tsx` já usado no
projeto — mesmo comportamento de overlay/animação de qualquer outro drawer do
sistema. O FAB usa o mesmo `Button` (`components/ui/button.tsx`) e paleta
`primary` do restante da UI.

## 4. Observar Convenções
Ícone de "sparkles" (lucide-react, já é a lib de ícones do projeto) — semântica
universal para "IA/assistente" (Copilot, Notion AI, Gemini usam o mesmo
glifo). Painel à direita, como Copilot Chat e a maioria dos assistentes
docked.

## 5. Feedback e Marcos
- Estado de "digitando"/streaming da resposta (skeleton de linhas, não spinner
  genérico).
- Erro de rede vira mensagem inline no chat ("Não consegui responder agora,
  tente de novo"), nunca um alert bloqueante.
- Mensagem vazia inicial mostra 3 chips de sugestão (ex.: "Modelar
  multi-tenancy", "Revisar regras de negócio do sistema atual", "Sugerir
  estrutura de telas").

## 6. Proximidade e Adaptação
O painel herda automaticamente o **sistema atualmente selecionado** (nome) via
`AppContext`, exibido como uma "pill" de contexto no topo do chat — o usuário
não precisa redigitar em qual sistema está trabalhando. Em telas sem sistema
selecionado, a pill não aparece e o assistente responde em modo genérico.
Responsivo: em telas estreitas o `Sheet` ocupa a largura total.

## 7. Interface é Conteúdo
Sem decoração supérflua: cabeçalho do painel só com título "Assistente de
Design" + pill de contexto + botão fechar. Mensagens em bolhas simples
(usuário à direita, assistente à esquerda), sem avatares desnecessários.

## 8. Princípios Gerais de Design Visual
- **Assunto óbvio**: título "Assistente de Design" + ícone sparkles no topo do
  painel.
- **Forma reforça significado**: bolhas do assistente usam `bg-secondary`
  (neutra), bolhas do usuário usam `bg-primary/10` — sem cores de alerta, já
  que não há estado de erro/sucesso no conteúdo em si.
- **Metáfora familiar**: layout de chat 1:1 (mensagens empilhadas + composer
  fixo na base) — nenhum usuário precisa aprender um padrão novo.

## 9. Matriz de Decisão de Design

| Decisão | Início Óbvio | Reversão Clara | Consistência | Convenção | Feedback | Proximidade | Conteúdo > Deco. |
|---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| FAB global fixo (canto inferior direito) | ✓ | ✓ | ✓ | ✓ | — | — | ✓ |
| Sheet docked à direita (não modal) | ✓ | ✓ | ✓ | ✓ | ✓ | — | ✓ |
| Composer fixo na base do painel | ✓ | — | ✓ | ✓ | ✓ | ✓ | ✓ |
| Pill de contexto (sistema atual) | — | — | ✓ | — | ✓ | ✓ | ✓ |
| Chips de sugestão em estado vazio | ✓ | — | — | ✓ | ✓ | — | ✓ |
| Skeleton de streaming em vez de spinner | — | — | ✓ | ✓ | ✓ | — | ✓ |
