import { ElevatedCard } from '../components/m3/ElevatedCard';
import { Skeleton, EmptyState, ErrorState } from '../components/ui/StateViews';
import { useApp } from '../app/AppContext';
import { useUltimosAcessos } from './useUltimosAcessos';

function formatarData(iso: string): string {
  return new Date(iso).toLocaleString('pt-BR', { dateStyle: 'short', timeStyle: 'short' });
}

export function CardUltimosAcessos() {
  const { client } = useApp();
  const { estado, recarregar } = useUltimosAcessos(client);

  return (
    <ElevatedCard>
      <h3 className="text-md font-heading font-bold mb-4">Últimos Acessos</h3>
      {estado.fase === 'erro' ? (
        <ErrorState mensagem="Não foi possível carregar os últimos acessos." onRepetir={recarregar} />
      ) : estado.fase === 'carregando' ? (
        <Skeleton itens={3} />
      ) : estado.fase === 'vazio' ? (
        <EmptyState titulo="Nenhum acesso registrado ainda" />
      ) : (
        <ul className="flex flex-col gap-3">
          {estado.eventos.map((evento, i) => (
            <li key={i} className="flex items-center justify-between text-sm">
              <span className="font-medium">{evento.usuario_nome}</span>
              <span className="text-muted-foreground">{evento.tenant_nome}</span>
              <span className="text-muted-foreground">{formatarData(evento.criado_em)}</span>
            </li>
          ))}
        </ul>
      )}
    </ElevatedCard>
  );
}
