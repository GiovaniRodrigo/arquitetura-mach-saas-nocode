// Catálogo de tipos de componente do canvas (RF09) e o modelo de estilos que
// cada nó guarda em `propriedades.estilos`. Não há mudança de schema no
// backend — `propriedades` continua sendo JSON livre — isto é só a forma que
// o Frontend organiza esse JSON para ter um inspector estruturado (Layout,
// Typography, Background/Border/Shadow) em vez de um blob solto.

import {
  LayoutGrid,
  Square,
  Heading1,
  Pilcrow,
  MousePointerClick,
  Link as LinkIcon,
  TextCursorInput,
  ChevronsUpDown,
  SquareCheck,
  Image as ImageIcon,
  GalleryHorizontal,
  PanelTop,
  PanelBottom,
  PanelLeft,
  PanelRight,
  LayoutTemplate,
  Menu as MenuIcon,
  ListCollapse,
  CreditCard,
  SeparatorHorizontal,
  Badge as BadgeIcon,
  Video as VideoIcon,
  GalleryThumbnails,
  Star,
  Shapes,
  Gauge,
  CircleUserRound,
  TextCursor,
  Circle,
  ToggleLeft,
  ChevronsRight,
  Info,
  LoaderCircle,
  RectangleHorizontal,
  type LucideIcon,
} from 'lucide-react';

export interface Estilos {
  posicao?: 'static' | 'relative' | 'absolute';
  x?: number;
  y?: number;
  largura?: string;
  altura?: string;
  padding?: string;
  margem?: string;
  display?: 'block' | 'inline' | 'inline-block' | 'flex' | 'none';
  direcao?: 'row' | 'column';
  justificar?: 'flex-start' | 'center' | 'flex-end' | 'space-between' | 'space-around';
  alinhar?: 'flex-start' | 'center' | 'flex-end' | 'stretch';
  espacamento?: string;
  fonteTamanho?: string;
  fontePeso?: 'normal' | 'medium' | 'bold';
  cor?: string;
  textoAlinhar?: 'left' | 'center' | 'right';
  fundoCor?: string;
  bordaLargura?: string;
  bordaCor?: string;
  bordaRaio?: string;
  sombra?: string;
  opacidade?: number;
}

export type Categoria = 'layout' | 'texto' | 'formulario' | 'midia';

export interface RegistroComponente {
  tipo: string;
  rotulo: string;
  categoria: Categoria;
  aceitaFilhos: boolean;
  /** Exibe o campo de conteúdo textual e a seção Typography no inspector. */
  temTexto: boolean;
  propriedadesPadrao: () => { texto?: string; estilos: Estilos };
  icone: LucideIcon;
}

const ESTILOS_BASE: Estilos = { posicao: 'relative', x: 0, y: 0 };

export const CATEGORIAS: { id: Categoria; rotulo: string }[] = [
  { id: 'layout', rotulo: 'Layout' },
  { id: 'texto', rotulo: 'Texto' },
  { id: 'formulario', rotulo: 'Formulário' },
  { id: 'midia', rotulo: 'Mídia' },
];

export const REGISTRO_COMPONENTES: RegistroComponente[] = [
  {
    tipo: 'section',
    rotulo: 'Section',
    categoria: 'layout',
    aceitaFilhos: true,
    temTexto: false,
    icone: LayoutGrid,
    propriedadesPadrao: () => ({
      estilos: { ...ESTILOS_BASE, display: 'flex', direcao: 'column', espacamento: '16px', padding: '24px' },
    }),
  },
  {
    tipo: 'container',
    rotulo: 'Container',
    categoria: 'layout',
    aceitaFilhos: true,
    temTexto: false,
    icone: Square,
    propriedadesPadrao: () => ({
      estilos: { ...ESTILOS_BASE, display: 'flex', direcao: 'row', espacamento: '8px', padding: '12px' },
    }),
  },
  {
    tipo: 'header',
    rotulo: 'Header',
    categoria: 'layout',
    aceitaFilhos: true,
    temTexto: false,
    icone: PanelTop,
    propriedadesPadrao: () => ({
      estilos: {
        ...ESTILOS_BASE,
        display: 'flex',
        direcao: 'row',
        justificar: 'space-between',
        alinhar: 'center',
        padding: '16px 24px',
        largura: '100%',
      },
    }),
  },
  {
    tipo: 'footer',
    rotulo: 'Footer',
    categoria: 'layout',
    aceitaFilhos: true,
    temTexto: false,
    icone: PanelBottom,
    propriedadesPadrao: () => ({
      estilos: {
        ...ESTILOS_BASE,
        display: 'flex',
        direcao: 'row',
        justificar: 'center',
        alinhar: 'center',
        padding: '16px 24px',
        largura: '100%',
      },
    }),
  },
  {
    tipo: 'sidebar',
    rotulo: 'Sidebar',
    categoria: 'layout',
    aceitaFilhos: true,
    temTexto: false,
    icone: PanelLeft,
    propriedadesPadrao: () => ({
      estilos: {
        ...ESTILOS_BASE,
        display: 'flex',
        direcao: 'column',
        espacamento: '8px',
        padding: '16px',
        largura: '240px',
        altura: '100%',
      },
    }),
  },
  {
    tipo: 'rightbar',
    rotulo: 'Rightbar',
    categoria: 'layout',
    aceitaFilhos: true,
    temTexto: false,
    icone: PanelRight,
    propriedadesPadrao: () => ({
      estilos: {
        ...ESTILOS_BASE,
        display: 'flex',
        direcao: 'column',
        espacamento: '8px',
        padding: '16px',
        largura: '280px',
        altura: '100%',
      },
    }),
  },
  {
    tipo: 'main',
    rotulo: 'Main',
    categoria: 'layout',
    aceitaFilhos: true,
    temTexto: false,
    icone: LayoutTemplate,
    propriedadesPadrao: () => ({
      estilos: { ...ESTILOS_BASE, display: 'flex', direcao: 'column', espacamento: '16px', padding: '24px', largura: '100%' },
    }),
  },
  {
    tipo: 'menu',
    rotulo: 'Menu',
    categoria: 'layout',
    aceitaFilhos: false,
    temTexto: false,
    icone: MenuIcon,
    propriedadesPadrao: () => ({
      estilos: { ...ESTILOS_BASE, display: 'flex', direcao: 'row', alinhar: 'center', espacamento: '24px' },
    }),
  },
  {
    tipo: 'accordion',
    rotulo: 'Accordion',
    categoria: 'layout',
    aceitaFilhos: false,
    temTexto: false,
    icone: ListCollapse,
    propriedadesPadrao: () => ({
      estilos: { ...ESTILOS_BASE, display: 'flex', direcao: 'column', espacamento: '4px', largura: '100%' },
    }),
  },
  {
    tipo: 'card',
    rotulo: 'Card',
    categoria: 'layout',
    aceitaFilhos: true,
    temTexto: false,
    icone: CreditCard,
    propriedadesPadrao: () => ({
      estilos: {
        ...ESTILOS_BASE,
        display: 'flex',
        direcao: 'column',
        espacamento: '12px',
        padding: '20px',
        bordaRaio: '12px',
        bordaLargura: '1px',
        bordaCor: '#e5e7eb',
        sombra: '0 1px 3px rgba(0,0,0,.1)',
        fundoCor: '#ffffff',
      },
    }),
  },
  {
    tipo: 'divisor',
    rotulo: 'Divisor',
    categoria: 'layout',
    aceitaFilhos: false,
    temTexto: false,
    icone: SeparatorHorizontal,
    propriedadesPadrao: () => ({
      estilos: { ...ESTILOS_BASE, largura: '100%', altura: '1px', fundoCor: '#e5e7eb', margem: '16px 0' },
    }),
  },
  {
    tipo: 'tabs',
    rotulo: 'Tabs',
    categoria: 'layout',
    aceitaFilhos: false,
    temTexto: false,
    icone: GalleryThumbnails,
    propriedadesPadrao: () => ({
      estilos: { ...ESTILOS_BASE, display: 'flex', direcao: 'column', largura: '100%' },
    }),
  },
  {
    tipo: 'progresso',
    rotulo: 'Progresso',
    categoria: 'layout',
    aceitaFilhos: false,
    temTexto: false,
    icone: Gauge,
    propriedadesPadrao: () => ({
      estilos: { ...ESTILOS_BASE, largura: '100%', altura: '8px', bordaRaio: '999px', fundoCor: '#6366f1' },
    }),
  },
  {
    tipo: 'breadcrumb',
    rotulo: 'Breadcrumb',
    categoria: 'layout',
    aceitaFilhos: false,
    temTexto: false,
    icone: ChevronsRight,
    propriedadesPadrao: () => ({
      estilos: { ...ESTILOS_BASE, display: 'flex', direcao: 'row', alinhar: 'center', espacamento: '6px' },
    }),
  },
  {
    tipo: 'spinner',
    rotulo: 'Spinner',
    categoria: 'layout',
    aceitaFilhos: false,
    temTexto: false,
    icone: LoaderCircle,
    propriedadesPadrao: () => ({ estilos: { ...ESTILOS_BASE, largura: '32px', altura: '32px', fundoCor: '#6366f1' } }),
  },
  {
    tipo: 'skeleton',
    rotulo: 'Skeleton',
    categoria: 'layout',
    aceitaFilhos: false,
    temTexto: false,
    icone: RectangleHorizontal,
    propriedadesPadrao: () => ({
      estilos: { ...ESTILOS_BASE, largura: '100%', altura: '16px', bordaRaio: '6px', fundoCor: '#e5e7eb' },
    }),
  },
  {
    tipo: 'heading',
    rotulo: 'Título',
    categoria: 'texto',
    aceitaFilhos: false,
    temTexto: true,
    icone: Heading1,
    propriedadesPadrao: () => ({
      texto: 'Título',
      estilos: { ...ESTILOS_BASE, fonteTamanho: '32px', fontePeso: 'bold' },
    }),
  },
  {
    tipo: 'paragrafo',
    rotulo: 'Parágrafo',
    categoria: 'texto',
    aceitaFilhos: false,
    temTexto: true,
    icone: Pilcrow,
    propriedadesPadrao: () => ({
      texto: 'Texto do parágrafo',
      estilos: { ...ESTILOS_BASE, fonteTamanho: '16px', fontePeso: 'normal' },
    }),
  },
  {
    tipo: 'botao',
    rotulo: 'Botão',
    categoria: 'texto',
    aceitaFilhos: false,
    temTexto: true,
    icone: MousePointerClick,
    propriedadesPadrao: () => ({
      texto: 'Botão',
      estilos: {
        ...ESTILOS_BASE,
        fonteTamanho: '14px',
        fontePeso: 'medium',
        fundoCor: '#6366f1',
        cor: '#ffffff',
        padding: '8px 16px',
        bordaRaio: '8px',
      },
    }),
  },
  {
    tipo: 'link',
    rotulo: 'Link',
    categoria: 'texto',
    aceitaFilhos: false,
    temTexto: true,
    icone: LinkIcon,
    propriedadesPadrao: () => ({
      texto: 'Link',
      estilos: { ...ESTILOS_BASE, cor: '#6366f1', fonteTamanho: '14px' },
    }),
  },
  {
    tipo: 'badge',
    rotulo: 'Badge',
    categoria: 'texto',
    aceitaFilhos: false,
    temTexto: true,
    icone: BadgeIcon,
    propriedadesPadrao: () => ({
      texto: 'Novo',
      estilos: {
        ...ESTILOS_BASE,
        fonteTamanho: '12px',
        fontePeso: 'medium',
        cor: '#ffffff',
        fundoCor: '#6366f1',
        padding: '2px 10px',
        bordaRaio: '999px',
      },
    }),
  },
  {
    tipo: 'alerta',
    rotulo: 'Alerta',
    categoria: 'texto',
    aceitaFilhos: false,
    temTexto: true,
    icone: Info,
    propriedadesPadrao: () => ({
      texto: 'Mensagem informativa',
      estilos: {
        ...ESTILOS_BASE,
        fonteTamanho: '14px',
        padding: '10px 14px',
        bordaRaio: '8px',
        bordaLargura: '1px',
      },
    }),
  },
  {
    tipo: 'input',
    rotulo: 'Input',
    categoria: 'formulario',
    aceitaFilhos: false,
    temTexto: false,
    icone: TextCursorInput,
    propriedadesPadrao: () => ({ estilos: { ...ESTILOS_BASE, largura: '240px' } }),
  },
  {
    tipo: 'select',
    rotulo: 'Select',
    categoria: 'formulario',
    aceitaFilhos: false,
    temTexto: false,
    icone: ChevronsUpDown,
    propriedadesPadrao: () => ({ estilos: { ...ESTILOS_BASE, largura: '240px' } }),
  },
  {
    tipo: 'checkbox',
    rotulo: 'Checkbox',
    categoria: 'formulario',
    aceitaFilhos: false,
    temTexto: true,
    icone: SquareCheck,
    propriedadesPadrao: () => ({ texto: 'Opção', estilos: { ...ESTILOS_BASE } }),
  },
  {
    tipo: 'radio',
    rotulo: 'Radio',
    categoria: 'formulario',
    aceitaFilhos: false,
    temTexto: true,
    icone: Circle,
    propriedadesPadrao: () => ({ texto: 'Opção', estilos: { ...ESTILOS_BASE } }),
  },
  {
    tipo: 'textarea',
    rotulo: 'Textarea',
    categoria: 'formulario',
    aceitaFilhos: false,
    temTexto: false,
    icone: TextCursor,
    propriedadesPadrao: () => ({ estilos: { ...ESTILOS_BASE, largura: '240px', altura: '80px' } }),
  },
  {
    tipo: 'toggle',
    rotulo: 'Toggle',
    categoria: 'formulario',
    aceitaFilhos: false,
    temTexto: false,
    icone: ToggleLeft,
    propriedadesPadrao: () => ({ estilos: { ...ESTILOS_BASE, largura: '36px', altura: '20px', fundoCor: '#6366f1' } }),
  },
  {
    tipo: 'avaliacao',
    rotulo: 'Avaliação',
    categoria: 'formulario',
    aceitaFilhos: false,
    temTexto: false,
    icone: Star,
    propriedadesPadrao: () => ({ estilos: { ...ESTILOS_BASE, fonteTamanho: '20px', cor: '#f59e0b' } }),
  },
  {
    tipo: 'imagem',
    rotulo: 'Imagem',
    categoria: 'midia',
    aceitaFilhos: false,
    temTexto: false,
    icone: ImageIcon,
    propriedadesPadrao: () => ({ estilos: { ...ESTILOS_BASE, largura: '200px', altura: '150px' } }),
  },
  {
    tipo: 'carrossel',
    rotulo: 'Carrossel',
    categoria: 'midia',
    aceitaFilhos: false,
    temTexto: false,
    icone: GalleryHorizontal,
    propriedadesPadrao: () => ({ estilos: { ...ESTILOS_BASE, largura: '100%', altura: '280px' } }),
  },
  {
    tipo: 'video',
    rotulo: 'Vídeo',
    categoria: 'midia',
    aceitaFilhos: false,
    temTexto: false,
    icone: VideoIcon,
    propriedadesPadrao: () => ({ estilos: { ...ESTILOS_BASE, largura: '100%', altura: '280px' } }),
  },
  {
    tipo: 'icone',
    rotulo: 'Ícone',
    categoria: 'midia',
    aceitaFilhos: false,
    temTexto: false,
    icone: Shapes,
    propriedadesPadrao: () => ({ estilos: { ...ESTILOS_BASE, largura: '32px', altura: '32px', cor: '#111827' } }),
  },
  {
    tipo: 'avatar',
    rotulo: 'Avatar',
    categoria: 'midia',
    aceitaFilhos: false,
    temTexto: true,
    icone: CircleUserRound,
    propriedadesPadrao: () => ({
      texto: 'AB',
      estilos: {
        ...ESTILOS_BASE,
        largura: '48px',
        altura: '48px',
        bordaRaio: '999px',
        fundoCor: '#6366f1',
        cor: '#ffffff',
        fonteTamanho: '16px',
        fontePeso: 'medium',
        display: 'flex',
        justificar: 'center',
        alinhar: 'center',
      },
    }),
  },
];

export function registroDoTipo(tipo: string): RegistroComponente | undefined {
  return REGISTRO_COMPONENTES.find((r) => r.tipo === tipo);
}

export function aceitaFilhos(tipo: string): boolean {
  return registroDoTipo(tipo)?.aceitaFilhos ?? false;
}
