# Interfaces: Pipeline CI/CD por Entrega de Artefatos Compilados

Esta demanda não expõe API HTTP própria; os contratos são o **formato dos artefatos**, o **layout do host**, a **interface dos scripts** e os **segredos** consumidos pelo pipeline.

---

## 1. Contrato do artefato (`dist/artifacts/<unidade>-<sha>.tar.gz`)

Cada tarball contém **apenas** conteúdo executável. É proibido incluir fonte, testes, `.git`, `node_modules`, `deps` ou stubs versionados (RN01, RN05).

| Unidade | Comando de build | Conteúdo do tarball |
|---------|------------------|---------------------|
| `gateway`, `iam`, `design`, `logic`, `deploy`, `export`, `workers` | `CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=<sha>" -o bin/<unidade> ./<caminho>/cmd` | `bin/<unidade>` (binário estático único) |
| `collab` | `MIX_ENV=prod mix release collab` | Árvore do release OTP (`bin/`, `lib/`, `releases/`, ERTS) |
| `player` | `npm ci && npm run build` | Conteúdo de `dist/` (HTML/JS/CSS minificados) |

**Nomenclatura**: `<unidade>-<sha>.tar.gz`, onde `<sha>` é o git sha curto (imutável, RN04).

---

## 2. Contrato do layout do host

```
/opt/machv4/
├── releases/
│   ├── <sha-atual>/
│   │   ├── bin/{gateway,iam,design,logic,deploy,export,workers}
│   │   ├── collab/           # release OTP autocontido
│   │   └── player/           # bundle estático (docroot do Nginx)
│   └── <sha-anterior>/       # retido para rollback (RELEASES_KEEP)
└── current -> releases/<sha-atual>   # symlink trocado atomicamente (RN04, RN07)
```

- A ativação é um `rename` atômico do symlink `current`; nunca aponta a um diretório parcialmente transferido (Critério 4).
- `releases/` retém as `RELEASES_KEEP` versões mais recentes.

---

## 3. Contrato das unidades `systemd`

Nome: `machv4-<unidade>.service` para `gateway, iam, design, logic, deploy, export, workers, collab`.

| Campo | Valor |
|-------|-------|
| `ExecStart` | `/opt/machv4/current/bin/<unidade>` (ou `/opt/machv4/current/collab/bin/collab start`) |
| `User` | usuário de serviço non-root (RNF01) |
| `EnvironmentFile` | `/etc/machv4/<unidade>.env` (DSN, OTLP, portas — nunca no artefato) |
| `Restart` | `on-failure` |

---

## 4. Interface dos scripts

```bash
# Compila e empacota todos os artefatos a partir de dist/ (nunca da raiz do repo)
build/build-artifacts.sh
#   entrada:  SHA (env, default: git rev-parse --short HEAD)
#   saída:    dist/artifacts/<unidade>-<SHA>.tar.gz ; exit 0 = sucesso

# Entrega os artefatos a um ambiente e ativa o release
build/deploy.sh --env <staging|production> --host <host> --user <user> --sha <sha>
#   efeito:   rsync -> releases/<sha> ; ln -sfn atômico ; systemctl restart 'machv4-*'

# Verifica a saúde dos serviços após a ativação
build/smoke-test.sh --host <host>
#   contrato: exit 0 = todos saudáveis ; exit ≠ 0 = falha (dispara rollback, RN08)

# Reverte para o release anterior (ou um sha informado), sem recompilar
build/rollback.sh --env <staging|production> --host <host> [--sha <sha>]
```

---

## 5. Contrato dos segredos e ambientes (GitHub)

| Environment | Secret | Uso |
|-------------|--------|-----|
| `staging`, `production` | `SSH_PRIVATE_KEY` | Chave dedicada por ambiente para o rsync/SSH (RNF01) |
| `staging`, `production` | `SSH_HOST` | Host alvo do deploy |
| `staging`, `production` | `SSH_USER` | Usuário de serviço non-root |
| `staging`, `production` | `SSH_KNOWN_HOSTS` | Fingerprint do host (evita TOFU no runner) |

- `staging`: acionado automaticamente por push em `main`.
- `production`: acionado por **disparo manual** do `cd.yml` (`workflow_dispatch`,
  normalmente com `--ref vX.Y.Z`) — o disparo é o gate humano (RN02, RN03).
  *Required reviewers* de environment exigem plano pago/repo público; o disparo
  manual cumpre o mesmo papel no plano free.

---

## 6. Contrato do smoke test (healthchecks)

| Serviço | Verificação |
|---------|-------------|
| `gateway` | `GET http://<host>:8080/healthz` → `200` |
| `iam`, `design`, `logic`, `deploy`, `export` | *health check* gRPC na porta do serviço (`50051`–`50055`) |
| `collab` | `GET http://<host>:4000/healthz` (Phoenix) → `200` |
| `player` | `GET http://<host>/` (Nginx serve `index.html`) → `200` |
| `workers` | unidade `systemd` ativa (`systemctl is-active machv4-workers`) |

Qualquer verificação com falha ⇒ `smoke-test.sh` retorna ≠ 0 ⇒ rollback automático (RN08).
