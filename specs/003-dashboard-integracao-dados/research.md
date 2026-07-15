# Pesquisa: Dashboard — Integração de Dados e Funcionalidade

---

## 1. Padrões Existentes no Projeto

O projeto já contém quase tudo o que a Fase 1 precisa — o dashboard apenas não reutiliza.
A pesquisa confirma que a maior parte do trabalho é **reuso e ligação**, não construção
do zero.

| Arquivo/Padrão | Localização | Relevância |
|----------------|-------------|-----------|
| `SeletorSistemas.tsx` | `player/src/systems/` | Referência canônica: `listarSistemas()`/`criarSistema()` com skeleton (`animate-pulse`), empty state, erro + "Tentar novamente", e `abrirSistema()` via query string. Fonte a extrair para `useSistemas`. |
| `ApiClient` | `player/src/api/client.ts` | Já expõe `listarSistemas`, `criarSistema`, `versaoAtiva`, `permissoes`, `criarExportacao`; `fetch` injetável para testes. Nenhuma chamada nova de rede é necessária na Fase 1. |
| `types.ts` | `player/src/api/types.ts` | `Sistema` só tem `id`/`nome` → confirma que RF07/RF09 exigem extensão de contrato (Fase 2). |
| `session.ts` | `player/src/auth/session.ts` | JWT persistido em `localStorage` (`mach_token`); base para ler claims de identidade (RF03). |
| Componentes M3 | `player/src/components/m3/` | `TonalCard`, `ElevatedCard`, `FabButton`, `NavPill` — reutilizar para manter RNF01. Nota: usam cores fixas (`bg-white`, `bg-slate-100`, `bg-blue-200`) que **não reagem a `dark:`** — precisarão de tokens temáticos para o dark mode (RF05). |
| `sidebar.tsx` / `SidebarProvider` | `player/src/components/ui/` | Responsividade e toggle da sidebar já resolvidos (usa `use-mobile`); preservar (RNF02). |
| `phoenixSocket.ts` | `player/src/collab/` | Canal Phoenix existente; base para presença em tempo real (RF08, Fase 2), hoje não usado pelo dashboard. |
| Padrão de teste | `*.test.tsx` | Vitest + Testing Library + `fetch`/`matchMedia` mockados; seguir o mesmo estilo nos novos testes. |

---

## 2. Tecnologias e Bibliotecas

| Tecnologia | Versão | Uso | Já instalada? |
|------------|--------|-----|---------------|
| React + react-router-dom | 6+ | SPA e rotas do dashboard | Sim |
| Tailwind CSS | 4 (`@tailwindcss/postcss`) | Estilos M3 e `dark:` para tema | Sim |
| lucide-react | — | Ícones (Home, Folder, Settings, LogOut) | Sim |
| Vitest + Testing Library | — | Testes unitários/comportamentais | Sim |
| Playwright | — | E2E (configurado; opcional para fluxos do dashboard) | Sim |
| phoenix (`@types/phoenix`) | — | WebSocket de colaboração (Fase 2) | Sim |

Nenhuma dependência nova é necessária para a Fase 1. O dark mode usa a estratégia de
classe do Tailwind (`dark:`), sem biblioteca adicional.

---

## 3. Referências Externas

| Referência | URL | O que resolve |
|------------|-----|--------------|
| Wireframe do dashboard | `.claude/planos/001-construtor-sistemas-mach-v4/ui/wireframes/dashboard.html` | Visão-alvo rica (tenant, Cmd+K, status/versão, presença, DLQ, filtros, empty/skeleton) — base de RF07–RF13 |
| Regras M3 | `.agents/Demandas/dashboard-m3-regras-negocio.md` | RF/RNF originais da estética M3 (hero, métricas, FAB, responsivo) |
| Docs de UI | `.claude/planos/001-.../ui/docs/04-sistema-cores-tipografia.md` | Sistema de cores/tipografia para tokens temáticos |
| Tailwind dark mode | https://tailwindcss.com/docs/dark-mode | Estratégia de classe `dark` no `<html>` (RNF04) |

---

## 4. Alternativas Consideradas

### Opção A: Reescrever `Projects` do zero com sua própria chamada à API
- **Prós**: Rápido no curtíssimo prazo.
- **Contras**: Duplica `SeletorSistemas` (viola RNF05/DRY); dois pontos de manutenção
  para estados de carregamento/erro.
- **Decisão**: Descartada.

### Opção B: Extrair `useSistemas` e reusar em `Projects` e `SeletorSistemas`
- **Prós**: Fonte única de verdade; testável isoladamente; remove duplicação.
- **Contras**: Exige refatorar `SeletorSistemas` (com cobertura de teste antes).
- **Decisão**: **Escolhida.**

### Opção C (tema): biblioteca de tema (ex.: next-themes-like)
- **Prós**: Pronto.
- **Contras**: Dependência extra; SPA simples com Tailwind não justifica.
- **Decisão**: Descartada — `ThemeContext` próprio + `initTheme()` síncrono.

### Opção D (identidade): buscar perfil do usuário em endpoint dedicado
- **Prós**: Dados sempre atuais.
- **Contras**: Endpoint pode não existir; round-trip extra só para exibir nome/avatar.
- **Decisão**: Ler claims do JWT para exibição (RF03/RNF06); endpoint de perfil fica
  como evolução futura.

### Opção E (métricas): endpoint de métricas agregadas
- **Prós**: Números precisos (tarefas pendentes, membros).
- **Contras**: Não existe hoje.
- **Decisão**: Derivar de `listarSistemas()` na Fase 1; especificar contrato para o futuro.
