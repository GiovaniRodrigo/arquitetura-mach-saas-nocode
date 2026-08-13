// Painel de propriedades / Inspector (RF09, seções 6/7/9/11/12/13): seções
// Layout, Typography (só para componentes com texto), Background/Border/
// Shadow, e um escape hatch "Avançado" em JSON bruto para o que não está
// coberto pelos campos estruturados. Cada campo estruturado aplica na hora
// (a mutação update_props substitui `propriedades` inteiro — nunca faz
// merge — por isso sempre reenviamos o objeto completo com só um campo
// trocado).

import { useState, type ChangeEvent } from 'react';
import type { Componente } from '../../../api/types';
import { registroDoTipo, type Estilos } from '../../../systems/componentRegistry';
import { ICONES_DISPONIVEIS, NOME_ICONE_PADRAO } from '../../../systems/iconePicker';
import { htmlParaTextoPlano } from '../../../systems/sanitizeHtml';

const TAMANHO_MAX_IMAGEM_BYTES = 2 * 1024 * 1024; // 2MB — a imagem vai inteira (base64) na árvore JSONB do design.

export interface InspectorProps {
  componente: Componente | undefined;
  onAtualizar: (blindIndex: string, propriedades: Record<string, unknown>) => void;
  onRemover: (blindIndex: string) => void;
  onDuplicar: (blindIndex: string) => void;
  onTrazerParaFrente: (blindIndex: string) => void;
  onEnviarParaTras: (blindIndex: string) => void;
}

export function Inspector({
  componente,
  onAtualizar,
  onRemover,
  onDuplicar,
  onTrazerParaFrente,
  onEnviarParaTras,
}: InspectorProps) {
  if (!componente) {
    return (
      <section aria-label="Propriedades" className="bg-card border border-border rounded-2xl p-4 h-full">
        <TituloSecao>Propriedades</TituloSecao>
        <p className="text-sm text-muted-foreground">Selecione um componente para editar suas propriedades.</p>
      </section>
    );
  }

  const registro = registroDoTipo(componente.tipo);
  const estilos: Estilos = (componente.propriedades?.estilos as Estilos | undefined) ?? {};

  function aplicarEstilo<K extends keyof Estilos>(campo: K, valor: Estilos[K]) {
    const valorFinal = valor === '' ? undefined : valor;
    onAtualizar(componente!.blind_index, {
      ...componente!.propriedades,
      estilos: { ...estilos, [campo]: valorFinal },
    });
  }

  function aplicarTexto(valor: string) {
    onAtualizar(componente!.blind_index, { ...componente!.propriedades, texto: valor });
  }

  function aplicarPropriedade(campo: string, valor: string) {
    onAtualizar(componente!.blind_index, { ...componente!.propriedades, [campo]: valor });
  }

  function aplicarCampo(campo: string, valor: unknown) {
    onAtualizar(componente!.blind_index, { ...componente!.propriedades, [campo]: valor });
  }

  const flex = estilos.display === 'flex';

  return (
    <section
      aria-label="Propriedades"
      className="bg-card border border-border rounded-2xl p-4 flex flex-col gap-4 h-full min-h-0 overflow-y-auto scrollbar-app"
    >
      <div className="flex items-center justify-between">
        <TituloSecao>Propriedades</TituloSecao>
        <span className="text-xs text-muted-foreground">{componente.tipo}</span>
      </div>

      <div className="flex gap-2">
        <button
          type="button"
          onClick={() => onDuplicar(componente.blind_index)}
          className="flex-1 text-xs px-2.5 py-1.5 rounded-lg border border-border hover:bg-secondary"
        >
          Duplicar
        </button>
        <button
          type="button"
          onClick={() => onRemover(componente.blind_index)}
          className="flex-1 text-xs px-2.5 py-1.5 rounded-lg border border-destructive/30 text-destructive hover:bg-destructive/10"
        >
          Remover
        </button>
      </div>

      {registro?.temTexto && (
        <Grupo titulo="Conteúdo">
          <Campo label="Texto">
            <input
              value={
                typeof componente.propriedades?.texto === 'string'
                  ? htmlParaTextoPlano(componente.propriedades.texto)
                  : ''
              }
              onChange={(e) => aplicarTexto(e.target.value)}
              className={campoClasse}
            />
          </Campo>
          <p className="text-[11px] text-muted-foreground">
            Duplo clique no texto no canvas para negrito/itálico/sublinhado por trecho.
          </p>
        </Grupo>
      )}

      {componente.tipo === 'imagem' && (
        <SecaoImagem componente={componente} onAplicar={aplicarPropriedade} />
      )}

      {componente.tipo === 'carrossel' && (
        <SecaoCarrossel componente={componente} onAplicar={aplicarCampo} />
      )}

      {componente.tipo === 'menu' && <SecaoMenu componente={componente} onAplicar={aplicarCampo} />}

      {componente.tipo === 'accordion' && (
        <SecaoAccordion componente={componente} onAplicar={aplicarCampo} />
      )}

      {componente.tipo === 'video' && <SecaoVideo componente={componente} onAplicar={aplicarPropriedade} />}

      {componente.tipo === 'tabs' && <SecaoTabs componente={componente} onAplicar={aplicarCampo} />}

      {componente.tipo === 'avaliacao' && (
        <SecaoAvaliacao componente={componente} onAplicar={aplicarCampo} />
      )}

      {componente.tipo === 'icone' && <SecaoIcone componente={componente} onAplicar={aplicarPropriedade} />}

      {componente.tipo === 'progresso' && (
        <SecaoProgresso componente={componente} onAplicar={aplicarCampo} />
      )}

      {componente.tipo === 'breadcrumb' && (
        <SecaoBreadcrumb componente={componente} onAplicar={aplicarCampo} />
      )}

      {componente.tipo === 'toggle' && <SecaoToggle componente={componente} onAplicar={aplicarCampo} />}

      {componente.tipo === 'alerta' && <SecaoAlerta componente={componente} onAplicar={aplicarCampo} />}

      {componente.tipo === 'avatar' && (
        <SecaoAvatar componente={componente} onAplicar={aplicarPropriedade} />
      )}

      <Grupo titulo="Layout">
        <Campo label="Posição">
          <Selecao
            valor={estilos.posicao ?? 'relative'}
            opcoes={['static', 'relative', 'absolute']}
            onChange={(v) => aplicarEstilo('posicao', v as Estilos['posicao'])}
          />
        </Campo>
        {estilos.posicao === 'absolute' && (
          <div className="flex gap-2">
            <Campo label="X">
              <input
                type="number"
                value={estilos.x ?? 0}
                onChange={(e) => aplicarEstilo('x', Number(e.target.value))}
                className={campoClasse}
              />
            </Campo>
            <Campo label="Y">
              <input
                type="number"
                value={estilos.y ?? 0}
                onChange={(e) => aplicarEstilo('y', Number(e.target.value))}
                className={campoClasse}
              />
            </Campo>
          </div>
        )}
        {estilos.posicao === 'absolute' && (
          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => onTrazerParaFrente(componente.blind_index)}
              className="flex-1 text-xs px-2.5 py-1.5 rounded-lg border border-border hover:bg-secondary"
            >
              Trazer para frente
            </button>
            <button
              type="button"
              onClick={() => onEnviarParaTras(componente.blind_index)}
              className="flex-1 text-xs px-2.5 py-1.5 rounded-lg border border-border hover:bg-secondary"
            >
              Enviar para trás
            </button>
          </div>
        )}
        <CampoMedida label="Largura" valor={estilos.largura} onChange={(v) => aplicarEstilo('largura', v)} />
        <CampoMedida label="Altura" valor={estilos.altura} onChange={(v) => aplicarEstilo('altura', v)} />
        <Campo label="Display">
          <Selecao
            valor={estilos.display ?? 'block'}
            opcoes={['block', 'inline-block', 'inline', 'flex', 'none']}
            onChange={(v) => aplicarEstilo('display', v as Estilos['display'])}
          />
        </Campo>
        {flex && (
          <>
            <Campo label="Direção">
              <Selecao
                valor={estilos.direcao ?? 'row'}
                opcoes={['row', 'column']}
                onChange={(v) => aplicarEstilo('direcao', v as Estilos['direcao'])}
              />
            </Campo>
            <Campo label="Justificar">
              <Selecao
                valor={estilos.justificar ?? 'flex-start'}
                opcoes={['flex-start', 'center', 'flex-end', 'space-between', 'space-around']}
                onChange={(v) => aplicarEstilo('justificar', v as Estilos['justificar'])}
              />
            </Campo>
            <Campo label="Alinhar">
              <Selecao
                valor={estilos.alinhar ?? 'stretch'}
                opcoes={['flex-start', 'center', 'flex-end', 'stretch']}
                onChange={(v) => aplicarEstilo('alinhar', v as Estilos['alinhar'])}
              />
            </Campo>
            <CampoMedida
              label="Gap"
              valor={estilos.espacamento}
              onChange={(v) => aplicarEstilo('espacamento', v)}
              unidadesPermitidas={['px', '%']}
              permiteAuto={false}
            />
          </>
        )}
        <Campo label="Padding">
          <input
            value={estilos.padding ?? ''}
            placeholder="8px 16px"
            onChange={(e) => aplicarEstilo('padding', e.target.value)}
            className={campoClasse}
          />
        </Campo>
        <Campo label="Margem">
          <input
            value={estilos.margem ?? ''}
            placeholder="0"
            onChange={(e) => aplicarEstilo('margem', e.target.value)}
            className={campoClasse}
          />
        </Campo>
      </Grupo>

      {registro?.temTexto && (
        <Grupo titulo="Typography">
          <CampoMedida
            label="Tamanho da fonte"
            valor={estilos.fonteTamanho}
            onChange={(v) => aplicarEstilo('fonteTamanho', v)}
            unidadesPermitidas={['px', 'rem']}
            permiteAuto={false}
          />
          <Campo label="Peso">
            <Selecao
              valor={estilos.fontePeso ?? 'normal'}
              opcoes={['normal', 'medium', 'bold']}
              onChange={(v) => aplicarEstilo('fontePeso', v as Estilos['fontePeso'])}
            />
          </Campo>
          <Campo label="Alinhamento do texto">
            <Selecao
              valor={estilos.textoAlinhar ?? 'left'}
              opcoes={['left', 'center', 'right']}
              onChange={(v) => aplicarEstilo('textoAlinhar', v as Estilos['textoAlinhar'])}
            />
          </Campo>
          <Campo label="Cor do texto">
            <input
              type="color"
              value={estilos.cor ?? '#000000'}
              onChange={(e) => aplicarEstilo('cor', e.target.value)}
              className="h-8 w-full rounded-lg border border-border bg-transparent"
            />
          </Campo>
        </Grupo>
      )}

      <Grupo titulo="Background / Border / Shadow">
        <Campo label="Cor de fundo">
          <input
            type="color"
            value={estilos.fundoCor ?? '#ffffff'}
            onChange={(e) => aplicarEstilo('fundoCor', e.target.value)}
            className="h-8 w-full rounded-lg border border-border bg-transparent"
          />
        </Campo>
        <CampoMedida
          label="Borda (largura)"
          valor={estilos.bordaLargura}
          onChange={(v) => aplicarEstilo('bordaLargura', v)}
          unidadesPermitidas={['px']}
          permiteAuto={false}
        />
        <Campo label="Borda (cor)">
          <input
            type="color"
            value={estilos.bordaCor ?? '#000000'}
            onChange={(e) => aplicarEstilo('bordaCor', e.target.value)}
            className="h-8 w-full rounded-lg border border-border bg-transparent"
          />
        </Campo>
        <CampoMedida
          label="Raio da borda"
          valor={estilos.bordaRaio}
          onChange={(v) => aplicarEstilo('bordaRaio', v)}
          unidadesPermitidas={['px', '%']}
          permiteAuto={false}
        />
        <Campo label="Sombra">
          <input
            value={estilos.sombra ?? ''}
            placeholder="0 4px 12px rgba(0,0,0,.15)"
            onChange={(e) => aplicarEstilo('sombra', e.target.value)}
            className={campoClasse}
          />
        </Campo>
        <Campo label={`Opacidade${estilos.opacidade !== undefined ? ` (${estilos.opacidade})` : ''}`}>
          <input
            type="range"
            min={0}
            max={1}
            step={0.05}
            value={estilos.opacidade ?? 1}
            onChange={(e) => aplicarEstilo('opacidade', Number(e.target.value))}
            className="w-full"
          />
        </Campo>
      </Grupo>

      <SecaoAvancada componente={componente} onAtualizar={onAtualizar} />
    </section>
  );
}

function TituloSecao({ children }: { children: React.ReactNode }) {
  return (
    <h3 className="text-sm font-heading font-bold text-muted-foreground uppercase tracking-wide">{children}</h3>
  );
}

function Grupo({ titulo, children }: { titulo: string; children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-2 border-t border-border pt-3 first:border-t-0 first:pt-0">
      <h4 className="text-xs font-heading font-bold text-muted-foreground uppercase tracking-wide">{titulo}</h4>
      {children}
    </div>
  );
}

function Campo({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="flex flex-col gap-1 text-xs">
      <span className="text-muted-foreground">{label}</span>
      {children}
    </label>
  );
}

const campoClasse = 'h-8 w-full rounded-lg border border-border bg-background px-2 text-xs text-foreground';

function Selecao({
  valor,
  opcoes,
  onChange,
}: {
  valor: string;
  opcoes: string[];
  onChange: (v: string) => void;
}) {
  return (
    <select value={valor} onChange={(e) => onChange(e.target.value)} className={campoClasse}>
      {opcoes.map((o) => (
        <option key={o} value={o}>
          {o}
        </option>
      ))}
    </select>
  );
}

const UNIDADES_MEDIDA = ['px', '%', 'vw', 'vh', 'rem'] as const;
type UnidadeMedida = (typeof UNIDADES_MEDIDA)[number] | 'auto';

/** Extrai {numero, unidade} de um valor CSS de medida simples (ex.: "200px", "auto", "50%"). */
export function parseMedida(valor: string | undefined): { numero: string; unidade: UnidadeMedida } {
  if (!valor || valor === 'auto') return { numero: '', unidade: 'auto' };
  const m = valor.match(/^(-?\d*\.?\d+)(px|%|vw|vh|rem)$/);
  if (m) return { numero: m[1], unidade: m[2] as UnidadeMedida };
  return { numero: '', unidade: 'px' };
}

/** Campo de medida (largura/altura/gap/fonte/borda/raio): número + select de unidade,
 * em vez de texto livre — evita erros de digitação de unidade e reflete o valor
 * CSS de forma estruturada (seção 6/9/12 — Width/Typography/Border). */
function CampoMedida({
  label,
  valor,
  onChange,
  unidadesPermitidas = UNIDADES_MEDIDA,
  permiteAuto = true,
}: {
  label: string;
  valor?: string;
  onChange: (v: string) => void;
  unidadesPermitidas?: readonly string[];
  permiteAuto?: boolean;
}) {
  const { numero, unidade } = parseMedida(valor);
  const opcoesUnidade = permiteAuto ? [...unidadesPermitidas, 'auto'] : unidadesPermitidas;

  function mudarUnidade(novaUnidade: string) {
    onChange(novaUnidade === 'auto' ? 'auto' : `${numero || 0}${novaUnidade}`);
  }

  function mudarNumero(novoNumero: string) {
    onChange(`${novoNumero}${unidade === 'auto' ? unidadesPermitidas[0] : unidade}`);
  }

  return (
    <Campo label={label}>
      <div className="flex gap-1">
        {unidade !== 'auto' && (
          <input
            type="number"
            value={numero}
            onChange={(e) => mudarNumero(e.target.value)}
            className="h-8 min-w-0 flex-1 rounded-lg border border-border bg-background px-2 text-xs text-foreground"
          />
        )}
        <select
          value={unidade}
          onChange={(e) => mudarUnidade(e.target.value)}
          className={`h-8 shrink-0 rounded-lg border border-border bg-background px-1 text-xs text-foreground ${
            unidade === 'auto' ? 'flex-1' : 'w-[4.5rem]'
          }`}
        >
          {opcoesUnidade.map((u) => (
            <option key={u} value={u}>
              {u}
            </option>
          ))}
        </select>
      </div>
    </Campo>
  );
}

/** Conteúdo do componente Imagem: upload de arquivo (vira data URL — sem
 * backend de assets ainda) ou URL externa, mais texto alternativo. */
function SecaoImagem({
  componente,
  onAplicar,
}: {
  componente: Componente;
  onAplicar: (campo: string, valor: string) => void;
}) {
  const [erro, setErro] = useState<string | null>(null);
  const src = typeof componente.propriedades?.src === 'string' ? componente.propriedades.src : '';
  const alt = typeof componente.propriedades?.alt === 'string' ? componente.propriedades.alt : '';

  function onArquivoSelecionado(e: ChangeEvent<HTMLInputElement>) {
    const arquivo = e.target.files?.[0];
    e.target.value = '';
    if (!arquivo) return;
    if (arquivo.size > TAMANHO_MAX_IMAGEM_BYTES) {
      setErro('Imagem muito grande (máx. 2MB).');
      return;
    }
    const leitor = new FileReader();
    leitor.onload = () => {
      if (typeof leitor.result === 'string') {
        setErro(null);
        onAplicar('src', leitor.result);
      }
    };
    leitor.onerror = () => setErro('Não foi possível ler o arquivo.');
    leitor.readAsDataURL(arquivo);
  }

  return (
    <Grupo titulo="Imagem">
      {src && (
        <img
          src={src}
          alt={alt}
          className="w-full h-24 object-cover rounded-lg border border-border"
        />
      )}
      <Campo label="Enviar arquivo">
        <input
          type="file"
          accept="image/*"
          onChange={onArquivoSelecionado}
          className="text-xs text-muted-foreground file:mr-2 file:px-2 file:py-1 file:rounded-lg file:border file:border-border file:bg-secondary file:text-secondary-foreground file:text-xs"
        />
      </Campo>
      {erro && <p className="text-xs text-destructive">{erro}</p>}
      <Campo label="ou URL da imagem">
        <input
          value={src}
          onChange={(e) => onAplicar('src', e.target.value)}
          placeholder="https://…"
          className={campoClasse}
        />
      </Campo>
      <Campo label="Texto alternativo">
        <input
          value={alt}
          onChange={(e) => onAplicar('alt', e.target.value)}
          placeholder="Descrição da imagem"
          className={campoClasse}
        />
      </Campo>
    </Grupo>
  );
}

interface SlideImagem {
  src: string;
  alt?: string;
}

/** Conteúdo do componente Carrossel: lista de imagens (upload múltiplo, cada
 * uma vira data URL como na Imagem) + opções de navegação/autoplay. */
function SecaoCarrossel({
  componente,
  onAplicar,
}: {
  componente: Componente;
  onAplicar: (campo: string, valor: unknown) => void;
}) {
  const [erro, setErro] = useState<string | null>(null);
  const imagens: SlideImagem[] = Array.isArray(componente.propriedades?.imagens)
    ? (componente.propriedades.imagens as SlideImagem[])
    : [];
  const autoplay = componente.propriedades?.autoplay === true;
  const intervalo =
    typeof componente.propriedades?.intervalo === 'number' ? componente.propriedades.intervalo : 3000;
  const mostrarSetas = componente.propriedades?.mostrarSetas !== false;
  const mostrarPontos = componente.propriedades?.mostrarPontos !== false;

  function onArquivosSelecionados(e: ChangeEvent<HTMLInputElement>) {
    const arquivos = Array.from(e.target.files ?? []);
    e.target.value = '';
    if (arquivos.length === 0) return;
    const grande = arquivos.find((a) => a.size > TAMANHO_MAX_IMAGEM_BYTES);
    if (grande) {
      setErro('Imagem muito grande (máx. 2MB).');
      return;
    }
    setErro(null);
    Promise.all(
      arquivos.map(
        (arquivo) =>
          new Promise<string>((resolve, reject) => {
            const leitor = new FileReader();
            leitor.onload = () => (typeof leitor.result === 'string' ? resolve(leitor.result) : reject());
            leitor.onerror = () => reject();
            leitor.readAsDataURL(arquivo);
          }),
      ),
    )
      .then((novasSrcs) => {
        onAplicar('imagens', [...imagens, ...novasSrcs.map((src) => ({ src }))]);
      })
      .catch(() => setErro('Não foi possível ler o arquivo.'));
  }

  function removerSlide(indice: number) {
    onAplicar(
      'imagens',
      imagens.filter((_, i) => i !== indice),
    );
  }

  return (
    <Grupo titulo="Carrossel">
      {imagens.length > 0 && (
        <ul className="flex flex-col gap-1.5">
          {imagens.map((slide, i) => (
            <li key={i} className="flex items-center gap-2">
              <img src={slide.src} alt={slide.alt ?? ''} className="w-10 h-8 object-cover rounded-md border border-border shrink-0" />
              <span className="text-xs text-muted-foreground flex-1 truncate">Slide {i + 1}</span>
              <button
                type="button"
                onClick={() => removerSlide(i)}
                aria-label={`Remover slide ${i + 1}`}
                className="text-xs text-destructive hover:underline"
              >
                Remover
              </button>
            </li>
          ))}
        </ul>
      )}
      <Campo label="Adicionar imagens">
        <input
          type="file"
          accept="image/*"
          multiple
          onChange={onArquivosSelecionados}
          className="text-xs text-muted-foreground file:mr-2 file:px-2 file:py-1 file:rounded-lg file:border file:border-border file:bg-secondary file:text-secondary-foreground file:text-xs"
        />
      </Campo>
      {erro && <p className="text-xs text-destructive">{erro}</p>}
      <label className="flex items-center gap-2 text-xs">
        <input type="checkbox" checked={autoplay} onChange={(e) => onAplicar('autoplay', e.target.checked)} />
        <span>Autoplay</span>
      </label>
      {autoplay && (
        <Campo label="Intervalo (ms)">
          <input
            type="number"
            min={500}
            step={500}
            value={intervalo}
            onChange={(e) => onAplicar('intervalo', Number(e.target.value))}
            className={campoClasse}
          />
        </Campo>
      )}
      <label className="flex items-center gap-2 text-xs">
        <input
          type="checkbox"
          checked={mostrarSetas}
          onChange={(e) => onAplicar('mostrarSetas', e.target.checked)}
        />
        <span>Mostrar setas</span>
      </label>
      <label className="flex items-center gap-2 text-xs">
        <input
          type="checkbox"
          checked={mostrarPontos}
          onChange={(e) => onAplicar('mostrarPontos', e.target.checked)}
        />
        <span>Mostrar pontos</span>
      </label>
    </Grupo>
  );
}

interface ItemMenu {
  label: string;
  url?: string;
}

/** Conteúdo do componente Menu: lista de itens (label + link), editáveis
 * inline, sem afetar a árvore de componentes (self-contained, como o
 * Carrossel). */
function SecaoMenu({
  componente,
  onAplicar,
}: {
  componente: Componente;
  onAplicar: (campo: string, valor: unknown) => void;
}) {
  const itens: ItemMenu[] = Array.isArray(componente.propriedades?.itens)
    ? (componente.propriedades.itens as ItemMenu[])
    : [];

  function atualizarItem(indice: number, campo: keyof ItemMenu, valor: string) {
    onAplicar(
      'itens',
      itens.map((item, i) => (i === indice ? { ...item, [campo]: valor } : item)),
    );
  }

  function removerItem(indice: number) {
    onAplicar(
      'itens',
      itens.filter((_, i) => i !== indice),
    );
  }

  function adicionarItem() {
    onAplicar('itens', [...itens, { label: 'Novo item', url: '#' }]);
  }

  return (
    <Grupo titulo="Menu">
      {itens.length > 0 && (
        <ul className="flex flex-col gap-2">
          {itens.map((item, i) => (
            <li key={i} className="flex flex-col gap-1 border border-border rounded-lg p-2">
              <div className="flex items-center gap-1">
                <input
                  value={item.label}
                  onChange={(e) => atualizarItem(i, 'label', e.target.value)}
                  placeholder="Rótulo"
                  aria-label={`Rótulo do item ${i + 1}`}
                  className={campoClasse}
                />
                <button
                  type="button"
                  onClick={() => removerItem(i)}
                  aria-label={`Remover item ${i + 1}`}
                  className="text-xs text-destructive shrink-0 px-1"
                >
                  Remover
                </button>
              </div>
              <input
                value={item.url ?? ''}
                onChange={(e) => atualizarItem(i, 'url', e.target.value)}
                placeholder="URL (ex.: /sobre)"
                aria-label={`URL do item ${i + 1}`}
                className={campoClasse}
              />
            </li>
          ))}
        </ul>
      )}
      <button
        type="button"
        onClick={adicionarItem}
        className="self-start text-xs bg-primary text-primary-foreground px-3 py-1.5 rounded-lg font-medium hover:bg-primary/90"
      >
        + Adicionar item
      </button>
    </Grupo>
  );
}

interface ItemAccordion {
  titulo: string;
  conteudo?: string;
}

/** Conteúdo do componente Accordion: lista de painéis (título + conteúdo). */
function SecaoAccordion({
  componente,
  onAplicar,
}: {
  componente: Componente;
  onAplicar: (campo: string, valor: unknown) => void;
}) {
  const itens: ItemAccordion[] = Array.isArray(componente.propriedades?.itens)
    ? (componente.propriedades.itens as ItemAccordion[])
    : [];

  function atualizarItem(indice: number, campo: keyof ItemAccordion, valor: string) {
    onAplicar(
      'itens',
      itens.map((item, i) => (i === indice ? { ...item, [campo]: valor } : item)),
    );
  }

  function removerItem(indice: number) {
    onAplicar(
      'itens',
      itens.filter((_, i) => i !== indice),
    );
  }

  function adicionarItem() {
    onAplicar('itens', [...itens, { titulo: 'Novo painel', conteudo: '' }]);
  }

  return (
    <Grupo titulo="Accordion">
      {itens.length > 0 && (
        <ul className="flex flex-col gap-2">
          {itens.map((item, i) => (
            <li key={i} className="flex flex-col gap-1 border border-border rounded-lg p-2">
              <div className="flex items-center gap-1">
                <input
                  value={item.titulo}
                  onChange={(e) => atualizarItem(i, 'titulo', e.target.value)}
                  placeholder="Título"
                  aria-label={`Título do painel ${i + 1}`}
                  className={campoClasse}
                />
                <button
                  type="button"
                  onClick={() => removerItem(i)}
                  aria-label={`Remover painel ${i + 1}`}
                  className="text-xs text-destructive shrink-0 px-1"
                >
                  Remover
                </button>
              </div>
              <textarea
                value={item.conteudo ?? ''}
                onChange={(e) => atualizarItem(i, 'conteudo', e.target.value)}
                placeholder="Conteúdo"
                aria-label={`Conteúdo do painel ${i + 1}`}
                rows={2}
                className="text-xs px-2 py-1.5 bg-background border border-border rounded-lg text-foreground"
              />
            </li>
          ))}
        </ul>
      )}
      <button
        type="button"
        onClick={adicionarItem}
        className="self-start text-xs bg-primary text-primary-foreground px-3 py-1.5 rounded-lg font-medium hover:bg-primary/90"
      >
        + Adicionar painel
      </button>
    </Grupo>
  );
}

/** Conteúdo do componente Vídeo: só URL (YouTube/Vimeo ou arquivo direto) —
 * sem upload, porque vídeo não cabe no limite prático de base64 na árvore
 * JSONB do jeito que a Imagem cabe. */
function SecaoVideo({
  componente,
  onAplicar,
}: {
  componente: Componente;
  onAplicar: (campo: string, valor: string) => void;
}) {
  const src = typeof componente.propriedades?.src === 'string' ? componente.propriedades.src : '';
  return (
    <Grupo titulo="Vídeo">
      <Campo label="URL (YouTube, Vimeo ou arquivo .mp4)">
        <input
          value={src}
          onChange={(e) => onAplicar('src', e.target.value)}
          placeholder="https://…"
          className={campoClasse}
        />
      </Campo>
    </Grupo>
  );
}

interface ItemTab {
  titulo: string;
  conteudo?: string;
}

/** Conteúdo do componente Tabs: lista de abas (título + conteúdo) — mesma
 * estrutura do Accordion, só que só um painel fica visível por vez. */
function SecaoTabs({
  componente,
  onAplicar,
}: {
  componente: Componente;
  onAplicar: (campo: string, valor: unknown) => void;
}) {
  const itens: ItemTab[] = Array.isArray(componente.propriedades?.itens)
    ? (componente.propriedades.itens as ItemTab[])
    : [];

  function atualizarItem(indice: number, campo: keyof ItemTab, valor: string) {
    onAplicar(
      'itens',
      itens.map((item, i) => (i === indice ? { ...item, [campo]: valor } : item)),
    );
  }

  function removerItem(indice: number) {
    onAplicar(
      'itens',
      itens.filter((_, i) => i !== indice),
    );
  }

  function adicionarItem() {
    onAplicar('itens', [...itens, { titulo: 'Nova aba', conteudo: '' }]);
  }

  return (
    <Grupo titulo="Tabs">
      {itens.length > 0 && (
        <ul className="flex flex-col gap-2">
          {itens.map((item, i) => (
            <li key={i} className="flex flex-col gap-1 border border-border rounded-lg p-2">
              <div className="flex items-center gap-1">
                <input
                  value={item.titulo}
                  onChange={(e) => atualizarItem(i, 'titulo', e.target.value)}
                  placeholder="Título"
                  aria-label={`Título da aba ${i + 1}`}
                  className={campoClasse}
                />
                <button
                  type="button"
                  onClick={() => removerItem(i)}
                  aria-label={`Remover aba ${i + 1}`}
                  className="text-xs text-destructive shrink-0 px-1"
                >
                  Remover
                </button>
              </div>
              <textarea
                value={item.conteudo ?? ''}
                onChange={(e) => atualizarItem(i, 'conteudo', e.target.value)}
                placeholder="Conteúdo"
                aria-label={`Conteúdo da aba ${i + 1}`}
                rows={2}
                className="text-xs px-2 py-1.5 bg-background border border-border rounded-lg text-foreground"
              />
            </li>
          ))}
        </ul>
      )}
      <button
        type="button"
        onClick={adicionarItem}
        className="self-start text-xs bg-primary text-primary-foreground px-3 py-1.5 rounded-lg font-medium hover:bg-primary/90"
      >
        + Adicionar aba
      </button>
    </Grupo>
  );
}

/** Conteúdo do componente Avaliação: 5 estrelas clicáveis para definir 0-5. */
function SecaoAvaliacao({
  componente,
  onAplicar,
}: {
  componente: Componente;
  onAplicar: (campo: string, valor: unknown) => void;
}) {
  const valor = typeof componente.propriedades?.valor === 'number' ? componente.propriedades.valor : 0;
  return (
    <Grupo titulo="Avaliação">
      <Campo label={`Estrelas (${valor}/5)`}>
        <div className="flex items-center gap-1">
          {[1, 2, 3, 4, 5].map((n) => (
            <button
              key={n}
              type="button"
              onClick={() => onAplicar('valor', n === valor ? 0 : n)}
              aria-label={`${n} estrela${n > 1 ? 's' : ''}`}
              aria-pressed={n <= valor}
              className="text-lg leading-none"
            >
              {n <= valor ? '★' : '☆'}
            </button>
          ))}
        </div>
      </Campo>
    </Grupo>
  );
}

/** Conteúdo do componente Ícone: escolha de um nome dentro do catálogo
 * fechado (services/frontend/src/systems/iconePicker.ts). */
function SecaoIcone({
  componente,
  onAplicar,
}: {
  componente: Componente;
  onAplicar: (campo: string, valor: string) => void;
}) {
  const valor = typeof componente.propriedades?.icone === 'string' ? componente.propriedades.icone : NOME_ICONE_PADRAO;
  return (
    <Grupo titulo="Ícone">
      <Campo label="Ícone">
        <select value={valor} onChange={(e) => onAplicar('icone', e.target.value)} className={campoClasse}>
          {Object.keys(ICONES_DISPONIVEIS).map((nome) => (
            <option key={nome} value={nome}>
              {nome}
            </option>
          ))}
        </select>
      </Campo>
    </Grupo>
  );
}

/** Conteúdo do componente Progresso: valor percentual 0-100 (a cor do
 * preenchimento reaproveita "Cor de fundo" da seção Background, mais abaixo). */
function SecaoProgresso({
  componente,
  onAplicar,
}: {
  componente: Componente;
  onAplicar: (campo: string, valor: unknown) => void;
}) {
  const valor = typeof componente.propriedades?.valor === 'number' ? componente.propriedades.valor : 50;
  return (
    <Grupo titulo="Progresso">
      <Campo label={`Valor (${valor}%)`}>
        <input
          type="range"
          min={0}
          max={100}
          step={1}
          value={valor}
          onChange={(e) => onAplicar('valor', Number(e.target.value))}
          className="w-full"
        />
      </Campo>
    </Grupo>
  );
}

interface ItemBreadcrumb {
  label: string;
  url?: string;
}

/** Conteúdo do componente Breadcrumb: mesmo shape do Menu (label + link). */
function SecaoBreadcrumb({
  componente,
  onAplicar,
}: {
  componente: Componente;
  onAplicar: (campo: string, valor: unknown) => void;
}) {
  const itens: ItemBreadcrumb[] = Array.isArray(componente.propriedades?.itens)
    ? (componente.propriedades.itens as ItemBreadcrumb[])
    : [];

  function atualizarItem(indice: number, campo: keyof ItemBreadcrumb, valor: string) {
    onAplicar(
      'itens',
      itens.map((item, i) => (i === indice ? { ...item, [campo]: valor } : item)),
    );
  }

  function removerItem(indice: number) {
    onAplicar(
      'itens',
      itens.filter((_, i) => i !== indice),
    );
  }

  function adicionarItem() {
    onAplicar('itens', [...itens, { label: 'Nova página', url: '#' }]);
  }

  return (
    <Grupo titulo="Breadcrumb">
      {itens.length > 0 && (
        <ul className="flex flex-col gap-2">
          {itens.map((item, i) => (
            <li key={i} className="flex flex-col gap-1 border border-border rounded-lg p-2">
              <div className="flex items-center gap-1">
                <input
                  value={item.label}
                  onChange={(e) => atualizarItem(i, 'label', e.target.value)}
                  placeholder="Rótulo"
                  aria-label={`Rótulo do item ${i + 1}`}
                  className={campoClasse}
                />
                <button
                  type="button"
                  onClick={() => removerItem(i)}
                  aria-label={`Remover item ${i + 1}`}
                  className="text-xs text-destructive shrink-0 px-1"
                >
                  Remover
                </button>
              </div>
              <input
                value={item.url ?? ''}
                onChange={(e) => atualizarItem(i, 'url', e.target.value)}
                placeholder="URL (ex.: /produtos)"
                aria-label={`URL do item ${i + 1}`}
                className={campoClasse}
              />
            </li>
          ))}
        </ul>
      )}
      <button
        type="button"
        onClick={adicionarItem}
        className="self-start text-xs bg-primary text-primary-foreground px-3 py-1.5 rounded-lg font-medium hover:bg-primary/90"
      >
        + Adicionar item
      </button>
    </Grupo>
  );
}

/** Conteúdo do componente Toggle: um único campo — ligado/desligado. */
function SecaoToggle({
  componente,
  onAplicar,
}: {
  componente: Componente;
  onAplicar: (campo: string, valor: unknown) => void;
}) {
  const ativo = componente.propriedades?.ativo === true;
  return (
    <Grupo titulo="Toggle">
      <label className="flex items-center gap-2 text-xs">
        <input type="checkbox" checked={ativo} onChange={(e) => onAplicar('ativo', e.target.checked)} />
        <span>Ativado</span>
      </label>
    </Grupo>
  );
}

const VARIANTES_ALERTA = ['info', 'sucesso', 'aviso', 'erro'] as const;

/** Conteúdo do componente Alerta: variante info/sucesso/aviso/erro. */
function SecaoAlerta({
  componente,
  onAplicar,
}: {
  componente: Componente;
  onAplicar: (campo: string, valor: unknown) => void;
}) {
  const variante = typeof componente.propriedades?.variante === 'string' ? componente.propriedades.variante : 'info';
  return (
    <Grupo titulo="Alerta">
      <Campo label="Variante">
        <Selecao valor={variante} opcoes={[...VARIANTES_ALERTA]} onChange={(v) => onAplicar('variante', v)} />
      </Campo>
    </Grupo>
  );
}

/** Conteúdo do componente Avatar: override opcional de imagem — sem src, as
 * iniciais (campo "Texto" da seção Conteúdo, genérica) seguem valendo. */
function SecaoAvatar({
  componente,
  onAplicar,
}: {
  componente: Componente;
  onAplicar: (campo: string, valor: string) => void;
}) {
  const [erro, setErro] = useState<string | null>(null);
  const src = typeof componente.propriedades?.src === 'string' ? componente.propriedades.src : '';

  function onArquivoSelecionado(e: ChangeEvent<HTMLInputElement>) {
    const arquivo = e.target.files?.[0];
    e.target.value = '';
    if (!arquivo) return;
    if (arquivo.size > TAMANHO_MAX_IMAGEM_BYTES) {
      setErro('Imagem muito grande (máx. 2MB).');
      return;
    }
    const leitor = new FileReader();
    leitor.onload = () => {
      if (typeof leitor.result === 'string') {
        setErro(null);
        onAplicar('src', leitor.result);
      }
    };
    leitor.onerror = () => setErro('Não foi possível ler o arquivo.');
    leitor.readAsDataURL(arquivo);
  }

  return (
    <Grupo titulo="Avatar">
      {src && <img src={src} alt="" className="w-12 h-12 rounded-full object-cover border border-border" />}
      <Campo label="Enviar imagem">
        <input
          type="file"
          accept="image/*"
          onChange={onArquivoSelecionado}
          className="text-xs text-muted-foreground file:mr-2 file:px-2 file:py-1 file:rounded-lg file:border file:border-border file:bg-secondary file:text-secondary-foreground file:text-xs"
        />
      </Campo>
      {erro && <p className="text-xs text-destructive">{erro}</p>}
      <Campo label="ou URL da imagem">
        <input
          value={src}
          onChange={(e) => onAplicar('src', e.target.value)}
          placeholder="https://…"
          className={campoClasse}
        />
      </Campo>
    </Grupo>
  );
}

function SecaoAvancada({
  componente,
  onAtualizar,
}: {
  componente: Componente;
  onAtualizar: (blindIndex: string, propriedades: Record<string, unknown>) => void;
}) {
  const [aberta, setAberta] = useState(false);
  const [rascunho, setRascunho] = useState<string | null>(null);
  const [erro, setErro] = useState<string | null>(null);

  const valor = rascunho ?? JSON.stringify(componente.propriedades ?? {}, null, 2);

  function aplicar() {
    try {
      onAtualizar(componente.blind_index, JSON.parse(valor || '{}'));
      setErro(null);
      setRascunho(null);
    } catch {
      setErro('JSON inválido.');
    }
  }

  return (
    <div className="flex flex-col gap-2 border-t border-border pt-3">
      <button
        type="button"
        onClick={() => setAberta((a) => !a)}
        className="text-xs font-heading font-bold text-muted-foreground uppercase tracking-wide text-left"
      >
        Avançado {aberta ? '▾' : '▸'}
      </button>
      {aberta && (
        <>
          <textarea
            value={valor}
            onChange={(e) => setRascunho(e.target.value)}
            rows={8}
            className="text-xs font-mono px-2 py-1.5 bg-background border border-border rounded-lg"
          />
          {erro && <p className="text-xs text-destructive">{erro}</p>}
          <button
            type="button"
            onClick={aplicar}
            className="self-start text-xs bg-primary text-primary-foreground px-3 py-1.5 rounded-lg font-medium hover:bg-primary/90"
          >
            Aplicar JSON
          </button>
        </>
      )}
    </div>
  );
}
