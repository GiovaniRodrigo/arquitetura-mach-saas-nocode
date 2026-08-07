import { describe, it, expect, vi } from 'vitest';
import { screen, waitFor, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import type { Tenant } from '../../api/types';
import { ApiError } from '../../api/client';
import { Clientes } from './Clientes';
import { renderDashboard, fakeClient, usuarioClienteFake } from '../../test/renderDashboard';

describe('Page: Clientes Dashboard (RF07, RN01)', () => {
  it('renderiza o título e a descrição', () => {
    renderDashboard(<Clientes />);
    expect(screen.getByText('Clientes')).toBeTruthy();
  });

  it('lista os tenants vinculados ao usuário autenticado (RF07)', async () => {
    const tenants: Tenant[] = [
      { id: 't1', nome: 'Acme' },
      { id: 't2', nome: 'Beta Ltda' },
    ];
    const client = fakeClient({ listarTenants: vi.fn().mockResolvedValue(tenants) });
    renderDashboard(<Clientes />, { client });

    expect(await screen.findByText('Acme')).toBeTruthy();
    expect(screen.getByText('Beta Ltda')).toBeTruthy();
    expect(screen.getAllByText(/Abrir cliente/i).length).toBe(2);
  });

  it('cada card de tenant linka para /dashboard/clientes/:tenantId', async () => {
    const tenants: Tenant[] = [{ id: 't1', nome: 'Acme' }];
    const client = fakeClient({ listarTenants: vi.fn().mockResolvedValue(tenants) });
    renderDashboard(<Clientes />, { client });

    const link = await screen.findByRole('link', { name: /abrir cliente/i });
    expect(link).toHaveAttribute('href', '/dashboard/clientes/t1');
  });

  it('exibe estado vazio quando não há tenants vinculados', async () => {
    const client = fakeClient({ listarTenants: vi.fn().mockResolvedValue([]) });
    renderDashboard(<Clientes />, { client });
    expect(await screen.findByText(/nenhum cliente ainda/i)).toBeTruthy();
  });

  it('exibe erro com retry quando a listagem falha', async () => {
    const client = fakeClient({ listarTenants: vi.fn().mockRejectedValue(new Error('boom')) });
    renderDashboard(<Clientes />, { client });
    await waitFor(() => expect(screen.getByRole('alert')).toBeTruthy());
    expect(screen.getByText('Tentar novamente')).toBeTruthy();
  });

  it('exibe o formulário de criação para dono/parceiro (RN10 de 003)', async () => {
    const client = fakeClient({ listarTenants: vi.fn().mockResolvedValue([]) });
    renderDashboard(<Clientes />, { client });

    expect(await screen.findByLabelText(/criar novo cliente/i)).toBeTruthy();
  });

  it('oculta o formulário de criação para cliente final (RN10 de 003)', async () => {
    const tenants: Tenant[] = [{ id: 't1', nome: 'Acme' }];
    const client = fakeClient({ listarTenants: vi.fn().mockResolvedValue(tenants) });
    renderDashboard(<Clientes />, { client, usuario: usuarioClienteFake });

    expect(await screen.findByText('Acme')).toBeTruthy();
    expect(screen.queryByLabelText(/criar novo cliente/i)).toBeNull();
  });

  it('cria um cliente e navega para a página dele (RF07)', async () => {
    const criarTenant = vi.fn().mockResolvedValue({ id: 't-novo', nome: 'Acme' });
    const client = fakeClient({ listarTenants: vi.fn().mockResolvedValue([]), criarTenant });
    renderDashboard(<Clientes />, { client });

    await screen.findByLabelText(/criar novo cliente/i);
    fireEvent.change(screen.getByLabelText(/criar novo cliente/i), { target: { value: 'Acme' } });
    fireEvent.click(screen.getByRole('button', { name: /criar/i }));

    await waitFor(() => expect(criarTenant).toHaveBeenCalledWith('Acme'));
  });

  it('exibe mensagem amigável quando a criação retorna 403', async () => {
    const criarTenant = vi.fn().mockRejectedValue(new ApiError(403, 'FORBIDDEN', 'sem permissão'));
    const client = fakeClient({ listarTenants: vi.fn().mockResolvedValue([]), criarTenant });
    renderDashboard(<Clientes />, { client });

    await screen.findByLabelText(/criar novo cliente/i);
    fireEvent.change(screen.getByLabelText(/criar novo cliente/i), { target: { value: 'Acme' } });
    fireEvent.click(screen.getByRole('button', { name: /criar/i }));

    expect(await screen.findByRole('alert')).toHaveTextContent(/não tem permissão/i);
  });
});
