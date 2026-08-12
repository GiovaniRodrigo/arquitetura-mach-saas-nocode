import { describe, expect, it } from 'vitest';
import { estilosParaCss } from './estilosCss';

describe('estilosParaCss', () => {
  it('devolve objeto vazio sem estilos', () => {
    expect(estilosParaCss(undefined)).toEqual({});
  });

  it('mapeia layout/flex para as propriedades CSS correspondentes', () => {
    const css = estilosParaCss({
      display: 'flex',
      direcao: 'column',
      justificar: 'center',
      alinhar: 'stretch',
      espacamento: '16px',
      largura: '100%',
      padding: '24px',
    });
    expect(css).toMatchObject({
      display: 'flex',
      flexDirection: 'column',
      justifyContent: 'center',
      alignItems: 'stretch',
      gap: '16px',
      width: '100%',
      padding: '24px',
    });
  });

  it('mapeia peso de fonte nomeado para número', () => {
    expect(estilosParaCss({ fontePeso: 'bold' }).fontWeight).toBe(700);
    expect(estilosParaCss({ fontePeso: 'medium' }).fontWeight).toBe(500);
    expect(estilosParaCss({ fontePeso: 'normal' }).fontWeight).toBe(400);
  });

  it('posição absoluta usa x/y como left/top', () => {
    const css = estilosParaCss({ posicao: 'absolute', x: 10, y: 20 });
    expect(css).toMatchObject({ position: 'absolute', left: 10, top: 20 });
  });

  it('posição relativa não define left/top', () => {
    const css = estilosParaCss({ posicao: 'relative', x: 10, y: 20 });
    expect(css.left).toBeUndefined();
    expect(css.top).toBeUndefined();
  });
});
