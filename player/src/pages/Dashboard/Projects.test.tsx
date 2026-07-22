import { describe, it, expect, vi } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import type { Sistema } from '../../api/types';
import { Projects } from './Projects';
import { renderDashboard, fakeClient, usuarioClienteFake } from '../../test/renderDashboard';

describe('Page: Projects Dashboard', () => {
  it('renderiza o título e a descrição', () => {
    renderDashboard(<Projects />);
    expect(screen.getByText('Projects')).toBeTruthy();
    expect(screen.getByText(/Gerencie seus projetos/i)).toBeTruthy();
  });

  it('lista os sistemas reais do tenant (RF02)', async () => {
    const sistemas: Sistema[] = [
      { id: 'a', nome: 'Alfa' },
      { id: 'b', nome: 'Beta' },
    ];
    const client = fakeClient({ listarSistemas: vi.fn().mockResolvedValue(sistemas) });
    renderDashboard(<Projects />, { client });

    expect(await screen.findByText('Alfa')).toBeTruthy();
    expect(screen.getByText('Beta')).toBeTruthy();
    expect(screen.getAllByText(/Abrir projeto/i).length).toBe(2);
  });

  it('exibe estado vazio com CTA quando não há sistemas (RF06)', async () => {
    const client = fakeClient({ listarSistemas: vi.fn().mockResolvedValue([]) });
    renderDashboard(<Projects />, { client });
    expect(await screen.findByText(/Nenhum projeto ainda/i)).toBeTruthy();
    expect(screen.getByText('Criar projeto')).toBeTruthy();
  });

  it('exibe erro com retry quando a listagem falha (RF06)', async () => {
    const client = fakeClient({ listarSistemas: vi.fn().mockRejectedValue(new Error('boom')) });
    renderDashboard(<Projects />, { client });
    await waitFor(() => expect(screen.getByRole('alert')).toBeTruthy());
    expect(screen.getByText('Tentar novamente')).toBeTruthy();
  });

  it('oculta os CTAs de criação para usuário sem permissão (RN10)', async () => {
    const sistemas: Sistema[] = [{ id: 'a', nome: 'Alfa' }];
    const client = fakeClient({ listarSistemas: vi.fn().mockResolvedValue(sistemas) });
    renderDashboard(<Projects />, { client, usuario: usuarioClienteFake });

    expect(await screen.findByText('Alfa')).toBeTruthy();
    expect(screen.queryByText('Criar novo projeto')).toBeNull();
  });

  it('oculta o CTA do empty state para usuário sem permissão (RN10)', async () => {
    const client = fakeClient({ listarSistemas: vi.fn().mockResolvedValue([]) });
    renderDashboard(<Projects />, { client, usuario: usuarioClienteFake });

    expect(await screen.findByText(/Nenhum projeto ainda/i)).toBeTruthy();
    expect(screen.queryByText('Criar projeto')).toBeNull();
  });
});
