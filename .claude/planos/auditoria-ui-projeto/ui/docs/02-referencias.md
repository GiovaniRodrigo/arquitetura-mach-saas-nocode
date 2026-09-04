# Popular References

| Reference | URL | Popularity | Applicability |
|---|---|---|---|
| SaaSUI | https://www.saasui.design/ | Dedicated catalog of real screenshots from Notion, Linear, Figma, Stripe (B2B SaaS market-standard reference) | Confirms "settings as cards stacked by section" (already used in `Configuracao.tsx`) as the prevailing pattern — the change recommended here is about internal consistency, not page architecture. |
| shadcn/ui — data table | https://ui.shadcn.com (`tasks` example) | Reference design system from the very React/Tailwind ecosystem the project already follows in `components/ui/` | Closest model to the current stack for evolving the systems table (`ClienteSistemas.tsx`) — `uppercase text-xs text-muted-foreground` header, `divide-y divide-border`, already in use, with a hover row as the natural next step. |
| Stripe Dashboard (Settings/Billing) | https://stripe.com/dashboard | One of the most cited references in B2B dashboard comparisons on the market | Inline confirmation pattern next to the button ("Password updated.") instead of a toast — already what `Perfil.tsx`/`SegurancaForm.tsx` do; reinforces keeping this pattern when formalizing the `--success` token. |
| Linear — Settings | https://linear.app | One of the most cited dev-tool products as a reference for minimalist/consistent UI in 2025-2026 | Reinforces the practice of reserving vivid color only for state (success/error/alert), never decorative — the same lesson already applied in the Monitor audit, now generalized. |
