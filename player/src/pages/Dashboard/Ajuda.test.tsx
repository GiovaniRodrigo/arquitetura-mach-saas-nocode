import { describe, it, expect } from 'vitest';
import { screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import { Ajuda } from './Ajuda';
import { renderDashboard } from '../../test/renderDashboard';

describe('Page: Ajuda (RF20/RF21)', () => {
  it('lista os artigos de documentação organizados por categoria', () => {
    renderDashboard(<Ajuda />);
    expect(screen.getByRole('heading', { name: /ajuda/i })).toBeInTheDocument();
    expect(screen.getAllByRole('heading', { level: 3 }).length).toBeGreaterThan(0);
  });

  it('filtra os artigos por termo de busca (RF21)', () => {
    renderDashboard(<Ajuda />);
    const busca = screen.getByRole('searchbox', { name: /buscar/i });
    fireEvent.change(busca, { target: { value: 'senha' } });
    const resultados = screen.getAllByRole('article');
    expect(resultados.length).toBeGreaterThan(0);
    for (const artigo of resultados) {
      expect(artigo.textContent?.toLowerCase()).toMatch(/senha/);
    }
  });

  it('exibe estado vazio quando a busca não encontra nada', () => {
    renderDashboard(<Ajuda />);
    const busca = screen.getByRole('searchbox', { name: /buscar/i });
    fireEvent.change(busca, { target: { value: 'termo-inexistente-xyz' } });
    expect(screen.getByText(/nenhum artigo encontrado/i)).toBeInTheDocument();
  });
});
