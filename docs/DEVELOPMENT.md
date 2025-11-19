# Telos Idea Matrix - Development Guide

This guide covers everything you need to know to develop, test, and contribute to the Telos Idea Matrix project.

## Table of Contents

- [Getting Started](#getting-started)
- [Project Structure](#project-structure)
- [Development Workflow](#development-workflow)
- [Testing](#testing)
- [Code Quality](#code-quality)
- [Performance](#performance)
- [Security](#security)
- [Deployment](#deployment)
- [Contributing](#contributing)

## Getting Started

### Prerequisites

- **Go 1.24.7 or higher**
- **Node.js 18+ and npm**
- **Docker and Docker Compose** (for containerized development)
- **Make** (for build automation)
- **Git**

### Initial Setup

```bash
# Clone the repository
git clone https://github.com/rayyacub/telos-idea-matrix.git
cd telos-idea-matrix

# Install Go dependencies
cd go
go mod download
cd ..

# Install frontend dependencies
cd web
npm install
cd ..

# Create a telos.md file (or use the example)
cp examples/telos.md ./telos.md

# Initialize database
mkdir -p ~/.telos
```

### Development Environment

#### Option 1: Local Development

```bash
# Terminal 1: Run API server with hot reload
cd go
make dev-api

# Terminal 2: Run frontend dev server
cd web
npm run dev
```

#### Option 2: Docker Development

```bash
# Build and start services
docker-compose up --build

# Access services:
# - API: http://localhost:8080
# - Frontend: http://localhost:5173
```

## Project Structure

```
telos-idea-matrix/
├── go/                          # Go backend
│   ├── cmd/
│   │   ├── cli/                 # CLI application
│   │   └── web/                 # Web API server
│   ├── internal/
│   │   ├── api/                 # HTTP handlers and middleware
│   │   ├── cli/                 # CLI commands
│   │   ├── config/              # Configuration management
│   │   ├── database/            # Database layer
│   │   │   └── migrations/      # SQL migrations
│   │   ├── models/              # Data models
│   │   ├── patterns/            # Anti-pattern detection
│   │   ├── scoring/             # Scoring engine
│   │   └── telos/               # Telos parsing
│   ├── pkg/                     # Public packages
│   ├── test/
│   │   └── integration/         # Integration tests
│   ├── Makefile                 # Build automation
│   └── go.mod                   # Go dependencies
│
├── web/                         # SvelteKit frontend
│   ├── src/
│   │   ├── lib/
│   │   │   └── components/      # Svelte components
│   │   ├── routes/              # SvelteKit routes
│   │   └── test/                # Test setup
│   ├── tests/                   # E2E tests
│   └── package.json             # npm dependencies
│
├── docs/                        # Documentation
│   ├── api/                     # API documentation
│   ├── ARCHITECTURE.md
│   ├── CLI_REFERENCE.md
│   └── DEVELOPMENT.md (this file)
│
├── scripts/                     # Utility scripts
│   └── deploy.sh                # Deployment script
│
├── Dockerfile                   # Multi-stage Docker build
├── docker-compose.yml           # Local development
└── README.md                    # Main documentation
```

## Development Workflow

### Making Changes

1. **Create a Feature Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make Your Changes**
   - Follow the [coding standards](#coding-standards)
   - Add tests for new functionality
   - Update documentation as needed

3. **Run Tests**
   ```bash
   # Go tests
   cd go && make test

   # Integration tests
   make test-integration

   # Frontend tests
   cd web && npm test
   ```

4. **Run Linters**
   ```bash
   # Go linter
   cd go && make lint

   # Frontend linter
   cd web && npm run check
   ```

5. **Commit Your Changes**
   ```bash
   git add .
   git commit -m "feat: add your feature description"
   ```

6. **Push and Create PR**
   ```bash
   git push origin feature/your-feature-name
   ```

### Coding Standards

#### Go Code

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Write descriptive variable names
- Keep functions small and focused
- Add godoc comments for exported functions

**Example:**

```go
// CalculateScore analyzes idea content against telos configuration
// and returns a detailed Analysis with scores and recommendations.
func (e *Engine) CalculateScore(content string) (*models.Analysis, error) {
    if content == "" {
        return nil, errors.New("content cannot be empty")
    }

    // Implementation...
}
```

#### TypeScript/Svelte Code

- Use TypeScript for type safety
- Follow [Svelte best practices](https://svelte.dev/docs)
- Use meaningful component names
- Keep components small and reusable

**Example:**

```typescript
<script lang="ts">
  interface IdeaCardProps {
    idea: Idea;
    onDelete?: (id: string) => void;
  }

  let { idea, onDelete }: IdeaCardProps = $props();
</script>
```

### Commit Message Format

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks
- `perf`: Performance improvements

**Examples:**
```
feat(api): add caching middleware for GET requests
fix(scoring): correct keyword matching algorithm
docs(cli): update CLI reference with new flags
test(integration): add concurrent access tests
```

## Testing

### Unit Tests

#### Go Unit Tests

```bash
# Run all tests
cd go && make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test ./internal/scoring -v

# Run specific test
go test ./internal/scoring -run TestCalculateScore
```

**Writing Tests:**

```go
func TestCalculateScore(t *testing.T) {
    telos := &models.Telos{
        CoreGoals: []models.Goal{
            {
                Name:     "Quality",
                Weight:   10.0,
                Keywords: []string{"quality", "testing"},
            },
        },
    }

    engine := NewEngine(telos)
    analysis, err := engine.CalculateScore("Build quality testing framework")

    assert.NoError(t, err)
    assert.Greater(t, analysis.FinalScore, 0.0)
    assert.Contains(t, analysis.MatchedGoals, "Quality")
}
```

#### Frontend Unit Tests

```bash
# Run tests
cd web && npm test

# Run tests in watch mode
npm test -- --watch

# Run tests with coverage
npm run test:coverage
```

**Writing Component Tests:**

```typescript
import { render, screen } from '@testing-library/svelte';
import { describe, it, expect } from 'vitest';
import IdeaCard from './IdeaCard.svelte';

describe('IdeaCard', () => {
  it('renders idea content', () => {
    const idea = {
      id: '123',
      content: 'Test idea',
      final_score: 8.5,
      status: 'active'
    };

    render(IdeaCard, { props: { idea } });
    expect(screen.getByText('Test idea')).toBeInTheDocument();
  });
});
```

### Integration Tests

Integration tests verify the entire stack works together:

```bash
cd go && make test-integration
```

**Writing Integration Tests:**

```go
//go:build integration
// +build integration

func TestEndToEndWorkflow(t *testing.T) {
    // Setup test database and server
    repo, _ := database.NewRepository(t.TempDir() + "/test.db")
    defer repo.Close()

    server := api.NewServer(repo, telosConfig)
    ts := httptest.NewServer(server.Router())
    defer ts.Close()

    // Test full workflow
    // 1. Create idea
    // 2. Retrieve idea
    // 3. Update idea
    // 4. Delete idea
}
```

### End-to-End Tests

Frontend E2E tests use Playwright:

```bash
cd web && npm run test:e2e

# Run with UI
npm run test:e2e:ui

# Run specific test file
npx playwright test dashboard.spec.ts
```

### Load Testing

```bash
# Run load tests
cd go && go test -tags=integration ./test/integration -run TestLoad

# Skip in short mode
go test -short ./test/integration
```

## Code Quality

### Linting

#### Go Linting

```bash
cd go && make lint

# Auto-fix issues
golangci-lint run --fix
```

Configuration: `.golangci.yml`

#### Frontend Linting

```bash
cd web && npm run check

# Type checking
npm run check:types

# Svelte check
npx svelte-check
```

### Code Coverage

```bash
# Go coverage
cd go && make test-coverage
open coverage.html

# Frontend coverage
cd web && npm run test:coverage
```

**Coverage Goals:**
- Minimum 70% coverage for new code
- Critical paths should have 90%+ coverage
- Integration tests for all API endpoints

### Security Scanning

```bash
# Go security scan
cd go && ~/go/bin/gosec ./...

# npm audit
cd web && npm audit

# Check for dependency vulnerabilities
npm audit fix
```

## Performance

### Benchmarking

```bash
# Run Go benchmarks
cd go && go test -bench=. ./internal/scoring

# Compare benchmarks
go test -bench=. -benchmem ./internal/scoring
```

**Writing Benchmarks:**

```go
func BenchmarkCalculateScore(b *testing.B) {
    engine := NewEngine(testTelos)
    content := "Build a testing framework"

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        engine.CalculateScore(content)
    }
}
```

### Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=.

# Memory profiling
go test -memprofile=mem.prof -bench=.

# Analyze profile
go tool pprof cpu.prof
```

### Database Optimization

- All queries use parameterized statements
- Indexes on frequently queried columns:
  - `idx_ideas_created_at`
  - `idx_ideas_final_score`
  - `idx_ideas_status`
  - `idx_ideas_status_score` (composite)

### API Performance

- Response caching (5-minute TTL)
- Rate limiting (100 req/min default)
- Connection pooling
- Efficient JSON serialization

## Security

### Best Practices

1. **Input Validation**
   - Validate all user inputs
   - Use parameterized queries
   - Sanitize file paths

2. **Authentication & Authorization**
   - CSRF token protection available
   - Rate limiting enabled
   - Security headers set

3. **Data Protection**
   - Sensitive data not logged
   - Database access restricted
   - CORS configured properly

### Security Headers

The API sets these security headers:
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Content-Security-Policy: default-src 'self'`

### Dependency Management

```bash
# Update Go dependencies
cd go && go get -u ./...
go mod tidy

# Update npm dependencies
cd web && npm update

# Check for vulnerabilities
npm audit
```

## Deployment

### Local Deployment

```bash
# Build and deploy
./scripts/deploy.sh deploy --env dev

# Or manually
make build
./bin/tm server
```

### Docker Deployment

```bash
# Build Docker image
docker-compose build

# Start services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### Production Deployment

```bash
# Build for production
./scripts/deploy.sh build

# Run tests
./scripts/deploy.sh test

# Deploy
./scripts/deploy.sh deploy --env prod

# Monitor
./scripts/deploy.sh status
./scripts/deploy.sh logs
```

### Environment Variables

**API Server:**
- `TELOS_FILE` - Path to telos configuration
- `PORT` - Server port (default: 8080)
- `TELOS_DB` - Database path
- `LOG_LEVEL` - Logging level

**Frontend:**
- `VITE_API_URL` - API base URL

## Contributing

### Pull Request Process

1. Fork the repository
2. Create your feature branch
3. Make your changes
4. Add tests
5. Run all tests and linters
6. Update documentation
7. Commit with conventional commit messages
8. Push to your fork
9. Create a Pull Request

### PR Checklist

- [ ] Tests added for new functionality
- [ ] All tests passing
- [ ] No linter errors
- [ ] Documentation updated
- [ ] Conventional commit messages
- [ ] PR description explains changes

### Code Review Guidelines

**Reviewers should check:**
- Code quality and style
- Test coverage
- Performance implications
- Security concerns
- Documentation completeness

## Debugging

### Go Debugging

```bash
# Use delve debugger
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug test
dlv test ./internal/scoring

# Debug server
dlv debug ./cmd/web/main.go
```

### Frontend Debugging

```bash
# Chrome DevTools
npm run dev
# Open http://localhost:5173
# Press F12 for DevTools

# Svelte DevTools
# Install browser extension
```

### Common Issues

**Database locked:**
```bash
# Kill all tm processes
killall tm

# Use separate database
tm --db /tmp/test.db list
```

**Port already in use:**
```bash
# Find process using port
lsof -i :8080

# Kill process
kill -9 <PID>

# Or use different port
tm server --port 3000
```

**Build fails:**
```bash
# Clean and rebuild
cd go && make clean && make build

# Clean npm cache
cd web && rm -rf node_modules && npm install
```

## Resources

- [Go Documentation](https://golang.org/doc/)
- [SvelteKit Documentation](https://kit.svelte.dev/docs)
- [Chi Router](https://github.com/go-chi/chi)
- [Vitest](https://vitest.dev/)
- [Playwright](https://playwright.dev/)

## Getting Help

- **Issues:** https://github.com/rayyacub/telos-idea-matrix/issues
- **Discussions:** https://github.com/rayyacub/telos-idea-matrix/discussions
- **Documentation:** https://github.com/rayyacub/telos-idea-matrix/docs
