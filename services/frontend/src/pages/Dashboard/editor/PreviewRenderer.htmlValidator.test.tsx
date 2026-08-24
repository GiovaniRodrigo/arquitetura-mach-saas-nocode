import { describe, it } from 'vitest';
import { render } from '@testing-library/react';
import type { Componente } from '../../../api/types';
import { validarFragmentoHtml } from '../../../test/htmlValidator';
import { PreviewRenderer } from './PreviewRenderer';

// Validação de HTML — segmento 2: telas construídas pelos CLIENTES no
// builder nocode (RF09). O PreviewRenderer é o renderizador usado em
// "Visualização do site" (PreviewOverlay) — o mais perto do resultado
// publicado que existe hoje sem um pipeline de publish separado (ver
// comentário no topo de PreviewRenderer.tsx), e é o mesmo catálogo de
// componentes que build/seed-demo-site.sh usa para popular o site de
// demonstração. Testado aqui como componente (RTL + jsdom), fora do e2e:
// abrir uma tela real no editor depende do canal de colaboração
// (services/collab/WebSocket Phoenix) — fora do escopo hermético dos e2e
// (ver nota em e2e/telas.spec.ts e e2e/htmlSistema.spec.ts).

function no(parcial: Partial<Componente> & { blind_index: string; tipo: string }): Componente {
  return { componente_filhos: [], propriedades: {}, ...parcial };
}

// Árvore representativa cobrindo cada `case` do switch de PreviewRenderer —
// os mesmos tipos de componente usados pelo site de demonstração seedado
// (header/footer/menu/carrossel/accordion/tabs, formulário completo, etc.).
const ARVORE: Componente = no({
  blind_index: 'root',
  tipo: 'main',
  componente_filhos: [
    no({
      blind_index: 'header',
      tipo: 'header',
      componente_filhos: [
        no({ blind_index: 'header-heading', tipo: 'heading', propriedades: { texto: '<b>Acme</b>' } }),
        no({
          blind_index: 'header-menu',
          tipo: 'menu',
          propriedades: { itens: [{ label: 'Início', url: '/' }, { label: 'Contato', url: '/contato' }] },
        }),
        no({ blind_index: 'header-cta', tipo: 'botao', propriedades: { texto: 'Entrar' } }),
      ],
    }),
    no({
      blind_index: 'breadcrumb',
      tipo: 'breadcrumb',
      propriedades: { itens: [{ label: 'Início', url: '/' }, { label: 'Produtos' }] },
    }),
    no({
      blind_index: 'hero',
      tipo: 'section',
      componente_filhos: [
        no({ blind_index: 'hero-titulo', tipo: 'heading', propriedades: { texto: 'Bem-vindo' } }),
        no({ blind_index: 'hero-texto', tipo: 'paragrafo', propriedades: { texto: 'Um parágrafo qualquer.' } }),
        no({ blind_index: 'hero-link', tipo: 'link', propriedades: { texto: 'Saiba mais' } }),
        no({ blind_index: 'hero-badge', tipo: 'badge', propriedades: { texto: 'Novo' } }),
        no({ blind_index: 'hero-icone', tipo: 'icone', propriedades: { icone: 'Star' } }),
        no({ blind_index: 'hero-avaliacao', tipo: 'avaliacao', propriedades: { valor: 4 } }),
        no({ blind_index: 'hero-progresso', tipo: 'progresso', propriedades: { valor: 60 } }),
        no({ blind_index: 'hero-avatar', tipo: 'avatar', propriedades: { src: 'https://exemplo.com/a.png', alt: 'Perfil' } }),
        no({ blind_index: 'hero-spinner', tipo: 'spinner' }),
        no({ blind_index: 'hero-skeleton', tipo: 'skeleton' }),
        no({ blind_index: 'hero-toggle', tipo: 'toggle', propriedades: { ativo: true } }),
        no({ blind_index: 'hero-alerta', tipo: 'alerta', propriedades: { variante: 'sucesso', texto: 'Tudo certo' } }),
      ],
    }),
    no({
      blind_index: 'midia',
      tipo: 'container',
      componente_filhos: [
        no({ blind_index: 'midia-imagem', tipo: 'imagem', propriedades: { src: 'https://exemplo.com/img.png', alt: 'Produto' } }),
        no({
          blind_index: 'midia-carrossel',
          tipo: 'carrossel',
          propriedades: { imagens: [{ src: 'https://exemplo.com/1.png', alt: 'Slide 1' }] },
        }),
        no({ blind_index: 'midia-video', tipo: 'video', propriedades: { src: 'https://www.youtube.com/watch?v=dQw4w9WgXcQ' } }),
      ],
    }),
    no({
      blind_index: 'interativos',
      tipo: 'container',
      componente_filhos: [
        no({
          blind_index: 'interativos-accordion',
          tipo: 'accordion',
          propriedades: { itens: [{ titulo: 'Pergunta', conteudo: 'Resposta' }] },
        }),
        no({
          blind_index: 'interativos-tabs',
          tipo: 'tabs',
          propriedades: { itens: [{ titulo: 'Aba 1', conteudo: 'Conteúdo 1' }] },
        }),
      ],
    }),
    no({
      blind_index: 'formulario',
      tipo: 'section',
      componente_filhos: [
        no({ blind_index: 'form-input', tipo: 'input', propriedades: { texto: 'Nome' } }),
        no({ blind_index: 'form-select', tipo: 'select', propriedades: { texto: 'Opção' } }),
        no({ blind_index: 'form-checkbox', tipo: 'checkbox', propriedades: { texto: 'Aceito os termos' } }),
        no({ blind_index: 'form-radio', tipo: 'radio', propriedades: { texto: 'Opção A' } }),
        no({ blind_index: 'form-textarea', tipo: 'textarea', propriedades: { texto: 'Mensagem' } }),
        no({ blind_index: 'form-submit', tipo: 'botao', propriedades: { texto: 'Enviar' } }),
      ],
    }),
    no({
      blind_index: 'sidebar',
      tipo: 'sidebar',
      componente_filhos: [no({ blind_index: 'sidebar-texto', tipo: 'paragrafo', propriedades: { texto: 'Filtros' } })],
    }),
    no({
      blind_index: 'rightbar',
      tipo: 'rightbar',
      componente_filhos: [no({ blind_index: 'rightbar-texto', tipo: 'paragrafo', propriedades: { texto: 'Ajuda' } })],
    }),
    no({
      blind_index: 'footer',
      tipo: 'footer',
      componente_filhos: [
        no({ blind_index: 'footer-copy', tipo: 'paragrafo', propriedades: { texto: '© 2026 Acme' } }),
      ],
    }),
  ],
});

describe('PreviewRenderer — HTML das telas dos clientes (html-validator/WHATWG)', () => {
  it('produz HTML válido para o catálogo completo de componentes do builder', async () => {
    const { container } = render(<PreviewRenderer no={ARVORE} />);
    // heading-level ignorado: ver justificativa em src/test/htmlValidator.ts
    // (o componente "heading" não carrega nível semântico no modelo de dados).
    await validarFragmentoHtml(container.innerHTML, 'Preview — catálogo de componentes do builder', [
      'heading-level',
    ]);
  });
});
