// Aplicação síncrona do tema salvo, executada no boot (antes do primeiro paint)
// para evitar flash de tema incorreto — FOUC (RNF04). O Tailwind usa a estratégia
// de classe `dark` no <html>.

export type Tema = 'claro' | 'escuro';
export const CHAVE_TEMA = 'mach_theme';

/** Lê a preferência salva; default "escuro" (dark-first, alinhado ao design). */
export function temaSalvo(): Tema {
  try {
    return localStorage.getItem(CHAVE_TEMA) === 'claro' ? 'claro' : 'escuro';
  } catch {
    return 'escuro';
  }
}

/** Aplica/retira a classe `dark` no <html> conforme o tema. */
export function aplicarTema(tema: Tema): void {
  document.documentElement.classList.toggle('dark', tema === 'escuro');
}

/** Chamado uma vez no boot, antes de renderizar a App. */
export function initTheme(): void {
  aplicarTema(temaSalvo());
}
