# Project Context

## Domain
**MACH V4 — Low-Code / No-Code Platform.** The software lets users build their own digital
applications through a visual interface (UI CRUD, business rules, instant publishing,
hierarchical multi-tenancy, real-time collaboration). The part audited here is the **Headless
Player** — the SPA that renders the *active* version of a system built on the platform, served at
`https://gfcode.com.br/ui/`.

## Target Audience
Two profiles, both served by the same shell:
- **Builders / operators** (semi-technical): owners and partners who create and publish systems.
- **End customers** (laypeople): users of the generated systems, who only consume the screens
  (nav + forms + submission).

Usage context: desktop and mobile browsers; session started via **social login** (Google/GitHub).
Since the rendered screen is the "face" of the product the builder's client delivers to *their*
customer, the perceived visual quality of the shell (login, states, chrome) directly impacts
trust in the product.

## Frontend Stack (observed in the code)
- **Vite + React 18 + TypeScript**, `react-router-dom`. Served under base `/ui/`
  (`player/vite.config.ts` → `base: "/ui/"`).
- **No CSS framework** (no Tailwind, MUI, etc.). All styling is *inline* in `Login.tsx`;
  the other screens (loading/empty/error and `CompositeRenderer`) **have no styling of their
  own** — they render with the browser's default appearance.
- Entry point: `player/src/main.tsx` (decides Login vs App based on the token), `App.tsx`
  (dynamic routes + screen rendering), `CompositeRenderer.tsx` (Composite tree → DOM).

## Technical Validation (done in the browser — summary)
`https://gfcode.com.br/ui/`: `/ui/` 200, `/ui/assets/index-*.js` 200 (correct base path), the SPA
mounts, **no console errors**, the Login screen renders, buttons point to
`/auth/{google,github}` (Nginx proxy ok). **Functionally correct** — this audit is about *design*
quality.

## Visual References Found (web)
| Reference | Why it's relevant | Popularity |
|---|---|---|
| **Appsmith** (no-code builder, login/app screen) | Same domain (low-code builder); login templates | ~30,000+ GitHub stars; 5,000+ Discord members |
| **Budibase** | No-code builder with auto-CRUD and auth screens | ~300,000 teams using the platform |
| **ToolJet** | AI-native low-code builder, UI ecosystem | 500+ GitHub contributors |
| **Bubble** | Reference customer-facing no-code builder | Millions of apps built; no-code market leader |
| **Retool** | Library of 100+ configurable components | Used by thousands of companies (internal-tools standard) |
| **Material Design 3** (Google) | Official guideline for buttons/states/colors | Android/Google's official design system |
| **Sign in with Google — Best Practices** | Official guideline for the social button | Official Google Identity documentation |

## Trends Identified (applicable to Login/shell)
1. **Minimalist social login with clear hierarchy**: 2–3 methods at most, highlighting the
   primary provider ("Continue with…") — avoids cognitive overload (Authgear/Lollypop 2025).
2. **Social buttons following each IDP's guideline**: Google's colored logo (4-color G),
   GitHub's mark; when the IDP doesn't publish a button spec, apply the logo + the product's own
   button pattern for consistency.
3. **Mobile-first with large touch targets** and a responsive layout — assume login on mobile.
4. **Centered card with brand identity** (logo + name + microcopy) instead of loose text.
5. **Visible, non-generic states**: skeleton instead of spinner, empty state with an action,
   error with retry — in line with Material 3 and NN/g.

## Sources
- Login/Signup UX 2025 — https://www.authgear.com/post/login-signup-ux-guide/
- SaaS login page design — https://lollypop.design/blog/2025/october/saas-login-page-design/
- Sign in with Google best practices — https://developers.google.com/identity/siwg/best-practices
- Designing with social login buttons — https://medium.com/@sabarivasan/designing-with-social-login-buttons-the-right-way-a-deep-dive-into-idp-guidelines-618742589c85
- Appsmith vs Budibase vs ToolJet — https://blog.tooljet.com/appsmith-vs-budibase-vs-tooljet/
- Open source low-code platforms — https://www.appsmith.com/blog/open-source-low-code-platforms
