import { describe, expect, it, vi } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import type { ApiClient } from '../api/client';
import { ApiError } from '../api/client';
import type { PontoAcessosMensal } from '../api/types';
import { useAcessosPorMes } from './useAcessosPorMes';

function fakeClient(over: Partial<ApiClient>): ApiClient {
  return over as unknown as ApiClient;
}

const pontos: PontoAcessosMensal[] = [
  { competencia: '2026-03', total: 12 },
  { competencia: '2026-08', total: 42 },
];

describe('useAcessosPorMes', () => {
  it('passa de carregando para pronto com os pontos', async () => {
    const client = fakeClient({ acessosPorMes: vi.fn().mockResolvedValue(pontos) });

    const { result } = renderHook(() => useAcessosPorMes(client));
    expect(result.current.estado.fase).toBe('carregando');

    await waitFor(() => expect(result.current.estado.fase).toBe('pronto'));
    expect(result.current.estado).toEqual({ fase: 'pronto', pontos });
  });

  it('entra em erro e recarrega com sucesso', async () => {
    const buscar = vi
      .fn()
      .mockRejectedValueOnce(new ApiError(500, 'INTERNAL', 'boom'))
      .mockResolvedValueOnce(pontos);
    const client = fakeClient({ acessosPorMes: buscar });

    const { result } = renderHook(() => useAcessosPorMes(client));
    await waitFor(() => expect(result.current.estado.fase).toBe('erro'));

    act(() => result.current.recarregar());
    await waitFor(() => expect(result.current.estado.fase).toBe('pronto'));
    expect(buscar).toHaveBeenCalledTimes(2);
  });
});
