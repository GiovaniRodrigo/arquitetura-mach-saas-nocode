# Design Principles Applied to the Player Screens

Evaluation of the Headless Player's real screens (`Login.tsx`, `App.tsx` states,
`CompositeRenderer.tsx`) against 9 principles. Each item lists **current state → gap → action**.

## Screens evaluated
- **T1 — Login** (no session): "MACH Platform" title, subtitle, 2 text social links.
- **T2 — Loading**: `<div>Loading…</div>`.
- **T3 — Empty / no system**: "Authenticated ✓. No system selected…".
- **T4 — Error**: `<div role="alert">{error}</div>` (raw error string).
- **T5 — Dynamic screen**: `nav` of links + rendered form + button (no styling of its own).

---

### 1. Obvious Start
- **T1**: there's a starting point (2 buttons), but **no highlighted primary CTA** — Google and
  GitHub carry identical visual weight. Action: promote 1 primary ("Continue with Google", solid,
  colored) with GitHub as secondary (outline).
- **T5**: `nav` of links with no active-item indicator or "first step" cue. Action: highlight the
  active route and the form's main CTA.

### 2. Clear Reversal
- **T4**: the error is a dead end — raw text, **no "Try again" button** nor "Back to login".
  Action: add retry and an exit link.
- **T5**: form submission with no visible "Cancel"/clear option. Action: Cancel/Submit pair.

### 3. Consistent Logic
- **Global**: only Login has styling; T2–T5 render with the browser's default appearance →
  **severe visual inconsistency** between screens of the same product. Action: a single design
  system (color/typography/spacing tokens) applied to every screen and to the CompositeRenderer.
- `:hover`/`:focus` states are undefined on the login links. Action: standardize them.

### 4. Follow Conventions
- **T1**: social buttons have **no IDP logos** — a deviation from Google's guideline (4-color G)
  and GitHub's (official mark). Action: include the official logos.
- Semantic icons are missing in T4/T5 (error with no alert icon). Action: use universal icons.

### 5. Feedback and Milestones
- **T2**: "Loading…" is generic text. Action: **skeleton screen** with the screen's silhouette.
- **T5**: submission with no clear loading/success state beyond textual `status`. Action: success
  toast + "submitting…" button state.
- Validation errors (`data-bi`) appear as `<p role="alert">` with no visual anchor to the field.
  Action: inline message under the field, with semantic color.

### 6. Proximity and Adaptation
- **T1**: a centered `max-width:360px` is fine, but **there are no media queries** nor
  guaranteed touch targets (12px padding → target < 44px in some cases). Action: mobile-first,
  targets ≥44px.
- **T5**: validation errors are grouped **after** the form, far from the fields. Action: move the
  message closer to the field it affects.

### 7. Interface Is Content
- **T1**: lean (good), but the leanness turns into a **lack of identity** (no logo/brand).
  Action: add just the essential brand elements (logo + name), with no superfluous decoration.
- **T3**: the empty state instructs via `<code>?sistema=…</code>` — technical jargon exposed to
  the end user. Action: turn it into a guided action ("Select system") instead of a query param.

### 8. General Visual Design Principles
- **Obvious subject**: the Login lacks a logo/MACH identity at the top, and the dynamic screens
  lack a context title. Action: header with branding.
- **Appropriate data**: T5 renders forms; fine. Wherever there are lists/CRUD, use a table.
- **Form and content**: error has no semantic color (red) and success has no green. Action: apply
  the semantic palette (doc 04).
- **Metaphors**: "system" and "active version" are new concepts to a layperson — support them
  with microcopy.

### 9. Design Decision Matrix (current state)

| Decision / Screen | Obvious Start | Clear Reversal | Consistency | Convention | Feedback | Proximity | Content > Deco. |
|---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| T1 — Login (social buttons) | ✗ | ✓ | ✗ | ✗ | ✗ | ✗ | ✓ |
| T2 — Loading | — | — | ✗ | ✗ | ✗ | — | ✓ |
| T3 — Empty / no system | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ |
| T4 — Error | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✓ |
| T5 — Dynamic screen (form) | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✓ |

Legend: ✓ meets · ✗ does not meet · — not applicable.

## Priorities (highest impact first)
1. **Single design system** applied to every screen (fixes Consistency in T2–T5).
2. **Login redesign**: card + logo, primary CTA, IDP logos, focus/hover, responsive.
3. **Real states**: skeleton (T2), empty with an action (T3), error with retry (T4).
4. **Form feedback**: inline validation + success toast + submission state (T5).
