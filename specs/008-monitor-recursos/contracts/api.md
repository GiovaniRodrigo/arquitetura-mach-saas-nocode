# Contratos de API: Monitor de Recursos

## `GET /api/v1/monitor/recursos`

**Descrição**: Fachada REST do Gateway sobre `MonitorService.ObterRecursos` (RF05).
Consumida pela tela Monitor do Frontend (RF06/RF07).

**Request:** headers `Authorization: Bearer <JWT>` (mesmo grupo autenticado das demais
rotas — RNF05); sem corpo.

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
Um array sempre com as 8 entradas fixas (IAM, Design, Logic, Deploy, Export, Gateway,
Workers, Collab), na mesma ordem — a falha de um serviço vira `status: "indisponivel"` na
sua própria entrada, nunca reduz o array (RN01, critério de aceitação 2).

**Erros:**
| Status | Código | Mensagem |
|--------|--------|----------|
| 401 | `UNAUTHORIZED` | Token ausente/inválido |
| 502/5xx | `MONITOR_INDISPONIVEL` | O serviço Monitor em si não respondeu — Frontend trata como erro único da tela (RNF02, critério de aceitação 3), distinto de uma entrada individual "indisponivel" dentro de um 200. |

---

## `GET /health` (Workers) — novo

**Descrição**: Endpoint de recursos do Workers (RF03). Não existe hoje nenhum servidor no
processo Workers; este é o primeiro.

**Request:** sem autenticação (uso interno, consultado só pelo Monitor — mesmo padrão do
`/health` do Gateway, que também é público/sem auth).

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

## `GET /healthz` (Collab) — estendido

**Descrição**: Já existe (`services/collab/lib/collab_web/endpoint.ex:48-52`), hoje
responde só status HTTP sem corpo relevante. Passa a incluir corpo JSON (RF02), mantendo
o status HTTP 200 atual (não quebra nenhum consumidor existente que só olhe o status code).

**Response 200 (novo corpo):**
```json
{
  "status": "servindo",
  "uptime_segundos": 3600,
  "memoria_alocada_bytes": 52428800
}
```
Sem `memoria_sistema_bytes`/`goroutines` — a BEAM não expõe um equivalente direto (RN02,
`plan.md` §4.3); o `ColetorHTTP` do Monitor trata esses dois campos como ausentes/0 para
este serviço.
