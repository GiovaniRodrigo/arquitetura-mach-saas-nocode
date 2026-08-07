import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { SistemaAbas } from './SistemaAbas';

function renderComRota(rota: string) {
  return render(
    <MemoryRouter initialEntries={[rota]}>
      <Routes>
        <Route path="/dashboard/clientes/:tenantId/sistemas/:sistemaId" element={<SistemaAbas />}>
          <Route path="telas" element={<div data-testid="conteudo-aba">Telas</div>} />
          <Route path="regras" element={<div data-testid="conteudo-aba">Regras</div>} />
          <Route path="versao" element={<div data-testid="conteudo-aba">Versão</div>} />
        </Route>
      </Routes>
    </MemoryRouter>,
  );
}

describe('SistemaAbas (RF09-RF12)', () => {
  it('renderiza as 3 abas de navegação', () => {
    renderComRota('/dashboard/clientes/t1/sistemas/s1/telas');
    expect(screen.getByRole('link', { name: /telas/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /regras de negócio/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /versão/i })).toBeInTheDocument();
  });

  it('renderiza o conteúdo da aba ativa via Outlet', () => {
    renderComRota('/dashboard/clientes/t1/sistemas/s1/regras');
    expect(screen.getByTestId('conteudo-aba')).toHaveTextContent('Regras');
  });

  it('destaca a aba ativa conforme a rota', () => {
    renderComRota('/dashboard/clientes/t1/sistemas/s1/versao');
    expect(screen.getByRole('link', { name: /versão/i })).toHaveAttribute('aria-current', 'page');
  });
});
