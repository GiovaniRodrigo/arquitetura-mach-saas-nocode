// Tela de seleção de sistemas exibida quando o utilizador está autenticado mas
// não escolheu um sistema (substitui o antigo empty state com ?sistema=<id>).
// Lista os sistemas do tenant e permite criar um novo (restrito a dono/parceiro
// no Gateway — um 403 é tratado com mensagem amigável). Ao escolher/criar,
// navega para ?sistema=<id>, reiniciando o Player já com o sistema no config.

import { useEffect, useState } from "react";
import { ApiClient, ApiError } from "../api/client";
import type { Sistema } from "../api/types";

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
    <main className="mach-seletor" style={estilos.wrap}>
      <h1 style={estilos.titulo}>Seus sistemas</h1>
      <p style={estilos.sub}>Escolha um sistema para abrir ou crie um novo.</p>

      {erro ? (
        <div role="alert" style={estilos.erro}>
          <p>Não foi possível carregar seus sistemas.</p>
          <button type="button" style={estilos.btnGhost} onClick={() => setTentativa((t) => t + 1)}>
            Tentar novamente
          </button>
        </div>
      ) : sistemas === null ? (
        <ul aria-busy="true" style={estilos.lista}>
          {[0, 1, 2].map((i) => (
            <li key={i} style={{ ...estilos.item, ...estilos.skeleton }} />
          ))}
        </ul>
      ) : sistemas.length === 0 ? (
        <p style={estilos.vazio}>Nenhum sistema ainda. Crie o primeiro abaixo.</p>
      ) : (
        <ul style={estilos.lista}>
          {sistemas.map((s) => (
            <li key={s.id}>
              <button type="button" style={estilos.item} onClick={() => abrirSistema(s.id)}>
                {s.nome}
              </button>
            </li>
          ))}
        </ul>
      )}

      <form onSubmit={criar} style={estilos.form}>
        <label htmlFor="novo-sistema" style={estilos.label}>
          Criar sistema
        </label>
        <div style={estilos.formRow}>
          <input
            id="novo-sistema"
            value={nome}
            onChange={(e) => setNome(e.target.value)}
            placeholder="Nome do sistema"
            style={estilos.input}
          />
          <button type="submit" disabled={!nome.trim() || criando} style={estilos.btnPrimary}>
            {criando ? "Criando…" : "Criar"}
          </button>
        </div>
        {erroCriar && (
          <p role="alert" style={estilos.erroInline}>
            {erroCriar}
          </p>
        )}
      </form>
    </main>
  );
}

const estilos: Record<string, React.CSSProperties> = {
  wrap: { maxWidth: 480, margin: "8vh auto", fontFamily: "system-ui, sans-serif", padding: "0 16px" },
  titulo: { fontSize: "1.5rem", marginBottom: 4 },
  sub: { color: "#475569", marginTop: 0, fontSize: ".95rem" },
  lista: { listStyle: "none", padding: 0, margin: "16px 0", display: "flex", flexDirection: "column", gap: 8 },
  item: {
    width: "100%", textAlign: "left", padding: "12px 14px", minHeight: 44,
    border: "1px solid #E2E8F0", borderRadius: 8, background: "#fff", cursor: "pointer", fontSize: "1rem",
  },
  skeleton: { background: "#EEF2F7", color: "transparent", cursor: "default", height: 44 },
  vazio: { color: "#475569", margin: "16px 0" },
  form: { marginTop: 24, borderTop: "1px solid #E2E8F0", paddingTop: 16 },
  label: { fontSize: ".85rem", fontWeight: 600, color: "#475569", display: "block", marginBottom: 6 },
  formRow: { display: "flex", gap: 8 },
  input: {
    flex: 1, minHeight: 44, padding: "0 12px", border: "1px solid #E2E8F0", borderRadius: 8, fontSize: "1rem",
  },
  btnPrimary: {
    minHeight: 44, padding: "0 18px", border: "none", borderRadius: 8,
    background: "#6366F1", color: "#fff", fontWeight: 600, cursor: "pointer",
  },
  btnGhost: {
    minHeight: 40, padding: "0 14px", border: "1px solid #E2E8F0", borderRadius: 8,
    background: "#fff", color: "#475569", cursor: "pointer", marginTop: 8,
  },
  erro: { border: "1px solid #FECACA", background: "#FEE2E2", borderRadius: 8, padding: 16, color: "#7F1D1D" },
  erroInline: { color: "#DC2626", fontSize: ".85rem", marginTop: 8 },
};
