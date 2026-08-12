// Canvas do editor (RF09): renderização recursiva da árvore em fluxo normal
// (não posições soltas), seleção com contorno, handles de resize e
// drag&drop nativo com indicadores before/inside/after (seção 4/5 da spec) —
// tanto para reordenar/reparentar nós existentes quanto para soltar um
// componente novo vindo da paleta (dataTransfer "novo:<tipo>").

import { useRef, useState, type DragEvent, type KeyboardEvent as ReactKeyboardEvent, type PointerEvent as ReactPointerEvent } from 'react';
import type { Componente } from '../../../api/types';
import { aceitaFilhos, registroDoTipo, type Estilos } from '../../../systems/componentRegistry';
import { estilosParaCss } from '../../../systems/estilosCss';
import { iconePorNome } from '../../../systems/iconePicker';
import { paraEmbedUrl } from '../../../systems/videoEmbed';
import { sanitizarHtml } from '../../../systems/sanitizeHtml';
import { FloatingToolbar } from './FloatingToolbar';

export interface CanvasProps {
  arvore: Componente;
  selecionado: string | null;
  onSelecionar: (blindIndex: string | null) => void;
  onAdicionar: (tipo: string, parentBlindIndex?: string, index?: number) => void;
  onMover: (blindIndex: string, novoParent: string, index?: number) => void;
  onResize: (blindIndex: string, largura: number, altura: number) => void;
  onEditarTexto: (blindIndex: string, html: string) => void;
  onMoverAbsoluto: (blindIndex: string, x: number, y: number) => void;
  larguraDevice: number;
  zoom: number;
}

type DropAlvo = { blindIndex: string; posicao: 'before' | 'after' | 'inside' } | null;

export function Canvas({
  arvore,
  selecionado,
  onSelecionar,
  onAdicionar,
  onMover,
  onResize,
  onEditarTexto,
  onMoverAbsoluto,
  larguraDevice,
  zoom,
}: CanvasProps) {
  const [dropAlvo, setDropAlvo] = useState<DropAlvo>(null);

  function calcularAlvo(e: DragEvent, no: Componente): DropAlvo {
    const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
    const relY = (e.clientY - rect.top) / rect.height;
    if (aceitaFilhos(no.tipo) && relY > 0.25 && relY < 0.75) {
      return { blindIndex: no.blind_index, posicao: 'inside' };
    }
    return { blindIndex: no.blind_index, posicao: relY < 0.5 ? 'before' : 'after' };
  }

  function soltar(e: DragEvent, parentBlindIndex: string, indexNo: number) {
    e.preventDefault();
    e.stopPropagation();
    const dados = e.dataTransfer.getData('text/plain');
    const alvo = dropAlvo;
    setDropAlvo(null);
    if (!dados) return;

    let parentDestino = parentBlindIndex;
    let indexDestino: number | undefined = indexNo;
    if (alvo?.posicao === 'inside') {
      parentDestino = alvo.blindIndex;
      indexDestino = undefined;
    } else if (alvo?.posicao === 'after') {
      indexDestino = indexNo + 1;
    }

    if (dados.startsWith('novo:')) {
      onAdicionar(dados.slice('novo:'.length), parentDestino, indexDestino);
    } else if (dados !== parentDestino) {
      onMover(dados, parentDestino, indexDestino);
    }
  }

  const filhos = arvore.componente_filhos ?? [];

  return (
    <div className="flex flex-col gap-2 min-w-0 h-full overflow-auto">
      <div
        role="presentation"
        aria-label="Canvas"
        // Fluxo de bloco normal (não flex) na raiz do artboard: só assim um
        // filho com Display "inline-block" consegue ficar lado a lado com o
        // vizinho — um pai flex sempre "blockifica" os itens e os empilha
        // conforme a própria direção do flex, ignorando o display do filho.
        className="mx-auto bg-background border border-border rounded-2xl min-h-[500px] p-4 [&>*+*]:mt-2 transition-[width]"
        style={{ width: larguraDevice * zoom }}
        onClick={() => onSelecionar(null)}
        onDragOver={(e) => e.preventDefault()}
        onDrop={(e) => soltar(e, arvore.blind_index, filhos.length)}
      >
        {filhos.length === 0 && (
          <p className="text-sm text-muted-foreground text-center py-12">
            Tela em branco — arraste um componente da paleta ou clique para adicionar.
          </p>
        )}
        {filhos.map((no, index) => (
          <NoRenderizado
            key={no.blind_index}
            no={no}
            parentBlindIndex={arvore.blind_index}
            index={index}
            zoom={zoom}
            selecionado={selecionado}
            dropAlvo={dropAlvo}
            onSelecionar={onSelecionar}
            onResize={onResize}
            onEditarTexto={onEditarTexto}
            onMoverAbsoluto={onMoverAbsoluto}
            onDragOverNo={(e, alvoNo) => {
              e.preventDefault();
              e.stopPropagation();
              setDropAlvo(calcularAlvo(e, alvoNo));
            }}
            onDropNo={soltar}
            onDragEnd={() => setDropAlvo(null)}
          />
        ))}
      </div>
    </div>
  );
}

function NoRenderizado({
  no,
  parentBlindIndex,
  index,
  zoom,
  selecionado,
  dropAlvo,
  onSelecionar,
  onResize,
  onEditarTexto,
  onMoverAbsoluto,
  onDragOverNo,
  onDropNo,
  onDragEnd,
}: {
  no: Componente;
  parentBlindIndex: string;
  index: number;
  zoom: number;
  selecionado: string | null;
  dropAlvo: DropAlvo;
  onSelecionar: (blindIndex: string | null) => void;
  onResize: (blindIndex: string, largura: number, altura: number) => void;
  onEditarTexto: (blindIndex: string, html: string) => void;
  onMoverAbsoluto: (blindIndex: string, x: number, y: number) => void;
  onDragOverNo: (e: DragEvent, no: Componente) => void;
  onDropNo: (e: DragEvent, parentBlindIndex: string, index: number) => void;
  onDragEnd: () => void;
}) {
  const [resizePreview, setResizePreview] = useState<{ largura: number; altura: number } | null>(null);
  const [movePreview, setMovePreview] = useState<{ x: number; y: number } | null>(null);
  const [editandoTexto, setEditandoTexto] = useState(false);
  const ativo = selecionado === no.blind_index;
  const filhos = no.componente_filhos ?? [];
  const podeReceberFilhos = aceitaFilhos(no.tipo);
  const alvoAqui = dropAlvo?.blindIndex === no.blind_index ? dropAlvo.posicao : null;

  const estilos = (no.propriedades?.estilos as Estilos | undefined) ?? {};
  const ehAbsoluto = estilos.posicao === 'absolute';
  // O wrapper externo (linha abaixo) é quem participa do fluxo do pai — se
  // ele ficar sempre "block", um filho com Display "inline-block"/"inline"
  // nunca fica lado a lado com o vizinho, mesmo com o display certo no nó
  // interno. Por isso ele precisa espelhar o mesmo display do nó.
  const exibicaoExterna =
    estilos.display === 'inline-block' || estilos.display === 'inline' ? 'inline-block' : undefined;
  const estiloBase = estilosParaCss(estilos);
  const estiloFinal = {
    ...estiloBase,
    ...(resizePreview ? { width: resizePreview.largura, height: resizePreview.altura } : {}),
    ...(movePreview ? { left: movePreview.x, top: movePreview.y } : {}),
  };

  function iniciarResize(e: ReactPointerEvent) {
    e.stopPropagation();
    e.preventDefault();
    const alvo = e.currentTarget.parentElement as HTMLElement;
    const rect = alvo.getBoundingClientRect();
    const startX = e.clientX;
    const startY = e.clientY;
    const startW = rect.width;
    const startH = rect.height;

    function calcular(clientX: number, clientY: number) {
      return {
        largura: Math.max(20, Math.round(startW + (clientX - startX) / zoom)),
        altura: Math.max(20, Math.round(startH + (clientY - startY) / zoom)),
      };
    }
    function mover(ev: PointerEvent) {
      setResizePreview(calcular(ev.clientX, ev.clientY));
    }
    function soltar(ev: PointerEvent) {
      window.removeEventListener('pointermove', mover);
      window.removeEventListener('pointerup', soltar);
      const final = calcular(ev.clientX, ev.clientY);
      setResizePreview(null);
      onResize(no.blind_index, final.largura, final.altura);
    }
    window.addEventListener('pointermove', mover);
    window.addEventListener('pointerup', soltar);
  }

  /** Posicionamento livre (RF09): só ativo com posição "absolute" — a
   * alça aparece no canto oposto ao de resize e move o nó relativo ao pai
   * (que já é position:relative por padrão, ver ESTILOS_BASE). */
  function iniciarMove(e: ReactPointerEvent) {
    e.stopPropagation();
    e.preventDefault();
    const startX = e.clientX;
    const startY = e.clientY;
    const startLeft = estilos.x ?? 0;
    const startTop = estilos.y ?? 0;

    function calcular(clientX: number, clientY: number) {
      return {
        x: Math.round(startLeft + (clientX - startX) / zoom),
        y: Math.round(startTop + (clientY - startY) / zoom),
      };
    }
    function mover(ev: PointerEvent) {
      setMovePreview(calcular(ev.clientX, ev.clientY));
    }
    function soltar(ev: PointerEvent) {
      window.removeEventListener('pointermove', mover);
      window.removeEventListener('pointerup', soltar);
      const final = calcular(ev.clientX, ev.clientY);
      setMovePreview(null);
      onMoverAbsoluto(no.blind_index, final.x, final.y);
    }
    window.addEventListener('pointermove', mover);
    window.addEventListener('pointerup', soltar);
  }

  return (
    <div
      className="relative"
      data-testid={`layer-${no.blind_index}`}
      style={exibicaoExterna ? { display: exibicaoExterna, verticalAlign: 'top' } : undefined}
    >
      {alvoAqui === 'before' && <div className="h-0.5 bg-primary rounded-full mb-0.5" />}

      <div
        draggable={!editandoTexto && !ehAbsoluto}
        onDragStart={(e) => {
          e.stopPropagation();
          e.dataTransfer.setData('text/plain', no.blind_index);
          e.dataTransfer.effectAllowed = 'move';
        }}
        onDragEnd={onDragEnd}
        onDragOver={(e) => onDragOverNo(e, no)}
        onDrop={(e) => onDropNo(e, parentBlindIndex, index)}
        onClick={(e) => {
          e.stopPropagation();
          onSelecionar(no.blind_index);
        }}
        role="button"
        tabIndex={0}
        aria-label={no.tipo}
        className={`relative text-xs select-none ${ativo ? 'outline outline-2 outline-primary outline-offset-[-1px]' : ''} ${
          alvoAqui === 'inside' ? 'ring-2 ring-primary/60 bg-primary/5' : ''
        } ${no.tipo === 'skeleton' ? 'animate-pulse' : ''} ${ehAbsoluto ? 'cursor-move' : ''}`}
        style={{
          ...estiloFinal,
          display: estiloFinal.display ?? (podeReceberFilhos ? 'flex' : 'block'),
          flexDirection: estiloFinal.flexDirection ?? (podeReceberFilhos ? 'column' : undefined),
          minHeight: podeReceberFilhos ? 40 : undefined,
          border: estiloFinal.borderWidth ? undefined : '1px dashed var(--color-border)',
          // "Cor de fundo" no Spinner é reaproveitada como cor do anel (ver
          // SpinnerPreview), não como preenchimento — sem isso o anel fica
          // invisível contra o próprio fundo da mesma cor.
          backgroundColor: no.tipo === 'spinner' ? 'transparent' : estiloFinal.backgroundColor,
        }}
      >
        {ativo && (
          <span className="absolute -top-5 left-0 text-[10px] bg-primary text-primary-foreground px-1.5 py-0.5 rounded-t-md whitespace-nowrap">
            {no.tipo}
          </span>
        )}

        {podeReceberFilhos ? (
          <>
            {filhos.length === 0 && (
              <span className="text-muted-foreground text-[11px] p-2">Solte componentes aqui</span>
            )}
            {filhos.map((filho, i) => (
              <NoRenderizado
                key={filho.blind_index}
                no={filho}
                parentBlindIndex={no.blind_index}
                index={i}
                zoom={zoom}
                selecionado={selecionado}
                dropAlvo={dropAlvo}
                onSelecionar={onSelecionar}
                onResize={onResize}
                onEditarTexto={onEditarTexto}
                onMoverAbsoluto={onMoverAbsoluto}
                onDragOverNo={onDragOverNo}
                onDropNo={onDropNo}
                onDragEnd={onDragEnd}
              />
            ))}
          </>
        ) : no.tipo === 'imagem' && typeof no.propriedades?.src === 'string' && no.propriedades.src ? (
          <img
            src={no.propriedades.src}
            alt={typeof no.propriedades?.alt === 'string' ? no.propriedades.alt : ''}
            className="w-full h-full object-cover pointer-events-none select-none"
            draggable={false}
          />
        ) : no.tipo === 'avatar' && typeof no.propriedades?.src === 'string' && no.propriedades.src ? (
          <img
            src={no.propriedades.src}
            alt={typeof no.propriedades?.alt === 'string' ? no.propriedades.alt : ''}
            className="w-full h-full object-cover rounded-[inherit] pointer-events-none select-none"
            draggable={false}
          />
        ) : no.tipo === 'carrossel' ? (
          <CarrosselPreview propriedades={no.propriedades} />
        ) : no.tipo === 'menu' ? (
          <MenuPreview propriedades={no.propriedades} />
        ) : no.tipo === 'accordion' ? (
          <AccordionPreview propriedades={no.propriedades} />
        ) : no.tipo === 'divisor' ? null : no.tipo === 'video' ? (
          <VideoPreview propriedades={no.propriedades} />
        ) : no.tipo === 'tabs' ? (
          <TabsPreview propriedades={no.propriedades} />
        ) : no.tipo === 'avaliacao' ? (
          <AvaliacaoPreview propriedades={no.propriedades} />
        ) : no.tipo === 'icone' ? (
          <IconePreview propriedades={no.propriedades} />
        ) : no.tipo === 'progresso' ? (
          <ProgressoPreview propriedades={no.propriedades} />
        ) : no.tipo === 'breadcrumb' ? (
          <BreadcrumbPreview propriedades={no.propriedades} />
        ) : no.tipo === 'skeleton' ? null : no.tipo === 'spinner' ? (
          <SpinnerPreview propriedades={no.propriedades} />
        ) : no.tipo === 'toggle' ? (
          <ToggleSwitchPreview propriedades={no.propriedades} />
        ) : no.tipo === 'alerta' ? (
          <AlertaPreview propriedades={no.propriedades} />
        ) : registroDoTipo(no.tipo)?.temTexto ? (
          <TextoEditavel
            texto={typeof no.propriedades?.texto === 'string' ? no.propriedades.texto : ''}
            editando={editandoTexto}
            onEditandoChange={setEditandoTexto}
            onCommit={(html) => onEditarTexto(no.blind_index, html)}
          />
        ) : (
          <span className="p-2 inline-block">{no.tipo}</span>
        )}

        {ativo && ehAbsoluto && (
          <div
            onPointerDown={iniciarMove}
            role="presentation"
            aria-label={`Mover ${no.tipo}`}
            title="Arraste para posicionar livremente"
            className="absolute -top-1.5 -left-1.5 w-3 h-3 bg-primary rounded-full cursor-move z-10"
          />
        )}

        {ativo && (
          <div
            onPointerDown={iniciarResize}
            role="presentation"
            aria-label={`Redimensionar ${no.tipo}`}
            className="absolute -bottom-1.5 -right-1.5 w-3 h-3 bg-primary rounded-sm cursor-nwse-resize z-10"
          />
        )}
      </div>

      {alvoAqui === 'after' && <div className="h-0.5 bg-primary rounded-full mt-0.5" />}
    </div>
  );
}

interface SlideImagem {
  src: string;
  alt?: string;
}

/** Preview do Carrossel no canvas: mostra o slide ativo com setas/pontos de
 * navegação (o índice é só estado local do editor — não é persistido). */
function CarrosselPreview({ propriedades }: { propriedades?: Record<string, unknown> }) {
  const imagens = Array.isArray(propriedades?.imagens) ? (propriedades!.imagens as SlideImagem[]) : [];
  const mostrarSetas = propriedades?.mostrarSetas !== false;
  const mostrarPontos = propriedades?.mostrarPontos !== false;
  const [ativo, setAtivo] = useState(0);

  if (imagens.length === 0) {
    return (
      <span className="text-muted-foreground text-[11px] p-2 flex items-center justify-center w-full h-full text-center">
        Carrossel — adicione imagens no painel de propriedades.
      </span>
    );
  }

  const indice = Math.min(ativo, imagens.length - 1);
  const slide = imagens[indice];

  return (
    <div className="relative w-full h-full overflow-hidden select-none">
      <img
        src={slide.src}
        alt={slide.alt ?? ''}
        className="w-full h-full object-cover pointer-events-none"
        draggable={false}
      />
      {mostrarSetas && imagens.length > 1 && (
        <>
          <button
            type="button"
            aria-label="Slide anterior"
            onClick={(e) => {
              e.stopPropagation();
              setAtivo((i) => (i - 1 + imagens.length) % imagens.length);
            }}
            className="absolute left-1 top-1/2 -translate-y-1/2 w-5 h-5 flex items-center justify-center rounded-full bg-background/70 text-foreground text-xs"
          >
            ‹
          </button>
          <button
            type="button"
            aria-label="Próximo slide"
            onClick={(e) => {
              e.stopPropagation();
              setAtivo((i) => (i + 1) % imagens.length);
            }}
            className="absolute right-1 top-1/2 -translate-y-1/2 w-5 h-5 flex items-center justify-center rounded-full bg-background/70 text-foreground text-xs"
          >
            ›
          </button>
        </>
      )}
      {mostrarPontos && imagens.length > 1 && (
        <div className="absolute bottom-1.5 left-1/2 -translate-x-1/2 flex gap-1">
          {imagens.map((_, i) => (
            <button
              key={i}
              type="button"
              aria-label={`Ir para o slide ${i + 1}`}
              onClick={(e) => {
                e.stopPropagation();
                setAtivo(i);
              }}
              className={`w-1.5 h-1.5 rounded-full ${i === indice ? 'bg-primary' : 'bg-background/70'}`}
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

/** Preview do Menu no canvas: itens lado a lado (não navegam — é edição). */
function MenuPreview({ propriedades }: { propriedades?: Record<string, unknown> }) {
  const itens = Array.isArray(propriedades?.itens) ? (propriedades!.itens as ItemMenu[]) : [];

  if (itens.length === 0) {
    return (
      <span className="text-muted-foreground text-[11px] p-2 flex items-center w-full h-full">
        Menu — adicione itens no painel de propriedades.
      </span>
    );
  }

  return (
    <nav className="flex items-center gap-6 w-full h-full">
      {itens.map((item, i) => (
        <span key={i} className="text-sm select-none">
          {item.label}
        </span>
      ))}
    </nav>
  );
}

interface ItemAccordion {
  titulo: string;
  conteudo?: string;
}

/** Preview do Accordion no canvas: clicar no título expande/colapsa (estado
 * local do editor — não persistido). */
function AccordionPreview({ propriedades }: { propriedades?: Record<string, unknown> }) {
  const itens = Array.isArray(propriedades?.itens) ? (propriedades!.itens as ItemAccordion[]) : [];
  const [aberto, setAberto] = useState<number | null>(0);

  if (itens.length === 0) {
    return (
      <span className="text-muted-foreground text-[11px] p-2 flex items-center w-full h-full">
        Accordion — adicione itens no painel de propriedades.
      </span>
    );
  }

  return (
    <div className="flex flex-col w-full">
      {itens.map((item, i) => (
        <div key={i} className="border-b border-border last:border-b-0">
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation();
              setAberto((a) => (a === i ? null : i));
            }}
            className="w-full flex items-center justify-between text-left text-sm font-medium py-2"
          >
            {item.titulo}
            <span className="text-xs">{aberto === i ? '▾' : '▸'}</span>
          </button>
          {aberto === i && <p className="text-xs text-muted-foreground pb-2">{item.conteudo}</p>}
        </div>
      ))}
    </div>
  );
}

/** Preview do Vídeo no canvas: embed do YouTube/Vimeo por URL, ou tag
 * <video> pausada (não toca no editor) para arquivo direto. */
function VideoPreview({ propriedades }: { propriedades?: Record<string, unknown> }) {
  const src = typeof propriedades?.src === 'string' ? propriedades.src : '';
  if (!src) {
    return (
      <span className="text-muted-foreground text-[11px] p-2 flex items-center justify-center w-full h-full text-center">
        Vídeo — informe a URL no painel de propriedades.
      </span>
    );
  }
  const embed = paraEmbedUrl(src);
  return embed ? (
    <div className="w-full h-full pointer-events-none bg-black/80 flex items-center justify-center text-white text-[11px]">
      {embed}
    </div>
  ) : (
    <video src={src} className="w-full h-full object-cover pointer-events-none" muted />
  );
}

interface ItemTab {
  titulo: string;
  conteudo?: string;
}

/** Preview do Tabs no canvas: clicar numa aba troca o painel visível (estado
 * local do editor — não persistido). */
function TabsPreview({ propriedades }: { propriedades?: Record<string, unknown> }) {
  const itens = Array.isArray(propriedades?.itens) ? (propriedades!.itens as ItemTab[]) : [];
  const [ativa, setAtiva] = useState(0);

  if (itens.length === 0) {
    return (
      <span className="text-muted-foreground text-[11px] p-2 flex items-center w-full h-full">
        Tabs — adicione abas no painel de propriedades.
      </span>
    );
  }

  const indice = Math.min(ativa, itens.length - 1);

  return (
    <div className="flex flex-col w-full">
      <div className="flex gap-1 border-b border-border">
        {itens.map((item, i) => (
          <button
            key={i}
            type="button"
            onClick={(e) => {
              e.stopPropagation();
              setAtiva(i);
            }}
            className={`text-xs px-2.5 py-1.5 border-b-2 -mb-px ${
              i === indice ? 'border-primary text-foreground' : 'border-transparent text-muted-foreground'
            }`}
          >
            {item.titulo}
          </button>
        ))}
      </div>
      <p className="text-xs text-muted-foreground p-2">{itens[indice]?.conteudo}</p>
    </div>
  );
}

/** Preview da Avaliação no canvas: 5 estrelas, clicáveis para definir o valor. */
function AvaliacaoPreview({ propriedades }: { propriedades?: Record<string, unknown> }) {
  const valor = typeof propriedades?.valor === 'number' ? propriedades.valor : 0;
  return (
    <div className="flex items-center gap-0.5 w-full h-full select-none">
      {[1, 2, 3, 4, 5].map((n) => (
        <span key={n} className="leading-none">
          {n <= valor ? '★' : '☆'}
        </span>
      ))}
    </div>
  );
}

function IconePreview({ propriedades }: { propriedades?: Record<string, unknown> }) {
  const Icone = iconePorNome(typeof propriedades?.icone === 'string' ? propriedades.icone : undefined);
  return <Icone className="w-full h-full" />;
}

function ProgressoPreview({ propriedades }: { propriedades?: Record<string, unknown> }) {
  const valor = typeof propriedades?.valor === 'number' ? Math.min(100, Math.max(0, propriedades.valor)) : 50;
  const estilos = (propriedades?.estilos as Estilos | undefined) ?? {};
  return (
    <div className="w-full h-full rounded-full bg-secondary overflow-hidden">
      <div
        className="h-full rounded-full"
        style={{ width: `${valor}%`, backgroundColor: estilos.fundoCor ?? '#6366f1' }}
      />
    </div>
  );
}

interface ItemBreadcrumb {
  label: string;
  url?: string;
}

/** Preview do Breadcrumb no canvas: itens separados por "›" (não navegam). */
function BreadcrumbPreview({ propriedades }: { propriedades?: Record<string, unknown> }) {
  const itens = Array.isArray(propriedades?.itens) ? (propriedades!.itens as ItemBreadcrumb[]) : [];

  if (itens.length === 0) {
    return (
      <span className="text-muted-foreground text-[11px] p-2 flex items-center w-full h-full">
        Breadcrumb — adicione itens no painel de propriedades.
      </span>
    );
  }

  return (
    <nav className="flex items-center gap-1.5 w-full h-full text-sm select-none">
      {itens.map((item, i) => (
        <span key={i} className="flex items-center gap-1.5">
          {i > 0 && <span className="text-muted-foreground">›</span>}
          <span className={i === itens.length - 1 ? 'text-muted-foreground' : ''}>{item.label}</span>
        </span>
      ))}
    </nav>
  );
}

/** Preview do Spinner no canvas: anel girando — a cor reaproveita "Cor de
 * fundo" (mesma convenção do Progresso) em vez de criar um campo novo. */
function SpinnerPreview({ propriedades }: { propriedades?: Record<string, unknown> }) {
  const estilos = (propriedades?.estilos as Estilos | undefined) ?? {};
  return (
    <div
      className="w-full h-full rounded-full animate-spin"
      style={{ border: '3px solid currentColor', borderTopColor: 'transparent', color: estilos.fundoCor ?? '#6366f1' }}
    />
  );
}

/** Preview do Toggle no canvas: só leitura — o estado é editado pelo
 * Inspector (mesmo racional do slider do Progresso). */
function ToggleSwitchPreview({ propriedades }: { propriedades?: Record<string, unknown> }) {
  const estilos = (propriedades?.estilos as Estilos | undefined) ?? {};
  const ativo = propriedades?.ativo === true;
  const cor = estilos.fundoCor ?? '#6366f1';
  return (
    <div
      className="w-full h-full rounded-full relative transition-colors"
      style={{ backgroundColor: ativo ? cor : '#d1d5db' }}
    >
      <div
        className="absolute top-0.5 w-[calc(50%-2px)] aspect-square rounded-full bg-white transition-all"
        style={{ left: ativo ? 'calc(50% + 1px)' : '2px' }}
      />
    </div>
  );
}

const VARIANTES_ALERTA: Record<string, { simbolo: string; cor: string; fundo: string }> = {
  info: { simbolo: 'ℹ', cor: '#1d4ed8', fundo: '#eff6ff' },
  sucesso: { simbolo: '✓', cor: '#15803d', fundo: '#f0fdf4' },
  aviso: { simbolo: '⚠', cor: '#b45309', fundo: '#fffbeb' },
  erro: { simbolo: '✕', cor: '#b91c1c', fundo: '#fef2f2' },
};

/** Preview do Alerta no canvas: cor por variante (info/sucesso/aviso/erro). */
function AlertaPreview({ propriedades }: { propriedades?: Record<string, unknown> }) {
  const variante = typeof propriedades?.variante === 'string' ? propriedades.variante : 'info';
  const esquema = VARIANTES_ALERTA[variante] ?? VARIANTES_ALERTA.info;
  const texto = typeof propriedades?.texto === 'string' ? propriedades.texto : '';
  return (
    <div
      className="w-full h-full flex items-center gap-2"
      style={{ color: esquema.cor, backgroundColor: esquema.fundo }}
    >
      <span>{esquema.simbolo}</span>
      <span>{texto}</span>
    </div>
  );
}

/** Texto editável inline (RF09 — barra de formatação flutuante estilo
 * Notion): duplo clique entra em modo de edição (contentEditable);
 * selecionar um trecho mostra a FloatingToolbar perto da seleção para
 * aplicar negrito/itálico/sublinhado só naquele trecho — não no bloco
 * inteiro. O HTML resultante é sempre sanitizado antes de subir via
 * onCommit (allowlist de tags, ver systems/sanitizeHtml.ts). */
function TextoEditavel({
  texto,
  editando,
  onEditandoChange,
  onCommit,
}: {
  texto: string;
  editando: boolean;
  onEditandoChange: (editando: boolean) => void;
  onCommit: (html: string) => void;
}) {
  const [rectSelecao, setRectSelecao] = useState<DOMRect | null>(null);
  const ref = useRef<HTMLDivElement>(null);

  function atualizarSelecao() {
    const sel = window.getSelection();
    if (!sel || sel.isCollapsed || !ref.current || !sel.anchorNode || !ref.current.contains(sel.anchorNode)) {
      setRectSelecao(null);
      return;
    }
    setRectSelecao(sel.getRangeAt(0).getBoundingClientRect());
  }

  function aplicarComando(comando: string) {
    document.execCommand(comando);
    if (ref.current) onCommit(sanitizarHtml(ref.current.innerHTML));
    atualizarSelecao();
  }

  function commitAoSair() {
    onEditandoChange(false);
    setRectSelecao(null);
    if (ref.current) onCommit(sanitizarHtml(ref.current.innerHTML));
  }

  function onKeyDown(e: ReactKeyboardEvent<HTMLDivElement>) {
    e.stopPropagation(); // não deixa Delete/Ctrl+Z/F virar atalho global do editor
    if (e.key === 'Escape') {
      e.currentTarget.blur();
      return;
    }
    const ctrl = e.ctrlKey || e.metaKey;
    if (!ctrl) return;
    const tecla = e.key.toLowerCase();
    if (tecla === 'b' || tecla === 'i' || tecla === 'u') {
      e.preventDefault();
      aplicarComando(tecla === 'b' ? 'bold' : tecla === 'i' ? 'italic' : 'underline');
    }
  }

  if (!editando) {
    return (
      <span
        className="p-2 inline-block outline-none"
        onDoubleClick={(e) => {
          e.stopPropagation();
          onEditandoChange(true);
        }}
        // eslint-disable-next-line react/no-danger -- sanitizado em sanitizeHtml.ts antes de persistir
        dangerouslySetInnerHTML={{ __html: texto || '&nbsp;' }}
      />
    );
  }

  return (
    <>
      <div
        ref={ref}
        contentEditable
        suppressContentEditableWarning
        autoFocus
        className="p-2 inline-block outline-none ring-1 ring-primary/40 rounded"
        onBlur={commitAoSair}
        onMouseUp={atualizarSelecao}
        onKeyUp={atualizarSelecao}
        onClick={(e) => e.stopPropagation()}
        onKeyDown={onKeyDown}
        // eslint-disable-next-line react/no-danger -- sanitizado em sanitizeHtml.ts antes de persistir
        dangerouslySetInnerHTML={{ __html: texto }}
      />
      <FloatingToolbar
        rect={rectSelecao}
        onBold={() => aplicarComando('bold')}
        onItalic={() => aplicarComando('italic')}
        onUnderline={() => aplicarComando('underline')}
      />
    </>
  );
}
