.PHONY: help dev up down build test test-backend test-frontend lint lint-backend lint-frontend clean migrate-up migrate-down

help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-18s %s\n", $$1, $$2}'

up: ## Start all services in background via docker-compose
	docker compose up -d

down: ## Stop all services via docker-compose
	docker compose down

build: ## Build all docker images
	docker compose build

dev-backend: ## Run backend locally in development mode
	cd backend && go run cmd/server/main.go

dev-frontend: ## Run frontend locally in development mode
	cd frontend && npm run dev

test: test-backend test-frontend ## Run all test suites

test-backend: ## Run backend unit & integration tests with race detector
	cd backend && go test -v -race -cover ./...

test-frontend: ## Run frontend tests
	cd frontend && npm test --if-present

lint: lint-backend lint-frontend ## Run all linters

lint-backend: ## Run golangci-lint on backend
	cd backend && go vet ./...

lint-frontend: ## Run frontend lint
	cd frontend && npm run lint --if-present

clean: ## Clean build artifacts and temporary files
	rm -rf backend/bin frontend/dist coverage.txt coverage.html
