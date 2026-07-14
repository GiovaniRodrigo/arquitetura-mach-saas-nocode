# Deploy por artefatos — layout do host e pré-requisitos (spec 002)

Este documento descreve o alvo da entrega contínua: um host Linux com `systemd` e
Nginx que recebe **apenas artefatos compilados** (binários Go, release OTP do
`collab`, bundle estático do `player`) via `rsync`/SSH. O host nunca clona o
repositório nem possui toolchain de build (RNF01, RN01).

## 1. Layout do sistema de arquivos

```
/opt/machv4/
├── releases/
│   ├── <sha>/                # um diretório por release, imutável (RN04)
│   │   ├── bin/{gateway,iam,design,logic,deploy,export,workers}
│   │   ├── collab/           # release OTP autocontido (bin/, lib/, releases/, erts-*)
│   │   └── player/           # dist estático (docroot do Nginx)
│   └── <sha-anterior>/       # retido para rollback (RELEASES_KEEP, padrão 5)
└── current -> releases/<sha> # symlink trocado atomicamente na ativação (RN04, RN07)

/etc/machv4/                  # EnvironmentFiles por serviço (segredos; fora do artefato)
├── gateway.env  iam.env  design.env  logic.env  deploy.env  export.env
├── workers.env
└── collab.env
```

## 2. Pré-requisitos do host (provisionamento — fora do escopo desta demanda)

- Usuário de serviço non-root `machv4` (dono de `/opt/machv4`).
- Usuário de deploy SSH (ex.: `deploy`) com permissão de escrita em `/opt/machv4`
  e autorização para `systemctl restart 'machv4-*'` via `sudo` sem senha
  (regra `sudoers` restrita apenas a esse comando).
- `systemd`, `rsync`, `tar` e Nginx instalados.
- Unidades copiadas de `infra/systemd/*.service` para `/etc/systemd/system/` e
  habilitadas (`systemctl enable machv4-gateway ... machv4-collab`).
- Nginx apontando para `infra/nginx/machv4.conf` (symlink em `sites-enabled/`).
- OTel Collector alcançável (endpoint em cada `*.env`).

## 3. EnvironmentFiles (`/etc/machv4/<serviço>.env`)

Cada arquivo carrega os segredos/endpoints em runtime — **nunca** versionados nem
incluídos no artefato. Exemplos mínimos:

```ini
# /etc/machv4/gateway.env
GATEWAY_ADDR=:8080
IAM_GRPC_ADDR=127.0.0.1:50051
DESIGN_GRPC_ADDR=127.0.0.1:50052
LOGIC_GRPC_ADDR=127.0.0.1:50053
OTEL_EXPORTER_OTLP_ENDPOINT=otel-collector.internal:4317

# /etc/machv4/collab.env
SECRET_KEY_BASE=<gerar com: mix phx.gen.secret>
PHX_HOST=app.exemplo.com
PORT=4000
REDIS_URL=redis://127.0.0.1:6379
DESIGN_GRPC_ADDR=127.0.0.1:50052
OTEL_EXPORTER_OTLP_ENDPOINT=otel-collector.internal:4317
```

## 4. GitHub Environments e segredos (task 13)

Configurar dois *environments* no repositório (**Settings → Environments**):

| Environment | Acionado por | Proteção |
|-------------|--------------|----------|
| `staging` | push em `main` | — (deploy automático) |
| `production` | tag `vX.Y.Z` | **Required reviewers** (aprovação manual — RN03) |

Secrets por environment (mesmos nomes; valores distintos):

| Secret | Descrição |
|--------|-----------|
| `SSH_PRIVATE_KEY` | Chave privada dedicada ao ambiente (par cadastrado no `authorized_keys` do usuário de deploy) |
| `SSH_HOST` | Host/IP alvo |
| `SSH_USER` | Usuário de deploy (ex.: `deploy`) |
| `SSH_KNOWN_HOSTS` | Saída de `ssh-keyscan <host>` (evita TOFU no runner) |

Princípios: chave por ambiente, escopo mínimo, `production` sempre atrás de
aprovação. O runner é o único ponto com toolchain e segredos de build; o host só
recebe os tarballs.

## 5. Operação manual

```bash
# Deploy de um sha específico (normalmente feito pelo cd.yml)
scripts/deploy.sh --env staging --host "$SSH_HOST" --user deploy --sha <sha>

# Rollback para o release anterior
scripts/rollback.sh --env production --host "$SSH_HOST" --user deploy
```
