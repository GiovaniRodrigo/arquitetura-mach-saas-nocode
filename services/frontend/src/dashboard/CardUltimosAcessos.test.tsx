import { describe, it, expect, vi } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { CardUltimosAcessos } from './CardUltimosAcessos';
import { renderDashboard, fakeClient } from '../test/renderDashboard';

describe('CardUltimosAcessos (RF04)', () => {
  it('lista até 10 logins mais recentes', async () => {
    const eventos = [
      { usuario_nome: 'Ana', tenant_nome: 'Acme', criado_em: '2026-08-06T12:00:00Z' },
      { usuario_nome: 'Bruno', tenant_nome: 'Beta', criado_em: '2026-08-06T11:00:00Z' },
    ];
    const client = fakeClient({ listarUltimosAcessos: vi.fn().mockResolvedValue(eventos) });
    renderDashboard(<CardUltimosAcessos />, { client });

    expect(await screen.findByText(/Ana/)).toBeInTheDocument();
    expect(screen.getByText(/Bruno/)).toBeInTheDocument();
  });

  it('exibe estado vazio quando não há eventos', async () => {
    const client = fakeClient({ listarUltimosAcessos: vi.fn().mockResolvedValue([]) });
    renderDashboard(<CardUltimosAcessos />, { client });
    expect(await screen.findByText(/nenhum acesso/i)).toBeInTheDocument();
  });

  it('exibe erro com retry quando a listagem falha', async () => {
    const client = fakeClient({ listarUltimosAcessos: vi.fn().mockRejectedValue(new Error('boom')) });
    renderDashboard(<CardUltimosAcessos />, { client });
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });
});
