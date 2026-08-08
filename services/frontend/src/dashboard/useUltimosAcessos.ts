// Estado do card "Últimos Acessos" (RF04): 10 logins mais recentes agregados
// dos tenants vinculados ao usuário autenticado, sem filtro de período (RN02).

import { useCallback, useEffect, useState } from 'react';
import type { ApiClient } from '../api/client';
import type { EventoLogin } from '../api/types';

export type EstadoUltimosAcessos =
  | { fase: 'carregando' }
  | { fase: 'pronto'; eventos: EventoLogin[] }
  | { fase: 'vazio' }
  | { fase: 'erro'; mensagem: string };

export interface UseUltimosAcessos {
  estado: EstadoUltimosAcessos;
  recarregar: () => void;
}

export function useUltimosAcessos(client: ApiClient): UseUltimosAcessos {
  const [estado, setEstado] = useState<EstadoUltimosAcessos>({ fase: 'carregando' });
  const [tentativa, setTentativa] = useState(0);

  useEffect(() => {
    let vivo = true;
    setEstado({ fase: 'carregando' });
    (async () => {
      try {
        const eventos = await client.listarUltimosAcessos();
        if (!vivo) return;
        setEstado(eventos.length === 0 ? { fase: 'vazio' } : { fase: 'pronto', eventos });
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
