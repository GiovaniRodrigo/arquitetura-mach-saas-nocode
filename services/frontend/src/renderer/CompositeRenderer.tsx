// CompositeRenderer: renderiza recursivamente a árvore de UI da versão ativa
// (RF01, RN04), respeitando o mapa de permissões (RN03). Nós sem `view` não são
// renderizados; nós sem `click` são renderizados desabilitados.

import type { JSX } from "react";
import type { Componente, MapaPermissoes } from "../api/types";
import { podeClicar, podeVer } from "../permissions/permissionMap";

export interface CompositeRendererProps {
  no: Componente;
  permissoes: MapaPermissoes;
  /** Notifica mudança de valor de um input (blind_index → valor). */
  onInput?: (blindIndex: string, valor: string) => void;
  /** Notifica clique/submit de um componente interativo. */
  onClick?: (blindIndex: string) => void;
}

export function CompositeRenderer({
  no,
  permissoes,
  onInput,
  onClick,
}: CompositeRendererProps): JSX.Element | null {
  // fail-closed: sem view, o nó (e subárvore) não é renderizado (RN03).
  if (!podeVer(permissoes, no.blind_index)) return null;

  const clicavel = podeClicar(permissoes, no.blind_index);
  const props = (no.propriedades ?? {}) as Record<string, unknown>;
  const rotulo = typeof props.label === "string" ? props.label : undefined;

  const filhos = (no.componente_filhos ?? []).map((f) => (
    <CompositeRenderer
      key={f.blind_index}
      no={f}
      permissoes={permissoes}
      onInput={onInput}
      onClick={onClick}
    />
  ));

  switch (no.tipo) {
    case "input_texto":
      return (
        <label data-bi={no.blind_index}>
          {rotulo}
          <input
            type="text"
            disabled={!clicavel}
            placeholder={typeof props.placeholder === "string" ? props.placeholder : undefined}
            onChange={(e) => onInput?.(no.blind_index, e.target.value)}
          />
        </label>
      );

    case "botao":
      return (
        <button
          data-bi={no.blind_index}
          disabled={!clicavel}
          onClick={() => clicavel && onClick?.(no.blind_index)}
        >
          {rotulo ?? "OK"}
        </button>
      );

    case "texto":
      return <span data-bi={no.blind_index}>{rotulo}</span>;

    case "container":
    default:
      return (
        <div data-bi={no.blind_index} data-tipo={no.tipo}>
          {filhos}
        </div>
      );
  }
}
