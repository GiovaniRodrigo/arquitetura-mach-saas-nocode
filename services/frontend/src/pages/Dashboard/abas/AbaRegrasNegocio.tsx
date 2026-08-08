import { useParams } from 'react-router-dom';
import { useCallback, useEffect, useState } from 'react';
import { ElevatedCard } from '../../../components/m3/ElevatedCard';
import { Skeleton, EmptyState, ErrorState } from '../../../components/ui/StateViews';
import { useApp } from '../../../app/AppContext';
import type { RegraNegocio, TipoRegraNegocio } from '../../../api/types';

// Aba Regras de Negócio (RF10/RF11, RN06): CRUD de validação de um componente
// (ex.: CPF numérico com 11 caracteres). Regra envolvendo múltiplos
// componentes (RF11) é modelagem de UI não trivial (seleção de N componentes
// + expressão) — fica como placeholder nesta fase (ver plan.md/Riscos).
export function AbaRegrasNegocio() {
  const { sistemaId = '' } = useParams<{ sistemaId: string }>();
  const { client } = useApp();
  const [estado, setEstado] = useState<
    | { fase: 'carregando' }
    | { fase: 'pronto'; regras: RegraNegocio[] }
    | { fase: 'vazio' }
    | { fase: 'erro'; mensagem: string }
  >({ fase: 'carregando' });
  const [tentativa, setTentativa] = useState(0);

  const [blindIndex, setBlindIndex] = useState('');
  const [tipo, setTipo] = useState<TipoRegraNegocio>('tamanho');
  const [tamanhoMax, setTamanhoMax] = useState('');
  const [criando, setCriando] = useState(false);

  useEffect(() => {
    let vivo = true;
    setEstado({ fase: 'carregando' });
    (async () => {
      try {
        const regras = await client.listarRegrasNegocio(sistemaId);
        if (!vivo) return;
        setEstado(regras.length === 0 ? { fase: 'vazio' } : { fase: 'pronto', regras });
      } catch (e) {
        if (vivo) setEstado({ fase: 'erro', mensagem: e instanceof Error ? e.message : String(e) });
      }
    })();
    return () => {
      vivo = false;
    };
  }, [client, sistemaId, tentativa]);

  const recarregar = useCallback(() => setTentativa((t) => t + 1), []);

  async function adicionarRegra(e: React.FormEvent) {
    e.preventDefault();
    if (!blindIndex) return;
    setCriando(true);
    try {
      const parametros = tipo === 'tamanho' ? { max: Number(tamanhoMax) } : {};
      await client.criarRegraNegocio(sistemaId, { blind_indexes: [blindIndex], tipo, parametros });
      setBlindIndex('');
      setTamanhoMax('');
      recarregar();
    } finally {
      setCriando(false);
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <ElevatedCard>
        <h3 className="text-md font-heading font-bold mb-4">Nova regra (componente único)</h3>
        <form onSubmit={adicionarRegra} className="flex flex-col gap-4">
          <div className="flex flex-col gap-1">
            <label htmlFor="regra-bi" className="text-sm font-medium">Componente (blind_index)</label>
            <input
              id="regra-bi"
              type="text"
              value={blindIndex}
              onChange={(e) => setBlindIndex(e.target.value)}
              placeholder="bi-cpf"
              className="px-3 py-2 text-sm bg-background border border-border rounded-lg"
            />
          </div>
          <div className="flex flex-col gap-1">
            <label htmlFor="regra-tipo" className="text-sm font-medium">Tipo de validação</label>
            <select
              id="regra-tipo"
              value={tipo}
              onChange={(e) => setTipo(e.target.value as TipoRegraNegocio)}
              className="px-3 py-2 text-sm bg-background border border-border rounded-lg"
            >
              <option value="tamanho">Tamanho</option>
              <option value="regex">Padrão (regex)</option>
              <option value="obrigatorio">Obrigatório</option>
            </select>
          </div>
          {tipo === 'tamanho' && (
            <div className="flex flex-col gap-1">
              <label htmlFor="regra-tamanho-max" className="text-sm font-medium">Tamanho máximo</label>
              <input
                id="regra-tamanho-max"
                type="number"
                value={tamanhoMax}
                onChange={(e) => setTamanhoMax(e.target.value)}
                className="px-3 py-2 text-sm bg-background border border-border rounded-lg"
              />
            </div>
          )}
          <button
            type="submit"
            disabled={criando || !blindIndex}
            className="self-start text-sm bg-primary text-primary-foreground px-4 py-2 rounded-lg font-medium hover:bg-primary/90 disabled:opacity-60"
          >
            {criando ? 'Adicionando…' : 'Adicionar regra'}
          </button>
        </form>
      </ElevatedCard>

      <ElevatedCard>
        <h3 className="text-md font-heading font-bold mb-4">Regras existentes</h3>
        {estado.fase === 'erro' ? (
          <ErrorState mensagem="Não foi possível carregar as regras de negócio." onRepetir={recarregar} />
        ) : estado.fase === 'carregando' ? (
          <Skeleton itens={2} />
        ) : estado.fase === 'vazio' ? (
          <EmptyState titulo="Nenhuma regra cadastrada ainda" />
        ) : (
          <ul className="flex flex-col gap-2">
            {estado.regras.map((r) => (
              <li key={r.id} className="text-sm text-muted-foreground">
                <span className="font-medium text-foreground">{r.blind_indexes.join(', ')}</span> — {r.tipo}
              </li>
            ))}
          </ul>
        )}
      </ElevatedCard>

      <ElevatedCard className="opacity-70">
        <h3 className="text-md font-heading font-bold mb-2">Regra com múltiplos componentes</h3>
        <p className="text-sm text-muted-foreground">Em breve: validação cruzada entre vários componentes (RF11).</p>
      </ElevatedCard>
    </div>
  );
}
