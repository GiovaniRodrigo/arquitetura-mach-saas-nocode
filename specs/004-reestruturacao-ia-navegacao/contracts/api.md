# Contratos de API: Reestruturação de IA e Regras de Negócio

Endpoints **assumidos** — hoje não existem no Gateway. Seguem a convenção já usada por
`player/src/api/client.ts` (`Authorization: Bearer`, tenant nunca no corpo — RN01 de
001) e o formato de erro de `ApiError` (`{ codigo, mensagem }`). Até serem
implementados no backend, a Fase 2 de `tasks.md` desenvolve a UI contra este contrato
com `fetch` mockado nos testes (mesmo padrão de `contracts/api.md` de 003, que também
assumia campos "a fornecer pelo Gateway").

---

## `GET /api/v1/dashboard/ultimos-acessos`

**Descrição**: 10 logins mais recentes entre os tenants vinculados ao usuário autenticado (RF04, RN02).

**Response 200:**
```json
{
  "eventos": [
    { "usuario_nome": "string", "tenant_nome": "string", "criado_em": "2026-08-06T12:00:00Z" }
  ]
}
```

## `GET /api/v1/dashboard/feedback?status=pendente|respondido`

**Descrição**: mensagens de feedback dos tenants vinculados (RF05, RN03).

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

**Response 200:** item atualizado (mesmo shape acima).

## `GET /api/v1/dashboard/resumo-financeiro`

**Descrição**: receita de assinatura/cobrança agregada dos tenants vinculados (RF06, RN04).

**Response 200:**
```json
{ "receita_total_centavos": 0, "moeda": "BRL", "competencia": "2026-08" }
```

## `GET /api/v1/tenants`

**Descrição**: tenants (clientes/negócios) vinculados ao usuário autenticado (RF07).

**Response 200:** `{ "tenants": [ { "id": "uuid", "nome": "string" } ] }`

## `GET /api/v1/sistemas?tenant_id={id}`

**Descrição**: extensão de `listarSistemas()` já existente, com filtro por tenant (RF08).

## `GET|POST /api/v1/sistemas/{id}/regras-negocio`

**Descrição**: CRUD de regras de validação de componente (RF10/RF11).

**Request (POST):**
```json
{ "blind_indexes": ["bi_1"], "tipo": "regex|tamanho|obrigatorio", "parametros": {} }
```

## `GET /api/v1/sistemas/{id}/versoes` · `POST /api/v1/sistemas/{id}/versoes/{versaoId}/publicar` · `POST /api/v1/sistemas/{id}/versoes/{versaoId}/reverter`

**Descrição**: lista/publica/reverte versão (RF12) — reaproveita a semântica de RN04/RN05 já definida em `001-construtor-sistemas-mach-v4`.

## `PUT /api/v1/configuracao/white-label`

**Request:** `{ "logo_url": "string", "cor_primaria": "#rrggbb", "cor_secundaria": "#rrggbb", "dominio_proprio": "string" }` (RF13, RNF03 — `dominio_validado` é assíncrono, backend responde `202` até a validação concluir)

## `PUT /api/v1/conta/senha`

**Request:** `{ "senha_atual": "string", "senha_nova": "string" }` (RF14, RNF02)

## `POST /api/v1/conta/mfa/ativar` · `POST /api/v1/conta/mfa/confirmar` · `DELETE /api/v1/conta/mfa`

**Descrição**: ativação TOTP em duas etapas — `ativar` devolve `{ "segredo_otp_auth_uri": "..." }` (exibição única); `confirmar` recebe `{ "codigo": "123456" }` (RF15, RNF01).

## `DELETE /api/v1/conta`

**Descrição**: exclui a conta autenticada (RF16, RN07).

**Erros:**
| Status | Código | Mensagem |
|--------|--------|----------|
| 409 | `TENANT_ATIVO_VINCULADO` | Existem tenants ativos vinculados a esta conta. |
| 401 | `REAUTENTICACAO_NECESSARIA` | Confirme sua senha para continuar. |

## `PATCH /api/v1/conta/perfil`

**Request:** `{ "nome": "string", "foto_url": "string" }` (RF17)

## `POST /api/v1/conta/email` · `POST /api/v1/conta/email/confirmar`

**Descrição**: `POST /email` envia o link/código ao novo endereço sem alterar o e-mail de login; `POST /email/confirmar` (com o token recebido) efetiva a troca (RF18, RN08).
