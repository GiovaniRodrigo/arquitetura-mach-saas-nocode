import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import type { ServicoStatus } from '../api/types';
import { CardServicoStatus } from './CardServicoStatus';

const servico: ServicoStatus = {
  nome: 'iam',
  status: 'servindo',
  cpu_millicores: 18,
  memoria_bytes: 15728640,
  requisicoes_por_segundo: 0.3,
  taxa_sucesso_percent: 100,
  latencia_p99_ms: 1.2,
};

describe('CardServicoStatus (RF06)', () => {
  it('renderiza nome, CPU e memória formatados quando servindo', () => {
    render(<CardServicoStatus servico={servico} />);

    expect(screen.getByText('IAM')).toBeInTheDocument();
    expect(screen.getByText('18m')).toBeInTheDocument();
    expect(screen.getByText(/15(\.0)? ?MB/)).toBeInTheDocument();
    expect(screen.getByText('100.0%')).toBeInTheDocument();
  });

  it('mostra indicador visual verde quando o status é "servindo"', () => {
    render(<CardServicoStatus servico={servico} />);
    expect(screen.getByTestId('indicador-status')).toHaveAttribute('data-status', 'servindo');
  });

  it('mostra indicador vermelho e oculta as métricas quando indisponível', () => {
    const indisponivel: ServicoStatus = {
      nome: 'logic',
      status: 'indisponivel',
      cpu_millicores: 0,
      memoria_bytes: 0,
      requisicoes_por_segundo: 0,
      taxa_sucesso_percent: 0,
      latencia_p99_ms: 0,
    };
    render(<CardServicoStatus servico={indisponivel} />);

    expect(screen.getByTestId('indicador-status')).toHaveAttribute('data-status', 'indisponivel');
    expect(screen.getByText(/indispon[ií]vel/i)).toBeInTheDocument();
    expect(screen.queryByText(/MB/)).not.toBeInTheDocument();
  });
});
