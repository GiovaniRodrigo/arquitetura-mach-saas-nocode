// Métricas do Dashboard derivadas da listagem de sistemas do tenant (RF01).
// Enquanto não existir um endpoint dedicado de métricas agregadas, o total de
// sistemas é garantido; contadores por status dependem do payload enriquecido de
// Sistema (Fase 2) e permanecem indefinidos até lá.

import { useEffect, useState } from 'react';
import type { ApiClient } from '../api/client';

export interface Metricas {
  sistemas_total: number;
}

export type EstadoMetricas =
  | { fase: 'carregando' }
  | { fase: 'pronto'; metricas: Metricas }
  | { fase: 'erro'; mensagem: string };

export function useMetricas(client: ApiClient): { estado: EstadoMetricas; recarregar: () => void } {
  const [estado, setEstado] = useState<EstadoMetricas>({ fase: 'carregando' });
  const [tentativa, setTentativa] = useState(0);

  useEffect(() => {
    let vivo = true;
    setEstado({ fase: 'carregando' });
    (async () => {
      try {
        const sistemas = await client.listarSistemas();
        if (vivo) setEstado({ fase: 'pronto', metricas: { sistemas_total: sistemas.length } });
      } catch (e) {
        if (vivo) setEstado({ fase: 'erro', mensagem: e instanceof Error ? e.message : String(e) });
      }
    })();
    return () => {
      vivo = false;
    };
  }, [client, tentativa]);

  return { estado, recarregar: () => setTentativa((t) => t + 1) };
}
