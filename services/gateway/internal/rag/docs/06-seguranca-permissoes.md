# Segurança e permissões em sistemas de cliente final

Todo sistema montado no builder é operado por perfis diferentes (dono,
parceiro, cliente final, atendente) — modele permissão desde o início:

- **Defina papéis pelo que a pessoa faz**, não por cargo formal — um
  "atendente" e uma "recepcionista" podem ser o mesmo papel se fazem as
  mesmas ações no sistema.
- **Toda ação destrutiva (excluir, cancelar, estornar) precisa de
  confirmação explícita e, idealmente, de um papel com permissão elevada** —
  não deixe o papel padrão excluir dados de outros usuários sem barreira.
- **Dados sensíveis (documentos pessoais, dados de saúde, dados
  financeiros) precisam de campo marcado como sensível** — isso deve
  refletir em mascaramento na listagem e em log de acesso, não só em
  "não mostrar por padrão".
- **Nunca exponha o identificador interno (ID sequencial) do banco em URLs
  públicas** sem necessidade — prefira identificadores opacos para recursos
  acessados por clientes finais fora do sistema autenticado.
