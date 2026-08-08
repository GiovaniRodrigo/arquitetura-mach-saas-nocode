import { describe, expect, it, vi } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import type { ApiClient } from '../api/client';
import { ApiError } from '../api/client';
import type { ResumoFinanceiro } from '../api/types';
import { useResumoFinanceiro } from './useResumoFinanceiro';

function fakeClient(over: Partial<ApiClient>): ApiClient {
  return over as unknown as ApiClient;
}

const resumo: ResumoFinanceiro = { receita_total_centavos: 150000, moeda: 'BRL', competencia: '2026-08' };

describe('useResumoFinanceiro (RF06, RN04)', () => {
  it('passa de carregando para pronto com o resumo', async () => {
    const client = fakeClient({ resumoFinanceiro: vi.fn().mockResolvedValue(resumo) });

    const { result } = renderHook(() => useResumoFinanceiro(client));
    expect(result.current.estado.fase).toBe('carregando');

    await waitFor(() => expect(result.current.estado.fase).toBe('pronto'));
    expect(result.current.estado).toEqual({ fase: 'pronto', resumo });
  });

  it('entra em erro e recarrega com sucesso', async () => {
    const buscar = vi
      .fn()
      .mockRejectedValueOnce(new ApiError(500, 'INTERNAL', 'boom'))
      .mockResolvedValueOnce(resumo);
    const client = fakeClient({ resumoFinanceiro: buscar });

    const { result } = renderHook(() => useResumoFinanceiro(client));
    await waitFor(() => expect(result.current.estado.fase).toBe('erro'));

    act(() => result.current.recarregar());
    await waitFor(() => expect(result.current.estado.fase).toBe('pronto'));
    expect(buscar).toHaveBeenCalledTimes(2);
  });
});
