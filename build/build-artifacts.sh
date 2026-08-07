#!/usr/bin/env bash
# Compila os artefatos de release do MACH V4 (spec 002, task 4).
#
# Fluxo: gera os stubs .proto -> compila os 7 binários Go estáticos, o release
# OTP do collab e o bundle estático do player -> empacota cada unidade em um
# tarball contendo APENAS conteúdo executável (RN01, RN05, RN06).
#
# Saída: dist/artifacts/<unidade>-<sha>.tar.gz
# Uso:   SHA=$(git rev-parse --short HEAD) build/build-artifacts.sh
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

SHA="${SHA:-$(git rev-parse --short HEAD)}"
STAGE="$ROOT/dist/release"        # espelha o layout do host: bin/ collab/ player/
OUT="$ROOT/dist/artifacts"

# --- Toolchains: em CI vêm das setup-actions; localmente, dos caminhos ~/.local.
[ -d "$HOME/.local/go/bin" ] && PATH="$HOME/.local/go/bin:$PATH"
[ -d "$HOME/.local/elixir1.17/bin" ] && PATH="$HOME/.local/elixir1.17/bin:$PATH"
if command -v go >/dev/null 2>&1; then PATH="$(go env GOPATH)/bin:$PATH"; fi
[ -d "$HOME/.mix/escripts" ] && PATH="$HOME/.mix/escripts:$PATH"
export PATH
export MIX_HOME="${MIX_HOME:-$HOME/.mix}" HEX_HOME="${HEX_HOME:-$HOME/.hex}"

echo "==> Limpando staging anterior"
rm -rf "$STAGE" "$OUT"
mkdir -p "$STAGE/bin" "$OUT"

# --- 1. Stubs .proto (gen/ é gitignored; nunca vai no artefato — RN05) -------
echo "==> buf generate"
buf generate

# --- 2. Binários Go estáticos (CGO_ENABLED=0 — RN06) -------------------------
declare -A GO_UNITS=(
  [gateway]=./gateway/cmd
  [iam]=./services/iam/cmd
  [design]=./services/design/cmd
  [logic]=./services/logic/cmd
  [deploy]=./services/deploy/cmd
  [export]=./services/export/cmd
  [workers]=./workers/cmd
)
for name in "${!GO_UNITS[@]}"; do
  echo "==> go build $name"
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags "-s -w -X main.version=$SHA" \
    -o "$STAGE/bin/$name" "${GO_UNITS[$name]}"
done

# --- 3. Release OTP do collab (autocontido: BEAM + ERTS, sem Mix/deps) -------
echo "==> mix release collab"
(
  cd "$ROOT/collab"
  export MIX_ENV=prod
  mix deps.get --only prod
  mix release collab --overwrite
)
cp -a "$ROOT/collab/_build/prod/rel/collab" "$STAGE/collab"
# Remove os templates de geradores (mix phx.gen.*) embarcados no priv/ das libs:
# são assets de desenvolvimento, inertes em runtime (RN01). O release mantém
# apenas .beam compilado + o runtime.exs (config obrigatória de boot do release).
find "$STAGE/collab" -type d -path '*/priv/templates' -prune -exec rm -rf {} +

# --- 4. Bundle estático do player (minificado, sem node_modules) -------------
echo "==> vite build (player)"
(
  cd "$ROOT/player"
  npm ci
  npm run build
)
cp -a "$ROOT/player/dist" "$STAGE/player"

# --- 5. Empacotamento por unidade (só executável) ---------------------------
echo "==> Empacotando artefatos ($SHA)"
for name in "${!GO_UNITS[@]}"; do
  tar czf "$OUT/$name-$SHA.tar.gz" -C "$STAGE" "bin/$name"
done
tar czf "$OUT/collab-$SHA.tar.gz" -C "$STAGE" collab
tar czf "$OUT/player-$SHA.tar.gz" -C "$STAGE" player

echo "==> Artefatos gerados em dist/artifacts:"
ls -1 "$OUT"
