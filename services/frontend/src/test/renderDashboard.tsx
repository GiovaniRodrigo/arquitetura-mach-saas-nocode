// Helper de teste: renderiza uma tela do dashboard com os provedores necessários
// (Router + AppProvider + ThemeProvider) e um ApiClient falso injetável.

import type { ReactElement } from 'react';
import { render } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
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
  podeCriarSistema: true,
};

/** Usuário cliente-final (sem papel dono/parceiro) — não vê CTAs de criação (RN10). */
export const usuarioClienteFake: UsuarioAutenticado = {
  nome: 'Bruno Cliente',
  email: 'bruno@x.com',
  iniciais: 'BC',
  podeCriarSistema: false,
};

export function renderDashboard(
  ui: ReactElement,
  opts: { client?: ApiClient; usuario?: UsuarioAutenticado; rota?: string; path?: string } = {},
) {
  const client =
    opts.client ??
    fakeClient({
      listarSistemas: async () => [],
      listarUltimosAcessos: async () => [],
      listarFeedback: async () => [],
      resumoFinanceiro: async () => ({ receita_total_centavos: 0, moeda: 'BRL', competencia: '' }),
      acessosPorMes: async () => [],
      receitaPorMes: async () => ({ pontos: [], moeda: 'BRL' }),
    });
  const usuario = opts.usuario ?? usuarioFake;
  const rota = opts.rota ?? '/dashboard';
  // `path` habilita rotas com parâmetros (:tenantId, :sistemaId) via useParams();
  // sem ele, a tela é montada direto na `rota` (comportamento pré-existente).
  return render(
    <MemoryRouter initialEntries={[rota]}>
      <ThemeProvider>
        <AppProvider client={client} usuario={usuario}>
          {opts.path ? <Routes><Route path={opts.path} element={ui} /></Routes> : ui}
        </AppProvider>
      </ThemeProvider>
    </MemoryRouter>,
  );
}
