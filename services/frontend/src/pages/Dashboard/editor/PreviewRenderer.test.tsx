import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { Componente } from '../../../api/types';
import { PreviewRenderer } from './PreviewRenderer';

describe('PreviewRenderer', () => {
  it('renderiza como flex um container/section/card sem `display` explícito nos estilos', () => {
    // Regressão: telas montadas fora do Canvas (ex.: build/seed-demo-site.sh)
    // não setam `display: 'flex'` explicitamente — só `direcao`/`justificar`
    // — porque o Canvas (editor/Canvas.tsx:270) já aplica esse fallback e a
    // tela parece correta lá. Sem o mesmo fallback aqui, o Preview renderiza
    // um <div> em display:block puro e ignora `flexDirection`/`gap`, então
    // os filhos empilham em vez de ficar lado a lado.
    const no: Componente = {
      blind_index: 'planos-linha',
      tipo: 'container',
      propriedades: { estilos: { direcao: 'row', espacamento: '24px' } },
      componente_filhos: [
        { blind_index: 'plano-1', tipo: 'card', propriedades: {} },
        { blind_index: 'plano-2', tipo: 'card', propriedades: {} },
      ],
    };
    const { container } = render(<PreviewRenderer no={no} />);
    const el = container.firstElementChild as HTMLElement;
    expect(el.style.display).toBe('flex');
    expect(el.style.flexDirection).toBe('row');
  });

  it('não força display em componentes que não aceitam filhos', () => {
    const no: Componente = {
      blind_index: 'titulo',
      tipo: 'heading',
      propriedades: { texto: 'Olá', estilos: {} },
    };
    const { container } = render(<PreviewRenderer no={no} />);
    const el = container.firstElementChild as HTMLElement;
    expect(el.style.display).toBe('');
  });

  it('preserva um `display` explícito em vez de sobrescrever com o fallback', () => {
    const no: Componente = {
      blind_index: 'grade',
      tipo: 'container',
      propriedades: { estilos: { display: 'inline-block' } },
    };
    const { container } = render(<PreviewRenderer no={no} />);
    const el = container.firstElementChild as HTMLElement;
    expect(el.style.display).toBe('inline-block');
  });
});
