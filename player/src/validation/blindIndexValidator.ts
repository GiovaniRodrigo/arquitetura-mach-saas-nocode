// Validação local por mapa de blind_index (RN08 — validação dupla): o Player
// pré-valida no cliente para dar feedback imediato, mas o Logic Engine revalida no
// servidor (a autoridade). As regras espelham as do servidor; as mensagens são
// genéricas e indexadas por blind_index — nunca nomes reais de campos (RNF08).

import type { CampoDef } from "../api/types";

export const MSG = {
  desconhecido: "campo não pertence ao schema",
  obrigatorio: "campo obrigatório ausente",
  numero: "valor não é um número válido",
  data: "valor não é uma data válida (AAAA-MM-DD)",
  booleano: "valor não é booleano (true/false)",
  min: "valor abaixo do mínimo permitido",
  max: "valor acima do máximo permitido",
  minLength: "texto mais curto que o mínimo permitido",
  maxLength: "texto mais longo que o máximo permitido",
} as const;

const RE_DATA = /^\d{4}-\d{2}-\d{2}$/;

/**
 * Valida `dados` contra `schema`, devolvendo erros indexados por blind_index.
 * Mapa vazio ⇒ válido. Rejeita chaves fora do schema (contorno do cliente).
 */
export function validar(
  schema: Record<string, CampoDef>,
  dados: Record<string, string>,
): Record<string, string> {
  const erros: Record<string, string> = {};

  for (const bi of Object.keys(dados)) {
    if (!(bi in schema)) erros[bi] = MSG.desconhecido;
  }

  for (const [bi, campo] of Object.entries(schema)) {
    const valor = dados[bi];
    if (valor === undefined || valor === "") {
      if (campo.obrigatorio) erros[bi] = MSG.obrigatorio;
      continue;
    }
    const msg = validarValor(campo, valor);
    if (msg) erros[bi] = msg;
  }

  return erros;
}

function validarValor(campo: CampoDef, valor: string): string | null {
  const lim = campo.limites ?? {};
  switch (campo.tipo) {
    case "number": {
      const n = Number(valor);
      if (valor.trim() === "" || Number.isNaN(n)) return MSG.numero;
      if (lim.min !== undefined && n < lim.min) return MSG.min;
      if (lim.max !== undefined && n > lim.max) return MSG.max;
      return null;
    }
    case "date":
      return RE_DATA.test(valor) && !Number.isNaN(Date.parse(valor)) ? null : MSG.data;
    case "bool":
      return valor === "true" || valor === "false" ? null : MSG.booleano;
    default: {
      if (lim.min_length !== undefined && valor.length < lim.min_length) return MSG.minLength;
      if (lim.max_length !== undefined && valor.length > lim.max_length) return MSG.maxLength;
      return null;
    }
  }
}
