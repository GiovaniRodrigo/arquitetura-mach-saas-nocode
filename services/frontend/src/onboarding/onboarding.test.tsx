import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { DashboardLayout } from '../layout/DashboardLayout';
import { AppProvider } from '../app/AppContext';
import { ThemeProvider } from '../theme/ThemeProvider';
import { telaOnboardingDe } from './conteudo';
import type { ApiClient } from '../api/client';
import type { UsuarioAutenticado } from '../auth/jwt';

Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation((query) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

const client = {} as ApiClient;
const usuario: UsuarioAutenticado = {
  nome: 'Ana Silva',
  email: 'ana@x.com',
  iniciais: 'AS',
  podeCriarSistema: true,
};

function renderLayout(initialRoute: string) {
  return render(
    <MemoryRouter initialEntries={[initialRoute]}>
      <ThemeProvider>
        <Routes>
          <Route
            path="/dashboard"
            element={
              <AppProvider client={client} usuario={usuario}>
                <DashboardLayout />
              </AppProvider>
            }
          >
            <Route index element={<div>Dashboard Content</div>} />
            <Route path="clientes" element={<div>Clientes Content</div>} />
            <Route path="ajuda" element={<div>Ajuda Content</div>} />
          </Route>
        </Routes>
      </ThemeProvider>
    </MemoryRouter>,
  );
}

describe('telaOnboardingDe', () => {
  it('resolve as telas do dashboard, clientes e sub-rotas do sistema', () => {
    expect(telaOnboardingDe('/dashboard')?.chave).toBe('dashboard');
    expect(telaOnboardingDe('/dashboard/clientes')?.chave).toBe('clientes');
    expect(telaOnboardingDe('/dashboard/clientes/t1')?.chave).toBe('cliente-sistemas');
    expect(telaOnboardingDe('/dashboard/clientes/t1/sistemas/s1/telas')?.chave).toBe('sistema-telas');
    expect(telaOnboardingDe('/dashboard/clientes/t1/sistemas/s1/regras')?.chave).toBe('sistema-regras');
    expect(telaOnboardingDe('/dashboard/clientes/t1/sistemas/s1/versao')?.chave).toBe('sistema-versao');
    expect(telaOnboardingDe('/dashboard/configuracao')?.chave).toBe('configuracao');
    expect(telaOnboardingDe('/dashboard/monitor')?.chave).toBe('monitor');
    expect(telaOnboardingDe('/dashboard/perfil')?.chave).toBe('perfil');
  });

  it('não tem tour para a própria tela de Ajuda', () => {
    expect(telaOnboardingDe('/dashboard/ajuda')).toBeNull();
  });
});

describe('Onboarding guiado no DashboardLayout', () => {
  beforeEach(() => {
    window.localStorage.clear();
  });

  it('mostra o onboarding automaticamente no primeiro acesso à tela', () => {
    renderLayout('/dashboard');
    expect(screen.getByRole('dialog')).toHaveTextContent('Bem-vindo ao Dashboard');
  });

  it('não mostra novamente numa tela já vista, mas o ícone de ajuda do cabeçalho reabre o tour', () => {
    window.localStorage.setItem('mach-onboarding-vistos', JSON.stringify(['dashboard']));
    renderLayout('/dashboard');
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: /ajuda guiada desta tela/i }));
    expect(screen.getByRole('dialog')).toHaveTextContent('Bem-vindo ao Dashboard');
  });

  it('avança os passos e fecha ao concluir o último', () => {
    renderLayout('/dashboard');
    expect(screen.getByText('Visão geral')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Próximo' }));
    expect(screen.getByText('Navegação')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Próximo' }));
    expect(screen.getByText('Busca rápida')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Concluir' }));
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('não exibe o ícone de ajuda guiada em telas sem onboarding (ex.: Ajuda)', () => {
    window.localStorage.setItem('mach-onboarding-vistos', JSON.stringify(['dashboard']));
    renderLayout('/dashboard/ajuda');
    expect(screen.queryByRole('button', { name: /ajuda guiada desta tela/i })).not.toBeInTheDocument();
  });
});
