// Tela de seleção de sistemas exibida quando o utilizador está autenticado mas
// não escolheu um sistema (substitui o antigo empty state com ?sistema=<id>).
// Lista os sistemas do tenant e permite criar um novo (restrito a dono/parceiro
// no Gateway — um 403 é tratado com mensagem amigável). Ao escolher/criar,
// navega para ?sistema=<id>, reiniciando o Frontend já com o sistema no config.
// A listagem/estados vêm do hook compartilhado useSistemas (RNF05).

import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { ApiClient, ApiError } from "../api/client";
import { useSistemas } from "./useSistemas";
import { Skeleton, EmptyState, ErrorState } from "../components/ui/StateViews";
import { Monitor, Plus, RefreshCcw } from "lucide-react";
import type { UsuarioAutenticado } from "../auth/jwt";

export function SeletorSistemas({
  client,
  usuario,
}: {
  client: ApiClient;
  usuario: UsuarioAutenticado;
}) {
  const { estado, recarregar, criar } = useSistemas(client);
  const navigate = useNavigate();
  const [nome, setNome] = useState("");
  const [criando, setCriando] = useState(false);
  const [erroCriar, setErroCriar] = useState<string | null>(null);

  // Abre o construtor do sistema (abas Telas/Regras de Negócio/Versão) no
  // tenant do usuário autenticado — não mais o runtime do sistema (abrirSistema).
  function abrirConstrutor(sistemaId: string) {
    navigate(`/dashboard/clientes/${usuario.tenantId}/sistemas/${sistemaId}/telas`);
  }

  async function submeter(e: React.FormEvent) {
    e.preventDefault();
    if (!nome.trim() || criando) return;
    setCriando(true);
    setErroCriar(null);
    try {
      const novo = await criar(nome.trim());
      abrirConstrutor(novo.id);
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
    <div className="w-full max-w-5xl mx-auto space-y-8">
      <div>
        <h1 className="font-heading text-4xl font-bold tracking-tight">Seus sistemas</h1>
        <p className="text-muted-foreground mt-2 text-lg">Escolha um sistema para abrir ou crie um novo.</p>
      </div>

      {estado.fase === "erro" ? (
        <ErrorState mensagem="Não foi possível carregar seus sistemas." onRepetir={recarregar} />
      ) : estado.fase === "carregando" ? (
        <Skeleton itens={3} />
      ) : estado.fase === "vazio" ? (
        <EmptyState titulo="Nenhum sistema ainda" descricao="Crie o primeiro abaixo." />
      ) : (
        <ul className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
          {estado.sistemas.map((s) => (
            <li key={s.id}>
              <button
                type="button"
                onClick={() => abrirConstrutor(s.id)}
                className="w-full h-full text-left p-6 bg-card border border-border hover:border-primary/50 hover:bg-secondary/80 rounded-2xl transition-all group flex flex-col gap-4"
              >
                <div className="w-12 h-12 bg-secondary group-hover:bg-primary/20 rounded-xl flex items-center justify-center text-muted-foreground group-hover:text-primary transition-colors">
                  <Monitor className="w-6 h-6" />
                </div>
                <div>
                  <h3 className="font-heading font-semibold text-xl text-foreground">{s.nome}</h3>
                  <p className="text-muted-foreground text-sm mt-1 flex items-center gap-2">
                    <span className="w-2 h-2 rounded-full bg-emerald-500"></span> Online
                  </p>
                </div>
              </button>
            </li>
          ))}
        </ul>
      )}

      {/* Sem permissão de criação (RN10), o formulário fica oculto — não expõe
          erro. O tratamento de 403 abaixo permanece como defesa em profundidade
          (ex.: claim ausente/desatualizado), já que a autorização real é do
          Gateway/IAM, nunca deste claim. */}
      {usuario.podeCriarSistema && (
        <form onSubmit={submeter} className="pt-8 border-t border-border max-w-md">
          <label htmlFor="novo-sistema" className="block text-sm font-semibold text-foreground mb-2">
            Criar novo sistema
          </label>
          <div className="flex flex-col sm:flex-row gap-3">
            <input
              id="novo-sistema"
              value={nome}
              onChange={(e) => setNome(e.target.value)}
              placeholder="Nome do sistema"
              className="flex-1 px-4 py-3 bg-card border border-border rounded-xl focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent transition-all"
            />
            <button
              type="submit"
              disabled={!nome.trim() || criando}
              className="flex items-center justify-center px-6 py-3 bg-primary hover:bg-primary/90 text-primary-foreground font-medium rounded-full transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
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
      )}
    </div>
  );
}
