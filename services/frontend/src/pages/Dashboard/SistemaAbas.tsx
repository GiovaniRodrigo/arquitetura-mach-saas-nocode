import { NavLink, Outlet, useParams } from 'react-router-dom';
import { useApp } from '../../app/AppContext';
import { useSistemas } from '../../systems/useSistemas';

// Casca de navegação do sistema selecionado (RF09-RF12): 3 abas —
// Telas (canvas), Regras de Negócio e Versão — via rotas aninhadas.
export function SistemaAbas() {
  const { tenantId, sistemaId } = useParams<{ tenantId: string; sistemaId: string }>();
  const base = `/dashboard/clientes/${tenantId}/sistemas/${sistemaId}`;

  // Não existe endpoint de "obter 1 sistema" — reaproveita a mesma listagem
  // já usada em ClienteSistemas (useSistemas) e localiza o sistema corrente
  // pelo id da rota, só para exibir o nome no cabeçalho.
  const { client } = useApp();
  const { estado } = useSistemas(client, tenantId);
  const sistema = estado.fase === 'pronto' ? estado.sistemas.find((s) => s.id === sistemaId) : undefined;
  const nomeSistema = estado.fase === 'carregando' ? 'Carregando…' : (sistema?.nome ?? 'Sistema');

  const abaClasse = ({ isActive }: { isActive: boolean }) =>
    `px-4 py-2 text-sm font-medium rounded-full transition-colors ${
      isActive ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:bg-secondary'
    }`;

  return (
    <div className="flex flex-col gap-6 max-w-5xl mx-auto pb-8">
      <h2 className="text-2xl font-heading font-bold" aria-busy={estado.fase === 'carregando'}>
        {nomeSistema}
      </h2>
      <nav className="flex gap-2 border-b border-border pb-2">
        <NavLink to={`${base}/telas`} className={abaClasse}>
          Telas
        </NavLink>
        <NavLink to={`${base}/regras`} className={abaClasse}>
          Regras de Negócio
        </NavLink>
        <NavLink to={`${base}/versao`} className={abaClasse}>
          Versão
        </NavLink>
      </nav>
      <Outlet />
    </div>
  );
}
