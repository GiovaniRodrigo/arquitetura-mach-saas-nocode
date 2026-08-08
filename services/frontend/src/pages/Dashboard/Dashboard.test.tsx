import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { Dashboard } from './Dashboard';
import { renderDashboard, fakeClient, usuarioClienteFake, usuarioFake } from '../../test/renderDashboard';
import { AppProvider } from '../../app/AppContext';
import { ThemeProvider } from '../../theme/ThemeProvider';

describe('Page: Dashboard Dashboard', () => {
  it('renderiza o Hero Card (RF03)', () => {
    renderDashboard(<Dashboard />);
    expect(screen.getByText('Build your Next Flow')).toBeTruthy();
    expect(screen.getByText('Get Started')).toBeTruthy();
  });

  it('renderiza os cards de métricas (RF04)', () => {
    renderDashboard(<Dashboard />);
    expect(screen.getByText('Sistemas')).toBeTruthy();
    expect(screen.getByText('Publicados')).toBeTruthy();
    expect(screen.getByText('Rascunhos')).toBeTruthy();
  });

  it('exibe a contagem real de sistemas (RF01)', async () => {
    const client = fakeClient({
      listarSistemas: vi.fn().mockResolvedValue([
        { id: 'a', nome: 'Alfa' },
        { id: 'b', nome: 'Beta' },
        { id: 'c', nome: 'Gama' },
      ]),
    });
    renderDashboard(<Dashboard />, { client });
    expect(await screen.findByText('3')).toBeTruthy();
  });

  it('renderiza o FAB (RF05)', () => {
    renderDashboard(<Dashboard />);
    expect(screen.getByText('Create')).toBeTruthy();
  });

  it('oculta "Get Started" e o FAB "Create" para usuário sem permissão de criação (RN10)', () => {
    renderDashboard(<Dashboard />, { usuario: usuarioClienteFake });
    expect(screen.queryByText('Get Started')).toBeNull();
    expect(screen.queryByText('Create')).toBeNull();
  });

  it('renderiza os cards de resumo consolidado (RF03-RF06)', () => {
    renderDashboard(<Dashboard />);
    expect(screen.getByText('Últimos Acessos')).toBeTruthy();
    expect(screen.getByText('Reclamações/Feedback')).toBeTruthy();
    expect(screen.getByText('Resumo Financeiro')).toBeTruthy();
  });

  it('clicar no card "Sistemas" navega para a tela de seleção de sistemas', () => {
    const client = fakeClient({
      listarSistemas: async () => [],
      listarUltimosAcessos: async () => [],
      listarFeedback: async () => [],
      resumoFinanceiro: async () => ({ receita_total_centavos: 0, moeda: 'BRL', competencia: '' }),
    });
    render(
      <MemoryRouter initialEntries={['/dashboard']}>
        <ThemeProvider>
          <AppProvider client={client} usuario={usuarioFake}>
            <Routes>
              <Route path="/dashboard" element={<Dashboard />} />
              <Route path="/sistemas" element={<div>Seus sistemas</div>} />
            </Routes>
          </AppProvider>
        </ThemeProvider>
      </MemoryRouter>,
    );

    fireEvent.click(screen.getByText('Sistemas'));

    expect(screen.getByText('Seus sistemas')).toBeTruthy();
  });
});
