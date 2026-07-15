import { TonalCard } from '../../components/m3/TonalCard';
import { ElevatedCard } from '../../components/m3/ElevatedCard';

export function Projects() {
  return (
    <div className="flex flex-col gap-6 max-w-5xl mx-auto pb-8">
      <TonalCard className="bg-primary/10 text-primary border-none">
        <h2 className="text-2xl font-heading font-bold mb-2">Projects</h2>
        <p className="text-primary/80 text-sm font-medium">
          Gerencie seus projetos, arquiteturas e deploys em um só lugar.
        </p>
      </TonalCard>

      <ElevatedCard className="bg-card text-card-foreground">
        <p className="text-sm text-muted-foreground">Nenhum projeto ainda.</p>
      </ElevatedCard>
    </div>
  );
}
