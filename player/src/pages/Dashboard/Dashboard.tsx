import { TonalCard } from '../../components/m3/TonalCard';
import { ElevatedCard } from '../../components/m3/ElevatedCard';
import { FabButton } from '../../components/m3/FabButton';
import { useApp } from '../../app/AppContext';
import { useMetricas } from '../../dashboard/useMetricas';
import { useNavigate } from 'react-router-dom';

export function Dashboard() {
  const { client, usuario } = useApp();
  const { estado } = useMetricas(client);
  const navigate = useNavigate();

  // "Get Started" e o FAB levam ao fluxo de criação de sistema (C1/C2).
  const criar = () => navigate('/sistemas');

  const totalTexto =
    estado.fase === 'pronto'
      ? String(estado.metricas.sistemas_total)
      : estado.fase === 'erro'
        ? '—'
        : '…';

  return (
    <div className="flex flex-col gap-6 max-w-5xl mx-auto pb-8">
      {/* Hero Card */}
      <TonalCard className="bg-primary/10 text-primary border-none relative overflow-hidden">
        <div className="absolute -top-24 -right-24 w-64 h-64 bg-primary/20 blur-3xl rounded-full pointer-events-none" />
        <h2 className="text-2xl md:text-3xl font-heading font-bold mb-2">Build your Next Flow</h2>
        <p className="text-primary/80 mb-6 max-w-xl text-sm md:text-base font-medium">
          Start creating projects and designing your business architecture with our intuitive node-based editor.
        </p>
        {usuario.podeCriarSistema && (
          <button
            type="button"
            onClick={criar}
            className="bg-primary text-primary-foreground px-6 py-2.5 rounded-full font-semibold hover:bg-primary/90 active:scale-95 transition-all shadow-sm"
          >
            Get Started
          </button>
        )}
      </TonalCard>

      {/* Metrics Section */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <ElevatedCard>
          <h3 className="text-sm font-medium text-muted-foreground mb-1">Sistemas</h3>
          <p
            className="text-3xl font-heading font-bold text-foreground"
            aria-busy={estado.fase === 'carregando'}
          >
            {totalTexto}
          </p>
        </ElevatedCard>
        <ElevatedCard>
          <h3 className="text-sm font-medium text-muted-foreground mb-1">Publicados</h3>
          <p className="text-3xl font-heading font-bold text-muted-foreground/60">—</p>
        </ElevatedCard>
        <ElevatedCard>
          <h3 className="text-sm font-medium text-muted-foreground mb-1">Rascunhos</h3>
          <p className="text-3xl font-heading font-bold text-muted-foreground/60">—</p>
        </ElevatedCard>
      </div>

      {/* FAB */}
      {usuario.podeCriarSistema && (
        <FabButton icon="+" onClick={criar} className="md:bottom-8 md:right-8">
          Create
        </FabButton>
      )}
    </div>
  );
}
