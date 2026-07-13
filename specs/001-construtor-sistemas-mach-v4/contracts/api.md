# Contratos de API: Construtor de Sistemas MACH V4

API REST pública exposta pelo **API Gateway em Go** (`:8080`). Todas as rotas (exceto login e health) exigem `Authorization: Bearer <JWT>`; o Gateway traduz cada chamada para gRPC interno injetando `tenant_id` como Metadata (RNF02). Erros de validação nunca expõem nomes reais de campos — apenas `blind_index` (RNF08).

**Erros comuns a todos os endpoints:**

| Status | Código | Quando |
|--------|--------|--------|
| 401 | `UNAUTHORIZED` | JWT ausente, inválido ou expirado |
| 403 | `FORBIDDEN` | Tenant/papel sem permissão para o recurso |
| 404 | `NOT_FOUND` | Recurso inexistente **no tenant do token** (RN01) |
| 429 | `RATE_LIMITED` | Limite de requisições do tenant excedido |

---

## Endpoints

### `POST /api/v1/auth/login`

**Descrição**: Autentica o utilizador e emite JWT com claims `tenant_id`, `sub`, `tipo` (RF03).

**Request:**
```json
{
  "email": "string — e-mail do utilizador",
  "password": "string — senha"
}
```

**Response 200:**
```json
{
  "token": "eyJhbGciOiJSUzI1NiIs...",
  "expires_in": 3600
}
```

**Erros:**
| Status | Código | Mensagem |
|--------|--------|----------|
| 401 | `INVALID_CREDENTIALS` | Credenciais inválidas |

---

### `POST /api/v1/designs`

**Descrição**: Cria a definição de UI de um ecrã (árvore recursiva Composite) via Design Engine (RF01).

**Request:**
```json
{
  "sistema_id": "uuid",
  "nome": "string — nome do ecrã",
  "arvore": {
    "blind_index": "string — hash do componente raiz",
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

**Erros:**
| Status | Código | Mensagem |
|--------|--------|----------|
| 422 | `INVALID_TREE` | Árvore recursiva estruturalmente inválida |

*Também disponíveis: `GET /api/v1/designs/{id}`, `PUT /api/v1/designs/{id}`, `DELETE /api/v1/designs/{id}` (RF01).*

---

### `POST /api/v1/regras`

**Descrição**: Cria uma regra de negócio (árvore de decisão) via Logic Engine (RF02).

**Request:**
```json
{
  "sistema_id": "uuid",
  "arvore_decisao": {
    "no": "condicao",
    "operador": "igual",
    "blind_index": "string — componente avaliado",
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

**Erros:**
| Status | Código | Mensagem |
|--------|--------|----------|
| 422 | `INVALID_DECISION_TREE` | Nó lógico desconhecido ou ciclo detectado |

*Também disponíveis: `GET /api/v1/regras/{id}`, `PUT /api/v1/regras/{id}`, `DELETE /api/v1/regras/{id}` (RF02).*

---

### `GET /api/v1/permissoes?sistema_id={uuid}`

**Descrição**: Retorna o mapa booleano de permissões calculado pelo IAM Service para o utilizador do token (RF03, RN03). Formato oficial de `doc/DATA_SECURITY.md`.

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

**Descrição**: Publica nova versão — insere linha em `versoes_sistema` e alterna a flag ativa atomicamente (RF04, RN04).

**Request:** *(corpo vazio — a versão é montada a partir do estado atual dos designs/regras)*

**Response 201:**
```json
{
  "versao_id": "uuid",
  "numero": 7,
  "ativa": true
}
```

**Erros:**
| Status | Código | Mensagem |
|--------|--------|----------|
| 409 | `PUBLISH_IN_PROGRESS` | Outra publicação em andamento para o sistema |

---

### `POST /api/v1/sistemas/{sistema_id}/rollback`

**Descrição**: Reativa a versão estável anterior (ou a indicada) alternando a flag — sem downtime (RF04, RN05, RNF05).

**Request:**
```json
{
  "versao_numero": 6
}
```
*`versao_numero` opcional; omitido = versão imediatamente anterior.*

**Response 200:**
```json
{
  "versao_ativa": 6
}
```

**Erros:**
| Status | Código | Mensagem |
|--------|--------|----------|
| 422 | `NO_PREVIOUS_VERSION` | Não há versão anterior para reverter |

---

### `GET /api/v1/sistemas/{sistema_id}/versao-ativa`

**Descrição**: Retorna a definição completa da versão ativa — consumida pelo Headless Player (RN04).

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

**Descrição**: Submete dados operacionais dinâmicos. Traduzido para `LogicEngineService.SalvarFormulario` (Unary RPC) — RF07, RN08. Mapa `blind_index → valor` conforme contrato `.proto` oficial.

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
  "mensagem_status": "Registo gravado"
}
```

**Response 422 (validação — RNF08):**
```json
{
  "sucesso": false,
  "erros_validacao": {
    "4a9e2d3...": "valor acima do máximo permitido"
  },
  "mensagem_status": "Falha de validação"
}
```

**Erros:**
| Status | Código | Mensagem |
|--------|--------|----------|
| 422 | `VALIDATION_ERROR` | Mapa `erros_validacao` indexado por blind_index |

---

### `POST /api/v1/exportacoes`

**Descrição**: Cria Job de exportação assíncrona e responde imediatamente (RF05).

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

**Descrição**: Consulta o estado do Job; quando `pronto`, inclui Presigned URL de expiração curta (RF05).

**Response 200:**
```json
{
  "job_id": "uuid",
  "status": "pronto",
  "arquivo_url": "https://storage.exemplo.com/exports/...?X-Amz-Expires=600",
  "expira_em": "2026-07-13T23:59:00Z"
}
```

**Erros:**
| Status | Código | Mensagem |
|--------|--------|----------|
| 410 | `EXPORT_EXPIRED` | Link de download expirado — solicite nova exportação |

---

## WebSocket (Motor de Colaboração — Elixir, `:4000`)

### `ws://host:4000/socket` → canal `screen:{sistema_id}:{screen_id}`

**Descrição**: Conexão persistente para edição colaborativa (RF06). `join` exige o mesmo JWT; o canal rejeita tenants sem acesso ao sistema (RN01).

**Eventos cliente → servidor:**
| Evento | Payload | Descrição |
|--------|---------|-----------|
| `mutation` | `{ "blind_index": "...", "op": "move\|update\|insert\|delete", "dados": {} }` | Mutação de componente; reinicia o debounce de 5s (RN06) |
| `lock` | `{ "blind_index": "..." }` | Solicita bloqueio otimista do componente (RN07) |
| `unlock` | `{ "blind_index": "..." }` | Libera bloqueio |

**Eventos servidor → cliente:**
| Evento | Payload | Descrição |
|--------|---------|-----------|
| `mutation` | broadcast da mutação aplicada | Sincronização em tempo real |
| `presence_state` / `presence_diff` | estado Phoenix Presence | Utilizadores online e cursores |
| `locked` | `{ "blind_index": "...", "por": "user_id" }` | Componente bloqueado — desativar input (RN07) |
| `flush_ok` | `{ "versao_rascunho": n }` | Confirmação da persistência em lote via gRPC (RN06) |
