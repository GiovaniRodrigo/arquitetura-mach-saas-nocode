import { describe, expect, it } from "vitest";
import type { MapaPermissoes, VersaoAtiva } from "../api/types";
import { rotasDe, slug } from "./dynamicRoutes";

const versao: VersaoAtiva = {
  versao_id: "v1",
  numero: 3,
  definicao: {
    campos: {},
    designs: [
      { id: "d1", nome: "Tela de Login", arvore: { blind_index: "bi-1", tipo: "container" } },
      { id: "d2", nome: "Relatórios", arvore: { blind_index: "bi-2", tipo: "container" } },
    ],
  },
};

describe("dynamicRoutes (RN03/RN04)", () => {
  it("slug normaliza acentos e espaços", () => {
    expect(slug("Relatórios de Vendas")).toBe("relatorios-de-vendas");
  });

  it("deriva uma rota por ecrã visível", () => {
    const mapa: MapaPermissoes = {
      "bi-1": { view: true, click: true },
      "bi-2": { view: true, click: false },
    };
    const rotas = rotasDe(versao, mapa);
    expect(rotas.map((r) => r.path)).toEqual(["/tela-de-login", "/relatorios"]);
    expect(rotas[0].screenId).toBe("d1");
  });

  it("omite ecrãs cujo raiz não é visível (RN03)", () => {
    const mapa: MapaPermissoes = { "bi-1": { view: true, click: true } };
    const rotas = rotasDe(versao, mapa);
    expect(rotas).toHaveLength(1);
    expect(rotas[0].screenId).toBe("d1");
  });
});
