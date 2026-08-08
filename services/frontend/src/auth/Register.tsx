// Tela de auto cadastro (spec 006, RF02/RF04): cria um tenant próprio e uma
// conta autenticada por e-mail/senha, autenticando o usuário automaticamente
// em sucesso — não depende de sessão nem de AppProvider (roda antes do login,
// como Login.tsx e Home.tsx).

import { useState, type FormEvent } from "react";
import { Link } from "react-router-dom";
import { salvarToken } from "./session";

export type FetchFn = typeof fetch;

interface RegisterProps {
  fetchFn?: FetchFn;
  /** Injetável para testes; em produção navega o browser (recarrega o app já autenticado). */
  redirecionarPara?: (url: string) => void;
}

export function Register({
  fetchFn = fetch,
  redirecionarPara = (url) => {
    window.location.href = url;
  },
}: RegisterProps) {
  const [nome, setNome] = useState("");
  const [email, setEmail] = useState("");
  const [senha, setSenha] = useState("");
  const [nomeTenant, setNomeTenant] = useState("");
  const [enviando, setEnviando] = useState(false);
  const [erro, setErro] = useState<string | null>(null);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setErro(null);
    setEnviando(true);
    try {
      const resp = await fetchFn("/api/v1/auth/registro", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ nome, email, senha, nome_tenant: nomeTenant }),
      });
      if (!resp.ok) {
        const corpo = await resp.json().catch(() => null);
        setErro(corpo?.mensagem ?? "Não foi possível concluir o cadastro.");
        return;
      }
      const { jwt } = await resp.json();
      salvarToken(jwt);
      // Mesmo destino do callback OAuth (session.ts/urlLogin): a raiz do SPA,
      // de onde a rota "*" autenticada redireciona para /dashboard.
      redirecionarPara(import.meta.env.BASE_URL);
    } finally {
      setEnviando(false);
    }
  }

  return (
    <main className="min-h-screen w-full flex flex-col md:flex-row bg-background text-foreground font-sans">
      <div className="md:w-1/2 bg-zinc-900 flex flex-col justify-center items-center p-12 text-zinc-50 border-r border-zinc-800">
        <div className="max-w-md">
          <h1 className="font-heading text-4xl md:text-5xl font-bold tracking-tight mb-4 text-white">
            Plataforma MACH
          </h1>
          <p className="text-zinc-400 text-lg md:text-xl">
            Crie sua conta e comece a construir seu sistema agora.
          </p>
        </div>
      </div>
      <div className="md:w-1/2 flex flex-col justify-center items-center p-8 md:p-12">
        <div className="w-full max-w-sm space-y-6">
          <div className="text-center">
            <h2 className="font-heading text-3xl font-bold tracking-tight">Criar conta</h2>
            <p className="text-muted-foreground mt-2">Teste grátis, sem cartão de crédito.</p>
          </div>

          <form className="space-y-4" onSubmit={onSubmit}>
            <div>
              <label htmlFor="register-nome" className="block text-sm font-medium mb-1">
                Nome
              </label>
              <input
                id="register-nome"
                type="text"
                required
                value={nome}
                onChange={(e) => setNome(e.target.value)}
                className="w-full px-4 py-2 rounded-lg border border-input bg-background"
              />
            </div>
            <div>
              <label htmlFor="register-email" className="block text-sm font-medium mb-1">
                E-mail
              </label>
              <input
                id="register-email"
                type="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full px-4 py-2 rounded-lg border border-input bg-background"
              />
            </div>
            <div>
              <label htmlFor="register-senha" className="block text-sm font-medium mb-1">
                Senha
              </label>
              <input
                id="register-senha"
                type="password"
                required
                minLength={8}
                value={senha}
                onChange={(e) => setSenha(e.target.value)}
                className="w-full px-4 py-2 rounded-lg border border-input bg-background"
              />
            </div>
            <div>
              <label htmlFor="register-nome-tenant" className="block text-sm font-medium mb-1">
                Nome do negócio
              </label>
              <input
                id="register-nome-tenant"
                type="text"
                required
                value={nomeTenant}
                onChange={(e) => setNomeTenant(e.target.value)}
                className="w-full px-4 py-2 rounded-lg border border-input bg-background"
              />
            </div>

            {erro && (
              <p role="alert" className="text-sm text-destructive">
                {erro}
              </p>
            )}

            <button
              type="submit"
              disabled={enviando}
              className="w-full px-4 py-3 bg-primary text-primary-foreground rounded-full font-semibold hover:bg-primary/90 active:scale-95 transition-all disabled:opacity-60"
            >
              {enviando ? "Criando conta…" : "Criar conta"}
            </button>
          </form>

          <p className="text-center text-sm text-muted-foreground">
            Já tem conta?{" "}
            <Link to="/login" className="font-medium text-foreground hover:underline">
              Entrar
            </Link>
          </p>
        </div>
      </div>
    </main>
  );
}
