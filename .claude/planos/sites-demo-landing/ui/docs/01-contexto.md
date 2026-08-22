# Contexto do Projeto

## Domínio

Este `/ui` cobre um alvo diferente das auditorias anteriores (`001-construtor-sistemas-mach-v4`,
`008-monitor-recursos`, `auditoria-ui-projeto`): não é o **dashboard interno** do MACH V4 (a
ferramenta que donos/parceiros usam para operar a plataforma), e sim o **produto final que o
Design Engine no-code gera** — o site publicado que o cliente do dono da conta vê.

Especificamente: a tela **"Home"** do sistema de demonstração "Loja Demo", uma landing page SaaS
B2B completa ("Brillance" — inspirada no template "Brillance SaaS Landing Page" do v0.app),
montada inteiramente com os componentes do catálogo do Design Engine
(`services/frontend/src/systems/componentRegistry.ts`) via o script
`build/seed-demo-site.sh` (707 linhas, gera 3 telas: Home, Produtos, Contato — o foco aqui é Home,
a mais elaborada e a que serve de vitrine do que o construtor no-code é capaz de produzir).

Estrutura atual da Home (linhas do script entre parênteses): navbar (170-189), hero (422-447),
logo cloud (448-465), grid de 6 feature cards em 2 linhas (466-489), dois showcases
texto/imagem alternados (264-299, instanciados em 396-397), faixa de estatísticas em 4 blocos
(301-311, 492-496), 3 cards de depoimento (313-339, 498-508), 3 cards de plano/pricing
(341-385, 509-526), FAQ em accordion (528-541), CTA final (542-551) e footer de 5 colunas
(206-248).

## Público-Alvo

Dois públicos distintos e por isso duas leituras diferentes desta auditoria:

1. **Quem edita** (dono/parceiro do MACH, semi-técnico): usa o Canvas para montar telas como
   esta arrastando os mesmos componentes — então toda recomendação de melhoria precisa ser
   **alcançável com o catálogo de componentes atual** (ver `03-principios-aplicados.md` para o
   levantamento exato do que o motor de renderização suporta), ou vir marcada explicitamente como
   "requer novo campo no registry".
2. **Quem visita o site publicado** (cliente final do dono/parceiro — aqui, o público-alvo
   fictício da "Brillance" é um comprador B2B avaliando ferramentas de gestão de equipe). É este
   público que a pesquisa de tendências em `02-referencias.md` mira.

## Referências Visuais Encontradas

| Referência | URL | Relevância |
|---|---|---|
| Notion / Linear / Framer (hero storytelling) | citado em SaaSFrame Blog, "10 SaaS Landing Page Trends for 2026" | Hero sections que mostram o produto e uma narrativa de "antes → depois" em vez de só declarar o que o produto é — a Home já segue parcialmente isso (headline + imagem do painel), dá pra reforçar. |
| Dribbble — tag `saas-hero` / `hero-section` | dribbble.com/tags/saas-hero | Milhares de designs — usada para validar o padrão de "eyebrow badge + heading + CTA duplo + prova social flutuante" que a Home já usa (`hero-eyebrow`, `hero-badge-flutuante`). |
| Dribbble — tag `pricing-page` / `pricing-card` | dribbble.com/tags/pricing-page (1.800+ designs) | Confirma o padrão "Mais Popular" com borda/sombra destacada, já implementado em `pricing_card()` — validado, não muda. |
| Nielsen Norman Group (via SaaS Website Best Practices, Lovable/Brand Vision) | ver `02-referencias.md` | Prova social: testemunhos com foto + nome completo são avaliados como mais confiáveis que iniciais anônimas; um badge de rating agregado ("4.8/5, 3.200 avaliações") é mais crível que uma citação isolada. |
| Figma — páginas de cor Indigo/Violet | figma.com/colors/indigo, figma.com/colors/violet | Indigo (a cor primária atual, `#4f46e5` = Tailwind indigo-600) é descrita como o meio-termo preferido em SaaS premium — mais personalidade que azul, mais acessível que violeta. Paleta atual **validada**, não é datada. |
| SaaS Landing Page — ranking de fontes | saaslandingpage.com/articles/the-12-most-popular-google-fonts-for-landing-pages | Inter aparece em 182 sites SaaS analisados — a fonte nº1 do mercado. O projeto já usa Inter como fonte padrão (`tailwind.config.js`). **Validado.** |

## Tendências Identificadas

1. **Tipografia com duas vozes** (heading em fonte "display" mais expressiva, corpo em fonte
   neutra) — em vez de uma família só para tudo.
2. **Prova social com rosto** — foto real (ou placeholder de foto) em vez de iniciais em avatar,
   e um número agregado de credibilidade (rating/contagem) perto da seção de depoimentos.
3. **Profundidade em camadas** (múltiplas sombras — ambiente + contato — em vez de uma sombra só)
   para dar sensação de elevação sem precisar de gradiente.
4. **Raio de borda um pouco maior** ("soft UI") em cards e botões de destaque, tendência 2025 em
   relação ao 8-12px mais "corporativo" de anos anteriores.
5. **Gradientes e overlays de cor** — tendência forte no mercado, mas **não suportada hoje** pelo
   modelo de estilos do Design Engine (`fundoCor` mapeia só para `background-color` sólido, nunca
   para `background: linear-gradient(...)`) — registrado como oportunidade de melhoria de
   plataforma, não como recomendação de conteúdo (ver `03-principios-aplicados.md`).
