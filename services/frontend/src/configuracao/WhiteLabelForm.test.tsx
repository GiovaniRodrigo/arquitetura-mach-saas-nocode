import { describe, it, expect, vi } from 'vitest';
import { screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { WhiteLabelForm } from './WhiteLabelForm';
import { renderDashboard, fakeClient } from '../test/renderDashboard';

describe('WhiteLabelForm (RF13, RNF03)', () => {
  it('salva logo, cores e domínio próprio', async () => {
    const atualizarWhiteLabel = vi.fn().mockResolvedValue({ validandoDominio: false });
    const client = fakeClient({ atualizarWhiteLabel });
    renderDashboard(<WhiteLabelForm />, { client });

    fireEvent.change(screen.getByLabelText(/logo/i), { target: { value: 'http://x/logo.png' } });
    fireEvent.change(screen.getByLabelText(/cor primária/i), { target: { value: '#112233' } });
    fireEvent.change(screen.getByLabelText(/cor secundária/i), { target: { value: '#445566' } });
    fireEvent.change(screen.getByLabelText(/domínio/i), { target: { value: 'app.parceiro.com' } });
    fireEvent.click(screen.getByRole('button', { name: /salvar/i }));

    await waitFor(() =>
      expect(atualizarWhiteLabel).toHaveBeenCalledWith({
        logo_url: 'http://x/logo.png',
        cor_primaria: '#112233',
        cor_secundaria: '#445566',
        dominio_proprio: 'app.parceiro.com',
      }),
    );
  });

  it('exibe estado "validando domínio" quando a API responde validação pendente', async () => {
    const atualizarWhiteLabel = vi.fn().mockResolvedValue({ validandoDominio: true });
    const client = fakeClient({ atualizarWhiteLabel });
    renderDashboard(<WhiteLabelForm />, { client });

    fireEvent.change(screen.getByLabelText(/domínio/i), { target: { value: 'app.parceiro.com' } });
    fireEvent.click(screen.getByRole('button', { name: /salvar/i }));

    expect(await screen.findByText(/validando domínio/i)).toBeInTheDocument();
  });
});
