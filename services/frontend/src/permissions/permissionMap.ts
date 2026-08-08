// Aplicação do mapa de permissões {view, click} por blind_index (RN03). A decisão
// vem pronta do IAM Service; o Frontend apenas a aplica — nunca reimplementa a lógica
// de regras. Padrão fail-closed: sem entrada no mapa ⇒ sem permissão.

import type { Componente, MapaPermissoes } from "../api/types";

/** Verdadeiro se o componente pode ser renderizado. */
export function podeVer(mapa: MapaPermissoes, blindIndex: string): boolean {
  return mapa[blindIndex]?.view ?? false;
}

/** Verdadeiro se o componente pode receber interação (clique/submit). */
export function podeClicar(mapa: MapaPermissoes, blindIndex: string): boolean {
  return mapa[blindIndex]?.click ?? false;
}

/**
 * Poda a árvore, removendo os nós sem permissão de view (e suas subárvores).
 * Devolve `null` se a raiz não for visível.
 */
export function filtrarVisiveis(no: Componente, mapa: MapaPermissoes): Componente | null {
  if (!podeVer(mapa, no.blind_index)) return null;

  const filhos = (no.componente_filhos ?? [])
    .map((f) => filtrarVisiveis(f, mapa))
    .filter((f): f is Componente => f !== null);

  return { ...no, componente_filhos: filhos };
}

/** Extrai todos os blind_index de uma árvore (para pedir o mapa ao IAM). */
export function blindIndexesDe(no: Componente): string[] {
  const acc: string[] = [no.blind_index];
  for (const f of no.componente_filhos ?? []) acc.push(...blindIndexesDe(f));
  return acc;
}
