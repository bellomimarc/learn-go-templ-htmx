.PHONY: help install generate build run docker-build docker-run test clean

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

test: generate
	@echo "🧪 Running tests..."
	go test -v ./...
	@echo "✅ Tests complete"

clean:
	@echo "🧹 Cleaning up..."
	rm -f saas-poc
	find . -name "*.templ.go" -delete
	go clean
	@echo "✅ Clean complete"
