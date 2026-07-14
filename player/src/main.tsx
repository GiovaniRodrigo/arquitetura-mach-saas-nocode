// Ponto de entrada da SPA. A configuração (URL do Gateway, JWT, sistema) é injetada
// pelo host em `window.__PLAYER_CONFIG__` no embed do Player.

import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { App, type PlayerConfig } from "./App";

declare global {
  interface Window {
    __PLAYER_CONFIG__?: PlayerConfig;
  }
}

const config: PlayerConfig = window.__PLAYER_CONFIG__ ?? {
  baseUrl: import.meta.env.VITE_GATEWAY_URL ?? "http://localhost:8080",
  token: "",
  sistemaId: "",
};

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App config={config} />
  </StrictMode>,
);
