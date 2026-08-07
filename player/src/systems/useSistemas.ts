// Fonte única de listagem/criação de sistemas do tenant, consumida pelo seletor
// (SeletorSistemas) e pelas telas do dashboard (Clientes/Dashboard). Encapsula a
// máquina de estados carregando | pronto | vazio | erro e a ação de recarregar,
// eliminando a duplicação que antes vivia em SeletorSistemas (RF02, RF06, RNF05).

import { useCallback, useEffect, useState } from 'react';
import type { ApiClient } from '../api/client';
import type { Sistema } from '../api/types';

/** Estados possíveis do carregamento de sistemas (RF06). */
export type EstadoSistemas =
  | { fase: 'carregando' }
  | { fase: 'pronto'; sistemas: Sistema[] }
  | { fase: 'vazio' }
  | { fase: 'erro'; mensagem: string };

export interface UseSistemas {
  estado: EstadoSistemas;
  /** Refaz a listagem (ex.: botão "Tentar novamente"). */
  recarregar: () => void;
  /** Cria um sistema; repropaga ApiError (ex.: 403 → tratado na UI). */
  criar: (nome: string) => Promise<Sistema>;
}

export function useSistemas(client: ApiClient, tenantId?: string): UseSistemas {
  const [estado, setEstado] = useState<EstadoSistemas>({ fase: 'carregando' });
  const [tentativa, setTentativa] = useState(0);

  useEffect(() => {
    let vivo = true;
    setEstado({ fase: 'carregando' });
    (async () => {
      try {
        const lista = await client.listarSistemas(tenantId);
        if (!vivo) return;
        setEstado(lista.length === 0 ? { fase: 'vazio' } : { fase: 'pronto', sistemas: lista });
      } catch (e) {
        if (vivo) setEstado({ fase: 'erro', mensagem: e instanceof Error ? e.message : String(e) });
      }
    })();
    return () => {
      vivo = false;
    };
  }, [client, tenantId, tentativa]);

  const recarregar = useCallback(() => setTentativa((t) => t + 1), []);
  const criar = useCallback((nome: string) => client.criarSistema(nome), [client]);

  return { estado, recarregar, criar };
}
