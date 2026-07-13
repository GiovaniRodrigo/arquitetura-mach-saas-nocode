# Mensageria Assíncrona (KEDA) e Observabilidade (OpenTelemetry)

## 1. Camada de Eventos Assíncronos e Autoscaling Dinâmico (KEDA)
Para garantir que tarefas pesadas em segundo plano (notificações, webhooks e integrações de terceiros configuradas pelo utilizador) não bloqueiem as threads síncronas do sistema, adota-se uma arquitetura orientada a eventos assíncronos via **RabbitMQ** orquestrada por **KEDA (Kubernetes Event-driven Autoscaling)**.

### Topologia de Mensageria e Fluxo
1. Quando uma regra de negócio processada pelo Logic Engine determina o disparo de uma tarefa em segundo plano, o serviço publica um evento estruturado em JSON contendo o `tenant_id` e o `component_blind_index` correspondente numa *Exchange* do RabbitMQ.
2. O RabbitMQ faz o roteamento dinâmico para as filas dedicadas (ex: `webhooks.disparo`, `notificacoes.envio`), liberando o fluxo gRPC imediatamente.

### Escalabilidade Elástica até Zero (Scale-to-Zero)
Os *workers* responsáveis pelo consumo das mensagens são controlados pelo KEDA no cluster Kubernetes utilizando um recurso customizado `ScaledObject`:
* **Métrica de Escalonamento:** O KEDA monitoriza o tamanho da fila (`QueueLength`) diretamente na API do RabbitMQ, ignorando a reatividade ineficiente de métricas de CPU/Memória.
* **Scale-to-Zero:** Se a fila estiver totalmente vazia (ex: períodos de baixa atividade noturna), o KEDA reduz o número de Pods ativos para **0**, eliminando custos de computação ociosos.
* **Comportamento sob Carga:** Ao detetar um acúmulo de mensagens, o KEDA escala horizontalmente e de forma agressiva até ao limite estipulado (ex: `maxReplicaCount: 50`), distribuindo a carga de forma justa.

### Isolamento contra o "Vizinho Barulhento" (Noisy Neighbor)
Para mitigar o risco de um único cliente inundar a plataforma com milhões de webhooks e monopolizar os *workers*, o Deploy Engine provisiona chaves de roteamento dinâmicas e o RabbitMQ implementa políticas de *Fair Queuing*. Falhas contínuas de integração de terceiros são desviadas de imediato para filas de erros (*Dead Letter Queues* - DLQ), acionando alertas no painel do respetivo *tenant* sem impactar os demais utilizadores.

## 2. Observabilidade e Rastreamento Distribuído (OpenTelemetry)
A rastreabilidade de ponta a ponta em fluxos de rede híbridos (HTTP, gRPC e AMQP/RabbitMQ) é garantida utilizando o padrão **OpenTelemetry** com armazenamento e visualização via **Jaeger**.

### Propagação de Contexto Transmídia
Para unificar a jornada de uma requisição distribuída, o sistema injeta e extrai o identificador global de rastreio (**Trace ID**) seguindo a especificação *W3C Trace Context* através do cabeçalho `traceparent`:
1. **Origem (API Gateway):** O middleware em Go gera o Trace ID raiz ao receber a requisição HTTP.
2. **Trânsito Síncrono (gRPC):** O interceptor gRPC do OpenTelemetry injeta o Trace ID nativamente como Metadados binários na chamada para o Logic Engine.
3. **Trânsito Assíncrono (AMQP):** Antes de publicar a mensagem de segundo plano no RabbitMQ, o Logic Engine grava o contexto do trace corrente dentro do mapa de *Headers* da mensagem AMQP.
4. **Destino (Workers):** O trabalhador de execução extrai o `traceparent` dos headers da mensagem, abre um sub-bloco de tempo (*Span*) e associa os logs ou erros de execução diretamente ao histórico unificado da transação.

### Tags de Contexto Multi-Tenant Seguras
Cada *Span* de monitorização é enriquecido com atributos estruturais anonimizados para permitir auditoria interna e depuração ágil por parte da equipa de engenharia:
* `platform.tenant_id`: Permite filtrar instantaneamente todos os *traces* de infraestrutura associados a um cliente específico.
* `platform.component.blind_index`: Identifica o componente exato gerador do gargalo sem nunca expor metadados ou dados sensíveis do negócio no Jaeger.
