# Project Context

## Domain

MAYS (Make Your SaaS) is a no-code builder for multi-tenant SaaS systems: the
user assembles screens via visual composition (a Figma/Webflow-style Canvas in
`pages/Dashboard/editor`), defines business rules, and versions and publishes
the resulting system. It's a **system architecture/design** tool operated by
non-programmers or low-code programmers.

This analysis's ask: add an **AI assistant specialized in system
design/architecture**, using RAG (Retrieval-Augmented Generation),
that talks with the user about the focus/description of what they're building
and returns recommendations on modeling, structure, and best practices.

## Target Audience

Account owners (owners/partners) building a client's system, inside the
authenticated Dashboard. Technical level varies — from layperson to developer —
but everyone already works with concepts like "system", "screen", "business
rule", "version". The assistant needs to be useful both for someone who
doesn't know how to name the problem ("I want a system to schedule
appointments") and for someone who already thinks in technical terms ("how do
I structure multi-tenancy here?").

## Visual References Found

See `02-referencias.md` for the full table with URLs and popularity.

- **GitHub Copilot Chat** — docked side panel, not floating over the
  content; composer fixed at the bottom of the panel.
- **Notion AI** — AI panel with a centered input and suggestion chips;
  history only appears after the first message (task-scoped assistant).
- **Intercom Messenger** — floating widget pattern anchored to a corner,
  optimized for a fast first response (customer support, not the
  right pattern for a continuous work assistant).
- **Attio / Hex** (leading 2026 dashboards) — treat AI output as a
  first-class surface (summaries, suggested actions) instead of a
  floating widget on top of the old UI.

## Trends Identified

1. **Docked panel, not a floating pop-up over the content** — the
   winning pattern in productivity products (Copilot Chat, Notion AI, Linear
   AI) is to open/close a side panel that pushes or partially overlaps the
   layout, always keeping the composer anchored at the bottom (avoids the
   most common UX bug in AI chat: a floating composer overlapping the last
   message).
2. **Persistent, global trigger**, not hidden inside a single tab — since
   the "system focus" spans several builder screens (Systems, Business
   Rules, Screens), the assistant's entry point should exist across the
   entire Dashboard, not just inside the Canvas editor.
3. **Automatic context from the current screen** — AI assistants in 2026
   products (Notion AI, Linear AI) inherit the context of what the user is
   viewing (here: the selected system) instead of requiring the user to
   re-explain everything.
4. **Task-scoped state, with entry suggestions** — avoids an empty chat
   screen; shows initial chips/suggestions ("Model multi-tenancy",
   "Review business rules") aligned with the product's domain.
5. **AI as a complementary layer, never blocking** — the panel can be
   closed at any time without losing the work on the Canvas behind it; it's
   never a modal.
