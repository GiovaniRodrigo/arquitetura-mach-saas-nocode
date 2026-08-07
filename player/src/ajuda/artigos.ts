// Conteúdo estático inicial da documentação (RF20). A assinatura de busca já é
// preparada para trocar por `client.buscarArtigos(termo)` quando o CMS existir
// (ver Fora de Escopo de specs/004-reestruturacao-ia-navegacao/spec.md).

export interface ArtigoDocumentacao {
  id: string;
  categoria: string;
  titulo: string;
  conteudo: string;
}

export const artigos: ArtigoDocumentacao[] = [
  {
    id: 'conta-criar',
    categoria: 'Conta',
    titulo: 'Como criar minha conta',
    conteudo: 'Cadastre-se com Google ou GitHub na tela inicial para começar a usar a plataforma.',
  },
  {
    id: 'conta-senha',
    categoria: 'Conta',
    titulo: 'Como alterar minha senha',
    conteudo: 'Acesse Configuração > Segurança e use a opção de atualizar senha para trocar sua senha atual.',
  },
  {
    id: 'seguranca-mfa',
    categoria: 'Segurança',
    titulo: 'Como ativar a autenticação em duas etapas (MFA)',
    conteudo: 'Em Configuração > Segurança, ative o MFA e escaneie o QR code com um aplicativo autenticador (TOTP).',
  },
  {
    id: 'construtor-telas',
    categoria: 'Construtor',
    titulo: 'Como criar telas e componentes',
    conteudo: 'Em Clientes, selecione um sistema e abra a aba Telas para criar e editar telas e componentes no canvas.',
  },
  {
    id: 'construtor-versao',
    categoria: 'Construtor',
    titulo: 'Como publicar uma nova versão',
    conteudo: 'Na aba Versão do sistema, publique a versão atual ou reverta para uma versão anterior a qualquer momento.',
  },
  {
    id: 'faturamento-resumo',
    categoria: 'Faturamento',
    titulo: 'Como consultar minha receita de assinatura',
    conteudo: 'O card Resumo Financeiro do Dashboard mostra a receita de assinatura dos seus clientes vinculados.',
  },
];

/** Filtra por termo no título ou conteúdo (RF21). Comparação sem acento/caixa. */
export function buscarArtigos(termo: string): ArtigoDocumentacao[] {
  const alvo = termo.trim().toLowerCase();
  if (!alvo) return artigos;
  return artigos.filter(
    (a) => a.titulo.toLowerCase().includes(alvo) || a.conteudo.toLowerCase().includes(alvo),
  );
}
