# Estrutura de telas e fluxos no builder no-code

Boas práticas para estruturar as "Telas" de um sistema montado visualmente:

- **Uma tela por objetivo do usuário final**, não por entidade. Uma tela de
  "Agendar Consulta" pode envolver Paciente, Profissional e Agenda ao mesmo
  tempo — não force uma tela por tabela do banco.
- **Liste antes de detalhar**: comece com telas de listagem (tabela com
  busca/filtro) das entidades principais; só depois desenhe telas de
  criação/edição — isso valida o modelo de dados antes de investir em
  formulário complexo.
- **Componentes reutilizáveis primeiro**: identifique padrões repetidos
  (cartão de status, seletor de data, tabela paginada) e trate-os como
  componentes reaproveitáveis no Canvas, não como telas coladas manualmente
  toda vez.
- **Estado vazio, erro e carregamento fazem parte do design da tela**, não
  são detalhe de implementação — planeje o que aparece quando não há dados
  ainda, quando a ação falha, e enquanto uma chamada está em andamento.
- **Navegação reflete a hierarquia do negócio**: se "Consulta" pertence a
  "Paciente", a navegação natural é entrar no paciente e ver suas consultas,
  não uma lista global de consultas sem contexto (a menos que exista um caso
  de uso real para isso, como a agenda do dia de um profissional).
