#!/usr/bin/env bash
# Startup guiado do stack MACH V4 para desenvolvimento local.
# Ordem: pré-checagens -> infra -> proto -> services Go -> workers -> gateway -> collab -> frontend.
#
# Uso:
#   ./build/dev-up.sh                # sobe tudo, com prompts de confirmação
#   ./build/dev-up.sh --no-frontend  # sobe tudo menos o frontend
#   ./build/dev-up.sh --yes          # não pergunta nada, assume "sim" em todos os prompts
#
# Logs de cada processo em background: .dev-logs/<nome>.log
# Ctrl+C encerra todos os processos que este script iniciou.

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
  local reply
  read -r -p "  ${C_YELLOW}?${C_RESET} $prompt [s/N] " reply </dev/tty
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

cleanup() {
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
}
trap cleanup EXIT INT TERM

run_bg() {
  local name="$1"; shift
  setsid "$@" >"$LOG_DIR/$name.log" 2>&1 &
  local pid=$!
  PIDS+=("$pid")
  NAMES+=("$name")
  info "log: $LOG_DIR/$name.log (pid $pid)"
}

# Roda um comando síncrono, mostra a saída na tela e também grava em
# $LOG_DIR/<name>.log, para que todos os logs (background ou não) fiquem
# concentrados na mesma pasta.
run_logged() {
  local name="$1"; shift
  info "log: $LOG_DIR/$name.log"
  "$@" > >(tee "$LOG_DIR/$name.log") 2> >(tee -a "$LOG_DIR/$name.log" >&2)
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
step "0/7  Pré-checagens de ferramentas"

export PATH="$HOME/.local/go/bin:$HOME/.local/elixir1.17/bin:$HOME/.mix/escripts:$PATH"
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

step "Verificando versão do Go"
GO_VER="$(go version | grep -oE 'go[0-9]+\.[0-9]+' | tr -d 'go')"
GO_MAJOR="${GO_VER%%.*}"; GO_MINOR="${GO_VER##*.}"
if [ "$GO_MAJOR" -lt 1 ] || { [ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -lt 23 ]; }; then
  warn "Go $GO_VER detectado no PATH — o repo precisa de 1.23+. Confirme \$HOME/.local/go/bin está antes no PATH."
  confirm "Continuar mesmo assim?" || abort "corrija o PATH do Go e rode de novo."
else
  ok "Go $GO_VER"
fi

# =============================================================================
step "1/7  Infraestrutura (Docker Compose)"

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
info "make migrate"
if ! make migrate; then
  fail "migrações falharam"
  confirm "Continuar mesmo assim (dados podem estar desatualizados)?" || abort "corrija as migrações."
fi

wait_for_port localhost 5432 postgres || abort "postgres não subiu"
wait_for_port localhost 5672 rabbitmq || abort "rabbitmq não subiu"
wait_for_port localhost "$MINIO_HOST_PORT" minio || abort "minio não subiu"
ok "Infra no ar (postgres, redis, rabbitmq, jaeger, otel-collector, minio)"

# =============================================================================
step "2/7  Contratos proto (buf generate)"
if make proto; then
  ok "gen/go, gen/elixir, gen/ts regenerados"
else
  abort "buf lint/generate falhou — veja o erro acima."
fi

# =============================================================================
step "3/7  Services gRPC (Go)"
run_bg iam    go run ./services/iam/cmd
run_bg design go run ./services/design/cmd
run_bg logic  go run ./services/logic/cmd
run_bg deploy go run ./services/deploy/cmd
run_bg export env "S3_ENDPOINT=localhost:$MINIO_HOST_PORT" go run ./services/export/cmd

wait_for_port localhost 50051 iam    || abort "iam não subiu"
wait_for_port localhost 50052 design || abort "design não subiu"
wait_for_port localhost 50053 logic  || abort "logic não subiu"
wait_for_port localhost 50054 deploy || abort "deploy não subiu"
wait_for_port localhost 50055 export || abort "export não subiu"
ok "5 services gRPC no ar"

# =============================================================================
step "4/7  Workers (consumidores RabbitMQ)"
run_bg workers go run ./services/workers/cmd
sleep 1
ok "workers iniciado (sem porta própria — acompanhe $LOG_DIR/workers.log)"

# =============================================================================
step "5/7  Gateway HTTP"
run_bg gateway go run ./services/gateway/cmd
wait_for_port localhost 8080 gateway || abort "gateway não subiu"
ok "Gateway em http://localhost:8080"

# =============================================================================
step "6/7  Collab (Phoenix)"
info "mix deps.get"
(cd services/collab && mix deps.get >/dev/null 2>&1) || warn "mix deps.get retornou erro — verifique $LOG_DIR ou rode manualmente"
run_bg collab bash -c "cd services/collab && mix phx.server"
wait_for_port localhost 4000 collab || abort "collab não subiu"
ok "Collab em http://localhost:4000"

# =============================================================================
if [ "$WITH_FRONTEND" = "1" ]; then
  step "7/7  Frontend (Vite)"
  if [ ! -d services/frontend/node_modules ]; then
    info "services/frontend/node_modules ausente — rodando npm install"
    (cd services/frontend && npm install) || abort "npm install falhou"
  fi
  ok "dependências do frontend prontas"
fi

# =============================================================================
step "Stack no ar"
cat <<EOF
  ${C_BOLD}Gateway${C_RESET}   http://localhost:8080
  ${C_BOLD}Collab${C_RESET}    http://localhost:4000
  ${C_BOLD}Jaeger${C_RESET}    http://localhost:16686
  ${C_BOLD}RabbitMQ${C_RESET}  http://localhost:15672  (mach/mach)
  ${C_BOLD}MinIO${C_RESET}     http://localhost:$MINIO_CONSOLE_HOST_PORT   (mach/machsecret)

  Logs dos processos em background: ${C_DIM}$LOG_DIR/*.log${C_RESET}
EOF

if [ "$WITH_FRONTEND" = "1" ]; then
  echo "  ${C_BOLD}Frontend${C_RESET}  http://localhost:5173  (iniciando em foreground abaixo)"
  echo
  confirm "Abrir o frontend agora (npm run dev, foreground, Ctrl+C encerra tudo)?" && {
    cd services/frontend && exec npm run dev
  }
  info "Frontend não iniciado. Rode manualmente: cd services/frontend && npm run dev"
fi

echo
echo "Ctrl+C encerra todos os processos em background iniciados por este script."
wait
