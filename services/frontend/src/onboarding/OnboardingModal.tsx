import { useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import type { TelaOnboarding } from './conteudo';

export interface OnboardingModalProps {
  tela: TelaOnboarding | null;
  aberto: boolean;
  onFechar: () => void;
}

export function OnboardingModal({ tela, aberto, onFechar }: OnboardingModalProps) {
  const [passo, setPasso] = useState(0);

  useEffect(() => {
    if (aberto) setPasso(0);
  }, [aberto, tela?.chave]);

  if (!tela) return null;

  const atual = tela.passos[passo];
  const ultimoPasso = passo === tela.passos.length - 1;

  return (
    <Dialog open={aberto} onOpenChange={(aberto) => { if (!aberto) onFechar(); }}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>{tela.titulo}</DialogTitle>
        </DialogHeader>

        <div className="flex flex-col gap-1.5">
          <h4 className="font-heading font-semibold text-foreground">{atual.titulo}</h4>
          <DialogDescription>{atual.descricao}</DialogDescription>
        </div>

        <div className="flex items-center justify-center gap-1.5" aria-hidden="true">
          {tela.passos.map((_, indice) => (
            <span
              key={indice}
              className={`h-1.5 rounded-full transition-all ${
                indice === passo ? 'w-6 bg-primary' : 'w-1.5 bg-border'
              }`}
            />
          ))}
        </div>

        <DialogFooter className="sm:justify-between">
          <Button type="button" variant="ghost" size="sm" onClick={() => setPasso((p) => p - 1)} disabled={passo === 0}>
            Anterior
          </Button>
          <Button type="button" size="sm" onClick={() => (ultimoPasso ? onFechar() : setPasso((p) => p + 1))}>
            {ultimoPasso ? 'Concluir' : 'Próximo'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
