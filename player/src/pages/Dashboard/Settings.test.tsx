import { describe, it, expect, beforeEach } from 'vitest';
import { screen, fireEvent } from '@testing-library/react';
import { Settings } from './Settings';
import { renderDashboard } from '../../test/renderDashboard';
import { CHAVE_TEMA } from '../../theme/initTheme';

describe('Page: Settings Dashboard', () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.classList.remove('dark');
  });

  it('renderiza o título e a descrição', () => {
    renderDashboard(<Settings />);
    expect(screen.getByText('Configurações')).toBeTruthy();
    expect(screen.getByText(/Ajuste as preferências/i)).toBeTruthy();
  });

  it('alterna o tema e persiste (RF05)', () => {
    renderDashboard(<Settings />);
    const botao = screen.getByRole('button', { name: /Alternar Tema/i });
    // Default é escuro; alternar leva a claro.
    fireEvent.click(botao);
    expect(localStorage.getItem(CHAVE_TEMA)).toBe('claro');
    fireEvent.click(botao);
    expect(localStorage.getItem(CHAVE_TEMA)).toBe('escuro');
    expect(document.documentElement.classList.contains('dark')).toBe(true);
  });

  it('"Editar Perfil" é um botão acionável (RF04/C5)', () => {
    renderDashboard(<Settings />);
    expect(screen.getByRole('button', { name: /Editar Perfil/i })).toBeTruthy();
  });
});
