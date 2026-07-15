import path from "path";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// Vite + Vitest do Headless Player. Testes correm em jsdom (renderização React) e
// usam globals estilo Jest (describe/it/expect).
export default defineConfig({
  // Servido sob /ui no domínio institucional (gfcode.com.br/ui) atrás do Nginx.
  base: "/ui/",
  plugins: [react()],
  test: {
    globals: true,
    environment: "jsdom",
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
});
