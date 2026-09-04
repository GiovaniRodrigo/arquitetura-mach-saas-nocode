# API Contracts: MACH V4 System Builder

Public REST API exposed by the **API Gateway in Go** (`:8080`). All routes (except login and health) require `Authorization: Bearer <JWT>`; the Gateway translates each call to internal gRPC, injecting `tenant_id` as Metadata (NFR02). Validation errors never expose real field names — only `blind_index` (NFR08).

**Common errors across all endpoints:**

| Status | Code | When |
|--------|--------|------|
| 401 | `UNAUTHORIZED` | JWT missing, invalid, or expired |
| 403 | `FORBIDDEN` | Tenant/role without permission for the resource |
| 404 | `NOT_FOUND` | Resource does not exist **within the token's tenant** (BR01) |
| 429 | `RATE_LIMITED` | Tenant request limit exceeded |

---

## Endpoints

### `POST /api/v1/auth/login`

**Description**: Authenticates the user and issues a JWT with `tenant_id`, `sub`, `tipo` claims (FR03).

**Request:**
```json
{
  "email": "string — user email",
  "password": "string — password"
}
```

**Response 200:**
```json
{
  "token": "eyJhbGciOiJSUzI1NiIs...",
  "expires_in": 3600
}
```

**Errors:**
| Status | Code | Message |
|--------|--------|----------|
| 401 | `INVALID_CREDENTIALS` | Invalid credentials |

---

### `POST /api/v1/designs`

**Description**: Creates the UI definition for a screen (recursive Composite tree) via the Design Engine (FR01).

**Request:**
```json
{
  "sistema_id": "uuid",
  "nome": "string — screen name",
  "arvore": {
    "blind_index": "string — root component hash",
    "tipo": "container",
    "propriedades": {},
    "componente_filhos": [
      { "blind_index": "...", "tipo": "input_texto", "propriedades": {}, "componente_filhos": [] }
    ]
  }
}
```

**Response 201:**
```json
{
  "design_id": "uuid",
  "sistema_id": "uuid"
}
```

**Errors:**
| Status | Code | Message |
|--------|--------|----------|
| 422 | `INVALID_TREE` | Structurally invalid recursive tree |

*Also available: `GET /api/v1/designs/{id}`, `PUT /api/v1/designs/{id}`, `DELETE /api/v1/designs/{id}` (FR01).*

---

### `POST /api/v1/regras`

**Description**: Creates a business rule (decision tree) via the Logic Engine (FR02).

**Request:**
```json
{
  "sistema_id": "uuid",
  "arvore_decisao": {
    "no": "condicao",
    "operador": "igual",
    "blind_index": "string — evaluated component",
    "valor": "string",
    "entao": { "no": "acao", "tipo": "webhook.disparo", "config": {} },
    "senao": null
  }
}
```

**Response 201:**
```json
{
  "regra_id": "uuid"
}
```

**Errors:**
| Status | Code | Message |
|--------|--------|----------|
| 422 | `INVALID_DECISION_TREE` | Unknown logic node or cycle detected |

*Also available: `GET /api/v1/regras/{id}`, `PUT /api/v1/regras/{id}`, `DELETE /api/v1/regras/{id}` (FR02).*

---

### `GET /api/v1/permissoes?sistema_id={uuid}`

**Description**: Returns the boolean permission map computed by the IAM Service for the token's user (FR03, BR03). Official format defined in `doc/DATA_SECURITY.md`.

**Response 200:**
```json
{
  "permissions": {
    "8f3b2a1...": { "view": true, "click": false },
    "4a9e2d3...": { "view": false, "click": false }
  }
}
```

---

### `POST /api/v1/sistemas/{sistema_id}/publicar`

**Description**: Publishes a new version — inserts a row into `versoes_sistema` and atomically toggles the active flag (FR04, BR04).

**Request:** *(empty body — the version is assembled from the current state of the designs/rules)*

**Response 201:**
```json
{
  "versao_id": "uuid",
  "numero": 7,
  "ativa": true
}
```

**Errors:**
| Status | Code | Message |
|--------|--------|----------|
| 409 | `PUBLISH_IN_PROGRESS` | Another publish already in progress for the system |

---

### `POST /api/v1/sistemas/{sistema_id}/rollback`

**Description**: Reactivates the previous stable version (or the one specified) by toggling the flag — with no downtime (FR04, BR05, NFR05).

**Request:**
```json
{
  "versao_numero": 6
}
```
*`versao_numero` is optional; if omitted, defaults to the immediately previous version.*

**Response 200:**
```json
{
  "versao_ativa": 6
}
```

**Errors:**
| Status | Code | Message |
|--------|--------|----------|
| 422 | `NO_PREVIOUS_VERSION` | No previous version to roll back to |

---

### `GET /api/v1/sistemas/{sistema_id}/versao-ativa`

**Description**: Returns the full definition of the active version — consumed by the Headless Player (BR04).

**Response 200:**
```json
{
  "versao_id": "uuid",
  "numero": 7,
  "definicao": { "arvore": { "...": "..." }, "campos": { "8f3b2a1...": { "tipo": "string", "obrigatorio": true, "limites": { "max_length": 120 } } } }
}
```

---

### `POST /api/v1/formularios`

**Description**: Submits dynamic operational data. Translated to `LogicEngineService.SalvarFormulario` (Unary RPC) — FR07, BR08. `blind_index → value` map per the official `.proto` contract.

**Request:**
```json
{
  "sistema_id": "uuid",
  "dados_formulario": {
    "8f3b2a1...": "João Silva",
    "4a9e2d3...": "42"
  }
}
```

**Response 200:**
```json
{
  "sucesso": true,
  "mensagem_status": "Record saved"
}
```

**Response 422 (validation — NFR08):**
```json
{
  "sucesso": false,
  "erros_validacao": {
    "4a9e2d3...": "value exceeds the maximum allowed"
  },
  "mensagem_status": "Validation failed"
}
```

**Errors:**
| Status | Code | Message |
|--------|--------|----------|
| 422 | `VALIDATION_ERROR` | `erros_validacao` map indexed by blind_index |

---

### `POST /api/v1/exportacoes`

**Description**: Creates an asynchronous export Job and responds immediately (FR05).

**Request:**
```json
{
  "sistema_id": "uuid"
}
```

**Response 202:**
```json
{
  "job_id": "uuid",
  "status": "criado"
}
```

---

### `GET /api/v1/exportacoes/{job_id}`

**Description**: Queries the Job status; when `pronto`, includes a short-lived Presigned URL (FR05).

**Response 200:**
```json
{
  "job_id": "uuid",
  "status": "pronto",
  "arquivo_url": "https://storage.example.com/exports/...?X-Amz-Expires=600",
  "expira_em": "2026-07-13T23:59:00Z"
}
```

**Errors:**
| Status | Code | Message |
|--------|--------|----------|
| 410 | `EXPORT_EXPIRED` | Download link expired — request a new export |

---

## WebSocket (Collaboration Engine — Elixir, `:4000`)

### `ws://host:4000/socket` → channel `screen:{sistema_id}:{screen_id}`

**Description**: Persistent connection for collaborative editing (FR06). `join` requires the same JWT; the channel rejects tenants without access to the system (BR01).

**Client → server events:**
| Event | Payload | Description |
|--------|---------|-----------|
| `mutation` | `{ "blind_index": "...", "op": "move\|update\|insert\|delete", "dados": {} }` | Component mutation; resets the 5s debounce (BR06) |
| `lock` | `{ "blind_index": "..." }` | Requests optimistic lock on the component (BR07) |
| `unlock` | `{ "blind_index": "..." }` | Releases the lock |

**Server → client events:**
| Event | Payload | Description |
|--------|---------|-----------|
| `mutation` | broadcast of the applied mutation | Real-time synchronization |
| `presence_state` / `presence_diff` | Phoenix Presence state | Online users and cursors |
| `locked` | `{ "blind_index": "...", "por": "user_id" }` | Component locked — disable input (BR07) |
| `flush_ok` | `{ "versao_rascunho": n }` | Confirmation of batched persistence via gRPC (BR06) |
