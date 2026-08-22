# Imagem do Collab (Elixir/Phoenix) — copia o release OTP autocontido (BEAM +
# ERTS) já produzido por build/build-artifacts.sh, sem precisar do toolchain
# Elixir na imagem (spec 009: migração para Kubernetes + service mesh).
#
# Base ubuntu:24.04, não debian:12-slim: o release embute o beam.smp compilado
# no host (Ubuntu 24.04, glibc 2.39) — debian:12 tem glibc 2.36, cedo demais
# (erro "GLIBC_2.38 not found" na subida). Precisa bater a versão do host.
FROM ubuntu:24.04
RUN apt-get update && apt-get install -y --no-install-recommends \
    libncurses6 libstdc++6 openssl locales ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && sed -i '/en_US.UTF-8/s/^# //g' /etc/locale.gen && locale-gen
ENV LANG=en_US.UTF-8 LANGUAGE=en_US:en LC_ALL=en_US.UTF-8
RUN useradd --system --create-home --uid 10001 collab
COPY --chown=collab:collab dist/release/collab /app
USER collab
WORKDIR /app
ENV HOME=/app PHX_SERVER=true
ENTRYPOINT ["/app/bin/collab", "start"]
