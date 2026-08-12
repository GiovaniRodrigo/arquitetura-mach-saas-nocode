// Renderização read-only da árvore para "Visualização do site" (RF09): o
// mesmo Componente + Estilos do editor, mas sem nenhuma affordance de edição
// (sem contorno de seleção, sem handles de resize, sem placeholder de drop) —
// o mais perto do resultado publicado que dá pra mostrar sem um pipeline de
// publish separado.

import { useEffect, useState } from 'react';
import type { Componente } from '../../../api/types';
import type { Estilos } from '../../../systems/componentRegistry';
import { estilosParaCss } from '../../../systems/estilosCss';
import { iconePorNome } from '../../../systems/iconePicker';
import { paraEmbedUrl } from '../../../systems/videoEmbed';
import { sanitizarHtml } from '../../../systems/sanitizeHtml';

export function PreviewRenderer({ no, selecionado }: { no: Componente; selecionado?: string }) {
  const estilos = (no.propriedades?.estilos as Estilos | undefined) ?? {};
  const estiloBase = estilosParaCss(estilos);
  // Destaque de "componente selecionado" (aba Regras de Negócio, RF10/RF11):
  // outline por fora do box em vez de border, pra não deslocar o layout do
  // preview ao selecionar/desselecionar um componente.
  const estilo =
    no.blind_index === selecionado
      ? { ...estiloBase, outline: '2px solid #6366f1', outlineOffset: '2px' }
      : estiloBase;
  const texto = typeof no.propriedades?.texto === 'string' ? no.propriedades.texto : undefined;
  // Componentes com temTexto editam via rich text inline no canvas (negrito/
  // itálico/sublinhado por trecho — ver TextoEditavel em Canvas.tsx), então o
  // texto pode conter HTML simples; sanitizado de novo aqui por defesa em
  // profundidade (ex.: propriedades tocadas via JSON bruto no Avançado).
  const textoHtml = texto !== undefined ? sanitizarHtml(texto) : undefined;
  const filhos = (no.componente_filhos ?? []).map((f) => (
    <PreviewRenderer key={f.blind_index} no={f} selecionado={selecionado} />
  ));

  switch (no.tipo) {
    case 'header':
      return <header style={estilo}>{filhos}</header>;
    case 'footer':
      return <footer style={estilo}>{filhos}</footer>;
    case 'sidebar':
    case 'rightbar':
      return <aside style={estilo}>{filhos}</aside>;
    case 'main':
      return <main style={estilo}>{filhos}</main>;
    case 'section':
      return <section style={estilo}>{filhos}</section>;
    case 'heading':
      return <h1 style={estilo} dangerouslySetInnerHTML={{ __html: textoHtml ?? '' }} />;
    case 'paragrafo':
      return <p style={estilo} dangerouslySetInnerHTML={{ __html: textoHtml ?? '' }} />;
    case 'botao':
      return <button type="button" style={estilo} dangerouslySetInnerHTML={{ __html: textoHtml ?? '' }} />;
    case 'link':
      return (
        <a
          href="#"
          style={estilo}
          onClick={(e) => e.preventDefault()}
          dangerouslySetInnerHTML={{ __html: textoHtml ?? '' }}
        />
      );
    case 'input':
      return <input style={estilo} placeholder={texto} readOnly />;
    case 'select':
      return (
        <select style={estilo} disabled>
          {texto && <option>{texto}</option>}
        </select>
      );
    case 'checkbox':
      return (
        <label style={{ ...estilo, display: 'flex', alignItems: 'center', gap: 6 }}>
          <input type="checkbox" readOnly />
          <span dangerouslySetInnerHTML={{ __html: textoHtml ?? '' }} />
        </label>
      );
    case 'radio':
      return (
        <label style={{ ...estilo, display: 'flex', alignItems: 'center', gap: 6 }}>
          <input type="radio" readOnly />
          <span dangerouslySetInnerHTML={{ __html: textoHtml ?? '' }} />
        </label>
      );
    case 'textarea':
      return <textarea style={estilo} placeholder={texto} readOnly />;
    case 'imagem':
      return typeof no.propriedades?.src === 'string' && no.propriedades.src ? (
        <img
          src={no.propriedades.src}
          alt={typeof no.propriedades?.alt === 'string' ? no.propriedades.alt : ''}
          style={{ ...estilo, objectFit: 'cover' }}
        />
      ) : null;
    case 'carrossel':
      return <CarrosselPublicado propriedades={no.propriedades} estilo={estilo} />;
    case 'menu':
      return <MenuPublicado propriedades={no.propriedades} estilo={estilo} />;
    case 'accordion':
      return <AccordionPublicado propriedades={no.propriedades} estilo={estilo} />;
    case 'badge':
      return <span style={estilo} dangerouslySetInnerHTML={{ __html: textoHtml ?? '' }} />;
    case 'video':
      return <VideoPublicado propriedades={no.propriedades} estilo={estilo} />;
    case 'tabs':
      return <TabsPublicado propriedades={no.propriedades} estilo={estilo} />;
    case 'avaliacao': {
      const valor = typeof no.propriedades?.valor === 'number' ? no.propriedades.valor : 0;
      return (
        <div style={estilo}>
          {[1, 2, 3, 4, 5].map((n) => (
            <span key={n}>{n <= valor ? '★' : '☆'}</span>
          ))}
        </div>
      );
    }
    case 'icone': {
      const Icone = iconePorNome(typeof no.propriedades?.icone === 'string' ? no.propriedades.icone : undefined);
      return <Icone style={estilo} />;
    }
    case 'progresso': {
      const valor =
        typeof no.propriedades?.valor === 'number' ? Math.min(100, Math.max(0, no.propriedades.valor)) : 50;
      return (
        <div style={{ ...estilo, backgroundColor: '#e5e7eb', overflow: 'hidden' }}>
          <div style={{ width: `${valor}%`, height: '100%', backgroundColor: estilos.fundoCor ?? '#6366f1' }} />
        </div>
      );
    }
    case 'avatar':
      return typeof no.propriedades?.src === 'string' && no.propriedades.src ? (
        <img
          src={no.propriedades.src}
          alt={typeof no.propriedades?.alt === 'string' ? no.propriedades.alt : ''}
          style={{ ...estilo, objectFit: 'cover' }}
        />
      ) : (
        <div style={estilo} dangerouslySetInnerHTML={{ __html: textoHtml ?? '' }} />
      );
    case 'breadcrumb': {
      const itens = Array.isArray(no.propriedades?.itens) ? (no.propriedades.itens as ItemMenu[]) : [];
      if (itens.length === 0) return null;
      return (
        <nav style={{ ...estilo, display: 'flex', alignItems: 'center' }}>
          {itens.map((item, i) => (
            <span key={i} style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
              {i > 0 && <span>›</span>}
              {i === itens.length - 1 ? <span>{item.label}</span> : <a href={item.url || '#'}>{item.label}</a>}
            </span>
          ))}
        </nav>
      );
    }
    case 'spinner':
      return (
        <div
          className="animate-spin"
          style={{
            ...estilo,
            backgroundColor: 'transparent',
            borderRadius: '50%',
            border: '3px solid currentColor',
            borderTopColor: 'transparent',
            color: estilos.fundoCor ?? '#6366f1',
          }}
        />
      );
    case 'skeleton':
      return <div style={estilo} className="animate-pulse" />;
    case 'toggle':
      return <TogglePublicado propriedades={no.propriedades} estilo={estilo} />;
    case 'alerta': {
      const variante = typeof no.propriedades?.variante === 'string' ? no.propriedades.variante : 'info';
      const esquema = ESQUEMAS_ALERTA[variante] ?? ESQUEMAS_ALERTA.info;
      return (
        <div style={{ ...estilo, display: 'flex', alignItems: 'center', gap: 8, color: esquema.cor, backgroundColor: esquema.fundo }}>
          <span>{esquema.simbolo}</span>
          <span>{texto}</span>
        </div>
      );
    }
    case 'container':
    default:
      return <div style={estilo}>{filhos}</div>;
  }
}

interface SlideImagem {
  src: string;
  alt?: string;
}

/** Carrossel funcional de verdade (com autoplay, se configurado) — a única
 * diferença do preview do editor é que aqui roda o autoplay real. */
function CarrosselPublicado({
  propriedades,
  estilo,
}: {
  propriedades?: Record<string, unknown>;
  estilo: React.CSSProperties;
}) {
  const imagens = Array.isArray(propriedades?.imagens) ? (propriedades!.imagens as SlideImagem[]) : [];
  const autoplay = propriedades?.autoplay === true;
  const intervalo = typeof propriedades?.intervalo === 'number' ? propriedades.intervalo : 3000;
  const mostrarSetas = propriedades?.mostrarSetas !== false;
  const mostrarPontos = propriedades?.mostrarPontos !== false;
  const [ativo, setAtivo] = useState(0);

  useEffect(() => {
    if (!autoplay || imagens.length <= 1) return;
    const id = setInterval(() => setAtivo((i) => (i + 1) % imagens.length), intervalo);
    return () => clearInterval(id);
  }, [autoplay, intervalo, imagens.length]);

  if (imagens.length === 0) return null;

  const indice = Math.min(ativo, imagens.length - 1);
  const slide = imagens[indice];

  return (
    <div style={{ ...estilo, position: 'relative', overflow: 'hidden' }}>
      <img src={slide.src} alt={slide.alt ?? ''} style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
      {mostrarSetas && imagens.length > 1 && (
        <>
          <button
            type="button"
            aria-label="Slide anterior"
            onClick={() => setAtivo((i) => (i - 1 + imagens.length) % imagens.length)}
            className="absolute left-2 top-1/2 -translate-y-1/2 w-7 h-7 flex items-center justify-center rounded-full bg-black/50 text-white text-sm"
          >
            ‹
          </button>
          <button
            type="button"
            aria-label="Próximo slide"
            onClick={() => setAtivo((i) => (i + 1) % imagens.length)}
            className="absolute right-2 top-1/2 -translate-y-1/2 w-7 h-7 flex items-center justify-center rounded-full bg-black/50 text-white text-sm"
          >
            ›
          </button>
        </>
      )}
      {mostrarPontos && imagens.length > 1 && (
        <div className="absolute bottom-2 left-1/2 -translate-x-1/2 flex gap-1.5">
          {imagens.map((_, i) => (
            <button
              key={i}
              type="button"
              aria-label={`Ir para o slide ${i + 1}`}
              onClick={() => setAtivo(i)}
              className={`w-2 h-2 rounded-full ${i === indice ? 'bg-white' : 'bg-white/50'}`}
            />
          ))}
        </div>
      )}
    </div>
  );
}

interface ItemMenu {
  label: string;
  url?: string;
}

function MenuPublicado({
  propriedades,
  estilo,
}: {
  propriedades?: Record<string, unknown>;
  estilo: React.CSSProperties;
}) {
  const itens = Array.isArray(propriedades?.itens) ? (propriedades!.itens as ItemMenu[]) : [];
  if (itens.length === 0) return null;

  return (
    <nav style={estilo}>
      {itens.map((item, i) => (
        <a key={i} href={item.url || '#'}>
          {item.label}
        </a>
      ))}
    </nav>
  );
}

interface ItemAccordion {
  titulo: string;
  conteudo?: string;
}

function AccordionPublicado({
  propriedades,
  estilo,
}: {
  propriedades?: Record<string, unknown>;
  estilo: React.CSSProperties;
}) {
  const itens = Array.isArray(propriedades?.itens) ? (propriedades!.itens as ItemAccordion[]) : [];
  const [aberto, setAberto] = useState<number | null>(0);
  if (itens.length === 0) return null;

  return (
    <div style={estilo}>
      {itens.map((item, i) => (
        <div key={i}>
          <button type="button" onClick={() => setAberto((a) => (a === i ? null : i))}>
            {item.titulo}
          </button>
          {aberto === i && <p>{item.conteudo}</p>}
        </div>
      ))}
    </div>
  );
}

function VideoPublicado({
  propriedades,
  estilo,
}: {
  propriedades?: Record<string, unknown>;
  estilo: React.CSSProperties;
}) {
  const src = typeof propriedades?.src === 'string' ? propriedades.src : '';
  if (!src) return null;

  const embed = paraEmbedUrl(src);
  return embed ? (
    <iframe
      src={embed}
      style={estilo}
      allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
      allowFullScreen
    />
  ) : (
    <video src={src} style={estilo} controls />
  );
}

interface ItemTab {
  titulo: string;
  conteudo?: string;
}

function TabsPublicado({
  propriedades,
  estilo,
}: {
  propriedades?: Record<string, unknown>;
  estilo: React.CSSProperties;
}) {
  const itens = Array.isArray(propriedades?.itens) ? (propriedades!.itens as ItemTab[]) : [];
  const [ativa, setAtiva] = useState(0);
  if (itens.length === 0) return null;

  const indice = Math.min(ativa, itens.length - 1);

  return (
    <div style={estilo}>
      <div>
        {itens.map((item, i) => (
          <button key={i} type="button" onClick={() => setAtiva(i)}>
            {item.titulo}
          </button>
        ))}
      </div>
      <p>{itens[indice]?.conteudo}</p>
    </div>
  );
}

/** Switch real e clicável — estado local inicializado por propriedades.ativo,
 * mesmo padrão de estado-não-persistido do AccordionPublicado/TabsPublicado. */
function TogglePublicado({
  propriedades,
  estilo,
}: {
  propriedades?: Record<string, unknown>;
  estilo: React.CSSProperties;
}) {
  const [ativo, setAtivo] = useState(propriedades?.ativo === true);
  const cor = (propriedades?.estilos as { fundoCor?: string } | undefined)?.fundoCor ?? '#6366f1';

  return (
    <button
      type="button"
      onClick={() => setAtivo((a) => !a)}
      style={{ ...estilo, position: 'relative', backgroundColor: ativo ? cor : '#d1d5db', border: 'none', padding: 0 }}
    >
      <span
        style={{
          position: 'absolute',
          top: 2,
          left: ativo ? '50%' : 2,
          width: 'calc(50% - 2px)',
          aspectRatio: '1',
          borderRadius: '50%',
          backgroundColor: '#ffffff',
        }}
      />
    </button>
  );
}

const ESQUEMAS_ALERTA: Record<string, { simbolo: string; cor: string; fundo: string }> = {
  info: { simbolo: 'ℹ', cor: '#1d4ed8', fundo: '#eff6ff' },
  sucesso: { simbolo: '✓', cor: '#15803d', fundo: '#f0fdf4' },
  aviso: { simbolo: '⚠', cor: '#b45309', fundo: '#fffbeb' },
  erro: { simbolo: '✕', cor: '#b91c1c', fundo: '#fef2f2' },
};
