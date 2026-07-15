# Tarefas: Dashboard — Integração de Dados e Funcionalidade

Ordenadas por dependência de execução. Fase 1 (RF01–RF06) entrega o dashboard funcional;
Fase 2 (RF07–RF13) depende de enriquecimento de contrato/backend. Cada tarefa é atômica
(≤ 1 dia) e referencia os arquivos afetados.

## Fase 1 — Integração, ações, tema, estados, identidade

- [ ] 1. Criar componentes de estado de UI reutilizáveis `Skeleton`, `EmptyState`, `ErrorState` no estilo M3, com `aria-busy`/`role="alert"` (RF06, RNF03) (`player/src/components/ui/StateViews.tsx`)
- [ ] 2. Escrever teste do hook de sistemas cobrindo os quatro estados (carregando/pronto/vazio/erro) com `fetch` mockado (RF02, RF06) (`player/src/systems/useSistemas.test.ts`)
- [ ] 3. Implementar o hook `useSistemas(client)` encapsulando `listarSistemas`/`criarSistema` + estados + `recarregar()` (RF02, RF06, RNF05) (`player/src/systems/useSistemas.ts`)
- [ ] 4. Refatorar `SeletorSistemas` para consumir `useSistemas` e `StateViews`, removendo a duplicação de estados (RNF05) (`player/src/systems/SeletorSistemas.tsx`)
- [ ] 5. Escrever teste do decodificador de JWT (extração de `name`/`email`/iniciais; token inválido → `null`) (RF03) (`player/src/auth/jwt.test.ts`)
- [ ] 6. Implementar `lerClaims(token)` para extrair identidade do JWT (apenas leitura, sem validar assinatura) (RF03, RNF06) (`player/src/auth/jwt.ts`)
- [ ] 7. Criar `AppContext` provendo `ApiClient` e `UsuarioAutenticado` à árvore do dashboard (RF01, RF02, RF03) (`player/src/app/AppContext.tsx`)
- [ ] 8. Envolver as rotas `/dashboard` no `AppContext` e remover o `<nav>` genérico duplicado das linhas 67–75 (RF03, navegação) (`player/src/App.tsx`)
- [ ] 9. Escrever teste do `ThemeProvider` (toggle alterna classe `dark`; persiste e relê `mach_theme`) (RF05, RNF04) (`player/src/theme/ThemeProvider.test.tsx`)
- [ ] 10. Implementar `ThemeProvider`/`ThemeContext` (estado, toggle, `localStorage`, classe `dark` no `<html>`) (RF05) (`player/src/theme/ThemeProvider.tsx`)
- [ ] 11. Implementar `initTheme()` síncrono e chamá-lo no boot antes do render; envolver App no `ThemeProvider` (RNF04) (`player/src/theme/initTheme.ts`, `player/src/main.tsx`)
- [ ] 12. Ajustar componentes M3 (`TonalCard`, `ElevatedCard`, `FabButton`) e o `DashboardLayout` para usar tokens temáticos que respondam a `dark:` (RF05, RNF01) (`player/src/components/m3/*.tsx`, `player/src/layout/DashboardLayout.tsx`)
- [ ] 13. Exibir nome/iniciais reais do usuário no cabeçalho via `AppContext`, substituindo "Welcome, User"/"U" (RF03) (`player/src/layout/DashboardLayout.tsx`)
- [ ] 14. Transformar o avatar inerte (C7) em menu do usuário — perfil/configurações/sair, reutilizando `encerrarSessao` (C8) (RF14) (`player/src/layout/DashboardLayout.tsx`)
- [ ] 15. Reescrever `Projects` para renderizar a grade de sistemas via `useSistemas` + `StateViews`; "Abrir projeto" (C4) navega e card "Criar novo projeto" (C3) inicia criação (RF02, RF04, RF06) (`player/src/pages/Dashboard/Projects.tsx`)
- [ ] 16. Implementar `useMetricas` derivando contadores de `listarSistemas()` (RF01) (`player/src/dashboard/useMetricas.ts`)
- [ ] 17. Reescrever `Overview` com métricas reais via `useMetricas`; "Get Started" (C1) e FAB "Create" (C2) acionam criação de sistema; estados loading/empty/erro; remover `alert()` (RF01, RF04, RF06) (`player/src/pages/Dashboard/Overview.tsx`)
- [ ] 18. Ligar `Settings`: "Alternar Tema" (C6) ao `ThemeContext` e "Editar Perfil" (C5) a uma rota/placeholder (RF04, RF05) (`player/src/pages/Dashboard/Settings.tsx`)
- [ ] 19. Auditar o Inventário de Controles (spec §2.1): confirmar que C1–C7 têm handler real (sem `alert()`/`<div>` clicável sem handler) e que C8–C10 seguem funcionais (RF04) (`player/src/pages/Dashboard/*.tsx`, `player/src/layout/DashboardLayout.tsx`)
- [ ] 20. Atualizar os testes das telas do dashboard para comportamento — cada controle C1–C7 dispara sua ação (spy/mock) e os estados de dados são cobertos — em vez de texto estático (RF01, RF02, RF04, RF06, RF14) (`player/src/pages/Dashboard/Overview.test.tsx`, `Projects.test.tsx`, `Settings.test.tsx`, `player/src/layout/DashboardLayout.test.tsx`)

## Fase 2 — Recursos avançados do wireframe (dependem de contrato/backend)

- [ ] 21. Estender o consumo de `Sistema` para exibir status (Publicado/Rascunho/Falha) e versão ativa nos cards, degradando quando ausente (RF07, RN04) (`player/src/pages/Dashboard/Projects.tsx`, `player/src/api/types.ts`)
- [ ] 22. Adicionar filtros por status (Todos/Publicados/Rascunhos) e alternância grade/lista em `Projects` (RF12) (`player/src/pages/Dashboard/Projects.tsx`)
- [ ] 23. Implementar presença de colaboradores por sistema (avatares empilhados) via `phoenixSocket` (RF08) (`player/src/dashboard/PresencaColaboradores.tsx`, `player/src/collab/phoenixSocket.ts`)
- [ ] 24. Exibir alerta de falha/contagem de DLQ por sistema (RF09, RN09) (`player/src/pages/Dashboard/Projects.tsx`)
- [ ] 25. Implementar command palette Cmd/Ctrl+K para busca de sistemas e ações (RF10) (`player/src/dashboard/CommandPalette.tsx`, `player/src/layout/DashboardLayout.tsx`)
- [ ] 26. Implementar seletor de tenant hierárquico e indicador de notificações na top bar (RF11, RF13, RN01) (`player/src/dashboard/TenantSwitcher.tsx`, `player/src/layout/DashboardLayout.tsx`)

## Encerramento

- [ ] 27. Rodar a suíte completa e o build: `npm run test` e `npm run build` (`tsc --noEmit` + `vite build`) devem passar sem erros (`player/`)
