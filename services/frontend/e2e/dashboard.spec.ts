import { test, expect } from '@playwright/test';
import { mockGatewayPadrao } from './mockApi';

test.describe('Dashboard Layout E2E', () => {
  test('deve renderizar a Sidebar e o Header sem sobreposição', async ({ page }) => {
    await mockGatewayPadrao(page);

    // Acessa o dashboard (o bypass auth local já deve permitir a navegação direta).
    // Sem barra inicial: a app é servida sob /ui/ (ver baseURL em playwright.config.ts).
    await page.goto('dashboard');

    // Verifica se a Sidebar e os itens de menu estão visíveis (não existe
    // item "Home" na Sidebar atual — os links são Dashboard/Clientes/
    // Configuração/Cadastro-Perfil/Ajuda, ver DashboardLayout.tsx).
    const clientesLink = page.getByRole('link', { name: 'Clientes' });
    await expect(clientesLink).toBeVisible();

    // Verifica se o Header (título da página) está visível.
    const headerTitle = page.getByRole('heading', { name: 'Dashboard' });
    await expect(headerTitle).toBeVisible();

    // Uma asserção crucial: o link deve estar no topo e receber cliques —
    // se o header estivesse cortando/sobrepondo, o click seria interceptado
    // por ele. Usa "Clientes" (não "Dashboard", já ativo) para exercitar uma
    // navegação real em vez de um click que não move a rota.
    await clientesLink.click();
    await expect(page).toHaveURL(/\/dashboard\/clientes$/);

    // Testa a funcionalidade responsiva/drawer abrindo e fechando
    const toggleButton = page.getByRole('button', { name: /toggle sidebar/i });
    
    // O Shadcn UI Sidebar pode alterar a largura ou atributos de visibilidade no colapso
    await expect(toggleButton).toBeVisible();
    await toggleButton.click();
    
    // Após clicar, a barra muda de estado. Em e2e, verificamos a existência física ou classes.
    // Como a Sidebar usa 'data-state="collapsed"', podemos verificar isso na raiz da sidebar se necessário.
  });
});
