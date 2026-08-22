# Regras de negócio: quando extrair da tela

Regras de negócio devem ser modeladas como entidade própria (fora do
desenho visual da tela) quando pelo menos uma destas condições é verdadeira:

- A regra depende de **mais de um dado** que não está todo visível na
  mesma tela (ex.: "não permitir agendar consulta se o paciente tem
  cobrança em aberto").
- A regra muda **com frequência maior** que o layout da tela — regras de
  desconto, de aprovação, de prazo, tendem a mudar por decisão de negócio,
  não de design.
- A regra precisa ser **auditável** — quem mudou o limite de crédito e
  quando, por exemplo.

Regras simples de validação de campo (obrigatório, formato, faixa de valor)
podem ficar direto na propriedade do componente na tela — não vale a pena
extrair regra de negócio formal para "campo obrigatório".

Ao descrever uma regra de negócio para o usuário, use o formato:
**gatilho → condição → ação** ("Quando o agendamento é confirmado, se o
paciente tem 3+ faltas nos últimos 30 dias, exigir pré-pagamento"). Isso
mapeia diretamente para como a aba "Regras de Negócio" do builder trata cada
regra.
