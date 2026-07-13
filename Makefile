.DEFAULT_GOAL := help
BUF ?= $(shell go env GOPATH)/bin/buf
MIGRATE_DSN ?= postgres://mach:mach@localhost:5432/machv4?sslmode=disable

.PHONY: help
help: ## Lista os alvos disponíveis
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

.PHONY: tools
tools: ## Instala ferramentas de build (buf) em $GOPATH/bin
	go install github.com/bufbuild/buf/cmd/buf@v1.42.0

.PHONY: proto
proto: ## Gera stubs Go/Elixir/TS a partir dos contratos .proto (buf)
	$(BUF) lint
	$(BUF) generate

.PHONY: proto-breaking
proto-breaking: ## Detecta quebras de contrato contra a branch main
	$(BUF) breaking --against '.git#branch=main'

.PHONY: up
up: ## Sobe a infraestrutura local (Postgres, Redis, RabbitMQ, Jaeger, MinIO)
	docker compose up -d

.PHONY: down
down: ## Derruba a infraestrutura local
	docker compose down

.PHONY: migrate
migrate: ## Aplica as migrações SQL no Postgres local
	docker compose run --rm migrate

.PHONY: test
test: ## Executa a suíte de testes Go
	go test ./...

.PHONY: tidy
tidy: ## Sincroniza dependências do módulo Go
	go mod tidy
