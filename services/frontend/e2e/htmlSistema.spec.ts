import { test, expect } from '@playwright/test';
import { mockGatewayPadrao, json } from './mockApi';
import { validarPaginaHtml } from './htmlValidator';

// Validação de HTML — segmento 1: telas do PRÓPRIO sistema (a casca fixa do
// dashboard: Clientes, Configuração, Monitor, Perfil, Ajuda, editor de
// sistemas), navegação real via browser com o Gateway mockado (mesmo padrão
// de dashboard.spec.ts/telas.spec.ts). O segmento 2 (telas construídas pelos
// clientes no builder nocode) é coberto à parte, como teste de componente em
// PreviewRenderer.htmlValidator.test.tsx — abrir uma tela específica no
// canvas depende do canal de colaboração (services/collab/WebSocket Phoenix),
// fora do escopo hermético dos e2e (ver nota em telas.spec.ts).

const TENANT = { id: 't1', nome: 'Acme' };
const SISTEMA = { id: 's1', nome: 'CRM' };

test.describe('Validação de HTML (html-validator/WHATWG) — telas do sistema', () => {
  test.beforeEach(async ({ page }) => {
    await mockGatewayPadrao(page);
    await page.route('**/api/v1/tenants', (route) => json(route, { tenants: [TENANT] }));
    await page.route(`**/api/v1/tenants/${TENANT.id}`, (route) => json(route, TENANT));
    await page.route(/\/api\/v1\/sistemas\?tenant_id=/, (route) => json(route, { sistemas: [SISTEMA] }));
    await page.route(/\/api\/v1\/designs\?sistema_id=/, (route) =>
      json(route, { telas: [{ id: 'd1', nome: 'Home' }] }),
    );
    // Monitor.tsx acessa `recursos.servicos.map(...)` sem guarda — o catch-all
    // de mockGatewayPadrao (200 `{}`) não basta aqui, quebra o render.
    await page.route('**/api/v1/monitor/recursos', (route) => json(route, { servicos: [] }));
  });

  const rotas: Array<{ rotulo: string; path: string; aguardar: (page: import('@playwright/test').Page) => Promise<unknown> }> = [
    { rotulo: 'Dashboard', path: 'dashboard', aguardar: (p) => expect(p.getByRole('heading', { name: 'Dashboard' })).toBeVisible() },
    {
      rotulo: 'Clientes',
      path: 'dashboard/clientes',
      aguardar: (p) => expect(p.getByRole('heading', { name: 'Clientes', level: 2 })).toBeVisible(),
    },
    {
      rotulo: 'Sistemas do cliente',
      path: `dashboard/clientes/${TENANT.id}`,
      aguardar: (p) => expect(p.getByRole('link', { name: /abrir sistema/i })).toBeVisible(),
    },
    {
      rotulo: 'Sistema — aba Telas',
      path: `dashboard/clientes/${TENANT.id}/sistemas/${SISTEMA.id}/telas`,
      aguardar: (p) => expect(p.getByRole('link', { name: 'Telas' })).toHaveAttribute('aria-current', 'page'),
    },
    {
      rotulo: 'Sistema — aba Regras de Negócio',
      path: `dashboard/clientes/${TENANT.id}/sistemas/${SISTEMA.id}/regras`,
      aguardar: (p) => expect(p.getByRole('link', { name: 'Regras de Negócio' })).toHaveAttribute('aria-current', 'page'),
    },
    {
      rotulo: 'Sistema — aba Versão',
      path: `dashboard/clientes/${TENANT.id}/sistemas/${SISTEMA.id}/versao`,
      aguardar: (p) => expect(p.getByRole('link', { name: 'Versão' })).toHaveAttribute('aria-current', 'page'),
    },
    {
      rotulo: 'Configuração',
      path: 'dashboard/configuracao',
      aguardar: (p) => expect(p.getByRole('heading', { name: 'Configurações', level: 2 })).toBeVisible(),
    },
    {
      rotulo: 'Monitor de Recursos',
      path: 'dashboard/monitor',
      aguardar: (p) => expect(p.getByRole('heading', { name: 'Monitor de Recursos', level: 2 })).toBeVisible(),
    },
    {
      rotulo: 'Cadastro/Perfil',
      path: 'dashboard/perfil',
      aguardar: (p) => expect(p.getByRole('heading', { name: 'Cadastro/Perfil', level: 2 })).toBeVisible(),
    },
    {
      rotulo: 'Ajuda',
      path: 'dashboard/ajuda',
      aguardar: (p) => expect(p.getByRole('heading', { name: 'Ajuda', level: 2 })).toBeVisible(),
    },
  ];

  for (const rota of rotas) {
    test(`HTML válido — ${rota.rotulo}`, async ({ page }) => {
      await page.goto(rota.path);
      await rota.aguardar(page);
      await validarPaginaHtml(page, rota.rotulo);
    });
  }
});
