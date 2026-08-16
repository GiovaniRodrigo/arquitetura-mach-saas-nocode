// Estado do onboarding guiado: mostra automaticamente o tour da tela no
// primeiro acesso (por tela, lembrado em localStorage) e permite reabri-lo a
// qualquer momento pelo ícone de ajuda do cabeçalho (ver DashboardLayout).

import { useCallback, useEffect, useState } from 'react';
import { useLocation } from 'react-router-dom';
import { telaOnboardingDe, type TelaOnboarding } from './conteudo';

/** Exportado para testes que precisam simular um usuário que já viu o onboarding. */
export const CHAVE_STORAGE_ONBOARDING = 'mach-onboarding-vistos';

function lerVistos(): Set<string> {
  try {
    const bruto = window.localStorage.getItem(CHAVE_STORAGE_ONBOARDING);
    return new Set(bruto ? (JSON.parse(bruto) as string[]) : []);
  } catch {
    return new Set();
  }
}

function marcarVisto(chave: string) {
  try {
    const vistos = lerVistos();
    vistos.add(chave);
    window.localStorage.setItem(CHAVE_STORAGE_ONBOARDING, JSON.stringify([...vistos]));
  } catch {
    // localStorage indisponível (ex.: modo privado) — onboarding volta a
    // aparecer a cada visita, sem quebrar a navegação.
  }
}

export interface UseOnboarding {
  tela: TelaOnboarding | null;
  aberto: boolean;
  abrirManual: () => void;
  fechar: () => void;
}

export function useOnboarding(): UseOnboarding {
  const location = useLocation();
  const tela = telaOnboardingDe(location.pathname);
  const [aberto, setAberto] = useState(false);

  useEffect(() => {
    if (!tela) {
      setAberto(false);
      return;
    }
    if (!lerVistos().has(tela.chave)) {
      setAberto(true);
      marcarVisto(tela.chave);
    } else {
      setAberto(false);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [tela?.chave]);

  const abrirManual = useCallback(() => setAberto(true), []);
  const fechar = useCallback(() => setAberto(false), []);

  return { tela, aberto, abrirManual, fechar };
}
