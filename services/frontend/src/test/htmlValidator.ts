// Helper de validação de HTML (html-validator) para testes de componente
// (vitest). Mesmo modo `validator: 'WHATWG'` usado nos e2e (100% offline via
// html-validate, sem bater no serviço público validator.w3.org) — ver
// e2e/htmlValidator.ts para o equivalente em Playwright.
import { expect } from 'vitest';
import htmlValidator from 'html-validator';

interface ErroValidacao {
  ruleId: string;
  message: string;
  line?: number;
  column?: number;
}

function formatarErros(erros: ErroValidacao[]): string {
  return erros.map((e) => `  [${e.ruleId}] linha ${e.line}:${e.column} — ${e.message}`).join('\n');
}

/**
 * Valida um fragmento HTML (ex.: `container.innerHTML` do React Testing
 * Library) — `isFragment: true` porque não é um documento completo (sem
 * <html>/<head>/doctype).
 *
 * `ignorar` (ruleId do html-validate) serve para regras que fazem sentido
 * para um DOCUMENTO completo mas não para um fragmento isolado renderizado
 * fora de contexto — ex.: `heading-level` (ordem de <h1>-<h6>) não se aplica
 * a uma árvore composable onde o componente "heading" não carrega nível
 * semântico algum no modelo de dados.
 */
export async function validarFragmentoHtml(html: string, rotulo: string, ignorar: string[] = []): Promise<void> {
  // @types/html-validator só modela o formato JSON do W3C (`{ messages }`);
  // no modo WHATWG a lib retorna `{ isValid, errorCount, warnings, errors }`
  // (ver node_modules/html-validator/lib/whatwg-validator.js) — daí o cast
  // via `unknown` em vez de um tipo diretamente compatível com a declaração
  // do pacote.
  const resultado = (await htmlValidator({
    validator: 'WHATWG',
    data: html,
    isFragment: true,
    ...(ignorar.length > 0 ? { ignore: ignorar } : {}),
  })) as unknown as { errors: ErroValidacao[] };
  expect(resultado.errors, `HTML inválido em "${rotulo}":\n${formatarErros(resultado.errors)}`).toEqual([]);
}
