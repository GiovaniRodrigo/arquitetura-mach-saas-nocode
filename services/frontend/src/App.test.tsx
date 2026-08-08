import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import { App } from './App';
import { ApiClient } from './api/client';

function respostaJSON(status: number, corpo: unknown): Response {
  return new Response(JSON.stringify(corpo), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

describe('App — sistema sem versão publicada', () => {
  it('exibe mensagem amigável e link de volta ao Dashboard, não o ApiError cru', async () => {
    const fetchFn = vi
      .fn()
      .mockResolvedValue(respostaJSON(404, { codigo: 'NOT_FOUND', mensagem: 'recurso inexistente' }));
    const client = new ApiClient('http://gw', 'tok', fetchFn as unknown as typeof fetch);
    const config = { baseUrl: 'http://gw', token: 'tok', sistemaId: 's-sem-versao' };

    render(<App config={config} client={client} />);

    const alerta = await screen.findByRole('alert');
    expect(alerta).toHaveTextContent(/ainda não foi publicado/i);
    expect(alerta).not.toHaveTextContent(/apierror/i);

    const link = screen.getByRole('link', { name: /dashboard/i });
    expect(link.getAttribute('href')).toContain('dashboard');
  });

  it('para outros erros de API, mostra a mensagem original do servidor', async () => {
    const fetchFn = vi
      .fn()
      .mockResolvedValue(respostaJSON(500, { codigo: 'INTERNAL', mensagem: 'erro interno' }));
    const client = new ApiClient('http://gw', 'tok', fetchFn as unknown as typeof fetch);
    const config = { baseUrl: 'http://gw', token: 'tok', sistemaId: 's-com-erro' };

    render(<App config={config} client={client} />);

    expect(await screen.findByRole('alert')).toHaveTextContent(/erro interno/i);
  });
});
