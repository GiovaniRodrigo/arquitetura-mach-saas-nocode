// Fonte única de listagem/criação de telas (Designs) de um sistema, consumida
// pela aba Telas (RF09). Mesma máquina de estados carregando | pronto | vazio |
// erro de useSistemas, para diferenciar "sem telas ainda" de erro de
// comunicação com o Gateway/Design Engine.

import { useCallback, useEffect, useState } from 'react';
import type { ApiClient } from '../api/client';
import type { DesignResumo } from '../api/types';
import { gerarBlindIndex } from '../collab/treeOps';

export type EstadoTelas =
  | { fase: 'carregando' }
  | { fase: 'pronto'; telas: DesignResumo[] }
  | { fase: 'vazio' }
  | { fase: 'erro'; mensagem: string };

export interface UseTelas {
  estado: EstadoTelas;
  recarregar: () => void;
  /** Cria uma tela vazia (raiz sem filhos) e devolve seu id. */
  criarTela: (nome: string) => Promise<string>;
  removerTela: (id: string) => Promise<void>;
}

export function useTelas(client: ApiClient, sistemaId: string): UseTelas {
  const [estado, setEstado] = useState<EstadoTelas>({ fase: 'carregando' });
  const [tentativa, setTentativa] = useState(0);

  useEffect(() => {
    let vivo = true;
    setEstado({ fase: 'carregando' });
    (async () => {
      try {
        const lista = await client.listarTelas(sistemaId);
        if (!vivo) return;
        setEstado(lista.length === 0 ? { fase: 'vazio' } : { fase: 'pronto', telas: lista });
      } catch (e) {
        if (vivo) setEstado({ fase: 'erro', mensagem: e instanceof Error ? e.message : String(e) });
      }
    })();
    return () => {
      vivo = false;
    };
  }, [client, sistemaId, tentativa]);

  const recarregar = useCallback(() => setTentativa((t) => t + 1), []);

  const criarTela = useCallback(
    async (nome: string) => {
      const { id } = await client.criarDesign({
        sistemaId,
        nome,
        arvore: { blind_index: gerarBlindIndex(), tipo: 'tela', componente_filhos: [] },
      });
      recarregar();
      return id;
    },
    [client, sistemaId, recarregar],
  );

  const removerTela = useCallback(
    async (id: string) => {
      await client.removerDesign(id);
      recarregar();
    },
    [client, recarregar],
  );

  return { estado, recarregar, criarTela, removerTela };
}
