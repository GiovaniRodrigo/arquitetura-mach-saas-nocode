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

O script idempotente [`provision-host.sh`](./provision-host.sh) cria e configura
tudo o que segue. Rode-o como root **em cada host** (staging e produção), passando
a chave pública do CI correspondente ao ambiente:

```bash
sudo ./infra/deploy/provision-host.sh --pubkey ~/ci_machv4_staging.pub
```

Ele executa, de forma reexecutável:

- Usuário de serviço non-root `machv4` (dono de `/opt/machv4`).
- Usuário de deploy SSH (ex.: `deploy`) membro do grupo `machv4`, com escrita em
  `/opt/machv4` (dir `2775`/setgid) e autorização para `systemctl restart 'machv4-*'`
  via `sudo` sem senha (regra `sudoers` restrita — `/etc/sudoers.d/machv4-deploy`,
  validada com `visudo -cf`).
- Chave pública do CI autorizada em `~deploy/.ssh/authorized_keys` (via `--pubkey`).
- `/etc/machv4/` (0750) com um stub `<serviço>.env` por unidade (a preencher — §3).
- Unidades de `infra/systemd/*.service` instaladas em `/etc/systemd/system/` e
  habilitadas (sem `start` — só após o 1º deploy criar `current`).
- Nginx apontando para `infra/nginx/machv4.conf` (`sites-enabled/` ou `conf.d/`),
  com `nginx -t` + reload.

Flags: `--service-user`, `--deploy-user`, `--base`, `--env-dir`, `--no-systemd`,
`--no-nginx`, `--repo-dir`, `--help`. Pré-requisitos do SO que **não** são criados
pelo script: `systemd`, `rsync`, `tar`, Nginx instalados, e o OTel Collector
alcançável (endpoint em cada `*.env`).

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

| Environment | Acionado por | Gate |
|-------------|--------------|------|
| `staging` | push em `main` | — (deploy automático) |
| `production` | **disparo manual** (`workflow_dispatch`) | o próprio ato de disparar é a aprovação humana — RN03 |

> Nota: *required reviewers* de environment exigem repo público ou plano pago
> (Pro/Team/Enterprise). Neste repositório privado no plano free, o gate de
> produção é o **disparo manual** do `cd.yml`. Para promover uma tag:
>
> ```bash
> gh workflow run cd.yml --ref vX.Y.Z
> ```

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
build/deploy.sh --env staging --host "$SSH_HOST" --user deploy --sha <sha>

# Rollback para o release anterior
build/rollback.sh --env production --host "$SSH_HOST" --user deploy
```
