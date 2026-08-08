import { test, expect } from '@playwright/test';

test.describe('Dashboard Layout E2E', () => {
  test('deve renderizar a Sidebar e o Header sem sobreposição', async ({ page }) => {
    // Acessa o dashboard (o bypass auth local já deve permitir a navegação direta)
    await page.goto('/dashboard');

    // Verifica se a Sidebar e os itens de menu estão visíveis
    const homeLink = page.getByRole('link', { name: /home/i });
    await expect(homeLink).toBeVisible();

    // Verifica se o Header e o Toggle estão visíveis
    const headerTitle = page.getByText(/welcome, user/i);
    await expect(headerTitle).toBeVisible();

    // Uma asserção crucial: o link deve estar no topo e receber cliques
    // Se o header estivesse cortando/sobrepondo, o click seria interceptado por ele
    await homeLink.click();

    // Testa a funcionalidade responsiva/drawer abrindo e fechando
    const toggleButton = page.getByRole('button', { name: /toggle sidebar/i });
    
    // O Shadcn UI Sidebar pode alterar a largura ou atributos de visibilidade no colapso
    await expect(toggleButton).toBeVisible();
    await toggleButton.click();
    
    // Após clicar, a barra muda de estado. Em e2e, verificamos a existência física ou classes.
    // Como a Sidebar usa 'data-state="collapsed"', podemos verificar isso na raiz da sidebar se necessário.
  });
});
