import { describe, it, expect, vi } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { CardReceitaPorMes } from './CardReceitaPorMes';
import { renderDashboard, fakeClient } from '../test/renderDashboard';

describe('CardReceitaPorMes', () => {
  it('exibe o título e a variação percentual entre o primeiro e o último mês', async () => {
    const client = fakeClient({
      receitaPorMes: vi.fn().mockResolvedValue({
        pontos: [
          { competencia: '2026-03', valor_centavos: 100000 },
          { competencia: '2026-08', valor_centavos: 134000 },
        ],
        moeda: 'BRL',
      }),
    });
    renderDashboard(<CardReceitaPorMes />, { client });

    expect(await screen.findByText('Receita de assinatura')).toBeInTheDocument();
    expect(await screen.findByText('+34%')).toBeInTheDocument();
  });

  it('exibe erro com retry quando a consulta falha', async () => {
    const client = fakeClient({ receitaPorMes: vi.fn().mockRejectedValue(new Error('boom')) });
    renderDashboard(<CardReceitaPorMes />, { client });
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });
});
