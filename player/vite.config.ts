import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// Vite + Vitest do Headless Player. Testes correm em jsdom (renderização React) e
// usam globals estilo Jest (describe/it/expect).
export default defineConfig({
  // Servido sob /app no domínio institucional (gfcode.com.br/app) atrás do Nginx.
  base: "/app/",
  plugins: [react()],
  test: {
    globals: true,
    environment: "jsdom",
  },
});
