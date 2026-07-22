import { describe, expect, it } from 'vitest';
import { lerClaims, iniciaisDe, usuarioDe } from './jwt';

/** Gera um JWT falso (só payload importa) com o header/assinatura fixos. */
function fakeJwt(payload: Record<string, unknown>): string {
  const b64 = (o: unknown) =>
    btoa(JSON.stringify(o)).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
  return `${b64({ alg: 'none' })}.${b64(payload)}.sig`;
}

describe('jwt', () => {
  it('lerClaims extrai name/email do payload', () => {
    const t = fakeJwt({ name: 'Ana Silva', email: 'ana@x.com' });
    expect(lerClaims(t)).toMatchObject({ name: 'Ana Silva', email: 'ana@x.com' });
  });

  it('lerClaims retorna null para token malformado', () => {
    expect(lerClaims('abc')).toBeNull();
    expect(lerClaims('')).toBeNull();
    expect(lerClaims('a.@@@.c')).toBeNull();
  });

  it('iniciaisDe deriva de nome e de e-mail', () => {
    expect(iniciaisDe('Ana Silva')).toBe('AS');
    expect(iniciaisDe('Ana')).toBe('A');
    expect(iniciaisDe(undefined, 'bob@x.com')).toBe('B');
    expect(iniciaisDe()).toBe('?');
  });

  it('usuarioDe monta objeto de exibição', () => {
    const u = usuarioDe(fakeJwt({ name: 'Ana Silva', email: 'ana@x.com' }));
    expect(u).toEqual({ nome: 'Ana Silva', email: 'ana@x.com', iniciais: 'AS', podeCriarSistema: false });
  });

  it('usuarioDe degrada com token vazio', () => {
    expect(usuarioDe('')).toEqual({
      nome: undefined,
      email: undefined,
      iniciais: '?',
      podeCriarSistema: false,
    });
  });

  it.each([
    ['dono', true],
    ['parceiro', true],
    ['cliente', false],
  ])('usuarioDe deriva podeCriarSistema=%s a partir do claim tipo=%s', (tipo, esperado) => {
    const u = usuarioDe(fakeJwt({ name: 'Ana', tipo }));
    expect(u.podeCriarSistema).toBe(esperado);
  });
});
