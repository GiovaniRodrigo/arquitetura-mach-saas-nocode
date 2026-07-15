// Contexto de aplicação: provê o ApiClient e a identidade do usuário (derivada do
// JWT) para toda a árvore do dashboard, evitando prop drilling (RF01, RF02, RF03).

import { createContext, useContext, type ReactNode } from 'react';
import type { ApiClient } from '../api/client';
import type { UsuarioAutenticado } from '../auth/jwt';

export interface AppContextValue {
  client: ApiClient;
  usuario: UsuarioAutenticado;
}

const AppContext = createContext<AppContextValue | null>(null);

export function AppProvider({
  client,
  usuario,
  children,
}: AppContextValue & { children: ReactNode }) {
  return <AppContext.Provider value={{ client, usuario }}>{children}</AppContext.Provider>;
}

/** Acessa client + usuário; lança se usado fora do AppProvider. */
export function useApp(): AppContextValue {
  const ctx = useContext(AppContext);
  if (!ctx) throw new Error('useApp deve ser usado dentro de <AppProvider>');
  return ctx;
}
