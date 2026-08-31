DATABASE_URL ?= postgres://saas_poc:saas_poc@localhost:5432/saas_poc?sslmode=disable
TEST_DATABASE_URL ?= postgres://saas_poc:saas_poc@localhost:5433/saas_poc_test?sslmode=disable

.PHONY: help install generate build run dev db-up db-down migrate migrate-down test test-integration docker-build docker-run clean

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

run: generate
	@echo "🚀 Starting server..."
	go run ./cmd/server/

docker-build:
	@echo "🐳 Building Docker image..."
	docker build -t saas-poc:latest .
	@echo "✅ Docker image built: saas-poc:latest"

docker-run:
	@echo "🐳 Running Docker container on http://localhost:8080..."
	docker run --rm -p 8080:8080 saas-poc:latest

dev:
	@echo "👀 Starting dev server with hot-reload..."
	go tool air

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
	TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test -v ./internal/features/dashboard || status=$$?; \
	docker compose rm -sf postgres-test; \
	exit $$status

clean:
	@echo "🧹 Cleaning up..."
	rm -f saas-poc
	find . -name "*.templ.go" -delete
	go clean
	@echo "✅ Clean complete"
