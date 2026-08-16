// Formata competência "AAAA-MM" (contrato dos gráficos mensais do Dashboard,
// ver useAcessosPorMes/useReceitaPorMes) para rótulo curto em pt-BR (ex.: "Ago").

const MESES_ABREVIADOS = [
  'Jan', 'Fev', 'Mar', 'Abr', 'Mai', 'Jun',
  'Jul', 'Ago', 'Set', 'Out', 'Nov', 'Dez',
];

export function mesAbreviado(competencia: string): string {
  const mes = Number(competencia.slice(5, 7));
  return MESES_ABREVIADOS[mes - 1] ?? competencia;
}

/**
 * Variação percentual entre o primeiro e o último ponto de uma série mensal
 * (mesma leitura do badge "+35%"/"+34%" dos KPIs de gráfico). `null` quando
 * não há como comparar (série vazia ou só um ponto).
 */
export function variacaoPercentual(valores: number[]): number | null {
  if (valores.length < 2) return null;
  const primeiro = valores[0];
  const ultimo = valores[valores.length - 1];
  if (primeiro === 0) return ultimo === 0 ? 0 : 100;
  return Math.round(((ultimo - primeiro) / primeiro) * 100);
}
