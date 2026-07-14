import { describe, expect, it } from "vitest";
import type { Componente, MapaPermissoes } from "../api/types";
import { blindIndexesDe, filtrarVisiveis, podeClicar, podeVer } from "./permissionMap";

const arvore: Componente = {
  blind_index: "bi-root",
  tipo: "container",
  componente_filhos: [
    { blind_index: "bi-visivel", tipo: "texto" },
    {
      blind_index: "bi-oculto",
      tipo: "container",
      componente_filhos: [{ blind_index: "bi-neto", tipo: "texto" }],
    },
  ],
};

describe("permissionMap (RN03)", () => {
  it("fail-closed: sem entrada no mapa ⇒ sem permissão", () => {
    expect(podeVer({}, "bi-x")).toBe(false);
    expect(podeClicar({}, "bi-x")).toBe(false);
  });

  it("blindIndexesDe percorre toda a árvore", () => {
    expect(blindIndexesDe(arvore).sort()).toEqual(
      ["bi-neto", "bi-oculto", "bi-root", "bi-visivel"].sort(),
    );
  });

  it("filtrarVisiveis poda nós sem view (e suas subárvores)", () => {
    const mapa: MapaPermissoes = {
      "bi-root": { view: true, click: false },
      "bi-visivel": { view: true, click: true },
      // bi-oculto sem view → removido junto com bi-neto
    };
    const podada = filtrarVisiveis(arvore, mapa);
    expect(podada).not.toBeNull();
    expect(podada!.componente_filhos).toHaveLength(1);
    expect(podada!.componente_filhos![0].blind_index).toBe("bi-visivel");
  });

  it("raiz invisível ⇒ null", () => {
    expect(filtrarVisiveis(arvore, {})).toBeNull();
  });
});
