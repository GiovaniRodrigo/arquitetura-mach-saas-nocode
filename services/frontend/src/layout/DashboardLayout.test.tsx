import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { DashboardLayout } from './DashboardLayout';
import { AppProvider } from '../app/AppContext';
import { ThemeProvider } from '../theme/ThemeProvider';
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

describe('Template: DashboardLayout', () => {
  const renderLayout = (initialRoute = '/dashboard') =>
    render(
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
              <Route index element={<div data-testid="outlet-content">Dashboard Content</div>} />
              <Route path="clientes" element={<div data-testid="outlet-content">Clientes Content</div>} />
              <Route path="configuracao" element={<div data-testid="outlet-content">Configuração Content</div>} />
              <Route path="perfil" element={<div data-testid="outlet-content">Perfil Content</div>} />
              <Route path="ajuda" element={<div data-testid="outlet-content">Ajuda Content</div>} />
            </Route>
          </Routes>
        </ThemeProvider>
      </MemoryRouter>,
    );

  it('renderiza a navegação principal (Sidebar) e os links', () => {
    renderLayout();
    expect(screen.getByText('MAYS - Make Your SaaS')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Dashboard' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Clientes' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Configuração' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Ajuda' })).toBeInTheDocument();
  });

  it('exibe o atalho fixo de Perfil/Cadastro no cabeçalho e navega ao clicar (RF17-RF19)', () => {
    renderLayout();
    const botao = screen.getByRole('button', { name: 'Perfil/Cadastro' });
    expect(botao).toBeInTheDocument();
    fireEvent.click(botao);
    expect(screen.getByTestId('outlet-content')).toHaveTextContent('Perfil Content');
  });

  it('exibe a identidade real do usuário no cabeçalho (RF03)', () => {
    renderLayout();
    expect(screen.getByRole('button', { name: /menu do usuário/i })).toHaveTextContent('AS');
    fireEvent.click(screen.getByRole('button', { name: /menu do usuário/i }));
    expect(screen.getByText('Ana Silva')).toBeInTheDocument();
    expect(screen.getByText('ana@x.com')).toBeInTheDocument();
  });

  it('exibe o nome da página atual no cabeçalho', () => {
    renderLayout();
    expect(screen.getByRole('heading', { name: 'Dashboard', level: 1 })).toBeInTheDocument();

    renderLayout('/dashboard/clientes');
    expect(screen.getByRole('heading', { name: 'Clientes', level: 1 })).toBeInTheDocument();

    renderLayout('/dashboard/ajuda');
    expect(screen.getByRole('heading', { name: 'Ajuda', level: 1 })).toBeInTheDocument();
  });

  it('abre o menu do avatar com as ações (RF14/C7)', () => {
    renderLayout();
    fireEvent.click(screen.getByRole('button', { name: /menu do usuário/i }));
    expect(screen.getByRole('menu')).toBeInTheDocument();
    expect(screen.getByRole('menuitem', { name: /Perfil/i })).toBeInTheDocument();
    expect(screen.getByRole('menuitem', { name: /Configurações/i })).toBeInTheDocument();
    expect(screen.getByRole('menuitem', { name: /Sair/i })).toBeInTheDocument();
  });

  it('o item "Perfil" do menu do avatar navega para /dashboard/perfil (RF17-RF19)', () => {
    renderLayout();
    fireEvent.click(screen.getByRole('button', { name: /menu do usuário/i }));
    fireEvent.click(screen.getByRole('menuitem', { name: /Perfil/i }));
    expect(screen.getByTestId('outlet-content')).toHaveTextContent('Perfil Content');
  });

  it('o item "Configurações" do menu do avatar navega para /dashboard/configuracao (RF13-RF16)', () => {
    renderLayout();
    fireEvent.click(screen.getByRole('button', { name: /menu do usuário/i }));
    fireEvent.click(screen.getByRole('menuitem', { name: /Configurações/i }));
    expect(screen.getByTestId('outlet-content')).toHaveTextContent('Configuração Content');
  });

  it('renderiza o Outlet corretamente', () => {
    renderLayout('/dashboard/clientes');
    expect(screen.getByTestId('outlet-content')).toHaveTextContent('Clientes Content');
  });

  it('alterna a Sidebar ao clicar no Trigger', () => {
    renderLayout();
    const trigger = screen.getByRole('button', { name: /toggle sidebar/i });
    expect(trigger).toBeInTheDocument();
    fireEvent.click(trigger);
    expect(screen.getByText('MAYS - Make Your SaaS')).toBeInTheDocument();
  });
});
