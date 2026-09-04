# Project Context

## Domain

MACH V4 is the no-code SaaS platform "MAYS — Make Your SaaS": a visual system builder
(screens, components, flows) made up of 8 microservices (IAM, Design, Logic, Deploy, Export,
Workers, Collab, Gateway) running on Kubernetes with a Linkerd service mesh.

The **Resource Monitor** screen (`/dashboard/monitor`, spec `008-monitor-recursos`/`009`) is an
**internal infrastructure observability** screen: for each of the 8 services, it shows whether
it's up, CPU, memory, requests/s, success rate, and p99 latency — data coming from Kubernetes's
metrics-server and Linkerd-viz's Prometheus via `services/gateway/internal/meshmetrics`.
It is not a product screen aimed at the SaaS end customer — it's an operational screen, equivalent
to a DevOps/SRE status board embedded in the platform's own dashboard.

## Target Audience

Any authenticated dashboard user (BR03 — there is currently no separate "platform administrator"
role), but the *actual expected usage* is by a technical profile: whoever operates/maintains the
MACH V4 platform, looking at the screen to diagnose "is something down?" or "is a service under
load?". The language and information density should follow observability dashboard conventions
(Grafana, Datadog, Vercel, Railway) — not the convention for layperson-facing product screens.

## Frontend Stack (already implemented)

- **React + TypeScript**, Tailwind CSS with HSL tokens via CSS custom properties (`--primary`,
  `--secondary`, `--destructive`, `--muted-foreground`, etc.), light/dark theme support
  (`.dark` on the root, toggle already implemented in `DashboardLayout.tsx`).
- **"M3" component system** (`src/components/m3/`): `ElevatedCard` (`--card` background,
  `rounded-3xl`, shadow), `TonalCard` (`--secondary` background, no shadow, for section emphasis),
  `FabButton`, `NavPill` — naming inspired by Material Design 3 (elevated/tonal/filled),
  but applied over a visual style closer to modern SaaS dashboards (very rounded corners,
  `shadow-sm`, indigo/teal palette) than to pure Android Material.
  `src/components/ui/`: shadcn-like primitives (`button`, `dialog`, `sheet`, `sidebar`,
  `switch`, `tooltip`) + `StateViews.tsx` already standardizes `Skeleton`/`EmptyState`/`ErrorState`
  reused across the dashboard's screens.
- **Icons**: `lucide-react` (the same package used throughout the sidebar/header).
- **Typography**: Inter (body), Outfit (`font-heading`, titles), JetBrains Mono (code/shortcuts
  such as `Ctrl K`).
- The current screen (`Monitor.tsx` + `CardServicoStatus.tsx` + `useRecursos.ts`) already implements:
  a tonal header card with a "Refresh" button, a grid of per-service cards (green/red indicator +
  metrics list), a single page-level error state (NFR02), and auto-refresh every 10s (FR07).
  This document proposes visual refinements on top of that base, not a rebuild.

## Visual References Found

| Reference | Popularity metric | Why it's relevant |
|---|---|---|
| [Uptime Kuma](https://github.com/louislam/uptime-kuma) | 90.1k stars on GitHub | Dominant reference for self-hosted service status dashboards; per-service card grid with color indicator, the same "N services, 1 down" aggregation the Monitor screen needs to communicate. |
| [Vercel Dashboard / Geist Design System](https://vercel.com/blog/dashboard-redesign) | Official design system of one of the most-used dev platforms on the market (de facto reference for technical dashboards) | Shows that, for a technical audience, color should be reserved for status (green/red/amber) — the rest of the UI is neutral (gray scale), avoiding visual "noise" competing with the data. |
| [Railway Observability Dashboard](https://docs.railway.com/observability) | PaaS platform popular among devs (recurring comparative reference alongside Fly.io/Vercel in market benchmarks) | Metric cards with a lightweight chart/visual indicator (not just a number) for CPU/memory/network — reinforces using a progress bar or sparkline instead of a bare number. |
| [Grafana](https://grafana.com) | Most widely used observability tool on the market (comparison baseline for every metrics dashboard) | "Single stat" panel with semantic background/border color based on threshold — inspires using color on the *whole card*, not just the indicator, when a service is unavailable. |
| [Nielsen Norman Group — Dashboard Design](https://www.nngroup.com/articles/dashboard-design/) | Evidence-based UX research (academic reference, not aesthetic) | Grounds the hierarchy: overall status > exceptions > details — the screen should make "how many services are healthy" obvious before requiring card-by-card reading. |

## Trends Identified

1. **Color reserved for meaning**: in popular technical dashboards (Vercel, Grafana), the UI is
   predominantly neutral/monochrome — color is used *only* for status (green/red/amber),
   never decoratively. The current screen already partially follows this (green/red dot); it can
   be reinforced by extending the color to the card's subtle border/background when unavailable.
2. **Aggregate summary at the top**: status dashboards (Uptime Kuma, status pages) always show
   a summary first ("7/8 operational") before the detailed grid — satisfies the Obvious Starting
   Point principle and NN/g's research on the "overview → exception → detail" hierarchy.
3. **Metrics as a bar/visual indicator, not just a number**: Railway and Grafana use a progress
   bar or sparkline for CPU/memory, letting the user scan "is it high?" without doing mental
   math — faster than reading "0.25 cores" in isolation.
4. **Skeleton loading instead of "Loading…" text**: `Skeleton` already exists in
   `StateViews.tsx`, used by other dashboard screens — the Monitor screen should reuse it
   (consistency, Consistent Logic) instead of the current plain text paragraph.
5. **"Last updated" timestamp**: dashboards with auto-refresh (Grafana, Vercel, Railway)
   always show "updated Xs ago" near the refresh button — communicates that auto-refresh
   (FR07) is actually working, without the user having to guess.
