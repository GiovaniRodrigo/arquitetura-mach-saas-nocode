// Sanitização do HTML inline produzido pela edição rica de texto (RF09,
// barra flutuante estilo Notion): permite só as tags/atributos que o
// document.execCommand('bold'|'italic'|'underline') realmente gera — tudo o
// mais (script, on*, iframe, style arbitrário) é descartado. Usa o parser
// HTML do próprio navegador (via <template>) em vez de regex — muito mais
// difícil de burlar do que strip por regex.

const TAGS_PERMITIDAS = new Set(['B', 'STRONG', 'I', 'EM', 'U', 'BR', 'SPAN']);
const PROPRIEDADES_STYLE_PERMITIDAS = new Set(['font-weight', 'font-style', 'text-decoration']);
// Tags cujo conteúdo é descartado por inteiro (não só desembrulhado como
// texto) — script/style não têm conteúdo textual seguro para preservar.
const TAGS_CONTEUDO_DESCARTADO = new Set(['SCRIPT', 'STYLE', 'NOSCRIPT']);

function sanitizarNo(no: Node): Node | null {
  if (no.nodeType === Node.TEXT_NODE) return no.cloneNode();

  if (no.nodeType !== Node.ELEMENT_NODE) return null;
  const el = no as Element;

  if (TAGS_CONTEUDO_DESCARTADO.has(el.tagName)) return null;

  if (!TAGS_PERMITIDAS.has(el.tagName)) {
    // Elemento não permitido: preserva só o conteúdo textual/filhos permitidos.
    const fragmento = document.createDocumentFragment();
    el.childNodes.forEach((filho) => {
      const sanitizado = sanitizarNo(filho);
      if (sanitizado) fragmento.appendChild(sanitizado);
    });
    return fragmento;
  }

  const limpo = document.createElement(el.tagName);
  if (el.tagName === 'SPAN') {
    const style = (el as HTMLElement).style;
    for (const prop of PROPRIEDADES_STYLE_PERMITIDAS) {
      const valor = style.getPropertyValue(prop);
      if (valor) limpo.style.setProperty(prop, valor);
    }
  }
  el.childNodes.forEach((filho) => {
    const sanitizado = sanitizarNo(filho);
    if (sanitizado) limpo.appendChild(sanitizado);
  });
  return limpo;
}

/** Sanitiza um HTML inline simples (negrito/itálico/sublinhado) removendo
 * qualquer tag/atributo fora do allowlist. Seguro contra XSS: usa o parser
 * de HTML do navegador para construir uma árvore de nós real (não string
 * concatenation), então reconstrói só com os nós permitidos. */
export function sanitizarHtml(html: string): string {
  const template = document.createElement('template');
  template.innerHTML = html;
  const fragmento = document.createDocumentFragment();
  template.content.childNodes.forEach((no) => {
    const sanitizado = sanitizarNo(no);
    if (sanitizado) fragmento.appendChild(sanitizado);
  });
  const saida = document.createElement('div');
  saida.appendChild(fragmento);
  return saida.innerHTML;
}

/** Extrai o texto puro (sem tags) de um HTML inline — usado para exibir o
 * valor no campo "Texto" simples do Inspector, que não edita rich text. */
export function htmlParaTextoPlano(html: string): string {
  const div = document.createElement('div');
  div.innerHTML = html;
  return div.textContent ?? '';
}
