# API Contracts: Dashboard — Data Integration and Functionality

Phase 1 **does not create new endpoints** — it consumes the ones that already exist
(defined in `specs/001-construtor-sistemas-mach-v4/contracts/api.md`). Identity always travels in
`Authorization: Bearer <JWT>`; the tenant is derived from the token by the Gateway (BR01). This
document records the current consumption and the **proposed extensions for Phase 2**.

---

## Endpoints consumed (existing — Phase 1)

### `GET /api/v1/sistemas`

**Description**: Lists the tenant's systems. Basis for `Projects` (FR02) and the `Overview`
metrics (FR01).

**Request:** headers `Authorization: Bearer <JWT>`; no body.

**Response 200:**
```json
{
  "sistemas": [
    { "id": "uuid", "nome": "string" }
  ]
}
```

**Errors:**
| Status | Code | Message |
|--------|--------|----------|
| 401    | `UNAUTHORIZED` | Missing/invalid token |
| 5xx    | `UNKNOWN` | Gateway failure → error state in the UI (retry) |

### `POST /api/v1/sistemas`

**Description**: Creates a system. Triggered by "Get Started"/FAB/"Create new project" card
(FR04). Restricted to owner/partner.

**Request:**
```json
{ "nome": "string — name of the new system" }
```

**Response 200/201:**
```json
{ "id": "uuid", "nome": "string" }
```

**Errors:**
| Status | Code | Message |
|--------|--------|----------|
| 403    | `FORBIDDEN` | End-customer without permission → CTA hidden/disabled (BR10) |
| 422    | `VALIDATION_ERROR` | Invalid name |

### `GET /api/v1/sistemas/{sistema_id}/versao-ativa`

**Description**: Consolidated active version; basis for the per-card status/version (FR07, BR04).
Optional consumption in Phase 2 (per card) or replaced by the enriched `Sistema` payload.

---

## Proposed extensions (Phase 2 — to be implemented on the Gateway)

### `GET /api/v1/sistemas` — enriched payload

**Description**: Extend each item with status, active version, timestamp, and DLQ for the
rich cards (FR07, FR09). Backward-compatible: the new fields are optional.

**Response 200 (proposal):**
```json
{
  "sistemas": [
    {
      "id": "uuid",
      "nome": "string",
      "status": "publicado | rascunho | falha",
      "versao_ativa": { "numero": 7, "rotulo": "v7 · ativa" },
      "atualizado_em": "2026-07-15T12:00:00Z",
      "dlq_eventos": 0
    }
  ]
}
```

### `GET /api/v1/metricas` — aggregated tenant metrics (proposal)

**Description**: Counters for the `Overview` (FR01) without the client having to derive them.

**Response 200 (proposal):**
```json
{
  "sistemas_total": 12,
  "sistemas_ativos": 8,
  "sistemas_rascunho": 4
}
```

**Errors:**
| Status | Code | Message |
|--------|--------|----------|
| 401    | `UNAUTHORIZED` | Missing/invalid token |

> Until `GET /api/v1/metricas` exists, `useMetricas` derives the counters from
> `GET /api/v1/sistemas` (total guaranteed; active/draft depend on the enriched payload).

---

## Real-time presence (Phase 2 — WebSocket)

The Phoenix channel is already available at `player/src/collab/phoenixSocket.ts`. For FR08, the
dashboard would subscribe to a per-system topic (e.g., `sistema:{id}`) and listen for presence
events (`presence_state`/`presence_diff`) to render the stacked avatars. Server-side
provisioning of the topics is a prerequisite and is outside this effort's front-end
scope.
