# API Contracts: Resource Monitor

## `GET /api/v1/monitor/recursos`

**Description**: The Gateway's REST facade over `MonitorService.ObterRecursos` (FR05).
Consumed by the Frontend's Monitor screen (FR06/FR07).

**Request:** headers `Authorization: Bearer <JWT>` (the same authenticated group as the
other routes — NFR05); no body.

**Response 200:**
```json
{
  "servicos": [
    {
      "nome": "iam",
      "tipo": "grpc",
      "status": "servindo",
      "uptime_segundos": 3600,
      "memoria_alocada_bytes": 15728640,
      "memoria_sistema_bytes": 41943040,
      "goroutines": 24,
      "mensagem_erro": ""
    },
    {
      "nome": "logic",
      "tipo": "grpc",
      "status": "indisponivel",
      "uptime_segundos": 0,
      "memoria_alocada_bytes": 0,
      "memoria_sistema_bytes": 0,
      "goroutines": 0,
      "mensagem_erro": "context deadline exceeded"
    }
  ],
  "coletado_em_unix": 1755158400
}
```
An array always with the 8 fixed entries (IAM, Design, Logic, Deploy, Export, Gateway,
Workers, Collab), in the same order — a service's failure becomes `status: "indisponivel"`
on its own entry, never shrinks the array (BR01, acceptance criterion 2).

**Errors:**
| Status | Code | Message |
|--------|--------|----------|
| 401 | `UNAUTHORIZED` | Missing/invalid token |
| 502/5xx | `MONITOR_INDISPONIVEL` | The Monitor service itself didn't respond — the Frontend treats this as a single screen-wide error (NFR02, acceptance criterion 3), distinct from an individual "indisponivel" entry inside a 200. |

---

## `GET /health` (Workers) — new

**Description**: Workers' resource endpoint (FR03). No server exists today in the
Workers process; this is the first one.

**Request:** no authentication (internal use, queried only by the Monitor — the same
pattern as the Gateway's `/health`, which is also public/no auth).

**Response 200:**
```json
{
  "status": "servindo",
  "uptime_segundos": 3600,
  "memoria_alocada_bytes": 8388608,
  "memoria_sistema_bytes": 20971520,
  "goroutines": 12
}
```

---

## `GET /healthz` (Collab) — extended

**Description**: Already exists (`services/collab/lib/collab_web/endpoint.ex:48-52`),
today it only answers with an HTTP status and no meaningful body. It now includes a JSON
body (FR02), keeping the current HTTP 200 status (doesn't break any existing consumer
that only looks at the status code).

**Response 200 (new body):**
```json
{
  "status": "servindo",
  "uptime_segundos": 3600,
  "memoria_alocada_bytes": 52428800
}
```
No `memoria_sistema_bytes`/`goroutines` — the BEAM doesn't expose a direct equivalent
(BR02, `plan.md` §4.3); the Monitor's `ColetorHTTP` treats these two fields as
absent/0 for this service.
