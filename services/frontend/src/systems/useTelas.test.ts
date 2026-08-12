import { describe, expect, it, vi } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import type { ApiClient } from '../api/client';
import { ApiError } from '../api/client';
import type { DesignResumo } from '../api/types';
import { useTelas } from './useTelas';

function fakeClient(over: Partial<ApiClient>): ApiClient {
  return over as unknown as ApiClient;
}

describe('useTelas (RF09)', () => {
  it('passa de carregando para pronto com a lista', async () => {
    const telas: DesignResumo[] = [{ id: 'd1', nome: 'Home' }];
    const client = fakeClient({ listarTelas: vi.fn().mockResolvedValue(telas) });

    const { result } = renderHook(() => useTelas(client, 's1'));
    expect(result.current.estado.fase).toBe('carregando');

    await waitFor(() => expect(result.current.estado.fase).toBe('pronto'));
    expect(result.current.estado).toEqual({ fase: 'pronto', telas });
  });

  it('entra em vazio quando não há telas ainda (não confunde com erro)', async () => {
    const client = fakeClient({ listarTelas: vi.fn().mockResolvedValue([]) });
    const { result } = renderHook(() => useTelas(client, 's1'));
    await waitFor(() => expect(result.current.estado.fase).toBe('vazio'));
  });

  it('entra em erro quando a listagem falha e recarrega com sucesso', async () => {
    const listar = vi
      .fn()
      .mockRejectedValueOnce(new ApiError(500, 'INTERNAL', 'boom'))
      .mockResolvedValueOnce([{ id: 'd1', nome: 'Home' }]);
    const client = fakeClient({ listarTelas: listar });

    const { result } = renderHook(() => useTelas(client, 's1'));
    await waitFor(() => expect(result.current.estado.fase).toBe('erro'));

    act(() => result.current.recarregar());
    await waitFor(() => expect(result.current.estado.fase).toBe('pronto'));
    expect(listar).toHaveBeenCalledTimes(2);
  });

  it('criarTela envia sistema_id/nome/uma árvore raiz vazia e recarrega a lista', async () => {
    const criarDesign = vi.fn().mockResolvedValue({ id: 'd-novo' });
    const listarTelas = vi.fn().mockResolvedValueOnce([]).mockResolvedValueOnce([{ id: 'd-novo', nome: 'Nova' }]);
    const client = fakeClient({ listarTelas, criarDesign });

    const { result } = renderHook(() => useTelas(client, 's1'));
    await waitFor(() => expect(result.current.estado.fase).toBe('vazio'));

    let idCriado = '';
    await act(async () => {
      idCriado = await result.current.criarTela('Nova');
    });

    expect(idCriado).toBe('d-novo');
    const [args] = criarDesign.mock.calls[0];
    expect(args.sistemaId).toBe('s1');
    expect(args.nome).toBe('Nova');
    expect(args.arvore.tipo).toBe('tela');
    expect(args.arvore.componente_filhos).toEqual([]);
    expect(typeof args.arvore.blind_index).toBe('string');
    expect(args.arvore.blind_index.length).toBeGreaterThan(0);

    await waitFor(() => expect(result.current.estado).toEqual({ fase: 'pronto', telas: [{ id: 'd-novo', nome: 'Nova' }] }));
  });

  it('removerTela delega para client.removerDesign e recarrega', async () => {
    const removerDesign = vi.fn().mockResolvedValue(undefined);
    const listarTelas = vi
      .fn()
      .mockResolvedValueOnce([{ id: 'd1', nome: 'Home' }])
      .mockResolvedValueOnce([]);
    const client = fakeClient({ listarTelas, removerDesign });

    const { result } = renderHook(() => useTelas(client, 's1'));
    await waitFor(() => expect(result.current.estado.fase).toBe('pronto'));

    await act(async () => {
      await result.current.removerTela('d1');
    });

    expect(removerDesign).toHaveBeenCalledWith('d1');
    await waitFor(() => expect(result.current.estado.fase).toBe('vazio'));
  });
});
