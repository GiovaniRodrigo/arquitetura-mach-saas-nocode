// Fonte de listagem dos tenants (clientes/negócios) vinculados ao usuário
// autenticado (RF07), no mesmo molde de useSistemas.ts (RNF05 de 003).

import { useCallback, useEffect, useState } from 'react';
import type { ApiClient } from '../api/client';
import type { Tenant } from '../api/types';

export type EstadoTenants =
  | { fase: 'carregando' }
  | { fase: 'pronto'; tenants: Tenant[] }
  | { fase: 'vazio' }
  | { fase: 'erro'; mensagem: string };

export interface UseTenants {
  estado: EstadoTenants;
  recarregar: () => void;
}

export function useTenants(client: ApiClient): UseTenants {
  const [estado, setEstado] = useState<EstadoTenants>({ fase: 'carregando' });
  const [tentativa, setTentativa] = useState(0);

  useEffect(() => {
    let vivo = true;
    setEstado({ fase: 'carregando' });
    (async () => {
      try {
        const tenants = await client.listarTenants();
        if (!vivo) return;
        setEstado(tenants.length === 0 ? { fase: 'vazio' } : { fase: 'pronto', tenants });
      } catch (e) {
        if (vivo) setEstado({ fase: 'erro', mensagem: e instanceof Error ? e.message : String(e) });
      }
    })();
    return () => {
      vivo = false;
    };
  }, [client, tentativa]);

  const recarregar = useCallback(() => setTentativa((t) => t + 1), []);

  return { estado, recarregar };
}
