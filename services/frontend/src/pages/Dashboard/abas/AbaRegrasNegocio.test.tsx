import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { AppProvider } from '../../../app/AppContext';
import { AbaRegrasNegocio } from './AbaRegrasNegocio';
import { fakeClient, usuarioFake } from '../../../test/renderDashboard';
import type { Design, DesignResumo } from '../../../api/types';

const TELA_CLIENTES: DesignResumo = { id: 'd1', nome: 'Clientes' };
const DESIGN_CLIENTES: Design = {
  id: 'd1',
  sistema_id: 's1',
  nome: 'Clientes',
  arvore: {
    blind_index: 'raiz',
    tipo: 'container',
    componente_filhos: [{ blind_index: 'bi-cpf', tipo: 'input', propriedades: { texto: 'CPF' } }],
  },
};

const TELA_PEDIDOS: DesignResumo = { id: 'd2', nome: 'Pedidos' };
const DESIGN_PEDIDOS: Design = {
  id: 'd2',
  sistema_id: 's1',
  nome: 'Pedidos',
  arvore: {
    blind_index: 'raiz2',
    tipo: 'container',
    componente_filhos: [{ blind_index: 'bi-total', tipo: 'input', propriedades: { texto: 'Total' } }],
  },
};

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

function clientePadrao(over: Parameters<typeof fakeClient>[0] = {}) {
  return fakeClient({
    listarTelas: vi.fn().mockResolvedValue([TELA_CLIENTES]),
    obterDesign: vi.fn().mockResolvedValue(DESIGN_CLIENTES),
    listarRegrasNegocio: vi.fn().mockResolvedValue([]),
    ...over,
  });
}

describe('AbaRegrasNegocio (RF10/RF11, RN06)', () => {
  it('lista os componentes de formulário encontrados nas telas do sistema', async () => {
    renderAba(clientePadrao());

    expect(await screen.findByText('bi-cpf')).toBeInTheDocument();
    expect(screen.getByText('CPF')).toBeInTheDocument();
  });

  it('ao selecionar um componente, mostra a prévia da tela dona dele', async () => {
    renderAba(clientePadrao());

    fireEvent.click(await screen.findByRole('button', { name: /cpf/i }));

    expect(await screen.findByText(/prévia — clientes/i)).toBeInTheDocument();
    expect(screen.getByText(/componente: cpf/i)).toBeInTheDocument();
  });

  it('cria uma regra de validação para o componente selecionado (RF10)', async () => {
    const criarRegraNegocio = vi.fn().mockResolvedValue({
      id: 'r2',
      blind_indexes: ['bi-cpf'],
      tipo: 'tamanho',
      parametros: { max: 11 },
    });
    renderAba(clientePadrao({ criarRegraNegocio }));

    fireEvent.click(await screen.findByRole('button', { name: /cpf/i }));
    fireEvent.change(await screen.findByLabelText(/^tipo$/i), { target: { value: 'tamanho' } });
    fireEvent.change(screen.getByLabelText(/tamanho m[aá]ximo/i), { target: { value: '11' } });
    fireEvent.click(screen.getByRole('button', { name: /salvar/i }));

    await waitFor(() =>
      expect(criarRegraNegocio).toHaveBeenCalledWith('s1', {
        blind_indexes: ['bi-cpf'],
        tipo: 'tamanho',
        parametros: { max: 11 },
      }),
    );
  });

  it('exibe a regra já cadastrada do componente selecionado', async () => {
    renderAba(
      clientePadrao({
        listarRegrasNegocio: vi.fn().mockResolvedValue([
          { id: 'r1', blind_indexes: ['bi-cpf'], tipo: 'tamanho', parametros: { max: 11 } },
        ]),
      }),
    );

    fireEvent.click(await screen.findByRole('button', { name: /cpf/i }));

    expect(await screen.findByText(/regra atual:/i)).toBeInTheDocument();
    expect(screen.getByText(/máx\. 11/i)).toBeInTheDocument();
  });

  it('exibe placeholder para regra multi-componente (RF11, fora do escopo funcional desta fase)', async () => {
    renderAba(clientePadrao());
    expect(await screen.findByText(/em breve/i)).toBeInTheDocument();
  });

  it('a busca de componentes é restrita à tela selecionada', async () => {
    renderAba(
      clientePadrao({
        listarTelas: vi.fn().mockResolvedValue([TELA_CLIENTES, TELA_PEDIDOS]),
        obterDesign: vi.fn().mockImplementation((id: string) =>
          Promise.resolve(id === TELA_CLIENTES.id ? DESIGN_CLIENTES : DESIGN_PEDIDOS),
        ),
      }),
    );

    // Tela "Clientes" (primeira) selecionada por padrão: só o campo dela aparece.
    expect(await screen.findByText('bi-cpf')).toBeInTheDocument();
    expect(screen.queryByText('bi-total')).not.toBeInTheDocument();

    fireEvent.change(screen.getByLabelText(/^tela$/i), { target: { value: TELA_PEDIDOS.id } });

    expect(await screen.findByText('bi-total')).toBeInTheDocument();
    expect(screen.queryByText('bi-cpf')).not.toBeInTheDocument();
  });
});
