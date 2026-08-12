import { describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import { PainelComponentes } from './PainelComponentes';

describe('PainelComponentes — tooltip no hover', () => {
  it('cada item da paleta tem o rótulo completo como tooltip nativo (title)', () => {
    render(<PainelComponentes onAdicionar={vi.fn()} />);
    expect(screen.getByRole('button', { name: 'Container' })).toHaveAttribute('title', 'Container');
    expect(screen.getByRole('button', { name: 'Rightbar' })).toHaveAttribute('title', 'Rightbar');
    expect(screen.getByRole('button', { name: 'Avaliação' })).toHaveAttribute('title', 'Avaliação');
  });
});
