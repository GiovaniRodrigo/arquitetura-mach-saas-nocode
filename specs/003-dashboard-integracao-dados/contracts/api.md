# Contratos de API: Dashboard — Integração de Dados e Funcionalidade

A Fase 1 **não cria endpoints novos** — consome os já existentes (definidos em
`specs/001-construtor-sistemas-mach-v4/contracts/api.md`). A identidade viaja sempre em
`Authorization: Bearer <JWT>`; o tenant é derivado do token pelo Gateway (RN01). Este
documento registra o consumo atual e as **extensões propostas para a Fase 2**.

---

## Endpoints consumidos (existentes — Fase 1)

### `GET /api/v1/sistemas`

**Descrição**: Lista os sistemas do tenant. Base de `Projects` (RF02) e das métricas do
`Overview` (RF01).

**Request:** headers `Authorization: Bearer <JWT>`; sem corpo.

**Response 200:**
```json
{
  "sistemas": [
    { "id": "uuid", "nome": "string" }
  ]
}
```

**Erros:**
| Status | Código | Mensagem |
|--------|--------|----------|
| 401    | `UNAUTHORIZED` | Token ausente/inválido |
| 5xx    | `UNKNOWN` | Falha do Gateway → estado de erro na UI (repetir) |

### `POST /api/v1/sistemas`

**Descrição**: Cria um sistema. Acionado por "Get Started"/FAB/card "Criar novo projeto"
(RF04). Restrito a dono/parceiro.

**Request:**
```json
{ "nome": "string — nome do novo sistema" }
```

**Response 200/201:**
```json
{ "id": "uuid", "nome": "string" }
```

**Erros:**
| Status | Código | Mensagem |
|--------|--------|----------|
| 403    | `FORBIDDEN` | Cliente-final sem permissão → CTA oculto/desabilitado (RN10) |
| 422    | `VALIDATION_ERROR` | Nome inválido |

### `GET /api/v1/sistemas/{sistema_id}/versao-ativa`

**Descrição**: Versão ativa consolidada; base do status/versão por card (RF07, RN04).
Consumo opcional na Fase 2 (por card) ou substituído pelo payload enriquecido de `Sistema`.

---

## Extensões propostas (Fase 2 — a implementar no Gateway)

### `GET /api/v1/sistemas` — payload enriquecido

**Descrição**: Estender cada item com status, versão ativa, timestamp e DLQ para os cards
ricos (RF07, RF09). Retrocompatível: campos novos são opcionais.

**Response 200 (proposta):**
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

### `GET /api/v1/metricas` — métricas agregadas do tenant (proposta)

**Descrição**: Contadores para o `Overview` (RF01) sem o cliente precisar derivá-los.

**Response 200 (proposta):**
```json
{
  "sistemas_total": 12,
  "sistemas_ativos": 8,
  "sistemas_rascunho": 4
}
```

**Erros:**
| Status | Código | Mensagem |
|--------|--------|----------|
| 401    | `UNAUTHORIZED` | Token ausente/inválido |

> Enquanto `GET /api/v1/metricas` não existir, `useMetricas` deriva os contadores de
> `GET /api/v1/sistemas` (total garantido; ativos/rascunho dependem do payload enriquecido).

---

## Presença em tempo real (Fase 2 — WebSocket)

Canal Phoenix já disponível em `player/src/collab/phoenixSocket.ts`. Para RF08, o
dashboard assinaria um tópico por sistema (ex.: `sistema:{id}`) e escutaria eventos de
presença (`presence_state`/`presence_diff`) para renderizar os avatares empilhados. O
provisionamento server-side dos tópicos é pré-requisito e está fora do escopo de
front-end desta demanda.
