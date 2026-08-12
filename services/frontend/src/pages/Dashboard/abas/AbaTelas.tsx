import { useParams } from 'react-router-dom';
import { useEffect, useState } from 'react';
import { useApp } from '../../../app/AppContext';
import { ApiError } from '../../../api/client';
import { useTelas } from '../../../systems/useTelas';
import { useCanvasDesign } from '../../../systems/useCanvasDesign';
import { ErrorState } from '../../../components/ui/StateViews';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogClose,
} from '../../../components/ui/dialog';
import { Input } from '../../../components/ui/input';
import { Button } from '../../../components/ui/button';
import { EditorTopbar, LARGURA_DEVICE, type Device, type ModoEditor } from '../editor/EditorTopbar';
import { PainelEsquerdo } from '../editor/PainelEsquerdo';
import { Canvas } from '../editor/Canvas';
import { Inspector } from '../editor/Inspector';
import { PreviewOverlay } from '../editor/PreviewOverlay';
import { encontrarNo } from '../../../collab/treeOps';
import { aceitaFilhos, type Estilos } from '../../../systems/componentRegistry';
import { useSessionStorageState } from '../../../systems/useSessionStorageState';
import type { DesignResumo } from '../../../api/types';

// Aba Telas (RF09/RF06): editor visual completo — topbar (troca de tela,
// undo/redo, device, zoom, status de salvamento), painel esquerdo
// (Componentes/Layers), canvas com árvore real (nesting, drag&drop
// before/inside/after, resize) e inspector seccionado (Layout/Typography/
// Background-Border-Shadow/Avançado). Sincronização em tempo real via
// services/collab (useCanvasDesign) — a persistência write-behind é
// responsabilidade do collab, nunca chamada diretamente daqui.

export function AbaTelas() {
  const { sistemaId = '' } = useParams<{ sistemaId: string }>();
  const { client } = useApp();
  const { estado, recarregar, criarTela } = useTelas(client, sistemaId);
  // sessionStorage: Telas/Regras de Negócio/Versão são rotas irmãs — trocar de
  // aba desmonta AbaTelas via <Outlet/>, perdendo o useState normal. Guardar
  // aqui mantém a tela aberta "do jeito que deixei" ao voltar.
  const [selecionadaId, setSelecionadaId] = useSessionStorageState<string | null>(
    `mach:sistema:${sistemaId}:tela-selecionada`,
    null,
  );
  const [criando, setCriando] = useState(false);
  const [dialogAberto, setDialogAberto] = useState(false);
  const [nomeNovaTela, setNomeNovaTela] = useState('');
  const [erroCriar, setErroCriar] = useState<string | null>(null);

  const telas = estado.fase === 'pronto' ? estado.telas : [];
  const telaSelecionada: DesignResumo | undefined = telas.find((t) => t.id === selecionadaId);

  async function criarNovaTela(e: React.FormEvent) {
    e.preventDefault();
    const nome = nomeNovaTela.trim();
    if (!nome) return;
    setCriando(true);
    setErroCriar(null);
    try {
      const id = await criarTela(nome);
      setSelecionadaId(id);
      setDialogAberto(false);
      setNomeNovaTela('');
    } catch (err) {
      setErroCriar(err instanceof ApiError ? err.message : String(err));
    } finally {
      setCriando(false);
    }
  }

  return (
    <div className="flex flex-col gap-4 h-full min-h-0">
      {estado.fase === 'erro' && telas.length === 0 ? (
        <ErrorState mensagem="Não foi possível carregar as telas." onRepetir={recarregar} />
      ) : (
        <>
          {!telaSelecionada && (
            <EditorTopbar
              telas={telas}
              telaSelecionadaId={selecionadaId}
              onSelecionarTela={setSelecionadaId}
              onNovaTela={() => setDialogAberto(true)}
              podeDesfazer={false}
              podeRefazer={false}
              onDesfazer={() => {}}
              onRefazer={() => {}}
              device="desktop"
              onDeviceChange={() => {}}
              zoom={1}
              onZoomChange={() => {}}
            />
          )}

          <Dialog
            open={dialogAberto}
            onOpenChange={(aberto) => {
              setDialogAberto(aberto);
              if (!aberto) setErroCriar(null);
            }}
          >
            <DialogContent>
              <form onSubmit={criarNovaTela} className="flex flex-col gap-4">
                <DialogHeader>
                  <DialogTitle>Nova tela</DialogTitle>
                </DialogHeader>
                <div className="flex flex-col gap-1">
                  <label htmlFor="nova-tela-nome" className="text-sm font-medium">
                    Nome da tela
                  </label>
                  <Input
                    id="nova-tela-nome"
                    autoFocus
                    value={nomeNovaTela}
                    onChange={(e) => setNomeNovaTela(e.target.value)}
                    placeholder="Ex.: Home"
                  />
                </div>
                {erroCriar && (
                  <p role="alert" className="text-destructive text-sm">
                    {erroCriar}
                  </p>
                )}
                <DialogFooter>
                  <DialogClose render={<Button type="button" variant="outline" />}>Cancelar</DialogClose>
                  <Button type="submit" disabled={criando || !nomeNovaTela.trim()}>
                    {criando ? 'Criando…' : 'Criar tela'}
                  </Button>
                </DialogFooter>
              </form>
            </DialogContent>
          </Dialog>

          {estado.fase === 'vazio' && !dialogAberto && (
            <p className="text-sm text-muted-foreground">
              Nenhuma tela criada ainda. Use "Nova tela" acima para começar.
            </p>
          )}

          {telaSelecionada ? (
            <TelaEditorArea
              key={telaSelecionada.id}
              sistemaId={sistemaId}
              tela={telaSelecionada}
              telas={telas}
              onSelecionarTela={setSelecionadaId}
              onNovaTela={() => setDialogAberto(true)}
            />
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-[280px_1fr_260px] gap-4 flex-1 min-h-0">
              <div className="bg-card border border-border rounded-2xl p-4 h-full" />
              <div className="bg-card/50 border border-dashed border-border rounded-2xl flex items-center justify-center p-8 h-full">
                <p className="text-sm text-muted-foreground text-center">Selecione ou crie uma tela para editar.</p>
              </div>
              <section aria-label="Propriedades" className="bg-card border border-border rounded-2xl p-4 h-full">
                <h3 className="text-sm font-heading font-bold text-muted-foreground uppercase tracking-wide mb-3">
                  Propriedades
                </h3>
                <p className="text-sm text-muted-foreground">Selecione um componente para editar suas propriedades.</p>
              </section>
            </div>
          )}
        </>
      )}
    </div>
  );
}

function TelaEditorArea({
  sistemaId,
  tela,
  telas,
  onSelecionarTela,
  onNovaTela,
}: {
  sistemaId: string;
  tela: DesignResumo;
  telas: DesignResumo[];
  onSelecionarTela: (id: string) => void;
  onNovaTela: () => void;
}) {
  const { client } = useApp();
  const canvas = useCanvasDesign(client, { designId: tela.id, sistemaId, nome: tela.nome });
  const [device, setDevice] = useSessionStorageState<Device>(`mach:sistema:${sistemaId}:device`, 'desktop');
  const [zoom, setZoom] = useSessionStorageState<number>(`mach:sistema:${sistemaId}:zoom`, 1);
  // Modo do editor — Edição/Visualização/Foco são mutuamente exclusivos (um
  // seletor de 3 posições, não três toggles independentes): Foco esconde os
  // painéis Componentes/Layers e Propriedades para dar o máximo de espaço ao
  // canvas; Visualização mostra o overlay de preview sem chrome de edição.
  // Persiste em sessionStorage pelo mesmo motivo do device/zoom (troca de aba
  // Telas/Regras/Versão desmonta o editor).
  const [modo, setModo] = useSessionStorageState<ModoEditor>(`mach:sistema:${sistemaId}:modo`, 'edicao');
  const focoAtivo = modo === 'foco';

  useEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      const alvo = e.target as HTMLElement | null;
      if (alvo && ['INPUT', 'TEXTAREA', 'SELECT'].includes(alvo.tagName)) return;
      if (alvo?.isContentEditable) return;
      const ctrl = e.ctrlKey || e.metaKey;
      if (ctrl && e.key.toLowerCase() === 'z') {
        e.preventDefault();
        if (e.shiftKey) canvas.refazer();
        else canvas.desfazer();
        return;
      }
      if (!ctrl && !e.altKey && e.key.toLowerCase() === 'f') {
        e.preventDefault();
        setModo((m) => (m === 'foco' ? 'edicao' : 'foco'));
      }
    }
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [canvas, setModo]);

  function adicionarNoAlvoPadrao(tipo: string) {
    if (!canvas.arvore) return;
    const noSelecionado = canvas.selecionado ? encontrarNo(canvas.arvore, canvas.selecionado) : null;
    const parent = noSelecionado && aceitaFilhos(noSelecionado.tipo) ? noSelecionado.blind_index : canvas.arvore.blind_index;
    canvas.adicionarComponente(tipo, parent);
  }

  function handleResize(blindIndex: string, largura: number, altura: number) {
    const no = canvas.arvore ? encontrarNo(canvas.arvore, blindIndex) : null;
    const estilosAtuais = (no?.propriedades?.estilos as Estilos | undefined) ?? {};
    canvas.atualizarPropriedades(blindIndex, {
      ...no?.propriedades,
      estilos: { ...estilosAtuais, largura: `${largura}px`, altura: `${altura}px` },
    });
  }

  function handleEditarTexto(blindIndex: string, html: string) {
    const no = canvas.arvore ? encontrarNo(canvas.arvore, blindIndex) : null;
    canvas.atualizarPropriedades(blindIndex, { ...no?.propriedades, texto: html });
  }

  function handleMoverAbsoluto(blindIndex: string, x: number, y: number) {
    const no = canvas.arvore ? encontrarNo(canvas.arvore, blindIndex) : null;
    const estilosAtuais = (no?.propriedades?.estilos as Estilos | undefined) ?? {};
    canvas.atualizarPropriedades(blindIndex, {
      ...no?.propriedades,
      estilos: { ...estilosAtuais, posicao: 'absolute', x, y },
    });
  }

  const componenteSelecionado =
    canvas.arvore && canvas.selecionado ? (encontrarNo(canvas.arvore, canvas.selecionado) ?? undefined) : undefined;

  return (
    <div className="flex flex-col gap-4 flex-1 min-h-0">
      <EditorTopbar
        telas={telas}
        telaSelecionadaId={tela.id}
        onSelecionarTela={onSelecionarTela}
        onNovaTela={onNovaTela}
        podeDesfazer={canvas.podeDesfazer}
        podeRefazer={canvas.podeRefazer}
        onDesfazer={canvas.desfazer}
        onRefazer={canvas.refazer}
        device={device}
        onDeviceChange={setDevice}
        zoom={zoom}
        onZoomChange={setZoom}
        statusSalvamento={canvas.estado.fase === 'pronto' ? canvas.statusSalvamento : undefined}
        modo={modo}
        onModoChange={setModo}
      />

      <div
        className={`grid grid-cols-1 gap-4 flex-1 min-h-0 ${focoAtivo ? '' : 'md:grid-cols-[280px_1fr_260px]'}`}
      >
        {!focoAtivo && (
          <PainelEsquerdo
            arvore={canvas.arvore}
            selecionado={canvas.selecionado}
            onSelecionar={canvas.selecionar}
            onRemover={canvas.removerComponente}
            onDuplicar={canvas.duplicarComponente}
            onAdicionar={adicionarNoAlvoPadrao}
          />
        )}

        {canvas.estado.fase === 'erro' ? (
          <ErrorState mensagem="Não foi possível carregar a tela." onRepetir={() => {}} />
        ) : canvas.estado.fase === 'carregando' || !canvas.arvore ? (
          <div className="bg-card/50 border border-dashed border-border rounded-2xl flex items-center justify-center p-8 h-full">
            <p className="text-sm text-muted-foreground">Carregando tela…</p>
          </div>
        ) : (
          <Canvas
            arvore={canvas.arvore}
            selecionado={canvas.selecionado}
            onSelecionar={canvas.selecionar}
            onAdicionar={canvas.adicionarComponente}
            onMover={canvas.moverNo}
            onResize={handleResize}
            onEditarTexto={handleEditarTexto}
            onMoverAbsoluto={handleMoverAbsoluto}
            larguraDevice={LARGURA_DEVICE[device]}
            zoom={zoom}
          />
        )}

        {!focoAtivo && (
          <Inspector
            componente={componenteSelecionado}
            onAtualizar={canvas.atualizarPropriedades}
            onRemover={canvas.removerComponente}
            onDuplicar={canvas.duplicarComponente}
          />
        )}
      </div>

      {modo === 'visualizacao' && canvas.arvore && (
        <PreviewOverlay
          nome={tela.nome}
          arvore={canvas.arvore}
          deviceInicial={device}
          onFechar={() => setModo('edicao')}
        />
      )}
    </div>
  );
}
