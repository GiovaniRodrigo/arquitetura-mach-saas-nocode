import { useState } from 'react';
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, within, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { Inspector, parseMedida } from './Inspector';
import type { Componente } from '../../../api/types';

/** Inspector é controlado pelo pai (não guarda estado); este wrapper simula o
 * ciclo real (useCanvasDesign) para testar que a UI reflete a mudança aplicada. */
function InspectorControlado({ inicial }: { inicial: Componente }) {
  const [componente, setComponente] = useState(inicial);
  return (
    <Inspector
      componente={componente}
      onAtualizar={(_bi, propriedades) => setComponente((c) => ({ ...c, propriedades }))}
      onRemover={vi.fn()}
      onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()}
    />
  );
}

describe('parseMedida', () => {
  it('reconhece número + unidade', () => {
    expect(parseMedida('200px')).toEqual({ numero: '200', unidade: 'px' });
    expect(parseMedida('50%')).toEqual({ numero: '50', unidade: '%' });
    expect(parseMedida('1.5rem')).toEqual({ numero: '1.5', unidade: 'rem' });
  });

  it('trata ausente/"auto" como auto', () => {
    expect(parseMedida(undefined)).toEqual({ numero: '', unidade: 'auto' });
    expect(parseMedida('auto')).toEqual({ numero: '', unidade: 'auto' });
  });

  it('cai para px em valor não reconhecido', () => {
    expect(parseMedida('calc(100% - 8px)')).toEqual({ numero: '', unidade: 'px' });
  });
});

describe('Inspector — campo de medida (unidade como select)', () => {
  function componenteFake(): Componente {
    return {
      blind_index: 'b1',
      tipo: 'container',
      propriedades: { estilos: { largura: '200px' } },
      componente_filhos: [],
    };
  }

  it('largura é editada por número + select de unidade, não texto livre', () => {
    const onAtualizar = vi.fn();
    render(
      <Inspector componente={componenteFake()} onAtualizar={onAtualizar} onRemover={vi.fn()} onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()} />,
    );

    const campo = screen.getByText('Largura').closest('label')!;
    expect(within(campo).getByRole('spinbutton')).toHaveValue(200);
    expect(within(campo).getByRole('combobox')).toHaveValue('px');

    fireEvent.change(within(campo).getByRole('spinbutton'), { target: { value: '300' } });
    expect(onAtualizar).toHaveBeenCalledWith(
      'b1',
      expect.objectContaining({ estilos: expect.objectContaining({ largura: '300px' }) }),
    );
  });

  it('trocar a unidade para auto esconde o número e aplica "auto"', () => {
    render(<InspectorControlado inicial={componenteFake()} />);

    const campo = screen.getByText('Largura').closest('label')!;
    fireEvent.change(within(campo).getByRole('combobox'), { target: { value: 'auto' } });

    expect(within(campo).getByRole('combobox')).toHaveValue('auto');
    expect(within(campo).queryByRole('spinbutton')).not.toBeInTheDocument();
  });

  it('trocar a unidade de px para % recompõe o valor com o número atual', () => {
    const onAtualizar = vi.fn();
    render(
      <Inspector componente={componenteFake()} onAtualizar={onAtualizar} onRemover={vi.fn()} onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()} />,
    );

    const campo = screen.getByText('Largura').closest('label')!;
    fireEvent.change(within(campo).getByRole('combobox'), { target: { value: '%' } });

    expect(onAtualizar).toHaveBeenCalledWith(
      'b1',
      expect.objectContaining({ estilos: expect.objectContaining({ largura: '200%' }) }),
    );
  });

  it('campos sem unidade "auto" aplicável (ex.: tamanho da fonte) não oferecem auto', () => {
    render(
      <Inspector
        componente={{
          blind_index: 'b1',
          tipo: 'heading',
          propriedades: { texto: 'X', estilos: { fonteTamanho: '16px' } },
          componente_filhos: [],
        }}
        onAtualizar={vi.fn()}
        onRemover={vi.fn()}
        onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()}
      />,
    );
    const campo = screen.getByText('Tamanho da fonte').closest('label')!;
    const opcoes = within(campo)
      .getAllByRole('option')
      .map((o) => (o as HTMLOptionElement).value);
    expect(opcoes).not.toContain('auto');
  });
});

describe('Inspector — componente Imagem (upload/URL)', () => {
  function imagemFake(): Componente {
    return { blind_index: 'b1', tipo: 'imagem', propriedades: { estilos: {} }, componente_filhos: [] };
  }

  it('define a imagem por URL', () => {
    const onAtualizar = vi.fn();
    render(
      <Inspector componente={imagemFake()} onAtualizar={onAtualizar} onRemover={vi.fn()} onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()} />,
    );
    const campo = screen.getByText('ou URL da imagem').closest('label')!;
    fireEvent.change(within(campo).getByRole('textbox'), { target: { value: 'https://exemplo.com/foto.png' } });
    expect(onAtualizar).toHaveBeenCalledWith('b1', expect.objectContaining({ src: 'https://exemplo.com/foto.png' }));
  });

  it('define o texto alternativo', () => {
    const onAtualizar = vi.fn();
    render(
      <Inspector componente={imagemFake()} onAtualizar={onAtualizar} onRemover={vi.fn()} onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()} />,
    );
    const campo = screen.getByText('Texto alternativo').closest('label')!;
    fireEvent.change(within(campo).getByRole('textbox'), { target: { value: 'Logo da empresa' } });
    expect(onAtualizar).toHaveBeenCalledWith('b1', expect.objectContaining({ alt: 'Logo da empresa' }));
  });

  it('envia um arquivo pequeno e aplica como data URL', async () => {
    const onAtualizar = vi.fn();
    render(
      <Inspector componente={imagemFake()} onAtualizar={onAtualizar} onRemover={vi.fn()} onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()} />,
    );
    const input = screen.getByLabelText('Enviar arquivo') as HTMLInputElement;
    const arquivo = new File(['conteudo-fake'], 'foto.png', { type: 'image/png' });
    fireEvent.change(input, { target: { files: [arquivo] } });

    await waitFor(() =>
      expect(onAtualizar).toHaveBeenCalledWith('b1', expect.objectContaining({ src: expect.stringMatching(/^data:/) })),
    );
  });

  it('rejeita arquivo maior que 2MB com mensagem de erro, sem aplicar', () => {
    const onAtualizar = vi.fn();
    render(
      <Inspector componente={imagemFake()} onAtualizar={onAtualizar} onRemover={vi.fn()} onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()} />,
    );
    const input = screen.getByLabelText('Enviar arquivo') as HTMLInputElement;
    const grande = new File([new Uint8Array(2 * 1024 * 1024 + 1)], 'grande.png', { type: 'image/png' });
    fireEvent.change(input, { target: { files: [grande] } });

    expect(screen.getByText(/muito grande/i)).toBeInTheDocument();
    expect(onAtualizar).not.toHaveBeenCalled();
  });

  it('a seção Imagem não aparece para outros tipos de componente', () => {
    render(
      <Inspector
        componente={{ blind_index: 'b1', tipo: 'container', propriedades: { estilos: {} }, componente_filhos: [] }}
        onAtualizar={vi.fn()}
        onRemover={vi.fn()}
        onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()}
      />,
    );
    expect(screen.queryByText('Enviar arquivo')).not.toBeInTheDocument();
  });
});

describe('Inspector — componente Carrossel (slides/autoplay)', () => {
  function carrosselFake(propriedades: Record<string, unknown> = {}): Componente {
    return { blind_index: 'b1', tipo: 'carrossel', propriedades: { estilos: {}, ...propriedades }, componente_filhos: [] };
  }

  it('envia múltiplos arquivos e acrescenta como slides', async () => {
    const onAtualizar = vi.fn();
    render(
      <Inspector componente={carrosselFake()} onAtualizar={onAtualizar} onRemover={vi.fn()} onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()} />,
    );
    const input = screen.getByLabelText('Adicionar imagens') as HTMLInputElement;
    const arquivos = [
      new File(['a'], 'a.png', { type: 'image/png' }),
      new File(['b'], 'b.png', { type: 'image/png' }),
    ];
    fireEvent.change(input, { target: { files: arquivos } });

    await waitFor(() =>
      expect(onAtualizar).toHaveBeenCalledWith(
        'b1',
        expect.objectContaining({
          imagens: [
            expect.objectContaining({ src: expect.stringMatching(/^data:/) }),
            expect.objectContaining({ src: expect.stringMatching(/^data:/) }),
          ],
        }),
      ),
    );
  });

  it('rejeita arquivo maior que 2MB com mensagem de erro, sem aplicar', () => {
    const onAtualizar = vi.fn();
    render(
      <Inspector componente={carrosselFake()} onAtualizar={onAtualizar} onRemover={vi.fn()} onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()} />,
    );
    const input = screen.getByLabelText('Adicionar imagens') as HTMLInputElement;
    const grande = new File([new Uint8Array(2 * 1024 * 1024 + 1)], 'grande.png', { type: 'image/png' });
    fireEvent.change(input, { target: { files: [grande] } });

    expect(screen.getByText(/muito grande/i)).toBeInTheDocument();
    expect(onAtualizar).not.toHaveBeenCalled();
  });

  it('remove um slide existente', () => {
    const onAtualizar = vi.fn();
    render(
      <Inspector
        componente={carrosselFake({ imagens: [{ src: 'data:a' }, { src: 'data:b' }] })}
        onAtualizar={onAtualizar}
        onRemover={vi.fn()}
        onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()}
      />,
    );
    fireEvent.click(screen.getByRole('button', { name: 'Remover slide 1' }));
    expect(onAtualizar).toHaveBeenCalledWith('b1', expect.objectContaining({ imagens: [{ src: 'data:b' }] }));
  });

  it('liga autoplay e revela o campo de intervalo', () => {
    render(<InspectorControlado inicial={carrosselFake()} />);
    expect(screen.queryByText('Intervalo (ms)')).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('checkbox', { name: 'Autoplay' }));
    expect(screen.getByText('Intervalo (ms)')).toBeInTheDocument();
  });

  it('desliga mostrar setas/pontos', () => {
    const onAtualizar = vi.fn();
    render(
      <Inspector componente={carrosselFake()} onAtualizar={onAtualizar} onRemover={vi.fn()} onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()} />,
    );
    fireEvent.click(screen.getByRole('checkbox', { name: 'Mostrar setas' }));
    expect(onAtualizar).toHaveBeenCalledWith('b1', expect.objectContaining({ mostrarSetas: false }));
    fireEvent.click(screen.getByRole('checkbox', { name: 'Mostrar pontos' }));
    expect(onAtualizar).toHaveBeenCalledWith('b1', expect.objectContaining({ mostrarPontos: false }));
  });

  it('a seção Carrossel não aparece para outros tipos de componente', () => {
    render(
      <Inspector
        componente={{ blind_index: 'b1', tipo: 'imagem', propriedades: { estilos: {} }, componente_filhos: [] }}
        onAtualizar={vi.fn()}
        onRemover={vi.fn()}
        onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()}
      />,
    );
    expect(screen.queryByText('Adicionar imagens')).not.toBeInTheDocument();
  });
});

describe('Inspector — componente Menu (itens/links)', () => {
  function menuFake(propriedades: Record<string, unknown> = {}): Componente {
    return { blind_index: 'b1', tipo: 'menu', propriedades: { estilos: {}, ...propriedades }, componente_filhos: [] };
  }

  it('adiciona um item ao clicar em "+ Adicionar item"', () => {
    const onAtualizar = vi.fn();
    render(<Inspector componente={menuFake()} onAtualizar={onAtualizar} onRemover={vi.fn()} onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()} />);
    fireEvent.click(screen.getByRole('button', { name: /adicionar item/i }));
    expect(onAtualizar).toHaveBeenCalledWith(
      'b1',
      expect.objectContaining({ itens: [{ label: 'Novo item', url: '#' }] }),
    );
  });

  it('edita o rótulo e a URL de um item existente', () => {
    const onAtualizar = vi.fn();
    render(
      <Inspector
        componente={menuFake({ itens: [{ label: 'Home', url: '/' }] })}
        onAtualizar={onAtualizar}
        onRemover={vi.fn()}
        onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()}
      />,
    );
    fireEvent.change(screen.getByLabelText('Rótulo do item 1'), { target: { value: 'Sobre' } });
    expect(onAtualizar).toHaveBeenCalledWith('b1', expect.objectContaining({ itens: [{ label: 'Sobre', url: '/' }] }));

    fireEvent.change(screen.getByLabelText('URL do item 1'), { target: { value: '/sobre' } });
    expect(onAtualizar).toHaveBeenCalledWith('b1', expect.objectContaining({ itens: [{ label: 'Home', url: '/sobre' }] }));
  });

  it('remove um item existente', () => {
    const onAtualizar = vi.fn();
    render(
      <Inspector
        componente={menuFake({ itens: [{ label: 'Home', url: '/' }, { label: 'Sobre', url: '/sobre' }] })}
        onAtualizar={onAtualizar}
        onRemover={vi.fn()}
        onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()}
      />,
    );
    fireEvent.click(screen.getByRole('button', { name: 'Remover item 1' }));
    expect(onAtualizar).toHaveBeenCalledWith('b1', expect.objectContaining({ itens: [{ label: 'Sobre', url: '/sobre' }] }));
  });

  it('a seção Menu não aparece para outros tipos de componente', () => {
    render(
      <Inspector
        componente={{ blind_index: 'b1', tipo: 'container', propriedades: { estilos: {} }, componente_filhos: [] }}
        onAtualizar={vi.fn()}
        onRemover={vi.fn()}
        onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()}
      />,
    );
    expect(screen.queryByText(/adicionar item/i)).not.toBeInTheDocument();
  });
});

describe('Inspector — componente Accordion (painéis)', () => {
  function accordionFake(propriedades: Record<string, unknown> = {}): Componente {
    return {
      blind_index: 'b1',
      tipo: 'accordion',
      propriedades: { estilos: {}, ...propriedades },
      componente_filhos: [],
    };
  }

  it('adiciona um painel ao clicar em "+ Adicionar painel"', () => {
    const onAtualizar = vi.fn();
    render(
      <Inspector componente={accordionFake()} onAtualizar={onAtualizar} onRemover={vi.fn()} onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()} />,
    );
    fireEvent.click(screen.getByRole('button', { name: /adicionar painel/i }));
    expect(onAtualizar).toHaveBeenCalledWith(
      'b1',
      expect.objectContaining({ itens: [{ titulo: 'Novo painel', conteudo: '' }] }),
    );
  });

  it('edita o título e o conteúdo de um painel existente', () => {
    const onAtualizar = vi.fn();
    render(
      <Inspector
        componente={accordionFake({ itens: [{ titulo: 'Pergunta 1', conteudo: '' }] })}
        onAtualizar={onAtualizar}
        onRemover={vi.fn()}
        onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()}
      />,
    );
    fireEvent.change(screen.getByLabelText('Título do painel 1'), { target: { value: 'Pergunta 2' } });
    expect(onAtualizar).toHaveBeenCalledWith(
      'b1',
      expect.objectContaining({ itens: [{ titulo: 'Pergunta 2', conteudo: '' }] }),
    );

    fireEvent.change(screen.getByLabelText('Conteúdo do painel 1'), { target: { value: 'Resposta' } });
    expect(onAtualizar).toHaveBeenCalledWith(
      'b1',
      expect.objectContaining({ itens: [{ titulo: 'Pergunta 1', conteudo: 'Resposta' }] }),
    );
  });

  it('remove um painel existente', () => {
    const onAtualizar = vi.fn();
    render(
      <Inspector
        componente={accordionFake({ itens: [{ titulo: 'A', conteudo: '' }, { titulo: 'B', conteudo: '' }] })}
        onAtualizar={onAtualizar}
        onRemover={vi.fn()}
        onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()}
      />,
    );
    fireEvent.click(screen.getByRole('button', { name: 'Remover painel 1' }));
    expect(onAtualizar).toHaveBeenCalledWith('b1', expect.objectContaining({ itens: [{ titulo: 'B', conteudo: '' }] }));
  });

  it('a seção Accordion não aparece para outros tipos de componente', () => {
    render(
      <Inspector
        componente={{ blind_index: 'b1', tipo: 'container', propriedades: { estilos: {} }, componente_filhos: [] }}
        onAtualizar={vi.fn()}
        onRemover={vi.fn()}
        onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()}
      />,
    );
    expect(screen.queryByText(/adicionar painel/i)).not.toBeInTheDocument();
  });
});

describe('Inspector — componente Vídeo (URL)', () => {
  it('define a URL do vídeo', () => {
    const onAtualizar = vi.fn();
    render(
      <Inspector
        componente={{ blind_index: 'b1', tipo: 'video', propriedades: { estilos: {} }, componente_filhos: [] }}
        onAtualizar={onAtualizar}
        onRemover={vi.fn()}
        onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()}
      />,
    );
    fireEvent.change(screen.getByRole('textbox', { name: /url/i }), {
      target: { value: 'https://youtu.be/abc123' },
    });
    expect(onAtualizar).toHaveBeenCalledWith('b1', expect.objectContaining({ src: 'https://youtu.be/abc123' }));
  });
});

describe('Inspector — componente Tabs (abas)', () => {
  it('adiciona, edita e remove uma aba', () => {
    const onAtualizar = vi.fn();
    render(
      <Inspector
        componente={{ blind_index: 'b1', tipo: 'tabs', propriedades: { estilos: {} }, componente_filhos: [] }}
        onAtualizar={onAtualizar}
        onRemover={vi.fn()}
        onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()}
      />,
    );
    fireEvent.click(screen.getByRole('button', { name: /adicionar aba/i }));
    expect(onAtualizar).toHaveBeenCalledWith(
      'b1',
      expect.objectContaining({ itens: [{ titulo: 'Nova aba', conteudo: '' }] }),
    );
  });

  it('remove uma aba existente', () => {
    const onAtualizar = vi.fn();
    render(
      <Inspector
        componente={{
          blind_index: 'b1',
          tipo: 'tabs',
          propriedades: { estilos: {}, itens: [{ titulo: 'A', conteudo: '' }, { titulo: 'B', conteudo: '' }] },
          componente_filhos: [],
        }}
        onAtualizar={onAtualizar}
        onRemover={vi.fn()}
        onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()}
      />,
    );
    fireEvent.click(screen.getByRole('button', { name: 'Remover aba 1' }));
    expect(onAtualizar).toHaveBeenCalledWith('b1', expect.objectContaining({ itens: [{ titulo: 'B', conteudo: '' }] }));
  });
});

describe('Inspector — componente Avaliação (estrelas)', () => {
  it('clicar numa estrela define o valor', () => {
    const onAtualizar = vi.fn();
    render(
      <Inspector
        componente={{ blind_index: 'b1', tipo: 'avaliacao', propriedades: { estilos: {} }, componente_filhos: [] }}
        onAtualizar={onAtualizar}
        onRemover={vi.fn()}
        onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()}
      />,
    );
    fireEvent.click(screen.getByRole('button', { name: '4 estrelas' }));
    expect(onAtualizar).toHaveBeenCalledWith('b1', expect.objectContaining({ valor: 4 }));
  });

  it('clicar na mesma estrela já selecionada zera o valor', () => {
    const onAtualizar = vi.fn();
    render(
      <Inspector
        componente={{
          blind_index: 'b1',
          tipo: 'avaliacao',
          propriedades: { estilos: {}, valor: 3 },
          componente_filhos: [],
        }}
        onAtualizar={onAtualizar}
        onRemover={vi.fn()}
        onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()}
      />,
    );
    fireEvent.click(screen.getByRole('button', { name: '3 estrelas' }));
    expect(onAtualizar).toHaveBeenCalledWith('b1', expect.objectContaining({ valor: 0 }));
  });
});

describe('Inspector — componente Ícone (catálogo fechado)', () => {
  it('troca o ícone selecionado', () => {
    const onAtualizar = vi.fn();
    render(
      <Inspector
        componente={{ blind_index: 'b1', tipo: 'icone', propriedades: { estilos: {} }, componente_filhos: [] }}
        onAtualizar={onAtualizar}
        onRemover={vi.fn()}
        onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()}
      />,
    );
    fireEvent.change(screen.getByRole('combobox', { name: /ícone/i }), { target: { value: 'Heart' } });
    expect(onAtualizar).toHaveBeenCalledWith('b1', expect.objectContaining({ icone: 'Heart' }));
  });
});

describe('Inspector — componente Progresso (percentual)', () => {
  it('arrastar o slider define o valor', () => {
    const onAtualizar = vi.fn();
    render(
      <Inspector
        componente={{ blind_index: 'b1', tipo: 'progresso', propriedades: { estilos: {} }, componente_filhos: [] }}
        onAtualizar={onAtualizar}
        onRemover={vi.fn()}
        onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()}
      />,
    );
    fireEvent.change(screen.getByRole('slider', { name: /valor/i }), { target: { value: '80' } });
    expect(onAtualizar).toHaveBeenCalledWith('b1', expect.objectContaining({ valor: 80 }));
  });
});

describe('Inspector — componente Breadcrumb (itens)', () => {
  it('adiciona um item ao clicar em "+ Adicionar item"', () => {
    const onAtualizar = vi.fn();
    render(
      <Inspector
        componente={{ blind_index: 'b1', tipo: 'breadcrumb', propriedades: { estilos: {} }, componente_filhos: [] }}
        onAtualizar={onAtualizar}
        onRemover={vi.fn()}
        onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()}
      />,
    );
    fireEvent.click(screen.getByRole('button', { name: /adicionar item/i }));
    expect(onAtualizar).toHaveBeenCalledWith(
      'b1',
      expect.objectContaining({ itens: [{ label: 'Nova página', url: '#' }] }),
    );
  });

  it('edita o rótulo de um item existente', () => {
    const onAtualizar = vi.fn();
    render(
      <Inspector
        componente={{
          blind_index: 'b1',
          tipo: 'breadcrumb',
          propriedades: { estilos: {}, itens: [{ label: 'Início', url: '/' }] },
          componente_filhos: [],
        }}
        onAtualizar={onAtualizar}
        onRemover={vi.fn()}
        onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()}
      />,
    );
    fireEvent.change(screen.getByLabelText('Rótulo do item 1'), { target: { value: 'Home' } });
    expect(onAtualizar).toHaveBeenCalledWith('b1', expect.objectContaining({ itens: [{ label: 'Home', url: '/' }] }));
  });

  it('remove um item existente', () => {
    const onAtualizar = vi.fn();
    render(
      <Inspector
        componente={{
          blind_index: 'b1',
          tipo: 'breadcrumb',
          propriedades: { estilos: {}, itens: [{ label: 'A', url: '/a' }, { label: 'B', url: '/b' }] },
          componente_filhos: [],
        }}
        onAtualizar={onAtualizar}
        onRemover={vi.fn()}
        onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()}
      />,
    );
    fireEvent.click(screen.getByRole('button', { name: 'Remover item 1' }));
    expect(onAtualizar).toHaveBeenCalledWith('b1', expect.objectContaining({ itens: [{ label: 'B', url: '/b' }] }));
  });
});

describe('Inspector — componente Toggle (ativo)', () => {
  it('liga e desliga o toggle', () => {
    const onAtualizar = vi.fn();
    render(
      <Inspector
        componente={{ blind_index: 'b1', tipo: 'toggle', propriedades: { estilos: {} }, componente_filhos: [] }}
        onAtualizar={onAtualizar}
        onRemover={vi.fn()}
        onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()}
      />,
    );
    fireEvent.click(screen.getByRole('checkbox', { name: 'Ativado' }));
    expect(onAtualizar).toHaveBeenCalledWith('b1', expect.objectContaining({ ativo: true }));
  });
});

describe('Inspector — componente Alerta (variante)', () => {
  it('troca a variante', () => {
    const onAtualizar = vi.fn();
    render(
      <Inspector
        componente={{ blind_index: 'b1', tipo: 'alerta', propriedades: { estilos: {}, texto: 'Oi' }, componente_filhos: [] }}
        onAtualizar={onAtualizar}
        onRemover={vi.fn()}
        onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()}
      />,
    );
    fireEvent.change(screen.getByRole('combobox', { name: /variante/i }), { target: { value: 'erro' } });
    expect(onAtualizar).toHaveBeenCalledWith('b1', expect.objectContaining({ variante: 'erro' }));
  });
});

describe('Inspector — componente Avatar (imagem opcional)', () => {
  function avatarFake(): Componente {
    return { blind_index: 'b1', tipo: 'avatar', propriedades: { estilos: {}, texto: 'AB' }, componente_filhos: [] };
  }

  it('define a imagem por URL', () => {
    const onAtualizar = vi.fn();
    render(<Inspector componente={avatarFake()} onAtualizar={onAtualizar} onRemover={vi.fn()} onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()} />);
    const campo = screen.getByText('ou URL da imagem').closest('label')!;
    fireEvent.change(within(campo).getByRole('textbox'), { target: { value: 'https://exemplo.com/foto.png' } });
    expect(onAtualizar).toHaveBeenCalledWith('b1', expect.objectContaining({ src: 'https://exemplo.com/foto.png' }));
  });

  it('envia um arquivo pequeno e aplica como data URL', async () => {
    const onAtualizar = vi.fn();
    render(<Inspector componente={avatarFake()} onAtualizar={onAtualizar} onRemover={vi.fn()} onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()} />);
    const input = screen.getByLabelText('Enviar imagem') as HTMLInputElement;
    const arquivo = new File(['conteudo-fake'], 'foto.png', { type: 'image/png' });
    fireEvent.change(input, { target: { files: [arquivo] } });

    await waitFor(() =>
      expect(onAtualizar).toHaveBeenCalledWith('b1', expect.objectContaining({ src: expect.stringMatching(/^data:/) })),
    );
  });

  it('rejeita arquivo maior que 2MB com mensagem de erro, sem aplicar', () => {
    const onAtualizar = vi.fn();
    render(<Inspector componente={avatarFake()} onAtualizar={onAtualizar} onRemover={vi.fn()} onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()} />);
    const input = screen.getByLabelText('Enviar imagem') as HTMLInputElement;
    const grande = new File([new Uint8Array(2 * 1024 * 1024 + 1)], 'grande.png', { type: 'image/png' });
    fireEvent.change(input, { target: { files: [grande] } });

    expect(screen.getByText(/muito grande/i)).toBeInTheDocument();
    expect(onAtualizar).not.toHaveBeenCalled();
  });
});

describe('Inspector — posicionamento livre (X/Y)', () => {
  function componenteFake(estilos: Record<string, unknown> = {}): Componente {
    return { blind_index: 'b1', tipo: 'container', propriedades: { estilos }, componente_filhos: [] };
  }

  it('campos X/Y não aparecem quando a posição não é absolute', () => {
    render(
      <Inspector
        componente={componenteFake({ posicao: 'relative' })}
        onAtualizar={vi.fn()}
        onRemover={vi.fn()}
        onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()}
      />,
    );
    expect(screen.queryByText('X')).not.toBeInTheDocument();
    expect(screen.queryByText('Y')).not.toBeInTheDocument();
  });

  it('trocar Posição para absolute revela os campos X/Y', () => {
    render(<InspectorControlado inicial={componenteFake({ posicao: 'relative' })} />);
    const campoPosicao = screen.getByText('Posição').closest('label')!;
    fireEvent.change(within(campoPosicao).getByRole('combobox'), { target: { value: 'absolute' } });

    expect(screen.getByText('X')).toBeInTheDocument();
    expect(screen.getByText('Y')).toBeInTheDocument();
  });

  it('edita X e Y quando a posição já é absolute', () => {
    const onAtualizar = vi.fn();
    render(
      <Inspector
        componente={componenteFake({ posicao: 'absolute', x: 10, y: 20 })}
        onAtualizar={onAtualizar}
        onRemover={vi.fn()}
        onDuplicar={vi.fn()} onTrazerParaFrente={vi.fn()} onEnviarParaTras={vi.fn()}
      />,
    );
    const campoX = screen.getByText('X').closest('label')!;
    expect(within(campoX).getByRole('spinbutton')).toHaveValue(10);
    fireEvent.change(within(campoX).getByRole('spinbutton'), { target: { value: '50' } });
    expect(onAtualizar).toHaveBeenCalledWith(
      'b1',
      expect.objectContaining({ estilos: expect.objectContaining({ x: 50 }) }),
    );

    const campoY = screen.getByText('Y').closest('label')!;
    fireEvent.change(within(campoY).getByRole('spinbutton'), { target: { value: '99' } });
    expect(onAtualizar).toHaveBeenCalledWith(
      'b1',
      expect.objectContaining({ estilos: expect.objectContaining({ y: 99 }) }),
    );
  });
});

describe('Inspector — trazer para frente / enviar para trás', () => {
  function componenteFake(estilos: Record<string, unknown> = {}): Componente {
    return { blind_index: 'b1', tipo: 'container', propriedades: { estilos }, componente_filhos: [] };
  }

  it('botões só aparecem quando a posição é absolute', () => {
    render(
      <Inspector
        componente={componenteFake({ posicao: 'relative' })}
        onAtualizar={vi.fn()}
        onRemover={vi.fn()}
        onDuplicar={vi.fn()}
        onTrazerParaFrente={vi.fn()}
        onEnviarParaTras={vi.fn()}
      />,
    );
    expect(screen.queryByRole('button', { name: /trazer para frente/i })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /enviar para trás/i })).not.toBeInTheDocument();
  });

  it('clicar em "Trazer para frente" chama onTrazerParaFrente com o blind_index do componente', () => {
    const onTrazerParaFrente = vi.fn();
    render(
      <Inspector
        componente={componenteFake({ posicao: 'absolute', x: 0, y: 0 })}
        onAtualizar={vi.fn()}
        onRemover={vi.fn()}
        onDuplicar={vi.fn()}
        onTrazerParaFrente={onTrazerParaFrente}
        onEnviarParaTras={vi.fn()}
      />,
    );
    fireEvent.click(screen.getByRole('button', { name: /trazer para frente/i }));
    expect(onTrazerParaFrente).toHaveBeenCalledWith('b1');
  });

  it('clicar em "Enviar para trás" chama onEnviarParaTras com o blind_index do componente', () => {
    const onEnviarParaTras = vi.fn();
    render(
      <Inspector
        componente={componenteFake({ posicao: 'absolute', x: 0, y: 0 })}
        onAtualizar={vi.fn()}
        onRemover={vi.fn()}
        onDuplicar={vi.fn()}
        onTrazerParaFrente={vi.fn()}
        onEnviarParaTras={onEnviarParaTras}
      />,
    );
    fireEvent.click(screen.getByRole('button', { name: /enviar para trás/i }));
    expect(onEnviarParaTras).toHaveBeenCalledWith('b1');
  });
});
