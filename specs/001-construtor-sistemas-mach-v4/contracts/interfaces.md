# Interfaces: MACH V4 System Builder

gRPC contracts (`.proto`) — the single source of truth for internal communication (NFR01). All services receive `tenant_id` exclusively via gRPC Metadata (key `x-tenant-context-bin`), never in the message body (NFR02, BR01).

---

## `construtor.logic.v1.LogicEngineService`

Official contract transcribed from `doc/CONTRACTS_PERFORMANCE.md §5`, extended with the rules CRUD (FR02, FR07).

```protobuf
syntax = "proto3";

package construtor.logic.v1;

// Represents the submission of dynamic operational data
message SalvarFormularioRequest {
  string sistema_id = 1;
  // Dynamic key-value map linking the input's Blind Index to the submitted value
  map<string, string> dados_formulario = 2;
}

// Response with the validation and persistence status
message SalvarFormularioResponse {
  bool sucesso = 1;
  // Error map indexed by the Blind Index of the failed component (NFR08)
  map<string, string> erros_validacao = 2;
  string mensagem_status = 3;
}

message Regra {
  string id = 1;
  string sistema_id = 2;
  // Serialized decision tree (logic nodes)
  bytes arvore_decisao = 3;
}

service LogicEngineService {
  // Unary action to validate and write data to the shared database (FR07, BR08)
  rpc SalvarFormulario (SalvarFormularioRequest) returns (SalvarFormularioResponse);

  rpc CriarRegra (Regra) returns (Regra);
  rpc ObterRegra (ObterRegraRequest) returns (Regra);
  rpc AtualizarRegra (Regra) returns (Regra);
  rpc RemoverRegra (ObterRegraRequest) returns (RemoverResponse);
}

message ObterRegraRequest { string id = 1; }
message RemoverResponse { bool sucesso = 1; }
```

**Expected implementations**: server in `services/logic/internal/server/grpc.go`; clients in the Gateway and the Export Engine.

---

## `construtor.design.v1.DesignEngineService`

CRUD of the recursive tree and batched persistence of collaboration (FR01, BR06).

```protobuf
syntax = "proto3";

package construtor.design.v1;

message Componente {
  string blind_index = 1;
  string tipo = 2;
  bytes propriedades = 3;              // Serialized JSON
  repeated Componente componente_filhos = 4;  // Composite pattern
}

message Design {
  string id = 1;
  string sistema_id = 2;
  string nome = 3;
  Componente arvore = 4;
}

message SalvarDesignRequest {
  // Consolidated payload sent by the Elixir engine after the 5s debounce (BR06)
  Design design = 1;
}

service DesignEngineService {
  rpc CriarDesign (Design) returns (Design);
  rpc ObterDesign (ObterDesignRequest) returns (Design);
  rpc AtualizarDesign (Design) returns (Design);
  rpc RemoverDesign (ObterDesignRequest) returns (RemoverResponse);
  // Single batched call from collaboration (write-behind)
  rpc SalvarDesign (SalvarDesignRequest) returns (SalvarDesignResponse);
}

message ObterDesignRequest { string id = 1; }
message SalvarDesignResponse { bool sucesso = 1; }
message RemoverResponse { bool sucesso = 1; }
```

**Expected implementations**: server in `services/design/internal/server/grpc.go`; clients in the Gateway, the Elixir engine (`collab/lib/collab/grpc/design_client.ex`), and the Export Engine.

---

## `construtor.iam.v1.IAMService`

Authentication and per-component permissions (FR03, BR03).

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
  // Blind indexes of the components on the screen being rendered
  repeated string blind_indexes = 2;
}

message PermissaoComponente {
  bool view = 1;
  bool click = 2;
}

message AvaliarPermissoesResponse {
  // Final boolean map — rule logic never leaves the server (BR03)
  map<string, PermissaoComponente> permissions = 1;
}

service IAMService {
  rpc ValidarToken (ValidarTokenRequest) returns (ValidarTokenResponse);
  rpc AvaliarPermissoes (AvaliarPermissoesRequest) returns (AvaliarPermissoesResponse);
}
```

**Expected implementations**: server in `services/iam/internal/server/grpc.go`; client in the Gateway (auth middleware and permissions route).

---

## `construtor.deploy.v1.DeployEngineService`

Publishing via active flag and instant rollback (FR04, BR04, BR05).

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
  // 0 = immediately previous version
  int32 versao_numero = 2;
}
message RollbackResponse { int32 versao_ativa = 1; }

message ObterVersaoAtivaRequest { string sistema_id = 1; }
message VersaoAtiva {
  string versao_id = 1;
  int32 numero = 2;
  bytes definicao_json = 3; // consolidated tree + campos_definicao
}

service DeployEngineService {
  rpc Publicar (PublicarRequest) returns (PublicarResponse);
  rpc Rollback (RollbackRequest) returns (RollbackResponse);
  rpc ObterVersaoAtiva (ObterVersaoAtivaRequest) returns (VersaoAtiva);
}
```

**Expected implementations**: server in `services/deploy/internal/server/grpc.go`; client in the Gateway.

---

## `construtor.export.v1.ExportEngineService`

Asynchronous export with collection via Server Streaming (FR05, NFR01).

```protobuf
syntax = "proto3";

package construtor.export.v1;

message CriarJobRequest { string sistema_id = 1; }
message Job {
  string id = 1;
  string status = 2;       // criado | coletando | pronto | erro | expirado
  string arquivo_url = 3;  // Presigned URL when ready
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
  // Server Streaming: consumed in chunks to avoid overloading RAM (NFR01)
  rpc ColetarDados (ColetarDadosRequest) returns (stream ChunkDados);
}

message ObterJobRequest { string id = 1; }
```

**Expected implementations**: server in `services/export/internal/{jobs,collector}`; client in the Gateway. The `ColetarDados` stream is also served by the Design/Logic Engines as sources.

---

## `construtor.common.v1` — Tenant Context (Metadata)

```protobuf
syntax = "proto3";

package construtor.common.v1;

// Serialized as binary gRPC Metadata (key "x-tenant-context-bin") — NFR02.
// NEVER included in business message bodies.
message TenantContext {
  string tenant_id = 1;
  string user_id = 2;
  string tipo = 3;      // dono | parceiro | cliente
  string trace_state = 4;
}
```

**Expected implementations**: `pkg/tenantctx` (Go — server/client interceptors) and an equivalent plug in the Elixir gRPC client.

---

## AMQP Message Contract (RabbitMQ)

Event published by the Logic Engine and consumed by the workers (FR08, BR09, NFR04):

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

| Property | Rule |
|-------------|-------|
| Routing key | `webhooks.disparo.<tenant_id>` / `notificacoes.envio.<tenant_id>` (fair queuing — BR09) |
| DLQ | `<fila>.dlq` after N attempts with backoff (NFR06) |
| Trace | `traceparent` required in the headers (NFR04) |
