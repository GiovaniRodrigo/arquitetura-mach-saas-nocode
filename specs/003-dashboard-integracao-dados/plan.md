# Plano de Implementação: Dashboard — Integração de Dados e Funcionalidade

A estratégia é evoluir o dashboard mockado em incrementos verticais, começando pela
integração de dados de maior valor (Projects/Overview) e reaproveitando o que já está
validado em `SeletorSistemas.tsx`. Extrai-se a lógica de listagem/criação de sistemas
para um hook compartilhado (`useSistemas`), injeta-se o `ApiClient`/identidade via
contexto React (hoje o `client` só existe em `App.tsx`), adiciona-se um `ThemeContext`
para o dark mode e padronizam-se os estados de UI (loading/empty/erro) em componentes
reutilizáveis. Recursos de Fase 2 (tenant, Cmd+K, presença, DLQ) ficam isolados atrás
de flags/campos opcionais para não bloquear a Fase 1.

---

## 1. Arquivos a Criar/Editar

### 1.1. Contexto de aplicação (client + usuário)

* **`player/src/app/AppContext.tsx`** (novo): provê `ApiClient` e o usuário autenticado
  (derivado do JWT) para toda a árvore do dashboard, evitando prop drilling. (RF01, RF02, RF03)
* **`player/src/auth/jwt.ts`** (novo): decodifica o payload do JWT (base64url, sem
  validar assinatura — só leitura de claims) para extrair nome/e-mail/iniciais. (RF03, RNF06)
* **`player/src/App.tsx`** (editar): envolver as rotas `/dashboard` com o `AppContext`
  provider; remover o `<nav>` genérico duplicado (linhas 67–75) que coexiste com a
  sidebar. (RF03, RF-nav)

### 1.2. Hook e serviço de dados

* **`player/src/systems/useSistemas.ts`** (novo): hook que encapsula
  `listarSistemas()`/`criarSistema()` com estados `carregando | pronto | vazio | erro`
  e ação `recarregar()`. Fonte única de verdade para `Projects` e `SeletorSistemas`. (RF02, RF06, RNF05)
* **`player/src/systems/SeletorSistemas.tsx`** (editar): refatorar para consumir
  `useSistemas`, eliminando a duplicação de estados. (RNF05)
* **`player/src/dashboard/useMetricas.ts`** (novo): hook que deriva as métricas do
  Overview a partir dos sistemas (e, quando disponível, do endpoint de métricas). (RF01)

### 1.3. Tema (dark mode)

* **`player/src/theme/ThemeProvider.tsx`** (novo): `ThemeContext` + provider que lê/grava
  `mach_theme` em `localStorage` e alterna a classe `dark` no `<html>`. (RF05, RNF04)
* **`player/src/theme/initTheme.ts`** (novo): script síncrono aplicado no boot (importado
  cedo em `main.tsx`) para evitar flash de tema. (RNF04)
* **`player/src/main.tsx`** (editar): chamar `initTheme()` antes do render e envolver
  a App no `ThemeProvider`. (RF05, RNF04)

### 1.4. Componentes de estado de UI

* **`player/src/components/ui/StateViews.tsx`** (novo): `Skeleton`, `EmptyState` e
  `ErrorState` (com botão repetir) reutilizáveis, no estilo M3. (RF06, RNF03)

### 1.5. Telas do dashboard

* **`player/src/pages/Dashboard/Overview.tsx`** (editar): métricas reais via
  `useMetricas`; "Get Started" e FAB acionam criação; estados loading/empty/erro. (RF01, RF04, RF06)
* **`player/src/pages/Dashboard/Projects.tsx`** (editar): grade de sistemas via
  `useSistemas`; card "Criar novo projeto" e "Abrir projeto" funcionais; status/versão
  por card (RF07); filtros/visualização (RF12, Fase 2). (RF02, RF04, RF06, RF07)
* **`player/src/pages/Dashboard/Settings.tsx`** (editar): "Alternar Tema" ligado ao
  `ThemeContext`; "Editar Perfil" navega para placeholder. (RF05)
* **`player/src/layout/DashboardLayout.tsx`** (editar): nome/avatar reais do usuário via
  `AppContext`; transformar o avatar inerte (C7) em menu do usuário (perfil/config/sair,
  reusando `encerrarSessao`); (Fase 2) seletor de tenant, notificações, Cmd+K. (RF03, RF14, RF11, RF13)

> **Cobertura de controles (RF04):** as edições de `Overview`, `Projects`, `Settings` e
> `DashboardLayout` acima devem, em conjunto, dar handler real a **todos** os controles
> C1–C7 do Inventário (spec §2.1) e preservar C8–C10. Os handlers de criação/abertura de
> sistema reutilizam `criarSistema`/`abrirSistema` (via `useSistemas`), sem duplicar lógica.

### 1.6. Fase 2 (isolada)

* **`player/src/dashboard/CommandPalette.tsx`** (novo, Fase 2): busca Cmd+K. (RF10)
* **`player/src/dashboard/PresencaColaboradores.tsx`** (novo, Fase 2): avatares empilhados
  via `collab/phoenixSocket.ts`. (RF08)
* **`player/src/dashboard/TenantSwitcher.tsx`** (novo, Fase 2): seletor de tenant. (RF11)

### 1.7. Testes

* **`player/src/systems/useSistemas.test.ts`** (novo): estados via `fetch` mockado. (RF02, RF06)
* **`player/src/theme/ThemeProvider.test.tsx`** (novo): toggle + persistência. (RF05)
* **`player/src/auth/jwt.test.ts`** (novo): extração de claims. (RF03)
* **`player/src/pages/Dashboard/*.test.tsx`** (editar): substituir asserts de texto
  estático por comportamento (estados, ações, dados). (RF01, RF02, RF04, RF06)

---

## 2. Estratégia Técnica

### 2.1. Fonte única de dados de sistemas (RNF05)

Hoje `SeletorSistemas` implementa manualmente `listarSistemas` + skeleton + empty + erro
+ retry. `Projects` recriaria tudo isso como mock. Em vez disso, extrai-se um hook:

```ts
// useSistemas.ts (esboço)
type Estado =
  | { fase: "carregando" }
  | { fase: "pronto"; sistemas: Sistema[] }
  | { fase: "vazio" }
  | { fase: "erro"; mensagem: string };

export function useSistemas(client: ApiClient) {
  const [estado, setEstado] = useState<Estado>({ fase: "carregando" });
  const [tentativa, setTentativa] = useState(0);
  useEffect(() => { /* listarSistemas → pronto | vazio | erro */ }, [client, tentativa]);
  return { estado, recarregar: () => setTentativa(t => t + 1) };
}
```

`SeletorSistemas` e `Projects` passam a consumir o mesmo hook e os mesmos componentes
de estado (`StateViews`).

### 2.2. Identidade via claims do JWT (RF03/RNF06)

O JWT já vive em `localStorage` (`auth/session.ts`). Um decodificador de payload lê
`name`/`email`/`sub` **apenas para exibição** — a autorização continua no Gateway. Não
se valida assinatura no cliente.

```ts
// jwt.ts (esboço)
export function lerClaims(token: string): { nome?: string; email?: string } | null {
  const [, payload] = token.split(".");
  if (!payload) return null;
  try { return JSON.parse(atob(payload.replace(/-/g,'+').replace(/_/g,'/'))); }
  catch { return null; }
}
```

### 2.3. Tema sem flash (RF05/RNF04)

`initTheme()` roda de forma síncrona no boot e aplica a classe `dark` no `<html>` a
partir do `localStorage` antes do primeiro paint; o `ThemeProvider` expõe o toggle para
o Settings. O Tailwind já suporta `dark:` via classe.

### 2.4. Métricas do Overview (RF01)

Enquanto não existir endpoint dedicado de métricas, `useMetricas` deriva contadores a
partir de `listarSistemas()` (total, e — quando o `Sistema` for enriquecido — ativos vs.
rascunhos). O contrato de métricas fica especificado em `contracts/api.md` para quando
o backend expuser.

### 2.5. Faseamento

- **Fase 1 (Alta)**: RF01–RF06 — integração, ações, tema, estados, identidade.
- **Fase 2 (Média/Baixa)**: RF07–RF13 — status/versão por card, presença, DLQ, Cmd+K,
  tenant switcher, filtros, notificações. Dependem de enriquecimento de contrato/backend.

---

## 3. Dependências e Pré-requisitos

- [ ] Gateway acessível com um JWT válido (fluxo OAuth de `auth/session.ts`) ou
      `VITE_BYPASS_AUTH=true` para desenvolvimento.
- [ ] Endpoint `GET /api/v1/sistemas` operante (já consumido por `SeletorSistemas`).
- [ ] (Fase 2 — RF07/RF09) Enriquecimento do payload de `Sistema` no Gateway com
      `status`, `versao_ativa` e métricas de DLQ — ver `data-model.md` e `contracts/api.md`.
- [ ] (Fase 2 — RF08) Tópicos de presença por sistema disponíveis no servidor de
      colaboração (Phoenix).

---

## 4. Riscos e Pontos de Atenção

| Risco | Impacto | Mitigação |
|-------|---------|-----------|
| `Sistema` atual só tem `id`/`nome`; RF07/RF09 exigem campos inexistentes. | Alto | Faseamento: Fase 1 não depende deles; Fase 2 tratada como opcional/degradável até o contrato ser estendido. |
| Não há endpoint de métricas agregadas para o Overview. | Médio | Derivar métricas de `listarSistemas()` no `useMetricas`; especificar contrato futuro. |
| Refatorar `SeletorSistemas` pode regredir comportamento já validado. | Médio | Extrair hook com testes cobrindo os quatro estados antes de trocar a tela. |
| Flash de tema (FOUC) se o provider aplicar o tema só após o mount. | Médio | `initTheme()` síncrono no boot, antes do `createRoot().render`. |
| Ler claims do JWT no cliente pode ser confundido com autorização. | Alto (segurança) | Claims usados só para exibição; autorização permanece no Gateway (RNF06). |
| Duplicação de navegação (`<nav>` de `App.tsx` + sidebar). | Baixo | Remover o `<nav>` genérico ao envolver o dashboard no layout. |
