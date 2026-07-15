// Tela de seleção de sistemas exibida quando o utilizador está autenticado mas
// não escolheu um sistema (substitui o antigo empty state com ?sistema=<id>).
// Lista os sistemas do tenant e permite criar um novo (restrito a dono/parceiro
// no Gateway — um 403 é tratado com mensagem amigável). Ao escolher/criar,
// navega para ?sistema=<id>, reiniciando o Player já com o sistema no config.

import { useEffect, useState } from "react";
import { ApiClient, ApiError } from "../api/client";
import type { Sistema } from "../api/types";
import { Monitor, Plus, RefreshCcw } from "lucide-react";

/** Recarrega o Player com o sistema selecionado no query string. */
function abrirSistema(id: string) {
  const params = new URLSearchParams(window.location.search);
  params.set("sistema", id);
  window.location.search = params.toString();
}

export function SeletorSistemas({ client }: { client: ApiClient }) {
  const [sistemas, setSistemas] = useState<Sistema[] | null>(null);
  const [erro, setErro] = useState<string | null>(null);
  const [nome, setNome] = useState("");
  const [criando, setCriando] = useState(false);
  const [erroCriar, setErroCriar] = useState<string | null>(null);
  const [tentativa, setTentativa] = useState(0);

  useEffect(() => {
    let vivo = true;
    setSistemas(null);
    setErro(null);
    (async () => {
      try {
        const lista = await client.listarSistemas();
        if (vivo) setSistemas(lista);
      } catch (e) {
        if (vivo) setErro(e instanceof Error ? e.message : String(e));
      }
    })();
    return () => {
      vivo = false;
    };
  }, [client, tentativa]);

  async function criar(e: React.FormEvent) {
    e.preventDefault();
    if (!nome.trim() || criando) return;
    setCriando(true);
    setErroCriar(null);
    try {
      const novo = await client.criarSistema(nome.trim());
      abrirSistema(novo.id);
    } catch (err) {
      if (err instanceof ApiError && err.status === 403) {
        setErroCriar("Você não tem permissão para criar sistemas.");
      } else {
        setErroCriar(err instanceof Error ? err.message : String(err));
      }
      setCriando(false);
    }
  }

  return (
    <main className="min-h-screen bg-background text-foreground font-sans p-6 md:p-12 flex flex-col items-center">
      <div className="w-full max-w-5xl space-y-8 mt-8">
        <div>
          <h1 className="font-heading text-4xl font-bold tracking-tight">Seus sistemas</h1>
          <p className="text-muted-foreground mt-2 text-lg">Escolha um sistema para abrir ou crie um novo.</p>
        </div>

        {erro ? (
          <div role="alert" className="bg-destructive/10 border border-destructive/20 text-destructive p-4 rounded-xl flex flex-col items-start gap-4">
            <p>Não foi possível carregar seus sistemas.</p>
            <button type="button" onClick={() => setTentativa((t) => t + 1)} className="flex items-center px-4 py-2 bg-background border border-border rounded-full hover:bg-zinc-800 transition-colors">
              <RefreshCcw className="w-4 h-4 mr-2" />
              Tentar novamente
            </button>
          </div>
        ) : sistemas === null ? (
          <ul aria-busy="true" className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {[0, 1, 2].map((i) => (
              <li key={i} className="h-40 bg-zinc-900 border border-zinc-800 rounded-2xl animate-pulse" />
            ))}
          </ul>
        ) : sistemas.length === 0 ? (
          <p className="text-muted-foreground bg-zinc-900/50 p-8 rounded-2xl text-center border border-zinc-800">Nenhum sistema ainda. Crie o primeiro abaixo.</p>
        ) : (
          <ul className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
            {sistemas.map((s) => (
              <li key={s.id}>
                <button 
                  type="button" 
                  onClick={() => abrirSistema(s.id)}
                  className="w-full h-full text-left p-6 bg-zinc-900 border border-zinc-800 hover:border-primary/50 hover:bg-zinc-800/80 rounded-2xl transition-all group flex flex-col gap-4"
                >
                  <div className="w-12 h-12 bg-zinc-800 group-hover:bg-primary/20 rounded-xl flex items-center justify-center text-zinc-400 group-hover:text-primary transition-colors">
                    <Monitor className="w-6 h-6" />
                  </div>
                  <div>
                    <h3 className="font-heading font-semibold text-xl text-zinc-100">{s.nome}</h3>
                    <p className="text-zinc-400 text-sm mt-1 flex items-center gap-2">
                      <span className="w-2 h-2 rounded-full bg-emerald-500"></span> Online
                    </p>
                  </div>
                </button>
              </li>
            ))}
          </ul>
        )}

        <form onSubmit={criar} className="mt-12 pt-8 border-t border-border max-w-md">
          <label htmlFor="novo-sistema" className="block text-sm font-semibold text-zinc-300 mb-2">
            Criar novo sistema
          </label>
          <div className="flex flex-col sm:flex-row gap-3">
            <input
              id="novo-sistema"
              value={nome}
              onChange={(e) => setNome(e.target.value)}
              placeholder="Nome do sistema"
              className="flex-1 px-4 py-3 bg-zinc-900 border border-zinc-800 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent transition-all"
            />
            <button 
              type="submit" 
              disabled={!nome.trim() || criando} 
              className="flex items-center justify-center px-6 py-3 bg-primary hover:bg-primary/90 text-white font-medium rounded-full transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {criando ? (
                <RefreshCcw className="w-5 h-5 mr-2 animate-spin" />
              ) : (
                <Plus className="w-5 h-5 mr-2" />
              )}
              Criar
            </button>
          </div>
          {erroCriar && (
            <p role="alert" className="text-destructive text-sm mt-3">
              {erroCriar}
            </p>
          )}
        </form>
      </div>
    </main>
  );
}
