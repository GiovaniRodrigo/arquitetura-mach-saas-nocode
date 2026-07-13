# Fluxos Críticos, Performance, Contratos gRPC e Evolução Técnica

## 1. Validação de Dados Distribuída
A validação ocorre de forma paralela em duas camadas utilizando *Blind Indexes*:
* **No Front-end (Headless Player):** O mapa de campos é verificado antes do envio. Se houver erros, a interface bloqueia o envio imediatamente, poupando largura de banda.
* **No Back-end (Logic Engine):** O servidor gRPC revalida sempre o payload contra o esquema salvo na base de dados para evitar submissões maliciosas via API. Caso detecte dados inválidos, devolve um mapa de erros estruturado pelo `blind_index` do componente falhado, permitindo que o front-end sinalize o input exato sem expor estruturas reais de tabelas.

## 2. Estratégia de Renderização em Lote (Batching)
Devido ao fluxo contínuo de dados gerado por WebSockets e ações do utilizador, o *Headless Player* adota uma estratégia de *batching* no DOM Virtual:
* O front-end acumula todas as alterações ocorridas num intervalo de **16 milissegundos** (alinhado à taxa de atualização de 60Hz dos monitores).
* O algoritmo de comparação (*Diffing*) do DOM Virtual executa **apenas uma vez por lote**, aplicando um único impacto otimizado no DOM real, eliminando quebras de fluidez (*jank*).

## 3. Mecanismo de Deploy e Rollback Instantâneo
* **Abordagem Interpretada:** O ato de publicar consiste em criar uma nova linha numa tabela de histórico de versões (ex: `versoes_sistema`) e alternar o estado de uma flag para *Ativa*. O *Headless Player* consome sempre a versão que possui a flag ativa.
* **Rollback em Milissegundos:** Se uma nova versão apresentar falhas, a reversão é executada alterando novamente a flag de status para a versão estável anterior, sem necessidade de recompile ou indisponibilidade.

## 4. Fluxo de Exportação de Dados em Segundo Plano
| Etapa do Fluxo | Tecnologia / Padrão Aplicado | Descrição Técnica |
| :--- | :--- | :--- |
| **1. Solicitação** | Gateway (Go) -> Export Engine | O utilizador clica em "Exportar". O sistema cria um Job e liberta o front-end imediatamente com uma notificação, evitando timeouts. |
| **2. Recolha** | gRPC Server Streaming | A *Export Engine* consome dados de UI, Regras e Registos Operacionais em pedaços (*chunks*) contínuos para não sobrecarregar a memória RAM. |
| **3. Armazenamento** | Cloud Storage (S3 / GCS) | Os dados consolidados são compactados num ficheiro e armazenados num bucket seguro e isolado na nuvem. |
| **4. Entrega Segura**| Presigned URL (Link Temporário) | É gerado um link encriptado de expiração rápida. O utilizador descarrega o ficheiro diretamente da nuvem, poupando os recursos e largura de banda do API Gateway. |

## 5. Contrato Protocol Buffers Oficial (`.proto`)
A comunicação padrão entre a Camada de Gateway e o *Logic Engine* para submissão de formulários utiliza o mecanismo **Unary RPC**. Coleções de campos dinâmicos trafegam via estruturas do tipo `map<string, string>` vinculando o *Blind Index* ao valor textual preenchido.

```protobuf
syntax = "proto3";

package construtor.logic.v1;

// Representa a submissão de dados operacionais dinâmicos
message SalvarFormularioRequest {
  // Mapa dinâmico de chave-valor ligando o Blind Index do input ao valor preenchido
  map<string, string> dados_formulario = 1;
}

// Resposta com o estado da validação e persistência
message SalvarFormularioResponse {
  bool sucesso = 1;
  // Mapa de erros indexado pelo Blind Index do componente falhado
  map<string, string> erros_validacao = 2;
  string mensagem_status = 3;
}

service LogicEngineService {
  // Ação Unary para validar e gravar dados na base de dados partilhada
  rpc SalvarFormulario (SalvarFormularioRequest) returns (SalvarFormularioResponse);
}
```

## 6. Roteiro de Evolução Futura: Abordagem Compilada
A arquitetura atual está preparada para evoluir para um modelo híbrido ou puramente compilado sem necessidade de refatorizar o núcleo da plataforma:
* Os microsserviços de *Design Engine* e *Logic Engine* permanecerão inalterados, pois continuam a ser os repositórios oficiais das regras de negócio e ecrãs.
* O *Deploy Engine* será expandido para incluir geradores de código automáticos.
* Ao clicar em "Publicar", o serviço lerá as definições via gRPC e disparará pipelines automatizados de CI/CD para compilar imagens Docker independentes ou funções Serverless na nuvem para clientes específicos (como instâncias *Single-Tenant* ou contas empresariais) que exijam isolamento total de performance, infraestrutura própria e aplicação de limites estritos de CPU/Memória nativos no Kubernetes.
