// Estado do card "Resumo Financeiro" (RF06): receita de assinatura/cobrança
// da plataforma pelos tenants vinculados (RN04).

import { useCallback, useEffect, useState } from 'react';
import type { ApiClient } from '../api/client';
import type { ResumoFinanceiro } from '../api/types';

export type EstadoResumoFinanceiro =
  | { fase: 'carregando' }
  | { fase: 'pronto'; resumo: ResumoFinanceiro }
  | { fase: 'erro'; mensagem: string };

export interface UseResumoFinanceiro {
  estado: EstadoResumoFinanceiro;
  recarregar: () => void;
}

export function useResumoFinanceiro(client: ApiClient): UseResumoFinanceiro {
  const [estado, setEstado] = useState<EstadoResumoFinanceiro>({ fase: 'carregando' });
  const [tentativa, setTentativa] = useState(0);

  useEffect(() => {
    let vivo = true;
    setEstado({ fase: 'carregando' });
    (async () => {
      try {
        const resumo = await client.resumoFinanceiro();
        if (vivo) setEstado({ fase: 'pronto', resumo });
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
