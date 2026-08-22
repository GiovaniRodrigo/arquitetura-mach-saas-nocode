import { useState } from 'react';
import { Plus, RefreshCcw } from 'lucide-react';
import { TonalCard } from '../../components/m3/TonalCard';
import { ElevatedCard } from '../../components/m3/ElevatedCard';
import { Input } from '../../components/ui/input';
import { Skeleton, EmptyState, ErrorState } from '../../components/ui/StateViews';
import { useApp } from '../../app/AppContext';
import { useTenants } from '../../clientes/useTenants';
import { ApiError } from '../../api/client';
import { Link, useNavigate } from 'react-router-dom';

// Clientes (RF07/RF08, RN01, RN05): lista os tenants vinculados ao usuário
// autenticado e permite criar um novo (restrito a dono/parceiro no Gateway —
// um 403 é tratado com mensagem amigável, mesmo padrão de SeletorSistemas).
// Selecionar um tenant navega para a lista de sistemas daquele cliente
// (ClienteSistemas) — as abas Telas/Regras/Versão só existem depois de
// escolher um sistema específico (RN05).
export function Clientes() {
  const { client, usuario } = useApp();
  const { estado, recarregar, criar } = useTenants(client);
  const navigate = useNavigate();

  const [nome, setNome] = useState('');
  const [criando, setCriando] = useState(false);
  const [erroCriar, setErroCriar] = useState<string | null>(null);

  async function submeter(e: React.FormEvent) {
    e.preventDefault();
    if (!nome.trim() || criando) return;
    setCriando(true);
    setErroCriar(null);
    try {
      const novo = await criar(nome.trim());
      setNome('');
      navigate(`/dashboard/clientes/${novo.id}`);
    } catch (err) {
      if (err instanceof ApiError && err.status === 403) {
        setErroCriar('Você não tem permissão para criar clientes.');
      } else {
        setErroCriar(err instanceof Error ? err.message : String(err));
      }
      setCriando(false);
    }
  }

  return (
    <div className="flex flex-col gap-6 max-w-5xl mx-auto pb-8">
      <TonalCard className="bg-primary/10 text-primary border-none">
        <h2 className="text-2xl font-heading font-bold mb-2">Clientes</h2>
        <p className="text-primary/80 text-sm font-medium">
          Selecione um cliente para gerenciar seus sistemas.
        </p>
      </TonalCard>

      {estado.fase === 'erro' ? (
        <ErrorState mensagem="Não foi possível carregar seus clientes." onRepetir={recarregar} />
      ) : estado.fase === 'carregando' ? (
        <Skeleton itens={4} />
      ) : estado.fase === 'vazio' ? (
        <EmptyState
          titulo="Nenhum cliente ainda"
          descricao="Os tenants vinculados a você aparecerão aqui."
        />
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {estado.tenants.map((t) => (
            <ElevatedCard key={t.id}>
              <div className="flex justify-between items-start mb-4">
                <h3 className="text-lg font-heading font-bold">{t.nome}</h3>
              </div>
              <Link
                to={`/dashboard/clientes/${t.id}`}
                className="text-sm font-medium text-primary hover:underline"
              >
                Abrir cliente &rarr;
              </Link>
            </ElevatedCard>
          ))}
        </div>
      )}

      {/* Sem permissão de criação (RN10 de 003), o formulário fica oculto —
          não expõe erro. O tratamento de 403 abaixo permanece como defesa em
          profundidade, já que a autorização real é do Gateway/IAM. */}
      {usuario.podeCriarSistema && (
        <ElevatedCard>
          <form onSubmit={submeter} className="flex flex-col gap-3">
            <label htmlFor="novo-cliente" className="text-sm font-semibold text-foreground">
              Criar novo cliente
            </label>
            <div className="flex flex-col sm:flex-row gap-3">
              <Input
                id="novo-cliente"
                value={nome}
                onChange={(e) => setNome(e.target.value)}
                placeholder="Nome do cliente/negócio"
                className="flex-1"
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
              <p role="alert" className="text-destructive text-sm">
                {erroCriar}
              </p>
            )}
          </form>
        </ElevatedCard>
      )}
    </div>
  );
}
