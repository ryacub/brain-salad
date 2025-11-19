# Migration Phase Prompts

This directory contains copy-paste ready prompts for executing each phase of the Go migration using subagents.

## How to Use

1. **Read** `../SUBAGENT_ORCHESTRATION.md` for the complete execution strategy
2. **Copy** the prompt for the phase you're working on
3. **Launch** a subagent with the copied prompt
4. **Validate** deliverables before moving to next phase

## Available Phases

### [Phase 0: Preparation](phase-0-preparation.md)
**Duration:** 3-5 days
- Set up Go project structure
- Configure CI/CD
- Document Rust behavior (`RUST_REFERENCE.md`)

### [Phase 1: Core Domain](phase-1-core-domain.md)
**Duration:** 7-10 days
- Implement data models
- Port Telos parser
- Port scoring engine (CRITICAL)
- Implement database layer
- **TDD Required:** Write tests FIRST

### [Phase 2: CLI Implementation](phase-2-cli.md)
**Duration:** 5-7 days
- Build all CLI commands with Cobra
- Wire to core domain logic
- Feature parity with Rust CLI

### [Phase 3: API Server](phase-3-api.md)
**Duration:** 5-7 days
- Build RESTful API with Chi
- All CRUD endpoints
- OpenAPI documentation
- Can run in parallel with Phase 2

### [Phase 4: SvelteKit Frontend](phase-4-frontend.md)
**Duration:** 10-14 days
- Build beautiful web UI
- SvelteKit + Tailwind + Skeleton UI
- E2E tests with Playwright

### [Phase 5: Integration & Polish](phase-5-integration.md)
**Duration:** 5-7 days
- End-to-end testing
- Performance optimization
- Security audit
- Documentation
- Docker deployment

### [Phase 6: Beta Release](phase-6-beta.md)
**Duration:** 5-7 days
- Deploy to staging
- User acceptance testing
- Bug fixes
- v1.0 release decision

## Execution Tips

### Sequential Execution (Recommended)
Execute phases in order: 0 → 1 → 2 → 3 → 4 → 5 → 6

### Parallel Execution (Advanced)
- Phases 2 (CLI) and 3 (API) can run in parallel after Phase 1
- Both depend only on Phase 1 (Core Domain)
- Merge carefully after both complete

### Validation Between Phases

Always run before proceeding:
```bash
make test           # All tests pass
make test-coverage  # Coverage meets target
make lint           # No linter errors
make build          # Builds successfully
```

### TDD Reminder

**CRITICAL:** Write tests FIRST for all code (RED → GREEN → REFACTOR)

1. 🔴 RED: Write failing test
2. 🟢 GREEN: Implement minimal code to pass
3. ♻️ REFACTOR: Improve quality while keeping tests green

## Progress Tracking

Create a `MIGRATION_PROGRESS.md` file to track your progress:

```markdown
# Migration Progress

## Phase 0: Preparation ✅
Completed: 2025-01-18
Notes: CI/CD working, RUST_REFERENCE.md complete

## Phase 1: Core Domain 🚧
Started: 2025-01-19
Status: Models complete, working on scoring engine
Coverage: 87%

## Phase 2: CLI ⏳
Not started

[etc.]
```

## Getting Help

If stuck:
1. Review the comprehensive plan: `../GO_MIGRATION_PLAN.md`
2. Check the orchestration guide: `../SUBAGENT_ORCHESTRATION.md`
3. Reference Rust behavior: `RUST_REFERENCE.md` (created in Phase 0)
4. Review language analysis: `../LANGUAGE_ANALYSIS.md`

## Quick Start

```bash
# 1. Read the orchestration guide
cat ../SUBAGENT_ORCHESTRATION.md

# 2. Copy Phase 0 prompt
cat phase-0-preparation.md

# 3. Launch subagent with the prompt
# (Use Task tool or copy-paste to new conversation)

# 4. Execute the phase following TDD

# 5. Validate before moving on
make test && make lint && make build

# 6. Move to next phase
cat phase-1-core-domain.md
```

---

**Ready to start? Begin with Phase 0!**

`cat phase-0-preparation.md`
