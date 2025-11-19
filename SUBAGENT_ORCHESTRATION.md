# Subagent Orchestration Guide: Go Migration

**Purpose:** Execute the Telos Idea Matrix migration from Rust to Go using specialized subagents for each phase.

**Timeline:** 6-8 weeks (8 phases, run sequentially or with controlled parallelism)

---

## Overview

This migration is broken into **8 independent phases**, each executable by a specialized subagent. Each phase:

- Has a **dedicated prompt** in `migration-prompts/`
- Is **self-contained** with all necessary context
- Produces **clear deliverables** that feed into the next phase
- Follows **Test-Driven Development** (write tests first)
- Has **validation criteria** before moving forward

---

## Execution Strategy

### Sequential Execution (Recommended for Solo Developer)

Execute one phase at a time, validate outputs, then proceed:

```
Phase 0 → Validate → Phase 1 → Validate → Phase 2 → ... → Phase 6
```

### Parallel Execution (Advanced)

Some phases can run in parallel once dependencies are met:

```
Phase 0 (Foundation)
    ↓
Phase 1 (Core Domain) ← Must complete first
    ↓
Phase 2 (CLI) ⟋ ⟍ Phase 3 (API)  ← Can run in parallel
    ↓           ↓
    └─────┬─────┘
          ↓
Phase 4 (Frontend) ← Depends on Phase 3 API
          ↓
Phase 5 (Integration)
          ↓
Phase 6 (Beta)
```

---

## Phase Breakdown

### Phase 0: Preparation & Foundation
**Duration:** 3-5 days
**Subagent Type:** `general-purpose`
**Prompt:** `migration-prompts/phase-0-preparation.md`

**Goals:**
- Set up Go project structure
- Configure CI/CD pipeline
- Document Rust behavior (reference implementation)
- Create development environment
- Set up testing infrastructure

**Deliverables:**
- [ ] Complete Go project structure (cmd/, internal/, pkg/)
- [ ] Working Makefile with build/test/lint targets
- [ ] GitHub Actions CI/CD pipeline
- [ ] `RUST_REFERENCE.md` documenting current behavior
- [ ] Development environment documented in `docs/DEVELOPMENT.md`
- [ ] `.golangci.yml` linter configuration
- [ ] Initial `go.mod` with dependencies

**Validation:**
```bash
make build        # Should build successfully
make test         # Should run (no tests yet, but framework works)
make lint         # Should pass (no code yet)
git log --oneline # Shows commits for setup
```

**Handoff to Next Phase:**
- Project structure ready for implementation
- Reference documentation complete
- CI/CD validating all pushes

---

### Phase 1: Core Domain Migration (TDD)
**Duration:** 7-10 days
**Subagent Type:** `general-purpose`
**Prompt:** `migration-prompts/phase-1-core-domain.md`

**Goals:**
- Implement data models (Idea, Telos, Analysis, Pattern)
- Port Telos parser (markdown → struct)
- Port scoring engine (CRITICAL: must match Rust exactly)
- Port pattern detector
- Implement database layer (SQLite)

**TDD Requirements:**
- Write tests FIRST for every component
- Coverage targets: >90% for models, >95% for scoring
- Compare output with Rust for validation

**Deliverables:**
- [ ] `internal/models/*.go` - All data structures
- [ ] `internal/telos/parser.go` - Telos markdown parser
- [ ] `internal/scoring/engine.go` - Scoring algorithm
- [ ] `internal/patterns/detector.go` - Pattern detection
- [ ] `internal/database/repository.go` - SQLite CRUD operations
- [ ] Comprehensive test suite (>90% coverage)
- [ ] Test fixtures in `testdata/`

**Validation:**
```bash
go test ./internal/... -v -cover
# All tests pass
# Coverage >90% overall

# Scoring validation
go test ./internal/scoring -v
# Compare output with Rust using same test ideas
# Scores should match within 0.1 points
```

**Handoff to Next Phase:**
- Core business logic fully functional
- Scoring algorithm validated against Rust
- Database layer working with migrations
- All tests passing

---

### Phase 2: CLI Implementation (Cobra)
**Duration:** 5-7 days
**Subagent Type:** `general-purpose`
**Prompt:** `migration-prompts/phase-2-cli.md`

**Dependencies:** Phase 1 complete

**Goals:**
- Implement all CLI commands using Cobra
- Wire commands to core domain logic
- Add colored output and UX polish
- Achieve feature parity with Rust CLI

**Commands to Implement:**
- `tm dump` - Capture idea with analysis
- `tm analyze` - Analyze existing or new idea
- `tm score` - Quick score without saving
- `tm review` - Browse/filter ideas
- `tm prune` - Clean up old ideas
- `tm analytics` - View statistics
- `tm link` - Manage idea relationships

**Deliverables:**
- [ ] `internal/cli/*.go` - All command implementations
- [ ] `cmd/cli/main.go` - CLI entry point
- [ ] Colored output (github.com/fatih/color)
- [ ] Interactive prompts where appropriate
- [ ] Complete help text for all commands
- [ ] CLI tests (>80% coverage)

**Validation:**
```bash
go build -o bin/tm ./cmd/cli

# Test each command
./bin/tm dump "Build a Go CLI tool"
./bin/tm review --min-score 7.0
./bin/tm analyze --last
./bin/tm prune --dry-run

# Compare with Rust CLI
# Ensure feature parity
```

**Handoff to Next Phase:**
- Fully functional CLI
- Feature parity with Rust version
- Excellent UX (colors, help text)
- Ready for users to test

---

### Phase 3: API Server (Chi Router)
**Duration:** 5-7 days
**Subagent Type:** `general-purpose`
**Prompt:** `migration-prompts/phase-3-api.md`

**Dependencies:** Phase 1 complete
**Can Run Parallel With:** Phase 2

**Goals:**
- Build RESTful API server using Chi
- Implement all CRUD endpoints
- Add middleware (CORS, logging, recovery)
- Generate OpenAPI documentation
- Enable SvelteKit frontend integration

**Endpoints to Implement:**
```
GET    /api/v1/ideas           - List ideas (with filters)
POST   /api/v1/ideas           - Create idea
GET    /api/v1/ideas/:id       - Get idea by ID
PUT    /api/v1/ideas/:id       - Update idea
DELETE /api/v1/ideas/:id       - Delete idea
POST   /api/v1/analyze         - Analyze text (no save)
GET    /api/v1/analytics/stats - Statistics
GET    /api/v1/analytics/trends - Trends
GET    /health                 - Health check
```

**Deliverables:**
- [ ] `internal/api/server.go` - Server setup
- [ ] `internal/api/handlers.go` - All endpoint handlers
- [ ] `internal/api/middleware/` - CORS, logging, etc.
- [ ] `cmd/web/main.go` - API server entry point
- [ ] `docs/api.yaml` - OpenAPI specification
- [ ] API tests (>85% coverage)
- [ ] Integration tests with database

**Validation:**
```bash
go build -o bin/tm-web ./cmd/web
./bin/tm-web &

# Test all endpoints
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/ideas
curl -X POST http://localhost:8080/api/v1/ideas \
  -H "Content-Type: application/json" \
  -d '{"title":"Test","description":"Testing API"}'

# Check OpenAPI docs
open http://localhost:8080/swagger
```

**Handoff to Next Phase:**
- Working API server
- All endpoints functional
- OpenAPI spec complete (for TypeScript types)
- CORS configured for frontend

---

### Phase 4: SvelteKit Frontend
**Duration:** 10-14 days
**Subagent Type:** `general-purpose`
**Prompt:** `migration-prompts/phase-4-frontend.md`

**Dependencies:** Phase 3 complete (API must be running)

**Goals:**
- Build beautiful web UI with SvelteKit
- Implement all CRUD operations
- Add filtering, sorting, analytics
- Mobile-responsive design
- E2E tests with Playwright

**Pages to Build:**
- `/` - Dashboard (list ideas, filters)
- `/ideas/[id]` - Idea detail view
- `/ideas/[id]/edit` - Edit idea
- `/analytics` - Charts and statistics
- `/settings` - Configuration

**Deliverables:**
- [ ] `web/src/routes/` - All pages
- [ ] `web/src/lib/components/` - Reusable components
- [ ] `web/src/lib/api/client.ts` - Type-safe API client
- [ ] `web/src/lib/api/types.ts` - TypeScript types (from OpenAPI)
- [ ] Tailwind CSS + Skeleton UI configured
- [ ] Responsive design (mobile/tablet/desktop)
- [ ] Component tests (Vitest)
- [ ] E2E tests (Playwright)

**Validation:**
```bash
cd web
npm install
npm run dev

# Open browser: http://localhost:5173
# Test all user flows:
# - Create idea
# - View idea list
# - Filter by score
# - Edit idea
# - Delete idea
# - View analytics

# Run tests
npm test                 # Component tests
npx playwright test      # E2E tests

# Build for production
npm run build
```

**Handoff to Next Phase:**
- Beautiful, functional web UI
- All features working
- Tests passing
- Production build succeeds

---

### Phase 5: Integration & Polish
**Duration:** 5-7 days
**Subagent Type:** `general-purpose`
**Prompt:** `migration-prompts/phase-5-integration.md`

**Dependencies:** Phases 2, 3, 4 complete

**Goals:**
- End-to-end integration testing
- Performance optimization
- Security audit
- Documentation
- Docker deployment

**Tasks:**
1. **Integration Testing**
   - Test CLI + API + Web together
   - Test concurrent access
   - Test data migration from Rust
   - Load testing

2. **Performance**
   - Database indexing
   - API response caching
   - Frontend bundle optimization
   - Benchmark critical paths

3. **Security**
   - SQL injection prevention
   - XSS protection
   - CSRF tokens
   - Rate limiting
   - Dependency vulnerability scan

4. **Documentation**
   - User guide (README.md)
   - API documentation
   - CLI reference
   - Development guide
   - Migration guide from Rust

5. **Deployment**
   - Multi-stage Dockerfile
   - docker-compose.yml
   - Deployment scripts
   - Health checks

**Deliverables:**
- [ ] End-to-end test suite
- [ ] Performance benchmarks (documented)
- [ ] Security audit report
- [ ] Complete documentation
- [ ] Docker images (CLI + Web)
- [ ] Deployment guide

**Validation:**
```bash
# Integration tests
make test-integration

# Performance
go test -bench=. ./...
lighthouse http://localhost:5173

# Security
gosec ./...
npm audit

# Docker
docker-compose build
docker-compose up
curl http://localhost:8080/health

# Documentation
# Review all docs for completeness
```

**Handoff to Next Phase:**
- Production-ready system
- Comprehensive documentation
- Docker deployment working
- All quality gates passing

---

### Phase 6: Beta Release & User Testing
**Duration:** 5-7 days
**Subagent Type:** `general-purpose`
**Prompt:** `migration-prompts/phase-6-beta.md`

**Dependencies:** Phase 5 complete

**Goals:**
- Deploy beta to staging
- Recruit beta testers
- Collect feedback
- Fix critical bugs
- Prepare for v1.0 release

**Tasks:**
1. **Beta Deployment**
   - Deploy to staging environment
   - Smoke test all features
   - Monitor logs and metrics

2. **User Testing**
   - Recruit 5-10 beta testers
   - Provide migration guide
   - Collect structured feedback

3. **Bug Fixes**
   - Triage reported issues
   - Fix critical bugs
   - Prioritize enhancements

4. **Go/No-Go Decision**
   - Review feedback
   - Check success criteria
   - Decide on v1.0 release

**Deliverables:**
- [ ] Beta deployment (staging URL)
- [ ] Beta tester feedback report
- [ ] Bug fix commits
- [ ] v1.0 release plan
- [ ] Migration guide (tested by users)

**Validation:**
```bash
# Deployment health
curl https://staging.example.com/health

# User acceptance
# >80% testers successfully migrated
# >80% satisfaction with new version
# No critical bugs reported

# Success criteria met
# All tests passing
# Performance acceptable
# Security audit passed
```

**Completion:**
- Beta successful
- Critical bugs fixed
- Ready for v1.0 release announcement

---

## How to Execute Each Phase

### Step 1: Copy the Phase Prompt

```bash
# Copy the prompt for the phase you're working on
cat migration-prompts/phase-0-preparation.md
```

### Step 2: Launch Subagent

Use the Task tool to launch a subagent with the copied prompt:

```markdown
I need you to execute Phase [N] of the Telos Idea Matrix Go migration.

**Context:**
[Paste the phase prompt here from migration-prompts/phase-N-*.md]

**Current Status:**
- Previous phases completed: [list]
- Codebase location: /home/user/brain-salad (Rust) and /home/user/telos-idea-matrix-go (Go)
- Reference docs: LANGUAGE_ANALYSIS.md, GO_MIGRATION_PLAN.md

**Your Task:**
Execute this phase following TDD principles:
1. Write tests FIRST (RED)
2. Implement minimal code (GREEN)
3. Refactor while keeping tests green (REFACTOR)

**Deliverables:**
[List from phase prompt]

**Validation:**
[Validation commands from phase prompt]

Ready to begin? Start with [first task from phase].
```

### Step 3: Validate Output

After subagent completes, validate deliverables:

```bash
# Run validation commands from phase
make test
make lint
make build

# Check coverage
go test -cover ./...

# Manual testing if needed
```

### Step 4: Commit & Document

```bash
# Commit phase work
git add .
git commit -m "Complete Phase [N]: [Phase Name]

Deliverables:
- [List key deliverables]

Tests: All passing
Coverage: [X]%
Status: Ready for next phase"

git push
```

### Step 5: Move to Next Phase

Once validated, proceed to the next phase.

---

## Parallel Execution Strategy

For faster completion, some phases can run in parallel:

### Parallel Set 1: CLI + API (After Phase 1)

**Terminal 1: Phase 2 (CLI)**
```bash
# Launch subagent for Phase 2
# Both CLI and API depend only on Phase 1 (Core Domain)
# They don't depend on each other
```

**Terminal 2: Phase 3 (API)**
```bash
# Launch subagent for Phase 3
# Can run simultaneously with Phase 2
```

**Merge:**
Once both complete, merge the branches and resolve any conflicts (should be minimal since they work on different files).

### Sequential Requirement

These MUST run sequentially:
- **Phase 0** → Everything depends on this
- **Phase 1** → CLI/API depend on this
- **Phase 4** → Depends on Phase 3 (needs API running)
- **Phase 5** → Depends on Phases 2, 3, 4 (integration)
- **Phase 6** → Depends on Phase 5 (deployment)

---

## Troubleshooting

### Subagent Gets Stuck

**Symptom:** Subagent doesn't make progress or asks for clarification

**Solution:**
1. Check if phase prompt has all necessary context
2. Provide missing information explicitly
3. Break task into smaller subtasks
4. Reference Rust implementation directly

### Tests Failing

**Symptom:** Tests don't pass after implementation

**Solution:**
1. Review test expectations (are they correct?)
2. Compare behavior with Rust version
3. Check test fixtures and data
4. Run tests with `-v` for detailed output
5. Use debugger (Delve for Go)

### Coverage Too Low

**Symptom:** Coverage below target (e.g., <85%)

**Solution:**
1. Identify untested code: `go test -coverprofile=coverage.out ./...`
2. Open coverage HTML: `go tool cover -html=coverage.out`
3. Write additional tests for uncovered lines
4. Focus on edge cases and error paths

### Integration Issues

**Symptom:** Components don't work together

**Solution:**
1. Review interface contracts
2. Check data types match between components
3. Add integration tests
4. Test with real data (not just mocks)

---

## Quality Gates

Before marking any phase as complete:

### Required Checks

```bash
# All tests pass
make test
✓ All tests passing

# Coverage meets target
make test-coverage
✓ Coverage ≥ [phase target]%

# No linter errors
make lint
✓ golangci-lint passes

# Code formatted
gofmt -l .
✓ No unformatted files

# Security scan
gosec ./...
✓ No critical vulnerabilities

# Builds successfully
make build
✓ Binary builds without errors
```

### Phase-Specific Gates

Each phase has additional validation criteria listed in its prompt.

---

## Progress Tracking

### Create a Tracking Document

```markdown
# Migration Progress

## Phase 0: Preparation ✅
- Started: 2025-01-15
- Completed: 2025-01-18
- Deliverables: All ✓
- Notes: CI/CD configured with GitHub Actions

## Phase 1: Core Domain ✅
- Started: 2025-01-19
- Completed: 2025-01-28
- Coverage: 92%
- Notes: Scoring validated against Rust (100% match)

## Phase 2: CLI 🚧
- Started: 2025-01-29
- Completed: [In Progress]
- Status: 6/7 commands implemented
- Next: Implement `tm analytics` command

## Phase 3: API ⏳
- Started: [Not started]
- Estimated start: 2025-02-01

[etc.]
```

### Update After Each Session

Keep this updated so you always know where you are.

---

## Success Metrics

### Technical Metrics

- ✅ All tests passing
- ✅ Coverage ≥85% overall
- ✅ No critical security vulnerabilities
- ✅ API response time <100ms (p95)
- ✅ Frontend Lighthouse score >90
- ✅ Docker images build successfully

### Functional Metrics

- ✅ Feature parity with Rust version
- ✅ Scoring output matches Rust (within 0.1)
- ✅ Database migration successful (no data loss)
- ✅ CLI commands work identically
- ✅ Web UI provides all CLI features plus more

### User Metrics (Phase 6)

- ✅ >80% beta testers successfully migrate
- ✅ >80% satisfaction with new version
- ✅ <5 critical bugs reported
- ✅ Performance meets or exceeds Rust version

---

## Timeline Summary

```
Week 1     | Phase 0: Preparation
────────────────────────────────────────
Week 2-3   | Phase 1: Core Domain (TDD)
────────────────────────────────────────
Week 3-4   | Phase 2: CLI Implementation
           | Phase 3: API Server (parallel)
────────────────────────────────────────
Week 5-7   | Phase 4: SvelteKit Frontend
────────────────────────────────────────
Week 7-8   | Phase 5: Integration & Polish
────────────────────────────────────────
Week 8     | Phase 6: Beta Release
────────────────────────────────────────
Total: 6-8 weeks to production-ready v2.0
```

---

## Next Steps

1. **Review** this orchestration guide
2. **Create** progress tracking document
3. **Start** with Phase 0:
   ```bash
   cat migration-prompts/phase-0-preparation.md
   # Copy prompt and launch subagent
   ```
4. **Execute** each phase sequentially
5. **Validate** deliverables before proceeding
6. **Track** progress in your tracking document

---

## Resources

- **Migration Plan:** `GO_MIGRATION_PLAN.md` (detailed technical spec)
- **Language Analysis:** `LANGUAGE_ANALYSIS.md` (why Go?)
- **Phase Prompts:** `migration-prompts/phase-*.md` (copy-paste ready)
- **Rust Reference:** Create `RUST_REFERENCE.md` in Phase 0

---

**Let's ship this migration! 🚀**

Follow TDD, validate religiously, and you'll have a production-ready Go + SvelteKit application in 6-8 weeks.
