// Espelho puro em TS de Collab.Session.Tree.apply/2 (services/collab/lib/collab/session/tree.ex):
// as mesmas 4 mutações, aplicadas tanto a mutações locais otimistas quanto a
// mutações remotas recebidas do canal (RF06/RF09). Mantém os dois lados do
// canal (Elixir e TS) com a mesma semântica de árvore sem duplicar estado
// autoritativo — o servidor de colaboração continua sendo a fonte de verdade.

import type { Componente } from "../api/types";

export type Mutacao =
  | { tipo: "set_tree"; arvore: Componente }
  | { tipo: "update_props"; blind_index: string; propriedades: Record<string, unknown> }
  | { tipo: "add_child"; parent: string; no: Componente; index?: number }
  | { tipo: "move"; blind_index: string; novo_parent: string; index?: number }
  | { tipo: "remove"; blind_index: string };

/** blind_index alvo de uma mutação — usado para o bloqueio otimista (RN07). */
export function alvoDaMutacao(mutacao: Mutacao): string | null {
  switch (mutacao.tipo) {
    case "update_props":
    case "remove":
    case "move":
      return mutacao.blind_index;
    case "add_child":
      return mutacao.parent;
    default:
      return null;
  }
}

/** Aplica uma mutação à árvore. Não muta `arvore`; devolve uma nova árvore. */
export function aplicarMutacao(arvore: Componente, mutacao: Mutacao): Componente {
  switch (mutacao.tipo) {
    case "set_tree":
      return mutacao.arvore;
    case "update_props":
      return atualizarNo(arvore, mutacao.blind_index, (no) => ({ ...no, propriedades: mutacao.propriedades })) ?? arvore;
    case "add_child":
      return (
        atualizarNo(arvore, mutacao.parent, (no) => ({
          ...no,
          componente_filhos: inserirEm(no.componente_filhos ?? [], mutacao.no, mutacao.index),
        })) ?? arvore
      );
    case "move":
      return aplicarMove(arvore, mutacao);
    case "remove":
      return removerNo(arvore, mutacao.blind_index);
  }
}

function aplicarMove(
  arvore: Componente,
  mutacao: { blind_index: string; novo_parent: string; index?: number },
): Componente {
  const { blind_index: blindIndex, novo_parent: novoParent, index } = mutacao;
  if (blindIndex === novoParent) return arvore;
  if (contemNo(encontrarNo(arvore, blindIndex), novoParent)) return arvore;

  const extraido = extrairNo(arvore, blindIndex);
  if (!extraido) return arvore;
  const [no, arvoreSemNo] = extraido;

  return (
    atualizarNo(arvoreSemNo, novoParent, (destino) => ({
      ...destino,
      componente_filhos: inserirEm(destino.componente_filhos ?? [], no, index),
    })) ?? arvore
  );
}

/** Insere `item` na posição `index` (ausente/fora dos limites = anexa ao final). */
function inserirEm(lista: Componente[], item: Componente, index?: number): Componente[] {
  if (index === undefined || index >= lista.length) return [...lista, item];
  if (index < 0) return [item, ...lista];
  return [...lista.slice(0, index), item, ...lista.slice(index)];
}

/** `no` (ou sua subárvore) contém um descendente (ou a si próprio) com este blind_index? */
function contemNo(no: Componente | null, blindIndex: string): boolean {
  if (!no) return false;
  if (no.blind_index === blindIndex) return true;
  return (no.componente_filhos ?? []).some((f) => contemNo(f, blindIndex));
}

/** Remove o nó `blindIndex` da árvore e devolve `[noExtraido, arvoreSemEle]`, ou `null`. */
function extrairNo(no: Componente, blindIndex: string): [Componente, Componente] | null {
  if (!no.componente_filhos) return null;
  const idx = no.componente_filhos.findIndex((f) => f.blind_index === blindIndex);
  if (idx !== -1) {
    const achado = no.componente_filhos[idx];
    const filhos = [...no.componente_filhos.slice(0, idx), ...no.componente_filhos.slice(idx + 1)];
    return [achado, { ...no, componente_filhos: filhos }];
  }
  for (let i = 0; i < no.componente_filhos.length; i++) {
    const resultado = extrairNo(no.componente_filhos[i], blindIndex);
    if (resultado) {
      const [achado, novoFilho] = resultado;
      const filhos = [...no.componente_filhos];
      filhos[i] = novoFilho;
      return [achado, { ...no, componente_filhos: filhos }];
    }
  }
  return null;
}

function atualizarNo(
  no: Componente,
  blindIndex: string,
  fn: (no: Componente) => Componente,
): Componente | null {
  if (no.blind_index === blindIndex) return fn(no);
  if (!no.componente_filhos) return null;
  for (let i = 0; i < no.componente_filhos.length; i++) {
    const atualizado = atualizarNo(no.componente_filhos[i], blindIndex, fn);
    if (atualizado) {
      const filhos = [...no.componente_filhos];
      filhos[i] = atualizado;
      return { ...no, componente_filhos: filhos };
    }
  }
  return null;
}

function removerNo(no: Componente, blindIndex: string): Componente {
  if (!no.componente_filhos) return no;
  const filhos = no.componente_filhos
    .filter((f) => f.blind_index !== blindIndex)
    .map((f) => removerNo(f, blindIndex));
  return { ...no, componente_filhos: filhos };
}

/** Busca um nó pelo blind_index em qualquer profundidade da árvore. */
export function encontrarNo(arvore: Componente | null, blindIndex: string): Componente | null {
  if (!arvore) return null;
  if (arvore.blind_index === blindIndex) return arvore;
  for (const filho of arvore.componente_filhos ?? []) {
    const achado = encontrarNo(filho, blindIndex);
    if (achado) return achado;
  }
  return null;
}

/** Gera um blind_index local para um nó novo (raiz de tela ou componente). */
export function gerarBlindIndex(): string {
  return `bi-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`;
}

/** Localiza o pai de um nó e seu índice entre os irmãos (para duplicar/mover ao lado). */
export function encontrarPai(arvore: Componente, blindIndex: string): { pai: Componente; index: number } | null {
  const filhos = arvore.componente_filhos ?? [];
  const idx = filhos.findIndex((f) => f.blind_index === blindIndex);
  if (idx !== -1) return { pai: arvore, index: idx };
  for (const filho of filhos) {
    const achado = encontrarPai(filho, blindIndex);
    if (achado) return achado;
  }
  return null;
}

/** Clona um nó e toda sua subárvore com blind_index novos (duplicar componente). */
export function clonarComNovosIndices(no: Componente): Componente {
  return {
    ...no,
    blind_index: gerarBlindIndex(),
    componente_filhos: (no.componente_filhos ?? []).map(clonarComNovosIndices),
  };
}
