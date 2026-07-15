import { describe, it, expect, vi } from 'vitest';
import { screen } from '@testing-library/react';
import { Overview } from './Overview';
import { renderDashboard, fakeClient } from '../../test/renderDashboard';

describe('Page: Overview Dashboard', () => {
  it('renderiza o Hero Card (RF03)', () => {
    renderDashboard(<Overview />);
    expect(screen.getByText('Build your Next Flow')).toBeTruthy();
    expect(screen.getByText('Get Started')).toBeTruthy();
  });

  it('renderiza os cards de métricas (RF04)', () => {
    renderDashboard(<Overview />);
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
    renderDashboard(<Overview />, { client });
    expect(await screen.findByText('3')).toBeTruthy();
  });

  it('renderiza o FAB (RF05)', () => {
    renderDashboard(<Overview />);
    expect(screen.getByText('Create')).toBeTruthy();
  });
});
