import { Bar, BarChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';
import { ElevatedCard } from '../components/m3/ElevatedCard';
import { Skeleton, ErrorState } from '../components/ui/StateViews';
import { useApp } from '../app/AppContext';
import { useReceitaPorMes } from './useReceitaPorMes';
import { mesAbreviado, variacaoPercentual } from './mesesFormato';

function formatarReceita(centavos: number, moeda: string): string {
  return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: moeda }).format(centavos / 100);
}

export function CardReceitaPorMes() {
  const { client } = useApp();
  const { estado, recarregar } = useReceitaPorMes(client);

  return (
    <ElevatedCard>
      <div className="flex items-start justify-between mb-1">
        <div>
          <h3 className="text-md font-heading font-bold">Receita de assinatura</h3>
          <p className="text-xs text-muted-foreground">Últimos 6 meses</p>
        </div>
        {estado.fase === 'pronto' && (() => {
          const variacao = variacaoPercentual(estado.pontos.map((p) => p.valor_centavos));
          if (variacao === null) return null;
          return (
            <span
              className={`text-xs font-semibold px-2 py-0.5 rounded-full ${
                variacao >= 0 ? 'bg-success/15 text-success' : 'bg-destructive/15 text-destructive'
              }`}
            >
              {variacao >= 0 ? '+' : ''}
              {variacao}%
            </span>
          );
        })()}
      </div>

      {estado.fase === 'erro' ? (
        <ErrorState mensagem="Não foi possível carregar a receita por mês." onRepetir={recarregar} />
      ) : estado.fase === 'carregando' ? (
        <Skeleton itens={1} />
      ) : (
        <div className="h-52 -ml-2 mt-4" aria-label="Gráfico de receita de assinatura por mês">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart
              data={estado.pontos.map((p) => ({ mes: mesAbreviado(p.competencia), valor: p.valor_centavos / 100 }))}
            >
              <XAxis
                dataKey="mes"
                axisLine={false}
                tickLine={false}
                tick={{ fontSize: 12, fill: 'hsl(var(--muted-foreground))' }}
              />
              <YAxis axisLine={false} tickLine={false} tick={{ fontSize: 12, fill: 'hsl(var(--muted-foreground))' }} width={48} />
              <Tooltip
                formatter={(valor) => [formatarReceita(Number(valor) * 100, estado.moeda), 'Receita']}
                contentStyle={{
                  background: 'hsl(var(--popover))',
                  color: 'hsl(var(--popover-foreground))',
                  border: '1px solid hsl(var(--border))',
                  borderRadius: '0.75rem',
                  fontSize: '0.75rem',
                }}
              />
              <Bar dataKey="valor" fill="#6366f1" radius={[6, 6, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      )}
    </ElevatedCard>
  );
}
