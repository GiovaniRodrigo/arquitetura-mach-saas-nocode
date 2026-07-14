// Ponto de entrada da SPA. A configuração (URL do Gateway, JWT, sistema) é injetada
// pelo host em `window.__PLAYER_CONFIG__` no embed do Player.

import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { App, type PlayerConfig } from "./App";
import { Login } from "./auth/Login";
import { capturarTokenDaURL, obterToken } from "./auth/session";

declare global {
  interface Window {
    __PLAYER_CONFIG__?: PlayerConfig;
  }
}

// Captura o token do callback OAuth (#token=…) antes de tudo.
capturarTokenDaURL();

const embed = window.__PLAYER_CONFIG__;
const token = embed?.token || obterToken();
const root = createRoot(document.getElementById("root")!);

if (!token) {
  // Sem sessão: oferece login social (Google/GitHub) via Gateway.
  root.render(
    <StrictMode>
      <Login />
    </StrictMode>,
  );
} else {
  const params = new URLSearchParams(window.location.search);
  const config: PlayerConfig = embed
    ? { ...embed, token }
    : {
        // Servido sob /app no mesmo domínio: baseUrl vazio → chamadas relativas a /api (proxy Nginx).
        baseUrl: import.meta.env.VITE_GATEWAY_URL ?? "",
        token,
        sistemaId: params.get("sistema") ?? "",
      };
  root.render(
    <StrictMode>
      <App config={config} />
    </StrictMode>,
  );
}
