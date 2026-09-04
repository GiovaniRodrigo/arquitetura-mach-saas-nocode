# Popular References

All the references below belong to the **same domain** as the product (visual/no-code builders, collaboration tools, and dev-tools design systems) and carry popularity metrics — per the skill's scoping rule (only recommend what has evidence of adoption).

## Products and Design Systems

| Reference | URL | Popularity | Applicability |
|---|---|---|---|
| **Figma — Multiplayer Editing** | https://www.figma.com/blog/multiplayer-editing-in-figma/ | ~US$20B valuation; co-editing raises project speed ~35%; fluid cursors with 10+ simultaneous editors | Model for the Builder's **named colored cursors, presence, and follow-mode** (FR06). Operation-optimized diffs = conceptual basis for batching (NFR07) |
| **Vercel Geist Design System** | https://vercel.com/geist/colors | Reference public design system 2024–25; adopted by thousands of shadcn projects | **Dark-first palette**: ink `#171717`, body `#0A0A0A`/`#fafafa`, 200-step gray scale for borders/dividers/disabled. Basis for the Dashboard/Builder neutral palette |
| **Linear** | https://linear.app | Market benchmark cited as the "dark-first" standard for 2025 | Restraint: 1 accent color over near-black surfaces; density without noise; `Cmd+K` command. Model for the command bar and the Dashboard's visual hierarchy |
| **Retool** | https://retool.com | Market leader in internal tooling | **3-column builder layout** (library · canvas · properties); Inter at dense sizes; orange accent `#EF5350` + blue `#3D5AFE`. Direct template for the Visual Builder |
| **Framer** | https://www.framer.com | Award-winning (Awwwards); interactive prototyping with animation at its core | Real-time transition/microinteraction preview on the canvas. Reference for the Builder's **Preview mode** and the Player's 60Hz animations (NFR07) |
| **WeWeb** | https://www.weweb.io/blog/drag-and-drop-app-builder-tools | Listed among the "25 best drag-and-drop app builders 2026" | **Publish/environments** patterns and multi-tenant project lists. Basis for the Dashboard and the Publish/Rollback flow (FR04) |
| **Budibase / Appsmith** | https://uibakery.io/blog/drag-and-drop-app-builders | "Top drag-and-drop app builders 2025" reviews (UI Bakery) | Open-source multi-tenant no-code builder patterns: data panel, configuration-based binding. Reference for the **Business Rules** panel (FR02) |
| **Typeform / Google Forms** | https://www.typeform.com | Reference for mass-form UX | Model for the **Headless Player**: one field in focus, inline validation, progress indicator, mobile-first (FR07) |

## Guidelines and UX Research

| Source | URL | Authority | What to apply |
|---|---|---|---|
| **Nielsen Norman Group — Form Design** | https://www.nngroup.com/articles/web-form-design/ | Evidence-based UX research | Top-aligned labels (faster completion); inline validation on *blur*; fewer fields = higher completion rate. Applied to the Player (FR07) |
| **Inline Validation — SubUX** | https://subux.pro/guides/article/inline-validation | Consolidated guide | Validate on **field blur**, not while typing; specific message near the field. Applied to the error map by `blind_index` (BR08) |
| **Material Design 3** | https://m3.material.io | Google HIG | Component states (hover/focus/disabled), elevation, text fields, chips. Conventions for the Player and the Builder's controls |
| **Apple HIG** | https://developer.apple.com/design/human-interface-guidelines | Apple HIG | Icon semantics, hierarchy of destructive actions/confirmation (Rollback, Delete) |
| **Multi-Step Form UX (Growform)** | https://www.growform.co/must-follow-ux-best-practices-when-designing-a-multi-step-form/ | Conversion guide | "Step 2 of 4" indicator, field order, removing superfluous fields. Applied to multi-step forms in the Player |

## Synthesis of decisions drawn from the references

- **Theme**: dark-first (Geist/Linear) for Dashboard/Builder; light-first for the Player (layperson audience).
- **Builder layout**: Retool's 3 columns (library · canvas · properties) + a Figma-style collaboration layer on top.
- **Typography**: Inter (body/dense UI, dev-tools standard) + Geist/Inter Display for titles; JetBrains Mono for `blind_index` and technical values.
- **Accent color**: a single accent (Linear restraint) — violet/indigo, distinct from generic blue, reserved for the primary CTA and active state.
- **Forms**: inline validation on blur (NN/g + SubUX), per-field errors, visible progress.

---

**Sources**:
- [Multiplayer Editing in Figma](https://www.figma.com/blog/multiplayer-editing-in-figma/)
- [Vercel Geist — Colors](https://vercel.com/geist/colors)
- [Geist Design System Breakdown (DesignSystems.one)](https://www.designsystems.one/design-systems/vercel-geist)
- [Drag-and-Drop App Builders (WeWeb)](https://www.weweb.io/blog/drag-and-drop-app-builder-tools)
- [Top Drag and Drop app builders 2025 (UI Bakery)](https://uibakery.io/blog/drag-and-drop-app-builders)
- [Inline Validation UX (SubUX)](https://subux.pro/guides/article/inline-validation)
- [Multi-Step Form UX Best Practices (Growform)](https://www.growform.co/must-follow-ux-best-practices-when-designing-a-multi-step-form/)
