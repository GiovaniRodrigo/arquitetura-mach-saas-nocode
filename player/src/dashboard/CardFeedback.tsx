import { ElevatedCard } from '../components/m3/ElevatedCard';
import { Skeleton, EmptyState, ErrorState } from '../components/ui/StateViews';
import { useApp } from '../app/AppContext';
import { useFeedback } from './useFeedback';

export function CardFeedback() {
  const { client } = useApp();
  const { estado, recarregar, marcarRespondido } = useFeedback(client);

  return (
    <ElevatedCard>
      <h3 className="text-md font-heading font-bold mb-4">Reclamações/Feedback</h3>
      {estado.fase === 'erro' ? (
        <ErrorState mensagem="Não foi possível carregar as mensagens." onRepetir={recarregar} />
      ) : estado.fase === 'carregando' ? (
        <Skeleton itens={3} />
      ) : estado.fase === 'vazio' ? (
        <EmptyState titulo="Nenhuma mensagem por aqui" />
      ) : (
        <ul className="flex flex-col gap-3">
          {estado.itens.map((item) => (
            <li key={item.id} className="flex flex-col gap-1 text-sm border-b border-border pb-3 last:border-none">
              <div className="flex items-center justify-between">
                <span className="font-medium">{item.tenant_nome}</span>
                <span
                  className={`text-xs font-semibold px-2 py-0.5 rounded-full ${
                    item.status === 'pendente'
                      ? 'bg-amber-500/15 text-amber-600 dark:text-amber-400'
                      : 'bg-emerald-500/15 text-emerald-600 dark:text-emerald-400'
                  }`}
                >
                  {item.status === 'pendente' ? 'Pendente' : 'Respondido'}
                </span>
              </div>
              <p className="text-muted-foreground">{item.mensagem}</p>
              {item.status === 'pendente' && (
                <button
                  type="button"
                  onClick={() => void marcarRespondido(item.id)}
                  className="self-start text-xs font-medium text-primary hover:underline"
                >
                  Marcar como respondido
                </button>
              )}
            </li>
          ))}
        </ul>
      )}
    </ElevatedCard>
  );
}
