# Phase 1: Core Domain Migration (TDD)

**Duration:** 7-10 days
**Goal:** Port all core business logic from Rust to Go with comprehensive test coverage

**CRITICAL:** Follow Test-Driven Development strictly. Write tests FIRST.

---

## Context

You are implementing the core business logic for Telos Idea Matrix in Go. This phase is the foundation for everything else - the CLI and API will simply wrap these components.

**Prerequisites:**
- Phase 0 complete (project structure exists)
- `RUST_REFERENCE.md` available for reference
- Rust source code at `/home/user/brain-salad/src/`

**Components to Implement:**
1. Data Models (Idea, Telos, Analysis, Pattern)
2. Telos Parser (markdown → structs)
3. Scoring Engine (idea → score)
4. Pattern Detector (idea → detected patterns)
5. Database Layer (SQLite CRUD)

---

## TDD Requirements

### Coverage Targets

```
Component            Target    Priority
─────────────────────────────────────────
Models & Types       >90%      CRITICAL
Scoring Engine       >95%      CRITICAL
Telos Parser         >90%      CRITICAL
Pattern Detector     >90%      CRITICAL
Database Repository  >85%      HIGH
─────────────────────────────────────────
OVERALL             >90%      REQUIRED
```

### TDD Workflow (RED → GREEN → REFACTOR)

For **every** function you write:

1. **🔴 RED**: Write failing test first
2. **🟢 GREEN**: Write minimal code to pass
3. **♻️ REFACTOR**: Improve code quality

**Never** write implementation before tests.

---

## Component 1: Data Models

### Goal
Define all Go structs matching Rust behavior from `RUST_REFERENCE.md`.

### Location
`internal/models/`

### Files to Create
- `idea.go` - Idea struct and methods
- `telos.go` - Telos, Goal, Strategy, Stack, Pattern structs
- `analysis.go` - Analysis and DetectedPattern structs
- `models_test.go` - All model tests

### TDD Implementation

#### Step 1: Write Tests First (RED)

Create `internal/models/models_test.go`:

```go
package models_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
)

// Test Idea creation and validation
func TestIdea_Validate_ValidIdea_ReturnsNoError(t *testing.T) {
	idea := &models.Idea{
		Title:  "Build a SaaS product",
		Status: "pending",
	}

	err := idea.Validate()
	assert.NoError(t, err)
}

func TestIdea_Validate_EmptyTitle_ReturnsError(t *testing.T) {
	idea := &models.Idea{
		Title:  "",
		Status: "pending",
	}

	err := idea.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "title")
}

func TestIdea_Validate_TitleTooShort_ReturnsError(t *testing.T) {
	idea := &models.Idea{
		Title:  "AB",  // Less than minimum (3 chars)
		Status: "pending",
	}

	err := idea.Validate()
	assert.Error(t, err)
}

func TestIdea_Validate_InvalidStatus_ReturnsError(t *testing.T) {
	idea := &models.Idea{
		Title:  "Valid title",
		Status: "invalid_status",
	}

	err := idea.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status")
}

func TestIdea_JSONSerialization_RoundTrip(t *testing.T) {
	original := &models.Idea{
		ID:          uuid.New(),
		Title:       "Test Idea",
		Description: "Testing JSON",
		Score:       8.5,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Status:      "pending",
		Tags:        []string{"test", "json"},
	}

	// Serialize
	jsonBytes, err := json.Marshal(original)
	assert.NoError(t, err)

	// Deserialize
	var decoded models.Idea
	err = json.Unmarshal(jsonBytes, &decoded)
	assert.NoError(t, err)

	// Compare
	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.Title, decoded.Title)
	assert.Equal(t, original.Score, decoded.Score)
}

// Test Telos struct
func TestTelos_Validate_ValidTelos_ReturnsNoError(t *testing.T) {
	telos := &models.Telos{
		Goals: []models.Goal{
			{ID: "G1", Description: "Build products", Priority: 1},
		},
		Strategies: []models.Strategy{
			{ID: "S1", Description: "Ship early"},
		},
		Stack: models.Stack{
			Primary:   []string{"Go", "TypeScript"},
			Secondary: []string{"Docker"},
		},
	}

	err := telos.Validate()
	assert.NoError(t, err)
}

// Add more tests for all edge cases...
```

Run tests - **they should fail** (code doesn't exist yet):
```bash
$ go test ./internal/models -v
# Error: cannot find package "internal/models"
```

#### Step 2: Implement Models (GREEN)

Create `internal/models/idea.go`:

```go
package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Idea represents a captured idea with analysis
type Idea struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Title       string     `json:"title" db:"title"`
	Description string     `json:"description" db:"description"`
	Score       float64    `json:"score" db:"score"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	Status      string     `json:"status" db:"status"`
	Tags        []string   `json:"tags"`
	Analysis    *Analysis  `json:"analysis,omitempty"`
}

// Validate validates the idea
func (i *Idea) Validate() error {
	if i.Title == "" {
		return errors.New("title is required")
	}
	if len(i.Title) < 3 {
		return errors.New("title must be at least 3 characters")
	}
	if len(i.Title) > 200 {
		return errors.New("title must be at most 200 characters")
	}

	validStatuses := map[string]bool{
		"pending":     true,
		"in-progress": true,
		"completed":   true,
		"archived":    true,
	}

	if i.Status != "" && !validStatuses[i.Status] {
		return errors.New("invalid status")
	}

	return nil
}
```

Create `internal/models/telos.go`:

```go
package models

import (
	"errors"
	"time"
)

// Telos represents the user's goals and values
type Telos struct {
	Goals           []Goal     `json:"goals"`
	Strategies      []Strategy `json:"strategies"`
	Stack           Stack      `json:"stack"`
	FailurePatterns []Pattern  `json:"failure_patterns"`
	LoadedAt        time.Time  `json:"loaded_at"`
}

// Goal represents a user goal
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

// Validate validates the telos
func (t *Telos) Validate() error {
	if len(t.Goals) == 0 {
		return errors.New("at least one goal is required")
	}
	return nil
}
```

Create `internal/models/analysis.go`:

```go
package models

import "time"

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
```

Run tests - **they should pass now**:
```bash
$ go test ./internal/models -v
=== RUN   TestIdea_Validate_ValidIdea_ReturnsNoError
--- PASS: TestIdea_Validate_ValidIdea_ReturnsNoError (0.00s)
=== RUN   TestIdea_Validate_EmptyTitle_ReturnsError
--- PASS: TestIdea_Validate_EmptyTitle_ReturnsError (0.00s)
...
PASS
ok      github.com/rayyacub/telos-idea-matrix/internal/models  0.123s
```

#### Step 3: Refactor

- Add godoc comments
- Extract validation constants
- Improve error messages

### Deliverables

- [ ] `internal/models/idea.go` - Idea struct with validation
- [ ] `internal/models/telos.go` - Telos and related structs
- [ ] `internal/models/analysis.go` - Analysis structs
- [ ] `internal/models/models_test.go` - Comprehensive tests
- [ ] All tests passing
- [ ] Coverage >90%

---

## Component 2: Telos Parser

### Goal
Parse `telos.md` markdown files into Telos structs.

### Location
`internal/telos/`

### Files to Create
- `parser.go` - Parser implementation
- `parser_test.go` - Parser tests
- `testdata/valid_telos.md` - Test fixture
- `testdata/minimal_telos.md` - Minimal test case
- `testdata/invalid_telos.md` - Error test case

### TDD Implementation

#### Step 1: Create Test Fixtures

Create `internal/telos/testdata/valid_telos.md`:

```markdown
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

#### Step 2: Write Tests First (RED)

Create `internal/telos/parser_test.go`:

```go
package telos_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/rayyacub/telos-idea-matrix/internal/telos"
)

func TestParseFile_ValidTelos_ParsesAllSections(t *testing.T) {
	parser := telos.NewParser()

	result, err := parser.ParseFile("testdata/valid_telos.md")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Goals, 2)
	assert.Len(t, result.Strategies, 2)
	assert.Len(t, result.FailurePatterns, 2)
	assert.Len(t, result.Stack.Primary, 3)
	assert.Len(t, result.Stack.Secondary, 2)
}

func TestParseFile_ValidTelos_ParsesGoalsWithDeadlines(t *testing.T) {
	parser := telos.NewParser()

	result, err := parser.ParseFile("testdata/valid_telos.md")

	assert.NoError(t, err)
	assert.Equal(t, "G1", result.Goals[0].ID)
	assert.Contains(t, result.Goals[0].Description, "SaaS product")
	assert.NotNil(t, result.Goals[0].Deadline)
	assert.Equal(t, 2025, result.Goals[0].Deadline.Year())
	assert.Equal(t, 12, int(result.Goals[0].Deadline.Month()))
	assert.Equal(t, 31, result.Goals[0].Deadline.Day())
}

func TestParseFile_ValidTelos_ParsesStrategies(t *testing.T) {
	parser := telos.NewParser()

	result, err := parser.ParseFile("testdata/valid_telos.md")

	assert.NoError(t, err)
	assert.Equal(t, "S1", result.Strategies[0].ID)
	assert.Contains(t, result.Strategies[0].Description, "Ship early")
}

func TestParseFile_ValidTelos_ParsesStack(t *testing.T) {
	parser := telos.NewParser()

	result, err := parser.ParseFile("testdata/valid_telos.md")

	assert.NoError(t, err)
	assert.Contains(t, result.Stack.Primary, "Go")
	assert.Contains(t, result.Stack.Primary, "TypeScript")
	assert.Contains(t, result.Stack.Secondary, "Docker")
}

func TestParseFile_ValidTelos_ParsesFailurePatterns(t *testing.T) {
	parser := telos.NewParser()

	result, err := parser.ParseFile("testdata/valid_telos.md")

	assert.NoError(t, err)
	assert.Equal(t, "Context switching", result.FailurePatterns[0].Name)
	assert.Contains(t, result.FailurePatterns[0].Description, "Starting new projects")
	// Keywords should be extracted from description
	assert.Contains(t, result.FailurePatterns[0].Keywords, "new")
	assert.Contains(t, result.FailurePatterns[0].Keywords, "projects")
}

func TestParseFile_MissingFile_ReturnsError(t *testing.T) {
	parser := telos.NewParser()

	_, err := parser.ParseFile("testdata/nonexistent.md")

	assert.Error(t, err)
}

func TestParseFile_EmptyFile_ReturnsError(t *testing.T) {
	parser := telos.NewParser()

	_, err := parser.ParseFile("testdata/empty.md")

	assert.Error(t, err)
}
```

Run tests - **they should fail**:
```bash
$ go test ./internal/telos -v
# Error: undefined: telos.NewParser
```

#### Step 3: Implement Parser (GREEN)

Create `internal/telos/parser.go`:

```go
package telos

import (
	"bufio"
	"fmt"
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
	patternRegex  *regexp.Regexp
}

// NewParser creates a new Telos parser
func NewParser() *Parser {
	return &Parser{
		goalRegex:     regexp.MustCompile(`^-\s+(G\d+):\s+(.+?)(?:\s+\(Deadline:\s+(.+?)\))?$`),
		strategyRegex: regexp.MustCompile(`^-\s+(S\d+):\s+(.+)$`),
		deadlineRegex: regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})`),
		patternRegex:  regexp.MustCompile(`^-\s+([^:]+):\s+(.+)$`),
	}
}

// ParseFile parses a telos.md file
func (p *Parser) ParseFile(path string) (*models.Telos, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	telos := &models.Telos{
		LoadedAt: time.Now(),
	}

	scanner := bufio.NewScanner(file)
	var currentSection string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

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
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	if err := telos.Validate(); err != nil {
		return nil, fmt.Errorf("invalid telos: %w", err)
	}

	return telos, nil
}

// parseGoal parses a goal line
func (p *Parser) parseGoal(line string) *models.Goal {
	matches := p.goalRegex.FindStringSubmatch(line)
	if len(matches) < 3 {
		return nil
	}

	goal := &models.Goal{
		ID:          matches[1],
		Description: matches[2],
		Priority:    0, // TODO: extract from order
	}

	// Parse deadline if present
	if len(matches) > 3 && matches[3] != "" {
		if deadline, err := time.Parse("2006-01-02", matches[3]); err == nil {
			goal.Deadline = &deadline
		}
	}

	return goal
}

// parseStrategy parses a strategy line
func (p *Parser) parseStrategy(line string) *models.Strategy {
	matches := p.strategyRegex.FindStringSubmatch(line)
	if len(matches) < 3 {
		return nil
	}

	return &models.Strategy{
		ID:          matches[1],
		Description: matches[2],
	}
}

// parseStack parses a stack line
func (p *Parser) parseStack(line string, stack *models.Stack) {
	if strings.HasPrefix(line, "- Primary:") {
		techs := strings.TrimPrefix(line, "- Primary:")
		stack.Primary = parseTechList(techs)
	} else if strings.HasPrefix(line, "- Secondary:") {
		techs := strings.TrimPrefix(line, "- Secondary:")
		stack.Secondary = parseTechList(techs)
	}
}

// parsePattern parses a failure pattern line
func (p *Parser) parsePattern(line string) *models.Pattern {
	matches := p.patternRegex.FindStringSubmatch(line)
	if len(matches) < 3 {
		return nil
	}

	return &models.Pattern{
		Name:        strings.TrimSpace(matches[1]),
		Description: strings.TrimSpace(matches[2]),
		Keywords:    extractKeywords(matches[2]),
	}
}

// Helper: parse comma-separated tech list
func parseTechList(text string) []string {
	var result []string
	parts := strings.Split(text, ",")
	for _, part := range parts {
		tech := strings.TrimSpace(part)
		if tech != "" {
			result = append(result, tech)
		}
	}
	return result
}

// Helper: extract keywords from description
func extractKeywords(text string) []string {
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true,
		"or": true, "but": true, "in": true, "on": true,
		"at": true, "to": true, "for": true, "of": true,
		"with": true, "from": true, "before": true,
	}

	words := strings.Fields(strings.ToLower(text))
	var keywords []string

	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?;:")

		if len(word) > 3 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}

	return keywords
}
```

Run tests - **should pass now**:
```bash
$ go test ./internal/telos -v
PASS
coverage: 87.3%
```

### Deliverables

- [ ] `internal/telos/parser.go` - Complete parser
- [ ] `internal/telos/parser_test.go` - All test cases
- [ ] Test fixtures in `testdata/`
- [ ] Coverage >90%

---

## Component 3: Scoring Engine (MOST CRITICAL)

### Goal
Implement scoring algorithm that **exactly matches** Rust version.

### Location
`internal/scoring/`

### TDD Implementation

**CRITICAL:** Reference `RUST_REFERENCE.md` for exact scoring algorithm.

#### Step 1: Create Test Fixtures

Create `internal/scoring/testdata/high_score_idea.json`:
```json
{
  "title": "Build a Go-based SaaS product using PostgreSQL",
  "description": "Create profitable SaaS leveraging Go backend and TypeScript frontend",
  "expected_score": 8.5,
  "expected_mission_alignment": 0.9,
  "expected_anti_pattern_score": 1.0,
  "expected_strategic_fit": 0.85
}
```

Create `internal/scoring/testdata/test_telos.md` (for testing).

#### Step 2: Write Tests First (RED)

Reference Rust tests - create equivalent Go tests.

Create `internal/scoring/engine_test.go` with comprehensive test suite covering:
- Perfect alignment (score >8)
- No alignment (score <4)
- Pattern detection
- Strategic fit
- Stack alignment
- Edge cases (empty telos, nil idea, etc.)

#### Step 3: Implement Scoring (GREEN)

Implement exact algorithm from `RUST_REFERENCE.md`.

#### Step 4: Validate Against Rust

**CRITICAL VALIDATION:**

```bash
# Run Rust version
cd /home/user/brain-salad
cargo build --release
./target/release/tm dump "Build a Go CLI tool" > rust_output.txt

# Run Go version (after implementation)
cd /home/user/telos-idea-matrix-go
go build -o bin/tm ./cmd/cli
./bin/tm dump "Build a Go CLI tool" > go_output.txt

# Compare scores
# Should match within 0.1 points
```

### Deliverables

- [ ] `internal/scoring/engine.go` - Scoring algorithm
- [ ] `internal/scoring/engine_test.go` - Comprehensive tests
- [ ] Test fixtures
- [ ] Validation: Scores match Rust within 0.1 points
- [ ] Coverage >95%

---

## Component 4: Pattern Detector

### Goal
Detect anti-patterns in ideas.

### Implementation
Similar TDD workflow as above.

### Deliverables

- [ ] `internal/patterns/detector.go`
- [ ] `internal/patterns/detector_test.go`
- [ ] Coverage >90%

---

## Component 5: Database Layer

### Goal
SQLite CRUD operations with migrations.

### Location
`internal/database/`

### TDD Implementation

Use integration tests (they need a real database).

Mark with build tag:
```go
//go:build integration
// +build integration

package database_test
```

Test helper:
```go
func setupTestDB(t *testing.T) (*Repository, func()) {
	t.Helper()
	tmpfile, err := os.CreateTemp("", "test_*.db")
	require.NoError(t, err)

	repo, err := NewRepository(tmpfile.Name())
	require.NoError(t, err)

	cleanup := func() {
		repo.Close()
		os.Remove(tmpfile.Name())
	}

	return repo, cleanup
}
```

### Deliverables

- [ ] `internal/database/repository.go` - All CRUD operations
- [ ] `internal/database/repository_test.go` - Integration tests
- [ ] `internal/database/migrations/001_initial.sql` - Schema
- [ ] Coverage >85%

---

## Validation

Before Phase 1 is complete:

### ✅ Checklist

- [ ] All components implemented
- [ ] All tests passing: `go test ./internal/... -v`
- [ ] Coverage >90%: `go test ./internal/... -cover`
- [ ] Integration tests passing: `go test -tags=integration ./... -v`
- [ ] Scoring matches Rust (validated with same test cases)
- [ ] No linter errors: `golangci-lint run`
- [ ] Code formatted: `gofmt -l .` (no output)

### 🧪 Validation Commands

```bash
# All tests
go test ./internal/... -v -cover

# Integration tests
go test -tags=integration ./internal/database -v

# Coverage report
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Validate scoring against Rust
# (Manual comparison with same test ideas)

# Lint
golangci-lint run

# Format check
gofmt -l .
```

### 📊 Coverage Report

Generate and review:
```bash
make test-coverage
open coverage.html
```

Ensure all critical paths covered.

---

## Success Criteria

Phase 1 complete when:

✅ All data models implemented with validation
✅ Telos parser works with real telos.md files
✅ Scoring engine matches Rust output exactly
✅ Pattern detection working
✅ Database layer functional (CRUD + migrations)
✅ All tests passing (unit + integration)
✅ Coverage >90% overall, >95% for scoring
✅ No linter errors
✅ Code well-documented (godoc comments)

---

## Handoff to Phase 2

**What Phase 2 needs:**
- Working core domain logic
- All business logic testable and tested
- Database migrations working
- Scoring validated against Rust

**Next steps:**
1. Commit all Phase 1 work
2. Review Phase 2 prompt: `migration-prompts/phase-2-cli.md`
3. Launch Phase 2 subagent to build CLI

---

## Troubleshooting

### Tests failing with different scores than Rust

- Double-check algorithm in `RUST_REFERENCE.md`
- Compare intermediate values (mission alignment, pattern scores, etc.)
- Use same test data
- Check floating-point precision

### Database tests failing

- Check temp file creation works
- Verify SQLite installed
- Check migrations run correctly
- Ensure proper cleanup

### Coverage too low

- Identify uncovered lines: `go tool cover -html=coverage.out`
- Add tests for edge cases
- Test error paths
- Use table-driven tests for multiple scenarios

---

## Time Estimates

- Models: 1 day
- Telos Parser: 2 days
- Scoring Engine: 3-4 days (most complex)
- Pattern Detector: 1 day
- Database Layer: 2 days
- **Total: 7-10 days**

Take your time. Quality over speed. Write tests first!
