// Card de status de um serviço monitorado (RF06): nome, indicador visual
// verde/vermelho, CPU/memória (metrics-server) e RPS/taxa de sucesso/latência
// (Prometheus do Linkerd) formatados de forma legível — nenhum dado vem de
// auto-relato do serviço, tudo é observado pelo mesh/cluster (spec 009). Um
// serviço "indisponivel" (RN01, aqui = pod não encontrado no cluster) some
// as métricas — nunca derruba a tela (RNF02, tratado no nível da página).

import { ElevatedCard } from '../components/m3/ElevatedCard';
import type { ServicoStatus } from '../api/types';

/** "15.0 MB" a partir de bytes. */
function formatarBytes(bytes: number): string {
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

/** "0.25" núcleos a partir de millicores, ou "180m" abaixo de 1 núcleo. */
function formatarCpu(millicores: number): string {
  return millicores >= 1000 ? `${(millicores / 1000).toFixed(2)} núcleos` : `${millicores}m`;
}

const NOMES_LEGIVEIS: Record<string, string> = {
  iam: 'IAM',
  design: 'Design',
  logic: 'Logic',
  deploy: 'Deploy',
  export: 'Export',
  gateway: 'Gateway',
  workers: 'Workers',
  collab: 'Collab',
};

export function CardServicoStatus({ servico }: { servico: ServicoStatus }) {
  const servindo = servico.status === 'servindo';
  const nomeLegivel = NOMES_LEGIVEIS[servico.nome] ?? servico.nome;

  return (
    <ElevatedCard>
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-md font-heading font-bold">{nomeLegivel}</h3>
        <span
          data-testid="indicador-status"
          data-status={servico.status}
          title={servindo ? 'Servindo' : 'Indisponível'}
          className={`inline-block w-3 h-3 rounded-full ${servindo ? 'bg-success' : 'bg-destructive'}`}
        />
      </div>

      {servindo ? (
        <dl className="flex flex-col gap-1 text-sm">
          <div className="flex justify-between">
            <dt className="text-muted-foreground">CPU</dt>
            <dd className="font-medium">{formatarCpu(servico.cpu_millicores)}</dd>
          </div>
          <div className="flex justify-between">
            <dt className="text-muted-foreground">Memória</dt>
            <dd className="font-medium">{formatarBytes(servico.memoria_bytes)}</dd>
          </div>
          <div className="flex justify-between">
            <dt className="text-muted-foreground">Requisições/s</dt>
            <dd className="font-medium">{servico.requisicoes_por_segundo.toFixed(2)}</dd>
          </div>
          <div className="flex justify-between">
            <dt className="text-muted-foreground">Taxa de sucesso</dt>
            <dd className="font-medium">{servico.taxa_sucesso_percent.toFixed(1)}%</dd>
          </div>
          <div className="flex justify-between">
            <dt className="text-muted-foreground">Latência p99</dt>
            <dd className="font-medium">{servico.latencia_p99_ms.toFixed(1)} ms</dd>
          </div>
        </dl>
      ) : (
        <p className="text-sm text-destructive">Indisponível — pod não encontrado no cluster.</p>
      )}
    </ElevatedCard>
  );
}
