# Contexto do Projeto

## Domínio

MAYS (Make Your SaaS) é um construtor no-code de sistemas SaaS multi-tenant: o
usuário monta telas por composição visual (Canvas tipo Figma/Webflow em
`pages/Dashboard/editor`), define regras de negócio, versiona e publica o
sistema resultante. É uma ferramenta de **arquitetura/design de sistemas**
operada por não-programadores ou programadores low-code.

Demanda desta análise: adicionar um **assistente de IA especialista em
design/arquitetura de sistemas**, usando RAG (Retrieval-Augmented Generation),
que conversa com o usuário sobre o foco/descrição do que ele está construindo e
devolve recomendações de modelagem, estrutura e boas práticas.

## Público-Alvo

Donos de conta (donos/parceiros) construindo o sistema de um cliente, dentro do
Dashboard autenticado. Nível técnico variável — de leigo a desenvolvedor — mas
todos já operam conceitos de "sistema", "tela", "regra de negócio", "versão".
O assistente precisa ser útil tanto para quem não sabe nomear o problema
("quero um sistema para agendar consultas") quanto para quem já pensa em termos
técnicos ("como estruturo multi-tenancy aqui?").

## Referências Visuais Encontradas

Ver `02-referencias.md` para tabela completa com URLs e popularidade.

- **GitHub Copilot Chat** — painel lateral (sidebar) docked, não flutuante
  sobre o conteúdo; composer fixo na base do painel.
- **Notion AI** — painel de IA com entrada centralizada e chips de sugestão;
  histórico só aparece após a primeira mensagem (assistente escopado à tarefa).
- **Intercom Messenger** — padrão de widget flutuante ancorado a um canto,
  otimizado para primeira resposta rápida (suporte ao cliente, não é o
  padrão certo para um assistente de trabalho contínuo).
- **Attio / Hex** (dashboards líderes 2026) — tratam a saída de IA como
  superfície de primeira classe (resumos, ações sugeridas) em vez de widget
  flutuante por cima da UI antiga.

## Tendências Identificadas

1. **Painel docked, não pop-up flutuante sobre o conteúdo** — o padrão
   vencedor em produtos de produtividade (Copilot Chat, Notion AI, Linear AI)
   é abrir/fechar um painel lateral que empurra ou sobrepõe parcialmente o
   layout, mantendo o composer sempre ancorado na base (evita o bug de UX mais
   comum em chat de IA: composer flutuante sobrepondo a última mensagem).
2. **Gatilho persistente e global**, não escondido dentro de uma única aba —
   já que o "foco do sistema" atravessa várias telas do builder (Sistemas,
   Regras de Negócio, Telas), o ponto de entrada do assistente deve existir em
   todo o Dashboard, não só dentro do editor Canvas.
3. **Contexto automático da tela atual** — assistentes de IA em produtos 2026
   (Notion AI, Linear AI) herdam o contexto do que o usuário está vendo (aqui:
   sistema selecionado) em vez de exigir que o usuário reexplique tudo.
4. **Estado escopado por tarefa, com sugestões de entrada** — evita tela de
   chat vazia; mostra chips/sugestões iniciais ("Modelar multi-tenancy",
   "Revisar regras de negócio") alinhadas ao domínio do produto.
5. **IA como camada complementar, nunca bloqueante** — o painel pode ser
   fechado a qualquer momento sem perder o trabalho no Canvas por trás; nunca
   é modal.
