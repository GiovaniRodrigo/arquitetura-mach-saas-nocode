// Paleta de componentes (RF09), categorizada. Clicar adiciona no container
// selecionado (ou na raiz); arrastar para o canvas solta na posição exata
// (before/inside/after — ver Canvas.tsx). O dataTransfer usa o prefixo
// "novo:" para o Canvas distinguir "criar componente" de "mover existente".

import { CATEGORIAS, REGISTRO_COMPONENTES } from '../../../systems/componentRegistry';

export interface PainelComponentesProps {
  onAdicionar: (tipo: string) => void;
}

export function PainelComponentes({ onAdicionar }: PainelComponentesProps) {
  return (
    <div className="flex flex-col gap-4 h-full overflow-y-auto scrollbar-app">
      {CATEGORIAS.map((cat) => {
        const itens = REGISTRO_COMPONENTES.filter((r) => r.categoria === cat.id);
        if (itens.length === 0) return null;
        return (
          <div key={cat.id} className="flex flex-col gap-1.5">
            <h4 className="text-xs font-heading font-bold text-muted-foreground uppercase tracking-wide px-1">
              {cat.rotulo}
            </h4>
            <div className="grid grid-cols-2 gap-1.5">
              {itens.map((item) => {
                const Icone = item.icone;
                return (
                  <button
                    key={item.tipo}
                    type="button"
                    title={item.rotulo}
                    draggable
                    onDragStart={(e) => {
                      e.dataTransfer.setData('text/plain', `novo:${item.tipo}`);
                      e.dataTransfer.effectAllowed = 'copy';
                    }}
                    onClick={() => onAdicionar(item.tipo)}
                    className="flex items-center gap-1.5 text-xs text-left px-2.5 py-2 rounded-lg border border-border bg-background hover:bg-secondary hover:border-primary/40 cursor-grab active:cursor-grabbing"
                  >
                    <Icone className="w-3.5 h-3.5 shrink-0 text-muted-foreground" />
                    <span className="truncate">{item.rotulo}</span>
                  </button>
                );
              })}
            </div>
          </div>
        );
      })}
    </div>
  );
}
