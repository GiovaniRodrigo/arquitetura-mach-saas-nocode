// Helper compartilhado dos specs e2e de validação de HTML (html-validator).
//
// Usa o modo `validator: 'WHATWG'` do pacote, que roda 100% offline via
// html-validate (sem bater no serviço público validator.w3.org) — mantém o
// mesmo padrão hermético dos demais e2e deste projeto (ver mockApi.ts: os
// specs não podem depender de rede/serviços externos).
import { expect, type Page } from '@playwright/test';
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

// aria-labelledby dangling: o Tooltip do @base-ui/react (components/ui/
// tooltip.tsx, usado por TooltipProvider em DashboardLayout) referencia via
// aria-labelledby o id do Popup — que só existe no DOM quando o tooltip está
// aberto (Portal condicional). É o comportamento padrão da biblioteca (mesmo
// padrão do Radix UI), não um bug do código deste projeto.
const IGNORAR_PADRAO = ['no-missing-references'];

/** Valida um documento HTML completo (ex.: `page.content()`). */
export async function validarHtml(html: string, rotulo: string): Promise<void> {
  // @types/html-validator só modela o formato JSON do W3C (`{ messages }`);
  // no modo WHATWG a lib retorna `{ isValid, errorCount, warnings, errors }`
  // (ver node_modules/html-validator/lib/whatwg-validator.js) — daí o cast
  // via `unknown` em vez de um tipo diretamente compatível com a declaração
  // do pacote.
  const resultado = (await htmlValidator({
    validator: 'WHATWG',
    data: html,
    ignore: IGNORAR_PADRAO,
  })) as unknown as { errors: ErroValidacao[] };
  expect(resultado.errors, `HTML inválido em "${rotulo}":\n${formatarErros(resultado.errors)}`).toEqual([]);
}

/** Valida o documento renderizado atualmente em `page`. */
export async function validarPaginaHtml(page: Page, rotulo: string): Promise<void> {
  await validarHtml(await page.content(), rotulo);
}
