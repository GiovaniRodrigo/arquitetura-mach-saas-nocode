import { describe, expect, it, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ThemeProvider, useTheme } from './ThemeProvider';
import { CHAVE_TEMA } from './initTheme';

function Sonda() {
  const { tema, alternarTema } = useTheme();
  return (
    <button onClick={alternarTema} data-testid="btn">
      {tema}
    </button>
  );
}

describe('ThemeProvider', () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.classList.remove('dark');
  });

  it('default é escuro e aplica a classe dark ao alternar de volta', () => {
    render(
      <ThemeProvider>
        <Sonda />
      </ThemeProvider>,
    );
    expect(screen.getByTestId('btn').textContent).toBe('escuro');

    fireEvent.click(screen.getByTestId('btn'));
    expect(screen.getByTestId('btn').textContent).toBe('claro');
    expect(document.documentElement.classList.contains('dark')).toBe(false);
    expect(localStorage.getItem(CHAVE_TEMA)).toBe('claro');

    fireEvent.click(screen.getByTestId('btn'));
    expect(screen.getByTestId('btn').textContent).toBe('escuro');
    expect(document.documentElement.classList.contains('dark')).toBe(true);
    expect(localStorage.getItem(CHAVE_TEMA)).toBe('escuro');
  });

  it('relê a preferência salva', () => {
    localStorage.setItem(CHAVE_TEMA, 'claro');
    render(
      <ThemeProvider>
        <Sonda />
      </ThemeProvider>,
    );
    expect(screen.getByTestId('btn').textContent).toBe('claro');
  });
});
