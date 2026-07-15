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

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <ElevatedCard className="bg-card text-card-foreground">
          <div className="flex justify-between items-start mb-4">
            <h3 className="text-lg font-heading font-bold">ERP Financeiro</h3>
            <span className="text-xs font-semibold bg-green-100 text-green-700 px-2 py-1 rounded-full">Ativo</span>
          </div>
          <p className="text-sm text-muted-foreground mb-4">Módulo de gestão financeira com integração bancária.</p>
          <div className="text-sm font-medium text-primary cursor-pointer hover:underline">Abrir projeto &rarr;</div>
        </ElevatedCard>

        <ElevatedCard className="bg-card text-card-foreground border-dashed border-2 border-border shadow-none hover:border-primary/50 cursor-pointer flex items-center justify-center flex-col min-h-[160px]">
          <div className="text-4xl text-muted-foreground mb-2">+</div>
          <p className="text-sm font-medium text-muted-foreground">Criar novo projeto</p>
        </ElevatedCard>
      </div>
    </div>
  );
}
