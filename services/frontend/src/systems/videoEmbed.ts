// Converte URLs de compartilhamento do YouTube/Vimeo para URL de embed —
// usado tanto no preview do canvas quanto na visualização do site. URLs que
// não batem com nenhum desses padrões são tratadas como arquivo de vídeo
// direto (renderizado via <video>, não <iframe>).

export function paraEmbedUrl(url: string): string | null {
  const youtube = url.match(/(?:youtu\.be\/|youtube\.com\/(?:watch\?v=|embed\/))([\w-]+)/);
  if (youtube) return `https://www.youtube.com/embed/${youtube[1]}`;

  const vimeo = url.match(/vimeo\.com\/(\d+)/);
  if (vimeo) return `https://player.vimeo.com/video/${vimeo[1]}`;

  return null;
}
