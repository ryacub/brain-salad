# Telos Idea Matrix: Go Migration Plan

**⚠️ NEW: Subagent Orchestration Available!**

For better execution effectiveness, see **`SUBAGENT_ORCHESTRATION.md`** which breaks this plan into discrete phases executable by specialized subagents. Copy-paste ready prompts available in `migration-prompts/`.

This document remains as the comprehensive technical specification.

---

**Migration Strategy:** Rust CLI → Go CLI + Go API + SvelteKit Frontend
**Timeline:** 6-8 weeks
**Risk Level:** Medium
**Rollback Strategy:** Keep Rust version until Go version reaches feature parity

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Migration Phases](#migration-phases)
3. [Test-Driven Development (TDD) Approach](#test-driven-development-tdd-approach)
4. [Detailed Task Breakdown](#detailed-task-breakdown)
5. [Testing Strategy](#testing-strategy)
6. [Deployment Strategy](#deployment-strategy)
7. [Success Criteria](#success-criteria)
8. [Risk Mitigation](#risk-mitigation)
9. [Rollback Plan](#rollback-plan)

---

## Architecture Overview

### Target Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Telos Idea Matrix                        │
└─────────────────────────────────────────────────────────────┘

┌──────────────────┐         ┌──────────────────┐
│   CLI Binary     │         │   Web Server     │
│   (tm)           │         │   (tm-web)       │
│                  │         │                  │
│   Cobra/Viper    │         │   Chi Router     │
└────────┬─────────┘         └────────┬─────────┘
         │                            │
         │                            │
         └────────────┬───────────────┘
                      │
         ┌────────────▼─────────────┐
         │   Shared Core (Go)       │
         │                          │
         │   • database/            │
         │   • scoring/             │
         │   • telos/               │
         │   • patterns/            │
         │   • types/               │
         └────────────┬─────────────┘
                      │
         ┌────────────▼─────────────┐
         │   SQLite Database        │
         └──────────────────────────┘

                      ▲
                      │ HTTP API
                      │
         ┌────────────┴─────────────┐
         │   SvelteKit Frontend     │
         │                          │
         │   • TypeScript           │
         │   • Tailwind CSS         │
         │   • Skeleton UI          │
         │   • Vite                 │
         └──────────────────────────┘
```

### Technology Stack

**Backend (Go):**
- **Language:** Go 1.21+
- **CLI Framework:** Cobra (commands) + Viper (config)
- **Web Framework:** Chi router (lightweight, composable)
- **Database:** database/sql + mattn/go-sqlite3
- **Validation:** go-playground/validator
- **Testing:** stdlib testing + testify
- **Serialization:** encoding/json (stdlib)

**Frontend (SvelteKit):**
- **Framework:** SvelteKit 2.x
- **Language:** TypeScript 5.x
- **Build Tool:** Vite
- **UI Components:** Skeleton UI
- **Styling:** Tailwind CSS 3.x
- **Icons:** Lucide Svelte
- **Charts:** Chart.js (if needed)
- **API Client:** Fetch API + typed client

**Infrastructure:**
- **Version Control:** Git
- **CI/CD:** GitHub Actions
- **Docker:** Multi-stage builds
- **Deployment:** Binary deployment or containers

---

## Migration Phases

### Phase 0: Preparation (Week 1)
**Goal:** Set up project structure and foundation
**Duration:** 5 days
**Risk:** Low

- [ ] Create new Go project structure
- [ ] Set up development environment
- [ ] Configure CI/CD pipeline
- [ ] Extract core domain logic from Rust
- [ ] Document current Rust behavior for reference

### Phase 1: Core Domain Migration (Week 2-3)
**Goal:** Port core business logic to Go
**Duration:** 10 days
**Risk:** Medium

- [ ] Implement data models and types
- [ ] Port Telos parser
- [ ] Port scoring engine
- [ ] Port pattern detector
- [ ] Implement database layer
- [ ] Write comprehensive tests

### Phase 2: CLI Implementation (Week 3-4)
**Goal:** Feature-complete CLI with Cobra
**Duration:** 7 days
**Risk:** Low

- [ ] Implement `dump` command
- [ ] Implement `analyze` command
- [ ] Implement `review` command
- [ ] Implement `score` command
- [ ] Implement `prune` command
- [ ] Implement `analytics` commands
- [ ] Implement `link` commands
- [ ] CLI integration tests

### Phase 3: API Server (Week 4-5)
**Goal:** RESTful API with full CRUD operations
**Duration:** 7 days
**Risk:** Low

- [ ] Set up Chi router and middleware
- [ ] Implement authentication (if needed)
- [ ] Implement API endpoints for ideas
- [ ] Implement API endpoints for analysis
- [ ] Implement API endpoints for analytics
- [ ] OpenAPI/Swagger documentation
- [ ] API integration tests

### Phase 4: SvelteKit Frontend (Week 5-7)
**Goal:** Beautiful, responsive web UI
**Duration:** 14 days
**Risk:** Medium

- [ ] Set up SvelteKit project
- [ ] Configure Tailwind + Skeleton UI
- [ ] Implement type-safe API client
- [ ] Build dashboard page
- [ ] Build idea detail page
- [ ] Build idea submission form
- [ ] Build analytics/charts page
- [ ] Build review/filtering interface
- [ ] Responsive design (mobile/tablet/desktop)
- [ ] E2E tests with Playwright

### Phase 5: Integration & Polish (Week 7-8)
**Goal:** Production-ready system
**Duration:** 7 days
**Risk:** Low

- [ ] End-to-end integration testing
- [ ] Performance optimization
- [ ] Security audit
- [ ] Documentation (user + developer)
- [ ] Docker images
- [ ] Deployment scripts
- [ ] Migration guide for users

### Phase 6: Beta Release (Week 8)
**Goal:** Limited release for testing
**Duration:** 3 days
**Risk:** Low

- [ ] Beta deployment
- [ ] User acceptance testing
- [ ] Bug fixes
- [ ] Gather feedback
- [ ] Prepare for v1.0 release

---

## Test-Driven Development (TDD) Approach

**CRITICAL:** This entire migration must follow Test-Driven Development practices. Write tests FIRST, then implement code to make them pass.

### TDD Principles

**Red → Green → Refactor Cycle:**

1. **🔴 RED**: Write a failing test
   - Define expected behavior
   - Write test that fails (code doesn't exist yet)
   - Verify test actually fails

2. **🟢 GREEN**: Write minimal code to pass
   - Implement just enough to make test pass
   - Don't worry about perfection
   - Get to green as quickly as possible

3. **♻️ REFACTOR**: Improve the code
   - Clean up implementation
   - Remove duplication
   - Improve readability
   - Keep tests passing

### Coverage Targets

```
Component                    Target    Priority
───────────────────────────────────────────────
Models & Types               >90%      CRITICAL
Scoring Engine               >95%      CRITICAL
Telos Parser                 >90%      CRITICAL
Pattern Detector             >90%      CRITICAL
Database Repository          >85%      HIGH
API Handlers                 >85%      HIGH
CLI Commands                 >80%      HIGH
Middleware                   >75%      MEDIUM
Utils/Helpers                >85%      MEDIUM
───────────────────────────────────────────────
OVERALL TARGET               >85%      REQUIRED
```

### TDD Workflow for Each Component

#### Example: Implementing the Scoring Engine

**Step 1: RED - Write Failing Tests First**

```go
// internal/scoring/engine_test.go
package scoring_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/rayyacub/telos-idea-matrix/internal/models"
    "github.com/rayyacub/telos-idea-matrix/internal/scoring"
)

func TestScoreIdea_PerfectAlignment_Returns10(t *testing.T) {
    // Arrange
    telos := &models.Telos{
        Goals: []models.Goal{
            {Description: "Build a SaaS product", Priority: 1},
        },
    }
    idea := &models.Idea{
        Title: "Create a new SaaS application",
        Description: "Build a profitable SaaS product",
    }

    engine := scoring.NewEngine(telos)

    // Act
    analysis, err := engine.Score(idea)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, analysis)
    assert.GreaterOrEqual(t, analysis.Score, 8.0, "Perfect alignment should score >8")
    assert.Equal(t, 1.0, analysis.MissionAlignment, "Mission alignment should be 100%")
}

func TestScoreIdea_NoAlignment_ReturnsLowScore(t *testing.T) {
    // Arrange
    telos := &models.Telos{
        Goals: []models.Goal{
            {Description: "Build software products", Priority: 1},
        },
    }
    idea := &models.Idea{
        Title: "Learn underwater basket weaving",
        Description: "Take a course on basket weaving techniques",
    }

    engine := scoring.NewEngine(telos)

    // Act
    analysis, err := engine.Score(idea)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, analysis)
    assert.Less(t, analysis.Score, 4.0, "No alignment should score <4")
    assert.Less(t, analysis.MissionAlignment, 0.3)
}

func TestScoreIdea_DetectsContextSwitching_ReducesScore(t *testing.T) {
    // Arrange
    telos := &models.Telos{
        Goals: []models.Goal{
            {Description: "Complete current project", Priority: 1},
        },
        FailurePatterns: []models.Pattern{
            {
                Name: "Context Switching",
                Description: "Starting new projects before finishing current ones",
                Keywords: []string{"new project", "start", "begin", "launch"},
            },
        },
    }
    idea := &models.Idea{
        Title: "Start a brand new project",
        Description: "I want to begin working on something new",
    }

    engine := scoring.NewEngine(telos)

    // Act
    analysis, err := engine.Score(idea)

    // Assert
    assert.NoError(t, err)
    assert.NotEmpty(t, analysis.DetectedPatterns)
    assert.Equal(t, "Context Switching", analysis.DetectedPatterns[0].Name)
    assert.Greater(t, analysis.DetectedPatterns[0].Confidence, 0.5)
}
```

**Run tests - they should FAIL:**
```bash
$ go test ./internal/scoring -v
# Output: cannot find package "github.com/rayyacub/telos-idea-matrix/internal/scoring"
# This is expected! We haven't written the code yet.
```

**Step 2: GREEN - Implement Minimal Code**

```go
// internal/scoring/engine.go
package scoring

import (
    "github.com/rayyacub/telos-idea-matrix/internal/models"
)

type Engine struct {
    telos *models.Telos
}

func NewEngine(telos *models.Telos) *Engine {
    return &Engine{telos: telos}
}

func (e *Engine) Score(idea *models.Idea) (*models.Analysis, error) {
    analysis := &models.Analysis{}

    // Calculate mission alignment (minimal implementation)
    analysis.MissionAlignment = e.calculateMissionAlignment(idea)

    // Detect patterns
    analysis.DetectedPatterns = e.detectPatterns(idea)
    analysis.AntiPatternScore = e.calculateAntiPatternScore(analysis.DetectedPatterns)

    // Calculate strategic fit
    analysis.StrategicFit = 0.5 // TODO: implement

    // Calculate final score
    analysis.Score = (
        analysis.MissionAlignment*0.40 +
        analysis.AntiPatternScore*0.35 +
        analysis.StrategicFit*0.25,
    ) * 10.0

    return analysis, nil
}

func (e *Engine) calculateMissionAlignment(idea *models.Idea) float64 {
    // Minimal implementation - just keyword matching
    // TODO: enhance with NLP, embeddings, etc.
    ideaText := strings.ToLower(idea.Title + " " + idea.Description)

    var totalScore float64
    for _, goal := range e.telos.Goals {
        goalWords := strings.Fields(strings.ToLower(goal.Description))
        matches := 0
        for _, word := range goalWords {
            if len(word) > 3 && strings.Contains(ideaText, word) {
                matches++
            }
        }
        if len(goalWords) > 0 {
            totalScore += float64(matches) / float64(len(goalWords))
        }
    }

    if len(e.telos.Goals) == 0 {
        return 0.5
    }

    return totalScore / float64(len(e.telos.Goals))
}

func (e *Engine) detectPatterns(idea *models.Idea) []models.DetectedPattern {
    var detected []models.DetectedPattern
    ideaText := strings.ToLower(idea.Title + " " + idea.Description)

    for _, pattern := range e.telos.FailurePatterns {
        matches := 0
        for _, keyword := range pattern.Keywords {
            if strings.Contains(ideaText, strings.ToLower(keyword)) {
                matches++
            }
        }

        if matches > 0 {
            confidence := float64(matches) / float64(len(pattern.Keywords))
            detected = append(detected, models.DetectedPattern{
                Name:        pattern.Name,
                Description: pattern.Description,
                Confidence:  confidence,
                Severity:    "medium", // TODO: calculate based on confidence
            })
        }
    }

    return detected
}

func (e *Engine) calculateAntiPatternScore(patterns []models.DetectedPattern) float64 {
    if len(patterns) == 0 {
        return 1.0
    }

    var totalPenalty float64
    for _, p := range patterns {
        totalPenalty += p.Confidence
    }

    penalty := totalPenalty / float64(len(patterns))
    if penalty > 1.0 {
        penalty = 1.0
    }

    return 1.0 - penalty
}
```

**Run tests again - they should PASS:**
```bash
$ go test ./internal/scoring -v
=== RUN   TestScoreIdea_PerfectAlignment_Returns10
--- PASS: TestScoreIdea_PerfectAlignment_Returns10 (0.00s)
=== RUN   TestScoreIdea_NoAlignment_ReturnsLowScore
--- PASS: TestScoreIdea_NoAlignment_ReturnsLowScore (0.00s)
=== RUN   TestScoreIdea_DetectsContextSwitching_ReducesScore
--- PASS: TestScoreIdea_DetectsContextSwitching_ReducesScore (0.00s)
PASS
ok      github.com/rayyacub/telos-idea-matrix/internal/scoring  0.123s
```

**Step 3: REFACTOR - Improve Code Quality**

```go
// Extract constants
const (
    MissionAlignmentWeight = 0.40
    AntiPatternWeight      = 0.35
    StrategicFitWeight     = 0.25
    MinWordLength          = 3  // Ignore short words
)

// Extract helper functions
func calculateTextOverlap(text1, text2 string) float64 {
    words1 := strings.Fields(strings.ToLower(text1))
    words2 := strings.Fields(strings.ToLower(text2))

    matches := 0
    for _, w1 := range words1 {
        if len(w1) <= MinWordLength {
            continue
        }
        for _, w2 := range words2 {
            if w1 == w2 {
                matches++
                break
            }
        }
    }

    if len(words1) == 0 {
        return 0
    }

    return float64(matches) / float64(len(words1))
}

// Refactored calculateMissionAlignment
func (e *Engine) calculateMissionAlignment(idea *models.Idea) float64 {
    if len(e.telos.Goals) == 0 {
        return 0.5 // Neutral if no goals defined
    }

    ideaText := idea.Title + " " + idea.Description

    var totalScore float64
    for _, goal := range e.telos.Goals {
        overlap := calculateTextOverlap(ideaText, goal.Description)
        totalScore += overlap
    }

    return totalScore / float64(len(e.telos.Goals))
}
```

**Run tests one more time - ensure refactoring didn't break anything:**
```bash
$ go test ./internal/scoring -v -cover
=== RUN   TestScoreIdea_PerfectAlignment_Returns10
--- PASS: TestScoreIdea_PerfectAlignment_Returns10 (0.00s)
=== RUN   TestScoreIdea_NoAlignment_ReturnsLowScore
--- PASS: TestScoreIdea_NoAlignment_ReturnsLowScore (0.00s)
=== RUN   TestScoreIdea_DetectsContextSwitching_ReducesScore
--- PASS: TestScoreIdea_DetectsContextSwitching_ReducesScore (0.00s)
PASS
coverage: 87.3% of statements
ok      github.com/rayyacub/telos-idea-matrix/internal/scoring  0.135s
```

**✅ Cycle complete! Move to next feature.**

### Test Organization Standards

#### Directory Structure

```
internal/
  scoring/
    engine.go              # Implementation
    engine_test.go         # Unit tests
    benchmarks_test.go     # Performance tests
    testdata/              # Test fixtures
      high_score_idea.json
      low_score_idea.json
      test_telos.md
    examples_test.go       # Example usage (godoc)
```

#### Test File Naming

```go
// Unit tests: *_test.go in same package
package scoring

func TestScoreIdea(t *testing.T) { }

// Integration tests: *_integration_test.go
//go:build integration
package scoring_test

func TestScoringWithDatabase(t *testing.T) { }

// Benchmarks: benchmark_*.go or *_test.go
func BenchmarkScoreIdea(b *testing.B) { }
```

#### Test Naming Convention

```go
// Format: Test<Function>_<Scenario>_<ExpectedBehavior>

func TestScoreIdea_PerfectAlignment_Returns10(t *testing.T) { }
func TestScoreIdea_NoGoalsDefined_ReturnsNeutralScore(t *testing.T) { }
func TestParseFile_ValidTelos_ParsesAllSections(t *testing.T) { }
func TestParseFile_MissingFile_ReturnsError(t *testing.T) { }
func TestCreateIdea_DuplicateID_ReturnsError(t *testing.T) { }
```

### Table-Driven Tests

Use table-driven tests for multiple scenarios:

```go
func TestCalculateMissionAlignment(t *testing.T) {
    tests := []struct {
        name         string
        idea         *models.Idea
        telos        *models.Telos
        expectedMin  float64
        expectedMax  float64
    }{
        {
            name: "perfect keyword match",
            idea: &models.Idea{
                Title: "Build a SaaS product",
            },
            telos: &models.Telos{
                Goals: []models.Goal{
                    {Description: "Launch a SaaS product"},
                },
            },
            expectedMin: 0.8,
            expectedMax: 1.0,
        },
        {
            name: "partial match",
            idea: &models.Idea{
                Title: "Build a mobile app",
            },
            telos: &models.Telos{
                Goals: []models.Goal{
                    {Description: "Create software products"},
                },
            },
            expectedMin: 0.3,
            expectedMax: 0.7,
        },
        {
            name: "no match",
            idea: &models.Idea{
                Title: "Learn underwater basket weaving",
            },
            telos: &models.Telos{
                Goals: []models.Goal{
                    {Description: "Build tech products"},
                },
            },
            expectedMin: 0.0,
            expectedMax: 0.2,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            engine := NewEngine(tt.telos)
            result := engine.calculateMissionAlignment(tt.idea)

            assert.GreaterOrEqual(t, result, tt.expectedMin)
            assert.LessOrEqual(t, result, tt.expectedMax)
        })
    }
}
```

### Test Fixtures and Helpers

#### Create Reusable Test Helpers

```go
// internal/scoring/testhelpers_test.go
package scoring_test

import (
    "testing"
    "github.com/rayyacub/telos-idea-matrix/internal/models"
)

func createTestTelos(t *testing.T) *models.Telos {
    t.Helper()
    return &models.Telos{
        Goals: []models.Goal{
            {ID: "G1", Description: "Build a SaaS product", Priority: 1},
            {ID: "G2", Description: "Launch by Q4 2025", Priority: 2},
        },
        Strategies: []models.Strategy{
            {ID: "S1", Description: "Ship early and iterate"},
            {ID: "S2", Description: "Focus on one tech stack"},
        },
        Stack: models.Stack{
            Primary:   []string{"Go", "TypeScript", "PostgreSQL"},
            Secondary: []string{"Docker", "GitHub Actions"},
        },
        FailurePatterns: []models.Pattern{
            {
                Name:        "Context Switching",
                Description: "Starting new projects before finishing",
                Keywords:    []string{"new project", "start", "begin"},
            },
        },
    }
}

func createHighScoreIdea(t *testing.T) *models.Idea {
    t.Helper()
    return &models.Idea{
        Title:       "Build a Go-based SaaS product",
        Description: "Ship early MVP using Go and PostgreSQL",
        Status:      "pending",
    }
}

func createLowScoreIdea(t *testing.T) *models.Idea {
    t.Helper()
    return &models.Idea{
        Title:       "Learn a new programming language",
        Description: "Start a tutorial on Haskell",
        Status:      "pending",
    }
}
```

#### Use Test Fixtures

```go
// testdata/valid_telos.md
# My Telos

## Goals
- G1: Build a profitable SaaS product (Deadline: 2025-12-31)
- G2: Establish personal brand (Deadline: 2025-06-30)

## Strategies
- S1: Ship early and often
- S2: Focus on one technology stack

## Stack
- Primary: Go, TypeScript, PostgreSQL
- Secondary: Docker, Kubernetes

## Failure Patterns
- Context switching: Starting new projects before finishing current ones
- Perfectionism: Over-engineering before validating market fit
```

```go
// Load fixtures in tests
func TestParseFile_ValidTelos_ParsesCorrectly(t *testing.T) {
    parser := NewParser()

    telos, err := parser.ParseFile("testdata/valid_telos.md")

    assert.NoError(t, err)
    assert.Len(t, telos.Goals, 2)
    assert.Len(t, telos.Strategies, 2)
    assert.Len(t, telos.FailurePatterns, 2)
    assert.Equal(t, "Build a profitable SaaS product", telos.Goals[0].Description)
}
```

### Integration Tests

Mark integration tests with build tags:

```go
//go:build integration
// +build integration

package database_test

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/rayyacub/telos-idea-matrix/internal/database"
)

func TestDatabaseIntegration_CreateAndRetrieveIdea(t *testing.T) {
    // Setup: Create temp database
    repo, cleanup := setupTestDB(t)
    defer cleanup()

    // Create idea
    idea := &models.Idea{
        Title:       "Test Integration",
        Description: "Testing database integration",
        Score:       7.5,
        Status:      "pending",
    }

    err := repo.CreateIdea(context.Background(), idea)
    assert.NoError(t, err)
    assert.NotEmpty(t, idea.ID)

    // Retrieve idea
    retrieved, err := repo.GetIdea(context.Background(), idea.ID)
    assert.NoError(t, err)
    assert.Equal(t, idea.Title, retrieved.Title)
    assert.Equal(t, idea.Score, retrieved.Score)
}

func setupTestDB(t *testing.T) (*database.Repository, func()) {
    t.Helper()

    tmpfile, err := os.CreateTemp("", "test_*.db")
    require.NoError(t, err)

    repo, err := database.NewRepository(tmpfile.Name())
    require.NoError(t, err)

    cleanup := func() {
        repo.Close()
        os.Remove(tmpfile.Name())
    }

    return repo, cleanup
}
```

Run integration tests separately:
```bash
# Unit tests only
go test ./...

# Integration tests only
go test -tags=integration ./...

# All tests
go test -tags=integration ./... -v
```

### Benchmarks

Write benchmarks for performance-critical code:

```go
func BenchmarkScoreIdea(b *testing.B) {
    telos := createTestTelos(b)
    idea := createHighScoreIdea(b)
    engine := NewEngine(telos)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := engine.Score(idea)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkCalculateMissionAlignment(b *testing.B) {
    telos := createTestTelos(b)
    idea := createHighScoreIdea(b)
    engine := NewEngine(telos)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = engine.calculateMissionAlignment(idea)
    }
}
```

Run benchmarks:
```bash
go test -bench=. ./internal/scoring
go test -bench=BenchmarkScoreIdea -benchmem ./internal/scoring
```

### Mocking External Dependencies

For components with external dependencies, use interfaces and mocks:

```go
// Define interface
type LLMClient interface {
    Analyze(ctx context.Context, text string) (*Analysis, error)
}

// Create mock for testing
type MockLLMClient struct {
    mock.Mock
}

func (m *MockLLMClient) Analyze(ctx context.Context, text string) (*Analysis, error) {
    args := m.Called(ctx, text)
    return args.Get(0).(*Analysis), args.Error(1)
}

// Use in tests
func TestAnalyzeWithLLM(t *testing.T) {
    mockClient := new(MockLLMClient)
    mockClient.On("Analyze", mock.Anything, "Test idea").
        Return(&Analysis{Score: 8.5}, nil)

    service := NewService(mockClient)
    result, err := service.AnalyzeIdea("Test idea")

    assert.NoError(t, err)
    assert.Equal(t, 8.5, result.Score)
    mockClient.AssertExpectations(t)
}
```

### TDD Checklist for Each Component

Before moving to the next component, ensure:

- [ ] All tests written BEFORE implementation
- [ ] All tests pass (`go test ./... -v`)
- [ ] Coverage meets target for component
- [ ] No skipped tests without justification
- [ ] Edge cases covered (nil, empty, invalid input)
- [ ] Error cases tested
- [ ] Happy path tested
- [ ] Integration tests written (if applicable)
- [ ] Benchmarks written (for critical paths)
- [ ] Documentation examples work (run with `go test`)
- [ ] Tests run in CI pipeline
- [ ] Code reviewed for test quality

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package
go test ./internal/scoring -v -cover

# Run with race detector
go test -race ./...

# Run integration tests
go test -tags=integration ./...

# Run benchmarks
go test -bench=. ./...

# Watch mode (using air or similar)
air test ./...

# Generate coverage HTML report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### TDD Anti-Patterns to Avoid

❌ **DON'T:**
- Write implementation before tests
- Write tests after implementation is "done"
- Skip tests for "simple" functions
- Test implementation details instead of behavior
- Have tests that depend on each other
- Write tests that are flaky (non-deterministic)
- Mock everything (mock only external dependencies)
- Have tests with no assertions
- Copy-paste tests without understanding them

✅ **DO:**
- Write failing test first (RED)
- Write minimal code to pass (GREEN)
- Refactor while keeping tests green (REFACTOR)
- Test behavior, not implementation
- Keep tests independent
- Make tests deterministic and fast
- Use real objects when possible
- Assert expected behavior clearly
- Understand what each test validates

### Validation Gates

**Before merging any PR:**
```bash
# All checks must pass
make lint              # No linter errors
make test              # All tests pass
make test-coverage     # Coverage >85%
make test-integration  # Integration tests pass
make benchmark         # No performance regression
```

**CI Pipeline enforces:**
- All tests pass
- Coverage >85% overall
- No security vulnerabilities (`gosec`)
- Code formatted (`gofmt`)
- No linter errors (`golangci-lint`)

---

## Detailed Task Breakdown

### Phase 0: Preparation

#### Task 0.1: Project Structure Setup
**Estimated Time:** 4 hours

```
telos-idea-matrix-go/
├── cmd/
│   ├── cli/                    # CLI binary entry point
│   │   └── main.go
│   └── web/                    # Web server entry point
│       └── main.go
├── internal/
│   ├── database/               # Database layer
│   │   ├── migrations/
│   │   ├── repository.go
│   │   └── sqlite.go
│   ├── scoring/                # Scoring engine
│   │   ├── engine.go
│   │   ├── weights.go
│   │   └── engine_test.go
│   ├── telos/                  # Telos parsing
│   │   ├── parser.go
│   │   ├── validator.go
│   │   └── parser_test.go
│   ├── patterns/               # Pattern detection
│   │   ├── detector.go
│   │   └── detector_test.go
│   ├── models/                 # Domain models
│   │   ├── idea.go
│   │   ├── telos.go
│   │   └── analysis.go
│   ├── config/                 # Configuration
│   │   ├── config.go
│   │   └── paths.go
│   ├── cli/                    # CLI commands
│   │   ├── dump.go
│   │   ├── analyze.go
│   │   ├── review.go
│   │   └── root.go
│   └── api/                    # API handlers
│       ├── handlers/
│       ├── middleware/
│       └── server.go
├── pkg/
│   └── client/                 # Go API client (optional)
│       └── client.go
├── web/                        # SvelteKit frontend
│   ├── src/
│   │   ├── routes/
│   │   ├── lib/
│   │   └── app.html
│   ├── static/
│   ├── package.json
│   └── svelte.config.js
├── test/
│   ├── integration/            # Integration tests
│   └── testdata/               # Test fixtures
├── scripts/
│   ├── build.sh
│   ├── test.sh
│   └── deploy.sh
├── docs/
│   ├── API.md
│   ├── CLI.md
│   └── DEVELOPMENT.md
├── .github/
│   └── workflows/
│       ├── ci.yml
│       └── release.yml
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
├── docker-compose.yml
└── README.md
```

**Action Items:**
- [ ] Create directory structure
- [ ] Initialize Go module: `go mod init github.com/rayyacub/telos-idea-matrix`
- [ ] Create initial `go.mod` with dependencies
- [ ] Set up `.gitignore`
- [ ] Create `Makefile` with build/test targets

#### Task 0.2: Development Environment
**Estimated Time:** 2 hours

**Tools to install:**
- Go 1.21+
- Node.js 18+ (for SvelteKit)
- SQLite3 CLI
- golangci-lint
- air (live reload for Go)

**Action Items:**
- [ ] Document setup in `docs/DEVELOPMENT.md`
- [ ] Create `.envrc` or `.env.example` for environment variables
- [ ] Set up VSCode/editor config (`.vscode/settings.json`)
- [ ] Configure linters (`.golangci.yml`)

#### Task 0.3: CI/CD Pipeline
**Estimated Time:** 4 hours

**GitHub Actions workflows:**

`.github/workflows/ci.yml`:
```yaml
name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test-go:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Run tests
        run: make test
      - name: Run linters
        run: make lint
      - name: Build binaries
        run: make build

  test-frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '18'
      - name: Install dependencies
        run: cd web && npm ci
      - name: Run tests
        run: cd web && npm test
      - name: Build
        run: cd web && npm run build
```

**Action Items:**
- [ ] Create CI workflow
- [ ] Create release workflow (goreleaser)
- [ ] Set up code coverage reporting
- [ ] Configure dependabot

#### Task 0.4: Extract Rust Domain Logic
**Estimated Time:** 6 hours

**Goal:** Document current behavior for faithful port

**Action Items:**
- [ ] Create `RUST_REFERENCE.md` documenting:
  - Scoring algorithm with examples
  - Pattern detection rules
  - Telos file format specification
  - Database schema
  - CLI command behaviors
  - Expected inputs/outputs
- [ ] Extract test cases from Rust tests
- [ ] Document edge cases and error handling
- [ ] Create test data fixtures

**Example documentation:**
```markdown
## Scoring Algorithm

### Inputs
- Idea text: String
- Telos configuration: Goals, Strategies, Stack, Patterns

### Calculation
1. Mission Alignment (40% weight)
   - Parse goals from Telos
   - Calculate keyword overlap
   - Score: 0.0 - 1.0

2. Anti-Pattern Detection (35% weight)
   - Check against known patterns
   - Pattern match counts
   - Score: 0.0 - 1.0 (inverted)

3. Strategic Fit (25% weight)
   - Match against strategies
   - Stack alignment
   - Score: 0.0 - 1.0

### Output
- Final score: Weighted average * 10
- Range: 0.0 - 10.0
- Precision: 1 decimal place
```

---

### Phase 1: Core Domain Migration

#### Task 1.1: Data Models
**Estimated Time:** 4 hours

**File:** `internal/models/idea.go`
```go
package models

import (
    "time"
    "github.com/google/uuid"
)

// Idea represents a captured idea with analysis
type Idea struct {
    ID          uuid.UUID  `json:"id" db:"id"`
    Title       string     `json:"title" db:"title" validate:"required,min=3,max=200"`
    Description string     `json:"description" db:"description"`
    Score       float64    `json:"score" db:"score"`
    CreatedAt   time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
    Status      string     `json:"status" db:"status" validate:"oneof=pending in-progress completed archived"`
    Tags        []string   `json:"tags"`
    Analysis    *Analysis  `json:"analysis,omitempty"`
}

// Analysis represents the scoring breakdown
type Analysis struct {
    Score              float64           `json:"score"`
    MissionAlignment   float64           `json:"mission_alignment"`
    AntiPatternScore   float64           `json:"anti_pattern_score"`
    StrategicFit       float64           `json:"strategic_fit"`
    DetectedPatterns   []DetectedPattern `json:"detected_patterns"`
    Recommendations    []string          `json:"recommendations"`
    AnalyzedAt         time.Time         `json:"analyzed_at"`
}

// DetectedPattern represents an anti-pattern match
type DetectedPattern struct {
    Name        string  `json:"name"`
    Description string  `json:"description"`
    Confidence  float64 `json:"confidence"`
    Severity    string  `json:"severity"` // "low", "medium", "high"
}

// Validate validates the idea
func (i *Idea) Validate() error {
    // Use go-playground/validator
    return validate.Struct(i)
}
```

**File:** `internal/models/telos.go`
```go
package models

import "time"

// Telos represents the user's goals and values
type Telos struct {
    Goals           []Goal     `json:"goals"`
    Strategies      []Strategy `json:"strategies"`
    Stack           Stack      `json:"stack"`
    FailurePatterns []Pattern  `json:"failure_patterns"`
    LoadedAt        time.Time  `json:"loaded_at"`
}

// Goal represents a user goal with deadline
type Goal struct {
    ID          string     `json:"id"`
    Description string     `json:"description"`
    Deadline    *time.Time `json:"deadline,omitempty"`
    Priority    int        `json:"priority"`
}

// Strategy represents a strategic approach
type Strategy struct {
    ID          string `json:"id"`
    Description string `json:"description"`
}

// Stack represents technology preferences
type Stack struct {
    Primary   []string `json:"primary"`
    Secondary []string `json:"secondary"`
}

// Pattern represents a failure pattern to avoid
type Pattern struct {
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Keywords    []string `json:"keywords"`
}
```

**Action Items:**
- [ ] Implement all model structs
- [ ] Add validation tags
- [ ] Add JSON serialization tags
- [ ] Add database tags
- [ ] Write unit tests for validation
- [ ] Document model constraints

#### Task 1.2: Telos Parser
**Estimated Time:** 8 hours

**File:** `internal/telos/parser.go`
```go
package telos

import (
    "bufio"
    "os"
    "regexp"
    "strings"
    "time"

    "github.com/rayyacub/telos-idea-matrix/internal/models"
)

// Parser parses telos.md files
type Parser struct {
    goalRegex     *regexp.Regexp
    strategyRegex *regexp.Regexp
    deadlineRegex *regexp.Regexp
}

// NewParser creates a new Telos parser
func NewParser() *Parser {
    return &Parser{
        goalRegex:     regexp.MustCompile(`^-\s+G\d+:\s+(.+?)(?:\s+\(Deadline:\s+(.+?)\))?$`),
        strategyRegex: regexp.MustCompile(`^-\s+S\d+:\s+(.+)$`),
        deadlineRegex: regexp.MustCompile(`\(Deadline:\s+(\d{4}-\d{2}-\d{2})\)`),
    }
}

// ParseFile parses a telos.md file
func (p *Parser) ParseFile(path string) (*models.Telos, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    telos := &models.Telos{
        LoadedAt: time.Now(),
    }

    scanner := bufio.NewScanner(file)
    var currentSection string

    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())

        // Detect sections
        if strings.HasPrefix(line, "## ") {
            currentSection = strings.TrimPrefix(line, "## ")
            continue
        }

        // Parse content based on section
        switch currentSection {
        case "Goals":
            if goal := p.parseGoal(line); goal != nil {
                telos.Goals = append(telos.Goals, *goal)
            }
        case "Strategies":
            if strategy := p.parseStrategy(line); strategy != nil {
                telos.Strategies = append(telos.Strategies, *strategy)
            }
        case "Stack":
            p.parseStack(line, &telos.Stack)
        case "Failure Patterns":
            if pattern := p.parsePattern(line); pattern != nil {
                telos.FailurePatterns = append(telos.FailurePatterns, *pattern)
            }
        }
    }

    if err := scanner.Err(); err != nil {
        return nil, err
    }

    return telos, nil
}

// parseGoal parses a goal line
func (p *Parser) parseGoal(line string) *models.Goal {
    matches := p.goalRegex.FindStringSubmatch(line)
    if len(matches) < 2 {
        return nil
    }

    goal := &models.Goal{
        Description: matches[1],
    }

    // Parse deadline if present
    if len(matches) > 2 && matches[2] != "" {
        if deadline, err := time.Parse("2006-01-02", matches[2]); err == nil {
            goal.Deadline = &deadline
        }
    }

    return goal
}

// ... implement other parse methods
```

**Action Items:**
- [ ] Implement full parser logic
- [ ] Handle all Markdown sections
- [ ] Parse deadlines correctly
- [ ] Extract keywords from patterns
- [ ] Write comprehensive tests with fixtures
- [ ] Handle malformed files gracefully

#### Task 1.3: Scoring Engine
**Estimated Time:** 10 hours

**File:** `internal/scoring/engine.go`
```go
package scoring

import (
    "strings"
    "github.com/rayyacub/telos-idea-matrix/internal/models"
)

// Weights for scoring components
const (
    MissionAlignmentWeight = 0.40
    AntiPatternWeight      = 0.35
    StrategicFitWeight     = 0.25
)

// Engine calculates idea scores
type Engine struct {
    telos *models.Telos
}

// NewEngine creates a new scoring engine
func NewEngine(telos *models.Telos) *Engine {
    return &Engine{telos: telos}
}

// Score calculates a comprehensive score for an idea
func (e *Engine) Score(idea *models.Idea) (*models.Analysis, error) {
    analysis := &models.Analysis{
        AnalyzedAt: time.Now(),
    }

    // Calculate mission alignment
    analysis.MissionAlignment = e.calculateMissionAlignment(idea)

    // Detect anti-patterns
    patterns := e.detectPatterns(idea)
    analysis.DetectedPatterns = patterns
    analysis.AntiPatternScore = e.calculateAntiPatternScore(patterns)

    // Calculate strategic fit
    analysis.StrategicFit = e.calculateStrategicFit(idea)

    // Weighted final score (0-10 scale)
    analysis.Score = (
        analysis.MissionAlignment*MissionAlignmentWeight +
        analysis.AntiPatternScore*AntiPatternWeight +
        analysis.StrategicFit*StrategicFitWeight,
    ) * 10.0

    // Generate recommendations
    analysis.Recommendations = e.generateRecommendations(analysis)

    return analysis, nil
}

// calculateMissionAlignment scores alignment with goals
func (e *Engine) calculateMissionAlignment(idea *models.Idea) float64 {
    if len(e.telos.Goals) == 0 {
        return 0.5 // Neutral if no goals defined
    }

    var totalScore float64
    ideaText := strings.ToLower(idea.Title + " " + idea.Description)

    for _, goal := range e.telos.Goals {
        goalText := strings.ToLower(goal.Description)

        // Simple keyword overlap (can be enhanced with NLP)
        overlap := calculateTextOverlap(ideaText, goalText)

        // Bonus for matching goal keywords
        keywords := extractKeywords(goalText)
        keywordMatches := 0
        for _, keyword := range keywords {
            if strings.Contains(ideaText, strings.ToLower(keyword)) {
                keywordMatches++
            }
        }

        goalScore := (overlap*0.6) + (float64(keywordMatches)/float64(len(keywords))*0.4)
        totalScore += goalScore
    }

    return totalScore / float64(len(e.telos.Goals))
}

// detectPatterns finds anti-patterns in the idea
func (e *Engine) detectPatterns(idea *models.Idea) []models.DetectedPattern {
    var detected []models.DetectedPattern
    ideaText := strings.ToLower(idea.Title + " " + idea.Description)

    for _, pattern := range e.telos.FailurePatterns {
        matches := 0
        for _, keyword := range pattern.Keywords {
            if strings.Contains(ideaText, strings.ToLower(keyword)) {
                matches++
            }
        }

        if matches > 0 {
            confidence := float64(matches) / float64(len(pattern.Keywords))
            severity := "low"
            if confidence > 0.7 {
                severity = "high"
            } else if confidence > 0.4 {
                severity = "medium"
            }

            detected = append(detected, models.DetectedPattern{
                Name:        pattern.Name,
                Description: pattern.Description,
                Confidence:  confidence,
                Severity:    severity,
            })
        }
    }

    return detected
}

// calculateAntiPatternScore calculates score reduction from patterns
func (e *Engine) calculateAntiPatternScore(patterns []models.DetectedPattern) float64 {
    if len(patterns) == 0 {
        return 1.0 // Perfect score if no patterns
    }

    var totalPenalty float64
    for _, p := range patterns {
        severityWeight := 0.3
        if p.Severity == "medium" {
            severityWeight = 0.6
        } else if p.Severity == "high" {
            severityWeight = 1.0
        }
        totalPenalty += p.Confidence * severityWeight
    }

    // Cap penalty at 1.0
    penalty := totalPenalty / float64(len(patterns))
    if penalty > 1.0 {
        penalty = 1.0
    }

    return 1.0 - penalty
}

// calculateStrategicFit scores alignment with strategies
func (e *Engine) calculateStrategicFit(idea *models.Idea) float64 {
    if len(e.telos.Strategies) == 0 {
        return 0.5
    }

    var totalScore float64
    ideaText := strings.ToLower(idea.Title + " " + idea.Description)

    for _, strategy := range e.telos.Strategies {
        strategyText := strings.ToLower(strategy.Description)
        overlap := calculateTextOverlap(ideaText, strategyText)
        totalScore += overlap
    }

    strategicScore := totalScore / float64(len(e.telos.Strategies))

    // Check stack alignment
    stackScore := e.calculateStackAlignment(idea)

    // Combine (70% strategy, 30% stack)
    return strategicScore*0.7 + stackScore*0.3
}

// calculateStackAlignment checks if idea uses preferred stack
func (e *Engine) calculateStackAlignment(idea *models.Idea) float64 {
    ideaText := strings.ToLower(idea.Title + " " + idea.Description)

    primaryMatches := 0
    for _, tech := range e.telos.Stack.Primary {
        if strings.Contains(ideaText, strings.ToLower(tech)) {
            primaryMatches++
        }
    }

    secondaryMatches := 0
    for _, tech := range e.telos.Stack.Secondary {
        if strings.Contains(ideaText, strings.ToLower(tech)) {
            secondaryMatches++
        }
    }

    totalStack := len(e.telos.Stack.Primary) + len(e.telos.Stack.Secondary)
    if totalStack == 0 {
        return 0.5
    }

    // Primary tech weighted higher
    score := (float64(primaryMatches)*1.0 + float64(secondaryMatches)*0.5) / float64(totalStack)
    if score > 1.0 {
        score = 1.0
    }

    return score
}

// generateRecommendations creates actionable suggestions
func (e *Engine) generateRecommendations(analysis *models.Analysis) []string {
    var recs []string

    if analysis.Score < 4.0 {
        recs = append(recs, "⚠️  Low alignment with your Telos - consider if this is a distraction")
    }

    if analysis.MissionAlignment < 0.4 {
        recs = append(recs, "⚡ This idea doesn't strongly advance your stated goals")
    }

    if len(analysis.DetectedPatterns) > 0 {
        for _, p := range analysis.DetectedPatterns {
            if p.Severity == "high" {
                recs = append(recs, "🚨 High-confidence anti-pattern detected: "+p.Name)
            }
        }
    }

    if analysis.StrategicFit < 0.4 {
        recs = append(recs, "🎯 Consider how this aligns with your current strategies")
    }

    if analysis.Score >= 8.0 {
        recs = append(recs, "✅ Strong alignment - good candidate for immediate action!")
    }

    return recs
}

// Helper functions
func calculateTextOverlap(text1, text2 string) float64 {
    words1 := strings.Fields(text1)
    words2 := strings.Fields(text2)

    matches := 0
    for _, w1 := range words1 {
        for _, w2 := range words2 {
            if w1 == w2 && len(w1) > 3 { // Ignore short words
                matches++
                break
            }
        }
    }

    if len(words1) == 0 {
        return 0
    }

    return float64(matches) / float64(len(words1))
}

func extractKeywords(text string) []string {
    // Simple keyword extraction (can be enhanced)
    words := strings.Fields(text)
    var keywords []string

    // Filter out common words
    stopWords := map[string]bool{
        "the": true, "a": true, "an": true, "and": true,
        "or": true, "but": true, "in": true, "on": true,
        "at": true, "to": true, "for": true,
    }

    for _, word := range words {
        word = strings.ToLower(word)
        if len(word) > 3 && !stopWords[word] {
            keywords = append(keywords, word)
        }
    }

    return keywords
}
```

**Action Items:**
- [ ] Implement all scoring methods
- [ ] Match Rust scoring behavior exactly
- [ ] Add comprehensive unit tests
- [ ] Test with real telos.md examples
- [ ] Benchmark performance
- [ ] Document algorithm clearly

#### Task 1.4: Database Layer
**Estimated Time:** 10 hours

**File:** `internal/database/repository.go`
```go
package database

import (
    "context"
    "database/sql"
    "time"

    "github.com/google/uuid"
    _ "github.com/mattn/go-sqlite3"

    "github.com/rayyacub/telos-idea-matrix/internal/models"
)

// Repository handles database operations
type Repository struct {
    db *sql.DB
}

// NewRepository creates a new database repository
func NewRepository(dbPath string) (*Repository, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }

    if err := db.Ping(); err != nil {
        return nil, err
    }

    repo := &Repository{db: db}

    // Run migrations
    if err := repo.migrate(); err != nil {
        return nil, err
    }

    return repo, nil
}

// migrate runs database migrations
func (r *Repository) migrate() error {
    schema := `
    CREATE TABLE IF NOT EXISTS ideas (
        id TEXT PRIMARY KEY,
        title TEXT NOT NULL,
        description TEXT,
        score REAL DEFAULT 0.0,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        status TEXT DEFAULT 'pending'
    );

    CREATE TABLE IF NOT EXISTS tags (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        idea_id TEXT NOT NULL,
        tag TEXT NOT NULL,
        FOREIGN KEY (idea_id) REFERENCES ideas(id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS analyses (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        idea_id TEXT NOT NULL,
        score REAL NOT NULL,
        mission_alignment REAL,
        anti_pattern_score REAL,
        strategic_fit REAL,
        analyzed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (idea_id) REFERENCES ideas(id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS detected_patterns (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        analysis_id INTEGER NOT NULL,
        name TEXT NOT NULL,
        description TEXT,
        confidence REAL,
        severity TEXT,
        FOREIGN KEY (analysis_id) REFERENCES analyses(id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS idea_links (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        source_id TEXT NOT NULL,
        target_id TEXT NOT NULL,
        link_type TEXT NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (source_id) REFERENCES ideas(id) ON DELETE CASCADE,
        FOREIGN KEY (target_id) REFERENCES ideas(id) ON DELETE CASCADE
    );

    CREATE INDEX IF NOT EXISTS idx_ideas_score ON ideas(score);
    CREATE INDEX IF NOT EXISTS idx_ideas_created_at ON ideas(created_at);
    CREATE INDEX IF NOT EXISTS idx_ideas_status ON ideas(status);
    CREATE INDEX IF NOT EXISTS idx_tags_idea_id ON tags(idea_id);
    `

    _, err := r.db.Exec(schema)
    return err
}

// CreateIdea inserts a new idea
func (r *Repository) CreateIdea(ctx context.Context, idea *models.Idea) error {
    if idea.ID == uuid.Nil {
        idea.ID = uuid.New()
    }

    now := time.Now()
    idea.CreatedAt = now
    idea.UpdatedAt = now

    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // Insert idea
    _, err = tx.ExecContext(ctx, `
        INSERT INTO ideas (id, title, description, score, created_at, updated_at, status)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    `, idea.ID.String(), idea.Title, idea.Description, idea.Score,
        idea.CreatedAt, idea.UpdatedAt, idea.Status)
    if err != nil {
        return err
    }

    // Insert tags
    for _, tag := range idea.Tags {
        _, err = tx.ExecContext(ctx, `
            INSERT INTO tags (idea_id, tag) VALUES (?, ?)
        `, idea.ID.String(), tag)
        if err != nil {
            return err
        }
    }

    // Insert analysis if present
    if idea.Analysis != nil {
        if err := r.saveAnalysis(ctx, tx, idea.ID, idea.Analysis); err != nil {
            return err
        }
    }

    return tx.Commit()
}

// GetIdea retrieves an idea by ID
func (r *Repository) GetIdea(ctx context.Context, id uuid.UUID) (*models.Idea, error) {
    var idea models.Idea

    err := r.db.QueryRowContext(ctx, `
        SELECT id, title, description, score, created_at, updated_at, status
        FROM ideas WHERE id = ?
    `, id.String()).Scan(
        &idea.ID, &idea.Title, &idea.Description, &idea.Score,
        &idea.CreatedAt, &idea.UpdatedAt, &idea.Status,
    )
    if err != nil {
        return nil, err
    }

    // Load tags
    tags, err := r.getTags(ctx, id)
    if err != nil {
        return nil, err
    }
    idea.Tags = tags

    // Load latest analysis
    analysis, err := r.getLatestAnalysis(ctx, id)
    if err != nil && err != sql.ErrNoRows {
        return nil, err
    }
    idea.Analysis = analysis

    return &idea, nil
}

// ListIdeas retrieves ideas with optional filtering
func (r *Repository) ListIdeas(ctx context.Context, opts ListOptions) ([]*models.Idea, error) {
    query := `
        SELECT id, title, description, score, created_at, updated_at, status
        FROM ideas
        WHERE 1=1
    `
    args := []interface{}{}

    if opts.MinScore > 0 {
        query += " AND score >= ?"
        args = append(args, opts.MinScore)
    }

    if opts.Status != "" {
        query += " AND status = ?"
        args = append(args, opts.Status)
    }

    query += " ORDER BY created_at DESC LIMIT ?"
    args = append(args, opts.Limit)

    rows, err := r.db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var ideas []*models.Idea
    for rows.Next() {
        var idea models.Idea
        err := rows.Scan(
            &idea.ID, &idea.Title, &idea.Description, &idea.Score,
            &idea.CreatedAt, &idea.UpdatedAt, &idea.Status,
        )
        if err != nil {
            return nil, err
        }

        // Load tags
        tags, err := r.getTags(ctx, idea.ID)
        if err != nil {
            return nil, err
        }
        idea.Tags = tags

        ideas = append(ideas, &idea)
    }

    return ideas, rows.Err()
}

// UpdateIdea updates an idea
func (r *Repository) UpdateIdea(ctx context.Context, idea *models.Idea) error {
    idea.UpdatedAt = time.Now()

    _, err := r.db.ExecContext(ctx, `
        UPDATE ideas
        SET title = ?, description = ?, score = ?, updated_at = ?, status = ?
        WHERE id = ?
    `, idea.Title, idea.Description, idea.Score, idea.UpdatedAt, idea.Status, idea.ID.String())

    return err
}

// DeleteIdea soft-deletes an idea
func (r *Repository) DeleteIdea(ctx context.Context, id uuid.UUID) error {
    _, err := r.db.ExecContext(ctx, "DELETE FROM ideas WHERE id = ?", id.String())
    return err
}

// Helper methods
func (r *Repository) getTags(ctx context.Context, ideaID uuid.UUID) ([]string, error) {
    rows, err := r.db.QueryContext(ctx, "SELECT tag FROM tags WHERE idea_id = ?", ideaID.String())
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var tags []string
    for rows.Next() {
        var tag string
        if err := rows.Scan(&tag); err != nil {
            return nil, err
        }
        tags = append(tags, tag)
    }

    return tags, rows.Err()
}

func (r *Repository) saveAnalysis(ctx context.Context, tx *sql.Tx, ideaID uuid.UUID, analysis *models.Analysis) error {
    result, err := tx.ExecContext(ctx, `
        INSERT INTO analyses (idea_id, score, mission_alignment, anti_pattern_score, strategic_fit, analyzed_at)
        VALUES (?, ?, ?, ?, ?, ?)
    `, ideaID.String(), analysis.Score, analysis.MissionAlignment,
        analysis.AntiPatternScore, analysis.StrategicFit, analysis.AnalyzedAt)
    if err != nil {
        return err
    }

    analysisID, err := result.LastInsertId()
    if err != nil {
        return err
    }

    // Save detected patterns
    for _, pattern := range analysis.DetectedPatterns {
        _, err = tx.ExecContext(ctx, `
            INSERT INTO detected_patterns (analysis_id, name, description, confidence, severity)
            VALUES (?, ?, ?, ?, ?)
        `, analysisID, pattern.Name, pattern.Description, pattern.Confidence, pattern.Severity)
        if err != nil {
            return err
        }
    }

    return nil
}

func (r *Repository) getLatestAnalysis(ctx context.Context, ideaID uuid.UUID) (*models.Analysis, error) {
    var analysis models.Analysis
    var analysisID int64

    err := r.db.QueryRowContext(ctx, `
        SELECT id, score, mission_alignment, anti_pattern_score, strategic_fit, analyzed_at
        FROM analyses
        WHERE idea_id = ?
        ORDER BY analyzed_at DESC
        LIMIT 1
    `, ideaID.String()).Scan(
        &analysisID, &analysis.Score, &analysis.MissionAlignment,
        &analysis.AntiPatternScore, &analysis.StrategicFit, &analysis.AnalyzedAt,
    )
    if err != nil {
        return nil, err
    }

    // Load patterns
    patterns, err := r.getPatterns(ctx, analysisID)
    if err != nil {
        return nil, err
    }
    analysis.DetectedPatterns = patterns

    return &analysis, nil
}

func (r *Repository) getPatterns(ctx context.Context, analysisID int64) ([]models.DetectedPattern, error) {
    rows, err := r.db.QueryContext(ctx, `
        SELECT name, description, confidence, severity
        FROM detected_patterns
        WHERE analysis_id = ?
    `, analysisID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var patterns []models.DetectedPattern
    for rows.Next() {
        var p models.DetectedPattern
        if err := rows.Scan(&p.Name, &p.Description, &p.Confidence, &p.Severity); err != nil {
            return nil, err
        }
        patterns = append(patterns, p)
    }

    return patterns, rows.Err()
}

// Close closes the database connection
func (r *Repository) Close() error {
    return r.db.Close()
}

// ListOptions for filtering ideas
type ListOptions struct {
    Limit    int
    MinScore float64
    Status   string
}
```

**Action Items:**
- [ ] Implement all CRUD operations
- [ ] Add transaction support
- [ ] Implement connection pooling
- [ ] Add database health check
- [ ] Write integration tests
- [ ] Add query optimization

---

### Phase 2: CLI Implementation

#### Task 2.1: Cobra CLI Setup
**Estimated Time:** 4 hours

**File:** `cmd/cli/main.go`
```go
package main

import (
    "fmt"
    "os"

    "github.com/rayyacub/telos-idea-matrix/internal/cli"
)

func main() {
    if err := cli.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

**File:** `internal/cli/root.go`
```go
package cli

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var (
    cfgFile string
    noAI    bool
)

var rootCmd = &cobra.Command{
    Use:   "tm",
    Short: "Telos Idea Matrix - Idea capture + Telos-aligned analysis",
    Long: `Telos Idea Matrix helps you escape decision paralysis by providing
instant, objective analysis of your ideas against your personal Telos.`,
    Version: "2.0.0",
}

func Execute() error {
    return rootCmd.Execute()
}

func init() {
    cobra.OnInitialize(initConfig)

    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.telos-matrix/config.yaml)")
    rootCmd.PersistentFlags().BoolVar(&noAI, "no-ai", false, "disable AI analysis, use rule-based only")
}

func initConfig() {
    if cfgFile != "" {
        viper.SetConfigFile(cfgFile)
    } else {
        home, err := os.UserHomeDir()
        if err != nil {
            fmt.Fprintln(os.Stderr, err)
            os.Exit(1)
        }

        viper.AddConfigPath(home + "/.telos-matrix")
        viper.SetConfigName("config")
        viper.SetConfigType("yaml")
    }

    viper.AutomaticEnv()
    viper.ReadInConfig()
}
```

**Action Items:**
- [ ] Set up Cobra command structure
- [ ] Configure Viper for config management
- [ ] Add global flags
- [ ] Set up command registration
- [ ] Add version command
- [ ] Add help text

#### Task 2.2: Implement Commands
**Estimated Time:** 12 hours

**Commands to implement:**
- [ ] `tm dump` - Capture idea
- [ ] `tm analyze` - Analyze idea
- [ ] `tm score` - Quick score
- [ ] `tm review` - Browse ideas
- [ ] `tm prune` - Clean old ideas
- [ ] `tm analytics` - View stats
- [ ] `tm link` - Manage relationships

**Example:** `internal/cli/dump.go`
```go
package cli

import (
    "context"
    "fmt"
    "strings"

    "github.com/spf13/cobra"

    "github.com/rayyacub/telos-idea-matrix/internal/config"
    "github.com/rayyacub/telos-idea-matrix/internal/database"
    "github.com/rayyacub/telos-idea-matrix/internal/models"
    "github.com/rayyacub/telos-idea-matrix/internal/scoring"
    "github.com/rayyacub/telos-idea-matrix/internal/telos"
)

var dumpCmd = &cobra.Command{
    Use:   "dump [idea text]",
    Short: "Capture an idea and get immediate analysis",
    Args:  cobra.MinimumNArgs(0),
    RunE:  runDump,
}

var (
    interactive bool
    quick       bool
)

func init() {
    rootCmd.AddCommand(dumpCmd)
    dumpCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "open editor for multi-line input")
    dumpCmd.Flags().BoolVarP(&quick, "quick", "q", false, "save without analysis")
}

func runDump(cmd *cobra.Command, args []string) error {
    ctx := context.Background()

    // Get idea text
    var ideaText string
    if len(args) > 0 {
        ideaText = strings.Join(args, " ")
    } else if interactive {
        // Open editor (implement with $EDITOR)
        ideaText = openEditor()
    } else {
        // Prompt for input
        fmt.Print("Enter your idea: ")
        fmt.Scanln(&ideaText)
    }

    if ideaText == "" {
        return fmt.Errorf("idea text is required")
    }

    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }

    // Parse telos
    parser := telos.NewParser()
    telosData, err := parser.ParseFile(cfg.TelosFile)
    if err != nil {
        return fmt.Errorf("failed to parse telos: %w", err)
    }

    // Create idea
    idea := &models.Idea{
        Title:  ideaText,
        Status: "pending",
    }

    // Score if not quick mode
    if !quick && !noAI {
        engine := scoring.NewEngine(telosData)
        analysis, err := engine.Score(idea)
        if err != nil {
            return fmt.Errorf("scoring failed: %w", err)
        }
        idea.Score = analysis.Score
        idea.Analysis = analysis
    }

    // Save to database
    repo, err := database.NewRepository(cfg.DatabasePath)
    if err != nil {
        return fmt.Errorf("database error: %w", err)
    }
    defer repo.Close()

    if err := repo.CreateIdea(ctx, idea); err != nil {
        return fmt.Errorf("failed to save idea: %w", err)
    }

    // Display result
    displayIdeaAnalysis(idea)

    return nil
}

func displayIdeaAnalysis(idea *models.Idea) {
    fmt.Println("\n✨ Idea Captured!")
    fmt.Printf("ID: %s\n", idea.ID)
    fmt.Printf("Title: %s\n", idea.Title)

    if idea.Analysis != nil {
        fmt.Printf("\n📊 Score: %.1f/10\n", idea.Analysis.Score)
        fmt.Printf("├─ Mission Alignment: %.1f%%\n", idea.Analysis.MissionAlignment*100)
        fmt.Printf("├─ Anti-Pattern Score: %.1f%%\n", idea.Analysis.AntiPatternScore*100)
        fmt.Printf("└─ Strategic Fit: %.1f%%\n", idea.Analysis.StrategicFit*100)

        if len(idea.Analysis.DetectedPatterns) > 0 {
            fmt.Println("\n⚠️  Detected Patterns:")
            for _, p := range idea.Analysis.DetectedPatterns {
                fmt.Printf("  • %s (%s confidence, %s severity)\n", p.Name, formatConfidence(p.Confidence), p.Severity)
            }
        }

        if len(idea.Analysis.Recommendations) > 0 {
            fmt.Println("\n💡 Recommendations:")
            for _, rec := range idea.Analysis.Recommendations {
                fmt.Printf("  %s\n", rec)
            }
        }
    }

    fmt.Println()
}

func formatConfidence(conf float64) string {
    if conf > 0.7 {
        return "high"
    } else if conf > 0.4 {
        return "medium"
    }
    return "low"
}

func openEditor() string {
    // Implement editor opening logic
    // Use os/exec to open $EDITOR
    return ""
}
```

**Action Items:**
- [ ] Implement all command handlers
- [ ] Add proper flag handling
- [ ] Implement interactive prompts (dialoguer)
- [ ] Add colored output (fatih/color)
- [ ] Add progress indicators
- [ ] Write tests for each command

---

### Phase 3: API Server

#### Task 3.1: API Server Setup
**Estimated Time:** 6 hours

**File:** `cmd/web/main.go`
```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "time"

    "github.com/rayyacub/telos-idea-matrix/internal/api"
    "github.com/rayyacub/telos-idea-matrix/internal/config"
    "github.com/rayyacub/telos-idea-matrix/internal/database"
)

func main() {
    // Load config
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Initialize database
    repo, err := database.NewRepository(cfg.DatabasePath)
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer repo.Close()

    // Create API server
    server := api.NewServer(repo, cfg)

    // HTTP server
    httpServer := &http.Server{
        Addr:         fmt.Sprintf(":%d", cfg.Port),
        Handler:      server.Router(),
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // Start server in goroutine
    go func() {
        log.Printf("🚀 Server starting on %s", httpServer.Addr)
        if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server failed: %v", err)
        }
    }()

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt)
    <-quit

    log.Println("🛑 Shutting down server...")
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := httpServer.Shutdown(ctx); err != nil {
        log.Fatalf("Server shutdown failed: %v", err)
    }

    log.Println("✅ Server stopped")
}
```

**File:** `internal/api/server.go`
```go
package api

import (
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/go-chi/cors"

    "github.com/rayyacub/telos-idea-matrix/internal/config"
    "github.com/rayyacub/telos-idea-matrix/internal/database"
)

// Server represents the API server
type Server struct {
    repo   *database.Repository
    config *config.Config
}

// NewServer creates a new API server
func NewServer(repo *database.Repository, cfg *config.Config) *Server {
    return &Server{
        repo:   repo,
        config: cfg,
    }
}

// Router creates the HTTP router
func (s *Server) Router() http.Handler {
    r := chi.NewRouter()

    // Middleware
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.Timeout(60 * time.Second))

    // CORS
    r.Use(cors.Handler(cors.Options{
        AllowedOrigins:   []string{"http://localhost:5173"}, // SvelteKit dev server
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
        AllowCredentials: true,
    }))

    // Health check
    r.Get("/health", s.handleHealth)

    // API routes
    r.Route("/api/v1", func(r chi.Router) {
        // Ideas
        r.Get("/ideas", s.handleListIdeas)
        r.Post("/ideas", s.handleCreateIdea)
        r.Get("/ideas/{id}", s.handleGetIdea)
        r.Put("/ideas/{id}", s.handleUpdateIdea)
        r.Delete("/ideas/{id}", s.handleDeleteIdea)

        // Analysis
        r.Post("/analyze", s.handleAnalyze)

        // Analytics
        r.Get("/analytics/stats", s.handleStats)
        r.Get("/analytics/trends", s.handleTrends)

        // Links
        r.Get("/ideas/{id}/links", s.handleGetLinks)
        r.Post("/links", s.handleCreateLink)
        r.Delete("/links/{id}", s.handleDeleteLink)
    })

    // Serve SvelteKit static files (production)
    // In development, SvelteKit runs on its own port
    if s.config.Env == "production" {
        r.Handle("/*", http.FileServer(http.Dir("./web/build")))
    }

    return r
}

// Health check handler
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
    respondJSON(w, http.StatusOK, map[string]string{
        "status": "healthy",
        "version": "2.0.0",
    })
}
```

**Action Items:**
- [ ] Set up Chi router
- [ ] Add middleware (logging, CORS, recovery)
- [ ] Implement graceful shutdown
- [ ] Add health check endpoint
- [ ] Configure CORS for SvelteKit
- [ ] Add request validation

#### Task 3.2: API Handlers
**Estimated Time:** 12 hours

**File:** `internal/api/handlers.go`
```go
package api

import (
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/google/uuid"

    "github.com/rayyacub/telos-idea-matrix/internal/database"
    "github.com/rayyacub/telos-idea-matrix/internal/models"
)

// handleListIdeas lists ideas with optional filtering
func (s *Server) handleListIdeas(w http.ResponseWriter, r *http.Request) {
    // Parse query params
    opts := database.ListOptions{
        Limit:    10,
        MinScore: parseFloat(r.URL.Query().Get("min_score"), 0),
        Status:   r.URL.Query().Get("status"),
    }

    if limit := r.URL.Query().Get("limit"); limit != "" {
        opts.Limit = parseInt(limit, 10)
    }

    // Query database
    ideas, err := s.repo.ListIdeas(r.Context(), opts)
    if err != nil {
        respondError(w, http.StatusInternalServerError, "Failed to fetch ideas")
        return
    }

    respondJSON(w, http.StatusOK, ideas)
}

// handleCreateIdea creates a new idea
func (s *Server) handleCreateIdea(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Title       string   `json:"title"`
        Description string   `json:"description"`
        Tags        []string `json:"tags"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    // Validate
    if req.Title == "" {
        respondError(w, http.StatusBadRequest, "Title is required")
        return
    }

    idea := &models.Idea{
        Title:       req.Title,
        Description: req.Description,
        Tags:        req.Tags,
        Status:      "pending",
    }

    // Score the idea
    if err := s.scoreIdea(idea); err != nil {
        respondError(w, http.StatusInternalServerError, "Scoring failed")
        return
    }

    // Save to database
    if err := s.repo.CreateIdea(r.Context(), idea); err != nil {
        respondError(w, http.StatusInternalServerError, "Failed to create idea")
        return
    }

    respondJSON(w, http.StatusCreated, idea)
}

// handleGetIdea retrieves a single idea
func (s *Server) handleGetIdea(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        respondError(w, http.StatusBadRequest, "Invalid ID")
        return
    }

    idea, err := s.repo.GetIdea(r.Context(), id)
    if err != nil {
        respondError(w, http.StatusNotFound, "Idea not found")
        return
    }

    respondJSON(w, http.StatusOK, idea)
}

// handleUpdateIdea updates an existing idea
func (s *Server) handleUpdateIdea(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        respondError(w, http.StatusBadRequest, "Invalid ID")
        return
    }

    var req struct {
        Title       string `json:"title"`
        Description string `json:"description"`
        Status      string `json:"status"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    // Fetch existing idea
    idea, err := s.repo.GetIdea(r.Context(), id)
    if err != nil {
        respondError(w, http.StatusNotFound, "Idea not found")
        return
    }

    // Update fields
    if req.Title != "" {
        idea.Title = req.Title
    }
    if req.Description != "" {
        idea.Description = req.Description
    }
    if req.Status != "" {
        idea.Status = req.Status
    }

    // Re-score if content changed
    if req.Title != "" || req.Description != "" {
        if err := s.scoreIdea(idea); err != nil {
            respondError(w, http.StatusInternalServerError, "Scoring failed")
            return
        }
    }

    // Update in database
    if err := s.repo.UpdateIdea(r.Context(), idea); err != nil {
        respondError(w, http.StatusInternalServerError, "Failed to update idea")
        return
    }

    respondJSON(w, http.StatusOK, idea)
}

// handleDeleteIdea deletes an idea
func (s *Server) handleDeleteIdea(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        respondError(w, http.StatusBadRequest, "Invalid ID")
        return
    }

    if err := s.repo.DeleteIdea(r.Context(), id); err != nil {
        respondError(w, http.StatusInternalServerError, "Failed to delete idea")
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

// handleAnalyze analyzes idea text without saving
func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Text string `json:"text"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    if req.Text == "" {
        respondError(w, http.StatusBadRequest, "Text is required")
        return
    }

    idea := &models.Idea{
        Title: req.Text,
    }

    if err := s.scoreIdea(idea); err != nil {
        respondError(w, http.StatusInternalServerError, "Analysis failed")
        return
    }

    respondJSON(w, http.StatusOK, idea.Analysis)
}

// Helper: score an idea
func (s *Server) scoreIdea(idea *models.Idea) error {
    // Load telos
    parser := telos.NewParser()
    telosData, err := parser.ParseFile(s.config.TelosFile)
    if err != nil {
        return err
    }

    // Score
    engine := scoring.NewEngine(telosData)
    analysis, err := engine.Score(idea)
    if err != nil {
        return err
    }

    idea.Score = analysis.Score
    idea.Analysis = analysis

    return nil
}

// Helper: respond with JSON
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

// Helper: respond with error
func respondError(w http.ResponseWriter, status int, message string) {
    respondJSON(w, status, map[string]string{"error": message})
}

// Helper: parse float from string
func parseFloat(s string, def float64) float64 {
    if s == "" {
        return def
    }
    var f float64
    fmt.Sscanf(s, "%f", &f)
    return f
}

// Helper: parse int from string
func parseInt(s string, def int) int {
    if s == "" {
        return def
    }
    var i int
    fmt.Sscanf(s, "%d", &i)
    return i
}
```

**Action Items:**
- [ ] Implement all CRUD handlers
- [ ] Add request validation
- [ ] Add error handling
- [ ] Implement analytics endpoints
- [ ] Add pagination support
- [ ] Write integration tests
- [ ] Add OpenAPI documentation

#### Task 3.3: OpenAPI Documentation
**Estimated Time:** 4 hours

Generate OpenAPI spec using `swaggo/swag` or write manually:

**File:** `docs/api.yaml`
```yaml
openapi: 3.0.0
info:
  title: Telos Idea Matrix API
  version: 2.0.0
  description: API for capturing and analyzing ideas against personal Telos

servers:
  - url: http://localhost:8080/api/v1
    description: Development server

paths:
  /ideas:
    get:
      summary: List ideas
      parameters:
        - name: limit
          in: query
          schema:
            type: integer
            default: 10
        - name: min_score
          in: query
          schema:
            type: number
        - name: status
          in: query
          schema:
            type: string
            enum: [pending, in-progress, completed, archived]
      responses:
        '200':
          description: List of ideas
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Idea'

    post:
      summary: Create idea
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [title]
              properties:
                title:
                  type: string
                description:
                  type: string
                tags:
                  type: array
                  items:
                    type: string
      responses:
        '201':
          description: Created idea
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Idea'

components:
  schemas:
    Idea:
      type: object
      properties:
        id:
          type: string
          format: uuid
        title:
          type: string
        description:
          type: string
        score:
          type: number
          format: double
        created_at:
          type: string
          format: date-time
        updated_at:
          type: string
          format: date-time
        status:
          type: string
          enum: [pending, in-progress, completed, archived]
        tags:
          type: array
          items:
            type: string
        analysis:
          $ref: '#/components/schemas/Analysis'

    Analysis:
      type: object
      properties:
        score:
          type: number
        mission_alignment:
          type: number
        anti_pattern_score:
          type: number
        strategic_fit:
          type: number
        detected_patterns:
          type: array
          items:
            $ref: '#/components/schemas/DetectedPattern'
        recommendations:
          type: array
          items:
            type: string
        analyzed_at:
          type: string
          format: date-time

    DetectedPattern:
      type: object
      properties:
        name:
          type: string
        description:
          type: string
        confidence:
          type: number
        severity:
          type: string
          enum: [low, medium, high]
```

**Action Items:**
- [ ] Generate OpenAPI spec
- [ ] Set up Swagger UI
- [ ] Document all endpoints
- [ ] Add request/response examples
- [ ] Generate TypeScript types from spec

---

### Phase 4: SvelteKit Frontend

#### Task 4.1: SvelteKit Setup
**Estimated Time:** 4 hours

```bash
# Initialize SvelteKit project
cd web
npm create svelte@latest .

# Install dependencies
npm install
npm install -D tailwindcss autoprefixer postcss
npm install -D @skeletonlabs/skeleton
npm install lucide-svelte
npm install @tanstack/svelte-query  # For API calls
npm install chart.js
```

**File:** `web/svelte.config.js`
```javascript
import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

export default {
    preprocess: vitePreprocess(),
    kit: {
        adapter: adapter({
            pages: 'build',
            assets: 'build',
            fallback: 'index.html'
        }),
        alias: {
            '$lib': 'src/lib',
            '$components': 'src/lib/components'
        }
    }
};
```

**File:** `web/tailwind.config.js`
```javascript
import { skeleton } from '@skeletonlabs/skeleton/plugin';
import * as themes from '@skeletonlabs/skeleton/themes';

export default {
    content: [
        './src/**/*.{html,js,svelte,ts}',
        './node_modules/@skeletonlabs/skeleton/**/*.{html,js,svelte,ts}'
    ],
    theme: {
        extend: {}
    },
    plugins: [
        skeleton({
            themes: [themes.cerberus]
        })
    ]
};
```

**File:** `web/src/app.css`
```css
@tailwind base;
@tailwind components;
@tailwind utilities;

:root {
    --color-primary: 136 58 234;
    --color-secondary: 76 29 149;
}
```

**Action Items:**
- [ ] Initialize SvelteKit project
- [ ] Configure Tailwind CSS
- [ ] Install Skeleton UI
- [ ] Set up TypeScript
- [ ] Configure Vite
- [ ] Set up project structure

#### Task 4.2: API Client
**Estimated Time:** 6 hours

**File:** `web/src/lib/api/client.ts`
```typescript
import type { Idea, Analysis, CreateIdeaRequest, UpdateIdeaRequest } from './types';

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

class ApiError extends Error {
    constructor(public status: number, message: string) {
        super(message);
    }
}

async function fetchJSON<T>(url: string, options?: RequestInit): Promise<T> {
    const response = await fetch(`${API_BASE}${url}`, {
        ...options,
        headers: {
            'Content-Type': 'application/json',
            ...options?.headers
        }
    });

    if (!response.ok) {
        const error = await response.json().catch(() => ({ error: 'Unknown error' }));
        throw new ApiError(response.status, error.error || 'Request failed');
    }

    return response.json();
}

export const api = {
    // Ideas
    async listIdeas(params?: { limit?: number; minScore?: number; status?: string }): Promise<Idea[]> {
        const query = new URLSearchParams();
        if (params?.limit) query.set('limit', params.limit.toString());
        if (params?.minScore) query.set('min_score', params.minScore.toString());
        if (params?.status) query.set('status', params.status);

        return fetchJSON(`/ideas?${query}`);
    },

    async getIdea(id: string): Promise<Idea> {
        return fetchJSON(`/ideas/${id}`);
    },

    async createIdea(data: CreateIdeaRequest): Promise<Idea> {
        return fetchJSON('/ideas', {
            method: 'POST',
            body: JSON.stringify(data)
        });
    },

    async updateIdea(id: string, data: UpdateIdeaRequest): Promise<Idea> {
        return fetchJSON(`/ideas/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data)
        });
    },

    async deleteIdea(id: string): Promise<void> {
        await fetch(`${API_BASE}/ideas/${id}`, { method: 'DELETE' });
    },

    // Analysis
    async analyze(text: string): Promise<Analysis> {
        return fetchJSON('/analyze', {
            method: 'POST',
            body: JSON.stringify({ text })
        });
    },

    // Analytics
    async getStats(): Promise<any> {
        return fetchJSON('/analytics/stats');
    },

    async getTrends(): Promise<any> {
        return fetchJSON('/analytics/trends');
    }
};
```

**File:** `web/src/lib/api/types.ts`
```typescript
export interface Idea {
    id: string;
    title: string;
    description?: string;
    score: number;
    created_at: string;
    updated_at: string;
    status: 'pending' | 'in-progress' | 'completed' | 'archived';
    tags: string[];
    analysis?: Analysis;
}

export interface Analysis {
    score: number;
    mission_alignment: number;
    anti_pattern_score: number;
    strategic_fit: number;
    detected_patterns: DetectedPattern[];
    recommendations: string[];
    analyzed_at: string;
}

export interface DetectedPattern {
    name: string;
    description: string;
    confidence: number;
    severity: 'low' | 'medium' | 'high';
}

export interface CreateIdeaRequest {
    title: string;
    description?: string;
    tags?: string[];
}

export interface UpdateIdeaRequest {
    title?: string;
    description?: string;
    status?: string;
}
```

**Action Items:**
- [ ] Implement API client
- [ ] Add TypeScript types
- [ ] Add error handling
- [ ] Add loading states
- [ ] Set up React Query for caching
- [ ] Write tests

#### Task 4.3: UI Components
**Estimated Time:** 16 hours

**File:** `web/src/lib/components/IdeaCard.svelte`
```svelte
<script lang="ts">
    import { fade } from 'svelte/transition';
    import { Trash2, Edit, BarChart } from 'lucide-svelte';
    import type { Idea } from '$lib/api/types';

    export let idea: Idea;
    export let onDelete: (id: string) => void;
    export let onEdit: (idea: Idea) => void;

    function getScoreColor(score: number): string {
        if (score >= 8) return 'variant-filled-success';
        if (score >= 6) return 'variant-filled-warning';
        return 'variant-filled-error';
    }

    function getSeverityColor(severity: string): string {
        switch (severity) {
            case 'high': return 'variant-filled-error';
            case 'medium': return 'variant-filled-warning';
            default: return 'variant-filled-surface';
        }
    }
</script>

<div
    class="card p-6 hover:scale-105 transition-transform"
    in:fade={{ duration: 300 }}
>
    <header class="flex justify-between items-start mb-4">
        <div class="flex-1">
            <h3 class="h3 mb-2">{idea.title}</h3>
            {#if idea.description}
                <p class="text-surface-600 dark:text-surface-400">
                    {idea.description}
                </p>
            {/if}
        </div>
        <span class="badge {getScoreColor(idea.score)} text-lg font-bold">
            {idea.score.toFixed(1)}
        </span>
    </header>

    {#if idea.analysis}
        <div class="space-y-2 mb-4">
            <div class="flex items-center gap-2">
                <span class="text-sm">Mission:</span>
                <div class="w-full bg-surface-200 rounded-full h-2">
                    <div
                        class="bg-primary-500 h-2 rounded-full"
                        style="width: {idea.analysis.mission_alignment * 100}%"
                    />
                </div>
                <span class="text-sm">{(idea.analysis.mission_alignment * 100).toFixed(0)}%</span>
            </div>

            <div class="flex items-center gap-2">
                <span class="text-sm">Strategic:</span>
                <div class="w-full bg-surface-200 rounded-full h-2">
                    <div
                        class="bg-secondary-500 h-2 rounded-full"
                        style="width: {idea.analysis.strategic_fit * 100}%"
                    />
                </div>
                <span class="text-sm">{(idea.analysis.strategic_fit * 100).toFixed(0)}%</span>
            </div>
        </div>

        {#if idea.analysis.detected_patterns.length > 0}
            <div class="mb-4">
                <p class="text-sm font-semibold mb-2">⚠️ Detected Patterns:</p>
                <div class="flex flex-wrap gap-2">
                    {#each idea.analysis.detected_patterns as pattern}
                        <span class="chip {getSeverityColor(pattern.severity)} text-xs">
                            {pattern.name}
                        </span>
                    {/each}
                </div>
            </div>
        {/if}
    {/if}

    <footer class="flex gap-2">
        <button
            class="btn variant-ghost-primary btn-sm"
            on:click={() => onEdit(idea)}
        >
            <Edit size={16} />
            Edit
        </button>
        <button
            class="btn variant-ghost-error btn-sm"
            on:click={() => onDelete(idea.id)}
        >
            <Trash2 size={16} />
            Delete
        </button>
        <a href="/ideas/{idea.id}" class="btn variant-ghost-secondary btn-sm">
            <BarChart size={16} />
            Details
        </a>
    </footer>
</div>
```

**File:** `web/src/lib/components/IdeaForm.svelte`
```svelte
<script lang="ts">
    import { createEventDispatcher } from 'svelte';
    import type { CreateIdeaRequest } from '$lib/api/types';

    const dispatch = createEventDispatcher();

    let title = '';
    let description = '';
    let tags = '';
    let loading = false;

    async function handleSubmit() {
        loading = true;

        const data: CreateIdeaRequest = {
            title,
            description: description || undefined,
            tags: tags ? tags.split(',').map(t => t.trim()) : []
        };

        dispatch('submit', data);

        // Reset form
        title = '';
        description = '';
        tags = '';
        loading = false;
    }
</script>

<form on:submit|preventDefault={handleSubmit} class="card p-6 space-y-4">
    <h2 class="h2">Capture New Idea</h2>

    <label class="label">
        <span>Title *</span>
        <input
            class="input"
            type="text"
            bind:value={title}
            placeholder="What's your idea?"
            required
        />
    </label>

    <label class="label">
        <span>Description</span>
        <textarea
            class="textarea"
            bind:value={description}
            placeholder="Add more details..."
            rows="4"
        />
    </label>

    <label class="label">
        <span>Tags</span>
        <input
            class="input"
            type="text"
            bind:value={tags}
            placeholder="rust, cli, productivity (comma-separated)"
        />
    </label>

    <button
        class="btn variant-filled-primary w-full"
        type="submit"
        disabled={loading || !title}
    >
        {loading ? 'Analyzing...' : 'Capture Idea'}
    </button>
</form>
```

**Action Items:**
- [ ] Build IdeaCard component
- [ ] Build IdeaForm component
- [ ] Build IdeaDetail component
- [ ] Build Dashboard component
- [ ] Build FilterBar component
- [ ] Build ScoreChart component
- [ ] Build PatternBadge component
- [ ] Make all components responsive

#### Task 4.4: Pages/Routes
**Estimated Time:** 12 hours

**File:** `web/src/routes/+page.svelte`
```svelte
<script lang="ts">
    import { onMount } from 'svelte';
    import { api } from '$lib/api/client';
    import IdeaCard from '$components/IdeaCard.svelte';
    import IdeaForm from '$components/IdeaForm.svelte';
    import type { Idea, CreateIdeaRequest } from '$lib/api/types';

    let ideas: Idea[] = [];
    let loading = true;
    let error: string | null = null;
    let minScore = 0;
    let status = '';

    onMount(async () => {
        await loadIdeas();
    });

    async function loadIdeas() {
        loading = true;
        error = null;

        try {
            ideas = await api.listIdeas({
                limit: 20,
                minScore: minScore > 0 ? minScore : undefined,
                status: status || undefined
            });
        } catch (e) {
            error = e instanceof Error ? e.message : 'Failed to load ideas';
        } finally {
            loading = false;
        }
    }

    async function handleCreateIdea(event: CustomEvent<CreateIdeaRequest>) {
        try {
            await api.createIdea(event.detail);
            await loadIdeas();
        } catch (e) {
            error = e instanceof Error ? e.message : 'Failed to create idea';
        }
    }

    async function handleDeleteIdea(id: string) {
        if (!confirm('Are you sure you want to delete this idea?')) return;

        try {
            await api.deleteIdea(id);
            await loadIdeas();
        } catch (e) {
            error = e instanceof Error ? e.message : 'Failed to delete idea';
        }
    }

    function handleEditIdea(idea: Idea) {
        // Navigate to edit page or open modal
        window.location.href = `/ideas/${idea.id}/edit`;
    }
</script>

<svelte:head>
    <title>Telos Idea Matrix</title>
</svelte:head>

<div class="container mx-auto p-8">
    <header class="mb-12">
        <h1 class="h1 gradient-heading mb-4">
            <span class="gradient-text">Telos Idea Matrix</span>
        </h1>
        <p class="text-xl text-surface-600 dark:text-surface-400">
            Stop drowning in ideas. Start shipping the ones that matter.
        </p>
    </header>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-8 mb-8">
        <div class="lg:col-span-2">
            <!-- Filters -->
            <div class="card p-4 mb-6">
                <div class="flex gap-4">
                    <label class="label flex-1">
                        <span>Min Score</span>
                        <input
                            type="range"
                            min="0"
                            max="10"
                            step="0.5"
                            bind:value={minScore}
                            on:change={loadIdeas}
                            class="range"
                        />
                        <span class="text-sm">{minScore.toFixed(1)}</span>
                    </label>

                    <label class="label flex-1">
                        <span>Status</span>
                        <select bind:value={status} on:change={loadIdeas} class="select">
                            <option value="">All</option>
                            <option value="pending">Pending</option>
                            <option value="in-progress">In Progress</option>
                            <option value="completed">Completed</option>
                            <option value="archived">Archived</option>
                        </select>
                    </label>
                </div>
            </div>

            <!-- Ideas Grid -->
            {#if loading}
                <div class="flex justify-center p-12">
                    <div class="spinner"></div>
                </div>
            {:else if error}
                <div class="alert variant-filled-error">
                    {error}
                </div>
            {:else if ideas.length === 0}
                <div class="card p-12 text-center">
                    <p class="text-xl">No ideas found. Start by capturing one!</p>
                </div>
            {:else}
                <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                    {#each ideas as idea (idea.id)}
                        <IdeaCard
                            {idea}
                            onDelete={handleDeleteIdea}
                            onEdit={handleEditIdea}
                        />
                    {/each}
                </div>
            {/if}
        </div>

        <!-- Sidebar: Idea Form -->
        <div>
            <IdeaForm on:submit={handleCreateIdea} />
        </div>
    </div>
</div>

<style>
    .gradient-text {
        background: linear-gradient(135deg, var(--color-primary-500) 0%, var(--color-secondary-500) 100%);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
        background-clip: text;
    }

    .spinner {
        border: 4px solid rgba(0, 0, 0, 0.1);
        border-left-color: var(--color-primary-500);
        border-radius: 50%;
        width: 48px;
        height: 48px;
        animation: spin 1s linear infinite;
    }

    @keyframes spin {
        to { transform: rotate(360deg); }
    }
</style>
```

**Other pages:**
- [ ] `/ideas/[id]` - Idea detail page
- [ ] `/ideas/[id]/edit` - Edit idea page
- [ ] `/analytics` - Analytics dashboard
- [ ] `/settings` - Settings page

**Action Items:**
- [ ] Build dashboard page
- [ ] Build idea detail page
- [ ] Build analytics page
- [ ] Build settings page
- [ ] Add navigation menu
- [ ] Add loading states
- [ ] Add error handling
- [ ] Make all pages responsive

---

### Phase 5: Integration & Polish

#### Task 5.1: End-to-End Testing
**Estimated Time:** 8 hours

**Install Playwright:**
```bash
cd web
npm install -D @playwright/test
npx playwright install
```

**File:** `web/tests/e2e/ideas.spec.ts`
```typescript
import { test, expect } from '@playwright/test';

test.describe('Ideas Management', () => {
    test('should create a new idea', async ({ page }) => {
        await page.goto('/');

        // Fill form
        await page.fill('input[placeholder="What\'s your idea?"]', 'Test Idea');
        await page.fill('textarea[placeholder="Add more details..."]', 'This is a test idea');

        // Submit
        await page.click('button:has-text("Capture Idea")');

        // Verify created
        await expect(page.locator('text=Test Idea')).toBeVisible();
    });

    test('should filter ideas by score', async ({ page }) => {
        await page.goto('/');

        // Set min score
        await page.fill('input[type="range"]', '7');

        // Wait for filter
        await page.waitForTimeout(500);

        // Verify all visible ideas have score >= 7
        const scores = await page.locator('.badge').allTextContents();
        scores.forEach(score => {
            expect(parseFloat(score)).toBeGreaterThanOrEqual(7);
        });
    });

    test('should delete an idea', async ({ page }) => {
        await page.goto('/');

        // Find first idea
        const firstIdea = page.locator('.card').first();
        const ideaText = await firstIdea.locator('h3').textContent();

        // Delete
        await firstIdea.locator('button:has-text("Delete")').click();
        await page.click('button:has-text("OK")'); // Confirm dialog

        // Verify deleted
        await expect(page.locator(`text=${ideaText}`)).not.toBeVisible();
    });
});
```

**Action Items:**
- [ ] Write E2E tests for all user flows
- [ ] Test idea creation
- [ ] Test idea editing
- [ ] Test idea deletion
- [ ] Test filtering
- [ ] Test analytics
- [ ] Set up CI for E2E tests

#### Task 5.2: Performance Optimization
**Estimated Time:** 6 hours

**Backend:**
- [ ] Add database indexes
- [ ] Implement response caching
- [ ] Add gzip compression middleware
- [ ] Optimize database queries
- [ ] Add connection pooling

**Frontend:**
- [ ] Lazy load routes
- [ ] Implement virtual scrolling for long lists
- [ ] Optimize images
- [ ] Add service worker for offline support
- [ ] Enable SvelteKit prerendering

**File:** `web/src/routes/+layout.ts`
```typescript
export const prerender = true;
export const ssr = true;
```

#### Task 5.3: Security Audit
**Estimated Time:** 4 hours

**Checklist:**
- [ ] Add CSRF protection
- [ ] Sanitize user inputs
- [ ] Add rate limiting
- [ ] Implement proper CORS
- [ ] Add security headers
- [ ] Scan dependencies for vulnerabilities
- [ ] Add SQL injection protection (parameterized queries)
- [ ] Add XSS protection

**File:** `internal/api/middleware/security.go`
```go
package middleware

import (
    "net/http"
)

func SecurityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        next.ServeHTTP(w, r)
    })
}
```

#### Task 5.4: Documentation
**Estimated Time:** 8 hours

**Documents to create:**
- [ ] `README.md` - Project overview and quick start
- [ ] `docs/CLI.md` - CLI command reference
- [ ] `docs/API.md` - API documentation
- [ ] `docs/DEVELOPMENT.md` - Developer setup guide
- [ ] `docs/DEPLOYMENT.md` - Deployment guide
- [ ] `docs/MIGRATION.md` - Migration from Rust version
- [ ] `CHANGELOG.md` - Version history

#### Task 5.5: Docker & Deployment
**Estimated Time:** 6 hours

**File:** `Dockerfile`
```dockerfile
# Build stage for Go
FROM golang:1.21-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o /tm-cli ./cmd/cli
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o /tm-web ./cmd/web

# Build stage for SvelteKit
FROM node:18-alpine AS node-builder
WORKDIR /app
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates sqlite
WORKDIR /root/

# Copy Go binaries
COPY --from=go-builder /tm-cli /usr/local/bin/
COPY --from=go-builder /tm-web /usr/local/bin/

# Copy SvelteKit build
COPY --from=node-builder /app/build /var/www/build

ENV PORT=8080
EXPOSE 8080

CMD ["tm-web"]
```

**File:** `docker-compose.yml`
```yaml
version: '3.8'

services:
  web:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
      - ./telos.md:/root/telos.md
    environment:
      - TELOS_FILE=/root/telos.md
      - DATABASE_PATH=/data/ideas.db
      - PORT=8080
    restart: unless-stopped

  cli:
    build: .
    entrypoint: ["tm-cli"]
    volumes:
      - ./data:/data
      - ./telos.md:/root/telos.md
    environment:
      - TELOS_FILE=/root/telos.md
      - DATABASE_PATH=/data/ideas.db
```

**Action Items:**
- [ ] Create multi-stage Dockerfile
- [ ] Create docker-compose.yml
- [ ] Test Docker builds
- [ ] Create deployment scripts
- [ ] Set up CI/CD for Docker images
- [ ] Document deployment process

---

### Phase 6: Beta Release

#### Task 6.1: Beta Deployment
**Estimated Time:** 4 hours

**Deployment checklist:**
- [ ] Deploy to staging environment
- [ ] Run smoke tests
- [ ] Verify database migrations
- [ ] Test with real telos.md
- [ ] Monitor logs and errors
- [ ] Set up monitoring/alerting

#### Task 6.2: User Acceptance Testing
**Estimated Time:** Variable (1 week)

**Test scenarios:**
- [ ] Import existing ideas from Rust version
- [ ] Capture new ideas
- [ ] Review and filter ideas
- [ ] Analyze patterns
- [ ] Export data
- [ ] CLI workflow
- [ ] Web UI workflow
- [ ] Mobile responsiveness

#### Task 6.3: Bug Fixes & Polish
**Estimated Time:** Variable

Based on feedback:
- [ ] Fix reported bugs
- [ ] Address UX issues
- [ ] Improve performance bottlenecks
- [ ] Refine UI/UX
- [ ] Update documentation

---

## Testing Strategy

### Unit Tests

**Go:**
```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package
go test ./internal/scoring -v
```

**SvelteKit:**
```bash
cd web
npm test
npm run test:coverage
```

**Coverage goals:**
- Go: >80% coverage
- TypeScript: >70% coverage

### Integration Tests

**Database tests:**
```go
func TestDatabaseIntegration(t *testing.T) {
    // Use temp database
    repo, cleanup := setupTestDB(t)
    defer cleanup()

    // Test CRUD operations
    idea := &models.Idea{Title: "Test"}
    err := repo.CreateIdea(context.Background(), idea)
    assert.NoError(t, err)
}
```

**API tests:**
```go
func TestAPIIntegration(t *testing.T) {
    // Set up test server
    server := setupTestServer(t)
    defer server.Close()

    // Test endpoints
    resp, err := http.Get(server.URL + "/api/v1/ideas")
    assert.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)
}
```

### E2E Tests

Run with Playwright:
```bash
cd web
npx playwright test
npx playwright test --ui  # Interactive mode
```

---

## Deployment Strategy

### Development

```bash
# Terminal 1: Go API
make dev-api

# Terminal 2: SvelteKit
cd web && npm run dev

# Terminal 3: CLI testing
make build-cli && ./bin/tm dump "Test idea"
```

### Staging

```bash
# Deploy to staging
make deploy-staging

# Run smoke tests
make test-staging
```

### Production

```bash
# Build all artifacts
make build-all

# Deploy
make deploy-production

# Verify deployment
make health-check
```

---

## Success Criteria

### Phase 0: Preparation
- [ ] Go project structure created
- [ ] CI/CD pipeline working
- [ ] Dev environment documented
- [ ] Rust behavior documented

### Phase 1: Core Domain
- [ ] All models implemented
- [ ] Telos parser working
- [ ] Scoring engine matches Rust
- [ ] Database layer functional
- [ ] 80%+ test coverage

### Phase 2: CLI
- [ ] All commands implemented
- [ ] Feature parity with Rust CLI
- [ ] Help text complete
- [ ] CLI tests passing

### Phase 3: API
- [ ] All endpoints implemented
- [ ] OpenAPI docs generated
- [ ] API tests passing
- [ ] CORS configured

### Phase 4: Frontend
- [ ] Dashboard working
- [ ] CRUD operations functional
- [ ] Responsive design
- [ ] E2E tests passing

### Phase 5: Integration
- [ ] Full stack integrated
- [ ] Performance optimized
- [ ] Security hardened
- [ ] Documentation complete

### Phase 6: Beta
- [ ] Beta deployed
- [ ] User feedback collected
- [ ] Critical bugs fixed
- [ ] Ready for v1.0

---

## Risk Mitigation

### Risk 1: Scoring Algorithm Mismatch
**Mitigation:**
- Document Rust behavior extensively
- Create test fixtures with expected outputs
- Validate Go output against Rust output
- Keep Rust version for comparison

### Risk 2: Data Migration Issues
**Mitigation:**
- Write migration script early
- Test with production data copies
- Keep backup of Rust database
- Support both formats temporarily

### Risk 3: Timeline Slippage
**Mitigation:**
- Build MVP first (CLI + basic web)
- Ship iteratively
- Cut scope if needed (defer analytics, advanced features)
- Keep Rust version as fallback

### Risk 4: Performance Regression
**Mitigation:**
- Benchmark critical paths
- Load test API endpoints
- Optimize database queries
- Use profiling tools

### Risk 5: User Adoption
**Mitigation:**
- Maintain feature parity
- Provide migration guide
- Support both versions temporarily
- Gather early feedback

---

## Rollback Plan

If migration fails:

1. **Keep Rust version available**
   - Don't delete Rust codebase
   - Keep Docker images published
   - Maintain branch for bug fixes

2. **Data portability**
   - Export/import between versions
   - Keep database schema compatible
   - Provide migration tools

3. **Gradual migration**
   - Run both versions side-by-side
   - Migrate users gradually
   - Collect feedback continuously

4. **Decision points**
   - Week 4: CLI ready? (Go/No-Go)
   - Week 6: API ready? (Go/No-Go)
   - Week 8: Beta feedback positive? (Go/No-Go)

---

## Resources & Dependencies

### Go Libraries
```
go get github.com/spf13/cobra
go get github.com/spf13/viper
go get github.com/go-chi/chi/v5
go get github.com/mattn/go-sqlite3
go get github.com/google/uuid
go get github.com/stretchr/testify
```

### Node.js Packages
```
npm install svelte @sveltejs/kit
npm install tailwindcss @skeletonlabs/skeleton
npm install lucide-svelte
npm install chart.js
npm install @playwright/test
```

### Tools
- Go 1.21+
- Node.js 18+
- Docker
- git
- make

---

## Timeline Summary

| Phase | Duration | Deliverables |
|-------|----------|--------------|
| Phase 0: Preparation | 1 week | Project structure, CI/CD |
| Phase 1: Core Domain | 2 weeks | Models, scoring, database |
| Phase 2: CLI | 1 week | Full CLI implementation |
| Phase 3: API | 1 week | RESTful API server |
| Phase 4: Frontend | 2 weeks | SvelteKit web UI |
| Phase 5: Integration | 1 week | Testing, docs, deployment |
| Phase 6: Beta | 1 week | Beta release, feedback |
| **Total** | **8 weeks** | **Production-ready v2.0** |

---

## Next Steps

1. Review this plan
2. Get stakeholder approval
3. Set up project repository
4. Begin Phase 0: Preparation
5. Schedule weekly checkpoints
6. Start coding!

---

**This plan is a living document. Update as you progress and learn.**
