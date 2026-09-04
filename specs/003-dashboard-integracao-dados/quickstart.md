# Quickstart: Dashboard — Data Integration and Functionality

Guide to run and test this implementation locally, inside the `player/` package.

---

## Prerequisites

- Node.js + npm installed
- Player dependencies installed (`npm install` in `player/`)
- One of these session options:
  - A reachable Gateway and a valid JWT (OAuth flow from `auth/session.ts`), **or**
  - `VITE_BYPASS_AUTH=true` for development without login

---

## Steps

```bash
# 1. Enter the front-end package
cd player

# 2. Install dependencies (if not already installed)
npm install

# 3. (Optional) Point at the Gateway and/or skip auth in development
#    Create player/.env.local with:
#    VITE_GATEWAY_URL=http://localhost:8080
#    VITE_BYPASS_AUTH=true

# 4. Start the development environment
npm run dev
```

Open the browser at `/dashboard` (the root route redirects there when there is no
active system). Check:

- **Overview**: metrics reflect the real number of systems; "Get Started"/FAB
  start system creation (no `alert()`).
- **Projects**: grid of real systems with skeleton → data/empty/error; "Open project"
  navigates to the system.
- **Settings**: "Toggle Theme" switches light/dark and persists after reloading.
- **Header**: shows the user's real name/initials.

---

## Verification

```bash
# This effort's tests (inside player/)
npm run test -- src/systems/useSistemas.test.ts
npm run test -- src/theme/ThemeProvider.test.tsx
npm run test -- src/auth/jwt.test.ts
npm run test -- src/pages/Dashboard

# Full suite + type-check + build
npm run test
npm run build   # tsc --noEmit && vite build
```

Quick "done" criteria:
- No dashboard data is hardcoded (metrics, cards, name/avatar).
- Every button performs a real action; no remaining `alert()`.
- Theme persists with no flash on reload.
- Loading/empty/error states are present on screens with data.

---

## Environment Variables

| Variable | Example Value | Description |
|----------|-----------------|-----------|
| `VITE_GATEWAY_URL` | `http://localhost:8080` | Gateway base URL (empty = relative calls via the Nginx proxy) |
| `VITE_BYPASS_AUTH` | `true` | Skips the login gate in development |
| `mach_token` (localStorage) | — | JWT persisted by the session; source of the identity claims (FR03) |
| `mach_theme` (localStorage) | `escuro` | Persisted theme preference (FR05) |
