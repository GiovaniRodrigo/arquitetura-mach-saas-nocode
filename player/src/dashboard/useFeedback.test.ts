import { describe, expect, it, vi } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import type { ApiClient } from '../api/client';
import { ApiError } from '../api/client';
import type { Feedback } from '../api/types';
import { useFeedback } from './useFeedback';

function fakeClient(over: Partial<ApiClient>): ApiClient {
  return over as unknown as ApiClient;
}

const item: Feedback = {
  id: 'f1',
  tenant_nome: 'Acme',
  mensagem: 'Não consigo publicar',
  status: 'pendente',
  criado_em: '2026-08-06T12:00:00Z',
};

describe('useFeedback (RF05, RN03)', () => {
  it('passa de carregando para pronto com os itens', async () => {
    const client = fakeClient({ listarFeedback: vi.fn().mockResolvedValue([item]) });

    const { result } = renderHook(() => useFeedback(client));
    expect(result.current.estado.fase).toBe('carregando');

    await waitFor(() => expect(result.current.estado.fase).toBe('pronto'));
    expect(result.current.estado).toEqual({ fase: 'pronto', itens: [item] });
  });

  it('entra em vazio quando não há itens', async () => {
    const client = fakeClient({ listarFeedback: vi.fn().mockResolvedValue([]) });
    const { result } = renderHook(() => useFeedback(client));
    await waitFor(() => expect(result.current.estado.fase).toBe('vazio'));
  });

  it('entra em erro e recarrega com sucesso', async () => {
    const listar = vi
      .fn()
      .mockRejectedValueOnce(new ApiError(500, 'INTERNAL', 'boom'))
      .mockResolvedValueOnce([item]);
    const client = fakeClient({ listarFeedback: listar });

    const { result } = renderHook(() => useFeedback(client));
    await waitFor(() => expect(result.current.estado.fase).toBe('erro'));

    act(() => result.current.recarregar());
    await waitFor(() => expect(result.current.estado.fase).toBe('pronto'));
    expect(listar).toHaveBeenCalledTimes(2);
  });

  it('marcarRespondido delega para client.atualizarStatusFeedback e recarrega', async () => {
    const atualizarStatusFeedback = vi.fn().mockResolvedValue({ ...item, status: 'respondido' });
    const client = fakeClient({
      listarFeedback: vi.fn().mockResolvedValue([item]),
      atualizarStatusFeedback,
    });

    const { result } = renderHook(() => useFeedback(client));
    await waitFor(() => expect(result.current.estado.fase).toBe('pronto'));

    await act(async () => {
      await result.current.marcarRespondido('f1');
    });
    expect(atualizarStatusFeedback).toHaveBeenCalledWith('f1', 'respondido');
  });
});
