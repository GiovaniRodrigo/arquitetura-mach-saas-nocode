// Topbar do editor (RF09): breadcrumb com troca de tela, undo/redo, seletor
// de device, zoom e status de salvamento. Concentra as operações globais do
// editor para tirá-las do canvas/paineis.

import { useState } from 'react';
import { ChevronDown, Undo2, Redo2, Monitor, Tablet, Smartphone, Plus, Check } from 'lucide-react';
import type { DesignResumo } from '../../../api/types';
import type { StatusSalvamento } from '../../../systems/useCanvasDesign';

export type Device = 'desktop' | 'tablet' | 'mobile';

export const LARGURA_DEVICE: Record<Device, number> = { desktop: 1440, tablet: 768, mobile: 375 };

/** Os 3 modos do editor são mutuamente exclusivos (um seletor de 3 posições,
 * não toggles independentes): Edição é o estado padrão; Foco esconde os
 * painéis laterais; Visualização abre o preview sem chrome de edição. */
export type ModoEditor = 'edicao' | 'visualizacao' | 'foco';

const MODOS: { valor: ModoEditor; rotulo: string; atalho?: string }[] = [
  { valor: 'edicao', rotulo: 'Edição' },
  { valor: 'visualizacao', rotulo: 'Visualização' },
  { valor: 'foco', rotulo: 'Foco', atalho: 'F' },
];

const NIVEIS_ZOOM = [0.25, 0.5, 0.75, 1, 1.25, 1.5];

export interface EditorTopbarProps {
  telas: DesignResumo[];
  telaSelecionadaId: string | null;
  onSelecionarTela: (id: string) => void;
  onNovaTela: () => void;
  podeDesfazer: boolean;
  podeRefazer: boolean;
  onDesfazer: () => void;
  onRefazer: () => void;
  device: Device;
  onDeviceChange: (device: Device) => void;
  zoom: number;
  onZoomChange: (zoom: number) => void;
  statusSalvamento?: StatusSalvamento;
  modo?: ModoEditor;
  onModoChange?: (modo: ModoEditor) => void;
}

export function EditorTopbar({
  telas,
  telaSelecionadaId,
  onSelecionarTela,
  onNovaTela,
  podeDesfazer,
  podeRefazer,
  onDesfazer,
  onRefazer,
  device,
  onDeviceChange,
  zoom,
  onZoomChange,
  statusSalvamento,
  modo,
  onModoChange,
}: EditorTopbarProps) {
  const [trocaAberta, setTrocaAberta] = useState(false);
  const telaAtual = telas.find((t) => t.id === telaSelecionadaId);

  return (
    <div className="flex items-center justify-between gap-4 bg-card border border-border rounded-2xl px-4 py-2">
      <div className="relative flex items-center gap-2 min-w-0">
        <button
          type="button"
          onClick={() => setTrocaAberta((a) => !a)}
          aria-expanded={trocaAberta}
          aria-haspopup="listbox"
          className="flex items-center gap-1 text-sm font-medium text-foreground hover:bg-secondary px-2 py-1 rounded-lg truncate"
        >
          {telaAtual?.nome ?? 'Selecione uma tela'}
          <ChevronDown className="w-3.5 h-3.5 shrink-0" />
        </button>

        {trocaAberta && (
          <>
            <button
              type="button"
              aria-label="Fechar seletor de telas"
              className="fixed inset-0 z-40 cursor-default"
              onClick={() => setTrocaAberta(false)}
            />
            <div
              role="listbox"
              aria-label="Telas"
              className="absolute top-full left-0 mt-1 z-50 w-56 bg-popover border border-border rounded-xl shadow-lg p-1 flex flex-col gap-0.5"
            >
              {telas.map((t) => (
                <button
                  key={t.id}
                  type="button"
                  role="option"
                  aria-selected={t.id === telaSelecionadaId}
                  onClick={() => {
                    onSelecionarTela(t.id);
                    setTrocaAberta(false);
                  }}
                  className={`text-left text-sm px-2.5 py-1.5 rounded-lg ${
                    t.id === telaSelecionadaId ? 'bg-primary text-primary-foreground' : 'hover:bg-secondary'
                  }`}
                >
                  {t.nome}
                </button>
              ))}
              <button
                type="button"
                onClick={() => {
                  setTrocaAberta(false);
                  onNovaTela();
                }}
                className="flex items-center gap-1 text-left text-sm px-2.5 py-1.5 rounded-lg text-primary hover:bg-secondary"
              >
                <Plus className="w-3.5 h-3.5" /> Nova tela
              </button>
            </div>
          </>
        )}
      </div>

      <div className="flex items-center gap-1">
        <button
          type="button"
          onClick={onDesfazer}
          disabled={!podeDesfazer}
          aria-label="Desfazer"
          title="Desfazer (Ctrl+Z)"
          className="p-1.5 rounded-lg hover:bg-secondary disabled:opacity-40 disabled:pointer-events-none"
        >
          <Undo2 className="w-4 h-4" />
        </button>
        <button
          type="button"
          onClick={onRefazer}
          disabled={!podeRefazer}
          aria-label="Refazer"
          title="Refazer (Ctrl+Shift+Z)"
          className="p-1.5 rounded-lg hover:bg-secondary disabled:opacity-40 disabled:pointer-events-none"
        >
          <Redo2 className="w-4 h-4" />
        </button>
      </div>

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
            onClick={() => onDeviceChange(d)}
            aria-pressed={device === d}
            title={`${d[0].toUpperCase()}${d.slice(1)} (${LARGURA_DEVICE[d]}px)`}
            className={`p-1.5 rounded-lg ${device === d ? 'bg-primary text-primary-foreground' : 'hover:bg-secondary'}`}
          >
            <Icone className="w-4 h-4" />
          </button>
        ))}
      </div>

      {modo && onModoChange && <SeletorModo modo={modo} onModoChange={onModoChange} />}

      <label className="flex items-center gap-1 text-xs text-muted-foreground">
        Zoom
        <select
          value={zoom}
          onChange={(e) => onZoomChange(Number(e.target.value))}
          className="bg-transparent border border-border rounded-lg px-1.5 py-1 text-xs text-foreground"
        >
          {NIVEIS_ZOOM.map((n) => (
            <option key={n} value={n}>
              {Math.round(n * 100)}%
            </option>
          ))}
        </select>
      </label>

      <span className="flex items-center justify-end gap-1 text-xs text-muted-foreground min-w-16 text-right">
        {statusSalvamento === 'salvando' && 'Salvando…'}
        {statusSalvamento === 'salvo' && (
          <>
            <Check className="w-3.5 h-3.5" /> Salvo
          </>
        )}
      </span>
    </div>
  );
}

/** Seletor de modo do editor: pill segmentado com Edição/Visualização/Foco
 * lado a lado num único contêiner — mutuamente exclusivos (clicar um
 * desativa os outros dois), o segmento ativo se destaca com fundo sólido. */
function SeletorModo({ modo, onModoChange }: { modo: ModoEditor; onModoChange: (modo: ModoEditor) => void }) {
  return (
    <div
      className="flex items-center gap-0.5 rounded-lg border border-border bg-background p-0.5"
      role="group"
      aria-label="Modo do editor"
    >
      {MODOS.map(({ valor, rotulo, atalho }) => {
        const ativo = modo === valor;
        return (
          <button
            key={valor}
            type="button"
            aria-pressed={ativo}
            onClick={() => onModoChange(valor)}
            title={atalho ? `${rotulo} (${atalho})` : rotulo}
            className={`px-2.5 py-1 rounded-md text-xs font-medium transition-colors ${
              ativo ? 'bg-secondary text-foreground' : 'text-muted-foreground hover:bg-secondary/50'
            }`}
          >
            {rotulo}
          </button>
        );
      })}
    </div>
  );
}
