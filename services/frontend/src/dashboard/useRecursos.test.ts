import { describe, expect, it, vi, afterEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import type { ApiClient } from '../api/client';
import { ApiError } from '../api/client';
import type { RecursosResponse } from '../api/types';
import { useRecursos } from './useRecursos';

function fakeClient(over: Partial<ApiClient>): ApiClient {
  return over as unknown as ApiClient;
}

const recursos: RecursosResponse = {
  servicos: [
    {
      nome: 'iam',
      status: 'servindo',
      cpu_millicores: 18,
      memoria_bytes: 15728640,
      requisicoes_por_segundo: 0.3,
      taxa_sucesso_percent: 100,
      latencia_p99_ms: 1.2,
    },
    {
      nome: 'logic',
      status: 'indisponivel',
      cpu_millicores: 0,
      memoria_bytes: 0,
      requisicoes_por_segundo: 0,
      taxa_sucesso_percent: 0,
      latencia_p99_ms: 0,
    },
  ],
  coletado_em_unix: 1755158400,
};

afterEach(() => {
  vi.useRealTimers();
});

describe('useRecursos (RF06, RF07, RNF02)', () => {
  it('passa de carregando para pronto com a lista de recursos', async () => {
    const client = fakeClient({ obterRecursos: vi.fn().mockResolvedValue(recursos) });

    const { result } = renderHook(() => useRecursos(client));
    expect(result.current.estado.fase).toBe('carregando');

    await waitFor(() => expect(result.current.estado.fase).toBe('pronto'));
    expect(result.current.estado).toEqual({ fase: 'pronto', recursos });
  });

  it('entra em erro quando o endpoint do Monitor falha e recarrega com sucesso (RNF02)', async () => {
    const buscar = vi
      .fn()
      .mockRejectedValueOnce(new ApiError(502, 'RECURSOS_INDISPONIVEIS', 'boom'))
      .mockResolvedValueOnce(recursos);
    const client = fakeClient({ obterRecursos: buscar });

    const { result } = renderHook(() => useRecursos(client));
    await waitFor(() => expect(result.current.estado.fase).toBe('erro'));

    act(() => result.current.recarregar());
    await waitFor(() => expect(result.current.estado.fase).toBe('pronto'));
    expect(buscar).toHaveBeenCalledTimes(2);
  });

  it('uma entrada individual "indisponivel" não gera estado de erro (RN01)', async () => {
    const client = fakeClient({ obterRecursos: vi.fn().mockResolvedValue(recursos) });

    const { result } = renderHook(() => useRecursos(client));
    await waitFor(() => expect(result.current.estado.fase).toBe('pronto'));
    if (result.current.estado.fase !== 'pronto') throw new Error('esperava pronto');
    expect(result.current.estado.recursos.servicos.find((s) => s.nome === 'logic')?.status).toBe(
      'indisponivel',
    );
  });

  it('atualiza automaticamente a cada intervalo configurado (RF07)', async () => {
    vi.useFakeTimers();
    const buscar = vi.fn().mockResolvedValue(recursos);
    const client = fakeClient({ obterRecursos: buscar });

    renderHook(() => useRecursos(client, { intervaloMs: 1000 }));
    await act(async () => {
      await Promise.resolve();
    });
    expect(buscar).toHaveBeenCalledTimes(1);

    await act(async () => {
      vi.advanceTimersByTime(1000);
      await Promise.resolve();
    });
    expect(buscar).toHaveBeenCalledTimes(2);

    await act(async () => {
      vi.advanceTimersByTime(1000);
      await Promise.resolve();
    });
    expect(buscar).toHaveBeenCalledTimes(3);

    await act(async () => {
      vi.advanceTimersByTime(1000);
      await Promise.resolve();
    });
    expect(buscar).toHaveBeenCalledTimes(4);
  });

  it('limpa o intervalo de auto-atualização ao desmontar (RF07)', async () => {
    vi.useFakeTimers();
    const buscar = vi.fn().mockResolvedValue(recursos);
    const client = fakeClient({ obterRecursos: buscar });

    const { unmount } = renderHook(() => useRecursos(client, { intervaloMs: 1000 }));
    await act(async () => {
      await Promise.resolve();
    });
    unmount();
    const chamadasAntes = buscar.mock.calls.length;

    await act(async () => {
      vi.advanceTimersByTime(5000);
      await Promise.resolve();
    });
    expect(buscar).toHaveBeenCalledTimes(chamadasAntes);
  });
});
