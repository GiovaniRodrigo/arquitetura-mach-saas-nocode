import { describe, it, expect, vi } from 'vitest';
import { screen } from '@testing-library/react';
import { Dashboard } from './Dashboard';
import { renderDashboard, fakeClient, usuarioClienteFake } from '../../test/renderDashboard';

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
});
