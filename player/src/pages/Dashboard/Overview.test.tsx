import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Overview } from './Overview';

describe('Page: Overview Dashboard', () => {
  it('renderiza o cabeçalho de início, sem dados mockados', () => {
    render(<Overview />);
    expect(screen.getByText('Início')).toBeTruthy();
    expect(screen.getByText(/Bem-vindo à Plataforma MACH/i)).toBeTruthy();
  });
});
