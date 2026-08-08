import { useMemo, useState } from 'react';
import { TonalCard } from '../../components/m3/TonalCard';
import { ElevatedCard } from '../../components/m3/ElevatedCard';
import { Search } from 'lucide-react';
import { artigos, buscarArtigos, type ArtigoDocumentacao } from '../../ajuda/artigos';

function agruparPorCategoria(itens: ArtigoDocumentacao[]): Map<string, ArtigoDocumentacao[]> {
  const grupos = new Map<string, ArtigoDocumentacao[]>();
  for (const item of itens) {
    const lista = grupos.get(item.categoria) ?? [];
    lista.push(item);
    grupos.set(item.categoria, lista);
  }
  return grupos;
}

export function Ajuda() {
  const [busca, setBusca] = useState('');
  const resultados = useMemo(() => buscarArtigos(busca), [busca]);
  const porCategoria = useMemo(() => agruparPorCategoria(resultados), [resultados]);

  return (
    <div className="flex flex-col gap-6 max-w-5xl mx-auto pb-8">
      <TonalCard className="bg-secondary text-secondary-foreground border-none">
        <h2 className="text-2xl font-heading font-bold mb-2">Ajuda</h2>
        <p className="text-muted-foreground text-sm font-medium">
          Documentação geral da plataforma — {artigos.length} artigos disponíveis.
        </p>
      </TonalCard>

      <div className="flex items-center gap-2 px-4 py-2.5 bg-card border border-border rounded-full">
        <Search className="w-4 h-4 text-muted-foreground shrink-0" />
        <input
          type="search"
          role="searchbox"
          aria-label="Buscar na documentação"
          value={busca}
          onChange={(e) => setBusca(e.target.value)}
          placeholder="Buscar na documentação…"
          className="flex-1 bg-transparent outline-none text-sm placeholder:text-muted-foreground"
        />
      </div>

      {resultados.length === 0 ? (
        <p className="text-sm text-muted-foreground text-center py-8">
          Nenhum artigo encontrado para "{busca}".
        </p>
      ) : (
        <div className="flex flex-col gap-6">
          {Array.from(porCategoria.entries()).map(([categoria, itens]) => (
            <section key={categoria} className="flex flex-col gap-3">
              <h3 className="text-sm font-heading font-bold text-muted-foreground uppercase tracking-wide">
                {categoria}
              </h3>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {itens.map((artigo) => (
                  <ElevatedCard key={artigo.id}>
                    <article aria-labelledby={`artigo-${artigo.id}`}>
                      <h4 id={`artigo-${artigo.id}`} className="font-heading font-semibold mb-1">
                        {artigo.titulo}
                      </h4>
                      <p className="text-sm text-muted-foreground">{artigo.conteudo}</p>
                    </article>
                  </ElevatedCard>
                ))}
              </div>
            </section>
          ))}
        </div>
      )}
    </div>
  );
}
