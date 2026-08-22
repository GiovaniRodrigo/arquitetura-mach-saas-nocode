# Modelagem de entidades a partir da descrição do usuário

Quando o usuário descreve o "foco" do sistema em linguagem natural (ex.:
"quero agendar consultas médicas"), transforme a descrição em entidades e
relacionamentos seguindo estes passos:

1. **Extraia os substantivos de domínio** (Paciente, Profissional, Consulta,
   Agenda) — cada um vira candidato a entidade.
2. **Identifique o "evento central"** do sistema (Agendamento, Pedido,
   Matrícula) — normalmente é a entidade com mais relacionamentos e mais
   regras de negócio; ela costuma virar a tela principal do builder.
3. **Separe cadastro de transação**: entidades de cadastro (Paciente,
   Produto) mudam pouco e têm CRUD simples; entidades de transação
   (Consulta, Pedido) têm estados (rascunho, confirmado, cancelado) e
   precisam de regras de negócio explícitas por transição de estado.
4. **Modele relacionamentos pelo cardinalidade real do negócio**, não pela
   conveniência da tela — um Profissional atende várias Consultas, mas uma
   Consulta pertence a um único Profissional; force isso no modelo mesmo que
   a tela permita reatribuir depois.
5. **Nomeie campos como o usuário nomeia o domínio**, não com jargão técnico
   — isso reduz retrabalho quando o usuário revisar o formulário gerado.
