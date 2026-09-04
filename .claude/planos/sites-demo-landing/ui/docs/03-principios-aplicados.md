# Applied Principles

> Before the principles: what the rendering engine actually supports. Every recommendation below
> was checked against `services/frontend/src/systems/estilosCss.ts` (converts `Estilos` → inline
> CSS) and `services/frontend/src/pages/Dashboard/editor/PreviewRenderer.tsx` (renders each
> component `tipo`). Confirmed 1:1 mapping: `fundoCor`→`background-color` (solid color
> **only**, no gradient), `sombra`→`box-shadow` (free-form string — **multiple layered shadows
> work**), `bordaRaio`→`border-radius` (accepts shorthand like `"16px 16px 0 0"` — asymmetric
> radius works), **no** exposed `transition`/`transform`/`animation`, **no** CSS Grid
> (`display` only accepts `block|inline|inline-block|flex|none`), **no** font-family field
> (all text inherits the app's global `Inter` font, even though `Outfit` is loaded and available).
> This defines what is a **content** recommendation (applicable today, just editing
> `seed-demo-site.sh`) vs. a **platform** recommendation (requires a new field in `componentRegistry`).

## 1. Obvious Start
The hero's primary CTA (`hero-cta-primario`, "Start for free") already has the highest visual
prominence on the entire page: solid brand color + colored shadow (`0 8px 16px rgba(79,70,229,.25)`,
line 433), while the secondary one (`hero-cta-secundario`, "Watch demo") is outlined. Correct,
no change. Recommended reinforcement: apply the same colored-shadow signature to the page's final
CTA too (`cta-final-botao`, line 548) — today it's just a solid `fundoCor` with no shadow,
losing the visual "weight" the hero's equivalent button has, despite being the page's last chance
at conversion.

## 2. Clear Reversal
Not applicable to a marketing landing page — there are no destructive actions on this screen (it
differs from the Contact form, already covered by another screen). No change recommended.

## 3. Consistent Logic
Finding: the two showcases (`showcase_section`, line 264) use `bordaRaio:"16px"` on the image
(line 291) while the feature/testimonial/pricing cards use `12-16px` varying by type, with no
single pattern. Recommendation: adopt a fixed 3-value scale (12px for small elements/badges, 16px
for cards, 20px for large blocks like the hero image) documented in `04-sistema-cores-tipografia.md`,
instead of ad-hoc per-section values.

## 4. Follow Conventions
The structure already follows the market convention validated in `02-referencias.md`: sticky-like
header with logo+menu+CTA, hero with an eyebrow badge, logo cloud right below the hero, "Most
Popular" with visual emphasis on pricing. No structural change — just visual refinement (items 5-9
below).

## 5. Feedback and Milestones
Not applicable to static marketing content (loading/error states belong to the dashboard, already
covered in `auditoria-ui-projeto`). The only real "state" on the Home is the FAQ accordion
(`AccordionPublicado`, already functional — opens/closes) and the pricing/testimonial cards, which
are static by nature.

## 6. Proximity and Adaptation
The page is fixed to a desktop width (`hero-imagem` has a fixed `largura:"860px"`, line 442; no
media query in the styles model) — there's no way to apply real responsiveness without a platform
change (grid/breakpoints), which is out of scope for content. Recorded as a known risk, not
as a recommendation for this round (the Canvas itself already has a device preview — mobile/tablet/
desktop — but the published tree uses absolute `largura` values, not fluid ones).

## 7. Interface Is Content
Main finding of this audit: testimonial avatars use only initials
(`testimonial_card`, line 313, `propriedades:{texto:$iniciais,...}`) — generic text standing in
for a "face". Research (`02-referencias.md`, NN/g via Lovable) shows that a real photo
(even a photo placeholder) communicates far more than abstract initials. Since the `avatar`
type already supports `src`/`alt` (`PreviewRenderer.tsx:130-138`), the swap is purely a content
change — see wireframe `depoimento-social-proof.html`.

## 8. General Visual Design Principles
- **Make the subject obvious**: already met (the "✨ New: AI automation" eyebrow badge at the
  top of the hero identifies the product context before the heading even loads).
- **Integrated form and content**: the Pro plan's "MOST POPULAR" badge already uses the primary
  color to reinforce emphasis — correct. New recommendation: use the same "dual shadow layer"
  technique (diffuse ambient + sharp contact) instead of a single shadow on every elevated element
  (hero image, highlighted pricing card, testimonial cards) — a popular technique in products like
  Linear/Stripe (cited in the 2025 trend references) and fully supported by `sombra` (a free-form
  `box-shadow` string that accepts multiple comma-separated values).
- **Metaphors for new concepts**: not applicable — the product is task management, a concept
  already familiar to the B2B target audience.

## 9. Design Decision Matrix

| Decision | Obvious Start | Clear Reversal | Consistency | Convention | Feedback | Proximity | Content > Deco. |
|---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| Layered shadow (hero image, highlighted pricing, cards) | ✓ | — | ✓ | ✓ | — | ✓ | ✓ |
| Testimonial avatars with a photo instead of initials | ✓ | — | ✓ | ✓ | — | — | ✓ |
| Aggregate rating badge above testimonials | ✓ | — | ✓ | ✓ | — | ✓ | ✓ |
| Colored shadow on the final CTA (parity with the hero) | ✓ | — | ✓ | ✓ | — | — | — |
| Fixed border-radius scale (12/16/20px) | — | — | ✓ | ✓ | — | ✓ | — |
| `fonteFamilia` field in the registry (Outfit on headings) | ✓ | — | ✓ | ✓ | — | — | ✓ |

## Out of scope for this round (recorded, not implemented)

- **Background gradients** (a strong 2025 trend, see `01-contexto.md`) — blocked by the current
  styles model (`fundoCor` only accepts a solid color). Would require a `fundoGradiente` field
  mapped to `background: linear-gradient(...)` in `estilosCss.ts` — a platform change, not a
  content change for this demo.
- **Real responsiveness** (breakpoints) — the `Estilos` model has no concept of per-device
  values; the Canvas simulates 3 widths in the editor, but the published tree is fixed. A larger
  platform change, out of scope for "improving the demo site's interface".
