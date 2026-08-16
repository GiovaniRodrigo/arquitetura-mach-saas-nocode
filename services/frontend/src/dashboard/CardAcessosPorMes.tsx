import { Area, AreaChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';
import { ElevatedCard } from '../components/m3/ElevatedCard';
import { Skeleton, ErrorState } from '../components/ui/StateViews';
import { useApp } from '../app/AppContext';
import { useAcessosPorMes } from './useAcessosPorMes';
import { mesAbreviado, variacaoPercentual } from './mesesFormato';

export function CardAcessosPorMes() {
  const { client } = useApp();
  const { estado, recarregar } = useAcessosPorMes(client);

  return (
    <ElevatedCard>
      <div className="flex items-start justify-between mb-1">
        <div>
          <h3 className="text-md font-heading font-bold">Acessos por mês</h3>
          <p className="text-xs text-muted-foreground">Últimos 6 meses</p>
        </div>
        {estado.fase === 'pronto' && (() => {
          const variacao = variacaoPercentual(estado.pontos.map((p) => p.total));
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
        <ErrorState mensagem="Não foi possível carregar os acessos por mês." onRepetir={recarregar} />
      ) : estado.fase === 'carregando' ? (
        <Skeleton itens={1} />
      ) : (
        <div className="h-52 -ml-2 mt-4" aria-label="Gráfico de acessos por mês">
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={estado.pontos.map((p) => ({ mes: mesAbreviado(p.competencia), total: p.total }))}>
              <defs>
                <linearGradient id="gradienteAcessos" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#6366f1" stopOpacity={0.35} />
                  <stop offset="95%" stopColor="#6366f1" stopOpacity={0} />
                </linearGradient>
              </defs>
              <XAxis
                dataKey="mes"
                axisLine={false}
                tickLine={false}
                tick={{ fontSize: 12, fill: 'hsl(var(--muted-foreground))' }}
              />
              <YAxis axisLine={false} tickLine={false} tick={{ fontSize: 12, fill: 'hsl(var(--muted-foreground))' }} width={32} />
              <Tooltip
                formatter={(valor) => [String(valor), 'Acessos']}
                contentStyle={{
                  background: 'hsl(var(--popover))',
                  color: 'hsl(var(--popover-foreground))',
                  border: '1px solid hsl(var(--border))',
                  borderRadius: '0.75rem',
                  fontSize: '0.75rem',
                }}
              />
              <Area
                type="monotone"
                dataKey="total"
                stroke="#6366f1"
                strokeWidth={2}
                fill="url(#gradienteAcessos)"
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      )}
    </ElevatedCard>
  );
}
