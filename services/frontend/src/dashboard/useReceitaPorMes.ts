// Estado do gráfico "Receita de assinatura": receita somada dos últimos 6
// meses dos tenants vinculados.

import { useCallback, useEffect, useState } from 'react';
import type { ApiClient } from '../api/client';
import type { PontoReceitaMensal } from '../api/types';

export type EstadoReceitaPorMes =
  | { fase: 'carregando' }
  | { fase: 'pronto'; pontos: PontoReceitaMensal[]; moeda: string }
  | { fase: 'erro'; mensagem: string };

export interface UseReceitaPorMes {
  estado: EstadoReceitaPorMes;
  recarregar: () => void;
}

export function useReceitaPorMes(client: ApiClient): UseReceitaPorMes {
  const [estado, setEstado] = useState<EstadoReceitaPorMes>({ fase: 'carregando' });
  const [tentativa, setTentativa] = useState(0);

  useEffect(() => {
    let vivo = true;
    setEstado({ fase: 'carregando' });
    (async () => {
      try {
        const { pontos, moeda } = await client.receitaPorMes();
        if (vivo) setEstado({ fase: 'pronto', pontos, moeda });
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
