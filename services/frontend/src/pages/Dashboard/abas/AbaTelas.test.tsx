import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor, within } from '@testing-library/react';
import '@testing-library/jest-dom';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { AbaTelas } from './AbaTelas';
import { AppProvider } from '../../../app/AppContext';
import { ApiError, type ApiClient } from '../../../api/client';
import type { Componente, Design, DesignResumo } from '../../../api/types';
import type { UsuarioAutenticado } from '../../../auth/jwt';

const entrarMock = vi.fn();
const mutarMock = vi.fn();
const sairMock = vi.fn();

vi.mock('../../../collab/phoenixSocket', () => ({
  CollabClient: vi.fn().mockImplementation(() => ({
    entrar: entrarMock,
    mutar: mutarMock,
    sair: sairMock,
    bloquear: vi.fn(),
    desbloquear: vi.fn(),
    cursor: vi.fn(),
  })),
}));

const usuarioFake: UsuarioAutenticado = { iniciais: 'A', podeCriarSistema: true, tenantId: 't1' };

function raiz(filhos: Componente[] = []): Componente {
  return { blind_index: 'root', tipo: 'tela', componente_filhos: filhos };
}

function fakeClient(over: Partial<ApiClient>): ApiClient {
  return over as unknown as ApiClient;
}

function renderComRota(client: ApiClient) {
  return render(
    <MemoryRouter initialEntries={['/dashboard/clientes/t1/sistemas/s1/telas']}>
      <AppProvider client={client} usuario={usuarioFake}>
        <Routes>
          <Route path="/dashboard/clientes/:tenantId/sistemas/:sistemaId/telas" element={<AbaTelas />} />
        </Routes>
      </AppProvider>
    </MemoryRouter>,
  );
}

function dataTransferFake(payload: string) {
  const dados: Record<string, string> = { 'text/plain': payload };
  return {
    setData: (_tipo: string, v: string) => {
      dados['text/plain'] = v;
    },
    getData: () => dados['text/plain'],
    effectAllowed: '',
    dropEffect: '',
  };
}

async function abrirTela(nome: string) {
  fireEvent.click(screen.getByRole('option', { name: new RegExp(`^${nome}$`, 'i') }));
  await waitFor(() => expect(screen.queryByRole('listbox', { name: /telas/i })).not.toBeInTheDocument());
}

beforeEach(() => {
  sessionStorage.clear();
  entrarMock.mockReset().mockImplementation(async (_id, handlers) => handlers.onJoin?.(raiz()));
  mutarMock.mockReset();
  sairMock.mockReset();
});

describe('AbaTelas (RF09/RF06 — editor visual completo)', () => {
  it('exibe "Nenhuma tela criada ainda" quando o sistema não tem telas', async () => {
    const client = fakeClient({ listarTelas: vi.fn().mockResolvedValue([]) });
    renderComRota(client);
    await waitFor(() => expect(screen.getByText(/nenhuma tela criada ainda/i)).toBeInTheDocument());
  });

  it('cria uma nova tela pelo modal e a seleciona automaticamente', async () => {
    const telas: DesignResumo[] = [];
    const listarTelas = vi.fn().mockImplementation(async () => telas);
    const criarDesign = vi.fn().mockImplementation(async ({ nome }: { nome: string }) => {
      telas.push({ id: 'd-novo', nome });
      return { id: 'd-novo' };
    });
    const obterDesign = vi.fn().mockResolvedValue({
      id: 'd-novo',
      sistema_id: 's1',
      nome: 'Home',
      arvore: raiz(),
    } as Design);
    const client = fakeClient({ listarTelas, criarDesign, obterDesign });

    renderComRota(client);
    await waitFor(() => expect(screen.getByText(/nenhuma tela criada ainda/i)).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /selecione uma tela/i }));
    fireEvent.click(screen.getByRole('button', { name: /nova tela/i }));

    const modal = await screen.findByRole('dialog', { name: /nova tela/i });
    fireEvent.change(within(modal).getByLabelText(/nome da tela/i), { target: { value: 'Home' } });
    fireEvent.click(within(modal).getByRole('button', { name: /criar tela/i }));

    await waitFor(() => expect(criarDesign).toHaveBeenCalledWith(expect.objectContaining({ sistemaId: 's1', nome: 'Home' })));
    await waitFor(() => expect(obterDesign).toHaveBeenCalledWith('d-novo'));
  });

  it('seleciona uma tela existente pelo dropdown do topbar e mostra o canvas em branco', async () => {
    const telas: DesignResumo[] = [{ id: 'd1', nome: 'Home' }];
    const client = fakeClient({
      listarTelas: vi.fn().mockResolvedValue(telas),
      obterDesign: vi.fn().mockResolvedValue({ id: 'd1', sistema_id: 's1', nome: 'Home', arvore: raiz() } as Design),
    });
    renderComRota(client);

    await waitFor(() => expect(screen.getByRole('button', { name: /selecione uma tela/i })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: /selecione uma tela/i }));
    await abrirTela('Home');

    await waitFor(() =>
      expect(screen.getByText(/tela em branco — arraste um componente/i)).toBeInTheDocument(),
    );
  });

  it('adiciona um componente pela paleta (aba Componentes) e o exibe no canvas', async () => {
    const telas: DesignResumo[] = [{ id: 'd1', nome: 'Home' }];
    const client = fakeClient({
      listarTelas: vi.fn().mockResolvedValue(telas),
      obterDesign: vi.fn().mockResolvedValue({ id: 'd1', sistema_id: 's1', nome: 'Home', arvore: raiz() } as Design),
    });
    renderComRota(client);

    fireEvent.click(await screen.findByRole('button', { name: /selecione uma tela/i }));
    await abrirTela('Home');

    fireEvent.click(await screen.findByRole('button', { name: /^Botão$/ }));

    await waitFor(() => expect(screen.getAllByText('Botão').length).toBeGreaterThan(1));
    expect(mutarMock).toHaveBeenCalledWith(expect.objectContaining({ tipo: 'add_child', parent: 'root' }));
  });

  it('seleciona um componente no canvas e edita/remove pelo Inspector', async () => {
    const arvoreComFilho = raiz([
      { blind_index: 'b1', tipo: 'botao', propriedades: { texto: 'Entrar', estilos: {} }, componente_filhos: [] },
    ]);
    const telas: DesignResumo[] = [{ id: 'd1', nome: 'Home' }];
    const client = fakeClient({
      listarTelas: vi.fn().mockResolvedValue(telas),
      obterDesign: vi.fn().mockResolvedValue({ id: 'd1', sistema_id: 's1', nome: 'Home', arvore: arvoreComFilho } as Design),
    });
    entrarMock.mockImplementation(async (_id, handlers) => handlers.onJoin?.(arvoreComFilho));

    renderComRota(client);
    fireEvent.click(await screen.findByRole('button', { name: /selecione uma tela/i }));
    await abrirTela('Home');

    fireEvent.click(await screen.findByRole('button', { name: /^botao$/i }));

    const painel = screen.getByRole('region', { name: /propriedades/i });
    await waitFor(() => expect(within(painel).getByText('botao')).toBeInTheDocument());
    expect(within(painel).getByLabelText(/^texto$/i)).toHaveValue('Entrar');

    fireEvent.click(within(painel).getByRole('button', { name: /^remover$/i }));
    expect(mutarMock).toHaveBeenCalledWith(expect.objectContaining({ tipo: 'remove', blind_index: 'b1' }));
  });

  it('painel Layers lista os componentes e permite selecionar/duplicar/remover', async () => {
    const arvoreComFilho = raiz([{ blind_index: 'b1', tipo: 'botao', propriedades: { estilos: {} }, componente_filhos: [] }]);
    const telas: DesignResumo[] = [{ id: 'd1', nome: 'Home' }];
    const client = fakeClient({
      listarTelas: vi.fn().mockResolvedValue(telas),
      obterDesign: vi.fn().mockResolvedValue({ id: 'd1', sistema_id: 's1', nome: 'Home', arvore: arvoreComFilho } as Design),
    });
    entrarMock.mockImplementation(async (_id, handlers) => handlers.onJoin?.(arvoreComFilho));

    renderComRota(client);
    fireEvent.click(await screen.findByRole('button', { name: /selecione uma tela/i }));
    await abrirTela('Home');

    fireEvent.click(screen.getByRole('tab', { name: /layers/i }));
    const arvore = screen.getByRole('tree', { name: /layers/i });
    expect(within(arvore).getByText('botao')).toBeInTheDocument();

    fireEvent.click(within(arvore).getByLabelText(/duplicar botao/i));
    expect(mutarMock).toHaveBeenCalledWith(expect.objectContaining({ tipo: 'add_child', parent: 'root', index: 1 }));
  });

  it('undo/redo no topbar chamam desfazer/refazer (set_tree) após uma mutação local', async () => {
    const telas: DesignResumo[] = [{ id: 'd1', nome: 'Home' }];
    const client = fakeClient({
      listarTelas: vi.fn().mockResolvedValue(telas),
      obterDesign: vi.fn().mockResolvedValue({ id: 'd1', sistema_id: 's1', nome: 'Home', arvore: raiz() } as Design),
    });
    renderComRota(client);
    fireEvent.click(await screen.findByRole('button', { name: /selecione uma tela/i }));
    await abrirTela('Home');

    const desfazer = () => screen.getByRole('button', { name: /desfazer/i });
    expect(desfazer()).toBeDisabled();

    fireEvent.click(await screen.findByRole('button', { name: /^Link$/ }));
    await waitFor(() => expect(desfazer()).not.toBeDisabled());

    fireEvent.click(desfazer());
    expect(mutarMock).toHaveBeenLastCalledWith(expect.objectContaining({ tipo: 'set_tree' }));
  });

  it('mantém a tela selecionada ao desmontar e remontar (troca entre Telas/Regras/Versão)', async () => {
    const telas: DesignResumo[] = [{ id: 'd1', nome: 'Home' }];
    const client = fakeClient({
      listarTelas: vi.fn().mockResolvedValue(telas),
      obterDesign: vi.fn().mockResolvedValue({ id: 'd1', sistema_id: 's1', nome: 'Home', arvore: raiz() } as Design),
    });

    const primeira = renderComRota(client);
    fireEvent.click(await screen.findByRole('button', { name: /selecione uma tela/i }));
    await abrirTela('Home');
    await waitFor(() => expect(screen.getByText(/tela em branco/i)).toBeInTheDocument());

    // Simula trocar para a aba "Regras de Negócio" (rota irmã) e voltar: o
    // Outlet desmonta AbaTelas por completo — sem sessionStorage, a seleção
    // se perderia e voltaria para "Selecione uma tela".
    primeira.unmount();
    renderComRota(client);

    await waitFor(() => expect(screen.getByRole('button', { name: /^Home$/ })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: /selecione uma tela/i })).not.toBeInTheDocument();
    await waitFor(() => expect(screen.getByText(/tela em branco/i)).toBeInTheDocument());
  });

  it('mantém o componente selecionado no Inspector ao desmontar e remontar', async () => {
    const arvoreComFilho = raiz([{ blind_index: 'b1', tipo: 'botao', propriedades: { estilos: {} }, componente_filhos: [] }]);
    const telas: DesignResumo[] = [{ id: 'd1', nome: 'Home' }];
    const client = fakeClient({
      listarTelas: vi.fn().mockResolvedValue(telas),
      obterDesign: vi.fn().mockResolvedValue({ id: 'd1', sistema_id: 's1', nome: 'Home', arvore: arvoreComFilho } as Design),
    });
    entrarMock.mockImplementation(async (_id, handlers) => handlers.onJoin?.(arvoreComFilho));

    const primeira = renderComRota(client);
    fireEvent.click(await screen.findByRole('button', { name: /selecione uma tela/i }));
    await abrirTela('Home');
    fireEvent.click(await screen.findByRole('button', { name: /^botao$/i }));

    const painel1 = screen.getByRole('region', { name: /propriedades/i });
    await waitFor(() => expect(within(painel1).getByText('botao')).toBeInTheDocument());
    await waitFor(() => expect(sessionStorage.getItem('mach:sistema:s1:tela-selecionada')).toBe('"d1"'));
    await waitFor(() => expect(sessionStorage.getItem('mach:tela:d1:componente-selecionado')).toBe('"b1"'));

    primeira.unmount();
    renderComRota(client);

    // Reconsulta a cada iteração (não captura um handle fixo): a região
    // "Propriedades" existe tanto no placeholder "nenhuma tela selecionada"
    // (enquanto a seleção ainda não restaurou) quanto no Inspector real —
    // são nós DOM diferentes, e o placeholder é substituído assim que
    // telaSelecionada resolve.
    await waitFor(() => {
      const painel = screen.getByRole('region', { name: /propriedades/i });
      expect(within(painel).getByText('botao')).toBeInTheDocument();
    });
  });

  it('arrastar um novo componente da paleta e soltar no canvas cria o nó (drag&drop nativo)', async () => {
    const telas: DesignResumo[] = [{ id: 'd1', nome: 'Home' }];
    const client = fakeClient({
      listarTelas: vi.fn().mockResolvedValue(telas),
      obterDesign: vi.fn().mockResolvedValue({ id: 'd1', sistema_id: 's1', nome: 'Home', arvore: raiz() } as Design),
    });
    renderComRota(client);
    fireEvent.click(await screen.findByRole('button', { name: /selecione uma tela/i }));
    await abrirTela('Home');

    const botaoPaleta = await screen.findByRole('button', { name: /^Imagem$/ });
    const canvasEl = screen.getByRole('presentation', { name: /canvas/i });
    const dataTransfer = dataTransferFake('');

    fireEvent.dragStart(botaoPaleta, { dataTransfer });
    fireEvent.dragOver(canvasEl, { dataTransfer });
    fireEvent.drop(canvasEl, { dataTransfer });

    expect(mutarMock).toHaveBeenCalledWith(expect.objectContaining({ tipo: 'add_child', no: expect.objectContaining({ tipo: 'imagem' }) }));
  });

  it('modo foco esconde os painéis Componentes/Layers e Propriedades', async () => {
    const telas: DesignResumo[] = [{ id: 'd1', nome: 'Home' }];
    const client = fakeClient({
      listarTelas: vi.fn().mockResolvedValue(telas),
      obterDesign: vi.fn().mockResolvedValue({ id: 'd1', sistema_id: 's1', nome: 'Home', arvore: raiz() } as Design),
    });
    renderComRota(client);
    fireEvent.click(await screen.findByRole('button', { name: /selecione uma tela/i }));
    await abrirTela('Home');

    expect(screen.getByRole('tablist')).toBeInTheDocument();
    expect(screen.getByRole('region', { name: /propriedades/i })).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Foco' }));

    expect(screen.queryByRole('tablist')).not.toBeInTheDocument();
    expect(screen.queryByRole('region', { name: /propriedades/i })).not.toBeInTheDocument();
    expect(screen.getByRole('presentation', { name: /canvas/i })).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Edição' }));
    expect(screen.getByRole('tablist')).toBeInTheDocument();
  });

  it('modo foco e visualização são mutuamente exclusivos', async () => {
    const telas: DesignResumo[] = [{ id: 'd1', nome: 'Home' }];
    const client = fakeClient({
      listarTelas: vi.fn().mockResolvedValue(telas),
      obterDesign: vi.fn().mockResolvedValue({ id: 'd1', sistema_id: 's1', nome: 'Home', arvore: raiz() } as Design),
    });
    renderComRota(client);
    fireEvent.click(await screen.findByRole('button', { name: /selecione uma tela/i }));
    await abrirTela('Home');

    fireEvent.click(screen.getByRole('button', { name: 'Foco' }));
    expect(screen.getByRole('button', { name: 'Foco' })).toHaveAttribute('aria-pressed', 'true');

    fireEvent.click(screen.getByRole('button', { name: 'Visualização' }));
    expect(screen.getByRole('button', { name: 'Visualização' })).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByRole('button', { name: 'Foco' })).toHaveAttribute('aria-pressed', 'false');
    expect(await screen.findByRole('dialog', { name: /visualização/i })).toBeInTheDocument();
  });

  it('a tecla F alterna o modo foco (fora de campos de texto)', async () => {
    const telas: DesignResumo[] = [{ id: 'd1', nome: 'Home' }];
    const client = fakeClient({
      listarTelas: vi.fn().mockResolvedValue(telas),
      obterDesign: vi.fn().mockResolvedValue({ id: 'd1', sistema_id: 's1', nome: 'Home', arvore: raiz() } as Design),
    });
    renderComRota(client);
    fireEvent.click(await screen.findByRole('button', { name: /selecione uma tela/i }));
    await abrirTela('Home');

    expect(screen.getByRole('tablist')).toBeInTheDocument();
    fireEvent.keyDown(window, { key: 'f' });
    expect(screen.queryByRole('tablist')).not.toBeInTheDocument();
  });

  it('abre e fecha a visualização do site', async () => {
    const arvoreComFilho = raiz([
      { blind_index: 'b1', tipo: 'heading', propriedades: { texto: 'Bem-vindo', estilos: {} }, componente_filhos: [] },
    ]);
    const telas: DesignResumo[] = [{ id: 'd1', nome: 'Home' }];
    const client = fakeClient({
      listarTelas: vi.fn().mockResolvedValue(telas),
      obterDesign: vi.fn().mockResolvedValue({ id: 'd1', sistema_id: 's1', nome: 'Home', arvore: arvoreComFilho } as Design),
    });
    entrarMock.mockImplementation(async (_id, handlers) => handlers.onJoin?.(arvoreComFilho));

    renderComRota(client);
    fireEvent.click(await screen.findByRole('button', { name: /selecione uma tela/i }));
    await abrirTela('Home');

    fireEvent.click(screen.getByRole('button', { name: 'Visualização' }));

    const dialog = await screen.findByRole('dialog', { name: /visualização/i });
    expect(within(dialog).getByRole('heading', { name: 'Bem-vindo' })).toBeInTheDocument();

    fireEvent.click(within(dialog).getByRole('button', { name: /fechar visualização/i }));
    expect(screen.queryByRole('dialog', { name: /visualização/i })).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Edição' })).toHaveAttribute('aria-pressed', 'true');
  });

  it('duplo clique no texto de um componente entra em edição inline (rich text)', async () => {
    const arvoreComFilho = raiz([
      { blind_index: 'b1', tipo: 'botao', propriedades: { texto: 'Entrar', estilos: {} }, componente_filhos: [] },
    ]);
    const telas: DesignResumo[] = [{ id: 'd1', nome: 'Home' }];
    const client = fakeClient({
      listarTelas: vi.fn().mockResolvedValue(telas),
      obterDesign: vi.fn().mockResolvedValue({ id: 'd1', sistema_id: 's1', nome: 'Home', arvore: arvoreComFilho } as Design),
    });
    entrarMock.mockImplementation(async (_id, handlers) => handlers.onJoin?.(arvoreComFilho));

    const { container } = renderComRota(client);
    fireEvent.click(await screen.findByRole('button', { name: /selecione uma tela/i }));
    await abrirTela('Home');

    expect(container.querySelector('[contenteditable="true"]')).not.toBeInTheDocument();

    fireEvent.doubleClick(await screen.findByText('Entrar'));
    const editavel = container.querySelector('[contenteditable="true"]');
    expect(editavel).toBeInTheDocument();
    expect(editavel).toHaveTextContent('Entrar');
  });

  it('sair da edição inline (blur) salva o texto sanitizado via update_props', async () => {
    const arvoreComFilho = raiz([
      { blind_index: 'b1', tipo: 'botao', propriedades: { texto: 'Entrar', estilos: {} }, componente_filhos: [] },
    ]);
    const telas: DesignResumo[] = [{ id: 'd1', nome: 'Home' }];
    const client = fakeClient({
      listarTelas: vi.fn().mockResolvedValue(telas),
      obterDesign: vi.fn().mockResolvedValue({ id: 'd1', sistema_id: 's1', nome: 'Home', arvore: arvoreComFilho } as Design),
    });
    entrarMock.mockImplementation(async (_id, handlers) => handlers.onJoin?.(arvoreComFilho));

    const { container } = renderComRota(client);
    fireEvent.click(await screen.findByRole('button', { name: /selecione uma tela/i }));
    await abrirTela('Home');

    fireEvent.doubleClick(await screen.findByText('Entrar'));
    const editavel = container.querySelector('[contenteditable="true"]')!;
    editavel.innerHTML = '<b>Entrar</b> agora <script>alert(1)</script>';
    fireEvent.blur(editavel);

    expect(mutarMock).toHaveBeenCalledWith(
      expect.objectContaining({
        tipo: 'update_props',
        blind_index: 'b1',
        propriedades: expect.objectContaining({ texto: '<b>Entrar</b> agora ' }),
      }),
    );
    expect(container.querySelector('[contenteditable="true"]')).not.toBeInTheDocument();
  });

  it('a alça de posicionamento livre só aparece em componentes selecionados com posição absolute', async () => {
    const arvoreComFilhos = raiz([
      { blind_index: 'b1', tipo: 'container', propriedades: { estilos: { posicao: 'relative' } }, componente_filhos: [] },
      {
        blind_index: 'b2',
        tipo: 'container',
        propriedades: { estilos: { posicao: 'absolute', x: 10, y: 20 } },
        componente_filhos: [],
      },
    ]);
    const telas: DesignResumo[] = [{ id: 'd1', nome: 'Home' }];
    const client = fakeClient({
      listarTelas: vi.fn().mockResolvedValue(telas),
      obterDesign: vi.fn().mockResolvedValue({ id: 'd1', sistema_id: 's1', nome: 'Home', arvore: arvoreComFilhos } as Design),
    });
    entrarMock.mockImplementation(async (_id, handlers) => handlers.onJoin?.(arvoreComFilhos));

    const { container } = renderComRota(client);
    fireEvent.click(await screen.findByRole('button', { name: /selecione uma tela/i }));
    await abrirTela('Home');
    const alca = () => container.querySelector('[aria-label="Mover container"]');
    // Usa data-testid (layer-<blind_index>, já existente em cada nó) em vez de
    // texto/role — "Container" também é o rótulo do botão da paleta, então
    // uma busca por nome ambígua pegaria o item errado.
    const noB1 = () => container.querySelector('[data-testid="layer-b1"] [role="button"]') as HTMLElement;
    const noB2 = () => container.querySelector('[data-testid="layer-b2"] [role="button"]') as HTMLElement;
    await waitFor(() => expect(noB1()).toBeInTheDocument());

    // Nenhum selecionado ainda: nenhuma alça de mover.
    expect(alca()).not.toBeInTheDocument();

    // Seleciona o container relative (b1): ainda sem alça (não é absolute).
    fireEvent.click(noB1());
    expect(alca()).not.toBeInTheDocument();

    // Seleciona o container absolute (b2): alça de mover aparece.
    fireEvent.click(noB2());
    expect(alca()).toBeInTheDocument();
  });

  it('componente com Display "inline-block" fica lado a lado com o vizinho no canvas', async () => {
    const arvoreComFilhos = raiz([
      { blind_index: 'b1', tipo: 'botao', propriedades: { texto: 'Um', estilos: { display: 'inline-block' } }, componente_filhos: [] },
      { blind_index: 'b2', tipo: 'botao', propriedades: { texto: 'Dois', estilos: { display: 'inline-block' } }, componente_filhos: [] },
    ]);
    const telas: DesignResumo[] = [{ id: 'd1', nome: 'Home' }];
    const client = fakeClient({
      listarTelas: vi.fn().mockResolvedValue(telas),
      obterDesign: vi.fn().mockResolvedValue({ id: 'd1', sistema_id: 's1', nome: 'Home', arvore: arvoreComFilhos } as Design),
    });
    entrarMock.mockImplementation(async (_id, handlers) => handlers.onJoin?.(arvoreComFilhos));

    const { container } = renderComRota(client);
    fireEvent.click(await screen.findByRole('button', { name: /selecione uma tela/i }));
    await abrirTela('Home');

    // O wrapper externo (data-testid=layer-*) é quem participa do fluxo do
    // canvas — se ele não espelhar o display "inline-block" do nó, os dois
    // botões voltam a empilhar mesmo com o CSS certo no elemento interno.
    const layerB1 = () => container.querySelector('[data-testid="layer-b1"]') as HTMLElement;
    const layerB2 = () => container.querySelector('[data-testid="layer-b2"]') as HTMLElement;
    await waitFor(() => expect(layerB1()).toBeInTheDocument());

    expect(layerB1().style.display).toBe('inline-block');
    expect(layerB2().style.display).toBe('inline-block');
  });

  it('exibe erro ao falhar a listagem de telas e recarrega com sucesso pelo botão "Tentar novamente"', async () => {
    const listarTelas = vi
      .fn()
      .mockRejectedValueOnce(new ApiError(500, 'INTERNAL', 'falha ao listar'))
      .mockResolvedValueOnce([{ id: 'd1', nome: 'Home' }]);
    const client = fakeClient({
      listarTelas,
      obterDesign: vi.fn().mockResolvedValue({ id: 'd1', sistema_id: 's1', nome: 'Home', arvore: raiz() } as Design),
    });
    renderComRota(client);

    const alerta = await screen.findByRole('alert');
    expect(within(alerta).getByText(/não foi possível carregar as telas/i)).toBeInTheDocument();

    fireEvent.click(within(alerta).getByRole('button', { name: /tentar novamente/i }));

    await waitFor(() => expect(screen.getByRole('button', { name: /selecione uma tela/i })).toBeInTheDocument());
    expect(listarTelas).toHaveBeenCalledTimes(2);
  });

  it('tela criada com sucesso mas recarga da lista falha em seguida: mostra erro (telas some da UI, RN não coberta hoje)', async () => {
    // useTelas.recarregar() sempre passa por "carregando" (telas=[]) antes do
    // resultado — se o listarTelas() de recarga falhar, a fase "erro" também
    // carrega telas=[], derrubando o editor mesmo que a tela recém-criada
    // exista no backend. Este teste documenta o comportamento atual (não uma
    // correção) para não regredir silenciosamente.
    const criarDesign = vi.fn().mockResolvedValue({ id: 'd-novo' });
    const listarTelas = vi
      .fn()
      .mockResolvedValueOnce([])
      .mockRejectedValueOnce(new ApiError(500, 'INTERNAL', 'falha ao recarregar'));
    const client = fakeClient({ listarTelas, criarDesign });
    renderComRota(client);
    await waitFor(() => expect(screen.getByText(/nenhuma tela criada ainda/i)).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /selecione uma tela/i }));
    fireEvent.click(screen.getByRole('button', { name: /nova tela/i }));
    const modal = await screen.findByRole('dialog', { name: /nova tela/i });
    fireEvent.change(within(modal).getByLabelText(/nome da tela/i), { target: { value: 'Home' } });
    fireEvent.click(within(modal).getByRole('button', { name: /criar tela/i }));

    await waitFor(() => expect(criarDesign).toHaveBeenCalled());
    const alerta = await screen.findByRole('alert');
    expect(within(alerta).getByText(/não foi possível carregar as telas/i)).toBeInTheDocument();
  });

  it('o botão "Criar tela" fica desabilitado com o nome vazio ou só espaços', async () => {
    const client = fakeClient({ listarTelas: vi.fn().mockResolvedValue([]) });
    renderComRota(client);
    await waitFor(() => expect(screen.getByText(/nenhuma tela criada ainda/i)).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /selecione uma tela/i }));
    fireEvent.click(screen.getByRole('button', { name: /nova tela/i }));
    const modal = await screen.findByRole('dialog', { name: /nova tela/i });
    const botaoCriar = within(modal).getByRole('button', { name: /criar tela/i });
    const input = within(modal).getByLabelText(/nome da tela/i);

    expect(botaoCriar).toBeDisabled();

    fireEvent.change(input, { target: { value: '   ' } });
    expect(botaoCriar).toBeDisabled();

    fireEvent.change(input, { target: { value: 'Home' } });
    expect(botaoCriar).not.toBeDisabled();
  });

  it('mantém o modal aberto e reabilita o botão quando a criação da tela falha', async () => {
    const criarDesign = vi.fn().mockRejectedValue(new ApiError(500, 'INTERNAL', 'falha ao criar'));
    const client = fakeClient({ listarTelas: vi.fn().mockResolvedValue([]), criarDesign });
    renderComRota(client);
    await waitFor(() => expect(screen.getByText(/nenhuma tela criada ainda/i)).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /selecione uma tela/i }));
    fireEvent.click(screen.getByRole('button', { name: /nova tela/i }));
    const modal = await screen.findByRole('dialog', { name: /nova tela/i });
    fireEvent.change(within(modal).getByLabelText(/nome da tela/i), { target: { value: 'Home' } });
    const botaoCriar = within(modal).getByRole('button', { name: /criar tela/i });

    fireEvent.click(botaoCriar);
    await waitFor(() => expect(criarDesign).toHaveBeenCalled());

    // A falha não deve fechar o modal silenciosamente, nem deixar o botão
    // preso em "Criando…", nem esconder o motivo do erro do usuário.
    await waitFor(() => expect(screen.getByRole('dialog', { name: /nova tela/i })).toBeInTheDocument());
    await waitFor(() => expect(within(modal).getByRole('button', { name: /criar tela/i })).not.toBeDisabled());
    expect(within(modal).getByRole('alert')).toHaveTextContent('falha ao criar');
  });

  it('exibe erro ao falhar o carregamento de uma tela específica, mesmo com a lista de telas ok', async () => {
    const telas: DesignResumo[] = [{ id: 'd1', nome: 'Home' }];
    const client = fakeClient({
      listarTelas: vi.fn().mockResolvedValue(telas),
      obterDesign: vi.fn().mockRejectedValue(new ApiError(404, 'NOT_FOUND', 'tela não encontrada')),
    });
    renderComRota(client);

    fireEvent.click(await screen.findByRole('button', { name: /selecione uma tela/i }));
    await abrirTela('Home');

    const alerta = await screen.findByRole('alert');
    expect(within(alerta).getByText(/não foi possível carregar a tela/i)).toBeInTheDocument();
    // O topbar (troca de tela/undo/redo) continua acessível mesmo com o
    // canvas em erro — só a área central é substituída pelo ErrorState.
    expect(screen.getByRole('button', { name: /^Home$/ })).toBeInTheDocument();
  });
});
