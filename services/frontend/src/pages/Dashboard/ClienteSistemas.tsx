import { useCallback, useEffect, useState } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { ArrowRight } from 'lucide-react';
import { TonalCard } from '../../components/m3/TonalCard';
import { ElevatedCard } from '../../components/m3/ElevatedCard';
import { Skeleton, EmptyState, ErrorState } from '../../components/ui/StateViews';
import { Input } from '../../components/ui/input';
import { Button } from '../../components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  DialogClose,
} from '../../components/ui/dialog';
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
  const [confirmandoExclusao, setConfirmandoExclusao] = useState(false);

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
      setConfirmandoExclusao(false);
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
              <Input
                id="cliente-nome"
                value={nome}
                onChange={(e) => setNome(e.target.value)}
                className="flex-1"
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
              <p role="status" className="text-sm text-success">
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

      <ElevatedCard>
        <div className="flex items-center justify-between gap-4 mb-1">
          <h3 className="text-lg font-heading font-bold">Sistemas</h3>
          {estadoTenant.fase === 'pronto' && (
            <button
              type="button"
              onClick={() => setConfirmandoExclusao(true)}
              disabled={excluindo}
              className="shrink-0 text-sm bg-destructive text-destructive-foreground px-4 py-2 rounded-lg font-medium hover:bg-destructive/90 disabled:opacity-60"
            >
              {excluindo ? 'Excluindo…' : 'Excluir cliente'}
            </button>
          )}
        </div>
        {erroExcluir && (
          <p role="alert" className="text-sm text-destructive mb-4">
            {erroExcluir}
          </p>
        )}

        {estadoTenant.fase === 'pronto' && (
          <Dialog open={confirmandoExclusao} onOpenChange={setConfirmandoExclusao}>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Excluir {estadoTenant.tenant.nome}?</DialogTitle>
                <DialogDescription>
                  Esta ação é permanente e exclui todos os sistemas e dados deste cliente.
                </DialogDescription>
              </DialogHeader>
              <DialogFooter>
                <DialogClose render={<Button type="button" variant="outline" />}>Cancelar</DialogClose>
                <Button type="button" variant="destructive" onClick={excluir} disabled={excluindo}>
                  {excluindo ? 'Excluindo…' : 'Confirmar exclusão'}
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        )}

        {estado.fase === 'erro' ? (
          <div className="mt-4">
            <ErrorState mensagem="Não foi possível carregar os sistemas deste cliente." onRepetir={recarregar} />
          </div>
        ) : estado.fase === 'carregando' ? (
          <div className="mt-4">
            <Skeleton itens={3} />
          </div>
        ) : estado.fase === 'vazio' ? (
          <div className="mt-4">
            <EmptyState
              titulo="Nenhum sistema para este cliente ainda"
              descricao={usuario.podeCriarSistema ? 'Crie o primeiro sistema para começar.' : undefined}
            />
          </div>
        ) : (
          <div className="overflow-x-auto -mx-6 -mb-6 mt-4">
            <table className="w-full text-left">
              <thead>
                <tr className="border-b border-border text-xs font-semibold uppercase text-muted-foreground">
                  <th scope="col" className="px-6 py-4">
                    Nome
                  </th>
                  <th scope="col" className="px-6 py-4 text-right">
                    Ações
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {estado.sistemas.map((s) => (
                  <tr key={s.id}>
                    <td className="px-6 py-4 font-medium">{s.nome}</td>
                    <td className="px-6 py-4 text-right">
                      <Link
                        to={`/dashboard/clientes/${tenantId}/sistemas/${s.id}`}
                        className="inline-flex items-center gap-1.5 text-sm font-medium px-4 py-2 rounded-full bg-primary/10 text-primary hover:bg-primary/20 transition-colors"
                      >
                        Abrir sistema
                        <ArrowRight className="w-4 h-4" />
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </ElevatedCard>
    </div>
  );
}
