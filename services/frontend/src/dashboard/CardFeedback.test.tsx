import { describe, it, expect, vi } from 'vitest';
import { screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { CardFeedback } from './CardFeedback';
import { renderDashboard, fakeClient } from '../test/renderDashboard';

const item = {
  id: 'f1',
  tenant_nome: 'Acme',
  mensagem: 'Não consigo publicar',
  status: 'pendente' as const,
  criado_em: '2026-08-06T12:00:00Z',
};

describe('CardFeedback (RF05, RN03)', () => {
  it('lista as mensagens com o status', async () => {
    const client = fakeClient({ listarFeedback: vi.fn().mockResolvedValue([item]) });
    renderDashboard(<CardFeedback />, { client });

    expect(await screen.findByText(/Não consigo publicar/)).toBeInTheDocument();
    expect(screen.getByText(/pendente/i)).toBeInTheDocument();
  });

  it('marca uma mensagem como respondida', async () => {
    const atualizarStatusFeedback = vi.fn().mockResolvedValue({ ...item, status: 'respondido' });
    const client = fakeClient({
      listarFeedback: vi.fn().mockResolvedValue([item]),
      atualizarStatusFeedback,
    });
    renderDashboard(<CardFeedback />, { client });

    await screen.findByText(/Não consigo publicar/);
    fireEvent.click(screen.getByRole('button', { name: /marcar como respondido/i }));

    await waitFor(() => expect(atualizarStatusFeedback).toHaveBeenCalledWith('f1', 'respondido'));
  });

  it('exibe estado vazio quando não há mensagens', async () => {
    const client = fakeClient({ listarFeedback: vi.fn().mockResolvedValue([]) });
    renderDashboard(<CardFeedback />, { client });
    expect(await screen.findByText(/nenhuma mensagem/i)).toBeInTheDocument();
  });
});
