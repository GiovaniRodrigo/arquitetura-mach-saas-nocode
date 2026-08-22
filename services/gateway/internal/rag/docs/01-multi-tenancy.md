# Multi-tenancy e isolamento de dados

Ao modelar um sistema que atende vários clientes (tenants) — clínicas, lojas,
escolas, franquias — a primeira decisão de arquitetura é qual entidade
representa o "dono do dado" isolado. Recomendações:

- Escolha uma entidade raiz de tenant (ex.: "Clínica", "Loja", "Unidade") e
  faça toda entidade sensível referenciá-la direta ou indiretamente. Nunca
  compartilhe uma tabela entre tenants sem uma coluna de tenant_id
  obrigatória e indexada.
- Prefira isolamento lógico (linha por tenant + Row-Level Security) a bancos
  físicos separados por tenant, a menos que haja exigência regulatória de
  isolamento físico — RLS escala melhor operacionalmente para dezenas/centenas
  de tenants.
- Nunca confie em filtro aplicado só na aplicação (WHERE tenant_id = ?)
  como única defesa: um endpoint esquecido do filtro vaza dados entre
  tenants. RLS no banco é a rede de segurança.
- Hierarquias de tenant (matriz/filial, dono/parceiro) precisam de uma regra
  explícita de "quem pode ver o quê" — normalmente resolvida checando se o
  tenant alvo é filho direto do tenant do usuário autenticado, nunca
  recursivamente sem necessidade.
