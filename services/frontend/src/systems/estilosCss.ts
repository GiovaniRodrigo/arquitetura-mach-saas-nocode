// Converte o objeto Estilos (guardado em propriedades.estilos) em CSS inline
// para o canvas. Puramente de apresentação — nenhuma lógica de domínio aqui.

import type { CSSProperties } from 'react';
import type { Estilos } from './componentRegistry';

export function estilosParaCss(estilos: Estilos | undefined): CSSProperties {
  if (!estilos) return {};
  const css: CSSProperties = {};

  if (estilos.largura) css.width = estilos.largura;
  if (estilos.altura) css.height = estilos.altura;
  if (estilos.padding) css.padding = estilos.padding;
  if (estilos.margem) css.margin = estilos.margem;
  if (estilos.display) css.display = estilos.display;
  if (estilos.direcao) css.flexDirection = estilos.direcao;
  if (estilos.justificar) css.justifyContent = estilos.justificar;
  if (estilos.alinhar) css.alignItems = estilos.alinhar;
  if (estilos.espacamento) css.gap = estilos.espacamento;
  if (estilos.fonteTamanho) css.fontSize = estilos.fonteTamanho;
  if (estilos.fontePeso) {
    css.fontWeight = estilos.fontePeso === 'bold' ? 700 : estilos.fontePeso === 'medium' ? 500 : 400;
  }
  if (estilos.cor) css.color = estilos.cor;
  if (estilos.textoAlinhar) css.textAlign = estilos.textoAlinhar;
  if (estilos.fundoCor) css.backgroundColor = estilos.fundoCor;
  if (estilos.bordaLargura) css.borderWidth = estilos.bordaLargura;
  if (estilos.bordaCor) css.borderColor = estilos.bordaCor;
  if (estilos.bordaLargura || estilos.bordaCor) css.borderStyle = 'solid';
  if (estilos.bordaRaio) css.borderRadius = estilos.bordaRaio;
  if (estilos.sombra) css.boxShadow = estilos.sombra;
  if (estilos.opacidade !== undefined) css.opacity = estilos.opacidade;
  if (estilos.posicao) css.position = estilos.posicao;
  if (estilos.posicao === 'absolute') {
    css.left = estilos.x ?? 0;
    css.top = estilos.y ?? 0;
  }
  return css;
}
