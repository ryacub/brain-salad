.PHONY: help build test lint clean dev-cli dev-api fmt

help:
	@echo "Available targets:"
	@echo "  build          - Build all binaries"
	@echo "  build-cli      - Build CLI binary"
	@echo "  build-api      - Build API server binary"
	@echo "  test           - Run all tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  lint           - Run linters"
	@echo "  fmt            - Format code"
	@echo "  clean          - Remove build artifacts"
	@echo "  dev-cli        - Run CLI in development mode"
	@echo "  dev-api        - Run API server in development mode"

build: build-cli build-api

build-cli:
	@echo "Building CLI..."
	@CGO_ENABLED=1 go build -o bin/tm ./cmd/cli

build-api:
	@echo "Building API server..."
	@CGO_ENABLED=1 go build -o bin/tm-web ./cmd/web

test:
	@echo "Running tests..."
	@go test ./... -v

test-coverage:
	@echo "Running tests with coverage..."
	@go test ./... -cover -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-integration:
	@echo "Running integration tests..."
	@go test -tags=integration ./... -v

lint:
	@echo "Running linters..."
	@golangci-lint run

fmt:
	@echo "Formatting code..."
	@gofmt -w .
	@go mod tidy

clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

dev-cli:
	@air -c .air-cli.toml

dev-api:
	@air -c .air-api.toml
