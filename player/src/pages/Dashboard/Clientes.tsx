import { TonalCard } from '../../components/m3/TonalCard';
import { ElevatedCard } from '../../components/m3/ElevatedCard';
import { Skeleton, EmptyState, ErrorState } from '../../components/ui/StateViews';
import { useApp } from '../../app/AppContext';
import { useSistemas } from '../../systems/useSistemas';
import { abrirSistema } from '../../systems/abrirSistema';
import { useNavigate } from 'react-router-dom';

export function Clientes() {
  const { client, usuario } = useApp();
  const { estado, recarregar } = useSistemas(client);
  const navigate = useNavigate();

  // Criação reutiliza a tela de seleção (que contém o formulário de criação).
  const irParaCriacao = () => navigate('/sistemas');

  // Sem permissão de criação (RN10), a UI oculta o card em vez de expor um erro.
  const cardNovo = usuario.podeCriarSistema ? (
    <ElevatedCard className="border-dashed border-2 border-border shadow-none hover:border-primary/50 cursor-pointer flex items-center justify-center flex-col min-h-[160px]">
      <button
        type="button"
        onClick={irParaCriacao}
        className="flex flex-col items-center justify-center w-full h-full"
      >
        <div className="text-4xl text-muted-foreground mb-2">+</div>
        <p className="text-sm font-medium text-muted-foreground">Criar novo projeto</p>
      </button>
    </ElevatedCard>
  ) : null;

  return (
    <div className="flex flex-col gap-6 max-w-5xl mx-auto pb-8">
      <TonalCard className="bg-primary/10 text-primary border-none">
        <h2 className="text-2xl font-heading font-bold mb-2">Clientes</h2>
        <p className="text-primary/80 text-sm font-medium">
          Gerencie seus projetos, arquiteturas e deploys em um só lugar.
        </p>
      </TonalCard>

      {estado.fase === 'erro' ? (
        <ErrorState mensagem="Não foi possível carregar seus projetos." onRepetir={recarregar} />
      ) : estado.fase === 'carregando' ? (
        <Skeleton itens={4} />
      ) : estado.fase === 'vazio' ? (
        <EmptyState
          titulo="Nenhum projeto ainda"
          descricao="Crie seu primeiro sistema para começar."
          acao={
            usuario.podeCriarSistema ? (
              <button
                type="button"
                onClick={irParaCriacao}
                className="bg-primary text-primary-foreground px-6 py-2.5 rounded-full font-semibold hover:bg-primary/90 active:scale-95 transition-all shadow-sm"
              >
                Criar projeto
              </button>
            ) : undefined
          }
        />
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {estado.sistemas.map((s) => (
            <ElevatedCard key={s.id}>
              <div className="flex justify-between items-start mb-4">
                <h3 className="text-lg font-heading font-bold">{s.nome}</h3>
                <span className="text-xs font-semibold bg-emerald-500/15 text-emerald-600 dark:text-emerald-400 px-2 py-1 rounded-full">
                  Online
                </span>
              </div>
              <button
                type="button"
                onClick={() => abrirSistema(s.id)}
                className="text-sm font-medium text-primary cursor-pointer hover:underline"
              >
                Abrir projeto &rarr;
              </button>
            </ElevatedCard>
          ))}
          {cardNovo}
        </div>
      )}
    </div>
  );
}
