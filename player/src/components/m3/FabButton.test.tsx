import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import { FabButton } from './FabButton';

describe('Atom: FabButton (M3)', () => {
  it('deve renderizar o botão com o texto correto', () => {
    render(<FabButton>Create</FabButton>);
    expect(screen.getByRole('button', { name: /create/i })).toBeInTheDocument();
  });

  it('deve renderizar o ícone quando fornecido', () => {
    render(<FabButton icon={<span data-testid="plus-icon">+</span>}>Create</FabButton>);
    expect(screen.getByTestId('plus-icon')).toBeInTheDocument();
  });

  it('deve repassar eventos de clique', () => {
    const handleClick = vi.fn();
    render(<FabButton onClick={handleClick}>Ação</FabButton>);
    
    fireEvent.click(screen.getByRole('button', { name: /ação/i }));
    expect(handleClick).toHaveBeenCalledTimes(1);
  });

  it('deve permitir classes CSS adicionais mantendo o formato de pílula (rounded-2xl)', () => {
    render(<FabButton className="bg-primary test-custom-class">Custom</FabButton>);
    const button = screen.getByRole('button', { name: /custom/i });
    
    expect(button).toHaveClass('test-custom-class');
    expect(button).toHaveClass('rounded-2xl');
    expect(button).toHaveClass('shadow-md');
  });
});
