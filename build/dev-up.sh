#!/usr/bin/env bash
# Startup guiado do stack MACH V4 para desenvolvimento local.
# Ordem: pré-checagens -> infra (Docker Compose) -> proto -> cluster
# Kubernetes (kind + Linkerd + metrics-server) -> build de imagens -> deploy
# dos 8 serviços no cluster -> frontend.
#
# Os 8 serviços da plataforma (IAM, Design, Logic, Deploy, Export, Workers,
# Collab, Gateway) rodam dentro de um cluster kind local, com sidecar
# Linkerd injetado — é o que alimenta a tela Monitor de Recursos
# (specs/008-monitor-recursos, specs/009 nota de arquitetura) com CPU/memória
# (metrics-server) e RPS/taxa de sucesso/latência (Prometheus do linkerd-viz)
# reais, sem instrumentar cada serviço. Não rodam mais como processo solto
# (`go run`) — só o Frontend (Vite) continua fora do cluster.
#
# Uso:
#   ./build/dev-up.sh                # sobe tudo, com prompts de confirmação
#   ./build/dev-up.sh --no-frontend  # sobe tudo menos o frontend
#   ./build/dev-up.sh --yes          # não pergunta nada, assume "sim" em todos os prompts
#
# Logs de cada processo em background: .dev-logs/<nome>.log
# Ctrl+C encerra os processos que este script iniciou em foreground/background
# (frontend). O cluster kind e os pods continuam no ar entre execuções — não
# é derrubado por Ctrl+C nem ao final do script (reaproveitado na próxima
# vez; `kind delete cluster --name machv4` derruba de vez).

set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."

# Carrega overrides locais (ex.: MINIO_HOST_PORT quando outro projeto já
# ocupa 9000/9001) do mesmo .env que o Docker Compose usa, para as duas
# ferramentas concordarem na mesma porta.
if [ -f .env ]; then
  set -a; source .env; set +a
fi
MINIO_HOST_PORT="${MINIO_HOST_PORT:-9000}"
MINIO_CONSOLE_HOST_PORT="${MINIO_CONSOLE_HOST_PORT:-9001}"

WITH_FRONTEND=1
ASSUME_YES=0
for arg in "$@"; do
  case "$arg" in
    --no-frontend) WITH_FRONTEND=0 ;;
    --yes|-y) ASSUME_YES=1 ;;
    *) echo "Argumento desconhecido: $arg" >&2; exit 1 ;;
  esac
done

# --- Cores / feedback -------------------------------------------------------
if [ -t 1 ]; then
  C_RESET=$'\033[0m'; C_BOLD=$'\033[1m'
  C_GREEN=$'\033[32m'; C_RED=$'\033[31m'; C_YELLOW=$'\033[33m'; C_BLUE=$'\033[34m'; C_DIM=$'\033[2m'
else
  C_RESET=""; C_BOLD=""; C_GREEN=""; C_RED=""; C_YELLOW=""; C_BLUE=""; C_DIM=""
fi

step()  { echo; echo "${C_BOLD}${C_BLUE}==> $*${C_RESET}"; }
ok()    { echo "  ${C_GREEN}✓${C_RESET} $*"; }
warn()  { echo "  ${C_YELLOW}!${C_RESET} $*"; }
fail()  { echo "  ${C_RED}✗${C_RESET} $*"; }
info()  { echo "  ${C_DIM}$*${C_RESET}"; }

confirm() {
  # confirm "pergunta" -> 0 (sim) / 1 (não)
  local prompt="$1"
  if [ "$ASSUME_YES" = "1" ]; then
    ok "$prompt -> assumindo 'sim' (--yes)"
    return 0
  fi
  if [ ! -r /dev/tty ]; then
    warn "$prompt -> sem tty interativo, assumindo 'não' (rode com --yes ou responda manualmente)"
    return 1
  fi
  local reply
  # Prompt escrito direto em /dev/tty (não via `read -p`, que sai por stderr):
  # em alguns terminais integrados essa mistura de canais some sem aviso,
  # deixando o script parecendo travado sem nenhum feedback na tela.
  printf "  %s?%s %s [s/N] " "$C_YELLOW" "$C_RESET" "$prompt" >/dev/tty
  read -r reply </dev/tty
  case "$reply" in
    s|S|sim|y|Y|yes) return 0 ;;
    *) return 1 ;;
  esac
}

abort() { fail "$*"; echo; echo "Startup abortado."; exit 1; }

# --- Estado ------------------------------------------------------------------
LOG_DIR=".dev-logs"
mkdir -p "$LOG_DIR"
PIDS=()
NAMES=()

CLEANING_UP=0
cleanup() {
  # Guarda de reentrância: o EXIT trap dispara de novo quando chamamos
  # `exit` dentro do handler de INT/TERM abaixo — sem isso, a limpeza roda
  # em dobro.
  [ "$CLEANING_UP" = "1" ] && return
  CLEANING_UP=1
  if [ "${#PIDS[@]}" -gt 0 ]; then
    echo
    step "Encerrando processos iniciados por este script"
    for i in "${!PIDS[@]}"; do
      local_pid="${PIDS[$i]}"
      # -$pid mira o process group inteiro (criado via setsid em run_bg): matar
      # só o "go run" deixa o binário compilado filho órfão segurando a porta.
      kill -TERM -- "-$local_pid" 2>/dev/null && ok "${NAMES[$i]} (pid $local_pid)" || true
    done
    sleep 1
    for i in "${!PIDS[@]}"; do
      kill -0 -- "-${PIDS[$i]}" 2>/dev/null && kill -KILL -- "-${PIDS[$i]}" 2>/dev/null
    done
    wait 2>/dev/null || true
  fi
  info "cluster kind 'machv4' continua no ar (kind delete cluster --name machv4 para derrubar)"
}
trap cleanup EXIT
# Ctrl+C (INT) num builtin bloqueante como `read` faz o bash rodar o trap e
# depois RETENTAR o builtin interrompido — sem o `exit` explícito aqui, o
# script nunca sai e o prompt de confirmação do frontend fica retomando pra
# sempre a cada Ctrl+C, sem devolver o terminal.
trap 'cleanup; exit 130' INT TERM

run_bg() {
  local name="$1"; shift
  setsid "$@" >"$LOG_DIR/$name.log" 2>&1 &
  local pid=$!
  PIDS+=("$pid")
  NAMES+=("$name")
  info "log: $LOG_DIR/$name.log (pid $pid)"
}

port_in_use() {
  local port="$1"
  (exec 3<>"/dev/tcp/127.0.0.1/$port") 2>/dev/null && { exec 3<&-; exec 3>&-; return 0; }
  return 1
}

wait_for_port() {
  local host="$1" port="$2" label="$3" tries=60
  printf "  %s aguardando %s:%s..." "$label" "$host" "$port"
  until (exec 3<>"/dev/tcp/$host/$port") 2>/dev/null; do
    exec 3<&- 2>/dev/null; exec 3>&- 2>/dev/null
    tries=$((tries - 1))
    if [ "$tries" -le 0 ]; then
      echo
      fail "timeout esperando $label em $host:$port — veja $LOG_DIR/$label.log"
      return 1
    fi
    sleep 1
    printf "."
  done
  exec 3<&- 2>/dev/null; exec 3>&- 2>/dev/null
  echo " ${C_GREEN}ok${C_RESET}"
}

# =============================================================================
step "0/6  Pré-checagens de ferramentas"

export PATH="$HOME/.local/go/bin:$HOME/.local/elixir1.17/bin:$HOME/.mix/escripts:$HOME/.local/bin:$HOME/.linkerd2/bin:$PATH"
export MIX_HOME="${MIX_HOME:-$HOME/.mix}"
export HEX_HOME="${HEX_HOME:-$HOME/.hex}"
if command -v go >/dev/null 2>&1; then
  export PATH="$(go env GOPATH)/bin:$PATH"
fi

MISSING=0
check_tool() {
  local bin="$1" hint="$2"
  if command -v "$bin" >/dev/null 2>&1; then
    ok "$bin  ${C_DIM}($(command -v "$bin"))${C_RESET}"
  else
    fail "$bin não encontrado — $hint"
    MISSING=1
  fi
}
check_tool docker  "instale o Docker / Docker Compose"
check_tool go      "export PATH=\$HOME/.local/go/bin:\$PATH (Go 1.26 local, o do apt é 1.22)"
check_tool node    "instale Node 20"
check_tool npm     "instale Node 20 (inclui npm)"
check_tool mix     "export PATH=\$HOME/.local/elixir1.17/bin:\$PATH (Elixir 1.17, o do apt é 1.14)"
check_tool buf     "make tools  (instala em \$(go env GOPATH)/bin)"

if [ "$MISSING" = "1" ]; then
  abort "ferramentas faltando — resolva os itens marcados com ✗ acima e rode de novo."
fi

docker info >/dev/null 2>&1 || abort "Docker daemon não está acessível — inicie o Docker e rode de novo."

step "Verificando versão do Go"
GO_VER="$(go version | grep -oE 'go[0-9]+\.[0-9]+' | tr -d 'go')"
GO_MAJOR="${GO_VER%%.*}"; GO_MINOR="${GO_VER##*.}"
if [ "$GO_MAJOR" -lt 1 ] || { [ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -lt 23 ]; }; then
  warn "Go $GO_VER detectado no PATH — o repo precisa de 1.23+. Confirme \$HOME/.local/go/bin está antes no PATH."
  confirm "Continuar mesmo assim?" || abort "corrija o PATH do Go e rode de novo."
else
  ok "Go $GO_VER"
fi

# kind/kubectl/linkerd não são pré-requisitos "traga você mesmo" como
# Go/Node/Elixir — instala automaticamente em $HOME/.local/bin (kind,
# kubectl) e $HOME/.linkerd2/bin (linkerd CLI) se ausentes, mesma ideia do
# `make tools` para o buf. Hardcoded para linux/amd64 (ambiente documentado
# do projeto, igual aos caminhos de Go/Elixir acima).
ensure_tool_kind() {
  command -v kind >/dev/null 2>&1 && { ok "kind  ${C_DIM}($(command -v kind))${C_RESET}"; return; }
  info "instalando kind em \$HOME/.local/bin"
  mkdir -p "$HOME/.local/bin"
  curl -sLo "$HOME/.local/bin/kind" "https://kind.sigs.k8s.io/dl/v0.24.0/kind-linux-amd64" \
    && chmod +x "$HOME/.local/bin/kind" || abort "download do kind falhou"
  ok "kind instalado"
}
ensure_tool_kubectl() {
  command -v kubectl >/dev/null 2>&1 && { ok "kubectl  ${C_DIM}($(command -v kubectl))${C_RESET}"; return; }
  info "instalando kubectl em \$HOME/.local/bin"
  mkdir -p "$HOME/.local/bin"
  local ver; ver="$(curl -sL https://dl.k8s.io/release/stable.txt)"
  curl -sLo "$HOME/.local/bin/kubectl" "https://dl.k8s.io/release/$ver/bin/linux/amd64/kubectl" \
    && chmod +x "$HOME/.local/bin/kubectl" || abort "download do kubectl falhou"
  ok "kubectl instalado"
}
ensure_tool_linkerd() {
  command -v linkerd >/dev/null 2>&1 && { ok "linkerd  ${C_DIM}($(command -v linkerd))${C_RESET}"; return; }
  info "instalando linkerd CLI em \$HOME/.linkerd2/bin"
  curl -sL https://run.linkerd.io/install | sh >/dev/null 2>&1 || abort "instalação do linkerd CLI falhou"
  ok "linkerd instalado"
}
ensure_tool_kind
ensure_tool_kubectl
ensure_tool_linkerd

# =============================================================================
step "1/6  Infraestrutura (Docker Compose)"
info "Traz Jaeger/OTel Collector para tracing; Postgres/Redis/RabbitMQ/MinIO"
info "aqui não são usados pelos 8 serviços (que rodam no cluster k8s com sua"
info "própria cópia, passo 5/6) — mantidos por compatibilidade com quem ainda"
info "roda algum serviço solto (go run) manualmente."

for p in 5432 6379 5672 15672 4317 4318 "$MINIO_HOST_PORT" "$MINIO_CONSOLE_HOST_PORT" 16686; do
  if port_in_use "$p"; then
    warn "porta $p já está em uso no host — pode ser outro projeto (ex.: outro MinIO). Ajuste MINIO_HOST_PORT/MINIO_CONSOLE_HOST_PORT em .env se for o caso"
    confirm "Continuar mesmo assim (o docker compose pode falhar ao subir esse serviço)?" || \
      abort "libere a porta $p ou ajuste docker-compose.yml e rode de novo."
  fi
done

info "make up"
if ! make up; then
  fail "docker compose up falhou"
  confirm "Tentar continuar assim mesmo?" || abort "resolva o erro acima do compose."
fi

wait_for_port localhost 4317 otel-collector || abort "otel-collector não subiu"
ok "Infra no ar (postgres, redis, rabbitmq, jaeger, otel-collector, minio)"

# =============================================================================
step "2/6  Contratos proto (buf generate)"
if make proto; then
  ok "gen/go, gen/elixir, gen/ts regenerados"
else
  abort "buf lint/generate falhou — veja o erro acima."
fi

# =============================================================================
step "3/6  Cluster Kubernetes (kind + Linkerd + metrics-server)"

for p in 8080 4000; do
  if port_in_use "$p"; then
    warn "porta $p já em uso — precisa estar livre para o Gateway (8080) / Collab (4000) do cluster"
    warn "se for um dev-up.sh antigo (processo solto), pare-o antes de continuar"
    confirm "Continuar mesmo assim (o kind pode falhar ao mapear essa porta)?" || \
      abort "libere a porta $p e rode de novo."
  fi
done

if kind get clusters 2>/dev/null | grep -qx machv4; then
  ok "cluster kind 'machv4' já existe — reaproveitando"
else
  info "kind create cluster (baixa a imagem do node na 1ª vez, pode demorar)"
  kind create cluster --config infra/k8s/kind-config.yaml || abort "kind create cluster falhou"
fi
kubectl config use-context kind-machv4 >/dev/null || abort "kubectl config use-context kind-machv4 falhou"
kubectl wait --for=condition=Ready node --all --timeout=120s >/dev/null || abort "node do cluster não ficou Ready"

info "Gateway API CRDs (pré-requisito do Linkerd)"
kubectl apply --server-side -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.5.1/standard-install.yaml >/dev/null 2>&1

if kubectl get ns linkerd >/dev/null 2>&1; then
  ok "Linkerd control plane já instalado"
else
  info "linkerd install --crds"
  linkerd install --crds >"$LOG_DIR/linkerd-crds.yaml" 2>"$LOG_DIR/linkerd-crds.err" \
    || abort "linkerd install --crds falhou — veja $LOG_DIR/linkerd-crds.err"
  kubectl apply -f "$LOG_DIR/linkerd-crds.yaml" >/dev/null || abort "kubectl apply (linkerd CRDs) falhou"
  info "linkerd install"
  linkerd install >"$LOG_DIR/linkerd-install.yaml" 2>"$LOG_DIR/linkerd-install.err" \
    || abort "linkerd install falhou — veja $LOG_DIR/linkerd-install.err"
  kubectl apply -f "$LOG_DIR/linkerd-install.yaml" >/dev/null || abort "kubectl apply (linkerd control plane) falhou"
  kubectl -n linkerd rollout status deploy --timeout=180s >/dev/null || abort "Linkerd control plane não ficou pronto"
fi

if kubectl get ns linkerd-viz >/dev/null 2>&1; then
  ok "linkerd-viz (Prometheus/dashboard) já instalado"
else
  info "linkerd viz install"
  linkerd viz install >"$LOG_DIR/linkerd-viz.yaml" 2>"$LOG_DIR/linkerd-viz.err" \
    || abort "linkerd viz install falhou — veja $LOG_DIR/linkerd-viz.err"
  kubectl apply -f "$LOG_DIR/linkerd-viz.yaml" >/dev/null || abort "kubectl apply (linkerd-viz) falhou"
  kubectl -n linkerd-viz rollout status deploy --timeout=180s >/dev/null || abort "linkerd-viz não ficou pronto"
fi

if kubectl -n kube-system get deploy metrics-server >/dev/null 2>&1; then
  ok "metrics-server já instalado"
else
  info "metrics-server (CPU/memória por pod — kubectl top)"
  kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml >/dev/null \
    || abort "kubectl apply (metrics-server) falhou"
  # kind usa certificados internos que o metrics-server padrão não reconhece.
  kubectl -n kube-system patch deployment metrics-server --type=json \
    -p '[{"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--kubelet-insecure-tls"}]' >/dev/null
  kubectl -n kube-system rollout status deploy/metrics-server --timeout=120s >/dev/null || abort "metrics-server não ficou pronto"
fi
ok "Cluster k8s + service mesh prontos (linkerd viz check para diagnóstico)"

# =============================================================================
step "4/6  Build de artefatos + imagens Docker"

SHA="$(git rev-parse --short HEAD 2>/dev/null || echo dev)"
mkdir -p dist/release/bin

declare -A GO_UNITS=(
  [iam]=./services/iam/cmd
  [design]=./services/design/cmd
  [logic]=./services/logic/cmd
  [deploy]=./services/deploy/cmd
  [export]=./services/export/cmd
  [workers]=./services/workers/cmd
  [gateway]=./services/gateway/cmd
)
for name in "${!GO_UNITS[@]}"; do
  info "go build $name"
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags "-s -w -X main.version=$SHA" -o "dist/release/bin/$name" "${GO_UNITS[$name]}" \
    || abort "go build $name falhou"
done

info "mix release collab"
(
  cd services/collab
  export MIX_ENV=prod
  mix deps.get --only prod >/dev/null 2>&1 && mix release collab --overwrite >/dev/null
) || abort "mix release collab falhou — veja o erro acima"
rm -rf dist/release/collab
cp -a services/collab/_build/prod/rel/collab dist/release/collab
find dist/release/collab -type d -path '*/priv/templates' -prune -exec rm -rf {} +

info "docker build (8 imagens machv4/<serviço>:dev)"
for name in iam design logic deploy export workers gateway; do
  docker build -q -f infra/docker/go-service.Dockerfile --build-arg BINARY="$name" -t "machv4/$name:dev" . >/dev/null \
    || abort "docker build $name falhou"
done
docker build -q -f infra/docker/collab.Dockerfile -t machv4/collab:dev . >/dev/null || abort "docker build collab falhou"

info "kind load docker-image (carregando as 8 imagens no cluster)"
kind load docker-image machv4/iam:dev machv4/design:dev machv4/logic:dev machv4/deploy:dev \
  machv4/export:dev machv4/workers:dev machv4/gateway:dev machv4/collab:dev --name machv4 >/dev/null \
  || abort "kind load docker-image falhou"
ok "8 imagens buildadas e carregadas no cluster"

# O kind tem seu próprio cache de imagens, separado do Docker do host: sem
# isso, postgres/redis/rabbitmq/minio seriam baixados de novo dentro do
# cluster mesmo o Docker Compose (passo 1/6) já tendo acabado de puxar
# exatamente essas imagens — pull duplicado e lento o bastante para estourar
# o timeout do rollout logo abaixo.
info "kind load docker-image (reaproveitando postgres/redis/rabbitmq/minio já baixados pelo compose)"
kind load docker-image postgres:16 redis:7 rabbitmq:3.13-management minio/minio:latest --name machv4 >/dev/null \
  || warn "kind load docker-image (infra) falhou — vão ser baixados de novo dentro do cluster, mais devagar"

# =============================================================================
step "5/6  Deploy no Kubernetes"

kubectl apply -f infra/k8s/00-namespace.yaml >/dev/null

# Segredos gerados sob demanda (nunca gravados em arquivo versionado) — só na
# 1ª vez, para não invalidar sessões/MFA a cada re-execução deste script.
ensure_secret() {
  local name="$1"; shift
  if kubectl -n machv4 get secret "$name" >/dev/null 2>&1; then
    info "secret $name já existe — mantendo"
  else
    kubectl -n machv4 create secret generic "$name" "$@" >/dev/null || abort "criar secret $name falhou"
    ok "secret $name criado"
  fi
}
ensure_secret iam-secrets --from-literal="IAM_MFA_ENCRYPTION_KEY=$(openssl rand -base64 32)"
ensure_secret collab-secrets --from-literal="SECRET_KEY_BASE=$(openssl rand -base64 48)"

# Par de chaves RS256 do IAM, compartilhado com o Collab (que, ao contrário
# do Gateway, verifica o JWT localmente em vez de delegar ao IAM via gRPC).
# Sem isto, o IAM cai no fallback de dev (gera um par efêmero a cada boot,
# nunca exposto a mais ninguém) e o Collab nunca consegue autenticar nenhuma
# ligação WebSocket (lib/collab/auth/token.ex) — gerado só na 1ª vez, pelo
# mesmo motivo dos outros secrets acima.
if kubectl -n machv4 get secret jwt-keys >/dev/null 2>&1; then
  info "secret jwt-keys já existe — mantendo"
else
  jwt_tmp="$(mktemp -d)"
  openssl genrsa -out "$jwt_tmp/private.pem" 2048 >/dev/null 2>&1
  openssl rsa -in "$jwt_tmp/private.pem" -pubout -out "$jwt_tmp/public.pem" >/dev/null 2>&1
  kubectl -n machv4 create secret generic jwt-keys \
    --from-file=private.pem="$jwt_tmp/private.pem" \
    --from-file=public.pem="$jwt_tmp/public.pem" >/dev/null || abort "criar secret jwt-keys falhou"
  rm -rf "$jwt_tmp"
  ok "secret jwt-keys criado"
fi

kubectl apply \
  -f infra/k8s/02-migrations-configmap.yaml \
  -f infra/k8s/03a-rabbitmq-definitions.yaml \
  -f infra/k8s/03b-rabbitmq-conf.yaml \
  -f infra/k8s/03-infra.yaml >/dev/null || abort "kubectl apply (infra k8s) falhou"

kubectl -n machv4 rollout status deploy/postgres --timeout=240s >/dev/null || abort "postgres (k8s) não subiu"
kubectl -n machv4 rollout status deploy/redis --timeout=240s >/dev/null || abort "redis (k8s) não subiu"
kubectl -n machv4 rollout status deploy/rabbitmq --timeout=240s >/dev/null || abort "rabbitmq (k8s) não subiu"
kubectl -n machv4 rollout status deploy/minio --timeout=240s >/dev/null || abort "minio (k8s) não subiu"

info "aplicando migrações (Job, idempotente — apagado e recriado a cada execução)"
kubectl -n machv4 delete job migrate --ignore-not-found >/dev/null
kubectl apply -f infra/k8s/03-infra.yaml >/dev/null # reaplica só o Job (o resto já está sem mudanças)
kubectl -n machv4 wait --for=condition=complete job/migrate --timeout=90s >/dev/null \
  || abort "migrações falharam — veja: kubectl -n machv4 logs job/migrate"
ok "migrações aplicadas"

kubectl apply -f infra/k8s/05-gateway-rbac.yaml -f infra/k8s/04-services.yaml >/dev/null \
  || abort "kubectl apply (serviços da plataforma) falhou"

# imagePullPolicy IfNotPresent não detecta sozinho que a tag ":dev" mudou de
# conteúdo — reinicia os 8 deployments para pegar a imagem recém-carregada.
for svc in iam design logic deploy export workers collab gateway; do
  kubectl -n machv4 rollout restart deploy/"$svc" >/dev/null
done
for svc in iam design logic deploy export workers collab gateway; do
  kubectl -n machv4 rollout status deploy/"$svc" --timeout=120s >/dev/null || abort "$svc (k8s) não subiu"
done
ok "8 serviços da plataforma no ar, com sidecar Linkerd"

wait_for_port localhost 8080 gateway || abort "gateway (NodePort 8080) não respondeu"
wait_for_port localhost 4000 collab  || abort "collab (NodePort 4000) não respondeu"

# =============================================================================
if [ "$WITH_FRONTEND" = "1" ]; then
  step "6/6  Frontend (Vite)"
  if [ ! -d services/frontend/node_modules ]; then
    info "services/frontend/node_modules ausente — rodando npm install"
    (cd services/frontend && npm install) || abort "npm install falhou"
  fi
  ok "dependências do frontend prontas"
fi

# =============================================================================
step "Stack no ar"
cat <<EOF
  ${C_BOLD}Gateway${C_RESET}   http://localhost:8080  ${C_DIM}(cluster kind 'machv4', NodePort)${C_RESET}
  ${C_BOLD}Collab${C_RESET}    http://localhost:4000  ${C_DIM}(cluster kind 'machv4', NodePort)${C_RESET}
  ${C_BOLD}Jaeger${C_RESET}    http://localhost:16686
  ${C_BOLD}RabbitMQ${C_RESET}  http://localhost:15672  (mach/mach)
  ${C_BOLD}MinIO${C_RESET}     http://localhost:$MINIO_CONSOLE_HOST_PORT   (mach/machsecret)

  Monitor de Recursos: /dashboard/monitor no Frontend (CPU/memória via
  metrics-server, RPS/sucesso/latência via Prometheus do linkerd-viz).

  kubectl -n machv4 get pods         # status dos 8 serviços
  linkerd -n machv4 viz stat deploy  # métricas do service mesh
  Logs dos passos deste script: ${C_DIM}$LOG_DIR/*.log${C_RESET}
EOF

if [ "$WITH_FRONTEND" = "1" ]; then
  echo "  ${C_BOLD}Frontend${C_RESET}  http://localhost:5183  (iniciando em foreground abaixo)"
  echo
  confirm "Abrir o frontend agora (npm run dev, foreground, Ctrl+C encerra)?" && {
    cd services/frontend && exec npm run dev
  }
  info "Frontend não iniciado. Rode manualmente: cd services/frontend && npm run dev"
fi

echo
echo "Ctrl+C encerra o frontend (o cluster kind continua no ar)."
wait
