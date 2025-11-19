#!/bin/bash

# Quick script to run go fmt and cleanup common issues
echo "Running go fmt to format code..."

cd /Users/rayyacub/Documents/CCResearch/telos-idea-matrix

# Run go fmt to format all code
go fmt ./...

# Run go vet for basic checks
echo "Running go vet for basic checks..."
go vet ./...

# Run golangci-lint if available (optional)
if command -v golangci-lint &> /dev/null; then
    echo "Running golangci-lint for comprehensive linting..."
    golangci-lint run --fix
else
    echo "golangci-lint not found, skipping comprehensive linting"
    echo "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
fi

echo "Go code cleanup completed!"