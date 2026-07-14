#!/usr/bin/env bash
# Smoke test pós-deploy (spec 002, task 9). Verifica a saúde dos serviços após a
# ativação de um release. Exit 0 = todos saudáveis; exit != 0 dispara o rollback
# automático no cd.yml (RF11, RN08).
#
# Uso: scripts/smoke-test.sh --host H [--user U]   # --user habilita checagem do worker
set -euo pipefail

HOST="" USER="" FAIL=0
while [ $# -gt 0 ]; do
  case "$1" in
    --host) HOST="$2"; shift 2;;
    --user) USER="$2"; shift 2;;
    *) echo "argumento desconhecido: $1" >&2; exit 2;;
  esac
done
[ -n "$HOST" ] || { echo "--host é obrigatório" >&2; exit 2; }

http() { # $1=url $2=rótulo
  if curl -fsS --max-time 5 -o /dev/null "$1"; then echo "  OK   $2"
  else echo "  FALHA $2 ($1)"; FAIL=1; fi
}
tcp() { # $1=porta $2=rótulo — liveness gRPC (conexão TCP)
  if timeout 5 bash -c ">/dev/tcp/$HOST/$1" 2>/dev/null; then echo "  OK   $2 (tcp:$1)"
  else echo "  FALHA $2 (tcp:$1)"; FAIL=1; fi
}

echo "==> smoke test em $HOST"
http "http://$HOST:8080/health"  "gateway"
http "http://$HOST:4000/healthz" "collab"
http "http://$HOST/"             "player (nginx)"
tcp 50051 "iam"
tcp 50052 "design"
tcp 50053 "logic"
tcp 50054 "deploy"
tcp 50055 "export"

# Workers não expõem porta: valida a unidade systemd via SSH (se --user informado).
if [ -n "$USER" ]; then
  if ssh -o BatchMode=yes "$USER@$HOST" "systemctl is-active --quiet machv4-workers"; then
    echo "  OK   workers (systemd)"
  else echo "  FALHA workers (systemd)"; FAIL=1; fi
fi

[ "$FAIL" = 0 ] && echo "==> smoke test OK" || echo "==> smoke test FALHOU"
exit "$FAIL"
