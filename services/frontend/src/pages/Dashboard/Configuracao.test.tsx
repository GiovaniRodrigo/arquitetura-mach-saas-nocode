import { describe, it, expect, beforeEach } from 'vitest';
import { screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import { Configuracao } from './Configuracao';
import { renderDashboard } from '../../test/renderDashboard';
import { CHAVE_TEMA } from '../../theme/initTheme';

describe('Page: Configuracao Dashboard', () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.classList.remove('dark');
  });

  it('renderiza o título e a descrição', () => {
    renderDashboard(<Configuracao />);
    expect(screen.getByText('Configurações')).toBeTruthy();
    expect(screen.getByText(/Ajuste as preferências/i)).toBeTruthy();
  });

  it('alterna o tema e persiste (RF05)', () => {
    renderDashboard(<Configuracao />);
    const botao = screen.getByRole('button', { name: /Alternar Tema/i });
    // Default é escuro; alternar leva a claro.
    fireEvent.click(botao);
    expect(localStorage.getItem(CHAVE_TEMA)).toBe('claro');
    fireEvent.click(botao);
    expect(localStorage.getItem(CHAVE_TEMA)).toBe('escuro');
    expect(document.documentElement.classList.contains('dark')).toBe(true);
  });

  it('não exibe mais o card "Perfil do Usuário" (movido para Cadastro/Perfil de topo)', () => {
    renderDashboard(<Configuracao />);
    expect(screen.queryByText('Perfil do Usuário')).toBeNull();
    expect(screen.queryByRole('button', { name: /Editar Perfil/i })).toBeNull();
  });

  it('inclui as seções White Label e Segurança, com âncora #seguranca (RF13-RF16)', () => {
    renderDashboard(<Configuracao />);
    expect(screen.getByRole('heading', { name: /white label/i })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: /^segurança$/i })).toBeInTheDocument();
    expect(document.getElementById('seguranca')).toBeInTheDocument();
  });
});
