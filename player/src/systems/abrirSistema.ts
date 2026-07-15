// Navegação para um sistema: recarrega o Player com ?sistema=<id>, reiniciando-o
// já com o sistema no config. Extraído do SeletorSistemas para ser reutilizado
// pelas telas do dashboard (RF04 — "Abrir projeto").

export function abrirSistema(id: string) {
  const params = new URLSearchParams(window.location.search);
  params.set('sistema', id);
  window.location.search = params.toString();
}
