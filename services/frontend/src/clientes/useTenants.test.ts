import { describe, expect, it, vi } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import type { ApiClient } from '../api/client';
import { ApiError } from '../api/client';
import type { Tenant } from '../api/types';
import { useTenants } from './useTenants';

function fakeClient(over: Partial<ApiClient>): ApiClient {
  return over as unknown as ApiClient;
}

describe('useTenants (RF07)', () => {
  it('passa de carregando para pronto com a lista', async () => {
    const tenants: Tenant[] = [{ id: 't1', nome: 'Acme' }];
    const client = fakeClient({ listarTenants: vi.fn().mockResolvedValue(tenants) });

    const { result } = renderHook(() => useTenants(client));
    expect(result.current.estado.fase).toBe('carregando');

    await waitFor(() => expect(result.current.estado.fase).toBe('pronto'));
    expect(result.current.estado).toEqual({ fase: 'pronto', tenants });
  });

  it('entra em vazio quando não há tenants', async () => {
    const client = fakeClient({ listarTenants: vi.fn().mockResolvedValue([]) });
    const { result } = renderHook(() => useTenants(client));
    await waitFor(() => expect(result.current.estado.fase).toBe('vazio'));
  });

  it('entra em erro e recarrega com sucesso', async () => {
    const listar = vi
      .fn()
      .mockRejectedValueOnce(new ApiError(500, 'INTERNAL', 'boom'))
      .mockResolvedValueOnce([{ id: 't1', nome: 'Acme' }]);
    const client = fakeClient({ listarTenants: listar });

    const { result } = renderHook(() => useTenants(client));
    await waitFor(() => expect(result.current.estado.fase).toBe('erro'));

    act(() => result.current.recarregar());
    await waitFor(() => expect(result.current.estado.fase).toBe('pronto'));
    expect(listar).toHaveBeenCalledTimes(2);
  });

  it('criar delega ao client.criarTenant e devolve o tenant criado', async () => {
    const novo: Tenant = { id: 't9', nome: 'Acme' };
    const criarTenant = vi.fn().mockResolvedValue(novo);
    const client = fakeClient({ listarTenants: vi.fn().mockResolvedValue([]), criarTenant });

    const { result } = renderHook(() => useTenants(client));
    await waitFor(() => expect(result.current.estado.fase).toBe('vazio'));

    const criado = await act(() => result.current.criar('Acme'));
    expect(criarTenant).toHaveBeenCalledWith('Acme');
    expect(criado).toEqual(novo);
  });

  it('criar repropaga ApiError (ex.: 403 de cliente final)', async () => {
    const erro = new ApiError(403, 'FORBIDDEN', 'sem permissão');
    const client = fakeClient({
      listarTenants: vi.fn().mockResolvedValue([]),
      criarTenant: vi.fn().mockRejectedValue(erro),
    });

    const { result } = renderHook(() => useTenants(client));
    await waitFor(() => expect(result.current.estado.fase).toBe('vazio'));

    await expect(result.current.criar('Acme')).rejects.toBe(erro);
  });
});
