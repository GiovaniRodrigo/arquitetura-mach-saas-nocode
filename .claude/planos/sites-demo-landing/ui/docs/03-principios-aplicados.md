# Princípios Aplicados

> Antes dos princípios: o que o motor de renderização realmente suporta. Toda recomendação abaixo
> foi checada contra `services/frontend/src/systems/estilosCss.ts` (converte `Estilos` → CSS
> inline) e `services/frontend/src/pages/Dashboard/editor/PreviewRenderer.tsx` (renderiza cada
> `tipo` de componente). Mapeamento 1:1 confirmado: `fundoCor`→`background-color` (cor sólida
> **só**, sem gradiente), `sombra`→`box-shadow` (string livre — **múltiplas sombras em camadas
> funcionam**), `bordaRaio`→`border-radius` (aceita shorthand tipo `"16px 16px 0 0"` — raio
> assimétrico funciona), **sem** `transition`/`transform`/`animation` expostos, **sem** CSS Grid
> (`display` só aceita `block|inline|inline-block|flex|none`), **sem** campo de família de fonte
> (todo texto herda a fonte global `Inter` do app, mesmo `Outfit` estando carregado e disponível).
> Isso define o que é recomendação de **conteúdo** (aplicável hoje, só editando
> `seed-demo-site.sh`) vs. recomendação de **plataforma** (exige um campo novo no `componentRegistry`).

## 1. Início Óbvio
O CTA primário do hero (`hero-cta-primario`, "Começar grátis") já tem a maior proeminência visual
da página inteira: cor sólida da marca + sombra colorida (`0 8px 16px rgba(79,70,229,.25)`,
linha 433), enquanto o secundário (`hero-cta-secundario`, "Ver demonstração") é outline. Correto,
sem mudança. Reforço recomendado: aplicar a mesma assinatura de sombra colorida também no CTA
final da página (`cta-final-botao`, linha 548) — hoje ele é só `fundoCor` sólida sem sombra,
perdendo o "peso" visual que o botão equivalente do hero tem, apesar de ser a última chance de
conversão da página.

## 2. Reversão Clara
Não se aplica a uma landing page de marketing — não há ações destrutivas nesta tela (é
diferente do formulário de Contato, já coberto por outra tela). Nenhuma alteração recomendada.

## 3. Lógica Consistente
Achado: os dois showcases (`showcase_section`, linha 264) usam `bordaRaio:"16px"` na imagem
(linha 291) enquanto os cards de feature/depoimento/pricing usam `12-16px` variando por tipo, sem
um padrão único. Recomendação: adotar uma escala fixa de 3 valores (12px pequenos/badges, 16px
cards, 20px blocos grandes como a imagem do hero) documentada em `04-sistema-cores-tipografia.md`,
em vez de valores ad-hoc por seção.

## 4. Observar Convenções
A estrutura já observa a convenção de mercado validada em `02-referencias.md`: header sticky-like
com logo+menu+CTA, hero com eyebrow badge, logo cloud logo abaixo do hero, "Mais Popular" com
destaque visual no pricing. Nenhuma mudança estrutural — só refinamento visual (itens 5-9 abaixo).

## 5. Feedback e Marcos
Não aplicável a conteúdo estático de marketing (loading/error states são do dashboard, já
cobertos em `auditoria-ui-projeto`). O único "estado" real da Home é o accordion do FAQ
(`AccordionPublicado`, já funcional — abre/fecha) e os cards de pricing/testemunho, que são
estáticos por natureza.

## 6. Proximidade e Adaptação
A página é fixa em largura de desktop (`hero-imagem` tem `largura:"860px"` fixo, linha 442; sem
media query no modelo de estilos) — não há como aplicar responsividade real sem uma mudança de
plataforma (grid/breakpoints), fora do escopo de conteúdo. Registrado como risco conhecido, não
como recomendação desta rodada (o Canvas em si já tem preview de device — mobile/tablet/desktop —
mas a árvore publicada usa valores de `largura` absolutos, não fluidos).

## 7. Interface é Conteúdo
Achado principal desta auditoria: os avatares de depoimento usam apenas iniciais
(`testimonial_card`, linha 313, `propriedades:{texto:$iniciais,...}`) — texto genérico fazendo às
vezes de "rosto". Pesquisa (`02-referencias.md`, NN/g via Lovable) mostra que uma foto real
(mesmo placeholder fotográfico) comunica muito mais do que iniciais abstratas. Como o tipo
`avatar` já suporta `src`/`alt` (`PreviewRenderer.tsx:130-138`), a troca é só de conteúdo — ver
wireframe `depoimento-social-proof.html`.

## 8. Princípios Gerais de Design Visual
- **Torne o assunto óbvio**: já atendido (eyebrow badge "✨ Novidade: automação com IA" no topo
  do hero identifica o contexto do produto antes mesmo do heading).
- **Forma e conteúdo integrados**: o badge "MAIS POPULAR" do plano Pro já usa a cor primária para
  reforçar destaque — correto. Recomendação nova: usar a mesma técnica de "camada de sombra dupla"
  (ambiente difusa + contato nítido) em vez de sombra única em todos os elementos elevados
  (hero image, cards de pricing destacado, cards de depoimento) — técnica popular em produtos como
  Linear/Stripe (citada nas referências de tendência 2025) e 100% suportada pelo `sombra` (string
  livre de `box-shadow`, aceita múltiplos valores separados por vírgula).
- **Metáforas para conceitos novos**: não aplicável — produto é gestão de tarefas, conceito já
  familiar ao público-alvo B2B.

## 9. Matriz de Decisão de Design

| Decisão | Início Óbvio | Reversão Clara | Consistência | Convenção | Feedback | Proximidade | Conteúdo > Deco. |
|---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| Sombra em camadas (hero image, pricing destaque, cards) | ✓ | — | ✓ | ✓ | — | ✓ | ✓ |
| Avatares de depoimento com foto em vez de iniciais | ✓ | — | ✓ | ✓ | — | — | ✓ |
| Badge de rating agregado acima dos depoimentos | ✓ | — | ✓ | ✓ | — | ✓ | ✓ |
| Sombra colorida no CTA final (paridade com hero) | ✓ | — | ✓ | ✓ | — | — | — |
| Escala fixa de border-radius (12/16/20px) | — | — | ✓ | ✓ | — | ✓ | — |
| Campo `fonteFamilia` no registry (Outfit em headings) | ✓ | — | ✓ | ✓ | — | — | ✓ |

## Fora de escopo desta rodada (registrado, não implementado)

- **Gradientes de fundo** (tendência forte em 2025, ver `01-contexto.md`) — bloqueado pelo modelo
  de estilos atual (`fundoCor` só aceita cor sólida). Exigiria um campo `fundoGradiente` mapeado
  para `background: linear-gradient(...)` em `estilosCss.ts` — mudança de plataforma, não de
  conteúdo desta demo.
- **Responsividade real** (breakpoints) — o modelo de `Estilos` não tem conceito de valores por
  device; o Canvas simula 3 larguras no editor, mas a árvore publicada é fixa. Mudança de
  plataforma maior, fora do escopo de "melhorar a interface do site demo".
