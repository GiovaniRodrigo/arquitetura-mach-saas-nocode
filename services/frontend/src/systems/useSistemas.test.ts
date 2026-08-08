import { describe, expect, it, vi } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import type { ApiClient } from '../api/client';
import { ApiError } from '../api/client';
import type { Sistema } from '../api/types';
import { useSistemas } from './useSistemas';

function fakeClient(over: Partial<ApiClient>): ApiClient {
  return over as unknown as ApiClient;
}

describe('useSistemas', () => {
  it('passa de carregando para pronto com a lista', async () => {
    const sistemas: Sistema[] = [{ id: 'a', nome: 'Alfa' }];
    const client = fakeClient({ listarSistemas: vi.fn().mockResolvedValue(sistemas) });

    const { result } = renderHook(() => useSistemas(client));
    expect(result.current.estado.fase).toBe('carregando');

    await waitFor(() => expect(result.current.estado.fase).toBe('pronto'));
    expect(result.current.estado).toEqual({ fase: 'pronto', sistemas });
  });

  it('entra em vazio quando a lista é vazia', async () => {
    const client = fakeClient({ listarSistemas: vi.fn().mockResolvedValue([]) });
    const { result } = renderHook(() => useSistemas(client));
    await waitFor(() => expect(result.current.estado.fase).toBe('vazio'));
  });

  it('entra em erro e recarrega com sucesso', async () => {
    const listar = vi
      .fn()
      .mockRejectedValueOnce(new ApiError(500, 'INTERNAL', 'boom'))
      .mockResolvedValueOnce([{ id: 'a', nome: 'Alfa' }]);
    const client = fakeClient({ listarSistemas: listar });

    const { result } = renderHook(() => useSistemas(client));
    await waitFor(() => expect(result.current.estado.fase).toBe('erro'));
    if (result.current.estado.fase === 'erro') {
      expect(result.current.estado.mensagem).toContain('boom');
    }

    act(() => result.current.recarregar());
    await waitFor(() => expect(result.current.estado.fase).toBe('pronto'));
    expect(listar).toHaveBeenCalledTimes(2);
  });

  it('com tenantId, filtra a listagem por cliente (RF08)', async () => {
    const listarSistemas = vi.fn().mockResolvedValue([{ id: 'a', nome: 'Alfa' }]);
    const client = fakeClient({ listarSistemas });

    const { result } = renderHook(() => useSistemas(client, 't1'));
    await waitFor(() => expect(result.current.estado.fase).toBe('pronto'));
    expect(listarSistemas).toHaveBeenCalledWith('t1');
  });

  it('criar delega para client.criarSistema', async () => {
    const novo: Sistema = { id: 'z', nome: 'Zeta' };
    const criarSistema = vi.fn().mockResolvedValue(novo);
    const client = fakeClient({
      listarSistemas: vi.fn().mockResolvedValue([]),
      criarSistema,
    });

    const { result } = renderHook(() => useSistemas(client));
    await waitFor(() => expect(result.current.estado.fase).toBe('vazio'));

    await act(async () => {
      const r = await result.current.criar('Zeta');
      expect(r).toEqual(novo);
    });
    expect(criarSistema).toHaveBeenCalledWith('Zeta');
  });
});
