import { describe, it, expect, vi } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { CardAcessosPorMes } from './CardAcessosPorMes';
import { renderDashboard, fakeClient } from '../test/renderDashboard';

describe('CardAcessosPorMes', () => {
  it('exibe o título e a variação percentual entre o primeiro e o último mês', async () => {
    const client = fakeClient({
      acessosPorMes: vi.fn().mockResolvedValue([
        { competencia: '2026-03', total: 20 },
        { competencia: '2026-08', total: 27 },
      ]),
    });
    renderDashboard(<CardAcessosPorMes />, { client });

    expect(await screen.findByText('Acessos por mês')).toBeInTheDocument();
    expect(await screen.findByText('+35%')).toBeInTheDocument();
  });

  it('exibe erro com retry quando a consulta falha', async () => {
    const client = fakeClient({ acessosPorMes: vi.fn().mockRejectedValue(new Error('boom')) });
    renderDashboard(<CardAcessosPorMes />, { client });
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });
});
