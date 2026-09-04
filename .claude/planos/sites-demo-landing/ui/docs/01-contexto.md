# Project Context

## Domain

This `/ui` covers a different target than the previous audits (`001-construtor-sistemas-mach-v4`,
`008-monitor-recursos`, `auditoria-ui-projeto`): it's not the MACH V4 **internal dashboard** (the
tool owners/partners use to operate the platform), but rather the **final product that the
no-code Design Engine generates** — the published site the account owner's client sees.

Specifically: the **"Home"** screen of the "Loja Demo" demo system, a complete B2B SaaS landing
page ("Brillance" — inspired by the "Brillance SaaS Landing Page" template from v0.app),
assembled entirely from the Design Engine's component catalog
(`services/frontend/src/systems/componentRegistry.ts`) via the
`build/seed-demo-site.sh` script (707 lines, generates 3 screens: Home, Produtos, Contato — the
focus here is Home, the most elaborate one and the one that showcases what the no-code builder
can produce).

Current Home structure (script line numbers in parentheses): navbar (170-189), hero (422-447),
logo cloud (448-465), a grid of 6 feature cards in 2 rows (466-489), two alternating text/image
showcases (264-299, instantiated at 396-397), a 4-block stats strip
(301-311, 492-496), 3 testimonial cards (313-339, 498-508), 3 plan/pricing cards
(341-385, 509-526), an FAQ accordion (528-541), a final CTA (542-551), and a 5-column footer
(206-248).

## Target Audience

Two distinct audiences, and therefore two different readings of this audit:

1. **The editor** (MACH owner/partner, semi-technical): uses the Canvas to assemble screens like
   this one by dragging the same components — so every improvement recommendation needs to be
   **achievable with the current component catalog** (see `03-principios-aplicados.md` for the
   exact survey of what the rendering engine supports), or explicitly flagged as
   "requires a new registry field".
2. **Whoever visits the published site** (the owner/partner's end customer — here, the "Brillance"
   fictional target audience is a B2B buyer evaluating team-management tools). This is the
   audience that the trend research in `02-referencias.md` targets.

## Visual References Found

| Reference | URL | Relevance |
|---|---|---|
| Notion / Linear / Framer (hero storytelling) | cited in SaaSFrame Blog, "10 SaaS Landing Page Trends for 2026" | Hero sections that show the product and a "before → after" narrative instead of just stating what the product is — the Home already partially follows this (headline + panel image), and it can be reinforced. |
| Dribbble — `saas-hero` / `hero-section` tag | dribbble.com/tags/saas-hero | Thousands of designs — used to validate the "eyebrow badge + heading + dual CTA + floating social proof" pattern the Home already uses (`hero-eyebrow`, `hero-badge-flutuante`). |
| Dribbble — `pricing-page` / `pricing-card` tag | dribbble.com/tags/pricing-page (1,800+ designs) | Confirms the "Most Popular" pattern with a highlighted border/shadow, already implemented in `pricing_card()` — validated, no change needed. |
| Nielsen Norman Group (via SaaS Website Best Practices, Lovable/Brand Vision) | see `02-referencias.md` | Social proof: testimonials with a photo + full name are rated as more trustworthy than anonymous initials; an aggregate rating badge ("4.8/5, 3,200 reviews") is more credible than an isolated quote. |
| Figma — Indigo/Violet color pages | figma.com/colors/indigo, figma.com/colors/violet | Indigo (the current primary color, `#4f46e5` = Tailwind indigo-600) is described as the preferred middle ground in premium SaaS — more personality than blue, more accessible than violet. Current palette **validated**, not dated. |
| SaaS Landing Page — font ranking | saaslandingpage.com/articles/the-12-most-popular-google-fonts-for-landing-pages | Inter appears in 182 SaaS sites analyzed — the market's #1 font. The project already uses Inter as its default font (`tailwind.config.js`). **Validated.** |

## Trends Identified

1. **Typography with two voices** (heading in a more expressive "display" font, body in a
   neutral font) — instead of a single family for everything.
2. **Social proof with a face** — a real photo (or photo placeholder) instead of avatar
   initials, plus an aggregate credibility number (rating/count) near the testimonials section.
3. **Layered depth** (multiple shadows — ambient + contact — instead of a single shadow)
   to give a sense of elevation without needing a gradient.
4. **A slightly larger border radius** ("soft UI") on cards and highlighted buttons, a 2025 trend
   compared to the more "corporate" 8-12px of previous years.
5. **Gradients and color overlays** — a strong market trend, but **not supported today** by
   the Design Engine's style model (`fundoCor` maps only to a solid `background-color`, never
   to `background: linear-gradient(...)`) — recorded as a platform improvement opportunity, not
   as a content recommendation (see `03-principios-aplicados.md`).
