import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { MemoryRouter } from 'react-router-dom';
import { Register } from './Register';
import { obterToken, encerrarSessao } from './session';

function respostaJSON(status: number, corpo: unknown): Response {
  return new Response(JSON.stringify(corpo), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

function preencherEEnviar() {
  fireEvent.change(screen.getByLabelText(/^nome$/i), { target: { value: 'Ana' } });
  fireEvent.change(screen.getByLabelText(/e-mail/i), { target: { value: 'ana@example.com' } });
  fireEvent.change(screen.getByLabelText(/senha/i), { target: { value: '12345678' } });
  fireEvent.change(screen.getByLabelText(/negócio|empresa|tenant/i), { target: { value: 'Ana LTDA' } });
  fireEvent.click(screen.getByRole('button', { name: /criar conta|cadastr/i }));
}

describe('Register (spec 006, RF02/RF04)', () => {
  beforeEach(() => {
    encerrarSessao();
  });

  it('renderiza o formulário de cadastro com nome, e-mail, senha e nome do negócio', () => {
    render(
      <MemoryRouter initialEntries={['/register']}>
        <Register fetchFn={vi.fn()} />
      </MemoryRouter>,
    );
    expect(screen.getByLabelText(/^nome$/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/e-mail/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/senha/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/negócio|empresa|tenant/i)).toBeInTheDocument();
  });

  it('em sucesso, salva o token e redireciona autenticado (RF04, Critério 8)', async () => {
    const fetchFn = vi.fn().mockResolvedValue(
      respostaJSON(201, { jwt: 'tok-123', user_id: 'u1', tenant_id: 't1', tipo: 'dono' }),
    );
    const redirecionarPara = vi.fn();

    render(
      <MemoryRouter initialEntries={['/register']}>
        <Register fetchFn={fetchFn} redirecionarPara={redirecionarPara} />
      </MemoryRouter>,
    );
    preencherEEnviar();

    await waitFor(() => expect(redirecionarPara).toHaveBeenCalled());
    expect(obterToken()).toBe('tok-123');

    const [url, init] = fetchFn.mock.calls[0];
    expect(url).toBe('/api/v1/auth/registro');
    const corpoEnviado = JSON.parse((init as RequestInit).body as string);
    expect(corpoEnviado).toEqual({
      nome: 'Ana',
      email: 'ana@example.com',
      senha: '12345678',
      nome_tenant: 'Ana LTDA',
    });
  });

  it('em e-mail duplicado, exibe erro e preserva os campos preenchidos (RF05)', async () => {
    const fetchFn = vi.fn().mockResolvedValue(
      respostaJSON(409, { codigo: 'EMAIL_DUPLICADO', mensagem: 'e-mail já cadastrado' }),
    );
    const redirecionarPara = vi.fn();

    render(
      <MemoryRouter initialEntries={['/register']}>
        <Register fetchFn={fetchFn} redirecionarPara={redirecionarPara} />
      </MemoryRouter>,
    );
    preencherEEnviar();

    expect(await screen.findByRole('alert')).toHaveTextContent(/e-mail já cadastrado/i);
    expect(redirecionarPara).not.toHaveBeenCalled();
    expect(obterToken()).toBe('');
    // Campos preservados — nada foi limpo após o erro.
    expect(screen.getByLabelText(/^nome$/i)).toHaveValue('Ana');
    expect(screen.getByLabelText(/e-mail/i)).toHaveValue('ana@example.com');
  });
});
