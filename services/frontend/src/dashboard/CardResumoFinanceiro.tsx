import { ElevatedCard } from '../components/m3/ElevatedCard';
import { Skeleton, ErrorState } from '../components/ui/StateViews';
import { useApp } from '../app/AppContext';
import { useResumoFinanceiro } from './useResumoFinanceiro';

function formatarReceita(centavos: number, moeda: string): string {
  return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: moeda }).format(centavos / 100);
}

export function CardResumoFinanceiro() {
  const { client } = useApp();
  const { estado, recarregar } = useResumoFinanceiro(client);

  return (
    <ElevatedCard>
      <h3 className="text-md font-heading font-bold mb-4">Resumo Financeiro</h3>
      {estado.fase === 'erro' ? (
        <ErrorState mensagem="Não foi possível carregar o resumo financeiro." onRepetir={recarregar} />
      ) : estado.fase === 'carregando' ? (
        <Skeleton itens={1} />
      ) : (
        <div className="flex flex-col gap-1">
          <p className="text-3xl font-heading font-bold text-foreground">
            {formatarReceita(estado.resumo.receita_total_centavos, estado.resumo.moeda)}
          </p>
          <p className="text-sm text-muted-foreground">
            Receita de assinatura em {estado.resumo.competencia}
          </p>
        </div>
      )}
    </ElevatedCard>
  );
}
