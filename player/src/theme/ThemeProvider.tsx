// Contexto de tema: expõe o tema atual e as ações de alternância/definição,
// persistindo em localStorage e aplicando a classe `dark` no <html> (RF05/RNF04).

import {
  createContext,
  useContext,
  useState,
  useCallback,
  useEffect,
  useRef,
  type ReactNode,
} from 'react';
import { type Tema, CHAVE_TEMA, temaSalvo, aplicarTema } from './initTheme';

export interface ThemeContextValue {
  tema: Tema;
  alternarTema: () => void;
  definirTema: (t: Tema) => void;
}

const ThemeContext = createContext<ThemeContextValue | null>(null);

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [tema, setTema] = useState<Tema>(() => temaSalvo());

  // Efeitos colaterais (classe no <html> + persistência) reagem à mudança de
  // estado, mantendo os setters puros. A persistência é pulada no primeiro
  // render para não gravar antes de uma escolha explícita do usuário.
  const primeiraRenderizacao = useRef(true);
  useEffect(() => {
    aplicarTema(tema);
    if (primeiraRenderizacao.current) {
      primeiraRenderizacao.current = false;
      return;
    }
    try {
      localStorage.setItem(CHAVE_TEMA, tema);
    } catch {
      /* localStorage indisponível — tema permanece só na sessão */
    }
  }, [tema]);

  const definirTema = useCallback((t: Tema) => setTema(t), []);
  // Updater funcional: alterna sempre a partir do valor mais recente, imune a
  // closures obsoletos mesmo em cliques rápidos sucessivos.
  const alternarTema = useCallback(
    () => setTema((atual) => (atual === 'escuro' ? 'claro' : 'escuro')),
    [],
  );

  return (
    <ThemeContext.Provider value={{ tema, alternarTema, definirTema }}>
      {children}
    </ThemeContext.Provider>
  );
}

/** Acessa o tema; lança se usado fora do ThemeProvider. */
export function useTheme(): ThemeContextValue {
  const ctx = useContext(ThemeContext);
  if (!ctx) throw new Error('useTheme deve ser usado dentro de <ThemeProvider>');
  return ctx;
}
