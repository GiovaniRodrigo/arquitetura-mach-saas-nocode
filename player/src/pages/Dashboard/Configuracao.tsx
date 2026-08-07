import { TonalCard } from '../../components/m3/TonalCard';
import { ElevatedCard } from '../../components/m3/ElevatedCard';
import { useTheme } from '../../theme/ThemeProvider';
import { Moon, Sun } from 'lucide-react';

export function Configuracao() {
  const { tema, alternarTema } = useTheme();

  return (
    <div className="flex flex-col gap-6 max-w-5xl mx-auto pb-8">
      <TonalCard className="bg-secondary text-secondary-foreground border-none">
        <h2 className="text-2xl font-heading font-bold mb-2">Configurações</h2>
        <p className="text-muted-foreground text-sm font-medium">
          Ajuste as preferências da sua conta e configurações globais do sistema.
        </p>
      </TonalCard>

      <div className="flex flex-col gap-4">
        <ElevatedCard>
          <h3 className="text-md font-heading font-bold mb-2">Aparência</h3>
          <p className="text-sm text-muted-foreground mb-4">
            Tema atual: <span className="font-medium">{tema === 'escuro' ? 'Escuro' : 'Claro'}</span>.
            Escolha entre o tema claro ou escuro (Dark Mode).
          </p>
          <button
            type="button"
            onClick={alternarTema}
            aria-pressed={tema === 'escuro'}
            className="inline-flex items-center gap-2 text-sm bg-secondary text-secondary-foreground px-4 py-2 rounded-lg font-medium hover:bg-secondary/80"
          >
            {tema === 'escuro' ? <Sun className="w-4 h-4" /> : <Moon className="w-4 h-4" />}
            Alternar Tema
          </button>
        </ElevatedCard>
      </div>
    </div>
  );
}
