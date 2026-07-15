import { TonalCard } from '../../components/m3/TonalCard';
import { ElevatedCard } from '../../components/m3/ElevatedCard';

export function Settings() {
  return (
    <div className="flex flex-col gap-6 max-w-5xl mx-auto pb-8">
      <TonalCard className="bg-secondary text-secondary-foreground border-none">
        <h2 className="text-2xl font-heading font-bold mb-2">Configurações</h2>
        <p className="text-muted-foreground text-sm font-medium">
          Ajuste as preferências da sua conta e configurações globais do sistema.
        </p>
      </TonalCard>

      <div className="flex flex-col gap-4">
        <ElevatedCard className="bg-card text-card-foreground">
          <h3 className="text-md font-heading font-bold mb-2">Perfil do Usuário</h3>
          <p className="text-sm text-muted-foreground mb-4">Atualize suas informações pessoais e foto de perfil.</p>
          <button className="text-sm bg-secondary text-secondary-foreground px-4 py-2 rounded-lg font-medium hover:bg-secondary/80">Editar Perfil</button>
        </ElevatedCard>

        <ElevatedCard className="bg-card text-card-foreground">
          <h3 className="text-md font-heading font-bold mb-2">Aparência</h3>
          <p className="text-sm text-muted-foreground mb-4">Escolha entre o tema claro ou escuro (Dark Mode).</p>
          <button className="text-sm bg-secondary text-secondary-foreground px-4 py-2 rounded-lg font-medium hover:bg-secondary/80">Alternar Tema</button>
        </ElevatedCard>
      </div>
    </div>
  );
}
