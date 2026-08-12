// Miniatura da tela inteira (RF10/RF11, aba Regras de Negócio): renderiza a
// árvore na largura "desktop" de referência e encolhe via CSS transform
// scale até caber na largura disponível do painel — mostra a tela completa
// (não recortada), só reduzida, com o componente selecionado destacado pelo
// PreviewRenderer.

import { useEffect, useRef, useState } from 'react';
import type { Componente } from '../../../api/types';
import { PreviewRenderer } from './PreviewRenderer';
import { LARGURA_DEVICE } from './EditorTopbar';

const LARGURA_BASE = LARGURA_DEVICE.desktop;

export function PreviewMiniatura({ arvore, selecionado }: { arvore: Componente; selecionado?: string }) {
  const outerRef = useRef<HTMLDivElement>(null);
  const contentRef = useRef<HTMLDivElement>(null);
  const [escala, setEscala] = useState(1);
  const [alturaConteudo, setAlturaConteudo] = useState(0);

  useEffect(() => {
    const outer = outerRef.current;
    const content = contentRef.current;
    if (!outer || !content) return;

    function recalcular() {
      const largura = outer!.clientWidth;
      setEscala(largura > 0 ? largura / LARGURA_BASE : 1);
      setAlturaConteudo(content!.scrollHeight);
    }
    recalcular();

    // jsdom (testes) não implementa ResizeObserver — a chamada inicial acima
    // já cobre o cenário de teste (sem layout real, sem resize a observar).
    if (typeof ResizeObserver === 'undefined') return;
    const observer = new ResizeObserver(recalcular);
    observer.observe(outer);
    observer.observe(content);
    return () => observer.disconnect();
  }, [arvore]);

  return (
    <div ref={outerRef} className="w-full overflow-hidden" style={{ height: alturaConteudo * escala }}>
      <div
        ref={contentRef}
        className="bg-white text-black"
        style={{ width: LARGURA_BASE, transform: `scale(${escala})`, transformOrigin: 'top left' }}
      >
        <PreviewRenderer no={arvore} selecionado={selecionado} />
      </div>
    </div>
  );
}
