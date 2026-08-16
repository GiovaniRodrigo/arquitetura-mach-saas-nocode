import { describe, expect, it, vi } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import type { ApiClient } from '../api/client';
import { ApiError } from '../api/client';
import type { PontoReceitaMensal } from '../api/types';
import { useReceitaPorMes } from './useReceitaPorMes';

function fakeClient(over: Partial<ApiClient>): ApiClient {
  return over as unknown as ApiClient;
}

const pontos: PontoReceitaMensal[] = [
  { competencia: '2026-03', valor_centavos: 100000 },
  { competencia: '2026-08', valor_centavos: 140000 },
];

describe('useReceitaPorMes', () => {
  it('passa de carregando para pronto com os pontos e a moeda', async () => {
    const client = fakeClient({
      receitaPorMes: vi.fn().mockResolvedValue({ pontos, moeda: 'BRL' }),
    });

    const { result } = renderHook(() => useReceitaPorMes(client));
    expect(result.current.estado.fase).toBe('carregando');

    await waitFor(() => expect(result.current.estado.fase).toBe('pronto'));
    expect(result.current.estado).toEqual({ fase: 'pronto', pontos, moeda: 'BRL' });
  });

  it('entra em erro e recarrega com sucesso', async () => {
    const buscar = vi
      .fn()
      .mockRejectedValueOnce(new ApiError(500, 'INTERNAL', 'boom'))
      .mockResolvedValueOnce({ pontos, moeda: 'BRL' });
    const client = fakeClient({ receitaPorMes: buscar });

    const { result } = renderHook(() => useReceitaPorMes(client));
    await waitFor(() => expect(result.current.estado.fase).toBe('erro'));

    act(() => result.current.recarregar());
    await waitFor(() => expect(result.current.estado.fase).toBe('pronto'));
    expect(buscar).toHaveBeenCalledTimes(2);
  });
});
