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

.PHONY: dev
dev: ## Startup guiado: sobe infra + proto + services + gateway + collab + frontend (build/dev-up.sh)
	./build/dev-up.sh

.PHONY: dev-no-frontend
dev-no-frontend: ## Como 'dev', mas sem subir o frontend (útil se ele já roda à parte)
	./build/dev-up.sh --no-frontend

.PHONY: migrate
migrate: ## Aplica as migrações SQL no Postgres local
	docker compose run --rm migrate

.PHONY: seed
seed: ## Popula o stack local com um site de demonstração (requer stack no ar: make dev). Vars: SEED_EMAIL, SEED_SENHA, SEED_SISTEMA_ID
	./build/seed-demo-site.sh

.PHONY: test
test: ## Executa a suíte de testes Go
	go test ./...

.PHONY: test-html
test-html: ## Valida o HTML das telas do sistema (e2e) e do builder nocode (componente) via html-validator
	cd services/frontend && npx vitest run src/pages/Dashboard/editor/PreviewRenderer.htmlValidator.test.tsx
	cd services/frontend && npx playwright test e2e/htmlSistema.spec.ts

.PHONY: tidy
tidy: ## Sincroniza dependências do módulo Go
	go mod tidy
