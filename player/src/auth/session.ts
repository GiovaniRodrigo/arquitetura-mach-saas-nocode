// Sessão do Player: o JWT MACH chega do callback OAuth do Gateway no fragmento
// da URL (#token=…), é persistido em localStorage e reenviado como Bearer pelo
// ApiClient. Mantém a identidade fora de query/rede intermediária.

const CHAVE = "mach_token";

/** Extrai o token de um fragmento de URL (`#token=…` ou `#a=b&token=…`). */
export function extrairToken(hash: string): string | null {
  const m = hash.match(/[#&]token=([^&]+)/);
  return m ? decodeURIComponent(m[1]) : null;
}

/** Captura o token do callback (se houver), persiste-o e limpa o fragmento da URL. */
export function capturarTokenDaURL(loc: Location = window.location): void {
  const token = extrairToken(loc.hash);
  if (token) {
    localStorage.setItem(CHAVE, token);
    window.history.replaceState(null, "", loc.pathname + loc.search);
  }
}

export function obterToken(): string {
  return localStorage.getItem(CHAVE) ?? "";
}

export function encerrarSessao(): void {
  localStorage.removeItem(CHAVE);
}

/** URL do endpoint de login do Gateway, devolvendo o token a este SPA (/app). */
export function urlLogin(provedor: "google" | "github", origin: string = window.location.origin): string {
  const redirect = encodeURIComponent(origin + "/app");
  return `/auth/${provedor}?redirect_uri=${redirect}`;
}
