# Plano de Implementação: Pipeline CI/CD por Entrega de Artefatos Compilados

A estratégia reaproveita o pipeline de validação da Fase 11 (`.github/workflows/ci.yml`) como *gate* e adiciona um pipeline de release (`cd.yml`) que compila os artefatos no runner, empacota apenas o conteúdo executável e entrega ao host via rsync/SSH com ativação atômica por symlink. A produção passa a executar os serviços sob `systemd` (binários Go e release OTP do `collab`) e o player sob Nginx, a partir de diretórios de release versionados pelo git sha. Toda a lógica de compilação e entrega vive em scripts idempotentes (`build/`) chamados tanto pelo CI quanto localmente, garantindo paridade.

---

## 1. Arquivos a Criar/Editar

### 1.1. Pipeline (GitHub Actions)

* **`.github/workflows/ci.yml`** *(editar)*: manter os jobs de validação (Fase 11) e expô-los como *gate* reutilizável (`workflow_call`) para o pipeline de release consumir. [RF01, RF10]
* **`.github/workflows/cd.yml`** *(criar)*: dispara em push a `main` (→ staging) e em tags `v*` (→ produção); invoca o CI como *gate*, compila os artefatos, empacota, publica e entrega. Usa *environments* `staging` e `production` (este com proteção/approval). [RF02–RF08, RF11, RN02, RN03]

### 1.2. Compilação de artefatos

* **`build/build-artifacts.sh`** *(criar)*: gera `gen/` (`buf generate`), compila os 7 binários Go (`CGO_ENABLED=0 go build`), o release OTP (`mix release`) e o `player/dist` (`vite build`); empacota cada um em `dist/artifacts/<unidade>-<sha>.tar.gz` contendo só o executável. [RF02, RF03, RN01, RN05, RN06]
* **`collab/mix.exs`** *(editar)*: adicionar a configuração `releases:` (nome `collab`, `include_executables_for: [:unix]`) para `mix release` produzir um pacote OTP autocontido com ERTS. [RF02]
* **`collab/rel/env.sh.eex`** *(criar)*: variáveis de ambiente do release em runtime (porta, endpoint do OTel, Redis, DSN gRPC). [RF02]

### 1.3. Entrega (CD)

* **`build/deploy.sh`** *(criar)*: recebe host/usuário/ambiente e o diretório de artefatos; faz `rsync -a --delete` para `releases/<sha>`, extrai os tarballs, troca o symlink `current` atomicamente (`ln -sfn`), reinicia as unidades `systemd` e serve o player. Idempotente. [RF07, RF08, RN01, RN04]
* **`build/smoke-test.sh`** *(criar)*: consulta os healthchecks de cada serviço após a ativação; código de saída ≠ 0 sinaliza falha. [RF11]
* **`build/rollback.sh`** *(criar)*: repointa `current` ao release imediatamente anterior (ou a um sha informado) e reinicia os serviços, sem recompilar. [RF09, RN07, RN08]

### 1.4. Runtime de produção

* **`infra/systemd/machv4-gateway.service`** e demais (`iam`, `design`, `logic`, `deploy`, `export`, `workers`, `collab`) *(criar)*: units `systemd` apontando para `/opt/machv4/current/bin/<unidade>`, executando como usuário non-root, com `Restart=on-failure` e variáveis de ambiente por `EnvironmentFile`. [RF08, RNF01]
* **`infra/nginx/machv4.conf`** *(criar)*: serve o `player` estático a partir de `/opt/machv4/current/player` e faz proxy reverso para o `gateway` (`:8080`). [RF08]
* **`infra/deploy/README.md`** *(criar)*: layout do host (`/opt/machv4/{releases,current}`), usuário de serviço, pré-requisitos. [RNF01, RNF04]

### 1.5. Segredos e ambientes

* **GitHub Environments `staging` e `production`** *(configurar)*: secrets `SSH_PRIVATE_KEY`, `SSH_HOST`, `SSH_USER`, `SSH_KNOWN_HOSTS`. O gate de produção é o **disparo manual** do `cd.yml` (`workflow_dispatch`), pois *required reviewers* exigem plano pago/repo público. [RF06, RN03, RNF01]

---

## 2. Estratégia Técnica

### 2.1. Artefatos autocontidos, produção sem toolchain

Cada unidade vira um pacote que **executa sem o ambiente de build**:

```bash
# Go — binário estático, sem libc do host, sem toolchain em produção (RN06)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -trimpath -ldflags "-s -w -X main.version=$SHA" \
  -o dist/bin/gateway ./gateway/cmd

# Elixir — release OTP autocontido (traz ERTS + BEAM compilado, sem mix/deps)
MIX_ENV=prod mix release collab   # -> _build/prod/rel/collab

# Player — bundle estático minificado (sem node_modules)
npm ci && npm run build           # -> player/dist
```

O empacotamento inclui **apenas** o diretório de saída de cada build; o `build-artifacts.sh` monta os tarballs a partir de `dist/`, nunca da raiz do repo, garantindo a RN01.

### 2.2. Ativação atômica por symlink e rollback sem recompilação

O host mantém `/opt/machv4/releases/<sha>/` e um symlink `current`. A ativação troca o alias em uma operação atômica; o rollback é a mesma troca para outro sha:

```bash
# ativação (deploy.sh)
ln -sfn "/opt/machv4/releases/$SHA" /opt/machv4/current.tmp
mv -Tf /opt/machv4/current.tmp /opt/machv4/current     # rename atômico
systemctl restart 'machv4-*.service'

# rollback (rollback.sh) — aponta ao release anterior, sem build (RN07)
ln -sfn "/opt/machv4/releases/$PREV" /opt/machv4/current.tmp
mv -Tf /opt/machv4/current.tmp /opt/machv4/current
systemctl restart 'machv4-*.service'
```

`releases/` retém as N versões mais recentes (limpeza no fim do deploy), viabilizando rollback imediato.

### 2.3. Separação rígida CI → artefato → CD

`ci.yml` só **valida** (gate). `cd.yml` só entra na fase de build/entrega se o gate passou (`needs`/`workflow_call`), materializando o fluxo `[Git] → [Runner compila] → [Produção recebe artefatos]`. O runner é o único ponto com toolchain, dependências de dev e segredos de build; o host de produção só recebe tarballs por rsync e nunca clona o repositório.

### 2.4. Smoke test com rollback automático

Após reiniciar os serviços, o `smoke-test.sh` valida os healthchecks; qualquer falha aciona `rollback.sh` no mesmo job, restaurando o release anterior (RN08) antes de marcar o deploy como falho.

---

## 3. Dependências e Pré-requisitos

- [ ] Host(s) de staging e produção provisionados: usuário de serviço non-root, `systemd`, `rsync`, Nginx e o diretório `/opt/machv4` com permissão. (Fora de escopo — pré-requisito.)
- [ ] Estratégia de migração de banco definida: as migrações (`infra/postgres/migrations/`) devem ser aplicadas antes de ativar um release que dependa de novo schema. (Pré-requisito; não automatizado nesta demanda.)
- [ ] OTel Collector alcançável a partir do host (endpoint por ambiente). [RNF06]
- [ ] Segredos SSH configurados nos GitHub Environments `staging` e `production`. [RNF01]
- [ ] Toolchains no runner: Go 1.26, OTP 26.2/Elixir 1.17.3, Node 20, `buf` 1.42.0 (já usados no `ci.yml` da Fase 11).

---

## 4. Riscos e Pontos de Atenção

| Risco | Impacto | Mitigação |
|-------|---------|-----------|
| Modelo por artefato/systemd perde o scale-to-zero do KEDA para os `workers` (manifesto k8s existente). | Alto | Rodar `workers` como serviço `systemd` sempre-ativo com réplica fixa **ou** manter os `workers` no substrato container/KEDA e aplicar o modelo por artefato só aos demais. Decisão documentada em `research.md`. |
| Acoplamento entre deploy e migração de schema pode gerar release incompatível com o banco. | Alto | Adotar migrações *backward-compatible* (expand/contract); aplicar migração antes do deploy no runbook do `quickstart.md`. |
| Host único de produção é ponto único de falha (SPOF). | Médio | Documentar como limitação; o rsync/symlink é replicável para múltiplos hosts em iteração futura. |
| Vazamento de artefato de dev para produção por script mal configurado. | Alto | `build-artifacts.sh` monta tarballs só de `dist/`; teste de aceitação inspeciona o host (critério 1). |
| Chave SSH comprometida concede acesso ao host de produção. | Alto | Chave dedicada por ambiente, escopo mínimo, guardada em GitHub Environment protegido; `production` exige aprovação (RN03). |
