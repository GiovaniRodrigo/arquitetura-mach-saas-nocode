import { useCallback, useEffect, useState } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { TonalCard } from '../../components/m3/TonalCard';
import { ElevatedCard } from '../../components/m3/ElevatedCard';
import { Skeleton, EmptyState, ErrorState } from '../../components/ui/StateViews';
import { useApp } from '../../app/AppContext';
import { useSistemas } from '../../systems/useSistemas';
import { ApiError } from '../../api/client';
import type { Tenant } from '../../api/types';

type EstadoTenant =
  | { fase: 'carregando' }
  | { fase: 'pronto'; tenant: Tenant }
  | { fase: 'erro'; mensagem: string };

// Sistemas do tenant selecionado (RF07/RF08, RN05): visualiza/edita o nome do
// cliente, permite excluí-lo (em cascata com seus sistemas e dados) e lista
// os sistemas; escolher um abre as abas Telas/Regras de Negócio/Versão.
export function ClienteSistemas() {
  const { tenantId = '' } = useParams<{ tenantId: string }>();
  const { client, usuario } = useApp();
  const navigate = useNavigate();
  const { estado, recarregar } = useSistemas(client, tenantId);

  const [estadoTenant, setEstadoTenant] = useState<EstadoTenant>({ fase: 'carregando' });
  const [nome, setNome] = useState('');
  const [salvando, setSalvando] = useState(false);
  const [salvo, setSalvo] = useState(false);
  const [erroSalvar, setErroSalvar] = useState<string | null>(null);
  const [excluindo, setExcluindo] = useState(false);
  const [erroExcluir, setErroExcluir] = useState<string | null>(null);

  const carregarTenant = useCallback(() => {
    let vivo = true;
    setEstadoTenant({ fase: 'carregando' });
    client
      .obterTenant(tenantId)
      .then((t) => {
        if (!vivo) return;
        setEstadoTenant({ fase: 'pronto', tenant: t });
        setNome(t.nome);
      })
      .catch((e) => {
        if (!vivo) return;
        setEstadoTenant({ fase: 'erro', mensagem: e instanceof Error ? e.message : String(e) });
      });
    return () => {
      vivo = false;
    };
  }, [client, tenantId]);

  useEffect(() => carregarTenant(), [carregarTenant]);

  async function salvarNome(e: React.FormEvent) {
    e.preventDefault();
    if (estadoTenant.fase !== 'pronto' || !nome.trim() || salvando) return;
    setSalvando(true);
    setSalvo(false);
    setErroSalvar(null);
    try {
      const atualizado = await client.atualizarTenant(tenantId, nome.trim());
      setEstadoTenant({ fase: 'pronto', tenant: atualizado });
      setSalvo(true);
    } catch (err) {
      setErroSalvar(err instanceof ApiError ? err.message : String(err));
    } finally {
      setSalvando(false);
    }
  }

  async function excluir() {
    setExcluindo(true);
    setErroExcluir(null);
    try {
      await client.excluirTenant(tenantId);
      navigate('/dashboard/clientes');
    } catch (err) {
      setErroExcluir(err instanceof ApiError ? err.message : String(err));
      setExcluindo(false);
    }
  }

  return (
    <div className="flex flex-col gap-6 max-w-5xl mx-auto pb-8">
      <TonalCard className="bg-primary/10 text-primary border-none">
        <h2 className="text-2xl font-heading font-bold mb-2">
          {estadoTenant.fase === 'pronto' ? estadoTenant.tenant.nome : 'Sistemas do cliente'}
        </h2>
        <p className="text-primary/80 text-sm font-medium">
          Selecione um sistema para editar telas, regras de negócio e versões.
        </p>
      </TonalCard>

      {estadoTenant.fase === 'erro' ? (
        <ErrorState mensagem="Não foi possível carregar este cliente." onRepetir={carregarTenant} />
      ) : estadoTenant.fase === 'carregando' ? (
        <Skeleton itens={1} />
      ) : (
        <ElevatedCard>
          <form onSubmit={salvarNome} className="flex flex-col gap-3">
            <label htmlFor="cliente-nome" className="text-sm font-semibold text-foreground">
              Nome do cliente
            </label>
            <div className="flex flex-col sm:flex-row gap-3">
              <input
                id="cliente-nome"
                value={nome}
                onChange={(e) => setNome(e.target.value)}
                className="flex-1 px-4 py-3 bg-background border border-border rounded-xl focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent transition-all"
              />
              <button
                type="submit"
                disabled={!nome.trim() || salvando || nome.trim() === estadoTenant.tenant.nome}
                className="flex items-center justify-center px-6 py-3 bg-primary hover:bg-primary/90 text-primary-foreground font-medium rounded-full transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {salvando ? 'Salvando…' : 'Salvar'}
              </button>
            </div>
            {salvo && (
              <p role="status" className="text-sm text-emerald-600 dark:text-emerald-400">
                Nome atualizado.
              </p>
            )}
            {erroSalvar && (
              <p role="alert" className="text-destructive text-sm">
                {erroSalvar}
              </p>
            )}
          </form>
        </ElevatedCard>
      )}

      {estado.fase === 'erro' ? (
        <ErrorState mensagem="Não foi possível carregar os sistemas deste cliente." onRepetir={recarregar} />
      ) : estado.fase === 'carregando' ? (
        <Skeleton itens={3} />
      ) : estado.fase === 'vazio' ? (
        <EmptyState
          titulo="Nenhum sistema para este cliente ainda"
          descricao={usuario.podeCriarSistema ? 'Crie o primeiro sistema para começar.' : undefined}
        />
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {estado.sistemas.map((s) => (
            <ElevatedCard key={s.id}>
              <h3 className="text-lg font-heading font-bold mb-4">{s.nome}</h3>
              <Link
                to={`/dashboard/clientes/${tenantId}/sistemas/${s.id}`}
                className="text-sm font-medium text-primary hover:underline"
              >
                Abrir sistema &rarr;
              </Link>
            </ElevatedCard>
          ))}
        </div>
      )}

      {estadoTenant.fase === 'pronto' && (
        <ElevatedCard>
          <h3 className="text-md font-heading font-bold mb-2 text-destructive">Excluir cliente</h3>
          <p className="text-sm text-muted-foreground mb-4">
            Esta ação é permanente e exclui todos os sistemas e dados deste cliente.
          </p>
          <button
            type="button"
            onClick={excluir}
            disabled={excluindo}
            className="self-start text-sm bg-destructive text-destructive-foreground px-4 py-2 rounded-lg font-medium hover:bg-destructive/90 disabled:opacity-60"
          >
            {excluindo ? 'Excluindo…' : 'Excluir cliente'}
          </button>
          {erroExcluir && (
            <p role="alert" className="text-sm text-destructive">
              {erroExcluir}
            </p>
          )}
        </ElevatedCard>
      )}
    </div>
  );
}
