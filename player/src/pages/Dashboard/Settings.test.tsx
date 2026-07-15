import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Settings } from './Settings';

describe('Page: Settings Dashboard', () => {
  it('deve renderizar o título e a descrição', () => {
    render(<Settings />);
    expect(screen.getByText('Configurações')).toBeTruthy();
    expect(screen.getByText(/Ajuste as preferências/i)).toBeTruthy();
  });

  it('deve renderizar os painéis de configuração', () => {
    render(<Settings />);
    expect(screen.getByText('Perfil do Usuário')).toBeTruthy();
    expect(screen.getByText('Aparência')).toBeTruthy();
  });
});
