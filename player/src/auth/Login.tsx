// Tela de login exibida quando não há sessão. Encaminha ao fluxo OAuth do
// Gateway (Google/GitHub); o token volta no fragmento e é capturado em session.ts.

import { urlLogin } from "./session";

export function Login() {
  return (
    <main className="mach-login" style={{ maxWidth: 360, margin: "10vh auto", textAlign: "center", fontFamily: "system-ui, sans-serif" }}>
      <h1>Plataforma MACH</h1>
      <p>Entre para acessar seus sistemas.</p>
      <a href={urlLogin("google")} className="btn" style={{ display: "block", padding: "12px", margin: "8px 0", border: "1px solid #ccc", borderRadius: 8, textDecoration: "none" }}>
        Entrar com Google
      </a>
      <a href={urlLogin("github")} className="btn" style={{ display: "block", padding: "12px", margin: "8px 0", border: "1px solid #ccc", borderRadius: 8, textDecoration: "none" }}>
        Entrar com GitHub
      </a>
    </main>
  );
}
