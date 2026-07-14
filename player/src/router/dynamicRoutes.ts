// Rotas dinâmicas da SPA derivadas da versão ativa (RN04): cada design (ecrã) do
// sistema vira uma rota. Ecrãs cujo componente raiz não é visível para o
// utilizador (RN03) são omitidos das rotas.

import type { MapaPermissoes, VersaoAtiva } from "../api/types";
import { podeVer } from "../permissions/permissionMap";

export interface RotaDinamica {
  path: string;
  screenId: string;
  nome: string;
  raizBlindIndex: string;
}

/** Converte um nome de ecrã em um segmento de rota estável. */
export function slug(nome: string): string {
  return nome
    .toLowerCase()
    .normalize("NFD")
    .replace(new RegExp("[\\u0300-\\u036f]", "g"), "") // remove diacríticos combinantes
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
}

/**
 * Deriva as rotas da versão ativa, aplicando o mapa de permissões: um ecrã só
 * gera rota se o seu componente raiz for visível.
 */
export function rotasDe(versao: VersaoAtiva, permissoes: MapaPermissoes): RotaDinamica[] {
  const usados = new Set<string>();

  return versao.definicao.designs
    .filter((d) => podeVer(permissoes, d.arvore.blind_index))
    .map((d) => {
      let path = "/" + (slug(d.nome) || d.id);
      // Garante unicidade de path caso dois ecrãs colidam no slug.
      while (usados.has(path)) path += "-" + d.id.slice(0, 4);
      usados.add(path);
      return { path, screenId: d.id, nome: d.nome, raizBlindIndex: d.arvore.blind_index };
    });
}
