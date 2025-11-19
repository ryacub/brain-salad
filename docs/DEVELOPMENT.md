# Development Guide

This guide covers everything you need to know to develop the Telos Idea Matrix Go implementation.

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Setup](#setup)
3. [Project Structure](#project-structure)
4. [Development Workflow](#development-workflow)
5. [Testing](#testing)
6. [Coding Standards](#coding-standards)
7. [TDD Workflow](#tdd-workflow)
8. [Common Tasks](#common-tasks)
9. [Resources](#resources)

---

## Prerequisites

### Required

- **Go 1.25.4+** - [Download](https://go.dev/dl/)
- **SQLite3** - Should be installed on most systems
- **make** - For running build targets

### Recommended

- **golangci-lint** - For linting ([Installation](https://golangci-lint.run/usage/install/))
- **air** - For hot reload during development ([Installation](https://github.com/cosmtrek/air))
- **git** - For version control

### Installation Check

```bash
# Check Go version
go version  # Should be 1.25.4 or higher

# Check SQLite
sqlite3 --version

# Check make
make --version

# Check golangci-lint (optional but recommended)
golangci-lint version
```

---

## Setup

### 1. Clone the Repository

```bash
git clone https://github.com/rayyacub/telos-idea-matrix.git
cd telos-idea-matrix
```

### 2. Install Dependencies

```bash
go mod download
go mod verify
```

### 3. Build the Project

```bash
make build
```

This creates two binaries:
- `bin/tm` - CLI tool
- `bin/tm-web` - API server

### 4. Run Tests

```bash
make test
```

### 5. Optional: Install Development Tools

```bash
# Install golangci-lint (macOS/Linux)
brew install golangci-lint

# Or using go install
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install air for hot reload
go install github.com/cosmtrek/air@latest
```

---

## Project Structure

```
telos-idea-matrix-go/
├── cmd/
│   ├── cli/                    # CLI entry point
│   │   └── main.go
│   └── web/                    # API server entry point
│       └── main.go
├── internal/                   # Private application code
│   ├── models/                 # Domain models
│   │   ├── idea.go
│   │   ├── telos.go
│   │   ├── analysis.go
│   │   └── models_test.go
│   ├── telos/                  # Telos markdown parser
│   │   ├── parser.go
│   │   ├── parser_test.go
│   │   └── testdata/
│   ├── scoring/                # Scoring engine
│   │   ├── engine.go
│   │   ├── engine_test.go
│   │   └── testdata/
│   ├── patterns/               # Pattern detection
│   │   ├── detector.go
│   │   └── detector_test.go
│   ├── database/               # Database layer
│   │   ├── repository.go
│   │   ├── repository_test.go
│   │   └── migrations/
│   ├── config/                 # Configuration management
│   │   ├── config.go
│   │   └── paths.go
│   ├── cli/                    # CLI command handlers
│   │   ├── root.go
│   │   ├── dump.go
│   │   ├── analyze.go
│   │   └── review.go
│   └── api/                    # API server handlers
│       ├── server.go
│       ├── handlers.go
│       └── middleware/
├── pkg/                        # Public libraries
│   └── client/                 # Optional Go API client
│       └── client.go
├── web/                        # SvelteKit frontend (Phase 4)
│   └── README.md
├── test/                       # Integration tests
│   └── integration/
├── docs/                       # Documentation
│   ├── DEVELOPMENT.md          # This file
│   ├── API.md                  # API documentation (TBD)
│   └── CLI.md                  # CLI documentation (TBD)
├── scripts/                    # Build and deployment scripts
│   ├── build.sh
│   ├── test.sh
│   └── deploy.sh
├── .github/                    # GitHub Actions CI/CD
│   └── workflows/
│       └── ci.yml
├── .gitignore
├── .golangci.yml               # Linter configuration
├── go.mod
├── go.sum
├── Makefile
├── README.md
├── deployments/                # Deployment configurations
│   ├── docker/                 # Docker files and compose configs
│   ├── nginx/                  # Nginx configuration
│   └── monitoring/             # Prometheus, Grafana configs
└── scripts/                    # Build and utility scripts
```

### Directory Conventions

- **cmd/**: Application entry points (main packages)
- **internal/**: Private code that cannot be imported by other projects
- **pkg/**: Public libraries that can be imported by other projects
- **test/**: Integration and end-to-end tests
- **docs/**: Documentation
- **scripts/**: Build, test, and deployment scripts

---

## Development Workflow

### 1. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
```

### 2. Write Tests First (TDD)

See [TDD Workflow](#tdd-workflow) section.

### 3. Implement Feature

Write minimal code to pass tests.

### 4. Run Tests

```bash
make test
```

### 5. Run Linters

```bash
make lint
```

### 6. Format Code

```bash
make fmt
```

### 7. Commit Changes

```bash
git add .
git commit -m "feat: add feature description"
```

Follow [Conventional Commits](https://www.conventionalcommits.org/):
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation
- `test:` - Tests
- `refactor:` - Code refactoring
- `chore:` - Build/tooling changes

### 8. Push and Create PR

```bash
git push origin feature/your-feature-name
```

Then create a Pull Request on GitHub.

---

## Testing

### Unit Tests

Unit tests live alongside the code they test with `_test.go` suffix.

```bash
# Run all tests
make test

# Run tests for specific package
go test ./internal/scoring

# Run tests with verbose output
go test -v ./...

# Run specific test
go test -v ./internal/scoring -run TestScoreDomainExpertise
```

### Test Coverage

```bash
# Generate coverage report
make test-coverage

# View coverage in browser
open coverage.html
```

**Coverage Target:** >85% for all packages

### Integration Tests

Integration tests use build tags:

```bash
# Run integration tests
make test-integration

# Or manually
go test -tags=integration ./...
```

### Test Structure

```go
func TestFunctionName(t *testing.T) {
    // Arrange
    input := "test data"
    expected := 42.0

    // Act
    result, err := FunctionName(input)

    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result != expected {
        t.Errorf("expected %v, got %v", expected, result)
    }
}
```

### Table-Driven Tests

Preferred for testing multiple scenarios:

```go
func TestScoring(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected float64
        wantErr  bool
    }{
        {
            name:     "high score idea",
            input:    "Build AI tool with Python...",
            expected: 8.5,
            wantErr:  false,
        },
        {
            name:     "low score idea",
            input:    "Learn Rust before...",
            expected: 2.0,
            wantErr:  false,
        },
        {
            name:     "empty input",
            input:    "",
            expected: 0.0,
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Calculate(tt.input)

            if (err != nil) != tt.wantErr {
                t.Errorf("wantErr %v, got err %v", tt.wantErr, err)
            }

            if !tt.wantErr && math.Abs(result-tt.expected) > 0.1 {
                t.Errorf("expected %v, got %v", tt.expected, result)
            }
        })
    }
}
```

### Testdata

Use `testdata/` directories for test fixtures:

```
internal/telos/testdata/
├── example_telos.md
├── minimal_telos.md
└── invalid_telos.md
```

---

## Coding Standards

### Go Conventions

Follow official Go conventions:

1. **gofmt** - Use standard formatting
2. **Effective Go** - Follow [Effective Go](https://go.dev/doc/effective_go)
3. **Code Review Comments** - Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

### Naming

- **Files**: `snake_case.go`
- **Packages**: `lowercase`, singular
- **Types**: `PascalCase`
- **Functions/Methods**: `PascalCase` (exported), `camelCase` (unexported)
- **Variables**: `camelCase`
- **Constants**: `PascalCase` or `SCREAMING_SNAKE_CASE`

### Documentation

All exported functions, types, and packages must have godoc comments:

```go
// Score represents a calculated idea score.
// Scores range from 0.0 to 10.0, where higher scores indicate
// better alignment with the user's telos.
type Score struct {
    Value float64
    Valid bool
}

// Calculate computes a score for the given idea text.
// It returns an error if the idea text is empty or invalid.
func Calculate(idea string) (*Score, error) {
    // Implementation...
}
```

### Error Handling

```go
// Good: Return descriptive errors
if idea == "" {
    return nil, fmt.Errorf("idea text cannot be empty")
}

// Good: Wrap errors with context
result, err := db.Save(idea)
if err != nil {
    return nil, fmt.Errorf("failed to save idea: %w", err)
}

// Bad: Ignoring errors
_ = db.Save(idea)

// Bad: Generic error messages
return nil, errors.New("error")
```

### Function Guidelines

1. **Keep functions small** - Ideally < 50 lines
2. **Single responsibility** - One clear purpose
3. **Minimize parameters** - Max 3-4 parameters; use structs for more
4. **Return early** - Reduce nesting

```go
// Good
func Validate(idea string) error {
    if idea == "" {
        return ErrEmptyIdea
    }

    if len(idea) > MaxLength {
        return ErrTooLong
    }

    return nil
}

// Bad
func Validate(idea string) error {
    var err error
    if idea != "" {
        if len(idea) <= MaxLength {
            // Success path deeply nested
        } else {
            err = ErrTooLong
        }
    } else {
        err = ErrEmptyIdea
    }
    return err
}
```

---

## TDD Workflow

We follow strict Test-Driven Development:

### RED → GREEN → REFACTOR

#### 1. RED: Write Failing Test

```go
func TestCalculateScore(t *testing.T) {
    engine := NewScoringEngine()

    score, err := engine.Calculate("Build AI tool with Python")

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if score.Value < 7.0 {
        t.Errorf("expected high score, got %v", score.Value)
    }
}
```

Run test - it should **FAIL**:
```bash
go test ./internal/scoring
```

#### 2. GREEN: Write Minimal Code

```go
func (e *ScoringEngine) Calculate(idea string) (*Score, error) {
    if idea == "" {
        return nil, errors.New("empty idea")
    }

    // Minimal implementation to pass test
    return &Score{Value: 8.0, Valid: true}, nil
}
```

Run test - it should **PASS**:
```bash
go test ./internal/scoring
```

#### 3. REFACTOR: Improve Code Quality

```go
func (e *ScoringEngine) Calculate(idea string) (*Score, error) {
    if err := validateIdea(idea); err != nil {
        return nil, err
    }

    mission := e.scoreMission(idea)
    antiChallenge := e.scoreAntiChallenge(idea)
    strategic := e.scoreStrategic(idea)

    rawScore := mission + antiChallenge + strategic
    finalScore := (rawScore / 10.0) * 10.0

    return &Score{Value: finalScore, Valid: true}, nil
}
```

Run tests - they should still **PASS**:
```bash
go test ./internal/scoring
```

### TDD Benefits

- Forces you to think about API design first
- Ensures code is testable
- Provides regression protection
- Documents expected behavior

---

## Common Tasks

### Add a New CLI Command

1. Create command file: `internal/cli/mycommand.go`
2. Implement command handler
3. Register in `internal/cli/root.go`
4. Add tests: `internal/cli/mycommand_test.go`

### Add a New Scoring Dimension

1. Update `internal/models/analysis.go` with new field
2. Add scoring function in `internal/scoring/engine.go`
3. Write tests in `internal/scoring/engine_test.go`
4. Update documentation in `docs/ARCHITECTURE.md` if behavior changes

### Add a Database Field

1. Update schema in `internal/database/repository.go`
2. Create migration if needed
3. Update `StoredIdea` struct
4. Update queries
5. Add tests

### Run Development Server with Hot Reload

```bash
# CLI hot reload
make dev-cli

# API server hot reload
make dev-api
```

### Generate Test Coverage Report

```bash
make test-coverage
open coverage.html
```

### Run Linters

```bash
make lint
```

### Format All Code

```bash
make fmt
```

### Clean Build Artifacts

```bash
make clean
```

---

## Resources

### Official Documentation

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Testing](https://go.dev/doc/tutorial/add-a-test)
- [Go Modules](https://go.dev/blog/using-go-modules)

### Libraries Used

- **Cobra** - CLI framework: https://github.com/spf13/cobra
- **Viper** - Configuration: https://github.com/spf13/viper
- **Chi** - HTTP router: https://github.com/go-chi/chi
- **SQLite** - Database: https://github.com/mattn/go-sqlite3
- **Testify** - Testing toolkit: https://github.com/stretchr/testify

### Tools

- **golangci-lint** - Linter aggregator: https://golangci-lint.run/
- **air** - Hot reload: https://github.com/cosmtrek/air
- **go-coverage** - Coverage visualization: https://go.dev/blog/cover

### Project-Specific

- **README.md** - Project overview
- **docs/ARCHITECTURE.md** - System design and architecture
- **docs/API.md** - API documentation
- **docs/CLI_REFERENCE.md** - CLI command reference
- **docs/CONFIGURATION.md** - Configuration guide
- **docs/DOCKER_GUIDE.md** - Docker deployment guide

---

## Getting Help

### Common Issues

**Issue:** `cannot find package`
```bash
# Solution: Download dependencies
go mod download
go mod tidy
```

**Issue:** `database locked`
```bash
# Solution: Close other connections, check for hanging processes
pkill tm
```

**Issue:** Linter errors
```bash
# Solution: Format code and run linters
make fmt
make lint
```

### Debugging

```go
// Use fmt.Printf for quick debugging
fmt.Printf("DEBUG: value=%v\n", myValue)

// Use log package for structured logging
import "log"
log.Printf("Processing idea: %s", idea)

// Use testify for better test output
import "github.com/stretchr/testify/assert"
assert.Equal(t, expected, actual, "scores should match")
```

### Contact

- **GitHub Issues**: https://github.com/rayyacub/telos-idea-matrix/issues
- **Pull Requests**: https://github.com/rayyacub/telos-idea-matrix/pulls

---

## Next Steps

Once Phase 0 is complete:

1. **Phase 1**: Implement core domain models
2. **Phase 2**: Implement CLI commands
3. **Phase 3**: Build API server
4. **Phase 4**: Create SvelteKit frontend

Refer to the migration plan in the repository for detailed phase breakdown.

---

Happy coding! 🚀
