import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { DashboardLayout } from './DashboardLayout';
import { AppProvider } from '../app/AppContext';
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
const usuario: UsuarioAutenticado = { nome: 'Ana Silva', email: 'ana@x.com', iniciais: 'AS' };

describe('Template: DashboardLayout', () => {
  const renderLayout = (initialRoute = '/dashboard') =>
    render(
      <MemoryRouter initialEntries={[initialRoute]}>
        <Routes>
          <Route
            path="/dashboard"
            element={
              <AppProvider client={client} usuario={usuario}>
                <DashboardLayout />
              </AppProvider>
            }
          >
            <Route index element={<div data-testid="outlet-content">Home Content</div>} />
            <Route path="projects" element={<div data-testid="outlet-content">Projects Content</div>} />
            <Route path="settings" element={<div data-testid="outlet-content">Settings Content</div>} />
          </Route>
        </Routes>
      </MemoryRouter>,
    );

  it('renderiza a navegação principal (Sidebar) e os links', () => {
    renderLayout();
    expect(screen.getByText('SaaS NoCode')).toBeInTheDocument();
    expect(screen.getByText('Home')).toBeInTheDocument();
    expect(screen.getByText('Projects')).toBeInTheDocument();
    expect(screen.getByText('Settings')).toBeInTheDocument();
  });

  it('exibe a identidade real do usuário no cabeçalho (RF03)', () => {
    renderLayout();
    expect(screen.getByText(/Bem-vindo, Ana Silva/)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /menu do usuário/i })).toHaveTextContent('AS');
  });

  it('abre o menu do avatar com as ações (RF14/C7)', () => {
    renderLayout();
    fireEvent.click(screen.getByRole('button', { name: /menu do usuário/i }));
    expect(screen.getByRole('menu')).toBeInTheDocument();
    expect(screen.getByRole('menuitem', { name: /Perfil/i })).toBeInTheDocument();
    expect(screen.getByRole('menuitem', { name: /Configurações/i })).toBeInTheDocument();
    expect(screen.getByRole('menuitem', { name: /Sair/i })).toBeInTheDocument();
  });

  it('renderiza o Outlet corretamente', () => {
    renderLayout('/dashboard/projects');
    expect(screen.getByTestId('outlet-content')).toHaveTextContent('Projects Content');
  });

  it('alterna a Sidebar ao clicar no Trigger', () => {
    renderLayout();
    const trigger = screen.getByRole('button', { name: /toggle sidebar/i });
    expect(trigger).toBeInTheDocument();
    fireEvent.click(trigger);
    expect(screen.getByText('SaaS NoCode')).toBeInTheDocument();
  });
});
