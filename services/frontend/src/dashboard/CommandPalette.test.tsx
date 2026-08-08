import { describe, it, expect, vi, beforeEach } from 'vitest';
import { screen, fireEvent, waitFor } from '@testing-library/react';
import type { Sistema } from '../api/types';
import { CommandPalette } from './CommandPalette';
import { renderDashboard, fakeClient } from '../test/renderDashboard';

const sistemas: Sistema[] = [
  { id: 'a', nome: 'ERP Financeiro' },
  { id: 'b', nome: 'CRM de Parceiros' },
];

function abrirComAtalho() {
  fireEvent.keyDown(document, { key: 'k', ctrlKey: true });
}

describe('CommandPalette (RF10)', () => {
  beforeEach(() => {
    // fecha qualquer instância remanescente entre testes
    fireEvent.keyDown(document, { key: 'Escape' });
  });

  it('abre com Ctrl+K e lista ações + sistemas', async () => {
    const client = fakeClient({ listarSistemas: vi.fn().mockResolvedValue(sistemas) });
    renderDashboard(<CommandPalette />, { client });

    expect(screen.queryByRole('dialog')).toBeNull();
    abrirComAtalho();

    expect(await screen.findByRole('dialog')).toBeTruthy();
    expect(screen.getByText('Ir para Clientes')).toBeTruthy();
    await waitFor(() => expect(screen.getByText('Abrir ERP Financeiro')).toBeTruthy());
  });

  it('filtra os resultados pela busca', async () => {
    const client = fakeClient({ listarSistemas: vi.fn().mockResolvedValue(sistemas) });
    renderDashboard(<CommandPalette />, { client });
    abrirComAtalho();
    await screen.findByRole('dialog');

    fireEvent.change(screen.getByLabelText('Buscar'), { target: { value: 'CRM' } });
    expect(screen.getByText('Abrir CRM de Parceiros')).toBeTruthy();
    expect(screen.queryByText('Abrir ERP Financeiro')).toBeNull();
    expect(screen.queryByText('Ir para Clientes')).toBeNull();
  });

  it('fecha com Esc', async () => {
    const client = fakeClient({ listarSistemas: vi.fn().mockResolvedValue([]) });
    renderDashboard(<CommandPalette />, { client });
    abrirComAtalho();
    await screen.findByRole('dialog');

    fireEvent.keyDown(screen.getByRole('dialog'), { key: 'Escape' });
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
  });
});
