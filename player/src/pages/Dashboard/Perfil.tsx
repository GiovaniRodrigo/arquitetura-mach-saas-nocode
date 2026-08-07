import { TonalCard } from '../../components/m3/TonalCard';
import { ElevatedCard } from '../../components/m3/ElevatedCard';
import { useApp } from '../../app/AppContext';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft } from 'lucide-react';

// Placeholder de edição de perfil (RF04/C5): o formulário completo está fora do
// escopo desta demanda; a tela exibe a identidade atual (claims do JWT) e um aviso.
export function Perfil() {
  const { usuario } = useApp();
  const navigate = useNavigate();

  return (
    <div className="flex flex-col gap-6 max-w-2xl mx-auto pb-8">
      <button
        type="button"
        onClick={() => navigate('/dashboard')}
        className="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground w-fit"
      >
        <ArrowLeft className="w-4 h-4" /> Voltar ao Dashboard
      </button>

      <TonalCard className="bg-secondary text-secondary-foreground border-none">
        <h2 className="text-2xl font-heading font-bold mb-2">Perfil do Usuário</h2>
        <p className="text-muted-foreground text-sm font-medium">
          Suas informações de identidade nesta sessão.
        </p>
      </TonalCard>

      <ElevatedCard>
        <div className="flex items-center gap-4 mb-4">
          <div className="w-14 h-14 rounded-full bg-primary/20 flex items-center justify-center text-primary font-bold text-lg">
            {usuario.iniciais}
          </div>
          <div>
            <p className="font-heading font-semibold text-foreground">{usuario.nome ?? 'Usuário'}</p>
            {usuario.email && <p className="text-sm text-muted-foreground">{usuario.email}</p>}
          </div>
        </div>
        <p className="text-sm text-muted-foreground">
          A edição dos dados de perfil estará disponível em breve.
        </p>
      </ElevatedCard>
    </div>
  );
}
