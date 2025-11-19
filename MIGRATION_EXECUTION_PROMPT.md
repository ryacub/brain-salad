# Execution Prompt: Go Migration of Telos Idea Matrix

**⚠️ RECOMMENDED APPROACH: Use the new Subagent Orchestration method instead!**

See **`SUBAGENT_ORCHESTRATION.md`** for the preferred execution strategy using specialized subagents for each phase.

This document remains as reference for the original single-agent approach. For better effectiveness and parallel execution, use the subagent method with phase-specific prompts in `migration-prompts/`.

---

## Original Approach (Single Agent)

**Use this prompt with Claude or another AI assistant to guide the migration implementation.**

---

## Context

You are migrating the Telos Idea Matrix from Rust to Go + SvelteKit. The current Rust implementation is a CLI tool for capturing and analyzing ideas against personal goals (Telos). The new architecture will be:

- **Go CLI** - Lightweight command-line tool for power users
- **Go API Server** - RESTful backend sharing core logic with CLI
- **SvelteKit Frontend** - Modern, beautiful web UI for managing ideas

**Key Requirements:**
1. **Feature parity** with Rust version
2. **Test-Driven Development** - Write tests first for all new code
3. **Faster iteration** - Leverage Go's quick compile times
4. **Shared codebase** - 80%+ code reuse between CLI and API
5. **Beautiful UI** - Modern, responsive SvelteKit frontend

**Reference Documents:**
- `GO_MIGRATION_PLAN.md` - Comprehensive migration plan
- `LANGUAGE_ANALYSIS.md` - Analysis of why Go is better for this project
- `RUST_REFERENCE.md` - Documentation of current Rust behavior (to be created)

---

## Your Task

Implement the migration following the plan in `GO_MIGRATION_PLAN.md`. Use **Test-Driven Development (TDD)** for all new code:

1. **Red**: Write failing tests first
2. **Green**: Write minimal code to make tests pass
3. **Refactor**: Clean up code while keeping tests green

---

## Phase-by-Phase Execution

### Phase 0: Preparation (Week 1)

**Goal:** Set up project structure and foundation

**Tasks:**
```bash
# 1. Create Go project
mkdir telos-idea-matrix-go
cd telos-idea-matrix-go
go mod init github.com/rayyacub/telos-idea-matrix

# 2. Create directory structure (see GO_MIGRATION_PLAN.md Task 0.1)

# 3. Set up development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/cosmtrek/air@latest  # Live reload

# 4. Create initial files
touch cmd/cli/main.go
touch cmd/web/main.go
touch internal/models/idea.go
touch Makefile
```

**Prompt for AI Assistant:**

```
I'm starting Phase 0 of the Go migration for Telos Idea Matrix.

Please help me:
1. Create the complete project directory structure as specified in GO_MIGRATION_PLAN.md Task 0.1
2. Set up the initial go.mod with these dependencies:
   - github.com/spf13/cobra (CLI)
   - github.com/spf13/viper (config)
   - github.com/go-chi/chi/v5 (routing)
   - github.com/mattn/go-sqlite3 (database)
   - github.com/google/uuid (IDs)
   - github.com/stretchr/testify (testing)
3. Create a Makefile with targets for: build, test, lint, run, clean
4. Create .gitignore for Go projects
5. Set up golangci-lint configuration (.golangci.yml)
6. Create a basic README.md

Use TDD approach: Start by creating test files (even if empty) alongside each new file.

Reference the existing Rust implementation in /home/user/brain-salad/src/ to understand:
- Current data models
- Scoring algorithm
- Database schema
- CLI commands

Create RUST_REFERENCE.md documenting:
- All struct definitions
- Scoring algorithm with examples
- Database schema
- Expected behavior for each CLI command
```

---

### Phase 1: Core Domain Migration (Week 2-3)

**Goal:** Port core business logic to Go with comprehensive test coverage

**TDD Workflow Example:**

```
For each component (models, telos parser, scoring engine, database):

1. Write Tests First (Red)
2. Implement Code (Green)
3. Refactor & Document (Refactor)
```

**Prompt for AI Assistant (Task 1.1: Data Models):**

```
Phase 1, Task 1.1: Implement data models with TDD

Following the TDD approach:

STEP 1 (RED): Write tests first
- Create internal/models/idea_test.go
- Write tests for Idea struct validation
- Test cases:
  * Valid idea creation
  * Invalid idea (missing title)
  * Invalid idea (title too short)
  * Invalid idea (invalid status)
  * UUID generation
  * JSON serialization/deserialization

STEP 2 (GREEN): Implement minimal code
- Create internal/models/idea.go
- Define Idea struct with tags:
  * json tags for API
  * db tags for database
  * validate tags for validation
- Implement Validate() method using go-playground/validator

STEP 3 (REFACTOR): Clean up
- Add documentation comments
- Extract validation logic if needed
- Ensure all tests pass

Reference: Look at the Rust implementation in src/types.rs to understand:
- What fields are needed
- What validation rules exist
- How ideas are structured

Create similar models for:
- Analysis
- DetectedPattern
- Telos
- Goal
- Strategy
- Stack
- Pattern

Run tests after each: go test ./internal/models -v
Target coverage: >90%
```

**Prompt for AI Assistant (Task 1.2: Telos Parser):**

```
Phase 1, Task 1.2: Implement Telos parser with TDD

STEP 1 (RED): Write tests first
- Create internal/telos/parser_test.go
- Create test fixtures in testdata/:
  * testdata/valid_telos.md (complete, valid telos file)
  * testdata/minimal_telos.md (minimal valid file)
  * testdata/invalid_telos.md (malformed file)
  * testdata/empty_telos.md (empty file)

- Test cases:
  * Parse valid telos file successfully
  * Extract all goals with deadlines
  * Extract all strategies
  * Parse tech stack (primary/secondary)
  * Extract failure patterns with keywords
  * Handle malformed markdown gracefully
  * Handle missing sections
  * Validate parsed data

STEP 2 (GREEN): Implement parser
- Create internal/telos/parser.go
- Implement Parser struct with regex patterns
- Implement ParseFile(path string) (*models.Telos, error)
- Implement helper methods:
  * parseGoal(line string) *models.Goal
  * parseStrategy(line string) *models.Strategy
  * parseStack(line string, stack *models.Stack)
  * parsePattern(line string) *models.Pattern

STEP 3 (REFACTOR): Improve
- Extract regex patterns to constants
- Add error messages
- Add logging for debugging
- Document expected markdown format

Reference: Check Rust implementation in src/telos.rs
- How does it parse the markdown?
- What regex patterns does it use?
- What edge cases does it handle?

Run: go test ./internal/telos -v -cover
Target coverage: >85%
```

**Prompt for AI Assistant (Task 1.3: Scoring Engine):**

```
Phase 1, Task 1.3: Implement scoring engine with TDD

This is the MOST CRITICAL component - scoring must match Rust exactly.

STEP 1 (RED): Write comprehensive tests
- Create internal/scoring/engine_test.go
- Create test fixtures with known scores:
  * testdata/high_score_idea.json (expected score: 8.5+)
  * testdata/low_score_idea.json (expected score: <4.0)
  * testdata/pattern_detected_idea.json (has anti-patterns)
  * testdata/test_telos.md (telos for test cases)

- Test cases (extract from Rust tests):
  * Mission alignment calculation
    - Perfect alignment (100%)
    - No alignment (0%)
    - Partial alignment (50%)
  * Anti-pattern detection
    - No patterns detected
    - Single pattern detected
    - Multiple patterns detected
    - Pattern confidence scoring
  * Strategic fit calculation
    - Stack alignment
    - Strategy alignment
  * Final score calculation
    - Weighted average correct
    - Score range 0-10
    - Precision to 1 decimal
  * Recommendations generation
    - High score recommendations
    - Low score warnings
    - Pattern warnings

STEP 2 (GREEN): Implement scoring engine
- Create internal/scoring/engine.go
- Implement Engine struct
- Implement Score(idea *models.Idea) (*models.Analysis, error)
- Implement helper methods:
  * calculateMissionAlignment(idea) float64
  * detectPatterns(idea) []models.DetectedPattern
  * calculateAntiPatternScore(patterns) float64
  * calculateStrategicFit(idea) float64
  * calculateStackAlignment(idea) float64
  * generateRecommendations(analysis) []string
  * calculateTextOverlap(text1, text2) float64
  * extractKeywords(text) []string

STEP 3 (REFACTOR): Optimize
- Add caching if needed
- Optimize text processing
- Add benchmarks for performance

VALIDATION: Compare Go output with Rust output
- Use same test ideas
- Scores should match within 0.1 points
- Detected patterns should be identical

Run: go test ./internal/scoring -v -cover -bench=.
Target coverage: >95% (this is critical business logic)
```

**Prompt for AI Assistant (Task 1.4: Database Layer):**

```
Phase 1, Task 1.4: Implement database layer with TDD

STEP 1 (RED): Write integration tests
- Create internal/database/repository_test.go
- Use temp database for each test (cleanup after)
- Test cases:
  * Database creation and migration
  * CRUD operations for ideas
    - CreateIdea
    - GetIdea (by ID)
    - ListIdeas (with filters)
    - UpdateIdea
    - DeleteIdea
  * Tag management
    - Add tags to idea
    - Query by tags
  * Analysis storage
    - Save analysis with idea
    - Retrieve latest analysis
    - Store detected patterns
  * Relationship management
    - Link ideas
    - Query relationships
  * Edge cases
    - Duplicate IDs
    - Non-existent IDs
    - Concurrent access
    - Transaction rollback

STEP 2 (GREEN): Implement repository
- Create internal/database/repository.go
- Implement Repository struct
- Implement NewRepository(dbPath string) (*Repository, error)
- Implement migrate() error
- Implement all CRUD methods
- Use transactions where needed
- Use parameterized queries (prevent SQL injection)

STEP 3 (REFACTOR): Optimize
- Add database indexes
- Add connection pooling
- Add prepared statements
- Add query logging

Schema should match Rust schema exactly:
- Check src/database_simple.rs for table definitions
- Ensure migration is compatible with existing Rust databases

Run: go test ./internal/database -v -cover
Target coverage: >80%
```

---

### Phase 2: CLI Implementation (Week 3-4)

**Goal:** Feature-complete CLI using Cobra

**Prompt for AI Assistant:**

```
Phase 2: Implement CLI with TDD

For each command (dump, analyze, score, review, prune, analytics, link):

STEP 1 (RED): Write CLI tests
- Create internal/cli/<command>_test.go
- Test cases:
  * Command execution with valid args
  * Command with invalid args
  * Flag parsing
  * Output format
  * Error handling

STEP 2 (GREEN): Implement command
- Create internal/cli/<command>.go
- Implement cobra.Command
- Wire up to core logic (models, scoring, database)
- Add flags and validation
- Format output

STEP 3 (REFACTOR): Improve UX
- Add colored output (github.com/fatih/color)
- Add progress indicators
- Improve error messages
- Add examples to help text

Example for 'dump' command:

// internal/cli/dump_test.go
func TestDumpCommand(t *testing.T) {
    // Setup
    repo := setupTestDB(t)
    defer repo.Close()

    // Test: Dump with idea text
    cmd := newDumpCommand(repo)
    cmd.SetArgs([]string{"Build a SaaS product"})

    err := cmd.Execute()
    assert.NoError(t, err)

    // Verify idea was saved
    ideas, _ := repo.ListIdeas(context.Background(), database.ListOptions{Limit: 1})
    assert.Len(t, ideas, 1)
    assert.Contains(t, ideas[0].Title, "SaaS")
}

Feature parity checklist with Rust CLI:
- [ ] All commands implemented
- [ ] All flags supported
- [ ] Output format matches (or improves)
- [ ] Error messages helpful
- [ ] Help text complete

Run: go test ./internal/cli/... -v
Build: go build -o bin/tm ./cmd/cli
Test manually: ./bin/tm dump "Test idea"
```

---

### Phase 3: API Server (Week 4-5)

**Goal:** RESTful API with OpenAPI docs

**Prompt for AI Assistant:**

```
Phase 3: Implement API server with TDD

STEP 1 (RED): Write API tests
- Create internal/api/handlers_test.go
- Use httptest for testing handlers
- Test cases for each endpoint:
  * GET /api/v1/ideas - List ideas
  * POST /api/v1/ideas - Create idea
  * GET /api/v1/ideas/{id} - Get idea
  * PUT /api/v1/ideas/{id} - Update idea
  * DELETE /api/v1/ideas/{id} - Delete idea
  * POST /api/v1/analyze - Analyze text
  * GET /api/v1/analytics/stats - Get stats

- Test HTTP methods, status codes, response bodies
- Test error cases (400, 404, 500)
- Test validation
- Test CORS headers

STEP 2 (GREEN): Implement handlers
- Create internal/api/server.go
- Create internal/api/handlers.go
- Set up Chi router
- Add middleware (logging, CORS, recovery)
- Implement each handler
- Add request/response validation

STEP 3 (REFACTOR): Add features
- Add pagination
- Add sorting
- Add filtering
- Add rate limiting
- Add OpenAPI/Swagger docs

Example test:

func TestCreateIdeaHandler(t *testing.T) {
    repo := setupTestDB(t)
    defer repo.Close()

    server := NewServer(repo, testConfig)

    body := `{"title": "Test Idea", "description": "Test"}`
    req := httptest.NewRequest("POST", "/api/v1/ideas", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()
    server.Router().ServeHTTP(w, req)

    assert.Equal(t, http.StatusCreated, w.Code)

    var idea models.Idea
    json.NewDecoder(w.Body).Decode(&idea)
    assert.Equal(t, "Test Idea", idea.Title)
    assert.Greater(t, idea.Score, 0.0)
}

OpenAPI documentation:
- Use swaggo/swag to generate from comments
- Or write docs/api.yaml manually
- Generate TypeScript types for frontend

Run: go test ./internal/api -v -cover
Start server: go run ./cmd/web
Test: curl http://localhost:8080/api/v1/ideas
```

---

### Phase 4: SvelteKit Frontend (Week 5-7)

**Goal:** Beautiful, responsive web UI

**Prompt for AI Assistant:**

```
Phase 4: Implement SvelteKit frontend with component tests

Setup:
cd web
npm create svelte@latest .
npm install
npm install -D tailwindcss @skeletonlabs/skeleton
npm install lucide-svelte @tanstack/svelte-query
npm install -D @playwright/test

STEP 1: Set up infrastructure
- Configure TailwindCSS with Skeleton UI theme
- Set up API client with TypeScript types
- Configure routing
- Set up state management (SvelteQuery)

STEP 2: Build components with tests
For each component:
- Write component tests (Vitest)
- Implement component
- Test interactions
- Test accessibility

Components to build:
1. IdeaCard.svelte
   - Test: Renders idea correctly
   - Test: Shows score badge with correct color
   - Test: Displays detected patterns
   - Test: Delete button works
   - Test: Edit button navigates

2. IdeaForm.svelte
   - Test: Form validation
   - Test: Submit creates idea
   - Test: Loading state
   - Test: Error handling

3. Dashboard.svelte
   - Test: Loads ideas on mount
   - Test: Filters work
   - Test: Sorting works
   - Test: Pagination works

4. Analytics.svelte
   - Test: Displays stats correctly
   - Test: Charts render
   - Test: Trends show

STEP 3: Build pages
- / (Dashboard)
- /ideas/[id] (Idea detail)
- /ideas/[id]/edit (Edit idea)
- /analytics (Analytics)
- /settings (Settings)

STEP 4: E2E tests (Playwright)
- Test full user flows:
  * Create idea → See it in list → Edit → Delete
  * Filter ideas by score
  * Search ideas
  * View analytics

Example component test:

// IdeaCard.test.ts
import { render, fireEvent } from '@testing-library/svelte';
import IdeaCard from './IdeaCard.svelte';

test('renders idea card', () => {
    const idea = {
        id: '123',
        title: 'Test Idea',
        score: 8.5,
        created_at: new Date().toISOString(),
        status: 'pending',
        tags: ['test']
    };

    const { getByText } = render(IdeaCard, { props: { idea } });

    expect(getByText('Test Idea')).toBeInTheDocument();
    expect(getByText('8.5')).toBeInTheDocument();
});

test('delete button calls onDelete', async () => {
    const onDelete = vi.fn();
    const idea = { /* ... */ };

    const { getByText } = render(IdeaCard, {
        props: { idea, onDelete }
    });

    await fireEvent.click(getByText('Delete'));

    expect(onDelete).toHaveBeenCalledWith(idea.id);
});

Example E2E test:

// tests/ideas.spec.ts
test('create and manage idea', async ({ page }) => {
    await page.goto('/');

    // Create idea
    await page.fill('input[placeholder="What\'s your idea?"]', 'Test Idea');
    await page.click('button:has-text("Capture Idea")');

    // Verify created
    await expect(page.locator('text=Test Idea')).toBeVisible();
    await expect(page.locator('.score-badge')).toBeVisible();

    // Edit idea
    await page.click('button:has-text("Edit")');
    await page.fill('input[name="title"]', 'Updated Idea');
    await page.click('button:has-text("Save")');

    // Verify updated
    await expect(page.locator('text=Updated Idea')).toBeVisible();

    // Delete idea
    await page.click('button:has-text("Delete")');
    await page.click('button:has-text("Confirm")');

    // Verify deleted
    await expect(page.locator('text=Updated Idea')).not.toBeVisible();
});

Run tests:
npm test                    # Unit/component tests
npx playwright test         # E2E tests
npm run build               # Production build
```

---

### Phase 5: Integration & Polish (Week 7-8)

**Prompt for AI Assistant:**

```
Phase 5: Integration testing, performance, security, documentation

Tasks:

1. End-to-End Integration Tests
   - Test full stack together (Go API + SvelteKit)
   - Test data migration from Rust version
   - Test CLI + Web simultaneously accessing same database
   - Load testing with many concurrent users

2. Performance Optimization
   Backend:
   - Add database indexes
   - Implement caching (in-memory for hot data)
   - Add gzip compression
   - Optimize slow queries

   Frontend:
   - Lazy load routes
   - Implement virtual scrolling
   - Optimize bundle size
   - Add service worker

   Run benchmarks:
   go test -bench=. ./...
   lighthouse http://localhost:5173

3. Security Audit
   - [ ] CSRF protection
   - [ ] Input sanitization
   - [ ] Rate limiting
   - [ ] SQL injection prevention (parameterized queries)
   - [ ] XSS prevention
   - [ ] Security headers
   - [ ] Dependency vulnerability scan

   Run:
   go install github.com/securego/gosec/v2/cmd/gosec@latest
   gosec ./...
   npm audit

4. Documentation
   Create comprehensive docs:
   - README.md (overview, quick start)
   - docs/CLI.md (all commands with examples)
   - docs/API.md (all endpoints, OpenAPI spec)
   - docs/DEVELOPMENT.md (setup, architecture, contributing)
   - docs/DEPLOYMENT.md (Docker, cloud deployment)
   - docs/MIGRATION.md (migrating from Rust version)
   - CHANGELOG.md (version history)

5. Docker & Deployment
   - Test multi-stage Dockerfile builds
   - Test docker-compose setup
   - Create deployment scripts
   - Set up health checks
   - Add monitoring/logging

   Test:
   docker-compose build
   docker-compose up
   curl http://localhost:8080/health
```

---

### Phase 6: Beta Release (Week 8)

**Prompt for AI Assistant:**

```
Phase 6: Beta deployment and user testing

Tasks:

1. Pre-deployment checklist
   - [ ] All tests passing (>80% coverage)
   - [ ] Security audit complete
   - [ ] Performance benchmarks acceptable
   - [ ] Documentation complete
   - [ ] Migration guide tested
   - [ ] Docker images built and tested

2. Deployment
   - Deploy to staging environment
   - Run smoke tests
   - Import sample data
   - Verify all features work
   - Monitor logs and metrics

3. User Acceptance Testing
   - Recruit beta testers
   - Provide migration guide
   - Collect feedback via:
     * GitHub Issues
     * Survey
     * Direct interviews
   - Monitor for bugs and errors

4. Bug Fixes & Iteration
   - Prioritize critical bugs
   - Fix and deploy patches
   - Update documentation based on feedback
   - Prepare for v1.0 release

5. Go/No-Go Decision
   Decision criteria:
   - All critical bugs fixed
   - User feedback positive (>80% satisfied)
   - Performance acceptable
   - Security audit passed
   - Documentation complete

   If GO: Prepare v1.0 release
   If NO-GO: Identify blockers, extend beta
```

---

## TDD Best Practices

### Test Organization

```
internal/
  scoring/
    engine.go           # Implementation
    engine_test.go      # Unit tests
    testdata/           # Test fixtures
      high_score_idea.json
      test_telos.md
```

### Test Naming Convention

```go
// Format: Test<Function>_<Scenario>_<ExpectedResult>
func TestScoreIdea_HighAlignment_ReturnsHighScore(t *testing.T) { }
func TestScoreIdea_NoGoals_ReturnsNeutralScore(t *testing.T) { }
func TestParseFile_ValidTelos_ParsesSuccessfully(t *testing.T) { }
func TestParseFile_MissingFile_ReturnsError(t *testing.T) { }
```

### Coverage Goals

```
Component               Coverage Target
--------------------------------------
Models                  >90%
Core Business Logic     >95%
Database Layer          >80%
API Handlers            >80%
CLI Commands            >75%
Utils/Helpers           >85%
--------------------------------------
Overall                 >85%
```

### Running Tests

```bash
# All tests
make test

# With coverage
make test-coverage

# Specific package
go test ./internal/scoring -v -cover

# Integration tests only
go test ./test/integration -v

# Benchmarks
go test -bench=. ./internal/scoring

# Watch mode (with air)
air test ./...
```

---

## Common Patterns

### Pattern 1: Table-Driven Tests

```go
func TestCalculateMissionAlignment(t *testing.T) {
    tests := []struct {
        name     string
        idea     *models.Idea
        telos    *models.Telos
        expected float64
    }{
        {
            name: "perfect alignment",
            idea: &models.Idea{Title: "Build SaaS product"},
            telos: &models.Telos{
                Goals: []models.Goal{
                    {Description: "Launch SaaS product"},
                },
            },
            expected: 1.0,
        },
        {
            name: "no alignment",
            idea: &models.Idea{Title: "Learn underwater basket weaving"},
            telos: &models.Telos{
                Goals: []models.Goal{
                    {Description: "Build software products"},
                },
            },
            expected: 0.0,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            engine := NewEngine(tt.telos)
            result := engine.calculateMissionAlignment(tt.idea)
            assert.InDelta(t, tt.expected, result, 0.1)
        })
    }
}
```

### Pattern 2: Test Fixtures

```go
// testdata_test.go
func loadTestTelos(t *testing.T, filename string) *models.Telos {
    t.Helper()
    parser := NewParser()
    telos, err := parser.ParseFile(filepath.Join("testdata", filename))
    require.NoError(t, err)
    return telos
}

// In tests:
func TestSomething(t *testing.T) {
    telos := loadTestTelos(t, "test_telos.md")
    // Use telos...
}
```

### Pattern 3: Test Database

```go
// test_helpers.go
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

// In tests:
func TestDatabaseOperations(t *testing.T) {
    repo, cleanup := setupTestDB(t)
    defer cleanup()

    // Use repo...
}
```

---

## Validation Checklist

Before considering each phase complete, verify:

### Phase 1: Core Domain
- [ ] All unit tests passing
- [ ] Coverage >90% for business logic
- [ ] Scoring matches Rust output (validated with fixtures)
- [ ] Database schema compatible with Rust version
- [ ] Benchmarks show acceptable performance

### Phase 2: CLI
- [ ] All commands implemented and tested
- [ ] Help text complete and accurate
- [ ] Error messages helpful
- [ ] Output format matches (or improves) Rust version
- [ ] Manual testing with real telos.md works

### Phase 3: API
- [ ] All endpoints tested
- [ ] OpenAPI documentation complete
- [ ] CORS configured correctly
- [ ] Error handling consistent
- [ ] Integration tests passing

### Phase 4: Frontend
- [ ] All components tested
- [ ] E2E tests covering main flows
- [ ] Responsive design works on mobile/tablet/desktop
- [ ] Accessibility tested (keyboard nav, screen readers)
- [ ] Performance acceptable (Lighthouse score >90)

### Phase 5: Integration
- [ ] Full stack integration tests passing
- [ ] Security audit complete
- [ ] Documentation reviewed and complete
- [ ] Docker deployment tested
- [ ] Migration from Rust tested with real data

---

## Example Usage

### Starting a New Phase

```bash
# Copy this prompt and fill in the details:

"I'm starting [Phase X: Name] of the Go migration for Telos Idea Matrix.

Current status:
- Completed phases: [list]
- Current task: [specific task from plan]

Please help me implement [specific component] using TDD:

1. First, help me write comprehensive tests for [component]
   - Test cases should cover: [list scenarios]
   - Reference Rust implementation at: [file path]
   - Use fixtures from: [testdata path]

2. Then, help me implement the minimal code to make tests pass

3. Finally, help me refactor and optimize

Constraints:
- Must maintain feature parity with Rust version
- Target test coverage: >X%
- Must pass golangci-lint with no errors

Ready to start? Please begin with the test file."
```

### Debugging Failing Tests

```bash
"Tests are failing for [component].

Current test output:
[paste test output]

Expected behavior (from Rust version):
[describe expected behavior]

Please help me:
1. Identify what's wrong
2. Fix the implementation
3. Ensure tests pass
4. Maintain code quality"
```

### Code Review Request

```bash
"I've implemented [component]. Please review for:

1. Test coverage - is it sufficient?
2. Code quality - any improvements?
3. Performance - any concerns?
4. Security - any vulnerabilities?
5. Consistency - matches project patterns?

Files:
- Implementation: [path]
- Tests: [path]

Test results:
[paste test output with coverage]"
```

---

## Success Criteria (Final)

The migration is successful when:

✅ **Functionality**
- [ ] All Rust CLI commands work in Go version
- [ ] Scoring algorithm produces identical results
- [ ] Database migration completes without data loss
- [ ] Web UI provides all CLI functionality plus more

✅ **Quality**
- [ ] Test coverage >85% overall
- [ ] No critical security vulnerabilities
- [ ] Performance acceptable (API <100ms p95)
- [ ] No memory leaks

✅ **Documentation**
- [ ] README clear and complete
- [ ] API docs generated (OpenAPI)
- [ ] Migration guide tested by users
- [ ] Contributing guide helps new developers

✅ **User Acceptance**
- [ ] Beta testers can migrate successfully
- [ ] >80% satisfaction with new version
- [ ] No critical bugs reported
- [ ] Performance meets or exceeds Rust version

✅ **Deployment**
- [ ] Docker images build successfully
- [ ] CI/CD pipeline working
- [ ] Staging deployment successful
- [ ] Rollback plan tested

---

## Quick Reference: Key Commands

```bash
# Development
make dev-cli          # Run CLI in watch mode
make dev-api          # Run API server in watch mode
cd web && npm run dev # Run SvelteKit dev server

# Testing
make test             # All Go tests
make test-coverage    # With coverage report
cd web && npm test    # Frontend tests
cd web && npx playwright test  # E2E tests

# Building
make build            # Build all binaries
make build-cli        # CLI only
make build-api        # API server only
cd web && npm run build  # Frontend build

# Docker
docker-compose build  # Build images
docker-compose up     # Start services
docker-compose down   # Stop services

# Quality
make lint            # Run linters
make fmt             # Format code
gosec ./...          # Security scan
```

---

**This is your guide. Follow it phase by phase, use TDD religiously, and you'll have a production-ready Go + SvelteKit application in 6-8 weeks.**

Good luck! 🚀
