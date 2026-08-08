import { describe, expect, it, vi } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import type { ApiClient } from '../api/client';
import { ApiError } from '../api/client';
import type { EventoLogin } from '../api/types';
import { useUltimosAcessos } from './useUltimosAcessos';

function fakeClient(over: Partial<ApiClient>): ApiClient {
  return over as unknown as ApiClient;
}

describe('useUltimosAcessos (RF04, RN02)', () => {
  it('passa de carregando para pronto com os eventos', async () => {
    const eventos: EventoLogin[] = [{ usuario_nome: 'Ana', tenant_nome: 'Acme', criado_em: '2026-08-06T12:00:00Z' }];
    const client = fakeClient({ listarUltimosAcessos: vi.fn().mockResolvedValue(eventos) });

    const { result } = renderHook(() => useUltimosAcessos(client));
    expect(result.current.estado.fase).toBe('carregando');

    await waitFor(() => expect(result.current.estado.fase).toBe('pronto'));
    expect(result.current.estado).toEqual({ fase: 'pronto', eventos });
  });

  it('entra em vazio quando não há eventos', async () => {
    const client = fakeClient({ listarUltimosAcessos: vi.fn().mockResolvedValue([]) });
    const { result } = renderHook(() => useUltimosAcessos(client));
    await waitFor(() => expect(result.current.estado.fase).toBe('vazio'));
  });

  it('entra em erro e recarrega com sucesso', async () => {
    const listar = vi
      .fn()
      .mockRejectedValueOnce(new ApiError(500, 'INTERNAL', 'boom'))
      .mockResolvedValueOnce([{ usuario_nome: 'Ana', tenant_nome: 'Acme', criado_em: '2026-08-06T12:00:00Z' }]);
    const client = fakeClient({ listarUltimosAcessos: listar });

    const { result } = renderHook(() => useUltimosAcessos(client));
    await waitFor(() => expect(result.current.estado.fase).toBe('erro'));

    act(() => result.current.recarregar());
    await waitFor(() => expect(result.current.estado.fase).toBe('pronto'));
    expect(listar).toHaveBeenCalledTimes(2);
  });
});
