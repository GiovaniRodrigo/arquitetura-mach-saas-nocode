# Institutional Design Sync (MACH V4 Player)

## 1. Overview
Refactor the Authentication (Login) and System Selection (App Launcher) interface of the MACH V4 Player to align with the Institutional Site's Design System (Custom Vanilla CSS), implementing TailwindCSS and the Shadcn UI base.

## 2. Functional Requirements (RF)
*   **FR01:** The Player must allow login via Google and GitHub, displaying the screen in Split Screen format.
*   **FR02:** The Player must list the systems available to the authenticated user, using a Card Grid (Bento Grid / App Launcher).
*   **FR03:** The user must be able to create a new system from the interface (existing functionality preserved visually).

## 3. Non-Functional Requirements (RNF)
*   **NFR01 (UI Stack):** Must use Tailwind CSS, `clsx`, `tailwind-merge`, and `lucide-react`.
*   **NFR02 (Typography):** Incorporate and use the "Outfit" (headings) and "Inter" (body) fonts imported from Google Fonts.
*   **NFR03 (Palette and Aesthetics):** Adopt primary color `#6366f1` (Indigo). Use Dark theme as the default. Adopt `rounded-full` radii for buttons and `rounded-2xl` for large cards, with smooth hover effects.
*   **NFR04 (Maintainability):** Remove any inline object-style style declarations (`const estilos`) from React components, migrating 100% to Tailwind utility classes.

## 4. Files Involved
*   **New/Added:**
    *   `player/tailwind.config.js`
    *   `player/postcss.config.js`
    *   `player/src/index.css`
    *   `player/src/lib/utils.ts`
*   **Modified:**
    *   `player/package.json` (Dependencies)
    *   `player/index.html` (Fonts and dark class on html)
    *   `player/src/main.tsx` (index.css import)
    *   `player/src/auth/Login.tsx` (Visual refactor)
    *   `player/src/systems/SeletorSistemas.tsx` (Visual refactor)

## 5. Acceptance Criteria
*   TypeScript and Linting tests pass successfully (if any).
*   The visuals match the documented design system.
*   Code free of old inline classes and styles.
