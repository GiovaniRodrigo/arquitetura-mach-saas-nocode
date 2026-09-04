# Popular References (web)

References from the **same domain** (no-code/low-code builders) and official social-login
guidelines. All with evidence of adoption — no generic blogs.

| Reference | URL | Popularity | Applicability to the MACH player |
|---|---|---|---|
| **Appsmith** — login/app templates | https://www.appsmith.com/blog/open-source-low-code-platforms | ~30,000+ GitHub stars; 5,000+ Discord | Login screen and builder-rendered app pattern; consistent grid |
| **Budibase** — auth + auto-CRUD | https://budibase.com/blog/alternatives/appsmith-vs-budibase/ | ~300,000 teams on the platform | Form/CRUD states and access screens for a lay end user |
| **ToolJet** — low-code UI builder | https://blog.tooljet.com/appsmith-vs-budibase-vs-tooljet/ | 500+ contributors | Consistent component library; action feedback |
| **Bubble** — customer-facing no-code | https://blog.tooljet.com/appsmith-vs-budibase-vs-tooljet/ | No-code market leader, millions of apps | Customer-facing shell; brand identity in the login |
| **Retool** — 100+ components | https://www.appsmith.com/blog/retool-alternatives | Thousands of companies (internal-tools standard) | Table/form states and component consistency |
| **Material Design 3** — buttons/states | https://m3.material.io/ | Google/Android's official design system | Button spec, focus/hover/disabled, semantic colors |
| **Sign in with Google (guideline)** | https://developers.google.com/identity/siwg/best-practices | Official Google Identity documentation | Hierarchy and placement of the primary social button |
| **Login/Signup UX Guide 2025** | https://www.authgear.com/post/login-signup-ux-guide/ | Reference guide cited in the industry | Limit to 2–3 methods; mobile-first; large targets |
| **SaaS login page design 2025 (Lollypop)** | https://lollypop.design/blog/2025/october/saas-login-page-design/ | Award-winning design studio | Centered card, microcopy, brand identity |

## Applicable synthesis
- **Hierarchy**: highlight 1 primary method; max 2–3 total (we have exactly 2 — good).
- **IDP branding**: Google requires the 4-color "G"; GitHub uses its official mark. Today the
  player uses **text only**, no logo — a deviation from both IDPs' guidelines.
- **Card + identity**: wrap the login in a centered card with the MACH logo/name.
- **Mobile-first**: targets ≥44px, fluid width with `max-width`.
- **States**: skeleton/loading, empty with an action, error with retry — currently absent.
