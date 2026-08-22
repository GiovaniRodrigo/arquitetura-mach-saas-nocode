import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import { Canvas, type CanvasProps } from './Canvas';
import type { Componente } from '../../../api/types';

function no(overrides: Partial<Componente> & { blind_index: string; tipo: string }): Componente {
  return { componente_filhos: [], propriedades: {}, ...overrides };
}

function noAbsoluto(
  blindIndex: string,
  tipo: string,
  x: number,
  y: number,
  extra: Record<string, unknown> = {},
): Componente {
  return no({
    blind_index: blindIndex,
    tipo,
    propriedades: { estilos: { posicao: 'absolute', x, y, ...extra } },
  });
}

function raiz(filhos: Componente[]): Componente {
  return { blind_index: 'root', tipo: 'tela', componente_filhos: filhos };
}

function propsPadrao(overrides: Partial<CanvasProps> = {}): CanvasProps {
  return {
    arvore: raiz([]),
    selecionado: null,
    onSelecionar: vi.fn(),
    onAdicionar: vi.fn(),
    onMover: vi.fn(),
    onResize: vi.fn(),
    onEditarTexto: vi.fn(),
    onMoverAbsoluto: vi.fn(),
    larguraDevice: 1440,
    zoom: 1,
    ...overrides,
  };
}

/** jsdom não implementa `PointerEvent` — mas os listeners nativos do React
 * (e o `window.addEventListener('pointermove'/'pointerup', ...)` que o
 * Canvas registra durante o arraste) casam pelo `type` string do evento, não
 * pela classe. `MouseEvent`, ao contrário do `Event` genérico que o
 * `fireEvent.pointerDown` do testing-library cai de volta em jsdom, aceita
 * `clientX`/`clientY` no init dict — por isso disparamos `MouseEvent`s
 * tipados como eventos de ponteiro em vez de `fireEvent.pointerDown/...`. */
function pointerEvent(tipo: string, clientX: number, clientY: number): Event {
  return new MouseEvent(tipo, { clientX, clientY, bubbles: true });
}

function arrastar(alca: Element, de: { x: number; y: number }, para: { x: number; y: number }) {
  fireEvent(alca, pointerEvent('pointerdown', de.x, de.y));
  fireEvent(window, pointerEvent('pointermove', para.x, para.y));
  fireEvent(window, pointerEvent('pointerup', para.x, para.y));
}

describe('Canvas — renderização de componentes (RF09)', () => {
  it('renderiza cada componente da árvore pelo seu tipo', () => {
    const arvore = raiz([
      no({ blind_index: 'b1', tipo: 'heading', propriedades: { texto: 'Título', estilos: {} } }),
      no({ blind_index: 'b2', tipo: 'imagem' }),
      no({ blind_index: 'b3', tipo: 'input' }),
    ]);
    render(<Canvas {...propsPadrao({ arvore })} />);

    expect(screen.getByRole('button', { name: 'heading' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'imagem' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'input' })).toBeInTheDocument();
  });

  it('mostra a mensagem de tela em branco quando não há componentes', () => {
    render(<Canvas {...propsPadrao({ arvore: raiz([]) })} />);
    expect(screen.getByText(/tela em branco/i)).toBeInTheDocument();
  });

  it('renderiza componentes aninhados dentro de containers', () => {
    const arvore = raiz([
      no({
        blind_index: 'pai',
        tipo: 'container',
        componente_filhos: [no({ blind_index: 'filho', tipo: 'botao', propriedades: { texto: 'Ok', estilos: {} } })],
      }),
    ]);
    render(<Canvas {...propsPadrao({ arvore })} />);

    expect(screen.getByTestId('layer-pai')).toBeInTheDocument();
    expect(screen.getByTestId('layer-filho')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'botao' })).toBeInTheDocument();
  });

  it('clicar em um componente seleciona-o; clicar no fundo do canvas limpa a seleção', () => {
    const onSelecionar = vi.fn();
    const arvore = raiz([no({ blind_index: 'b1', tipo: 'heading', propriedades: { texto: 'X', estilos: {} } })]);
    render(<Canvas {...propsPadrao({ arvore, onSelecionar })} />);

    fireEvent.click(screen.getByRole('button', { name: 'heading' }));
    expect(onSelecionar).toHaveBeenCalledWith('b1');

    fireEvent.click(screen.getByRole('presentation', { name: 'Canvas' }));
    expect(onSelecionar).toHaveBeenLastCalledWith(null);
  });
});

describe('Canvas — posicionamento livre dos componentes (RF09, posicao: absolute)', () => {
  it('só mostra a alça de mover quando o componente está selecionado e com posição absoluta', () => {
    const arvore = raiz([noAbsoluto('b1', 'imagem', 0, 0)]);
    const { rerender } = render(<Canvas {...propsPadrao({ arvore, selecionado: null })} />);
    expect(screen.queryByRole('presentation', { name: 'Mover imagem' })).not.toBeInTheDocument();

    rerender(<Canvas {...propsPadrao({ arvore, selecionado: 'b1' })} />);
    expect(screen.getByRole('presentation', { name: 'Mover imagem' })).toBeInTheDocument();

    const arvoreRelativa = raiz([no({ blind_index: 'b2', tipo: 'imagem' })]);
    rerender(<Canvas {...propsPadrao({ arvore: arvoreRelativa, selecionado: 'b2' })} />);
    expect(screen.queryByRole('presentation', { name: 'Mover imagem' })).not.toBeInTheDocument();
    expect(screen.getByRole('presentation', { name: 'Redimensionar imagem' })).toBeInTheDocument();
  });

  it('arrastar a alça de mover reposiciona o componente pelo delta do ponteiro (não trava em grade/fluxo)', () => {
    const onMoverAbsoluto = vi.fn();
    const arvore = raiz([noAbsoluto('b1', 'imagem', 10, 10)]);
    render(<Canvas {...propsPadrao({ arvore, selecionado: 'b1', onMoverAbsoluto })} />);

    const alca = screen.getByRole('presentation', { name: 'Mover imagem' });
    arrastar(alca, { x: 100, y: 100 }, { x: 137, y: 163 });

    expect(onMoverAbsoluto).toHaveBeenCalledWith('b1', 10 + 37, 10 + 63);
  });

  it('permite posicionar em qualquer lugar, inclusive fora da área visível (coordenadas negativas, sem clamping)', () => {
    const onMoverAbsoluto = vi.fn();
    const arvore = raiz([noAbsoluto('b1', 'card', 0, 0)]);
    render(<Canvas {...propsPadrao({ arvore, selecionado: 'b1', onMoverAbsoluto })} />);

    const alca = screen.getByRole('presentation', { name: 'Mover card' });
    arrastar(alca, { x: 200, y: 200 }, { x: -150, y: -400 });

    // Diferente do resize (que impõe mínimo de 20px), a posição livre não tem
    // limite algum — o componente pode ser movido para qualquer coordenada.
    expect(onMoverAbsoluto).toHaveBeenCalledWith('b1', -350, -600);
  });

  it('atualiza a posição em tela durante o arraste (preview), só confirmando via onMoverAbsoluto no pointerup', () => {
    const onMoverAbsoluto = vi.fn();
    const arvore = raiz([noAbsoluto('b1', 'imagem', 0, 0)]);
    render(<Canvas {...propsPadrao({ arvore, selecionado: 'b1', onMoverAbsoluto })} />);

    const alca = screen.getByRole('presentation', { name: 'Mover imagem' });
    const componente = screen.getByRole('button', { name: 'imagem' });

    fireEvent(alca, pointerEvent('pointerdown', 0, 0));
    fireEvent(window, pointerEvent('pointermove', 50, 20));

    expect(componente).toHaveStyle({ left: '50px', top: '20px' });
    expect(onMoverAbsoluto).not.toHaveBeenCalled();

    fireEvent(window, pointerEvent('pointerup', 50, 20));
    expect(onMoverAbsoluto).toHaveBeenCalledWith('b1', 50, 20);
  });

  it('compensa o zoom do canvas ao calcular a nova posição', () => {
    const onMoverAbsoluto = vi.fn();
    const arvore = raiz([noAbsoluto('b1', 'imagem', 0, 0)]);

    const { unmount } = render(<Canvas {...propsPadrao({ arvore, selecionado: 'b1', onMoverAbsoluto, zoom: 2 })} />);
    arrastar(screen.getByRole('presentation', { name: 'Mover imagem' }), { x: 0, y: 0 }, { x: 60, y: 40 });
    expect(onMoverAbsoluto).toHaveBeenCalledWith('b1', 30, 20);
    unmount();

    onMoverAbsoluto.mockClear();
    render(<Canvas {...propsPadrao({ arvore, selecionado: 'b1', onMoverAbsoluto, zoom: 0.5 })} />);
    arrastar(screen.getByRole('presentation', { name: 'Mover imagem' }), { x: 0, y: 0 }, { x: 60, y: 40 });
    expect(onMoverAbsoluto).toHaveBeenCalledWith('b1', 120, 80);
  });

  it('componente com posição absoluta não é arrastável pelo drag&drop nativo (evita conflito com a alça de mover)', () => {
    const arvore = raiz([noAbsoluto('b1', 'imagem', 0, 0)]);
    render(<Canvas {...propsPadrao({ arvore, selecionado: 'b1' })} />);
    expect(screen.getByRole('button', { name: 'imagem' })).toHaveAttribute('draggable', 'false');
  });
});

describe('Canvas — paridade de display com o site publicado', () => {
  // Regressão: o Canvas desenha todo nó como <div> e forçava display:block em
  // tudo que não aceita filhos. Um `badge` (que o PreviewRenderer publica como
  // <span>, inline) virava então uma barra de largura total no editor — ex.: o
  // "MAIS POPULAR" da home demo, pílula no publicado e barra na edição.
  function displayDe(blindIndex: string): string {
    const el = screen
      .getByTestId(`layer-${blindIndex}`)
      .querySelector('[role="button"]') as HTMLElement;
    return el.style.display;
  }

  it('desenha badge como inline-block, para encolher até o conteúdo igual ao publicado', () => {
    const arvore = raiz([
      no({
        blind_index: 'selo',
        tipo: 'badge',
        propriedades: { texto: 'MAIS POPULAR', estilos: { padding: '4px 12px', bordaRaio: '999px' } },
      }),
    ]);
    render(<Canvas {...propsPadrao({ arvore })} />);
    expect(displayDe('selo')).toBe('inline-block');
  });

  it('mantém block em componentes que o publicado renderiza como bloco', () => {
    const arvore = raiz([
      no({ blind_index: 'texto', tipo: 'paragrafo', propriedades: { texto: 'Olá', estilos: {} } }),
    ]);
    render(<Canvas {...propsPadrao({ arvore })} />);
    expect(displayDe('texto')).toBe('block');
  });

  it('respeita um display explícito em vez do fallback por tipo', () => {
    const arvore = raiz([
      no({ blind_index: 'selo', tipo: 'badge', propriedades: { texto: 'X', estilos: { display: 'block' } } }),
    ]);
    render(<Canvas {...propsPadrao({ arvore })} />);
    expect(displayDe('selo')).toBe('block');
  });
});
