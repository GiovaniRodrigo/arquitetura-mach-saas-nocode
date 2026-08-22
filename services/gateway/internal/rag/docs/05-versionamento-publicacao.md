# Versionamento e publicação de sistemas no-code

Ao construir um sistema que será publicado para o cliente final, trate cada
alteração relevante (novo campo, nova regra, novo layout) como parte de uma
versão explícita, não como edição direta do ambiente em produção:

- **Rascunho vs. versão publicada**: edite livremente em rascunho; só publique
  quando o fluxo tiver sido testado de ponta a ponta (criar, listar, editar,
  excluir o registro principal).
- **Rollback precisa ser trivial**: se uma versão publicada quebra um fluxo
  do cliente, o caminho de volta para a versão anterior deve ser uma ação de
  um clique, não uma reconstrução manual.
- **Mudanças estruturais (remover campo, renomear entidade) são arriscadas**
  em sistemas já publicados com dados reais — prefira depreciar um campo
  (esconder da tela, manter no banco) a excluir, até confirmar que nenhum
  dado histórico depende dele.
- **Comunique o que mudou** em termos que o dono do sistema entende
  ("agora dá para agendar retorno automático") e não em termos técnicos
  ("adicionado campo tipo_consulta").
