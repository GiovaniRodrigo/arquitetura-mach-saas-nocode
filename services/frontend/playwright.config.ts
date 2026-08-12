import { defineConfig, devices } from '@playwright/test';

/**
 * See https://playwright.dev/docs/test-configuration.
 */
export default defineConfig({
  testDir: './e2e',
  /* Maximum time one test can run for. */
  timeout: 30 * 1000,
  expect: {
    /**
     * Maximum time expect() should wait for the condition to be met.
     */
    timeout: 5000
  },
  /* Run tests in files in parallel */
  fullyParallel: true,
  /* Fail the build on CI if you accidentally left test.only in the source code. */
  forbidOnly: !!process.env.CI,
  /* Retry on CI only */
  retries: process.env.CI ? 2 : 0,
  /* Opt out of parallel tests on CI. */
  workers: process.env.CI ? 1 : undefined,
  /* Reporter to use. */
  reporter: 'html',
  /* Shared settings for all the projects below. */
  use: {
    /* Base URL to use in actions like `await page.goto('/')`. A SPA é servida
     * sob /ui/ (vite.config.ts `base`) — goto() deve usar caminhos SEM barra
     * inicial (ex.: 'dashboard', não '/dashboard'), senão o path absoluto
     * ignora o /ui/ do baseURL e cai fora da app (404). */
    baseURL: 'http://localhost:5190/ui/',

    /* Collect trace when retrying the failed test. */
    trace: 'on-first-retry',
  },

  /* Configure projects for major browsers */
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    // Removendo Firefox e Webkit por agilidade no desenvolvimento inicial
  ],

  /* Run your local dev server before starting the tests */
  webServer: {
    // Porta dedicada (5190), diferente da porta fixa de dev (5183, RN de
    // spec 006-auto-cadastro): evita colidir com um `npm run dev` já rodando
    // localmente (reuseExistingServer reaproveitaria essa instância — que não
    // tem VITE_BYPASS_AUTH — em vez de subir a sua própria).
    command: 'npm run dev -- --port 5190 --strictPort',
    url: 'http://localhost:5190',
    reuseExistingServer: !process.env.CI,
    // Sem isso a SPA renderiza a landing pública (Home/Login) em vez do
    // dashboard — main.tsx só monta <App/> com token real ou bypass=true.
    // Os specs de e2e mockam o Gateway via page.route(), então nunca chega
    // a existir um token real neste ambiente.
    env: { VITE_BYPASS_AUTH: 'true' },
  },
});
