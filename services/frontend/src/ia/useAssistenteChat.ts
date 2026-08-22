// Estado do chat do Assistente de Design (RAG): histórico da conversa,
// envio de mensagem e tratamento de erro. Histórico fica em sessionStorage
// (useSessionStorageState) — sobrevive a fechar/reabrir o painel e trocar de
// aba do Dashboard, mas não é persistido no servidor (spec chat-ia-rag).

import { useCallback, useState } from "react";
import type { ApiClient } from "../api/client";
import { ApiError } from "../api/client";
import type { MensagemChatIA } from "../api/types";
import { useSessionStorageState } from "../systems/useSessionStorageState";

export interface UseAssistenteChat {
  historico: MensagemChatIA[];
  enviando: boolean;
  erro: string | null;
  enviar: (texto: string) => Promise<void>;
  limpar: () => void;
}

const CHAVE_SESSION_STORAGE = "assistenteDesignChatHistorico";

export function useAssistenteChat(client: ApiClient, sistemaNome?: string): UseAssistenteChat {
  const [historico, setHistorico] = useSessionStorageState<MensagemChatIA[]>(CHAVE_SESSION_STORAGE, []);
  const [enviando, setEnviando] = useState(false);
  const [erro, setErro] = useState<string | null>(null);

  const enviar = useCallback(
    async (texto: string) => {
      const mensagem = texto.trim();
      if (!mensagem || enviando) return;

      const comPergunta = [...historico, { papel: "usuario" as const, conteudo: mensagem }];
      setHistorico(comPergunta);
      setErro(null);
      setEnviando(true);
      try {
        const resposta = await client.enviarMensagemChatIA(comPergunta, sistemaNome);
        setHistorico([...comPergunta, { papel: "assistente" as const, conteudo: resposta }]);
      } catch (e) {
        setErro(e instanceof ApiError ? e.message : "Não consegui responder agora. Tente novamente em instantes.");
      } finally {
        setEnviando(false);
      }
    },
    [client, historico, sistemaNome, enviando, setHistorico],
  );

  const limpar = useCallback(() => {
    setHistorico([]);
    setErro(null);
  }, [setHistorico]);

  return { historico, enviando, erro, enviar, limpar };
}
