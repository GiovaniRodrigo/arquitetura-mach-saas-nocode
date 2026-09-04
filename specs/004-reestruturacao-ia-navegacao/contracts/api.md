# API Contracts: AI and Business Rules Restructuring

**Assumed** endpoints — they do not exist in the Gateway today. They follow the convention already used by
`player/src/api/client.ts` (`Authorization: Bearer`, tenant never in the body — BR01 from
001) and the `ApiError` error format (`{ codigo, mensagem }`). Until they are
implemented in the backend, Phase 2 of `tasks.md` builds the UI against this contract
with mocked `fetch` in tests (same pattern as `contracts/api.md` from 003, which also
assumed fields "to be provided by the Gateway").

---

## `GET /api/v1/dashboard/ultimos-acessos`

**Description**: the 10 most recent logins across the tenants linked to the authenticated user (FR04, BR02).

**Response 200:**
```json
{
  "eventos": [
    { "usuario_nome": "string", "tenant_nome": "string", "criado_em": "2026-08-06T12:00:00Z" }
  ]
}
```

## `GET /api/v1/dashboard/feedback?status=pendente|respondido`

**Description**: feedback messages from the linked tenants (FR05, BR03).

**Response 200:**
```json
{
  "itens": [
    { "id": "uuid", "tenant_nome": "string", "mensagem": "string", "status": "pendente", "criado_em": "2026-08-06T12:00:00Z" }
  ]
}
```

## `PATCH /api/v1/dashboard/feedback/{id}`

**Request:** `{ "status": "respondido" }`

**Response 200:** updated item (same shape as above).

## `GET /api/v1/dashboard/resumo-financeiro`

**Description**: aggregated subscription/billing revenue from the linked tenants (FR06, BR04).

**Response 200:**
```json
{ "receita_total_centavos": 0, "moeda": "BRL", "competencia": "2026-08" }
```

## `GET /api/v1/tenants`

**Description**: tenants (customers/businesses) linked to the authenticated user (FR07).

**Response 200:** `{ "tenants": [ { "id": "uuid", "nome": "string" } ] }`

## `GET /api/v1/sistemas?tenant_id={id}`

**Description**: extension of the existing `listarSistemas()`, filtered by tenant (FR08).

## `GET|POST /api/v1/sistemas/{id}/regras-negocio`

**Description**: CRUD for component validation rules (FR10/FR11).

**Request (POST):**
```json
{ "blind_indexes": ["bi_1"], "tipo": "regex|tamanho|obrigatorio", "parametros": {} }
```

## `GET /api/v1/sistemas/{id}/versoes` · `POST /api/v1/sistemas/{id}/versoes/{versaoId}/publicar` · `POST /api/v1/sistemas/{id}/versoes/{versaoId}/reverter`

**Description**: list/publish/roll back a version (FR12) — reuses the BR04/BR05 semantics already defined in `001-construtor-sistemas-mach-v4`.

## `PUT /api/v1/configuracao/white-label`

**Request:** `{ "logo_url": "string", "cor_primaria": "#rrggbb", "cor_secundaria": "#rrggbb", "dominio_proprio": "string" }` (FR13, NFR03 — `dominio_validado` is asynchronous, the backend responds `202` until validation completes)

## `PUT /api/v1/conta/senha`

**Request:** `{ "senha_atual": "string", "senha_nova": "string" }` (FR14, NFR02)

## `POST /api/v1/conta/mfa/ativar` · `POST /api/v1/conta/mfa/confirmar` · `DELETE /api/v1/conta/mfa`

**Description**: two-step TOTP activation — `ativar` returns `{ "segredo_otp_auth_uri": "..." }` (shown once); `confirmar` receives `{ "codigo": "123456" }` (FR15, NFR01).

## `DELETE /api/v1/conta`

**Description**: deletes the authenticated account (FR16, BR07).

**Errors:**
| Status | Code | Message |
|--------|--------|----------|
| 409 | `TENANT_ATIVO_VINCULADO` | There are active tenants linked to this account. |
| 401 | `REAUTENTICACAO_NECESSARIA` | Confirm your password to continue. |

## `PATCH /api/v1/conta/perfil`

**Request:** `{ "nome": "string", "foto_url": "string" }` (FR17)

## `POST /api/v1/conta/email` · `POST /api/v1/conta/email/confirmar`

**Description**: `POST /email` sends the link/code to the new address without changing the login email; `POST /email/confirmar` (with the received token) applies the change (FR18, BR08).
