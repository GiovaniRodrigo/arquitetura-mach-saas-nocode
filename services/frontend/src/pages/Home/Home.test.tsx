import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import { MemoryRouter } from 'react-router-dom';
import { Home } from './Home';

describe('Page: Home (RF01/RF02)', () => {
  it('renderiza a apresentação pública do produto sem exigir sessão', () => {
    render(
      <MemoryRouter>
        <Home />
      </MemoryRouter>,
    );
    expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument();
  });

  it('oferece o CTA "Entrar" apontando para o login', () => {
    render(
      <MemoryRouter>
        <Home />
      </MemoryRouter>,
    );
    const entrar = screen.getAllByRole('link', { name: /entrar/i });
    expect(entrar.length).toBeGreaterThan(0);
    for (const link of entrar) expect(link).toHaveAttribute('href', '/login');
  });

  it('oferece o CTA de cadastro/trial apontando para o cadastro (spec 006, RF07)', () => {
    render(
      <MemoryRouter>
        <Home />
      </MemoryRouter>,
    );
    const cadastrar = screen.getByRole('link', { name: /testar grátis|cadastr/i });
    expect(cadastrar).toHaveAttribute('href', '/register');
  });
});
