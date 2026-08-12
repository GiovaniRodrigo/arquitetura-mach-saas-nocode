// Componentes de formulário (categoria 'formulario' do componentRegistry)
// coletados de todas as telas de um sistema, para a aba Regras de Negócio
// escolher visualmente o componente-alvo de uma regra (RF10/RF11) em vez de
// digitar o blind_index de cabeça.

import { useCallback, useEffect, useState } from 'react';
import type { ApiClient } from '../api/client';
import type { Componente, DesignResumo } from '../api/types';
import { registroDoTipo } from './componentRegistry';

export interface ComponenteCampo {
  blindIndex: string;
  tipo: string;
  rotulo: string;
  telaId: string;
  telaNome: string;
}

export type EstadoComponentesFormulario =
  | { fase: 'carregando' }
  | { fase: 'pronto'; telas: DesignResumo[]; itens: ComponenteCampo[]; arvores: Record<string, Componente> }
  | { fase: 'vazio' }
  | { fase: 'erro'; mensagem: string };

function rotularCampo(no: Componente): string {
  const texto = typeof no.propriedades?.texto === 'string' ? no.propriedades.texto.trim() : '';
  if (texto) return texto;
  const bruto = no.blind_index.replace(/^bi-/, '').replace(/[-_]+/g, ' ').trim();
  return bruto ? bruto.replace(/\b\w/g, (c) => c.toUpperCase()) : no.blind_index;
}

function coletarCampos(no: Componente, telaId: string, telaNome: string, out: ComponenteCampo[]): void {
  if (registroDoTipo(no.tipo)?.categoria === 'formulario') {
    out.push({ blindIndex: no.blind_index, tipo: no.tipo, rotulo: rotularCampo(no), telaId, telaNome });
  }
  for (const filho of no.componente_filhos ?? []) coletarCampos(filho, telaId, telaNome, out);
}

export function useComponentesFormulario(client: ApiClient, sistemaId: string) {
  const [estado, setEstado] = useState<EstadoComponentesFormulario>({ fase: 'carregando' });
  const [tentativa, setTentativa] = useState(0);

  useEffect(() => {
    let vivo = true;
    setEstado({ fase: 'carregando' });
    (async () => {
      try {
        const telas = await client.listarTelas(sistemaId);
        const designs = await Promise.all(telas.map((tela) => client.obterDesign(tela.id)));
        if (!vivo) return;
        const arvores: Record<string, Componente> = {};
        const itens: ComponenteCampo[] = [];
        telas.forEach((tela, i) => {
          arvores[tela.id] = designs[i].arvore;
          coletarCampos(designs[i].arvore, tela.id, tela.nome, itens);
        });
        setEstado(telas.length === 0 ? { fase: 'vazio' } : { fase: 'pronto', telas, itens, arvores });
      } catch (e) {
        if (vivo) setEstado({ fase: 'erro', mensagem: e instanceof Error ? e.message : String(e) });
      }
    })();
    return () => {
      vivo = false;
    };
  }, [client, sistemaId, tentativa]);

  const recarregar = useCallback(() => setTentativa((t) => t + 1), []);

  return { estado, recarregar };
}
