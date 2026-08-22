import { useParams } from 'react-router-dom';
import { ElevatedCard } from '../../../components/m3/ElevatedCard';
import { Skeleton, EmptyState, ErrorState } from '../../../components/ui/StateViews';
import { useApp } from '../../../app/AppContext';
import { useCallback, useEffect, useState } from 'react';
import type { Versao } from '../../../api/types';

// Aba Versão (RF12): lista as versões do sistema e permite publicar uma nova
// versão ativa ou reverter para uma anterior — reaproveita a semântica de
// publicação/rollback por flag ativa já definida em 001 (RN04/RN05).
export function AbaVersao() {
  const { sistemaId = '' } = useParams<{ sistemaId: string }>();
  const { client } = useApp();
  const [estado, setEstado] = useState<
    | { fase: 'carregando' }
    | { fase: 'pronto'; versoes: Versao[] }
    | { fase: 'vazio' }
    | { fase: 'erro'; mensagem: string }
  >({ fase: 'carregando' });
  const [tentativa, setTentativa] = useState(0);
  const [emAcao, setEmAcao] = useState<string | null>(null);

  useEffect(() => {
    let vivo = true;
    setEstado({ fase: 'carregando' });
    (async () => {
      try {
        const versoes = await client.listarVersoes(sistemaId);
        if (!vivo) return;
        setEstado(versoes.length === 0 ? { fase: 'vazio' } : { fase: 'pronto', versoes });
      } catch (e) {
        if (vivo) setEstado({ fase: 'erro', mensagem: e instanceof Error ? e.message : String(e) });
      }
    })();
    return () => {
      vivo = false;
    };
  }, [client, sistemaId, tentativa]);

  const recarregar = useCallback(() => setTentativa((t) => t + 1), []);

  async function publicar(versaoId: string) {
    setEmAcao(versaoId);
    try {
      await client.publicarVersao(sistemaId, versaoId);
      recarregar();
    } finally {
      setEmAcao(null);
    }
  }

  async function reverter(versaoId: string) {
    setEmAcao(versaoId);
    try {
      await client.reverterVersao(sistemaId, versaoId);
      recarregar();
    } finally {
      setEmAcao(null);
    }
  }

  if (estado.fase === 'erro') {
    return <ErrorState mensagem="Não foi possível carregar as versões." onRepetir={recarregar} />;
  }
  if (estado.fase === 'carregando') return <Skeleton itens={3} />;
  if (estado.fase === 'vazio') {
    return <EmptyState titulo="Nenhuma versão publicada ainda" />;
  }

  const versaoAtiva = estado.versoes.find((v) => v.ativa);

  return (
    <ul className="flex flex-col gap-3">
      {estado.versoes.map((v) => {
        // Mais nova que a ativa (ou nenhuma ativa ainda) → promover (publicar).
        // Mais antiga que a ativa → restaurar como ativa (reverter) — RN05 de 001.
        const ehMaisNova = !versaoAtiva || v.numero > versaoAtiva.numero;
        return (
          <li key={v.id}>
            <ElevatedCard>
              <div className="flex items-center justify-between">
                <div>
                  <p className="font-heading font-semibold">
                    Versão {v.numero} {v.ativa && <span className="text-success">· ativa</span>}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    {new Date(v.criado_em).toLocaleString('pt-BR', { dateStyle: 'short', timeStyle: 'short' })}
                  </p>
                </div>
                {!v.ativa && (
                  <button
                    type="button"
                    disabled={emAcao === v.id}
                    onClick={() => void (ehMaisNova ? publicar(v.id) : reverter(v.id))}
                    className="text-sm bg-secondary text-secondary-foreground px-4 py-2 rounded-lg font-medium hover:bg-secondary/80 disabled:opacity-60"
                  >
                    {ehMaisNova ? 'Publicar' : 'Reverter'}
                  </button>
                )}
              </div>
            </ElevatedCard>
          </li>
        );
      })}
    </ul>
  );
}
