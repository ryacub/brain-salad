# Phase 0: Preparation & Foundation

**Duration:** 3-5 days
**Goal:** Set up complete Go project infrastructure and document Rust behavior

---

## Context

You are setting up the foundation for migrating the Telos Idea Matrix from Rust to Go. This phase creates the project structure, tooling, CI/CD, and reference documentation needed for all subsequent phases.

**Current State:**
- Rust implementation at: `/home/user/brain-salad/`
- Rust source code: `/home/user/brain-salad/src/`
- Rust tests: `/home/user/brain-salad/src/**/*_test.rs` (if they exist)

**Target State:**
- New Go project at: `/home/user/telos-idea-matrix-go/`
- Complete project structure
- Working CI/CD pipeline
- Development environment ready
- Rust behavior fully documented

---

## Deliverables

### 1. Project Structure

Create complete Go project structure:

```
telos-idea-matrix-go/
├── cmd/
│   ├── cli/
│   │   └── main.go              # CLI entry point
│   └── web/
│       └── main.go              # API server entry point
├── internal/
│   ├── models/
│   │   ├── idea.go
│   │   ├── telos.go
│   │   ├── analysis.go
│   │   └── models_test.go       # Test file (empty for now)
│   ├── telos/
│   │   ├── parser.go
│   │   ├── parser_test.go
│   │   └── testdata/            # Test fixtures
│   │       └── .gitkeep
│   ├── scoring/
│   │   ├── engine.go
│   │   ├── engine_test.go
│   │   └── testdata/
│   │       └── .gitkeep
│   ├── patterns/
│   │   ├── detector.go
│   │   └── detector_test.go
│   ├── database/
│   │   ├── repository.go
│   │   ├── repository_test.go
│   │   └── migrations/
│   │       └── .gitkeep
│   ├── config/
│   │   ├── config.go
│   │   └── paths.go
│   ├── cli/
│   │   ├── root.go
│   │   ├── dump.go
│   │   ├── analyze.go
│   │   ├── review.go
│   │   └── *_test.go files
│   └── api/
│       ├── server.go
│       ├── handlers.go
│       ├── middleware/
│       └── *_test.go files
├── pkg/
│   └── client/                  # Optional Go API client
│       └── client.go
├── web/                         # SvelteKit frontend (placeholder)
│   └── README.md                # Instructions for Phase 4
├── test/
│   └── integration/
│       └── .gitkeep
├── docs/
│   ├── DEVELOPMENT.md
│   ├── API.md                   # Placeholder
│   └── CLI.md                   # Placeholder
├── scripts/
│   ├── build.sh
│   ├── test.sh
│   └── deploy.sh
├── .github/
│   └── workflows/
│       ├── ci.yml
│       └── release.yml
├── .gitignore
├── .golangci.yml
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
├── docker-compose.yml
├── README.md
└── RUST_REFERENCE.md            # YOU WILL CREATE THIS
```

### 2. Initialize Go Module

```bash
cd /home/user/telos-idea-matrix-go
go mod init github.com/rayyacub/telos-idea-matrix
```

### 3. Add Dependencies

```bash
# CLI Framework
go get github.com/spf13/cobra@latest
go get github.com/spf13/viper@latest

# Web Framework
go get github.com/go-chi/chi/v5@latest
go get github.com/go-chi/cors@latest

# Database
go get github.com/mattn/go-sqlite3@latest

# Testing
go get github.com/stretchr/testify@latest

# Utilities
go get github.com/google/uuid@latest
go get github.com/fatih/color@latest
```

### 4. Create Makefile

Create a `Makefile` with these targets:

```makefile
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
```

### 5. Create .gitignore

```gitignore
# Binaries
bin/
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary, built with `go test -c`
*.test

# Output of the go coverage tool
*.out
coverage.html

# Dependency directories
vendor/

# Go workspace file
go.work

# IDEs
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Environment
.env
.env.local

# Database
*.db
!testdata/*.db

# Logs
*.log
logs/

# Build artifacts
dist/
build/

# Frontend
web/node_modules/
web/.svelte-kit/
web/build/
web/.env
```

### 6. Configure golangci-lint

Create `.golangci.yml`:

```yaml
run:
  timeout: 5m
  tests: true

linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - typecheck
    - unused
    - gofmt
    - goimports
    - misspell
    - goconst
    - dupl
    - revive

linters-settings:
  errcheck:
    check-blank: true
  govet:
    check-shadowing: true
  gofmt:
    simplify: true
  revive:
    rules:
      - name: exported
        severity: warning
      - name: error-strings
      - name: error-naming
      - name: if-return
      - name: var-naming

issues:
  exclude-use-default: false
  max-issues-per-linter: 0
  max-same-issues: 0
```

### 7. Create GitHub Actions CI/CD

Create `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches: [ main, develop, 'claude/**' ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Install dependencies
        run: go mod download

      - name: Run tests
        run: make test

      - name: Run linters
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest

      - name: Build
        run: make build

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out

  integration-test:
    name: Integration Tests
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run integration tests
        run: make test-integration
```

### 8. Document Rust Reference Behavior

**CRITICAL:** Create `RUST_REFERENCE.md` documenting:

#### Read Rust Source Code

Examine these files in the Rust codebase:
- `/home/user/brain-salad/src/types.rs` - Data structures
- `/home/user/brain-salad/src/scoring.rs` - Scoring algorithm
- `/home/user/brain-salad/src/telos.rs` - Telos parser
- `/home/user/brain-salad/src/patterns_simple.rs` - Pattern detection
- `/home/user/brain-salad/src/database_simple.rs` - Database schema
- `/home/user/brain-salad/src/commands/*.rs` - CLI commands

#### Document in RUST_REFERENCE.md

```markdown
# Rust Implementation Reference

This document describes the behavior of the Rust implementation
to ensure the Go migration maintains feature parity.

## Data Models

### Idea Struct
[Document the Rust Idea struct fields, types, and validation]

### Telos Struct
[Document Goals, Strategies, Stack, FailurePatterns]

### Analysis Struct
[Document scoring breakdown structure]

## Scoring Algorithm

### Formula
```
Final Score = (
    MissionAlignment * 0.40 +
    AntiPatternScore * 0.35 +
    StrategicFit * 0.25
) * 10.0
```

### Mission Alignment Calculation
[Document exact algorithm from Rust]

### Pattern Detection
[Document pattern matching logic]

### Strategic Fit
[Document strategy and stack alignment]

## Telos Parser

### Expected Format
[Document telos.md format]

### Parsing Rules
[Document how each section is parsed]

### Example
[Include example telos.md and expected parsed output]

## Database Schema

### Tables
[Document all tables, columns, types, indexes]

### SQL Schema
```sql
[Copy exact schema from Rust]
```

## CLI Commands

### tm dump
[Behavior, flags, output format]

### tm analyze
[Behavior, flags, output format]

### tm review
[Behavior, flags, output format]

[etc. for all commands]

## Test Cases

### High Score Example
Input: [idea text]
Expected Score: 8.5
Breakdown:
- Mission: 0.9
- Anti-pattern: 0.8
- Strategic: 0.85

### Low Score Example
Input: [idea text]
Expected Score: 2.3
Breakdown:
- Mission: 0.1
- Anti-pattern: 0.2
- Strategic: 0.4

[More test cases]
```

### 9. Create Development Guide

Create `docs/DEVELOPMENT.md`:

```markdown
# Development Guide

## Prerequisites

- Go 1.21+
- SQLite3
- golangci-lint
- make

## Setup

1. Clone the repository
2. Install dependencies: `go mod download`
3. Install tools: `make install-tools`
4. Run tests: `make test`
5. Build: `make build`

## Project Structure

[Explain directory layout]

## Development Workflow

1. Create feature branch
2. Write tests first (TDD)
3. Implement feature
4. Run tests: `make test`
5. Run linters: `make lint`
6. Commit with descriptive message
7. Push and create PR

## Testing

### Unit Tests
`go test ./internal/...`

### Integration Tests
`go test -tags=integration ./...`

### Coverage
`make test-coverage`

## Coding Standards

- Follow Go conventions
- Use gofmt for formatting
- Write godoc comments for exported functions
- Keep functions small and focused
- Write tests first (TDD)
- Target >85% coverage

## TDD Workflow

1. RED: Write failing test
2. GREEN: Write minimal code to pass
3. REFACTOR: Improve code quality

## Resources

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
```

### 10. Create Initial README

Create `README.md`:

```markdown
# Telos Idea Matrix (Go)

A command-line tool and web app for evaluating ideas against your personal goals and values.

**Status:** 🚧 In Development (Migration from Rust)

## Quick Start

### CLI

```bash
# Build
make build-cli

# Run
./bin/tm dump "Your idea here"
```

### Web UI

```bash
# Build
make build-api

# Run
./bin/tm-web
# Open http://localhost:8080
```

## Development

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md)

## Architecture

- **CLI:** Go + Cobra
- **API:** Go + Chi
- **Frontend:** SvelteKit + TypeScript
- **Database:** SQLite

## License

MIT
```

---

## Tasks

Execute these tasks in order:

### Task 1: Create Project Structure (1 hour)

```bash
# Create new directory
mkdir -p /home/user/telos-idea-matrix-go
cd /home/user/telos-idea-matrix-go

# Create all directories
mkdir -p cmd/cli cmd/web
mkdir -p internal/{models,telos/testdata,scoring/testdata,patterns,database/migrations,config,cli,api/middleware}
mkdir -p pkg/client
mkdir -p web
mkdir -p test/integration
mkdir -p docs
mkdir -p scripts
mkdir -p .github/workflows

# Create placeholder files
touch cmd/cli/main.go
touch cmd/web/main.go
touch internal/models/{idea.go,telos.go,analysis.go,models_test.go}
touch internal/telos/{parser.go,parser_test.go,testdata/.gitkeep}
touch internal/scoring/{engine.go,engine_test.go,testdata/.gitkeep}
touch internal/patterns/{detector.go,detector_test.go}
touch internal/database/{repository.go,repository_test.go,migrations/.gitkeep}
touch internal/config/{config.go,paths.go}
touch internal/cli/root.go
touch internal/api/{server.go,handlers.go}
touch pkg/client/client.go
touch web/README.md
touch test/integration/.gitkeep
touch docs/DEVELOPMENT.md
touch scripts/{build.sh,test.sh,deploy.sh}

# Initialize git
git init
git checkout -b main
```

### Task 2: Initialize Go Module (15 min)

```bash
go mod init github.com/rayyacub/telos-idea-matrix
```

Add all dependencies (see Deliverable #3 above).

### Task 3: Create Build Files (30 min)

Create:
- Makefile (see Deliverable #4)
- .gitignore (see Deliverable #5)
- .golangci.yml (see Deliverable #6)

### Task 4: Set Up CI/CD (30 min)

Create `.github/workflows/ci.yml` (see Deliverable #7).

### Task 5: Document Rust Behavior (2-3 hours)

**MOST IMPORTANT TASK**

Read the Rust source code and create `RUST_REFERENCE.md`:

1. Read `/home/user/brain-salad/src/types.rs`
   - Document all struct definitions
   - Note field types and validation

2. Read `/home/user/brain-salad/src/scoring.rs`
   - Document scoring algorithm precisely
   - Include example calculations
   - Note all constants and weights

3. Read `/home/user/brain-salad/src/telos.rs`
   - Document markdown parsing logic
   - Note regex patterns used
   - Include example telos.md

4. Read `/home/user/brain-salad/src/database_simple.rs`
   - Copy exact SQL schema
   - Document all tables and columns

5. Read `/home/user/brain-salad/src/commands/*.rs`
   - Document each command's behavior
   - Note all flags and options

6. Create test cases
   - Find or create example ideas
   - Run Rust version to get scores
   - Document expected outputs

### Task 6: Create Documentation (1 hour)

- `docs/DEVELOPMENT.md` (see Deliverable #9)
- `README.md` (see Deliverable #10)

### Task 7: Initial Commit (15 min)

```bash
git add .
git commit -m "Initial project setup

- Complete Go project structure
- Makefile with build/test/lint targets
- GitHub Actions CI/CD pipeline
- Rust reference documentation
- Development guide
- golangci-lint configuration

Ready for Phase 1: Core Domain Migration"

git remote add origin [your-repo-url]
git push -u origin main
```

---

## Validation

Before considering Phase 0 complete, verify:

### ✅ Checklist

- [ ] Project structure created (all directories exist)
- [ ] `go.mod` created with all dependencies
- [ ] `Makefile` created with all targets
- [ ] `.gitignore` created
- [ ] `.golangci.yml` created
- [ ] GitHub Actions CI/CD created (`.github/workflows/ci.yml`)
- [ ] `RUST_REFERENCE.md` complete and thorough
- [ ] `docs/DEVELOPMENT.md` created
- [ ] `README.md` created
- [ ] Initial git commit made

### 🧪 Tests

```bash
# These should work (even with no code yet)
make help           # Shows available targets
make fmt            # Formats code (no-op if no .go files yet)
go mod download     # Downloads dependencies
go mod verify       # Verifies dependencies

# These will work in CI but may not have code yet
# That's OK - we're just setting up infrastructure
```

### 📋 Deliverables Review

1. Open `RUST_REFERENCE.md` - Is it comprehensive?
2. Run `tree /home/user/telos-idea-matrix-go` - Does structure match spec?
3. Check `.github/workflows/ci.yml` - Does it run on push?
4. Review `Makefile` - All targets present?
5. Check `go.mod` - All dependencies listed?

---

## Success Criteria

Phase 0 is complete when:

✅ Complete project structure exists
✅ All build tooling configured
✅ CI/CD pipeline functional
✅ Rust behavior fully documented
✅ Development environment documented
✅ Initial commit made to repository
✅ Ready to start implementing in Phase 1

---

## Handoff to Phase 1

Once Phase 0 is complete, you're ready for **Phase 1: Core Domain Migration**.

**What Phase 1 needs from you:**
- Working Go project structure
- `RUST_REFERENCE.md` for implementation guidance
- Test infrastructure ready
- CI/CD validating all commits

**Next steps:**
1. Review Phase 1 prompt: `migration-prompts/phase-1-core-domain.md`
2. Launch Phase 1 subagent
3. Begin TDD implementation of data models

---

## Notes

- Don't write actual implementation code yet
- Focus on infrastructure and documentation
- Make RUST_REFERENCE.md as detailed as possible
- Commit early and often
- If stuck, refer to GO_MIGRATION_PLAN.md for details
