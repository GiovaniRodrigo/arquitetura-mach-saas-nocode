// Painel de Layers (RF09, seção 16): árvore navegável dos componentes da
// tela — expandir/colapsar, selecionar, duplicar, excluir.

import { useState } from 'react';
import { ChevronRight, ChevronDown, Copy, Trash2 } from 'lucide-react';
import type { Componente } from '../../../api/types';

export interface PainelLayersProps {
  arvore: Componente | null;
  selecionado: string | null;
  onSelecionar: (blindIndex: string) => void;
  onRemover: (blindIndex: string) => void;
  onDuplicar: (blindIndex: string) => void;
}

export function PainelLayers({ arvore, selecionado, onSelecionar, onRemover, onDuplicar }: PainelLayersProps) {
  const [colapsados, setColapsados] = useState<Set<string>>(new Set());

  function alternarColapso(blindIndex: string) {
    setColapsados((atual) => {
      const novo = new Set(atual);
      if (novo.has(blindIndex)) novo.delete(blindIndex);
      else novo.add(blindIndex);
      return novo;
    });
  }

  if (!arvore || (arvore.componente_filhos?.length ?? 0) === 0) {
    return <p className="text-sm text-muted-foreground px-1">Nenhum componente ainda.</p>;
  }

  return (
    <ul role="tree" aria-label="Layers" className="flex flex-col gap-0.5 h-full overflow-y-auto scrollbar-app">
      {arvore.componente_filhos!.map((no) => (
        <LayerItem
          key={no.blind_index}
          no={no}
          nivel={0}
          colapsados={colapsados}
          onAlternarColapso={alternarColapso}
          selecionado={selecionado}
          onSelecionar={onSelecionar}
          onRemover={onRemover}
          onDuplicar={onDuplicar}
        />
      ))}
    </ul>
  );
}

function LayerItem({
  no,
  nivel,
  colapsados,
  onAlternarColapso,
  selecionado,
  onSelecionar,
  onRemover,
  onDuplicar,
}: {
  no: Componente;
  nivel: number;
  colapsados: Set<string>;
  onAlternarColapso: (blindIndex: string) => void;
  selecionado: string | null;
  onSelecionar: (blindIndex: string) => void;
  onRemover: (blindIndex: string) => void;
  onDuplicar: (blindIndex: string) => void;
}) {
  const filhos = no.componente_filhos ?? [];
  const temFilhos = filhos.length > 0;
  const colapsado = colapsados.has(no.blind_index);
  const ativo = selecionado === no.blind_index;

  return (
    <li role="treeitem" aria-selected={ativo}>
      <div
        className={`group flex items-center gap-1 rounded-lg px-1.5 py-1 cursor-pointer ${
          ativo ? 'bg-primary text-primary-foreground' : 'hover:bg-secondary text-foreground'
        }`}
        style={{ paddingLeft: 6 + nivel * 14 }}
        onClick={() => onSelecionar(no.blind_index)}
      >
        {temFilhos ? (
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation();
              onAlternarColapso(no.blind_index);
            }}
            aria-label={colapsado ? 'Expandir' : 'Colapsar'}
            className="shrink-0"
          >
            {colapsado ? <ChevronRight className="w-3.5 h-3.5" /> : <ChevronDown className="w-3.5 h-3.5" />}
          </button>
        ) : (
          <span className="w-3.5 h-3.5 shrink-0" />
        )}
        <span className="text-xs truncate flex-1">{no.tipo}</span>
        <button
          type="button"
          onClick={(e) => {
            e.stopPropagation();
            onDuplicar(no.blind_index);
          }}
          aria-label={`Duplicar ${no.tipo}`}
          className="opacity-0 group-hover:opacity-100 shrink-0"
        >
          <Copy className="w-3 h-3" />
        </button>
        <button
          type="button"
          onClick={(e) => {
            e.stopPropagation();
            onRemover(no.blind_index);
          }}
          aria-label={`Remover ${no.tipo}`}
          className="opacity-0 group-hover:opacity-100 shrink-0"
        >
          <Trash2 className="w-3 h-3" />
        </button>
      </div>

      {temFilhos && !colapsado && (
        <ul role="group">
          {filhos.map((filho) => (
            <LayerItem
              key={filho.blind_index}
              no={filho}
              nivel={nivel + 1}
              colapsados={colapsados}
              onAlternarColapso={onAlternarColapso}
              selecionado={selecionado}
              onSelecionar={onSelecionar}
              onRemover={onRemover}
              onDuplicar={onDuplicar}
            />
          ))}
        </ul>
      )}
    </li>
  );
}
