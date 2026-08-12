import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { SistemaAbas } from './SistemaAbas';
import { AppProvider } from '../../app/AppContext';
import type { ApiClient } from '../../api/client';
import type { UsuarioAutenticado } from '../../auth/jwt';

const usuarioFake: UsuarioAutenticado = { iniciais: 'A', podeCriarSistema: true, tenantId: 't1' };

function fakeClient(over: Partial<ApiClient>): ApiClient {
  return over as unknown as ApiClient;
}

function renderComRota(
  rota: string,
  client: ApiClient = fakeClient({ listarSistemas: vi.fn().mockResolvedValue([]) }),
) {
  return render(
    <MemoryRouter initialEntries={[rota]}>
      <AppProvider client={client} usuario={usuarioFake}>
        <Routes>
          <Route path="/dashboard/clientes/:tenantId/sistemas/:sistemaId" element={<SistemaAbas />}>
            <Route path="telas" element={<div data-testid="conteudo-aba">Telas</div>} />
            <Route path="regras" element={<div data-testid="conteudo-aba">Regras</div>} />
            <Route path="versao" element={<div data-testid="conteudo-aba">Versão</div>} />
          </Route>
        </Routes>
      </AppProvider>
    </MemoryRouter>,
  );
}

describe('SistemaAbas (RF09-RF12)', () => {
  it('renderiza as 3 abas de navegação', () => {
    renderComRota('/dashboard/clientes/t1/sistemas/s1/telas');
    expect(screen.getByRole('link', { name: /telas/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /regras de negócio/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /versão/i })).toBeInTheDocument();
  });

  it('renderiza o conteúdo da aba ativa via Outlet', () => {
    renderComRota('/dashboard/clientes/t1/sistemas/s1/regras');
    expect(screen.getByTestId('conteudo-aba')).toHaveTextContent('Regras');
  });

  it('destaca a aba ativa conforme a rota', () => {
    renderComRota('/dashboard/clientes/t1/sistemas/s1/versao');
    expect(screen.getByRole('link', { name: /versão/i })).toHaveAttribute('aria-current', 'page');
  });

  it('as abas Telas e Regras de Negócio usam a largura total (editor e prévia precisam do espaço)', () => {
    const telas = renderComRota('/dashboard/clientes/t1/sistemas/s1/telas');
    expect(telas.getByTestId('conteudo-aba').parentElement).not.toHaveClass('max-w-5xl');
    telas.unmount();

    const regras = renderComRota('/dashboard/clientes/t1/sistemas/s1/regras');
    expect(regras.getByTestId('conteudo-aba').parentElement).not.toHaveClass('max-w-5xl');
  });

  it('a aba Versão continua com largura limitada para leitura', () => {
    const versao = renderComRota('/dashboard/clientes/t1/sistemas/s1/versao');
    expect(versao.getByTestId('conteudo-aba').parentElement).toHaveClass('max-w-5xl');
  });

  it('exibe o nome do sistema no cabeçalho', async () => {
    const client = fakeClient({
      listarSistemas: vi.fn().mockResolvedValue([
        { id: 's1', nome: 'CRM' },
        { id: 's2', nome: 'ERP' },
      ]),
    });
    renderComRota('/dashboard/clientes/t1/sistemas/s1/telas', client);

    expect(await screen.findByRole('heading', { name: 'CRM' })).toBeInTheDocument();
  });
});
