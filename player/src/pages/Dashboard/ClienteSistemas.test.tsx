import { describe, it, expect, vi } from 'vitest';
import { screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import type { Sistema } from '../../api/types';
import { ClienteSistemas } from './ClienteSistemas';
import { renderDashboard, fakeClient } from '../../test/renderDashboard';

describe('Page: ClienteSistemas (RF08, RN05)', () => {
  it('lista os sistemas do tenant selecionado, filtrando por tenant_id', async () => {
    const sistemas: Sistema[] = [{ id: 's1', nome: 'ERP' }];
    const listarSistemas = vi.fn().mockResolvedValue(sistemas);
    const client = fakeClient({ listarSistemas });
    renderDashboard(<ClienteSistemas />, { client, rota: '/dashboard/clientes/t1', path: '/dashboard/clientes/:tenantId' });

    expect(await screen.findByText('ERP')).toBeInTheDocument();
    expect(listarSistemas).toHaveBeenCalledWith('t1');
  });

  it('"Abrir sistema" navega para as abas do sistema (Telas/Regras/Versão)', async () => {
    const sistemas: Sistema[] = [{ id: 's1', nome: 'ERP' }];
    const client = fakeClient({ listarSistemas: vi.fn().mockResolvedValue(sistemas) });
    renderDashboard(<ClienteSistemas />, { client, rota: '/dashboard/clientes/t1', path: '/dashboard/clientes/:tenantId' });

    const link = await screen.findByRole('link', { name: /abrir sistema/i });
    expect(link).toHaveAttribute('href', '/dashboard/clientes/t1/sistemas/s1');
  });

  it('exibe estado vazio quando o tenant não tem sistemas', async () => {
    const client = fakeClient({ listarSistemas: vi.fn().mockResolvedValue([]) });
    renderDashboard(<ClienteSistemas />, { client, rota: '/dashboard/clientes/t1', path: '/dashboard/clientes/:tenantId' });
    expect(await screen.findByText(/nenhum sistema/i)).toBeInTheDocument();
  });
});
