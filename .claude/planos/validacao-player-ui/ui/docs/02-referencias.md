# Referências Populares (web)

Referências do **mesmo domínio** (builders no-code/low-code) e guidelines oficiais de login
social. Todas com evidência de adoção — sem blogs genéricos.

| Referência | URL | Popularidade | Aplicabilidade ao player MACH |
|---|---|---|---|
| **Appsmith** — login/app templates | https://www.appsmith.com/blog/open-source-low-code-platforms | ~30.000+ stars GitHub; Discord 5.000+ | Padrão de tela de login e de app renderizado por builder; grid consistente |
| **Budibase** — auth + auto-CRUD | https://budibase.com/blog/alternatives/appsmith-vs-budibase/ | ~300.000 equipes na plataforma | Estados de formulário/CRUD e telas de acesso para usuário final leigo |
| **ToolJet** — UI builder low-code | https://blog.tooljet.com/appsmith-vs-budibase-vs-tooljet/ | 500+ contribuidores | Biblioteca de componentes consistente; feedback de ações |
| **Bubble** — no-code customer-facing | https://blog.tooljet.com/appsmith-vs-budibase-vs-tooljet/ | Líder de mercado no-code, milhões de apps | Casca voltada ao cliente final; identidade de marca no login |
| **Retool** — 100+ componentes | https://www.appsmith.com/blog/retool-alternatives | Milhares de empresas (padrão internal tools) | Estados de tabela/form e consistência de componentes |
| **Material Design 3** — botões/estados | https://m3.material.io/ | Design system oficial Google/Android | Especificação de botão, foco/hover/disabled, cores semânticas |
| **Sign in with Google (guideline)** | https://developers.google.com/identity/siwg/best-practices | Documentação oficial Google Identity | Hierarquia e posicionamento do botão social primário |
| **Login/Signup UX Guide 2025** | https://www.authgear.com/post/login-signup-ux-guide/ | Guia de referência citado no setor | Limitar a 2–3 métodos; mobile-first; alvos grandes |
| **SaaS login page design 2025 (Lollypop)** | https://lollypop.design/blog/2025/october/saas-login-page-design/ | Estúdio de design premiado | Card centrado, microcopy, identidade de marca |

## Síntese aplicável
- **Hierarquia**: destacar 1 método primário; máximo 2–3 no total (temos exatamente 2 — bom).
- **Marca dos IDPs**: Google exige o "G" de 4 cores; GitHub usa o mark oficial. Hoje o player
  usa **só texto**, sem logo — desvio das guidelines de ambos os IDPs.
- **Card + identidade**: envolver o login num card centrado com logo/nome MACH.
- **Mobile-first**: alvos ≥44px, largura fluida com `max-width`.
- **Estados**: skeleton/loading, empty com ação, erro com retry — hoje ausentes.
