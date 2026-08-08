import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import { AbaTelas } from './AbaTelas';

describe('AbaTelas (RF09 — casca de navegação)', () => {
  it('renderiza o layout de 3 colunas (sidebar de telas, canvas, propriedades)', () => {
    render(<AbaTelas />);
    expect(screen.getByRole('complementary', { name: /telas/i })).toBeInTheDocument();
    expect(screen.getByRole('region', { name: /propriedades/i })).toBeInTheDocument();
  });

  it('exibe estado vazio "nenhuma tela criada ainda" — sem editor funcional nesta fase', () => {
    render(<AbaTelas />);
    expect(screen.getByText(/nenhuma tela criada ainda/i)).toBeInTheDocument();
  });
});
