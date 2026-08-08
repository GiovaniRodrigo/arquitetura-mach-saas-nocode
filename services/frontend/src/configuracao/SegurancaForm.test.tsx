import { describe, it, expect, vi } from 'vitest';
import { screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { ApiError } from '../api/client';
import { SegurancaForm } from './SegurancaForm';
import { renderDashboard, fakeClient } from '../test/renderDashboard';

describe('SegurancaForm (RF14-RF16, RN07, RNF01, RNF02)', () => {
  it('troca a senha informando senha atual e nova (RF14)', async () => {
    const atualizarSenha = vi.fn().mockResolvedValue(undefined);
    const client = fakeClient({ atualizarSenha });
    renderDashboard(<SegurancaForm />, { client });

    fireEvent.change(screen.getByLabelText('Senha atual'), { target: { value: 'atual123' } });
    fireEvent.change(screen.getByLabelText(/^nova senha/i), { target: { value: 'nova123' } });
    fireEvent.click(screen.getByRole('button', { name: /atualizar senha/i }));

    await waitFor(() => expect(atualizarSenha).toHaveBeenCalledWith('atual123', 'nova123'));
  });

  it('ativa o MFA em duas etapas: QR code exibido uma única vez e some após confirmar (RF15, RNF01)', async () => {
    const ativarMfa = vi.fn().mockResolvedValue({ segredoOtpAuthUri: 'otpauth://totp/x' });
    const confirmarMfa = vi.fn().mockResolvedValue(undefined);
    const client = fakeClient({ ativarMfa, confirmarMfa });
    renderDashboard(<SegurancaForm />, { client });

    fireEvent.click(screen.getByRole('button', { name: /ativar mfa/i }));
    expect(await screen.findByText(/otpauth:\/\/totp\/x/)).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText(/código/i), { target: { value: '654321' } });
    fireEvent.click(screen.getByRole('button', { name: /confirmar/i }));

    await waitFor(() => expect(confirmarMfa).toHaveBeenCalledWith('654321'));
    await waitFor(() => expect(screen.queryByText(/otpauth:\/\/totp\/x/)).not.toBeInTheDocument());
  });

  it('bloqueia a exclusão de conta quando há tenant ativo vinculado (RF16, RN07)', async () => {
    const excluirConta = vi
      .fn()
      .mockRejectedValue(new ApiError(409, 'TENANT_ATIVO_VINCULADO', 'Existem tenants ativos vinculados a esta conta.'));
    const client = fakeClient({ excluirConta });
    renderDashboard(<SegurancaForm />, { client });

    fireEvent.change(screen.getByLabelText(/senha atual \(excluir conta\)/i), { target: { value: 'minha-senha' } });
    fireEvent.click(screen.getByRole('button', { name: /excluir conta/i }));

    expect(await screen.findByRole('alert')).toHaveTextContent(/tenants ativos vinculados/i);
  });

  it('excluirConta envia a senha atual digitada como reautenticação (RF16, RNF02)', async () => {
    const excluirConta = vi.fn().mockResolvedValue(undefined);
    const client = fakeClient({ excluirConta });
    renderDashboard(<SegurancaForm />, { client });

    fireEvent.change(screen.getByLabelText(/senha atual \(excluir conta\)/i), { target: { value: 'senha-correta' } });
    fireEvent.click(screen.getByRole('button', { name: /excluir conta/i }));

    await waitFor(() => expect(excluirConta).toHaveBeenCalledWith('senha-correta'));
  });

  it('desativarMfa envia a senha atual digitada como reautenticação (RF15, RNF02)', async () => {
    const ativarMfa = vi.fn().mockResolvedValue({ segredoOtpAuthUri: 'otpauth://totp/x' });
    const confirmarMfa = vi.fn().mockResolvedValue(undefined);
    const desativarMfa = vi.fn().mockResolvedValue(undefined);
    const client = fakeClient({ ativarMfa, confirmarMfa, desativarMfa });
    renderDashboard(<SegurancaForm />, { client });

    fireEvent.click(screen.getByRole('button', { name: /ativar mfa/i }));
    await screen.findByText(/otpauth:\/\/totp\/x/);
    fireEvent.change(screen.getByLabelText(/código/i), { target: { value: '654321' } });
    fireEvent.click(screen.getByRole('button', { name: /confirmar/i }));
    await waitFor(() => expect(confirmarMfa).toHaveBeenCalled());

    fireEvent.change(screen.getByLabelText(/senha atual \(desativar mfa\)/i), { target: { value: 'senha-do-usuario' } });
    fireEvent.click(screen.getByRole('button', { name: /desativar mfa/i }));

    await waitFor(() => expect(desativarMfa).toHaveBeenCalledWith('senha-do-usuario'));
  });
});
