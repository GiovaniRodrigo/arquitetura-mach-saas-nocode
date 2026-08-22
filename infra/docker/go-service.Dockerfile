# Imagem genérica para os 7 binários Go estáticos do MACH V4 (spec 009:
# migração para Kubernetes + service mesh). Não compila nada — copia o binário
# já produzido por build/build-artifacts.sh (CGO_ENABLED=0, estático), o mesmo
# artefato usado no deploy tradicional (spec 002). Parametrizado por BINARY
# para não duplicar 7 Dockerfiles idênticos.
ARG BINARY
FROM gcr.io/distroless/static-debian12:nonroot
ARG BINARY
COPY dist/release/bin/${BINARY} /app/servico
USER nonroot:nonroot
ENTRYPOINT ["/app/servico"]
