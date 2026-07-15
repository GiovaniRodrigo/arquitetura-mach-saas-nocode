// Componentes de estado de UI reutilizáveis (M3) para as fases de carregamento,
// vazio e erro consumidas pelas telas do dashboard e pelo seletor de sistemas
// (RF06, RNF03). Padronizam `aria-busy` (loading) e `role="alert"` (erro).

import type { ReactNode } from 'react';
import { RefreshCcw } from 'lucide-react';

/** Skeleton em grade — placeholder animado durante o carregamento (aria-busy). */
export function Skeleton({ itens = 3, className = '' }: { itens?: number; className?: string }) {
  return (
    <ul
      aria-busy="true"
      aria-live="polite"
      className={`grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6 ${className}`}
    >
      {Array.from({ length: itens }).map((_, i) => (
        <li
          key={i}
          className="h-40 bg-card border border-border rounded-3xl animate-pulse"
        />
      ))}
    </ul>
  );
}

/** Estado vazio com título, descrição e ação opcional (ex.: CTA de criação). */
export function EmptyState({
  titulo,
  descricao,
  acao,
}: {
  titulo: string;
  descricao?: string;
  acao?: ReactNode;
}) {
  return (
    <div className="flex flex-col items-center justify-center text-center bg-card/50 border border-border rounded-3xl p-12 gap-3">
      <h3 className="text-lg font-heading font-bold text-foreground">{titulo}</h3>
      {descricao && <p className="text-sm text-muted-foreground max-w-md">{descricao}</p>}
      {acao && <div className="mt-2">{acao}</div>}
    </div>
  );
}

/** Estado de erro com alerta acessível e ação de repetir (RF06). */
export function ErrorState({
  mensagem,
  onRepetir,
}: {
  mensagem: string;
  onRepetir: () => void;
}) {
  return (
    <div
      role="alert"
      className="bg-destructive/10 border border-destructive/20 text-destructive p-6 rounded-3xl flex flex-col items-start gap-4"
    >
      <p className="font-medium">{mensagem}</p>
      <button
        type="button"
        onClick={onRepetir}
        className="flex items-center px-4 py-2 bg-background border border-border rounded-full hover:bg-secondary transition-colors text-foreground"
      >
        <RefreshCcw className="w-4 h-4 mr-2" />
        Tentar novamente
      </button>
    </div>
  );
}
