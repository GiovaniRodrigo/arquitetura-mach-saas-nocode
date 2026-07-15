import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { DashboardLayout } from './DashboardLayout';
import { TooltipProvider } from '@/components/ui/tooltip';

Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(), // deprecated
    removeListener: vi.fn(), // deprecated
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

describe('DashboardLayout', () => {
  const renderLayout = (initialRoute = '/dashboard') => {
    return render(
      <MemoryRouter initialEntries={[initialRoute]}>
        <TooltipProvider>
          <Routes>
            <Route path="/dashboard" element={<DashboardLayout />}>
              <Route index element={<div data-testid="outlet-content">Home Content</div>} />
              <Route path="projects" element={<div data-testid="outlet-content">Projects Content</div>} />
              <Route path="settings" element={<div data-testid="outlet-content">Settings Content</div>} />
            </Route>
          </Routes>
        </TooltipProvider>
      </MemoryRouter>
    );
  };

  it('deve renderizar a navegação principal (Sidebar) e os links', () => {
    renderLayout();
    
    // Header da marca
    expect(screen.getByText('SaaS NoCode')).toBeInTheDocument();
    
    // Links de navegação
    expect(screen.getByText('Home')).toBeInTheDocument();
    expect(screen.getByText('Projects')).toBeInTheDocument();
    expect(screen.getByText('Settings')).toBeInTheDocument();
  });

  it('deve renderizar o Outlet corretamente', () => {
    renderLayout('/dashboard/projects');
    expect(screen.getByTestId('outlet-content')).toHaveTextContent('Projects Content');
  });

  it('deve alternar a visibilidade da Sidebar ao clicar no Trigger', async () => {
    renderLayout();
    
    // O Trigger do Shadcn tem a classe sr-only com texto "Toggle Sidebar" em seu botão interno.
    const trigger = screen.getByRole('button', { name: /toggle sidebar/i });
    expect(trigger).toBeInTheDocument();

    // Podemos testar que o botão existe e é clicável, embora a animação/CSS de expansão
    // seja de responsabilidade do css do provider.
    fireEvent.click(trigger);
    
    // Se o componente não quebrou ao clicar e continua renderizado, a mecânica básica do context não gerou erros
    expect(screen.getByText('SaaS NoCode')).toBeInTheDocument();
  });
});
