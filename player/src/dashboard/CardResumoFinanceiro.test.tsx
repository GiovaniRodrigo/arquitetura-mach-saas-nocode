import { describe, it, expect, vi } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { CardResumoFinanceiro } from './CardResumoFinanceiro';
import { renderDashboard, fakeClient } from '../test/renderDashboard';

describe('CardResumoFinanceiro (RF06, RN04)', () => {
  it('exibe a receita formatada na moeda correta', async () => {
    const client = fakeClient({
      resumoFinanceiro: vi.fn().mockResolvedValue({ receita_total_centavos: 150000, moeda: 'BRL', competencia: '2026-08' }),
    });
    renderDashboard(<CardResumoFinanceiro />, { client });

    expect(await screen.findByText(/1\.500,00/)).toBeInTheDocument();
  });

  it('exibe erro com retry quando a consulta falha', async () => {
    const client = fakeClient({ resumoFinanceiro: vi.fn().mockRejectedValue(new Error('boom')) });
    renderDashboard(<CardResumoFinanceiro />, { client });
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });
});
