import { test, expect } from '@playwright/test';
import { mockGatewayPadrao, json } from './mockApi';

// E2E da rota /dashboard/clientes/:tenantId/sistemas/:sistemaId/telas (RF09).
//
// Escopo: navegação real (multi-página, via browser) e o fluxo de listar/
// criar telas, mockando o Gateway (`/api/v1/**`) via page.route() — sem
// backend real. Fica de fora, de propósito, tudo que depende do canal de
// colaboração (services/collab, WebSocket Phoenix em /socket): abrir uma
// tela específica no canvas, drag&drop, undo/redo etc. Esses fluxos já têm
// cobertura de integração em AbaTelas.test.tsx (collab mockado via
// vi.mock) — reproduzir o protocolo Phoenix (join/ref/heartbeat) num
// WebSocket real só para o e2e não paga o custo/fragilidade hoje; ver
// ressalva no relatório de cobertura.

const TENANT = { id: 't1', nome: 'Acme' };
const SISTEMA = { id: 's1', nome: 'CRM' };

test.describe('Rota /dashboard/clientes/:tenantId/sistemas/:sistemaId/telas (RF09)', () => {
  test('navega de Clientes até a aba Telas e cria uma nova tela pelo modal', async ({ page }) => {
    await mockGatewayPadrao(page);
    await page.route('**/api/v1/tenants', (route) => json(route, { tenants: [TENANT] }));
    await page.route(`**/api/v1/tenants/${TENANT.id}`, (route) => json(route, TENANT));
    await page.route(/\/api\/v1\/sistemas\?tenant_id=/, (route) => json(route, { sistemas: [SISTEMA] }));

    let telas: Array<{ id: string; nome: string }> = [];
    await page.route(/\/api\/v1\/designs\?sistema_id=/, (route) => json(route, { telas }));
    await page.route('**/api/v1/designs', (route) => {
      if (route.request().method() !== 'POST') return route.fallback();
      const novo = { id: 'd-novo', nome: 'Home' };
      telas = [...telas, novo];
      return json(route, { design_id: novo.id });
    });

    // Jornada real: Clientes -> Abrir cliente -> Abrir sistema -> aba Telas
    // (index redireciona para /telas, ver App.tsx).
    await page.goto('dashboard/clientes');
    await page.getByRole('link', { name: /abrir cliente/i }).click();
    await expect(page).toHaveURL(new RegExp(`/dashboard/clientes/${TENANT.id}$`));

    await page.getByRole('link', { name: /abrir sistema/i }).click();
    await expect(page).toHaveURL(new RegExp(`/dashboard/clientes/${TENANT.id}/sistemas/${SISTEMA.id}/telas$`));

    // Casca do sistema (SistemaAbas): 3 abas de navegação, Telas ativa.
    await expect(page.getByRole('link', { name: 'Telas' })).toHaveAttribute('aria-current', 'page');
    await expect(page.getByText(/nenhuma tela criada ainda/i)).toBeVisible();

    await page.getByRole('button', { name: /selecione uma tela/i }).click();
    await page.getByRole('button', { name: /nova tela/i }).click();

    const modal = page.getByRole('dialog', { name: /nova tela/i });
    await expect(modal).toBeVisible();
    await modal.getByLabel(/nome da tela/i).fill('Home');
    await modal.getByRole('button', { name: /criar tela/i }).click();

    await expect(modal).not.toBeVisible();
    // A tela recém-criada fica selecionada automaticamente (EditorTopbar
    // passa a mostrar seu nome em vez de "Selecione uma tela").
    await expect(page.getByRole('button', { name: /^Home$/ })).toBeVisible();
  });

  test('acessar a URL da aba Telas diretamente (deep link/refresh) renderiza a casca do editor', async ({
    page,
  }) => {
    await mockGatewayPadrao(page);
    await page.route(`**/api/v1/tenants/${TENANT.id}`, (route) => json(route, TENANT));
    await page.route(/\/api\/v1\/sistemas\?tenant_id=/, (route) => json(route, { sistemas: [SISTEMA] }));
    await page.route(/\/api\/v1\/designs\?sistema_id=/, (route) =>
      json(route, { telas: [{ id: 'd1', nome: 'Home' }] }),
    );

    // Carrega a rota aninhada direto pela URL, sem passar pelas telas
    // anteriores — é o caso de recarregar a página ou abrir um link salvo.
    await page.goto(`dashboard/clientes/${TENANT.id}/sistemas/${SISTEMA.id}/telas`);

    await expect(page.getByRole('heading', { name: SISTEMA.nome })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Telas' })).toHaveAttribute('aria-current', 'page');
    await expect(page.getByRole('link', { name: 'Regras de Negócio' })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Versão' })).toBeVisible();
    // A tela existente aparece como opção no dropdown do topbar (ainda não
    // selecionada — abrir o canvas depende do canal de colaboração, fora do
    // escopo deste spec).
    await expect(page.getByRole('button', { name: /selecione uma tela/i })).toBeVisible();
  });
});
