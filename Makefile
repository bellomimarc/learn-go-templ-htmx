DATABASE_URL ?= postgres://saas_poc:saas_poc@localhost:5432/saas_poc?sslmode=disable
TEST_DATABASE_URL ?= postgres://saas_poc:saas_poc@localhost:5433/saas_poc_test?sslmode=disable
ZITADEL_COMPOSE = docker compose --project-name saas-poc-zitadel --env-file .zitadel/zitadel.env -f deploy/zitadel/docker-compose.yml

.PHONY: help install generate build run dev db-up db-down migrate migrate-down test test-integration docker-build docker-run idp-env idp-up idp-down idp-reset idp-credentials clean

help:
	@echo "╔════════════════════════════════════════════════════════╗"
	@echo "║       SaaS Gestionale PoC - Minimal Stack 2026        ║"
	@echo "╚════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "Available commands:"
	@echo "  make install      - Download dependencies"
	@echo "  make generate     - Generate Templ code"
	@echo "  make build        - Build the application"
	@echo "  make run          - Run the application"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run Docker container"
	@echo "  make dev          - Run with hot-reload (requires air)"
	@echo "  make test         - Run tests"
	@echo "  make db-up        - Start PostgreSQL 18 for development"
	@echo "  make db-down      - Stop development PostgreSQL"
	@echo "  make migrate      - Apply database migrations"
	@echo "  make migrate-down - Roll back one database migration"
	@echo "  make test-integration - Run TODO tests against PostgreSQL"
	@echo "  make idp-up       - Start and configure local ZITADEL"
	@echo "  make idp-down     - Stop local ZITADEL"
	@echo "  make idp-reset    - Delete local ZITADEL data and credentials"
	@echo "  make idp-credentials - Show generated local user credentials"
	@echo "  make clean        - Clean build artifacts"
	@echo ""

install:
	@echo "📦 Installing dependencies..."
	go mod download
	go mod tidy
	@echo "✅ Dependencies installed"

generate:
	@echo "🔨 Generating Templ code..."
	go tool templ generate
	@echo "✅ Templ code generated"

build: generate
	@echo "🏗️  Building application..."
	go build -o saas-poc ./cmd/server/
	@echo "✅ Build complete: ./saas-poc"

run: db-up idp-up generate
	@echo "🚀 Starting server..."
	@set -a; . ./.env; set +a; go run ./cmd/server/

docker-build:
	@echo "🐳 Building Docker image..."
	docker build -t saas-poc:latest .
	@echo "✅ Docker image built: saas-poc:latest"

docker-run:
	@echo "🐳 Running Docker container on http://localhost:8080..."
	docker run --rm -p 8080:8080 saas-poc:latest

dev: db-up idp-up
	@echo "👀 Starting dev server with hot-reload..."
	@set -a; . ./.env; set +a; go tool air

db-up:
	@echo "Starting PostgreSQL 18..."
	docker compose up -d --wait postgres

db-down:
	@echo "Stopping PostgreSQL..."
	docker compose stop postgres

migrate: db-up
	@echo "Applying database migrations..."
	go tool goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	@echo "Rolling back one database migration..."
	go tool goose -dir migrations postgres "$(DATABASE_URL)" down

test: generate
	@echo "🧪 Running tests..."
	go test -v ./...
	@echo "✅ Tests complete"

test-integration: generate
	@echo "Running TODO integration tests against PostgreSQL 18..."
	@docker compose up -d --wait postgres-test; \
	status=0; \
	go tool goose -dir migrations postgres "$(TEST_DATABASE_URL)" up && \
	TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test -v ./internal/delivery/dashboard || status=$$?; \
	docker compose rm -sf postgres-test; \
	exit $$status

idp-env:
	@mkdir -p .zitadel
	@if [ ! -f .zitadel/zitadel.env ]; then \
		umask 077; \
		printf '%s\n' \
			"ZITADEL_MASTERKEY=$$(openssl rand -hex 16)" \
			"ZITADEL_DB_PASSWORD=$$(openssl rand -hex 24)" \
			"ZITADEL_ADMIN_PASSWORD=Zitadel1-$$(openssl rand -hex 12)!" \
			"SUPER_ADMIN_PASSWORD=Super1-$$(openssl rand -hex 12)!" \
			"REGULAR_USER_PASSWORD=Regular1-$$(openssl rand -hex 12)!" \
			> .zitadel/zitadel.env; \
	fi

idp-up: idp-env
	@echo "Starting and configuring ZITADEL on http://auth.localhost:8081..."
	@if [ ! -f .env ]; then cp .env.example .env; fi
	@$(ZITADEL_COMPOSE) up -d --wait zitadel-proxy
	@$(ZITADEL_COMPOSE) run --rm zitadel-config
	@test -s .zitadel/client_id
	@client_id="$$(cat .zitadel/client_id)"; \
	tmp_file="$$(mktemp .env.XXXXXX)"; \
	awk -v client_id="$$client_id" 'BEGIN { updated = 0 } /^ZITADEL_CLIENT_ID=/ { print "ZITADEL_CLIENT_ID=" client_id; updated = 1; next } { print } END { if (!updated) print "ZITADEL_CLIENT_ID=" client_id }' .env > "$$tmp_file" && mv "$$tmp_file" .env
	@echo "ZITADEL is ready. Run 'make idp-credentials' for local logins."

idp-down: idp-env
	@$(ZITADEL_COMPOSE) down

idp-reset: idp-env
	@$(ZITADEL_COMPOSE) down --volumes --remove-orphans
	@rm -rf .zitadel

idp-credentials: idp-env
	@set -a; . ./.zitadel/zitadel.env; set +a; \
	printf 'super-admin@example.test  %s\nregular-user@example.test  %s\n' "$$SUPER_ADMIN_PASSWORD" "$$REGULAR_USER_PASSWORD"

clean:
	@echo "🧹 Cleaning up..."
	rm -f saas-poc
	find . -name "*.templ.go" -delete
	go clean
	@echo "✅ Clean complete"
