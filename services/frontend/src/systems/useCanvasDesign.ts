// Canvas colaborativo de uma tela (RF09/RF06): busca a árvore canônica via
// REST, junta-se ao canal de colaboração (services/collab) e mantém a árvore
// local sincronizada — mutações próprias aplicadas otimisticamente e enviadas
// ao canal, mutações remotas aplicadas ao chegar. A persistência em si
// (write-behind debounced) é responsabilidade exclusiva do collab; este hook
// nunca chama AtualizarDesign diretamente.
//
// Também mantém undo/redo (pilha de snapshots de mutações LOCAIS apenas —
// mutações remotas de outros usuários não entram no histórico) e um indicador
// de status de salvamento aproximado, alinhado ao debounce de 5s do collab
// (RN06): qualquer mutação local marca "salvando" e volta a "salvo" depois de
// DEBOUNCE_FLUSH_MS + margem sem nova mutação.

import { useCallback, useEffect, useRef, useState } from 'react';
import type { ApiClient } from '../api/client';
import type { Componente } from '../api/types';
import { obterToken } from '../auth/session';
import { CollabClient } from '../collab/phoenixSocket';
import {
  aplicarMutacao,
  clonarComNovosIndices,
  encontrarNo,
  encontrarPai,
  gerarBlindIndex,
  type Mutacao,
} from '../collab/treeOps';
import { registroDoTipo } from './componentRegistry';
import { useSessionStorageState } from './useSessionStorageState';

export type EstadoCanvas =
  | { fase: 'carregando' }
  | { fase: 'pronto' }
  | { fase: 'erro'; mensagem: string };

export type StatusSalvamento = 'salvo' | 'salvando';

const DEBOUNCE_FLUSH_MS = 5000;
const MARGEM_STATUS_MS = 600;

export interface UseCanvasDesign {
  estado: EstadoCanvas;
  arvore: Componente | null;
  selecionado: string | null;
  statusSalvamento: StatusSalvamento;
  podeDesfazer: boolean;
  podeRefazer: boolean;
  selecionar: (blindIndex: string | null) => void;
  adicionarComponente: (tipo: string, parentBlindIndex?: string, index?: number) => void;
  moverNo: (blindIndex: string, novoParent: string, index?: number) => void;
  atualizarPropriedades: (blindIndex: string, propriedades: Record<string, unknown>) => void;
  removerComponente: (blindIndex: string) => void;
  duplicarComponente: (blindIndex: string) => void;
  desfazer: () => void;
  refazer: () => void;
}

/** `/socket` é proxiado ao collab tanto em dev (vite.config.ts) quanto em produção (Nginx). */
function urlSocket(): string {
  const protocolo = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  return `${protocolo}//${window.location.host}/socket`;
}

function vazia(tree: unknown): boolean {
  return !tree || typeof tree !== 'object' || Object.keys(tree).length === 0;
}

export function useCanvasDesign(
  client: ApiClient,
  params: { designId: string; sistemaId: string; nome: string },
): UseCanvasDesign {
  const { designId, sistemaId, nome } = params;
  const [estado, setEstado] = useState<EstadoCanvas>({ fase: 'carregando' });
  const [arvore, setArvore] = useState<Componente | null>(null);
  // sessionStorage (não useState puro): sobrevive a trocar de aba (Telas ↔
  // Regras/Versão desmonta este hook via <Outlet/>) e voltar, mantendo a
  // seleção "do jeito que deixei".
  const [selecionado, setSelecionado] = useSessionStorageState<string | null>(
    `mach:tela:${designId}:componente-selecionado`,
    null,
  );
  const [statusSalvamento, setStatusSalvamento] = useState<StatusSalvamento>('salvo');
  const [historico, setHistorico] = useState<Componente[]>([]);
  const [futuro, setFuturo] = useState<Componente[]>([]);
  const collabRef = useRef<CollabClient | null>(null);
  const timerStatusRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    let vivo = true;
    setEstado({ fase: 'carregando' });
    setArvore(null);
    setHistorico([]);
    setFuturo([]);
    setStatusSalvamento('salvo');

    const collab = new CollabClient(urlSocket(), obterToken());
    collabRef.current = collab;

    (async () => {
      try {
        const design = await client.obterDesign(designId);
        if (!vivo) return;

        await collab.entrar(
          designId,
          {
            onMutation: (mutacao) => {
              if (!vivo) return;
              setArvore((atual) => (atual ? aplicarMutacao(atual, mutacao as Mutacao) : atual));
            },
            onJoin: (treeRecebida) => {
              if (!vivo) return;
              if (vazia(treeRecebida) && design.arvore) {
                // Sessão nova do collab (sem snapshot em Redis) — hidrata com o
                // que está persistido, para não perder o que já existe.
                collab.mutar({ tipo: 'set_tree', arvore: design.arvore });
                setArvore(design.arvore);
              } else {
                setArvore(treeRecebida as Componente);
              }
            },
          },
          { sistema_id: sistemaId, design_id: designId, nome },
        );
        if (!vivo) return;
        setEstado({ fase: 'pronto' });
      } catch (e) {
        if (vivo) setEstado({ fase: 'erro', mensagem: e instanceof Error ? e.message : String(e) });
      }
    })();

    return () => {
      vivo = false;
      collab.sair();
      collabRef.current = null;
    };
  }, [client, designId, sistemaId, nome]);

  useEffect(
    () => () => {
      if (timerStatusRef.current) clearTimeout(timerStatusRef.current);
    },
    [],
  );

  const marcarSalvando = useCallback(() => {
    setStatusSalvamento('salvando');
    if (timerStatusRef.current) clearTimeout(timerStatusRef.current);
    timerStatusRef.current = setTimeout(() => setStatusSalvamento('salvo'), DEBOUNCE_FLUSH_MS + MARGEM_STATUS_MS);
  }, []);

  // Mutação originada pelo próprio usuário: aplica otimista, envia ao collab,
  // e empilha o estado anterior no histórico de undo (limpando o de redo).
  const aplicarLocal = useCallback(
    (mutacao: Mutacao) => {
      if (!arvore) return;
      setHistorico((h) => [...h, arvore]);
      setFuturo([]);
      setArvore(aplicarMutacao(arvore, mutacao));
      collabRef.current?.mutar(mutacao);
      marcarSalvando();
    },
    [arvore, marcarSalvando],
  );

  const selecionar = useCallback((blindIndex: string | null) => setSelecionado(blindIndex), []);

  const adicionarComponente = useCallback(
    (tipo: string, parentBlindIndex?: string, index?: number) => {
      if (!arvore) return;
      const registro = registroDoTipo(tipo);
      const props = registro?.propriedadesPadrao() ?? { estilos: {} };
      const no: Componente = {
        blind_index: gerarBlindIndex(),
        tipo,
        propriedades: props,
        componente_filhos: [],
      };
      aplicarLocal({ tipo: 'add_child', parent: parentBlindIndex ?? arvore.blind_index, no, index });
      setSelecionado(no.blind_index);
    },
    [arvore, aplicarLocal],
  );

  const moverNo = useCallback(
    (blindIndex: string, novoParent: string, index?: number) => {
      aplicarLocal({ tipo: 'move', blind_index: blindIndex, novo_parent: novoParent, index });
    },
    [aplicarLocal],
  );

  const atualizarPropriedades = useCallback(
    (blindIndex: string, propriedades: Record<string, unknown>) => {
      aplicarLocal({ tipo: 'update_props', blind_index: blindIndex, propriedades });
    },
    [aplicarLocal],
  );

  const removerComponente = useCallback(
    (blindIndex: string) => {
      aplicarLocal({ tipo: 'remove', blind_index: blindIndex });
      setSelecionado((s) => (s === blindIndex ? null : s));
    },
    [aplicarLocal],
  );

  const duplicarComponente = useCallback(
    (blindIndex: string) => {
      if (!arvore) return;
      const no = encontrarNo(arvore, blindIndex);
      const pai = encontrarPai(arvore, blindIndex);
      if (!no || !pai) return;
      const clone = clonarComNovosIndices(no);
      aplicarLocal({ tipo: 'add_child', parent: pai.pai.blind_index, no: clone, index: pai.index + 1 });
      setSelecionado(clone.blind_index);
    },
    [arvore, aplicarLocal],
  );

  const desfazer = useCallback(() => {
    setHistorico((h) => {
      if (h.length === 0 || !arvore) return h;
      const anterior = h[h.length - 1];
      setFuturo((f) => [...f, arvore]);
      setArvore(anterior);
      collabRef.current?.mutar({ tipo: 'set_tree', arvore: anterior });
      marcarSalvando();
      return h.slice(0, -1);
    });
  }, [arvore, marcarSalvando]);

  const refazer = useCallback(() => {
    setFuturo((f) => {
      if (f.length === 0 || !arvore) return f;
      const proximo = f[f.length - 1];
      setHistorico((h) => [...h, arvore]);
      setArvore(proximo);
      collabRef.current?.mutar({ tipo: 'set_tree', arvore: proximo });
      marcarSalvando();
      return f.slice(0, -1);
    });
  }, [arvore, marcarSalvando]);

  return {
    estado,
    arvore,
    selecionado,
    statusSalvamento,
    podeDesfazer: historico.length > 0,
    podeRefazer: futuro.length > 0,
    selecionar,
    adicionarComponente,
    moverNo,
    atualizarPropriedades,
    removerComponente,
    duplicarComponente,
    desfazer,
    refazer,
  };
}
