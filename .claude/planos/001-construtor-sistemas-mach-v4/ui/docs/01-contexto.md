# Project Context

## Domain

**MACH V4 System Builder** is a **Low-Code/No-Code multi-tenant** (SaaS) platform that lets users build digital applications through a visual interface, without writing code. The architecture follows the **MACH** pillars (Microservices, API-first, Cloud-native, Headless), with 5 core microservices, real-time collaboration (Elixir/Phoenix), instant publish/rollback, and asynchronous data export.

The product belongs to the **visual app builders / internal tooling builders** category — the same competitive space as Retool, Webflow, Framer, Builder.io, Plasmic, Appsmith, and Budibase. The technical differentiator is security by design (Blind Index / LGPD anonymization, multi-tenant isolation) and Figma-style real-time collaboration.

### UI surfaces identified in the spec

| Surface | Actors | Associated FRs | State in the spec |
|---|---|---|---|
| **Systems Dashboard** — list of the tenant's systems, versions, publication status | Creator, Administrator | FR01, FR04 | Implicit (product front end) |
| **Visual Builder** — drag-and-drop canvas, component tree, properties panel, real-time collaboration with cursors/presence/locking | Creator, Collaborator | FR01, FR02, FR06 | Visual editor outside *back-end* scope, but is the face of the product |
| **Headless Player** — renderer for published dynamic forms, with Blind Index validation | End Customer | FR07 | **In scope** (web SPA) |
| Supporting panels — Publish/Rollback, Export, Permissions | Creator, Administrator | FR04, FR05, FR03 | In scope (REST contracts) |

> **Scope note**: spec `001` covers the back end, gRPC contracts, and the Headless Player. The drag-and-drop visual editor is declared as its "own demand" (spec §8). This UI work anticipates the design of those screens to guide the future demand, while keeping the Headless Player (in scope) as the priority deliverable.

## Target Audience

Three distinct profiles, with opposing UI needs — the design must serve all three without compromising any of them:

1. **Creator/Collaborator** (technical-intermediate): assembles systems in the Builder. Expects information density, keyboard shortcuts, immediate feedback, fluid collaboration. Mental reference: Figma, Retool, Linear.
2. **Administrator (Owner/Partner)**: manages hierarchical tenants, component-level permissions, exports. Expects clarity, auditability, control. Mental reference: enterprise governance dashboards.
3. **End Customer** (layperson): only fills out forms published via the Headless Player. Expects simplicity, zero friction, clear validation. Mental reference: Typeform, Google Forms.

Usage context: **desktop-first** for Builder/Dashboard (building work on large screens), **mobile-first** for the Headless Player (forms consumed on any device).

## Visual References Found

| Reference | Why it's relevant | Popularity metric |
|---|---|---|
| **Figma — Multiplayer / Live Cursors** | Gold standard for real-time collaboration (named colored cursors, presence, follow-mode) — directly applicable to FR06 | ~US$20B valuation; co-editing increases project speed by ~35% (Product Brief/Medium) |
| **Vercel Geist Design System** | Reference *dark-first* system: pure black `oklch(0 0 0)`, 200-step gray scale, Geist font | Widely copied public design system of 2024–25; foundation of thousands of shadcn projects |
| **Linear** | Restraint case study: near-black surfaces + 1 accent color; density without noise | Benchmark cited as the "dark-first" market standard for 2025 |
| **Retool** | Internal tooling builder: canvas + properties panel + Inter at dense sizes; orange accent `#EF5350` + blue `#3D5AFE` | Market leader in internal tooling; direct reference for builder layout |
| **Framer** | Canvas with animation/microinteractions at its core; real-time transition preview | Award-winning platform (Awwwards) for interactive prototyping |
| **WeWeb / Budibase / Appsmith** | Multi-tenant no-code builder patterns (app list, environments, publish) | "Best drag-and-drop app builders 2025/2026" reviews (WeWeb, UI Bakery, Zapier) |

## Trends Identified (2025–2026)

1. **Dark-first as the default, not a toggle** — Linear and Vercel treat dark as the canonical surface; light is the alternative. Building tools are used for long hours; dark reduces fatigue. *(Apply: Dashboard and Builder dark-first; Player light by default for the layperson End Customer.)*
2. **Collaboration as communication** — the cursor is no longer just presence, it's intent: follow cursor, comment in-context, visual component locking. *(Apply directly to FR06/BR07.)*
3. **AI-assisted building** — UI generation via prompt/image (Mendix Maia, Framer) and contextual layout suggestions. *(Reserve space in the Builder for a future "AI panel".)*
4. **Visual + code (enterprise governance)** — auditability, SSO, visible versioning, custom component import. *(Apply to the versions/rollback panel — FR04 — and to the tenant hierarchy.)*
5. **Performance-focused defaults** — visual apps need to load fast; render batching (NFR07: 16ms batches) and skeleton screens instead of spinners.

---

**Sources**:
- [Figma's Collaborative Canvas (Medium/Product Brief)](https://medium.com/@productbrief/figmas-collaborative-canvas-how-real-time-design-built-a-20-billion-creative-empire-efefc6126a93)
- [Figma's Live Cursor UI (Designilo, 2025)](https://designilo.com/2025/07/20/figmas-live-cursor-ui-a-new-era-of-design-dev-collaboration/)
- [Vercel Geist — Colors](https://vercel.com/geist/colors)
- [Geist Design System Breakdown (DesignSystems.one)](https://www.designsystems.one/design-systems/vercel-geist)
- [19 Best Dark Mode Dashboard Templates 2026 (AdminLTE)](https://adminlte.io/blog/dark-dashboard-templates/)
- [25 Best Drag-and-Drop App Builders (WeWeb)](https://www.weweb.io/blog/drag-and-drop-app-builder-tools)
- [10 Top Drag and Drop app builders 2025 (UI Bakery)](https://uibakery.io/blog/drag-and-drop-app-builders)
