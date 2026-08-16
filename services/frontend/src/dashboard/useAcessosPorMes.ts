// Estado do gráfico "Acessos por mês": contagem de logins dos últimos 6
// meses dos tenants vinculados.

import { useCallback, useEffect, useState } from 'react';
import type { ApiClient } from '../api/client';
import type { PontoAcessosMensal } from '../api/types';

export type EstadoAcessosPorMes =
  | { fase: 'carregando' }
  | { fase: 'pronto'; pontos: PontoAcessosMensal[] }
  | { fase: 'erro'; mensagem: string };

export interface UseAcessosPorMes {
  estado: EstadoAcessosPorMes;
  recarregar: () => void;
}

export function useAcessosPorMes(client: ApiClient): UseAcessosPorMes {
  const [estado, setEstado] = useState<EstadoAcessosPorMes>({ fase: 'carregando' });
  const [tentativa, setTentativa] = useState(0);

  useEffect(() => {
    let vivo = true;
    setEstado({ fase: 'carregando' });
    (async () => {
      try {
        const pontos = await client.acessosPorMes();
        if (vivo) setEstado({ fase: 'pronto', pontos });
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
