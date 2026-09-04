# Interfaces: Dashboard — Data Integration and Functionality

Code contracts (TypeScript) introduced in the Player. All live under `player/src/`.

---

## `EstadoDados<T>` (FR06)

UI state machine shared by screens that load data.

```ts
type EstadoDados<T> =
  | { fase: "carregando" }
  | { fase: "pronto"; dados: T }
  | { fase: "vazio" }
  | { fase: "erro"; mensagem: string };
```

**Expected implementations**: `useSistemas`, `useMetricas`, and the
`StateViews` components that render each phase.

---

## `useSistemas` (FR02, FR06, NFR05)

Single source of truth for listing/creating systems; consumed by `Projects` and `SeletorSistemas`.

```ts
interface UseSistemas {
  estado: EstadoDados<Sistema[]>;
  recarregar(): void;
  criar(nome: string): Promise<Sistema>;   // repropagates ApiError (e.g., 403 → BR10)
}

function useSistemas(client: ApiClient): UseSistemas;
```

**Expected implementations**: `player/src/systems/useSistemas.ts`.

---

## `Metricas` / `useMetricas` (FR01)

```ts
interface Metricas {
  sistemas_total: number;
  sistemas_ativos?: number;    // requires the enriched payload (Phase 2)
  sistemas_rascunho?: number;  // requires the enriched payload (Phase 2)
}

function useMetricas(client: ApiClient): EstadoDados<Metricas>;
```

**Expected implementations**: `player/src/dashboard/useMetricas.ts`.

---

## `UsuarioAutenticado` / `lerClaims` (FR03, NFR06)

```ts
interface UsuarioAutenticado {
  nome?: string;
  email?: string;
  iniciais: string;   // derived; fallback "?"
}

/** Reads claims from the JWT payload for display only (does not validate the signature). */
function lerClaims(token: string): { nome?: string; email?: string } | null;

/** Builds the display object from the persisted token. */
function usuarioDe(token: string): UsuarioAutenticado;
```

**Expected implementations**: `player/src/auth/jwt.ts`.

---

## `AppContext` (FR01, FR02, FR03)

Provides the `ApiClient` and the user to the dashboard tree, avoiding prop drilling.

```ts
interface AppContextValue {
  client: ApiClient;
  usuario: UsuarioAutenticado;
}

const AppContext: React.Context<AppContextValue | null>;
function useApp(): AppContextValue;   // throws if used outside the provider
```

**Expected implementations**: `player/src/app/AppContext.tsx`.

---

## `ThemeContext` / `ThemeProvider` (FR05, NFR04)

```ts
type Tema = "claro" | "escuro";

interface ThemeContextValue {
  tema: Tema;
  alternarTema(): void;
  definirTema(t: Tema): void;
}

function useTheme(): ThemeContextValue;

/** Applies the theme saved in localStorage to the <html> `dark` class, synchronously at boot. */
function initTheme(): void;
```

**Expected implementations**: `player/src/theme/ThemeProvider.tsx`,
`player/src/theme/initTheme.ts`. Persistence key: `mach_theme`.

---

## `StateViews` (FR06, NFR03)

Reusable M3 components for the `EstadoDados` phases.

```tsx
function Skeleton(props: { linhas?: number; className?: string }): JSX.Element;      // aria-busy
function EmptyState(props: { titulo: string; descricao?: string; acao?: React.ReactNode }): JSX.Element;
function ErrorState(props: { mensagem: string; onRepetir(): void }): JSX.Element;    // role="alert"
```

**Expected implementations**: `player/src/components/ui/StateViews.tsx`.
