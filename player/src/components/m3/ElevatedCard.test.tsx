import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import { ElevatedCard } from './ElevatedCard';

describe('Atom: ElevatedCard (M3)', () => {
  it('deve renderizar os children corretamente', () => {
    render(
      <ElevatedCard>
        <h1 data-testid="card-title">Métrica</h1>
      </ElevatedCard>
    );
    expect(screen.getByTestId('card-title')).toHaveTextContent('Métrica');
  });

  it('deve incluir as classes de elevação do Material Design 3', () => {
    render(<ElevatedCard data-testid="card">Conteúdo</ElevatedCard>);
    const card = screen.getByTestId('card');
    
    expect(card).toHaveClass('shadow-sm');
    expect(card).toHaveClass('hover:shadow-md');
    expect(card).toHaveClass('rounded-3xl'); // M3 radius
  });

  it('deve permitir a injeção de className adicional', () => {
    render(<ElevatedCard data-testid="card" className="bg-card text-primary">Estilizado</ElevatedCard>);
    const card = screen.getByTestId('card');
    
    expect(card).toHaveClass('bg-card');
    expect(card).toHaveClass('text-primary');
  });
});
