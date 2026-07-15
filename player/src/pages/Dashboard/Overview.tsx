import { TonalCard } from '../../components/m3/TonalCard';

export function Overview() {
  return (
    <div className="flex flex-col gap-6 max-w-5xl mx-auto pb-8">
      <TonalCard className="bg-primary/10 text-primary border-none">
        <h2 className="text-2xl font-heading font-bold mb-2">Início</h2>
        <p className="text-primary/80 text-sm font-medium">Bem-vindo à Plataforma MACH.</p>
      </TonalCard>
    </div>
  );
}
