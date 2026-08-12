import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import type { ApiClient } from '../api/client';
import type { Design } from '../api/types';
import { useCanvasDesign } from './useCanvasDesign';

const entrarMock = vi.fn();
const mutarMock = vi.fn();
const sairMock = vi.fn();
let handlersCapturados: any;
let paramsCapturados: any;

vi.mock('../collab/phoenixSocket', () => ({
  CollabClient: vi.fn().mockImplementation(() => ({
    entrar: (screenId: string, handlers: any, params: any) => {
      handlersCapturados = handlers;
      paramsCapturados = params;
      return entrarMock(screenId, handlers, params);
    },
    mutar: mutarMock,
    sair: sairMock,
    bloquear: vi.fn(),
    desbloquear: vi.fn(),
    cursor: vi.fn(),
  })),
}));

function fakeClient(over: Partial<ApiClient>): ApiClient {
  return over as unknown as ApiClient;
}

const arvoreVaziaBase = { blind_index: 'root', tipo: 'tela', componente_filhos: [] };

function designFake(arvore: Design['arvore'] = arvoreVaziaBase): Design {
  return { id: 'd1', sistema_id: 's1', nome: 'Home', arvore };
}

beforeEach(() => {
  sessionStorage.clear();
  entrarMock.mockReset().mockResolvedValue(undefined);
  mutarMock.mockReset();
  sairMock.mockReset();
  handlersCapturados = undefined;
  paramsCapturados = undefined;
});

afterEach(() => {
  vi.useRealTimers();
});

async function montarPronto(client: ApiClient, arvoreInicial: Design['arvore'] = arvoreVaziaBase) {
  entrarMock.mockImplementation(async (_id, handlers) => handlers.onJoin?.(arvoreInicial));
  const hook = renderHook(() => useCanvasDesign(client, { designId: 'd1', sistemaId: 's1', nome: 'Home' }));
  await waitFor(() => expect(hook.result.current.estado.fase).toBe('pronto'));
  return hook;
}

describe('useCanvasDesign (RF09/RF06)', () => {
  it('busca o design via REST, junta-se ao collab com os params corretos e fica pronto', async () => {
    const design = designFake();
    const client = fakeClient({ obterDesign: vi.fn().mockResolvedValue(design) });
    const { result } = await montarPronto(client, design.arvore);

    expect(result.current.arvore).toEqual(design.arvore);
    expect(paramsCapturados).toEqual({ sistema_id: 's1', design_id: 'd1', nome: 'Home' });
  });

  it('adicionarComponente usa os padrões do componentRegistry e seleciona o novo nó', async () => {
    const client = fakeClient({ obterDesign: vi.fn().mockResolvedValue(designFake()) });
    const { result } = await montarPronto(client);

    act(() => result.current.adicionarComponente('botao'));

    expect(result.current.arvore?.componente_filhos).toHaveLength(1);
    const novo = result.current.arvore?.componente_filhos?.[0];
    expect(novo?.tipo).toBe('botao');
    expect(novo?.propriedades?.texto).toBe('Botão');
    expect(result.current.selecionado).toBe(novo?.blind_index);
    expect(mutarMock).toHaveBeenCalledWith(
      expect.objectContaining({ tipo: 'add_child', parent: 'root' }),
    );
  });

  it('adicionarComponente aceita um parent específico (nesting dentro de um container)', async () => {
    const arvore = {
      ...arvoreVaziaBase,
      componente_filhos: [{ blind_index: 'cont-1', tipo: 'container', componente_filhos: [] }],
    };
    const client = fakeClient({ obterDesign: vi.fn().mockResolvedValue(designFake(arvore)) });
    const { result } = await montarPronto(client, arvore);

    act(() => result.current.adicionarComponente('botao', 'cont-1'));

    expect(result.current.arvore?.componente_filhos?.[0].componente_filhos).toHaveLength(1);
  });

  it('moverNo emite uma mutação move e reflete o reordenamento local', async () => {
    const arvore = {
      ...arvoreVaziaBase,
      componente_filhos: [
        { blind_index: 'b1', tipo: 'x' },
        { blind_index: 'b2', tipo: 'x' },
      ],
    };
    const client = fakeClient({ obterDesign: vi.fn().mockResolvedValue(designFake(arvore)) });
    const { result } = await montarPronto(client, arvore);

    act(() => result.current.moverNo('b1', 'root', 2));

    expect(result.current.arvore?.componente_filhos?.map((f) => f.blind_index)).toEqual(['b2', 'b1']);
    expect(mutarMock).toHaveBeenCalledWith({ tipo: 'move', blind_index: 'b1', novo_parent: 'root', index: 2 });
  });

  it('removerComponente remove o nó e limpa a seleção se estava selecionado', async () => {
    const arvore = { ...arvoreVaziaBase, componente_filhos: [{ blind_index: 'b1', tipo: 'botao' }] };
    const client = fakeClient({ obterDesign: vi.fn().mockResolvedValue(designFake(arvore)) });
    const { result } = await montarPronto(client, arvore);

    act(() => result.current.selecionar('b1'));
    act(() => result.current.removerComponente('b1'));

    expect(result.current.arvore?.componente_filhos).toHaveLength(0);
    expect(result.current.selecionado).toBeNull();
  });

  it('mutação remota (onMutation) atualiza a árvore local sem entrar no histórico de undo', async () => {
    const client = fakeClient({ obterDesign: vi.fn().mockResolvedValue(designFake()) });
    const { result } = await montarPronto(client);

    act(() => {
      handlersCapturados.onMutation(
        { tipo: 'add_child', parent: 'root', no: { blind_index: 'b2', tipo: 'texto' } },
        'outro-usuario',
      );
    });

    await waitFor(() => expect(result.current.arvore?.componente_filhos).toHaveLength(1));
    expect(result.current.podeDesfazer).toBe(false);
  });

  it('desfazer/refazer restauram snapshots e propagam via set_tree', async () => {
    const client = fakeClient({ obterDesign: vi.fn().mockResolvedValue(designFake()) });
    const { result } = await montarPronto(client);

    act(() => result.current.adicionarComponente('botao'));
    expect(result.current.arvore?.componente_filhos).toHaveLength(1);
    expect(result.current.podeDesfazer).toBe(true);
    expect(result.current.podeRefazer).toBe(false);

    act(() => result.current.desfazer());
    expect(result.current.arvore?.componente_filhos).toHaveLength(0);
    expect(result.current.podeDesfazer).toBe(false);
    expect(result.current.podeRefazer).toBe(true);
    expect(mutarMock).toHaveBeenLastCalledWith(
      expect.objectContaining({ tipo: 'set_tree' }),
    );

    act(() => result.current.refazer());
    expect(result.current.arvore?.componente_filhos).toHaveLength(1);
    expect(result.current.podeDesfazer).toBe(true);
    expect(result.current.podeRefazer).toBe(false);
  });

  it('desfazer/refazer sem histórico não fazem nada', async () => {
    const client = fakeClient({ obterDesign: vi.fn().mockResolvedValue(designFake()) });
    const { result } = await montarPronto(client);

    act(() => result.current.desfazer());
    act(() => result.current.refazer());
    expect(result.current.arvore?.componente_filhos).toHaveLength(0);
  });

  it('statusSalvamento vira "salvando" numa mutação local e volta a "salvo" após o debounce', async () => {
    const client = fakeClient({ obterDesign: vi.fn().mockResolvedValue(designFake()) });
    const { result } = await montarPronto(client);

    expect(result.current.statusSalvamento).toBe('salvo');
    vi.useFakeTimers();
    act(() => result.current.adicionarComponente('botao'));
    expect(result.current.statusSalvamento).toBe('salvando');

    act(() => vi.advanceTimersByTime(5601));
    expect(result.current.statusSalvamento).toBe('salvo');
  });

  it('sai do collab ao desmontar', async () => {
    const client = fakeClient({ obterDesign: vi.fn().mockResolvedValue(designFake()) });
    const { unmount } = await montarPronto(client);
    unmount();
    expect(sairMock).toHaveBeenCalled();
  });

  it('erro ao buscar o design entra em fase erro', async () => {
    const client = fakeClient({ obterDesign: vi.fn().mockRejectedValue(new Error('falha de rede')) });
    const { result } = renderHook(() =>
      useCanvasDesign(client, { designId: 'd1', sistemaId: 's1', nome: 'Home' }),
    );
    await waitFor(() => expect(result.current.estado.fase).toBe('erro'));
  });
});
