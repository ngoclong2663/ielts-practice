# Run from the project root:  make be   |   make fe
.DEFAULT_GOAL := help

help: ## Show available commands
	@grep -E '^[a-z0-9-]+:.*## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "}; {printf "  make %-10s %s\n", $$1, $$2}'

be: ## Apply DB migrations and seed data, then run the Go API on :8080
	$(MAKE) -C be mgup
	$(MAKE) -C be seed
	$(MAKE) -C be run

fe: ## Install deps if needed, then run the Next.js dev server on :3000
	cd fe && ( [ -d node_modules ] || pnpm install ) && pnpm dev

install: ## Download Go modules and install frontend packages
	cd be && go mod download
	cd fe && pnpm install

migrate: ## Apply DB migrations only
	$(MAKE) -C be mgup

seed: ## Apply seed data (writing task packs, figures to S3); safe to re-run
	$(MAKE) -C be seed

test: ## Backend tests + frontend lint, typecheck and i18n check
	$(MAKE) -C be test
	cd fe && pnpm lint && pnpm typecheck && pnpm i18n:check

e2e: ## End-to-end tests (Playwright) on an isolated API/web/database
	@bash scripts/e2e.sh

build: ## Production builds of both apps
	$(MAKE) -C be build
	cd fe && pnpm build

.PHONY: help be fe install migrate seed test e2e build
