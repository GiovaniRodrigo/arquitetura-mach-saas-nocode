import { describe, expect, it } from "vitest";
import type { CampoDef } from "../api/types";
import { MSG, validar } from "./blindIndexValidator";

const schema: Record<string, CampoDef> = {
  "bi-nome": { tipo: "string", obrigatorio: true, limites: { max_length: 10 } },
  "bi-idade": { tipo: "number", obrigatorio: true, limites: { min: 0, max: 120 } },
  "bi-nasc": { tipo: "date", obrigatorio: false },
  "bi-ativo": { tipo: "bool", obrigatorio: false },
};

describe("validar (RN08)", () => {
  it("aceita payload válido", () => {
    expect(
      validar(schema, { "bi-nome": "Ana", "bi-idade": "30", "bi-nasc": "1994-05-01", "bi-ativo": "true" }),
    ).toEqual({});
  });

  it("rejeita chave fora do schema (contorno do cliente)", () => {
    const erros = validar(schema, { "bi-nome": "Ana", "bi-idade": "1", "bi-x": "injetado" });
    expect(erros["bi-x"]).toBe(MSG.desconhecido);
  });

  it("exige campos obrigatórios", () => {
    const erros = validar(schema, { "bi-idade": "1" });
    expect(erros["bi-nome"]).toBe(MSG.obrigatorio);
  });

  it("valida tipos e limites", () => {
    const erros = validar(schema, {
      "bi-nome": "nome-muito-comprido",
      "bi-idade": "999",
      "bi-nasc": "01/01/2000",
      "bi-ativo": "sim",
    });
    expect(erros["bi-nome"]).toBe(MSG.maxLength);
    expect(erros["bi-idade"]).toBe(MSG.max);
    expect(erros["bi-nasc"]).toBe(MSG.data);
    expect(erros["bi-ativo"]).toBe(MSG.booleano);
  });

  it("mensagens são genéricas — chaves são blind_indexes, não nomes reais (RNF08)", () => {
    const erros = validar(schema, { "bi-idade": "-5" });
    for (const [bi, msg] of Object.entries(erros)) {
      expect(bi.startsWith("bi-")).toBe(true);
      expect(msg.length).toBeGreaterThan(0);
    }
  });
});
