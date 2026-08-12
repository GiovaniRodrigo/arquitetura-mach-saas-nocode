import path from "path";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// Vite + Vitest do Frontend. Testes correm em jsdom (renderização React) e
// usam globals estilo Jest (describe/it/expect).
export default defineConfig({
  // Servido sob /ui no domínio institucional (gfcode.com.br/ui) atrás do Nginx.
  base: "/ui/",
  plugins: [react()],
  server: {
    // Porta fixa (RN da spec 006-auto-cadastro/quickstart.md) — sem isso o Vite
    // cai no default 5173, que diverge do que os demais scripts/docs assumem.
    port: 5183,
    strictPort: true,
    // Em produção o Nginx faz esse proxy (baseUrl vazio no client, ver main.tsx);
    // localmente o próprio Vite assume o papel para `npm run dev` funcionar sem
    // precisar de VITE_GATEWAY_URL.
    proxy: {
      "/api": "http://localhost:8080",
      "/socket": { target: "http://localhost:4000", ws: true },
    },
  },
  test: {
    globals: true,
    environment: "jsdom",
    include: ["src/**/*.test.{ts,tsx}"],
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
});
