// Helper compartilhado dos specs e2e: mocka o Gateway (`/api/v1/**`) via
// page.route(), para não depender de um backend real rodando localmente.
//
// Crítico: SEM isso, qualquer chamada não interceptada por um teste cai no
// backend real (se houver algum rodando na mesma máquina) — e com
// VITE_BYPASS_AUTH=true (sem token real), o Gateway responde 401, o que
// dispara `encerrarSessao()+window.location.reload()` em api/client.ts
// (`aoNaoAutorizado`) e entra num loop de reload que nunca estabiliza (a
// causa raiz do "element was detached from the DOM, retrying" visto ao
// depurar este arquivo). Por isso `mockGatewayPadrao` cobre também as
// chamadas "de fundo" do dashboard (métricas/últimos acessos/feedback/
// resumo financeiro) mesmo em specs que não testam essas telas.
import type { Page, Route } from '@playwright/test';

function json(route: Route, body: unknown, status = 200) {
  return route.fulfill({ status, contentType: 'application/json', body: JSON.stringify(body) });
}

/**
 * Registra os mocks "de fundo" com respostas vazias neutras — inclui um
 * catch-all de `**\/api/v1/**` (200 `{}`) para QUALQUER endpoint não listado
 * abaixo, porque nunca dá pra prever de antemão toda chamada de fundo que
 * uma página do dashboard dispara (menus, cards, etc.), e um único endpoint
 * esquecido já é suficiente para cair no backend real e disparar o loop de
 * reload. Chame antes de `page.goto()`; specs que precisam de dados
 * específicos podem registrar `page.route()` adicionais DEPOIS desta
 * chamada — o último registrado vence quando os padrões se sobrepõem.
 */
export async function mockGatewayPadrao(page: Page): Promise<void> {
  await page.route('**/api/v1/**', (route) => json(route, {}));
  await page.route('**/api/v1/tenants', (route) => json(route, { tenants: [] }));
  await page.route('**/api/v1/sistemas**', (route) => json(route, { sistemas: [] }));
  await page.route('**/api/v1/designs**', (route) => json(route, { telas: [] }));
  await page.route('**/api/v1/dashboard/ultimos-acessos', (route) => json(route, { eventos: [] }));
  await page.route('**/api/v1/dashboard/feedback**', (route) => json(route, { itens: [] }));
  await page.route('**/api/v1/dashboard/resumo-financeiro', (route) =>
    json(route, { receita_total_centavos: 0, moeda: 'BRL', competencia: '' }),
  );
}

export { json };
