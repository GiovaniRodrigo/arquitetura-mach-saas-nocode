import { TonalCard } from '../../components/m3/TonalCard';

export function Settings() {
  return (
    <div className="flex flex-col gap-6 max-w-5xl mx-auto pb-8">
      <TonalCard className="bg-secondary text-secondary-foreground border-none">
        <h2 className="text-2xl font-heading font-bold mb-2">Configurações</h2>
        <p className="text-muted-foreground text-sm font-medium">
          Ajuste as preferências da sua conta e configurações globais do sistema.
        </p>
      </TonalCard>
    </div>
  );
}
