import { useParams } from 'react-router-dom';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { Search } from 'lucide-react';
import { Skeleton, ErrorState } from '../../../components/ui/StateViews';
import { useApp } from '../../../app/AppContext';
import { useComponentesFormulario, type ComponenteCampo } from '../../../systems/useComponentesFormulario';
import { useSessionStorageState } from '../../../systems/useSessionStorageState';
import { PreviewMiniatura } from '../editor/PreviewMiniatura';
import type { RegraNegocio, TipoRegraNegocio } from '../../../api/types';

// Aba Regras de Negócio (RF10/RF11, RN06): CRUD de validação de um componente
// (ex.: CPF numérico com 11 caracteres). Layout de dois painéis — a esquerda
// (mais larga, é onde o trabalho acontece), seletor da tela (cadastrada na
// aba Telas) + busca dos componentes de formulário dessa tela; a direita
// (mais estreita, só consulta de referência), miniatura read-only da tela
// inteira com destaque visual do componente selecionado, e o formulário de
// regra logo abaixo. Regra envolvendo múltiplos componentes (RF11) é
// modelagem de UI não trivial (seleção de N componentes + expressão) — fica
// como placeholder nesta fase (ver plan.md/Riscos).
export function AbaRegrasNegocio() {
  const { sistemaId = '' } = useParams<{ sistemaId: string }>();
  const { client } = useApp();
  const { estado: estadoComponentes, recarregar: recarregarComponentes } = useComponentesFormulario(client, sistemaId);

  const [estadoRegras, setEstadoRegras] = useState<
    | { fase: 'carregando' }
    | { fase: 'pronto'; regras: RegraNegocio[] }
    | { fase: 'vazio' }
    | { fase: 'erro'; mensagem: string }
  >({ fase: 'carregando' });
  const [tentativaRegras, setTentativaRegras] = useState(0);

  useEffect(() => {
    let vivo = true;
    setEstadoRegras({ fase: 'carregando' });
    (async () => {
      try {
        const regras = await client.listarRegrasNegocio(sistemaId);
        if (!vivo) return;
        setEstadoRegras(regras.length === 0 ? { fase: 'vazio' } : { fase: 'pronto', regras });
      } catch (e) {
        if (vivo) setEstadoRegras({ fase: 'erro', mensagem: e instanceof Error ? e.message : String(e) });
      }
    })();
    return () => {
      vivo = false;
    };
  }, [client, sistemaId, tentativaRegras]);

  const recarregarRegras = useCallback(() => setTentativaRegras((t) => t + 1), []);

  const regrasPorComponente = useMemo(() => {
    const mapa = new Map<string, RegraNegocio[]>();
    if (estadoRegras.fase !== 'pronto') return mapa;
    for (const regra of estadoRegras.regras) {
      for (const bi of regra.blind_indexes) {
        mapa.set(bi, [...(mapa.get(bi) ?? []), regra]);
      }
    }
    return mapa;
  }, [estadoRegras]);

  const [busca, setBusca] = useState('');
  const [selecionado, setSelecionado] = useState<string | null>(null);
  // sessionStorage pelo mesmo motivo do device/zoom da aba Telas: trocar de
  // aba Telas/Regras/Versão desmonta este componente via <Outlet/>.
  const [telaSelecionadaId, setTelaSelecionadaId] = useSessionStorageState<string | null>(
    `mach:sistema:${sistemaId}:regra-tela-selecionada`,
    null,
  );

  const telas = estadoComponentes.fase === 'pronto' ? estadoComponentes.telas : [];
  const telaAtualId = telas.some((t) => t.id === telaSelecionadaId) ? telaSelecionadaId : (telas[0]?.id ?? null);

  const itensDaTela =
    estadoComponentes.fase === 'pronto' ? estadoComponentes.itens.filter((i) => i.telaId === telaAtualId) : [];
  const itensFiltrados = itensDaTela.filter((item) => {
    const alvo = `${item.rotulo} ${item.blindIndex}`.toLowerCase();
    return alvo.includes(busca.trim().toLowerCase());
  });
  const componenteSelecionado: ComponenteCampo | undefined = itensDaTela.find((i) => i.blindIndex === selecionado);
  const telaAtual = telas.find((t) => t.id === telaAtualId);
  const arvoreDaTela =
    estadoComponentes.fase === 'pronto' && telaAtualId ? estadoComponentes.arvores[telaAtualId] : undefined;

  function selecionarTela(id: string) {
    setTelaSelecionadaId(id);
    setSelecionado(null);
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-[1fr_320px] gap-4 flex-1 min-h-0">
      <div className="bg-card border border-border rounded-2xl p-4 h-full flex flex-col gap-3 min-h-0">
        <h3 className="text-sm font-heading font-bold text-muted-foreground uppercase tracking-wide">Componentes</h3>
        {telas.length > 0 && (
          <div className="flex flex-col gap-1">
            <label htmlFor="regra-tela" className="text-xs font-medium text-muted-foreground">Tela</label>
            <select
              id="regra-tela"
              value={telaAtualId ?? ''}
              onChange={(e) => selecionarTela(e.target.value)}
              className="px-3 py-2 text-sm bg-background border border-border rounded-lg"
            >
              {telas.map((tela) => (
                <option key={tela.id} value={tela.id}>
                  {tela.nome}
                </option>
              ))}
            </select>
          </div>
        )}
        <div className="relative">
          <Search className="w-4 h-4 absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <input
            type="text"
            value={busca}
            onChange={(e) => setBusca(e.target.value)}
            placeholder="Buscar componente..."
            aria-label="Buscar componente"
            className="w-full pl-8 pr-3 py-2 text-sm bg-background border border-border rounded-lg"
          />
        </div>
        <div className="flex flex-col gap-1.5 overflow-y-auto scrollbar-app">
          {estadoComponentes.fase === 'carregando' ? (
            <Skeleton itens={3} />
          ) : estadoComponentes.fase === 'erro' ? (
            <ErrorState mensagem="Não foi possível carregar os componentes." onRepetir={recarregarComponentes} />
          ) : itensFiltrados.length === 0 ? (
            <p className="text-sm text-muted-foreground px-1">
              {itensDaTela.length === 0 ? 'Nenhum componente de formulário nesta tela.' : 'Nenhum componente encontrado.'}
            </p>
          ) : (
            itensFiltrados.map((item) => {
              const temRegra = (regrasPorComponente.get(item.blindIndex)?.length ?? 0) > 0;
              const ativo = item.blindIndex === selecionado;
              return (
                <button
                  key={item.blindIndex}
                  type="button"
                  onClick={() => setSelecionado(item.blindIndex)}
                  className={`flex items-center gap-2 text-left px-2.5 py-2 rounded-lg border text-sm ${
                    ativo ? 'border-primary bg-primary/10' : 'border-border bg-background hover:bg-secondary'
                  }`}
                >
                  <span
                    aria-hidden
                    className={`w-2 h-2 rounded-full shrink-0 ${temRegra ? 'bg-primary' : 'border border-muted-foreground'}`}
                  />
                  <span className="flex flex-col overflow-hidden">
                    <span className="font-medium truncate">{item.rotulo}</span>
                    <span className="text-xs text-muted-foreground truncate">{item.blindIndex}</span>
                  </span>
                </button>
              );
            })
          )}
        </div>
      </div>

      <div className="flex flex-col gap-4 min-h-0 overflow-y-auto scrollbar-app">
        <div className="bg-card border border-border rounded-2xl p-4">
          <h3 className="text-xs font-heading font-bold text-muted-foreground uppercase tracking-wide mb-3">
            {telaAtual ? `Prévia — ${telaAtual.nome}` : 'Prévia'}
          </h3>
          {!arvoreDaTela ? (
            <div className="border border-dashed border-border rounded-xl flex items-center justify-center p-8">
              <p className="text-sm text-muted-foreground text-center">Nenhuma tela cadastrada neste sistema.</p>
            </div>
          ) : (
            <div className="border border-border rounded-xl overflow-hidden">
              <PreviewMiniatura arvore={arvoreDaTela} selecionado={componenteSelecionado?.blindIndex} />
            </div>
          )}
        </div>

        {componenteSelecionado && (
          <FormularioRegra
            key={componenteSelecionado.blindIndex}
            sistemaId={sistemaId}
            componente={componenteSelecionado}
            regrasExistentes={regrasPorComponente.get(componenteSelecionado.blindIndex) ?? []}
            onSalva={recarregarRegras}
          />
        )}

        <div className="bg-card border border-border rounded-2xl p-4 opacity-70">
          <h3 className="text-md font-heading font-bold mb-2">Regra com múltiplos componentes</h3>
          <p className="text-sm text-muted-foreground">Em breve: validação cruzada entre vários componentes (RF11).</p>
        </div>
      </div>
    </div>
  );
}

function FormularioRegra({
  sistemaId,
  componente,
  regrasExistentes,
  onSalva,
}: {
  sistemaId: string;
  componente: ComponenteCampo;
  regrasExistentes: RegraNegocio[];
  onSalva: () => void;
}) {
  const { client } = useApp();
  const [tipo, setTipo] = useState<TipoRegraNegocio>('tamanho');
  const [tamanhoMax, setTamanhoMax] = useState('');
  const [salvando, setSalvando] = useState(false);

  async function salvar(e: React.FormEvent) {
    e.preventDefault();
    setSalvando(true);
    try {
      const parametros = tipo === 'tamanho' ? { max: Number(tamanhoMax) } : {};
      await client.criarRegraNegocio(sistemaId, { blind_indexes: [componente.blindIndex], tipo, parametros });
      setTamanhoMax('');
      onSalva();
    } finally {
      setSalvando(false);
    }
  }

  return (
    <div className="bg-card border border-border rounded-2xl p-4 flex flex-col gap-4">
      <div>
        <h3 className="text-md font-heading font-bold">Componente: {componente.rotulo}</h3>
        <p className="text-xs text-muted-foreground">{componente.blindIndex}</p>
      </div>

      {regrasExistentes.length > 0 && (
        <ul className="flex flex-col gap-1">
          {regrasExistentes.map((r) => (
            <li key={r.id} className="text-sm text-muted-foreground">
              Regra atual: <span className="font-medium text-foreground">{r.tipo}</span>
              {r.tipo === 'tamanho' && typeof r.parametros.max === 'number' ? ` (máx. ${r.parametros.max})` : ''}
            </li>
          ))}
        </ul>
      )}

      <form onSubmit={salvar} className="flex flex-col gap-4">
        <div className="flex flex-col gap-1">
          <label htmlFor="regra-tipo" className="text-sm font-medium">Tipo</label>
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
          disabled={salvando || (tipo === 'tamanho' && !tamanhoMax)}
          className="self-start text-sm bg-primary text-primary-foreground px-4 py-2 rounded-lg font-medium hover:bg-primary/90 disabled:opacity-60"
        >
          {salvando ? 'Salvando…' : 'Salvar'}
        </button>
      </form>
    </div>
  );
}
