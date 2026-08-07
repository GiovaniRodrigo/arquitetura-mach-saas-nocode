import { useParams, Link } from 'react-router-dom';
import { TonalCard } from '../../components/m3/TonalCard';
import { ElevatedCard } from '../../components/m3/ElevatedCard';
import { Skeleton, EmptyState, ErrorState } from '../../components/ui/StateViews';
import { useApp } from '../../app/AppContext';
import { useSistemas } from '../../systems/useSistemas';

// Sistemas do tenant selecionado (RF08, RN05): um cliente pode ter vários
// sistemas; escolher um abre as abas Telas/Regras de Negócio/Versão.
export function ClienteSistemas() {
  const { tenantId = '' } = useParams<{ tenantId: string }>();
  const { client, usuario } = useApp();
  const { estado, recarregar } = useSistemas(client, tenantId);

  return (
    <div className="flex flex-col gap-6 max-w-5xl mx-auto pb-8">
      <TonalCard className="bg-primary/10 text-primary border-none">
        <h2 className="text-2xl font-heading font-bold mb-2">Sistemas do cliente</h2>
        <p className="text-primary/80 text-sm font-medium">
          Selecione um sistema para editar telas, regras de negócio e versões.
        </p>
      </TonalCard>

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
    </div>
  );
}
