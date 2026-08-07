// Estado do card "Reclamações/Feedback" (RF05): mensagens dos tenants vinculados,
// com status pendente/respondido (RN03).

import { useCallback, useEffect, useState } from 'react';
import type { ApiClient } from '../api/client';
import type { Feedback } from '../api/types';

export type EstadoFeedback =
  | { fase: 'carregando' }
  | { fase: 'pronto'; itens: Feedback[] }
  | { fase: 'vazio' }
  | { fase: 'erro'; mensagem: string };

export interface UseFeedback {
  estado: EstadoFeedback;
  recarregar: () => void;
  /** Marca uma mensagem como respondida (RN03) e recarrega a lista. */
  marcarRespondido: (id: string) => Promise<void>;
}

export function useFeedback(client: ApiClient): UseFeedback {
  const [estado, setEstado] = useState<EstadoFeedback>({ fase: 'carregando' });
  const [tentativa, setTentativa] = useState(0);

  useEffect(() => {
    let vivo = true;
    setEstado({ fase: 'carregando' });
    (async () => {
      try {
        const itens = await client.listarFeedback();
        if (!vivo) return;
        setEstado(itens.length === 0 ? { fase: 'vazio' } : { fase: 'pronto', itens });
      } catch (e) {
        if (vivo) setEstado({ fase: 'erro', mensagem: e instanceof Error ? e.message : String(e) });
      }
    })();
    return () => {
      vivo = false;
    };
  }, [client, tentativa]);

  const recarregar = useCallback(() => setTentativa((t) => t + 1), []);
  const marcarRespondido = useCallback(
    async (id: string) => {
      await client.atualizarStatusFeedback(id, 'respondido');
      setTentativa((t) => t + 1);
    },
    [client],
  );

  return { estado, recarregar, marcarRespondido };
}
