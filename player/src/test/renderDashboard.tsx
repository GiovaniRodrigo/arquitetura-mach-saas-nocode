// Helper de teste: renderiza uma tela do dashboard com os provedores necessários
// (Router + AppProvider + ThemeProvider) e um ApiClient falso injetável.

import type { ReactElement } from 'react';
import { render } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import type { ApiClient } from '../api/client';
import type { UsuarioAutenticado } from '../auth/jwt';
import { AppProvider } from '../app/AppContext';
import { ThemeProvider } from '../theme/ThemeProvider';

export function fakeClient(over: Partial<ApiClient>): ApiClient {
  return over as unknown as ApiClient;
}

export const usuarioFake: UsuarioAutenticado = {
  nome: 'Ana Silva',
  email: 'ana@x.com',
  iniciais: 'AS',
};

export function renderDashboard(
  ui: ReactElement,
  opts: { client?: ApiClient; usuario?: UsuarioAutenticado; rota?: string } = {},
) {
  const client = opts.client ?? fakeClient({ listarSistemas: async () => [] });
  const usuario = opts.usuario ?? usuarioFake;
  return render(
    <MemoryRouter initialEntries={[opts.rota ?? '/dashboard']}>
      <ThemeProvider>
        <AppProvider client={client} usuario={usuario}>
          {ui}
        </AppProvider>
      </ThemeProvider>
    </MemoryRouter>,
  );
}
