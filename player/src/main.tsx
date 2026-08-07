// Ponto de entrada da SPA. A configuração (URL do Gateway, JWT, sistema) é injetada
// pelo host em `window.__PLAYER_CONFIG__` no embed do Player.

import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import "./index.css";
import { App, type PlayerConfig } from "./App";
import { Login } from "./auth/Login";
import { Home } from "./pages/Home/Home";
import { capturarTokenDaURL, obterToken } from "./auth/session";
import { ThemeProvider } from "./theme/ThemeProvider";
import { initTheme } from "./theme/initTheme";

// Aplica o tema salvo antes do primeiro paint, evitando flash (RNF04).
initTheme();

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

const isBypass = import.meta.env.VITE_BYPASS_AUTH === "true";

if (!token && !isBypass) {
  // Sem sessão: landing pública (Home, RF01) por padrão; /login oferece o
  // login social (Google/GitHub) via Gateway (RF02).
  root.render(
    <StrictMode>
      <ThemeProvider>
        <BrowserRouter basename={import.meta.env.BASE_URL}>
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route path="*" element={<Home />} />
          </Routes>
        </BrowserRouter>
      </ThemeProvider>
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
      <ThemeProvider>
        <App config={config} />
      </ThemeProvider>
    </StrictMode>,
  );
}
