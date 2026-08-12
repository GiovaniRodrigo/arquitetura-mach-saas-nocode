// Overlay de "Visualização do site" (RF09): mostra a tela renderizada sem
// nenhuma chrome de edição, num frame do tamanho do device escolhido — o
// mais perto do resultado publicado que dá pra ver sem sair do editor.

import { useState } from 'react';
import { X, Monitor, Tablet, Smartphone } from 'lucide-react';
import type { Componente } from '../../../api/types';
import { PreviewRenderer } from './PreviewRenderer';
import { LARGURA_DEVICE, type Device } from './EditorTopbar';

export interface PreviewOverlayProps {
  nome: string;
  arvore: Componente;
  deviceInicial: Device;
  onFechar: () => void;
}

export function PreviewOverlay({ nome, arvore, deviceInicial, onFechar }: PreviewOverlayProps) {
  const [device, setDevice] = useState<Device>(deviceInicial);

  return (
    <div role="dialog" aria-label={`Visualização — ${nome}`} className="fixed inset-0 z-50 bg-background flex flex-col">
      <div className="flex items-center justify-between gap-4 border-b border-border px-4 py-2 shrink-0">
        <span className="text-sm font-medium text-foreground truncate">Visualização — {nome}</span>

        <div className="flex items-center gap-1" role="group" aria-label="Dispositivo">
          {(
            [
              ['desktop', Monitor],
              ['tablet', Tablet],
              ['mobile', Smartphone],
            ] as const
          ).map(([d, Icone]) => (
            <button
              key={d}
              type="button"
              onClick={() => setDevice(d)}
              aria-pressed={device === d}
              title={`${d[0].toUpperCase()}${d.slice(1)} (${LARGURA_DEVICE[d]}px)`}
              className={`p-1.5 rounded-lg ${device === d ? 'bg-primary text-primary-foreground' : 'hover:bg-secondary'}`}
            >
              <Icone className="w-4 h-4" />
            </button>
          ))}
        </div>

        <button
          type="button"
          onClick={onFechar}
          aria-label="Fechar visualização"
          className="p-1.5 rounded-lg hover:bg-secondary"
        >
          <X className="w-4 h-4" />
        </button>
      </div>

      <div className="flex-1 min-h-0 overflow-auto p-6 flex justify-center">
        <div
          className="mx-auto bg-white text-black min-h-full transition-[width]"
          style={{ width: LARGURA_DEVICE[device] }}
        >
          <PreviewRenderer no={arvore} />
        </div>
      </div>
    </div>
  );
}
