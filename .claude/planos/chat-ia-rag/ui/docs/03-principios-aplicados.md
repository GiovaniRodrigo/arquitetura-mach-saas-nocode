# Applied Principles — AI Assistant (RAG Chat)

## 1. Obvious Start
A single circular floating button (FAB), "sparkles" icon, fixed in the
bottom-right corner, at `z-50`, visible on **every** Dashboard page
(including the fixed-viewport Canvas editor). It's the only entry point — no
ambiguity about where to "talk to the AI".

## 2. Clear Reversal
The panel is a non-modal, non-blocking `Sheet` (drawer): `Esc`, a click
outside, or the FAB itself (now in a "close" state) close it at any time
without discarding the conversation history (kept in memory/sessionStorage).
No AI action applies anything automatically to the Canvas — it's only a text
recommendation.

## 3. Consistent Logic
The `Sheet` reuses the `components/ui/sheet.tsx` component already used in
the project — same overlay/animation behavior as any other drawer in the
system. The FAB uses the same `Button` (`components/ui/button.tsx`) and
`primary` palette as the rest of the UI.

## 4. Follow Conventions
The "sparkles" icon (lucide-react, already the project's icon library) —
universal semantics for "AI/assistant" (Copilot, Notion AI, Gemini use the
same glyph). Panel on the right, like Copilot Chat and most docked
assistants.

## 5. Feedback and Milestones
- "Typing"/streaming response state (line skeleton, not a generic spinner).
- Network errors become an inline chat message ("Couldn't respond right now,
  try again"), never a blocking alert.
- The initial empty message shows 3 suggestion chips (e.g., "Model
  multi-tenancy", "Review the current system's business rules", "Suggest
  a screen structure").

## 6. Proximity and Adaptation
The panel automatically inherits the **currently selected system** (name)
via `AppContext`, shown as a context "pill" at the top of the chat — the user
doesn't need to retype which system they're working on. On screens with no
selected system, the pill doesn't appear and the assistant responds in
generic mode. Responsive: on narrow screens the `Sheet` takes up the full
width.

## 7. Interface Is Content
No superfluous decoration: the panel header only has the title "Design
Assistant" + context pill + close button. Messages in simple bubbles
(user on the right, assistant on the left), with no unnecessary avatars.

## 8. General Visual Design Principles
- **Obvious subject**: "Design Assistant" title + sparkles icon at the top of
  the panel.
- **Form reinforces meaning**: assistant bubbles use `bg-secondary`
  (neutral), user bubbles use `bg-primary/10` — no alert colors, since
  there's no error/success state in the content itself.
- **Familiar metaphor**: 1:1 chat layout (stacked messages + composer
  fixed at the bottom) — no user needs to learn a new pattern.

## 9. Design Decision Matrix

| Decision | Obvious Start | Clear Reversal | Consistency | Convention | Feedback | Proximity | Content > Deco. |
|---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| Fixed global FAB (bottom-right corner) | ✓ | ✓ | ✓ | ✓ | — | — | ✓ |
| Sheet docked to the right (not modal) | ✓ | ✓ | ✓ | ✓ | ✓ | — | ✓ |
| Composer fixed at the bottom of the panel | ✓ | — | ✓ | ✓ | ✓ | ✓ | ✓ |
| Context pill (current system) | — | — | ✓ | — | ✓ | ✓ | ✓ |
| Suggestion chips in empty state | ✓ | — | — | ✓ | ✓ | — | ✓ |
| Streaming skeleton instead of spinner | — | — | ✓ | ✓ | ✓ | — | ✓ |
