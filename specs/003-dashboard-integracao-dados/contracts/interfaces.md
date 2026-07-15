# Interfaces: Dashboard — Integração de Dados e Funcionalidade

Contratos de código (TypeScript) introduzidos no Player. Todos residem em `player/src/`.

---

## `EstadoDados<T>` (RF06)

Máquina de estados de UI compartilhada por telas que carregam dados.

```ts
type EstadoDados<T> =
  | { fase: "carregando" }
  | { fase: "pronto"; dados: T }
  | { fase: "vazio" }
  | { fase: "erro"; mensagem: string };
```

**Implementações esperadas**: `useSistemas`, `useMetricas`, e os componentes de
`StateViews` que renderizam cada fase.

---

## `useSistemas` (RF02, RF06, RNF05)

Fonte única de listagem/criação de sistemas; consumido por `Projects` e `SeletorSistemas`.

```ts
interface UseSistemas {
  estado: EstadoDados<Sistema[]>;
  recarregar(): void;
  criar(nome: string): Promise<Sistema>;   // repropaga ApiError (ex.: 403 → RN10)
}

function useSistemas(client: ApiClient): UseSistemas;
```

**Implementações esperadas**: `player/src/systems/useSistemas.ts`.

---

## `Metricas` / `useMetricas` (RF01)

```ts
interface Metricas {
  sistemas_total: number;
  sistemas_ativos?: number;    // requer payload enriquecido (Fase 2)
  sistemas_rascunho?: number;  // requer payload enriquecido (Fase 2)
}

function useMetricas(client: ApiClient): EstadoDados<Metricas>;
```

**Implementações esperadas**: `player/src/dashboard/useMetricas.ts`.

---

## `UsuarioAutenticado` / `lerClaims` (RF03, RNF06)

```ts
interface UsuarioAutenticado {
  nome?: string;
  email?: string;
  iniciais: string;   // derivadas; fallback "?"
}

/** Lê claims do payload do JWT apenas para exibição (não valida assinatura). */
function lerClaims(token: string): { nome?: string; email?: string } | null;

/** Monta o objeto de exibição a partir do token persistido. */
function usuarioDe(token: string): UsuarioAutenticado;
```

**Implementações esperadas**: `player/src/auth/jwt.ts`.

---

## `AppContext` (RF01, RF02, RF03)

Provê `ApiClient` e usuário à árvore do dashboard, evitando prop drilling.

```ts
interface AppContextValue {
  client: ApiClient;
  usuario: UsuarioAutenticado;
}

const AppContext: React.Context<AppContextValue | null>;
function useApp(): AppContextValue;   // lança se usado fora do provider
```

**Implementações esperadas**: `player/src/app/AppContext.tsx`.

---

## `ThemeContext` / `ThemeProvider` (RF05, RNF04)

```ts
type Tema = "claro" | "escuro";

interface ThemeContextValue {
  tema: Tema;
  alternarTema(): void;
  definirTema(t: Tema): void;
}

function useTheme(): ThemeContextValue;

/** Aplica o tema salvo em localStorage à classe `dark` do <html>, síncrono no boot. */
function initTheme(): void;
```

**Implementações esperadas**: `player/src/theme/ThemeProvider.tsx`,
`player/src/theme/initTheme.ts`. Chave de persistência: `mach_theme`.

---

## `StateViews` (RF06, RNF03)

Componentes M3 reutilizáveis para as fases de `EstadoDados`.

```tsx
function Skeleton(props: { linhas?: number; className?: string }): JSX.Element;      // aria-busy
function EmptyState(props: { titulo: string; descricao?: string; acao?: React.ReactNode }): JSX.Element;
function ErrorState(props: { mensagem: string; onRepetir(): void }): JSX.Element;    // role="alert"
```

**Implementações esperadas**: `player/src/components/ui/StateViews.tsx`.
