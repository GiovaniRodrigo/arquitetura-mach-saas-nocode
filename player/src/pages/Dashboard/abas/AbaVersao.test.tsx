import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { AppProvider } from '../../../app/AppContext';
import { AbaVersao } from './AbaVersao';
import { fakeClient, usuarioFake } from '../../../test/renderDashboard';

function renderAba(client: ReturnType<typeof fakeClient>) {
  return render(
    <MemoryRouter initialEntries={['/dashboard/clientes/t1/sistemas/s1/versao']}>
      <AppProvider client={client} usuario={usuarioFake}>
        <Routes>
          <Route path="/dashboard/clientes/:tenantId/sistemas/:sistemaId/versao" element={<AbaVersao />} />
        </Routes>
      </AppProvider>
    </MemoryRouter>,
  );
}

describe('AbaVersao (RF12)', () => {
  it('lista as versões do sistema, mais recente primeiro, com a ativa destacada', async () => {
    const client = fakeClient({
      listarVersoes: vi.fn().mockResolvedValue([
        { id: 'v2', numero: 2, ativa: true, criado_em: '2026-08-06T12:00:00Z' },
        { id: 'v1', numero: 1, ativa: false, criado_em: '2026-08-01T12:00:00Z' },
      ]),
    });
    renderAba(client);

    expect(await screen.findByText(/versão 2/i)).toBeInTheDocument();
    expect(screen.getByText(/ativa/i)).toBeInTheDocument();
  });

  it('publica uma versão ao clicar em "Publicar"', async () => {
    const publicarVersao = vi.fn().mockResolvedValue(undefined);
    const client = fakeClient({
      listarVersoes: vi.fn().mockResolvedValue([
        { id: 'v1', numero: 1, ativa: false, criado_em: '2026-08-01T12:00:00Z' },
      ]),
      publicarVersao,
    });
    renderAba(client);

    fireEvent.click(await screen.findByRole('button', { name: /publicar/i }));
    await waitFor(() => expect(publicarVersao).toHaveBeenCalledWith('s1', 'v1'));
  });

  it('reverte para uma versão anterior ao clicar em "Reverter"', async () => {
    const reverterVersao = vi.fn().mockResolvedValue(undefined);
    const client = fakeClient({
      listarVersoes: vi.fn().mockResolvedValue([
        { id: 'v2', numero: 2, ativa: true, criado_em: '2026-08-06T12:00:00Z' },
        { id: 'v1', numero: 1, ativa: false, criado_em: '2026-08-01T12:00:00Z' },
      ]),
      reverterVersao,
    });
    renderAba(client);

    await screen.findByText(/versão 2/i);
    fireEvent.click(screen.getByRole('button', { name: /reverter/i }));
    await waitFor(() => expect(reverterVersao).toHaveBeenCalledWith('s1', 'v1'));
  });
});
