import { describe, expect, it } from 'vitest';
import { REGISTRO_COMPONENTES, CATEGORIAS, registroDoTipo, aceitaFilhos } from './componentRegistry';

describe('componentRegistry', () => {
  it('todo componente pertence a uma categoria conhecida', () => {
    const ids = CATEGORIAS.map((c) => c.id);
    for (const r of REGISTRO_COMPONENTES) {
      expect(ids).toContain(r.categoria);
    }
  });

  it('todo tipo é único no catálogo', () => {
    const tipos = REGISTRO_COMPONENTES.map((r) => r.tipo);
    expect(new Set(tipos).size).toBe(tipos.length);
  });

  it('propriedadesPadrao devolve estilos sempre presentes', () => {
    for (const r of REGISTRO_COMPONENTES) {
      const props = r.propriedadesPadrao();
      expect(props.estilos).toBeDefined();
      if (r.temTexto) {
        expect(typeof props.texto).toBe('string');
      }
    }
  });

  it('registroDoTipo encontra pelo tipo e devolve undefined se não existir', () => {
    expect(registroDoTipo('botao')?.rotulo).toBe('Botão');
    expect(registroDoTipo('inexistente')).toBeUndefined();
  });

  it('aceitaFilhos reflete containers (layout) vs. componentes folha', () => {
    expect(aceitaFilhos('container')).toBe(true);
    expect(aceitaFilhos('section')).toBe(true);
    expect(aceitaFilhos('botao')).toBe(false);
    expect(aceitaFilhos('inexistente')).toBe(false);
  });

  it('header/footer/sidebar/main/rightbar são containers de layout', () => {
    for (const tipo of ['header', 'footer', 'sidebar', 'main', 'rightbar']) {
      expect(registroDoTipo(tipo)?.categoria).toBe('layout');
      expect(aceitaFilhos(tipo)).toBe(true);
    }
  });

  it('carrossel é um componente de mídia sem filhos', () => {
    expect(registroDoTipo('carrossel')?.categoria).toBe('midia');
    expect(aceitaFilhos('carrossel')).toBe(false);
  });

  it('menu e accordion são componentes de layout autocontidos (sem filhos na árvore)', () => {
    for (const tipo of ['menu', 'accordion']) {
      expect(registroDoTipo(tipo)?.categoria).toBe('layout');
      expect(aceitaFilhos(tipo)).toBe(false);
    }
  });

  it('card aceita filhos, divisor/tabs/progresso são componentes de layout autocontidos', () => {
    expect(aceitaFilhos('card')).toBe(true);
    for (const tipo of ['divisor', 'tabs', 'progresso']) {
      expect(registroDoTipo(tipo)?.categoria).toBe('layout');
      expect(aceitaFilhos(tipo)).toBe(false);
    }
  });

  it('badge é um componente de texto com conteúdo padrão', () => {
    expect(registroDoTipo('badge')?.categoria).toBe('texto');
    expect(registroDoTipo('badge')?.temTexto).toBe(true);
    expect(registroDoTipo('badge')?.propriedadesPadrao().texto).toBe('Novo');
  });

  it('avaliacao é um componente de formulário sem filhos', () => {
    expect(registroDoTipo('avaliacao')?.categoria).toBe('formulario');
    expect(aceitaFilhos('avaliacao')).toBe(false);
  });

  it('video e icone são componentes de mídia sem filhos', () => {
    for (const tipo of ['video', 'icone']) {
      expect(registroDoTipo(tipo)?.categoria).toBe('midia');
      expect(aceitaFilhos(tipo)).toBe(false);
    }
  });

  it('avatar é um componente de mídia com conteúdo padrão (iniciais)', () => {
    expect(registroDoTipo('avatar')?.categoria).toBe('midia');
    expect(aceitaFilhos('avatar')).toBe(false);
    expect(registroDoTipo('avatar')?.temTexto).toBe(true);
    expect(registroDoTipo('avatar')?.propriedadesPadrao().texto).toBe('AB');
  });

  it('radio é um componente de formulário com conteúdo padrão', () => {
    expect(registroDoTipo('radio')?.categoria).toBe('formulario');
    expect(registroDoTipo('radio')?.temTexto).toBe(true);
  });

  it('textarea e toggle são componentes de formulário sem filhos', () => {
    for (const tipo of ['textarea', 'toggle']) {
      expect(registroDoTipo(tipo)?.categoria).toBe('formulario');
      expect(aceitaFilhos(tipo)).toBe(false);
    }
  });

  it('breadcrumb/spinner/skeleton são componentes de layout autocontidos', () => {
    for (const tipo of ['breadcrumb', 'spinner', 'skeleton']) {
      expect(registroDoTipo(tipo)?.categoria).toBe('layout');
      expect(aceitaFilhos(tipo)).toBe(false);
    }
  });

  it('alerta é um componente de texto com conteúdo padrão', () => {
    expect(registroDoTipo('alerta')?.categoria).toBe('texto');
    expect(registroDoTipo('alerta')?.temTexto).toBe(true);
  });
});
