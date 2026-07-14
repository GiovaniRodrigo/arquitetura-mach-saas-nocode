#!/usr/bin/env bash
# Rollback de um release (spec 002, task 10). Repointa o symlink `current` ao
# release anterior (ou a um sha informado) e reinicia os serviços — SEM novo build
# (RF09, RN07, RN08).
#
# Uso (remoto): scripts/rollback.sh --env production --host H --user U [--sha SHA]
# Uso (local):  scripts/rollback.sh --dest /caminho/opt/machv4 [--sha SHA]  # ensaio
set -euo pipefail

ENVNAME="" HOST="" USER="" SHA="" DEST=""
while [ $# -gt 0 ]; do
  case "$1" in
    --env) ENVNAME="$2"; shift 2;;
    --host) HOST="$2"; shift 2;;
    --user) USER="$2"; shift 2;;
    --sha) SHA="$2"; shift 2;;
    --dest) DEST="$2"; shift 2;;
    *) echo "argumento desconhecido: $1" >&2; exit 2;;
  esac
done

if [ -n "$DEST" ]; then MODE=local; BASE="$DEST"; RESTART=0
else MODE=remote; BASE=/opt/machv4; RESTART=1
  [ -n "$HOST" ] && [ -n "$USER" ] || { echo "modo remoto exige --host e --user" >&2; exit 2; }
  TARGET="${USER}@${HOST}"
fi

remote_sh() {
  if [ "$MODE" = local ]; then bash -eo pipefail -c "$1"
  else ssh -o BatchMode=yes "$TARGET" "$1"; fi
}

echo "==> [$MODE] rollback em ${ENVNAME:-$BASE}"
remote_sh "$(cat <<EOF
set -e
BASE='$BASE'
want='$SHA'
cur=\$(readlink "\$BASE/current" 2>/dev/null | xargs -r basename)
if [ -n "\$want" ]; then
  target="\$want"
else
  # Primeiro release por data que não seja o ativo (o anterior imediato).
  target=\$(ls -1dt "\$BASE"/releases/*/ 2>/dev/null | xargs -rn1 basename | grep -vx "\$cur" | head -n1)
fi
[ -n "\$target" ] || { echo "nenhum release anterior para rollback" >&2; exit 1; }
[ -d "\$BASE/releases/\$target" ] || { echo "release \$target inexistente" >&2; exit 1; }
echo "revertendo \$cur -> \$target"
ln -sfn "\$BASE/releases/\$target" "\$BASE/current.tmp"
mv -Tf "\$BASE/current.tmp" "\$BASE/current"
if [ '$RESTART' = 1 ]; then sudo systemctl restart 'machv4-*.service'; fi
echo "current -> \$(readlink "\$BASE/current")"
EOF
)"
echo "==> rollback concluído"
