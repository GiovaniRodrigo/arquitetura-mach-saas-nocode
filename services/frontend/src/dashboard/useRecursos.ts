// Estado da tela "Monitor" (RF06, RF07): status de recursos (uptime/memória) dos
// 8 serviços da plataforma, coletados pelo serviço Monitor via o Gateway. Uma
// entrada individual "indisponivel" (RN01) faz parte do estado 'pronto' —
// apenas a falha do endpoint inteiro (RNF02) entra em 'erro'.

import { useCallback, useEffect, useState } from 'react';
import type { ApiClient } from '../api/client';
import type { RecursosResponse } from '../api/types';

export type EstadoRecursos =
  | { fase: 'carregando' }
  | { fase: 'pronto'; recursos: RecursosResponse }
  | { fase: 'erro'; mensagem: string };

export interface UseRecursos {
  estado: EstadoRecursos;
  recarregar: () => void;
}

export interface UseRecursosOpcoes {
  /** Intervalo do auto-refresh em ms (RF07). Default: 10s. */
  intervaloMs?: number;
}

const INTERVALO_PADRAO_MS = 10000;

export function useRecursos(client: ApiClient, opcoes: UseRecursosOpcoes = {}): UseRecursos {
  const intervaloMs = opcoes.intervaloMs ?? INTERVALO_PADRAO_MS;
  const [estado, setEstado] = useState<EstadoRecursos>({ fase: 'carregando' });
  const [tentativa, setTentativa] = useState(0);

  const recarregar = useCallback(() => setTentativa((t) => t + 1), []);

  useEffect(() => {
    let vivo = true;
    setEstado({ fase: 'carregando' });
    (async () => {
      try {
        const recursos = await client.obterRecursos();
        if (vivo) setEstado({ fase: 'pronto', recursos });
      } catch (e) {
        if (vivo) setEstado({ fase: 'erro', mensagem: e instanceof Error ? e.message : String(e) });
      }
    })();
    return () => {
      vivo = false;
    };
  }, [client, tentativa]);

  // Auto-atualização (RF07, cenário 4): dispara a mesma busca em intervalo fixo
  // enquanto a tela estiver montada, limpo no cleanup para não vazar timers.
  useEffect(() => {
    const id = setInterval(() => recarregar(), intervaloMs);
    return () => clearInterval(id);
  }, [intervaloMs, recarregar]);

  return { estado, recarregar };
}
