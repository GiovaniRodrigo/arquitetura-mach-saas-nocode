// Painel esquerdo do editor (RF09): abas Componentes | Layers.

import { useState } from 'react';
import { Layers as LayersIcon } from 'lucide-react';
import type { Componente } from '../../../api/types';
import { PainelComponentes } from './PainelComponentes';
import { PainelLayers } from './PainelLayers';

export interface PainelEsquerdoProps {
  arvore: Componente | null;
  selecionado: string | null;
  onSelecionar: (blindIndex: string) => void;
  onRemover: (blindIndex: string) => void;
  onDuplicar: (blindIndex: string) => void;
  onAdicionar: (tipo: string) => void;
}

type Aba = 'componentes' | 'layers';

export function PainelEsquerdo({
  arvore,
  selecionado,
  onSelecionar,
  onRemover,
  onDuplicar,
  onAdicionar,
}: PainelEsquerdoProps) {
  const [aba, setAba] = useState<Aba>('componentes');

  return (
    <div className="bg-card border border-border rounded-2xl p-3 flex flex-col gap-3 h-full min-h-0">
      <div className="flex gap-1 shrink-0" role="tablist">
        <button
          type="button"
          role="tab"
          aria-selected={aba === 'componentes'}
          onClick={() => setAba('componentes')}
          className={`flex-1 text-xs font-medium px-2 py-1.5 rounded-lg ${
            aba === 'componentes' ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:bg-secondary'
          }`}
        >
          Componentes
        </button>
        <button
          type="button"
          role="tab"
          aria-selected={aba === 'layers'}
          onClick={() => setAba('layers')}
          className={`flex-1 flex items-center justify-center gap-1 text-xs font-medium px-2 py-1.5 rounded-lg ${
            aba === 'layers' ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:bg-secondary'
          }`}
        >
          <LayersIcon className="w-3.5 h-3.5" />
          Layers
        </button>
      </div>

      <div className="flex-1 min-h-0">
        {aba === 'componentes' ? (
          <PainelComponentes onAdicionar={onAdicionar} />
        ) : (
          <PainelLayers
            arvore={arvore}
            selecionado={selecionado}
            onSelecionar={onSelecionar}
            onRemover={onRemover}
            onDuplicar={onDuplicar}
          />
        )}
      </div>
    </div>
  );
}
