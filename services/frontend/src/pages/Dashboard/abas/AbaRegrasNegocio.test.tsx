import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { AppProvider } from '../../../app/AppContext';
import { AbaRegrasNegocio } from './AbaRegrasNegocio';
import { fakeClient, usuarioFake } from '../../../test/renderDashboard';

function renderAba(client: ReturnType<typeof fakeClient>) {
  return render(
    <MemoryRouter initialEntries={['/dashboard/clientes/t1/sistemas/s1/regras']}>
      <AppProvider client={client} usuario={usuarioFake}>
        <Routes>
          <Route path="/dashboard/clientes/:tenantId/sistemas/:sistemaId/regras" element={<AbaRegrasNegocio />} />
        </Routes>
      </AppProvider>
    </MemoryRouter>,
  );
}

describe('AbaRegrasNegocio (RF10/RF11, RN06)', () => {
  it('lista as regras existentes do sistema', async () => {
    const client = fakeClient({
      listarRegrasNegocio: vi.fn().mockResolvedValue([
        { id: 'r1', blind_indexes: ['bi-cpf'], tipo: 'tamanho', parametros: { max: 11 } },
      ]),
    });
    renderAba(client);

    expect(await screen.findByText(/bi-cpf/)).toBeInTheDocument();
  });

  it('cria uma regra de validação de componente único (RF10)', async () => {
    const criarRegraNegocio = vi.fn().mockResolvedValue({
      id: 'r2',
      blind_indexes: ['bi-cpf'],
      tipo: 'tamanho',
      parametros: { max: 11 },
    });
    const client = fakeClient({
      listarRegrasNegocio: vi.fn().mockResolvedValue([]),
      criarRegraNegocio,
    });
    renderAba(client);

    await screen.findByText(/nenhuma regra/i);
    fireEvent.change(screen.getByLabelText(/componente \(blind_index\)/i), { target: { value: 'bi-cpf' } });
    fireEvent.change(screen.getByLabelText(/tipo de valida[cç][aã]o/i), { target: { value: 'tamanho' } });
    fireEvent.change(screen.getByLabelText(/tamanho m[aá]ximo/i), { target: { value: '11' } });
    fireEvent.click(screen.getByRole('button', { name: /adicionar regra/i }));

    await waitFor(() =>
      expect(criarRegraNegocio).toHaveBeenCalledWith('s1', {
        blind_indexes: ['bi-cpf'],
        tipo: 'tamanho',
        parametros: { max: 11 },
      }),
    );
  });

  it('exibe placeholder para regra multi-componente (RF11, fora do escopo funcional desta fase)', async () => {
    const client = fakeClient({ listarRegrasNegocio: vi.fn().mockResolvedValue([]) });
    renderAba(client);
    expect(await screen.findByText(/em breve/i)).toBeInTheDocument();
  });
});
