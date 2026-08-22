// Tela "Monitor de Recursos" (RF06-RF08): saúde/infra dos 8 serviços da
// plataforma, lida do Kubernetes/Linkerd via GET /api/v1/monitor/recursos
// (spec 009). RNF02: erro do endpoint inteiro vira um estado único de tela,
// nunca 8 cards quebrados — uma entrada individual "indisponivel" (RN01)
// dentro de uma resposta 200 é normal e faz parte do estado 'pronto'.

import { RefreshCw } from 'lucide-react';
import { TonalCard } from '../../components/m3/TonalCard';
import { useApp } from '../../app/AppContext';
import { useRecursos } from '../../dashboard/useRecursos';
import { CardServicoStatus } from '../../dashboard/CardServicoStatus';

export function Monitor() {
  const { client } = useApp();
  const { estado, recarregar } = useRecursos(client);

  return (
    <div className="flex flex-col gap-6 max-w-5xl mx-auto pb-8">
      <TonalCard className="bg-secondary text-secondary-foreground border-none">
        <div className="flex items-center justify-between gap-4">
          <div>
            <h2 className="text-2xl font-heading font-bold mb-2">Monitor de Recursos</h2>
            <p className="text-muted-foreground text-sm font-medium">
              Saúde, CPU, memória e tráfego dos serviços da plataforma, em tempo real.
            </p>
          </div>
          <button
            type="button"
            onClick={recarregar}
            className="inline-flex items-center gap-2 text-sm bg-secondary text-secondary-foreground px-4 py-2 rounded-lg font-medium hover:bg-secondary/80 border border-border shrink-0"
          >
            <RefreshCw className="w-4 h-4" />
            Atualizar
          </button>
        </div>
      </TonalCard>

      {estado.fase === 'carregando' && <p className="text-sm text-muted-foreground">Carregando…</p>}

      {estado.fase === 'erro' && (
        <div role="alert" className="text-sm text-destructive">
          <p>Não foi possível consultar os recursos da plataforma: {estado.mensagem}</p>
        </div>
      )}

      {estado.fase === 'pronto' && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {estado.recursos.servicos.map((servico) => (
            <CardServicoStatus key={servico.nome} servico={servico} />
          ))}
        </div>
      )}
    </div>
  );
}
