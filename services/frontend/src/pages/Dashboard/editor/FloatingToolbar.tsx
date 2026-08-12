// Barra de formatação flutuante (RF09) — estilo "proximity toolbar" (Notion):
// aparece perto da seleção de texto dentro de um componente em edição
// inline, com Negrito/Itálico/Sublinhado aplicados só ao trecho selecionado.
// Renderizada via portal em document.body para não ser cortada pelo
// overflow/scroll dos painéis do editor.

import { createPortal } from 'react-dom';
import { Bold, Italic, Underline } from 'lucide-react';

export interface FloatingToolbarProps {
  /** Posição (viewport) da seleção de texto — null esconde a barra. */
  rect: DOMRect | null;
  onBold: () => void;
  onItalic: () => void;
  onUnderline: () => void;
}

export function FloatingToolbar({ rect, onBold, onItalic, onUnderline }: FloatingToolbarProps) {
  if (!rect) return null;

  const top = rect.top - 44;
  const left = rect.left + rect.width / 2;

  return createPortal(
    <div
      role="toolbar"
      aria-label="Formatação de texto"
      className="fixed z-[100] flex items-center gap-0.5 bg-popover border border-border rounded-lg shadow-lg p-1 -translate-x-1/2"
      style={{ top: Math.max(8, top), left }}
      // mousedown (não click) + preventDefault: evita que o foco/seleção do
      // contentEditable seja perdido antes do comando de formatação rodar.
      onMouseDown={(e) => e.preventDefault()}
    >
      <button
        type="button"
        aria-label="Negrito"
        title="Negrito (Ctrl+B)"
        onClick={onBold}
        className="p-1.5 rounded-md hover:bg-secondary text-foreground"
      >
        <Bold className="w-3.5 h-3.5" />
      </button>
      <button
        type="button"
        aria-label="Itálico"
        title="Itálico (Ctrl+I)"
        onClick={onItalic}
        className="p-1.5 rounded-md hover:bg-secondary text-foreground"
      >
        <Italic className="w-3.5 h-3.5" />
      </button>
      <button
        type="button"
        aria-label="Sublinhado"
        title="Sublinhado (Ctrl+U)"
        onClick={onUnderline}
        className="p-1.5 rounded-md hover:bg-secondary text-foreground"
      >
        <Underline className="w-3.5 h-3.5" />
      </button>
    </div>,
    document.body,
  );
}
