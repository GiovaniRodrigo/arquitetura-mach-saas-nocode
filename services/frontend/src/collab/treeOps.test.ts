import { describe, expect, it } from "vitest";
import { aplicarMutacao, alvoDaMutacao } from "./treeOps";
import type { Componente } from "../api/types";

function raiz(filhos: Componente[] = []): Componente {
  return { blind_index: "root", tipo: "tela", componente_filhos: filhos };
}

describe("treeOps (espelho de Collab.Session.Tree.apply/2)", () => {
  it("set_tree substitui a árvore inteira", () => {
    const nova = raiz([{ blind_index: "b1", tipo: "botao" }]);
    const resultado = aplicarMutacao(raiz(), { tipo: "set_tree", arvore: nova });
    expect(resultado).toBe(nova);
  });

  it("add_child anexa o nó como filho do parent (na raiz)", () => {
    const filho: Componente = { blind_index: "b1", tipo: "botao", propriedades: { x: 10, y: 20 } };
    const resultado = aplicarMutacao(raiz(), { tipo: "add_child", parent: "root", no: filho });
    expect(resultado.componente_filhos).toEqual([filho]);
  });

  it("add_child anexa num nó aninhado (não só na raiz)", () => {
    const arvore = raiz([{ blind_index: "container-1", tipo: "container", componente_filhos: [] }]);
    const neto: Componente = { blind_index: "texto-1", tipo: "texto" };
    const resultado = aplicarMutacao(arvore, { tipo: "add_child", parent: "container-1", no: neto });
    expect(resultado.componente_filhos?.[0].componente_filhos).toEqual([neto]);
  });

  it("update_props substitui integralmente as propriedades do nó (não faz merge)", () => {
    const arvore = raiz([{ blind_index: "b1", tipo: "botao", propriedades: { x: 10, y: 20, texto: "Entrar" } }]);
    const resultado = aplicarMutacao(arvore, {
      tipo: "update_props",
      blind_index: "b1",
      propriedades: { x: 30, y: 40 },
    });
    expect(resultado.componente_filhos?.[0].propriedades).toEqual({ x: 30, y: 40 });
  });

  it("remove apaga o nó e sua subárvore", () => {
    const arvore = raiz([
      { blind_index: "b1", tipo: "botao" },
      { blind_index: "b2", tipo: "container", componente_filhos: [{ blind_index: "b3", tipo: "texto" }] },
    ]);
    const resultado = aplicarMutacao(arvore, { tipo: "remove", blind_index: "b1" });
    expect(resultado.componente_filhos?.map((f) => f.blind_index)).toEqual(["b2"]);
  });

  it("remove alcança um nó aninhado em qualquer profundidade", () => {
    const arvore = raiz([
      { blind_index: "b2", tipo: "container", componente_filhos: [{ blind_index: "b3", tipo: "texto" }] },
    ]);
    const resultado = aplicarMutacao(arvore, { tipo: "remove", blind_index: "b3" });
    expect(resultado.componente_filhos?.[0].componente_filhos).toEqual([]);
  });

  it("update_props/add_child em blind_index inexistente devolve a árvore original (no-op)", () => {
    const arvore = raiz([{ blind_index: "b1", tipo: "botao" }]);
    const resultado = aplicarMutacao(arvore, {
      tipo: "update_props",
      blind_index: "inexistente",
      propriedades: {},
    });
    expect(resultado).toEqual(arvore);
  });

  it("add_child com index insere na posição exata (drop BEFORE/AFTER)", () => {
    const arvore = raiz([{ blind_index: "b1", tipo: "x" }, { blind_index: "b2", tipo: "x" }]);
    const resultado = aplicarMutacao(arvore, {
      tipo: "add_child",
      parent: "root",
      no: { blind_index: "meio", tipo: "x" },
      index: 1,
    });
    expect(resultado.componente_filhos?.map((f) => f.blind_index)).toEqual(["b1", "meio", "b2"]);
  });

  it("move reordena um nó dentro do mesmo parent", () => {
    const arvore = raiz([
      { blind_index: "b1", tipo: "x" },
      { blind_index: "b2", tipo: "x" },
      { blind_index: "b3", tipo: "x" },
    ]);
    const resultado = aplicarMutacao(arvore, {
      tipo: "move",
      blind_index: "b1",
      novo_parent: "root",
      index: 2,
    });
    expect(resultado.componente_filhos?.map((f) => f.blind_index)).toEqual(["b2", "b3", "b1"]);
  });

  it("move reparenta um nó para dentro de outro container, preservando sua subárvore", () => {
    const arvore = raiz([
      { blind_index: "container", tipo: "container", componente_filhos: [] },
      { blind_index: "solto", tipo: "x", componente_filhos: [{ blind_index: "filho-do-solto", tipo: "y" }] },
    ]);
    const resultado = aplicarMutacao(arvore, {
      tipo: "move",
      blind_index: "solto",
      novo_parent: "container",
    });
    expect(resultado.componente_filhos?.map((f) => f.blind_index)).toEqual(["container"]);
    const container = resultado.componente_filhos?.[0];
    expect(container?.componente_filhos?.map((f) => f.blind_index)).toEqual(["solto"]);
    expect(container?.componente_filhos?.[0].componente_filhos?.map((f) => f.blind_index)).toEqual([
      "filho-do-solto",
    ]);
  });

  it("move rejeita mover um nó para dentro de si mesmo ou da própria subárvore (no-op)", () => {
    const arvore = raiz([{ blind_index: "b1", tipo: "x" }]);
    expect(aplicarMutacao(arvore, { tipo: "move", blind_index: "b1", novo_parent: "b1" })).toEqual(arvore);

    const comFilho = raiz([{ blind_index: "pai", tipo: "x", componente_filhos: [{ blind_index: "filho", tipo: "y" }] }]);
    expect(
      aplicarMutacao(comFilho, { tipo: "move", blind_index: "pai", novo_parent: "filho" }),
    ).toEqual(comFilho);
  });

  it("alvoDaMutacao devolve o blind_index correto por tipo", () => {
    expect(alvoDaMutacao({ tipo: "update_props", blind_index: "b1", propriedades: {} })).toBe("b1");
    expect(alvoDaMutacao({ tipo: "remove", blind_index: "b2" })).toBe("b2");
    expect(alvoDaMutacao({ tipo: "add_child", parent: "root", no: { blind_index: "x", tipo: "y" } })).toBe("root");
    expect(alvoDaMutacao({ tipo: "move", blind_index: "z", novo_parent: "root" })).toBe("z");
    expect(alvoDaMutacao({ tipo: "set_tree", arvore: raiz() })).toBeNull();
  });
});
