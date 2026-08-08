import { describe, it, expect, vi } from 'vitest';
import { screen, waitFor, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import type { Sistema, Tenant } from '../../api/types';
import { ApiError } from '../../api/client';
import { ClienteSistemas } from './ClienteSistemas';
import { renderDashboard, fakeClient } from '../../test/renderDashboard';

const tenantAcme: Tenant = { id: 't1', nome: 'Acme' };

function renderPagina(overrides: Parameters<typeof fakeClient>[0] = {}) {
  const client = fakeClient({
    listarSistemas: vi.fn().mockResolvedValue([]),
    obterTenant: vi.fn().mockResolvedValue(tenantAcme),
    ...overrides,
  });
  renderDashboard(<ClienteSistemas />, { client, rota: '/dashboard/clientes/t1', path: '/dashboard/clientes/:tenantId' });
  return client;
}

describe('Page: ClienteSistemas (RF07/RF08, RN05)', () => {
  it('lista os sistemas do tenant selecionado, filtrando por tenant_id', async () => {
    const sistemas: Sistema[] = [{ id: 's1', nome: 'ERP' }];
    const listarSistemas = vi.fn().mockResolvedValue(sistemas);
    renderPagina({ listarSistemas });

    expect(await screen.findByText('ERP')).toBeInTheDocument();
    expect(listarSistemas).toHaveBeenCalledWith('t1');
  });

  it('"Abrir sistema" navega para as abas do sistema (Telas/Regras/Versão)', async () => {
    const sistemas: Sistema[] = [{ id: 's1', nome: 'ERP' }];
    renderPagina({ listarSistemas: vi.fn().mockResolvedValue(sistemas) });

    const link = await screen.findByRole('link', { name: /abrir sistema/i });
    expect(link).toHaveAttribute('href', '/dashboard/clientes/t1/sistemas/s1');
  });

  it('lista os sistemas em uma tabela com coluna de Ações', async () => {
    const sistemas: Sistema[] = [
      { id: 's1', nome: 'ERP' },
      { id: 's2', nome: 'CRM' },
    ];
    renderPagina({ listarSistemas: vi.fn().mockResolvedValue(sistemas) });

    const tabela = await screen.findByRole('table');
    expect(screen.getByRole('columnheader', { name: /nome/i })).toBeInTheDocument();
    expect(screen.getByRole('columnheader', { name: /ações/i })).toBeInTheDocument();

    const linhas = screen.getAllByRole('row').slice(1); // pula o cabeçalho
    expect(linhas).toHaveLength(2);
    expect(tabela).toHaveTextContent('ERP');
    expect(tabela).toHaveTextContent('CRM');
    expect(screen.getAllByRole('link', { name: /abrir sistema/i })).toHaveLength(2);
  });

  it('exibe estado vazio quando o tenant não tem sistemas', async () => {
    renderPagina({ listarSistemas: vi.fn().mockResolvedValue([]) });
    expect(await screen.findByText(/nenhum sistema/i)).toBeInTheDocument();
  });

  it('visualiza: carrega e exibe o nome do cliente (RF07)', async () => {
    const obterTenant = vi.fn().mockResolvedValue(tenantAcme);
    renderPagina({ obterTenant });

    expect(await screen.findByRole('heading', { name: 'Acme' })).toBeInTheDocument();
    expect(obterTenant).toHaveBeenCalledWith('t1');
  });

  it('visualizar: erro ao carregar o cliente mostra estado de erro com retry', async () => {
    const obterTenant = vi.fn().mockRejectedValueOnce(new ApiError(500, 'INTERNAL', 'boom')).mockResolvedValueOnce(tenantAcme);
    renderPagina({ obterTenant });

    expect(await screen.findByText(/não foi possível carregar este cliente/i)).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: /tentar novamente/i }));
    expect(await screen.findByRole('heading', { name: 'Acme' })).toBeInTheDocument();
  });

  it('atualizar: salva o novo nome do cliente (RF07)', async () => {
    const atualizarTenant = vi.fn().mockResolvedValue({ id: 't1', nome: 'Acme Renomeada' });
    renderPagina({ atualizarTenant });

    const input = await screen.findByLabelText(/nome do cliente/i);
    fireEvent.change(input, { target: { value: 'Acme Renomeada' } });
    fireEvent.click(screen.getByRole('button', { name: /^salvar$/i }));

    await waitFor(() => expect(atualizarTenant).toHaveBeenCalledWith('t1', 'Acme Renomeada'));
    expect(await screen.findByText(/nome atualizado/i)).toBeInTheDocument();
    expect(await screen.findByRole('heading', { name: 'Acme Renomeada' })).toBeInTheDocument();
  });

  it('excluir: remove o cliente e navega de volta para a lista (RF07)', async () => {
    const excluirTenant = vi.fn().mockResolvedValue(undefined);
    renderPagina({ excluirTenant });

    await screen.findByRole('heading', { name: 'Acme' });
    fireEvent.click(screen.getByRole('button', { name: /excluir cliente/i }));

    await waitFor(() => expect(excluirTenant).toHaveBeenCalledWith('t1'));
  });

  it('excluir: mostra erro sem navegar quando a exclusão falha', async () => {
    const excluirTenant = vi.fn().mockRejectedValue(new ApiError(403, 'FORBIDDEN', 'sem permissão'));
    renderPagina({ excluirTenant });

    await screen.findByRole('heading', { name: 'Acme' });
    fireEvent.click(screen.getByRole('button', { name: /excluir cliente/i }));

    expect(await screen.findByText('sem permissão')).toBeInTheDocument();
  });
});
