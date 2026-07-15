// Leitura dos claims do JWT MACH apenas para EXIBIÇÃO (nome/e-mail/iniciais no
// cabeçalho — RF03). Não valida a assinatura: a autorização permanece no Gateway
// (RNF06). Nunca confie nestes valores para decisões de acesso.

export interface UsuarioAutenticado {
  nome?: string;
  email?: string;
  /** Iniciais para o avatar; fallback "?". */
  iniciais: string;
}

interface ClaimsJwt {
  name?: string;
  email?: string;
  preferred_username?: string;
  sub?: string;
}

/** Decodifica o payload (2ª parte) de um JWT em base64url. Retorna null se inválido. */
export function lerClaims(token: string): ClaimsJwt | null {
  const partes = token.split('.');
  if (partes.length < 2 || !partes[1]) return null;
  try {
    const base64 = partes[1].replace(/-/g, '+').replace(/_/g, '/');
    const json = decodeURIComponent(
      atob(base64)
        .split('')
        .map((c) => '%' + c.charCodeAt(0).toString(16).padStart(2, '0'))
        .join(''),
    );
    return JSON.parse(json) as ClaimsJwt;
  } catch {
    return null;
  }
}

/** Deriva as iniciais de um nome ("Ana Silva" → "AS") ou de um e-mail ("a@x" → "A"). */
export function iniciaisDe(nome?: string, email?: string): string {
  const base = (nome ?? email ?? '').trim();
  if (!base) return '?';
  const palavras = base.split(/\s+/).filter(Boolean);
  if (palavras.length >= 2) {
    return (palavras[0][0] + palavras[palavras.length - 1][0]).toUpperCase();
  }
  return base[0].toUpperCase();
}

/** Monta o objeto de exibição a partir do token persistido. */
export function usuarioDe(token: string): UsuarioAutenticado {
  const claims = token ? lerClaims(token) : null;
  const nome = claims?.name ?? claims?.preferred_username;
  const email = claims?.email;
  return { nome, email, iniciais: iniciaisDe(nome, email) };
}
