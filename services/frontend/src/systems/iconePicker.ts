// Catálogo fechado de ícones selecionáveis pelo componente "Ícone" (RF09):
// como propriedades.icone guarda só o nome (string), mapeamos para o
// componente real do lucide-react aqui — um único lugar usado por Canvas,
// Inspector e PreviewRenderer.

import {
  Star,
  Heart,
  Check,
  ArrowRight,
  Mail,
  Phone,
  MapPin,
  ShieldCheck,
  Zap,
  Smile,
  ThumbsUp,
  Search,
  type LucideIcon,
} from 'lucide-react';

export const ICONES_DISPONIVEIS: Record<string, LucideIcon> = {
  Star,
  Heart,
  Check,
  ArrowRight,
  Mail,
  Phone,
  MapPin,
  ShieldCheck,
  Zap,
  Smile,
  ThumbsUp,
  Search,
};

export const NOME_ICONE_PADRAO = 'Star';

export function iconePorNome(nome: string | undefined): LucideIcon {
  return (nome && ICONES_DISPONIVEIS[nome]) || ICONES_DISPONIVEIS[NOME_ICONE_PADRAO];
}
