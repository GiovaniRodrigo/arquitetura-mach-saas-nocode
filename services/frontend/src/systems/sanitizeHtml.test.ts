import { describe, expect, it } from 'vitest';
import { sanitizarHtml, htmlParaTextoPlano } from './sanitizeHtml';

describe('sanitizarHtml', () => {
  it('mantém tags permitidas (negrito/itálico/sublinhado)', () => {
    expect(sanitizarHtml('<b>forte</b> e <i>itálico</i> e <u>sublinhado</u>')).toBe(
      '<b>forte</b> e <i>itálico</i> e <u>sublinhado</u>',
    );
  });

  it('mantém texto puro sem tags', () => {
    expect(sanitizarHtml('Texto simples')).toBe('Texto simples');
  });

  it('remove tags fora do allowlist mas preserva o texto interno', () => {
    expect(sanitizarHtml('<script>alert(1)</script>Olá')).toBe('Olá');
    expect(sanitizarHtml('<div onclick="alert(1)">Oi</div>')).toBe('Oi');
    expect(sanitizarHtml('<iframe src="evil"></iframe>Texto')).toBe('Texto');
  });

  it('remove atributos não permitidos de tags permitidas', () => {
    expect(sanitizarHtml('<b onclick="alert(1)" class="x">forte</b>')).toBe('<b>forte</b>');
  });

  it('em span, mantém só propriedades de style do allowlist', () => {
    const saida = sanitizarHtml('<span style="font-weight: bold; color: red">x</span>');
    expect(saida).toContain('font-weight: bold');
    expect(saida).not.toContain('color');
  });

  it('remove span sem estilo permitido, preservando o texto', () => {
    expect(sanitizarHtml('<span style="color: red">x</span>')).toBe('<span>x</span>');
  });

  it('aceita tags aninhadas permitidas', () => {
    expect(sanitizarHtml('<b><i>forte e itálico</i></b>')).toBe('<b><i>forte e itálico</i></b>');
  });

  it('string vazia devolve string vazia', () => {
    expect(sanitizarHtml('')).toBe('');
  });
});

describe('htmlParaTextoPlano', () => {
  it('extrai só o texto, sem tags', () => {
    expect(htmlParaTextoPlano('<b>forte</b> e <i>itálico</i>')).toBe('forte e itálico');
  });

  it('texto simples permanece igual', () => {
    expect(htmlParaTextoPlano('Texto simples')).toBe('Texto simples');
  });

  it('remove conteúdo de script também (não deve executar)', () => {
    expect(htmlParaTextoPlano('<script>alert(1)</script>Olá')).toContain('Olá');
  });
});
