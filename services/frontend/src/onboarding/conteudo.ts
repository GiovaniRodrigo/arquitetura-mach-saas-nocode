// Conteúdo estático do onboarding guiado por tela (mesmo espírito do
// src/ajuda/artigos.ts): cada entrada descreve os passos mostrados na
// primeira visita da tela e reabertos manualmente pelo ícone de ajuda do
// cabeçalho (ver useOnboarding + OnboardingModal).

export interface PassoOnboarding {
  titulo: string;
  descricao: string;
}

export interface TelaOnboarding {
  /** Chave estável usada para lembrar quais telas já foram vistas. */
  chave: string;
  titulo: string;
  passos: PassoOnboarding[];
}

const telas: TelaOnboarding[] = [
  {
    chave: 'dashboard',
    titulo: 'Bem-vindo ao Dashboard',
    passos: [
      {
        titulo: 'Visão geral',
        descricao: 'Aqui você acompanha os indicadores principais da sua conta: clientes, sistemas ativos e resumo financeiro.',
      },
      {
        titulo: 'Navegação',
        descricao: 'Use o menu lateral para acessar Clientes, Configuração e o Monitor de Recursos da plataforma.',
      },
      {
        titulo: 'Busca rápida',
        descricao: 'Pressione Ctrl+K (ou clique em "Buscar…" no cabeçalho) a qualquer momento para navegar direto a um sistema ou tela.',
      },
    ],
  },
  {
    chave: 'clientes',
    titulo: 'Tela de Clientes',
    passos: [
      {
        titulo: 'Seus clientes',
        descricao: 'Esta lista reúne todos os clientes (tenants) cadastrados na sua conta.',
      },
      {
        titulo: 'Abrir um cliente',
        descricao: 'Clique em um cliente para ver seus sistemas e entrar no construtor de telas.',
      },
    ],
  },
  {
    chave: 'cliente-sistemas',
    titulo: 'Sistemas do Cliente',
    passos: [
      {
        titulo: 'Sistemas vinculados',
        descricao: 'Aqui ficam os sistemas criados para este cliente. Escolha um para editar telas, regras de negócio e versões.',
      },
      {
        titulo: 'Editar cliente',
        descricao: 'Você pode renomear ou excluir o cliente por aqui — a exclusão remove também seus sistemas e dados.',
      },
    ],
  },
  {
    chave: 'sistema-telas',
    titulo: 'Construtor de Telas',
    passos: [
      {
        titulo: 'Canvas',
        descricao: 'Arraste componentes do painel esquerdo para o canvas central para montar a tela.',
      },
      {
        titulo: 'Camadas e inspetor',
        descricao: 'Use o painel de camadas para organizar a hierarquia e o inspetor à direita para ajustar propriedades do componente selecionado.',
      },
      {
        titulo: 'Posição livre',
        descricao: 'Componentes com posição livre podem ser trazidos para frente ou enviados para trás pela barra flutuante.',
      },
    ],
  },
  {
    chave: 'sistema-regras',
    titulo: 'Regras de Negócio',
    passos: [
      {
        titulo: 'Regras do sistema',
        descricao: 'Defina aqui as validações e comportamentos que se aplicam aos formulários e telas deste sistema.',
      },
    ],
  },
  {
    chave: 'sistema-versao',
    titulo: 'Versões do Sistema',
    passos: [
      {
        titulo: 'Publicar',
        descricao: 'Publique a versão atual para que ela passe a ser exibida aos usuários finais do sistema.',
      },
      {
        titulo: 'Reverter',
        descricao: 'Se precisar, volte para qualquer versão publicada anteriormente a qualquer momento.',
      },
    ],
  },
  {
    chave: 'configuracao',
    titulo: 'Configuração',
    passos: [
      {
        titulo: 'Segurança',
        descricao: 'Atualize sua senha e ative a autenticação em duas etapas (MFA) para proteger sua conta.',
      },
      {
        titulo: 'White label',
        descricao: 'Personalize marca, cores e domínio da plataforma para seus próprios clientes.',
      },
    ],
  },
  {
    chave: 'monitor',
    titulo: 'Monitor de Recursos',
    passos: [
      {
        titulo: 'Status dos serviços',
        descricao: 'Acompanhe uptime e memória dos serviços da plataforma, atualizados automaticamente a cada poucos segundos.',
      },
      {
        titulo: 'Indisponibilidade',
        descricao: 'Um serviço individual fora do ar aparece destacado sem interromper a visão dos demais.',
      },
    ],
  },
  {
    chave: 'perfil',
    titulo: 'Perfil e Cadastro',
    passos: [
      {
        titulo: 'Seus dados',
        descricao: 'Revise e atualize seu nome, e-mail e demais dados de cadastro nesta tela.',
      },
    ],
  },
];

const onboardingPorChave = new Map(telas.map((t) => [t.chave, t]));

/** Chaves de todas as telas com onboarding — usado por testes para simular um usuário que já viu tudo. */
export const chavesOnboarding: string[] = telas.map((t) => t.chave);

/** Resolve a tela de onboarding a partir do pathname atual, ou null se a rota não tem tour. */
export function telaOnboardingDe(pathname: string): TelaOnboarding | null {
  if (/\/dashboard\/clientes\/[^/]+\/sistemas\/[^/]+\/telas/.test(pathname)) {
    return onboardingPorChave.get('sistema-telas') ?? null;
  }
  if (/\/dashboard\/clientes\/[^/]+\/sistemas\/[^/]+\/regras/.test(pathname)) {
    return onboardingPorChave.get('sistema-regras') ?? null;
  }
  if (/\/dashboard\/clientes\/[^/]+\/sistemas\/[^/]+\/versao/.test(pathname)) {
    return onboardingPorChave.get('sistema-versao') ?? null;
  }
  if (/\/dashboard\/clientes\/[^/]+$/.test(pathname)) {
    return onboardingPorChave.get('cliente-sistemas') ?? null;
  }
  if (pathname.startsWith('/dashboard/clientes')) {
    return onboardingPorChave.get('clientes') ?? null;
  }
  if (pathname.startsWith('/dashboard/configuracao')) {
    return onboardingPorChave.get('configuracao') ?? null;
  }
  if (pathname.startsWith('/dashboard/monitor')) {
    return onboardingPorChave.get('monitor') ?? null;
  }
  if (pathname.startsWith('/dashboard/perfil')) {
    return onboardingPorChave.get('perfil') ?? null;
  }
  if (pathname === '/dashboard' || pathname === '/dashboard/') {
    return onboardingPorChave.get('dashboard') ?? null;
  }
  return null;
}
