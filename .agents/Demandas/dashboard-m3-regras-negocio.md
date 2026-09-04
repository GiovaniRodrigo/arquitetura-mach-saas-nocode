# Material Design 3 (M3) Dashboard - Business Rules

## File Architecture
- `player/src/layout/DashboardLayout.tsx`: Main container with Navigation Rail (side menu) and top Top App Bar, providing the `<Outlet />` for the pages.
- `player/src/pages/Dashboard/Overview.tsx`: The main dashboard page, implementing the "Hero Card" and the rounded "Metrics Cards".
- `player/src/components/m3/`: New folder for pure M3 components (`FabButton.tsx`, `TonalCard.tsx`, `ElevatedCard.tsx`, `NavPill.tsx`).

## Functional Requirements (RF)
- **FR01:** The system must render a side navigation menu (Navigation Drawer/Rail) exclusive to logged-in users.
- **FR02:** The system must display a header ("Top App Bar") containing a welcome message and the user's avatar.
- **FR03:** The main screen (Overview) must display a welcome card with a call to action ("Hero Card").
- **FR04:** The main screen must display indicator cards (Status Cards) based on platform data.
- **FR05:** The screen must provide a Floating Action Button (FAB) for quickly creating projects/flows.

## Non-Functional Requirements (RNF)
- **NFR01:** The entire interface must use the **Material Design 3 (M3)** visual aesthetic, with pronounced corners (`rounded-3xl` and `rounded-full`), tonal colors, and soft elevations.
- **NFR02:** The layout must be responsive (the side menu must be hidden or adapted on mobile devices).
- **NFR03:** Interactive components must have clear states (`hover`, `focus`, `active:scale-95`) to provide tactile feedback to the user.
