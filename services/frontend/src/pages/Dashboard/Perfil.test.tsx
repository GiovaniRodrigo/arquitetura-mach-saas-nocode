import { describe, it, expect, vi } from 'vitest';
import { screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { Perfil } from './Perfil';
import { renderDashboard, fakeClient } from '../../test/renderDashboard';

describe('Page: Perfil (RF17-RF19, RN08)', () => {
  it('salva nome e foto diretamente ao editar (RF17)', async () => {
    const atualizarPerfil = vi.fn().mockResolvedValue(undefined);
    const client = fakeClient({ atualizarPerfil });
    renderDashboard(<Perfil />, { client });

    fireEvent.change(screen.getByLabelText(/nome/i), { target: { value: 'Ana Nova' } });
    fireEvent.click(screen.getByRole('button', { name: /salvar alterações/i }));

    await waitFor(() =>
      expect(atualizarPerfil).toHaveBeenCalledWith(
        expect.objectContaining({ nome: 'Ana Nova' }),
      ),
    );
  });

  it('troca de e-mail exige confirmação e não altera o e-mail exibido de imediato (RF18, RN08)', async () => {
    const solicitarTrocaEmail = vi.fn().mockResolvedValue(undefined);
    const client = fakeClient({ solicitarTrocaEmail });
    renderDashboard(<Perfil />, { client });

    fireEvent.change(screen.getByLabelText(/novo e-mail/i), { target: { value: 'novo@x.com' } });
    fireEvent.click(screen.getByRole('button', { name: /alterar e-mail/i }));

    await waitFor(() => expect(solicitarTrocaEmail).toHaveBeenCalledWith('novo@x.com'));
    expect(await screen.findByText(/confirme.*novo@x\.com/i)).toBeInTheDocument();
    // O e-mail atual (da sessão/JWT) continua exibido — a troca só efetiva após confirmação.
    expect(screen.getByText('ana@x.com')).toBeInTheDocument();
  });

  it('o link "Alterar senha" leva a Configuração > Segurança (RF19)', () => {
    renderDashboard(<Perfil />);
    const link = screen.getByRole('link', { name: /alterar senha/i });
    expect(link).toHaveAttribute('href', '/dashboard/configuracao#seguranca');
  });
});
