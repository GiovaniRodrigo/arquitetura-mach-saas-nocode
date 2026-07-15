# Contexto do Projeto

## Domínio
**MACH V4 — Plataforma Low-Code / No-Code.** O software permite que utilizadores construam as
suas próprias aplicações digitais por uma interface visual (CRUD de UI, regras de negócio,
publicação instantânea, multi-tenancy hierárquico, colaboração em tempo real). A parte
auditada aqui é o **Headless Player** — a SPA que renderiza a versão *ativa* de um sistema
construído na plataforma, servida em `https://gfcode.com.br/ui/`.

## Público-Alvo
Dois perfis, ambos atendidos pela mesma casca:
- **Builders / operadores** (semi-técnicos): donos e parceiros que criam e publicam sistemas.
- **Clientes finais** (leigos): usuários dos sistemas gerados, que apenas consomem as telas
  (nav + formulários + submissão).

Contexto de uso: navegador desktop e mobile; sessão iniciada por **login social** (Google/GitHub).
Como a tela renderizada é a "cara" do produto que o cliente do builder entrega ao *seu* cliente,
a percepção de qualidade visual da casca (login, estados, chrome) impacta diretamente a
confiança no produto.

## Stack Frontend (observada no código)
- **Vite + React 18 + TypeScript**, `react-router-dom`. Base servida sob `/ui/`
  (`player/vite.config.ts` → `base: "/ui/"`).
- **Sem framework de CSS** (não há Tailwind, MUI, etc.). Todo estilo é *inline* em `Login.tsx`;
  as demais telas (loading/empty/erro e `CompositeRenderer`) **não têm estilo próprio** — saem
  com aparência default do navegador.
- Entrada: `player/src/main.tsx` (decide Login vs App conforme token), `App.tsx` (rotas
  dinâmicas + render de ecrãs), `CompositeRenderer.tsx` (árvore Composite → DOM).

## Validação técnica (feita no navegador — resumo)
`https://gfcode.com.br/ui/`: `/ui/` 200, `/ui/assets/index-*.js` 200 (base path correto), SPA
monta, **sem erros de console**, tela de Login renderiza, botões apontam para
`/auth/{google,github}` (proxy Nginx ok). **Funcionalmente correto** — esta auditoria trata da
qualidade de *design*.

## Referências Visuais Encontradas (web)
| Referência | Por que é relevante | Popularidade |
|---|---|---|
| **Appsmith** (builder no-code, tela de login/app) | Mesmo domínio (builder low-code); templates de login | ~30.000+ stars no GitHub; Discord 5.000+ membros |
| **Budibase** | Builder no-code com auto-CRUD e telas de auth | ~300.000 times/equipes usando a plataforma |
| **ToolJet** | Builder low-code AI-native, ecossistema de UI | 500+ contribuidores no GitHub |
| **Bubble** | Builder no-code customer-facing de referência | Milhões de apps criados; líder de mercado no-code |
| **Retool** | Biblioteca de 100+ componentes configuráveis | Usado por milhares de empresas (padrão de internal tools) |
| **Material Design 3** (Google) | Guideline oficial de botões/estados/cores | Design system oficial do Android/Google |
| **Sign in with Google — Best Practices** | Guideline oficial do botão social | Documentação oficial Google Identity |

## Tendências Identificadas (aplicáveis ao Login/casca)
1. **Login social minimalista com hierarquia clara**: 2–3 métodos no máximo, destacando o
   provedor primário ("Continue com…") — evita sobrecarga cognitiva (Authgear/Lollypop 2025).
2. **Botões sociais seguindo guideline de cada IDP**: logo colorido do Google (G de 4 cores),
   marca do GitHub; quando o IDP não publica spec de botão, aplicar o logo + padrão de botão
   do próprio produto para consistência.
3. **Mobile-first com alvos de toque grandes** e layout responsivo — assumir login no celular.
4. **Card centrado com identidade de marca** (logo + nome + microcopy) em vez de texto solto.
5. **Estados visíveis e não-genéricos**: skeleton em vez de spinner, empty state com ação,
   erro com retry — em linha com Material 3 e NN/g.

## Fontes
- Login/Signup UX 2025 — https://www.authgear.com/post/login-signup-ux-guide/
- SaaS login page design — https://lollypop.design/blog/2025/october/saas-login-page-design/
- Sign in with Google best practices — https://developers.google.com/identity/siwg/best-practices
- Designing with social login buttons — https://medium.com/@sabarivasan/designing-with-social-login-buttons-the-right-way-a-deep-dive-into-idp-guidelines-618742589c85
- Appsmith vs Budibase vs ToolJet — https://blog.tooljet.com/appsmith-vs-budibase-vs-tooljet/
- Open source low-code platforms — https://www.appsmith.com/blog/open-source-low-code-platforms
