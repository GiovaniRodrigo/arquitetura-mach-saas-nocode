#!/usr/bin/env bash
# Entrega os artefatos compilados a um ambiente e ativa o release (spec 002, task 8).
#
# Transfere APENAS os tarballs de dist/artifacts para o host, extrai em
# releases/<sha>, troca o symlink `current` atomicamente e reinicia os serviços
# (RF07, RF08, RN01, RN04). Nenhum código-fonte é enviado.
#
# Uso (remoto):  build/deploy.sh --env staging --host H --user U --sha SHA
# Uso (local):   build/deploy.sh --sha SHA --dest /caminho/opt/machv4   # ensaio
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ARTIFACTS="$ROOT/dist/artifacts"
KEEP="${RELEASES_KEEP:-5}"
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
[ -n "$SHA" ] || { echo "--sha é obrigatório" >&2; exit 2; }
ls "$ARTIFACTS"/*.tar.gz >/dev/null 2>&1 || { echo "sem artefatos em $ARTIFACTS (rode build-artifacts.sh)" >&2; exit 1; }

if [ -n "$DEST" ]; then MODE=local; BASE="$DEST"; RESTART=0
else MODE=remote; BASE=/opt/machv4; RESTART=1
  [ -n "$HOST" ] && [ -n "$USER" ] || { echo "modo remoto exige --host e --user" >&2; exit 2; }
  TARGET="${USER}@${HOST}"
fi

remote_sh() { # executa o snippet $1 no alvo (local ou via ssh)
  if [ "$MODE" = local ]; then bash -eo pipefail -c "$1"
  else ssh -o BatchMode=yes "$TARGET" "$1"; fi
}

INCOMING="$BASE/.incoming-$SHA"
echo "==> [$MODE] preparando incoming em $INCOMING"
remote_sh "rm -rf '$INCOMING' && mkdir -p '$INCOMING'"

echo "==> enviando artefatos ($SHA)"
if [ "$MODE" = local ]; then cp "$ARTIFACTS"/*.tar.gz "$INCOMING/"
else rsync -az --delete "$ARTIFACTS"/ "$TARGET:$INCOMING/"; fi

echo "==> ativando release $SHA (troca atômica de symlink)"
remote_sh "$(cat <<EOF
set -e
REL='$BASE/releases/$SHA'
rm -rf "\$REL"; mkdir -p "\$REL"
for t in '$INCOMING'/*.tar.gz; do tar xzf "\$t" -C "\$REL"; done
rm -rf '$INCOMING'
ln -sfn "\$REL" '$BASE/current.tmp'
mv -Tf '$BASE/current.tmp' '$BASE/current'
if [ '$RESTART' = 1 ]; then sudo systemctl restart 'machv4-*.service'; fi
# Retém apenas as $KEEP versões mais recentes (rollback rápido).
ls -1dt '$BASE'/releases/*/ 2>/dev/null | tail -n +$((KEEP+1)) | xargs -r rm -rf
echo "current -> \$(readlink '$BASE/current')"
EOF
)"
echo "==> deploy concluído ($ENVNAME:$SHA)"
