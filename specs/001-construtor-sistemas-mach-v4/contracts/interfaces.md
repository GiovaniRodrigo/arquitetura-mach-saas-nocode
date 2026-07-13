# Interfaces: Construtor de Sistemas MACH V4

Contratos gRPC (`.proto`) — a fonte única de verdade da comunicação interna (RNF01). Todos os serviços recebem `tenant_id` exclusivamente via gRPC Metadata (chave `x-tenant-context-bin`), nunca no corpo das mensagens (RNF02, RN01).

---

## `construtor.logic.v1.LogicEngineService`

Contrato oficial transcrito de `doc/CONTRACTS_PERFORMANCE.md §5`, ampliado com o CRUD de regras (RF02, RF07).

```protobuf
syntax = "proto3";

package construtor.logic.v1;

// Representa a submissão de dados operacionais dinâmicos
message SalvarFormularioRequest {
  string sistema_id = 1;
  // Mapa dinâmico de chave-valor ligando o Blind Index do input ao valor preenchido
  map<string, string> dados_formulario = 2;
}

// Resposta com o estado da validação e persistência
message SalvarFormularioResponse {
  bool sucesso = 1;
  // Mapa de erros indexado pelo Blind Index do componente falhado (RNF08)
  map<string, string> erros_validacao = 2;
  string mensagem_status = 3;
}

message Regra {
  string id = 1;
  string sistema_id = 2;
  // Árvore de decisão serializada (nós lógicos)
  bytes arvore_decisao = 3;
}

service LogicEngineService {
  // Ação Unary para validar e gravar dados na base de dados partilhada (RF07, RN08)
  rpc SalvarFormulario (SalvarFormularioRequest) returns (SalvarFormularioResponse);

  rpc CriarRegra (Regra) returns (Regra);
  rpc ObterRegra (ObterRegraRequest) returns (Regra);
  rpc AtualizarRegra (Regra) returns (Regra);
  rpc RemoverRegra (ObterRegraRequest) returns (RemoverResponse);
}

message ObterRegraRequest { string id = 1; }
message RemoverResponse { bool sucesso = 1; }
```

**Implementações esperadas**: servidor em `services/logic/internal/server/grpc.go`; clientes no Gateway e no Export Engine.

---

## `construtor.design.v1.DesignEngineService`

CRUD da árvore recursiva e persistência em lote da colaboração (RF01, RN06).

```protobuf
syntax = "proto3";

package construtor.design.v1;

message Componente {
  string blind_index = 1;
  string tipo = 2;
  bytes propriedades = 3;              // JSON serializado
  repeated Componente componente_filhos = 4;  // padrão Composite
}

message Design {
  string id = 1;
  string sistema_id = 2;
  string nome = 3;
  Componente arvore = 4;
}

message SalvarDesignRequest {
  // Payload consolidado enviado pelo motor Elixir após o debounce de 5s (RN06)
  Design design = 1;
}

service DesignEngineService {
  rpc CriarDesign (Design) returns (Design);
  rpc ObterDesign (ObterDesignRequest) returns (Design);
  rpc AtualizarDesign (Design) returns (Design);
  rpc RemoverDesign (ObterDesignRequest) returns (RemoverResponse);
  // Chamada única em lote da colaboração (write-behind)
  rpc SalvarDesign (SalvarDesignRequest) returns (SalvarDesignResponse);
}

message ObterDesignRequest { string id = 1; }
message SalvarDesignResponse { bool sucesso = 1; }
message RemoverResponse { bool sucesso = 1; }
```

**Implementações esperadas**: servidor em `services/design/internal/server/grpc.go`; clientes no Gateway, no motor Elixir (`collab/lib/collab/grpc/design_client.ex`) e no Export Engine.

---

## `construtor.iam.v1.IAMService`

Autenticação e permissões por componente (RF03, RN03).

```protobuf
syntax = "proto3";

package construtor.iam.v1;

message ValidarTokenRequest { string jwt = 1; }
message ValidarTokenResponse {
  bool valido = 1;
  string tenant_id = 2;
  string user_id = 3;
  string tipo = 4; // dono | parceiro | cliente
}

message AvaliarPermissoesRequest {
  string sistema_id = 1;
  // Blind indexes dos componentes do ecrã em renderização
  repeated string blind_indexes = 2;
}

message PermissaoComponente {
  bool view = 1;
  bool click = 2;
}

message AvaliarPermissoesResponse {
  // Mapa booleano final — a lógica das regras nunca sai do servidor (RN03)
  map<string, PermissaoComponente> permissions = 1;
}

service IAMService {
  rpc ValidarToken (ValidarTokenRequest) returns (ValidarTokenResponse);
  rpc AvaliarPermissoes (AvaliarPermissoesRequest) returns (AvaliarPermissoesResponse);
}
```

**Implementações esperadas**: servidor em `services/iam/internal/server/grpc.go`; cliente no Gateway (middleware de auth e rota de permissões).

---

## `construtor.deploy.v1.DeployEngineService`

Publicação por flag ativa e rollback instantâneo (RF04, RN04, RN05).

```protobuf
syntax = "proto3";

package construtor.deploy.v1;

message PublicarRequest { string sistema_id = 1; }
message PublicarResponse {
  string versao_id = 1;
  int32 numero = 2;
}

message RollbackRequest {
  string sistema_id = 1;
  // 0 = versão imediatamente anterior
  int32 versao_numero = 2;
}
message RollbackResponse { int32 versao_ativa = 1; }

message ObterVersaoAtivaRequest { string sistema_id = 1; }
message VersaoAtiva {
  string versao_id = 1;
  int32 numero = 2;
  bytes definicao_json = 3; // árvore + campos_definicao consolidados
}

service DeployEngineService {
  rpc Publicar (PublicarRequest) returns (PublicarResponse);
  rpc Rollback (RollbackRequest) returns (RollbackResponse);
  rpc ObterVersaoAtiva (ObterVersaoAtivaRequest) returns (VersaoAtiva);
}
```

**Implementações esperadas**: servidor em `services/deploy/internal/server/grpc.go`; cliente no Gateway.

---

## `construtor.export.v1.ExportEngineService`

Exportação assíncrona com coleta via Server Streaming (RF05, RNF01).

```protobuf
syntax = "proto3";

package construtor.export.v1;

message CriarJobRequest { string sistema_id = 1; }
message Job {
  string id = 1;
  string status = 2;       // criado | coletando | pronto | erro | expirado
  string arquivo_url = 3;  // Presigned URL quando pronto
  string expira_em = 4;    // RFC 3339
}

message ColetarDadosRequest { string sistema_id = 1; }
message ChunkDados {
  string origem = 1; // design | regras | operacional
  bytes conteudo = 2;
}

service ExportEngineService {
  rpc CriarJob (CriarJobRequest) returns (Job);
  rpc ObterJob (ObterJobRequest) returns (Job);
  // Server Streaming: consumo em chunks para não sobrecarregar a RAM (RNF01)
  rpc ColetarDados (ColetarDadosRequest) returns (stream ChunkDados);
}

message ObterJobRequest { string id = 1; }
```

**Implementações esperadas**: servidor em `services/export/internal/{jobs,collector}`; cliente no Gateway. O streaming `ColetarDados` é servido também por Design/Logic Engines como fontes.

---

## `construtor.common.v1` — Contexto de Tenant (Metadata)

```protobuf
syntax = "proto3";

package construtor.common.v1;

// Serializado como Metadata binário gRPC (chave "x-tenant-context-bin") — RNF02.
// NUNCA incluído no corpo de mensagens de negócio.
message TenantContext {
  string tenant_id = 1;
  string user_id = 2;
  string tipo = 3;      // dono | parceiro | cliente
  string trace_state = 4;
}
```

**Implementações esperadas**: `pkg/tenantctx` (Go — interceptores server/client) e plug equivalente no cliente gRPC Elixir.

---

## Contrato de Mensagem AMQP (RabbitMQ)

Evento publicado pelo Logic Engine e consumido pelos workers (RF08, RN09, RNF04):

```json
{
  "headers": {
    "traceparent": "00-<trace-id>-<span-id>-01",
    "x-tenant-id": "uuid",
    "x-component-blind-index": "8f3b2a1..."
  },
  "body": {
    "tipo": "webhook.disparo",
    "payload": { "url_destino": "https://...", "corpo": {} }
  }
}
```

| Propriedade | Regra |
|-------------|-------|
| Routing key | `webhooks.disparo.<tenant_id>` / `notificacoes.envio.<tenant_id>` (fair queuing — RN09) |
| DLQ | `<fila>.dlq` após N tentativas com backoff (RNF06) |
| Trace | `traceparent` obrigatório nos headers (RNF04) |
