# Sistema de Cores e Tipografia

> A paleta e a fonte de corpo atuais (definidas em `build/seed-demo-site.sh:39-49` e herdadas do
> `tailwind.config.js` do app) já batem com o que a pesquisa de mercado aponta como correto para
> este domínio — ver `02-referencias.md`. Este documento **não propõe trocar** matiz nem fonte de
> corpo; formaliza a escala derivada (hoje só parcialmente presente, alguns valores ad-hoc) e
> propõe um único ganho de plataforma (fonte de display nos headings).

## Paleta de Cores

### Primária (validada — Tailwind `indigo-600`, referência: Figma "Indigo" color page)
- Principal: `#4f46e5` (já em uso, `COR_PRIMARIA`)
- Escura (hover/estados de ênfase, faixa de estatísticas): `#4338ca` (já em uso, `COR_PRIMARIA_ESCURA`)
- Clara (fundos de badge/ícone): `#eef2ff` (já em uso, `COR_PRIMARIA_CLARA`)

Escala completa recomendada para consistência (só os tons 50/600/700 são usados hoje — os
demais existem no Tailwind indigo e podem ser usados se a paleta precisar de mais variação, ex.
estados de foco em formulário):

| Tom | Hex | Uso sugerido |
|---|---|---|
| 50 | `#eef2ff` | fundo de badge, ícone (já em uso como `COR_PRIMARIA_CLARA`) |
| 100 | `#e0e7ff` | hover de fundo claro (não usado hoje) |
| 500 | `#6366f1` | cor default do `componentRegistry.ts` (botão/badge sem estilo custom) |
| 600 | `#4f46e5` | primária (já em uso) |
| 700 | `#4338ca` | escura/ênfase (já em uso como `COR_PRIMARIA_ESCURA`) |

### Neutros (já em uso, sem mudança)
- Texto: `#111827` (`COR_TEXTO`)
- Texto mudo: `#6b7280` (`COR_TEXTO_MUTED`)
- Borda: `#e5e7eb` (`COR_BORDA`)
- Fundo claro (seções alternadas): `#f9fafb` (`COR_FUNDO_CLARO`)
- Escuro (footer/faixa de stats): `#111827` (`COR_ESCURO`)
- Mudo no escuro: `#9ca3af` (`COR_MUTED_NO_ESCURO`)

### Semântica (ausente hoje na Home — só relevante se a demo ganhar estados de formulário/alerta)
Os tokens já existem no catálogo de componentes via `alerta` (`PreviewRenderer.tsx:171-180`,
`ESQUEMAS_ALERTA`) — reaproveitar, não reinventar:
- Sucesso: `#15803d` texto / `#f0fdf4` fundo
- Alerta: `#b45309` texto / `#fffbeb` fundo
- Erro: `#b91c1c` texto / `#fef2f2` fundo
- Info: `#1d4ed8` texto / `#eff6ff` fundo

## Tipografia

### Famílias
- **Corpo (validado, sem mudança)**: Inter — carregada em `index.css:1` e `tailwind.config.js:18`
  como fonte padrão. Referência: Inter aparece em 182 de sites SaaS analisados pelo ranking
  saaslandingpage.com (ver `02-referencias.md`) — é a escolha nº1 do mercado, o projeto já acertou.
- **Display/headings (recomendação nova)**: Outfit — **já carregada** em `index.css:1`
  (`Outfit:wght@400;500;600;700;800`) mas **nunca usada** em nenhum componente, porque `Estilos`
  (`componentRegistry.ts:45-68`) não tem campo de família de fonte — todo texto herda a fonte do
  app (Inter). Referência: Outfit descrita como "geometric sans-serif com personalidade mais
  amigável que Inter, terminações arredondadas" (saaslandingpage.com) — a combinação Inter (corpo)
  + Outfit (headings) é exatamente o padrão "duas vozes tipográficas" das referências de tendência
  2025 (`01-contexto.md`), com **custo zero de rede** (fonte já baixada pelo app, só falta ligar).
  Requer: adicionar `fonteFamilia?: 'inter' | 'outfit'` a `Estilos` e um `if (estilos.fonteFamilia)
  css.fontFamily = ...` em `estilosCss.ts` — mudança pequena e localizada, fora do escopo desta
  auditoria de conteúdo (fica registrada aqui para uma iteração futura, se o usuário topar).

### Escala (consolidando os tamanhos já usados na Home, sem introduzir novos)
| Papel | Tamanho | Peso | Onde já aparece |
|---|---|---|---|
| Hero heading | 44px | bold (700) | `hero-heading` |
| H2 de seção | 28-30px | bold (700) | `features-heading`, `planos-heading`, `depoimentos-heading`, `faq-heading` |
| H3 de card | 16-20px | bold (700) | `feat-N-titulo`, `plano-N-nome` |
| Corpo | 14-17px | normal (400) | parágrafos gerais |
| Legenda/label | 12-13px | medium (500) | badges, rótulos de stat, copy de footer |

Nenhuma mudança recomendada na escala — já é consistente e segue uma progressão razoável
(44 → 28-30 → 16-20 → 14-17 → 12-13). O ganho está em variar a **família**, não o **tamanho**.

## Espaçamento e Raio (ajuste de consistência, ver `03-principios-aplicados.md` item 3)

Escala de `border-radius` recomendada (hoje varia livremente entre 8-16px por seção sem critério):

| Papel | Raio | Onde aplicar |
|---|---|---|
| Pequeno (badge, botão, input) | `8px` | já é o padrão predominante — manter |
| Médio (card) | `16px` | unificar `feature_card`/`testimonial_card` (hoje 12px) e `pricing_card` (já 16px) neste valor |
| Grande (imagem hero, showcase) | `20px` | subir de 16px — reforça hierarquia (imagens "flutuam" mais que cards) |

## Sombra (nova — camadas em vez de valor único)

Padrão recomendado para qualquer elemento "elevado" (imagem do hero, card de pricing em
destaque, cards de depoimento), substituindo a sombra única atual por duas camadas (ambiente +
contato) — técnica descrita nas referências de tendência 2025 como comum em produtos como
Linear/Stripe, e 100% suportada hoje pelo campo `sombra` (string livre → `box-shadow`):

```
sombra: "0 1px 2px rgba(17,24,39,.06), 0 24px 48px -12px rgba(17,24,39,.18)"
```

Para elementos com a cor da marca (CTA primário, pricing em destaque), a camada de contato usa a
cor primária em vez de neutro puro:

```
sombra: "0 1px 2px rgba(79,70,229,.15), 0 20px 40px -8px rgba(79,70,229,.35)"
```
