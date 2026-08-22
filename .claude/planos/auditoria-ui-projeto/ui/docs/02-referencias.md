# Referências Populares

| Referência | URL | Popularidade | Aplicabilidade |
|---|---|---|---|
| SaaSUI | https://www.saasui.design/ | Catálogo dedicado de screenshots reais de Notion, Linear, Figma, Stripe (referência de padrão de mercado B2B SaaS) | Confirma "settings em cards empilhados por seção" (já usado em `Configuracao.tsx`) como padrão vigente — a mudança recomendada aqui é de consistência interna, não de arquitetura de página. |
| shadcn/ui — data table | https://ui.shadcn.com (exemplo `tasks`) | Design system de referência do próprio ecossistema React/Tailwind que o projeto já segue em `components/ui/` | Modelo mais próximo à stack atual para evoluir a tabela de sistemas (`ClienteSistemas.tsx`) — cabeçalho `uppercase text-xs text-muted-foreground`, `divide-y divide-border`, já usados, e linha de hover como próximo passo natural. |
| Stripe Dashboard (Settings/Billing) | https://stripe.com/dashboard | Uma das referências mais citadas em comparações de dashboards B2B do mercado | Padrão de confirmação inline junto ao botão ("Senha atualizada.") em vez de toast — já é o que `Perfil.tsx`/`SegurancaForm.tsx` fazem; reforça manter esse padrão ao formalizar o token `--success`. |
| Linear — Settings | https://linear.app | Um dos produtos dev-tool mais citados como referência de UI minimalista/consistente em 2025-2026 | Reforça a prática de reservar cor viva só para estado (sucesso/erro/alerta), nunca decorativo — mesma lição já aplicada na auditoria do Monitor, agora generalizada. |
