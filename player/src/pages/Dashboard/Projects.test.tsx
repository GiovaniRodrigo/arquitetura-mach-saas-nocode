import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Projects } from './Projects';

describe('Page: Projects Dashboard', () => {
  it('deve renderizar o título e a descrição', () => {
    render(<Projects />);
    expect(screen.getByText('Projects')).toBeTruthy();
    expect(screen.getByText(/Gerencie seus projetos/i)).toBeTruthy();
  });

  it('exibe estado vazio, sem projetos mockados', () => {
    render(<Projects />);
    expect(screen.getByText(/Nenhum projeto ainda/i)).toBeTruthy();
  });
});
