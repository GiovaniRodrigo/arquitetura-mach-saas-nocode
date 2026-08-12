import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useSessionStorageState } from './useSessionStorageState';

beforeEach(() => sessionStorage.clear());
afterEach(() => sessionStorage.clear());

describe('useSessionStorageState', () => {
  it('começa com o padrão quando não há valor salvo', () => {
    const { result } = renderHook(() => useSessionStorageState('chave-1', 'x'));
    expect(result.current[0]).toBe('x');
  });

  it('persiste no sessionStorage e uma nova instância com a mesma chave lê o valor salvo', () => {
    const { result, unmount } = renderHook(() => useSessionStorageState<string | null>('tela-sel', null));
    act(() => result.current[1]('d1'));
    expect(sessionStorage.getItem('tela-sel')).toBe('"d1"');

    unmount();
    const { result: result2 } = renderHook(() => useSessionStorageState<string | null>('tela-sel', null));
    expect(result2.current[0]).toBe('d1');
  });

  it('chaves diferentes não interferem entre si', () => {
    const { result: a } = renderHook(() => useSessionStorageState('a', 'padrao-a'));
    const { result: b } = renderHook(() => useSessionStorageState('b', 'padrao-b'));
    act(() => a.current[1]('mudou-a'));
    expect(b.current[0]).toBe('padrao-b');
  });
});
