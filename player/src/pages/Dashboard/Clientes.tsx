import { TonalCard } from '../../components/m3/TonalCard';
import { ElevatedCard } from '../../components/m3/ElevatedCard';
import { Skeleton, EmptyState, ErrorState } from '../../components/ui/StateViews';
import { useApp } from '../../app/AppContext';
import { useTenants } from '../../clientes/useTenants';
import { Link } from 'react-router-dom';

// Clientes (RF07/RF08, RN01, RN05): lista os tenants vinculados ao usuário
// autenticado. Selecionar um tenant navega para a lista de sistemas daquele
// cliente (ClienteSistemas) — as abas Telas/Regras/Versão só existem depois
// de escolher um sistema específico (RN05).
export function Clientes() {
  const { client } = useApp();
  const { estado, recarregar } = useTenants(client);

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
    </div>
  );
}
