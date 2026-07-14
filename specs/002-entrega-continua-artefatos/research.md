# Pesquisa: Pipeline CI/CD por Entrega de Artefatos Compilados

---

## 1. Padrões Existentes no Projeto

| Arquivo/Padrão | Localização | Relevância |
|----------------|-------------|-----------|
| Pipeline de validação (Fase 11) | `.github/workflows/ci.yml` | Base reutilizável do *gate* de CI: jobs `proto`/`go`/`elixir`/`player`/`integration`. O `cd.yml` o consome via `workflow_call`. |
| Geração de stubs `.proto` | `Makefile` (`make proto`), `buf.gen.yaml` | O build de artefatos precisa gerar `gen/` antes de compilar Go/Elixir. `gen/` é gitignored (RN05). |
| Entrypoints Go | `gateway/cmd`, `services/{iam,design,logic,deploy,export}/cmd`, `workers/cmd` | As 7 unidades binárias a compilar (`go build ./<caminho>/cmd`). |
| Serviço Elixir | `collab/mix.exs` | Precisa ganhar config `releases:` para `mix release` produzir OTP autocontido. |
| Player | `player/package.json` (`build`), `player/vite.config.ts` | `vite build` → `player/dist` (bundle estático minificado). |
| Manifesto KEDA/k8s | `infra/k8s/keda/scaledobject-workers.yaml` | Substrato **alternativo** (container/Kubernetes) — referência para a decisão de arquitetura (seção 4). Usa `ghcr.io/machv4/workers:latest`, namespace `mach`. |
| Instrumentação OTel | serviços Go + `collab` (Fase 9) | Os binários já exportam OTLP; o runtime de produção só precisa apontar `OTEL_EXPORTER_OTLP_ENDPOINT` ao Collector do ambiente. |
| Toolchains do runner | memórias do projeto | Go 1.26 (`$HOME/.local/go`), OTP 26.2/Elixir 1.17.3, `buf` 1.42.0, `protoc-gen-elixir` via escript. |

---

## 2. Tecnologias e Bibliotecas

| Tecnologia | Versão | Uso | Já instalada? |
|------------|--------|-----|---------------|
| GitHub Actions | — | Orquestração de CI/CD, *environments* e aprovação manual | Sim (Fase 11) |
| `go build` (`CGO_ENABLED=0`) | Go 1.26 | Binários estáticos, sem toolchain em produção | Sim (runner) |
| `mix release` | Elixir 1.17.3 / OTP 26.2 | Release OTP autocontido do `collab` | Sim (runner) |
| Vite | 5.x | `vite build` → `dist/` estático | Sim (player) |
| rsync sobre SSH | — | Transferência incremental (delta) só de artefatos | Padrão do runner Ubuntu |
| `webfactory/ssh-agent` (ou `ssh-agent` nativo) | — | Injeção da chave SSH no job de deploy | Não (adicionar) |
| systemd | — | Supervisão dos serviços no host (restart, EnvironmentFile) | Assumido no host |
| Nginx | — | Servir o player estático + proxy para o gateway | Assumido no host |

---

## 3. Referências Externas

| Referência | URL | O que resolve |
|------------|-----|--------------|
| Deploying Elixir releases | https://hexdocs.pm/mix/Mix.Tasks.Release.html | Configuração de `mix release` autocontido |
| Go — build de binários estáticos | https://pkg.go.dev/cmd/go#hdr-Compile_packages_and_dependencies | Flags `-trimpath`, `-ldflags`, `CGO_ENABLED=0` |
| GitHub Environments / workflow_dispatch | https://docs.github.com/actions/deployment/targeting-different-environments | Escopo de secrets por ambiente; gate de produção por disparo manual (RN03). *Required reviewers* precisam de plano pago/repo público. |
| rsync deploy pattern | https://rsync.samba.org/documentation.html | Entrega incremental e `--delete` |
| Zero-downtime symlink swap | https://12factor.net (build/release/run) | Separação build → release → run; ativação atômica |

---

## 4. Alternativas Consideradas

### Opção A: Entrega por artefatos compilados (rsync/SSH + systemd) — **Escolhida**
- **Prós**: produção sem toolchain, fonte ou `.git` (RNF01); binários Go estáticos e release OTP são naturalmente autocontidos; ativação/rollback atômicos por symlink; simples, sem orquestrador. Alinha-se exatamente ao fluxo pedido `[Git] → [Runner] → [Produção só recebe artefatos]`.
- **Contras**: sem scale-to-zero nativo dos `workers`; host único tende a SPOF; migração de schema fica desacoplada do deploy.
- **Decisão**: **Escolhida** — atende ao requisito central de enviar somente artefatos.

### Opção B: Imagens de container + Kubernetes/GHCR + KEDA
- **Prós**: já há um manifesto KEDA (`scaledobject-workers.yaml`) e imagens `ghcr.io/machv4/*`; scale-to-zero dos workers; portabilidade.
- **Contras**: a imagem é o artefato, mas o modelo foge do requisito de "apenas arquivos compilados via rsync"; exige cluster, registry e credenciais; maior complexidade operacional.
- **Decisão**: **Descartada** para esta demanda; permanece como caminho válido especificamente para os `workers` (autoscaling por profundidade de fila). Ver risco correspondente em `plan.md`.

### Opção C: Clonar o repositório no host e buildar em produção (`git pull` + `go build`/`npm install`)
- **Prós**: trivial de configurar.
- **Contras**: viola RN01/RNF01 (fonte, `.git`, toolchain e dev-deps em produção); builds não reprodutíveis; superfície de ataque maior.
- **Decisão**: **Descartada** — é justamente o antipadrão que esta demanda elimina.

### Opção D: Publicar o player em object storage/CDN (S3/MinIO) em vez de Nginx no host
- **Prós**: descarrega o host do tráfego estático; cache/CDN; já há MinIO na stack (Fase 8).
- **Contras**: introduz um segundo alvo de deploy e configuração de CDN/domínio.
- **Decisão**: **Descartada** como padrão; registrada como evolução. O padrão adotado serve o `dist/` pelo Nginx do host.
