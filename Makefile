.PHONY: test test-unit test-integration test-coverage build run clean lint fmt deps

# Default target
all: fmt lint test build

# Dependencies
deps:
	go mod download
	go mod tidy

# Formatting
fmt:
	go fmt ./...

# Linting
lint:
	golangci-lint run

# Testing
test: test-unit

test-unit:
	go test -v ./...

test-integration:
	go test -v -tags=integration ./...

test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-coverage-console:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# Build
build:
	go build -o bin/goodreads-scraper ./cmd/api

build-docker:
	docker build -t goodreads-scraper:latest .

# Run
run:
	go run ./cmd/api

run-docker:
	docker compose up -d

# Development
dev:
	air -c .air.toml

# Benchmarks
bench:
	go test -bench=. -benchmem ./...

# Clean
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html
	docker compose down
	docker image prune -f

# Database/Cache operations
cache-stats:
	curl -s http://localhost:8080/api/v1/cache/stats | jq .

# API testing
test-api:
	@echo "Testing health endpoint..."
	curl -s http://localhost:8080/health | jq .
	@echo "\nTesting portfolio endpoint..."
	curl -s "http://localhost:8080/api/v1/portfolio/$(USER)" | jq .

# Load testing (requires hey: go install github.com/rakyll/hey@latest)
load-test:
	hey -n 100 -c 10 http://localhost:8080/health

load-test-api:
	hey -n 50 -c 5 "http://localhost:8080/api/v1/portfolio/$(USER)"

# Security scanning (requires gosec: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest)
security:
	gosec ./...

# Generate mocks (requires mockery: go install github.com/vektra/mockery/v2@latest)
mocks:
	mockery --all --output ./mocks

# Documentation
docs:
	@echo "API Documentation:"
	@echo "Health: GET /health"
	@echo "Portfolio: GET /api/v1/portfolio/:username"
	@echo "Reading Stats: GET /api/v1/reading-stats/:username"
	@echo "Shelf: GET /api/v1/reading-stats/:username/:shelf"
	@echo "Debug: GET /debug/:username"
	@echo ""
	@echo "Environment Variables:"
	@echo "PORT=8080"
	@echo "CACHE_TTL=6h"
	@echo "SCRAPE_TIMEOUT=30s"
	@echo "RATE_LIMIT_PER_MINUTE=60"
	@echo "SCRAPE_RATE_LIMIT=10"
	@echo "TRUSTED_PROXIES=127.0.0.1,::1"

# Help
help:
	@echo "Available targets:"
	@echo "  deps              - Download and tidy dependencies"
	@echo "  fmt               - Format code"
	@echo "  lint              - Run linter"
	@echo "  test              - Run all tests"
	@echo "  test-unit         - Run unit tests only"
	@echo "  test-integration  - Run integration tests only"
	@echo "  test-coverage     - Run tests with coverage report"
	@echo "  build             - Build binary"
	@echo "  build-docker      - Build Docker image"
	@echo "  run               - Run locally"
	@echo "  run-docker        - Run with Docker Compose"
	@echo "  dev               - Run with hot reload (requires air)"
	@echo "  bench             - Run benchmarks"
	@echo "  clean             - Clean build artifacts"
	@echo "  cache-stats       - Get cache statistics"
	@echo "  test-api          - Test API endpoints (USER=username)"
	@echo "  load-test         - Load test health endpoint"
	@echo "  load-test-api     - Load test API (USER=username)"
	@echo "  security          - Run security scan"
	@echo "  mocks             - Generate mocks"
	@echo "  docs              - Show API documentation"
	@echo "  help              - Show this help" 