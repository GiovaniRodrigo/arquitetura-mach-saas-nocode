import { describe, it, expect, vi } from 'vitest';
import { screen, waitFor, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import type { RecursosResponse } from '../../api/types';
import { ApiError } from '../../api/client';
import { Monitor } from './Monitor';
import { renderDashboard, fakeClient } from '../../test/renderDashboard';

const recursos: RecursosResponse = {
  servicos: [
    {
      nome: 'iam',
      status: 'servindo',
      cpu_millicores: 18,
      memoria_bytes: 15728640,
      requisicoes_por_segundo: 0.3,
      taxa_sucesso_percent: 100,
      latencia_p99_ms: 1.2,
    },
    {
      nome: 'logic',
      status: 'indisponivel',
      cpu_millicores: 0,
      memoria_bytes: 0,
      requisicoes_por_segundo: 0,
      taxa_sucesso_percent: 0,
      latencia_p99_ms: 0,
    },
  ],
  coletado_em_unix: 1755158400,
};

describe('Page: Monitor de Recursos (RF06-RF08)', () => {
  it('renderiza o título e um card por serviço', async () => {
    const client = fakeClient({ obterRecursos: vi.fn().mockResolvedValue(recursos) });
    renderDashboard(<Monitor />, { client });

    expect(screen.getByText('Monitor de Recursos')).toBeTruthy();
    expect(await screen.findByText('IAM')).toBeTruthy();
    expect(screen.getByText('Logic')).toBeTruthy();
  });

  it('botão "Atualizar" chama o endpoint novamente', async () => {
    const buscar = vi.fn().mockResolvedValue(recursos);
    const client = fakeClient({ obterRecursos: buscar });
    renderDashboard(<Monitor />, { client });

    await screen.findByText('IAM');
    expect(buscar).toHaveBeenCalledTimes(1);

    fireEvent.click(screen.getByRole('button', { name: /atualizar/i }));
    await waitFor(() => expect(buscar).toHaveBeenCalledTimes(2));
  });

  it('erro do endpoint inteiro mostra um estado único, sem renderizar cards (RNF02)', async () => {
    const client = fakeClient({
      obterRecursos: vi.fn().mockRejectedValue(new ApiError(502, 'RECURSOS_INDISPONIVEIS', 'boom')),
    });
    renderDashboard(<Monitor />, { client });

    expect(await screen.findByRole('alert')).toHaveTextContent(/não foi possível consultar/i);
    expect(screen.queryByTestId('indicador-status')).not.toBeInTheDocument();
  });
});
