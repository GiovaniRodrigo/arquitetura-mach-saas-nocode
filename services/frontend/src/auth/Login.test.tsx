import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { MemoryRouter } from 'react-router-dom';
import { Login } from './Login';
import { obterToken, encerrarSessao } from './session';

function respostaJSON(status: number, corpo: unknown): Response {
  return new Response(JSON.stringify(corpo), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

describe('Login (spec 001 RF03 + spec 006 RF01/RF06)', () => {
  beforeEach(() => {
    encerrarSessao();
  });

  it('continua oferecendo os botões OAuth existentes (RNF04)', () => {
    render(
      <MemoryRouter>
        <Login fetchFn={vi.fn()} />
      </MemoryRouter>,
    );
    expect(screen.getByRole('link', { name: /entrar com google/i })).toHaveAttribute(
      'href',
      expect.stringContaining('/auth/google?redirect_uri='),
    );
    expect(screen.getByRole('link', { name: /entrar com github/i })).toBeInTheDocument();
  });

  it('exibe um link "Cadastre-se" para /register (RF01)', () => {
    render(
      <MemoryRouter>
        <Login fetchFn={vi.fn()} />
      </MemoryRouter>,
    );
    expect(screen.getByRole('link', { name: /cadastr/i })).toHaveAttribute('href', '/register');
  });

  it('autentica por e-mail/senha e salva o token em sucesso (RF06)', async () => {
    const fetchFn = vi.fn().mockResolvedValue(
      respostaJSON(200, { jwt: 'tok-abc', user_id: 'u1', tenant_id: 't1', tipo: 'dono' }),
    );
    const redirecionarPara = vi.fn();

    render(
      <MemoryRouter>
        <Login fetchFn={fetchFn} redirecionarPara={redirecionarPara} />
      </MemoryRouter>,
    );

    fireEvent.change(screen.getByLabelText(/e-mail/i), { target: { value: 'ana@example.com' } });
    fireEvent.change(screen.getByLabelText(/senha/i), { target: { value: '12345678' } });
    fireEvent.click(screen.getByRole('button', { name: /^entrar$/i }));

    await waitFor(() => expect(redirecionarPara).toHaveBeenCalled());
    expect(obterToken()).toBe('tok-abc');

    const [url, init] = fetchFn.mock.calls[0];
    expect(url).toBe('/api/v1/auth/login');
    expect(JSON.parse((init as RequestInit).body as string)).toEqual({ email: 'ana@example.com', senha: '12345678' });
  });

  it('em credenciais inválidas, exibe erro genérico (RN04)', async () => {
    const fetchFn = vi.fn().mockResolvedValue(
      respostaJSON(401, { codigo: 'CREDENCIAIS_INVALIDAS', mensagem: 'credenciais inválidas' }),
    );

    render(
      <MemoryRouter>
        <Login fetchFn={fetchFn} redirecionarPara={vi.fn()} />
      </MemoryRouter>,
    );

    fireEvent.change(screen.getByLabelText(/e-mail/i), { target: { value: 'ana@example.com' } });
    fireEvent.change(screen.getByLabelText(/senha/i), { target: { value: 'errada' } });
    fireEvent.click(screen.getByRole('button', { name: /^entrar$/i }));

    expect(await screen.findByRole('alert')).toHaveTextContent(/credenciais inválidas/i);
    expect(obterToken()).toBe('');
  });
});
