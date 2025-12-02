# Brain Salad Pruning Plan

> Ruthless codebase simplification while preserving all functionality

**Created:** 2025-12-02
**Status:** Proposed
**Estimated Reduction:** ~3,700 LOC (-16%)

---

## Executive Summary

This plan identifies opportunities to reduce codebase complexity by:
- Deleting unused/redundant packages
- Removing deprecated legacy commands
- Consolidating duplicate code
- Simplifying architecture

**Guiding Principle:** If it's not actively used or adds no unique value, delete it.

---

## Phase 1: Delete Unused Packages

**Target:** ~1,170 LOC reduction

### 1.1 `internal/tasks/` - DELETE ENTIRE PACKAGE

| File | LOC | Status |
|------|-----|--------|
| `task.go` | ~80 | Delete |
| `manager.go` | ~150 | Delete |
| `scheduler.go` | ~120 | Delete |
| `manager_test.go` | ~100 | Delete |

**Reason:** Only used by `cmd/web/main.go` for 3 trivial scheduled tasks:
- Database VACUUM (1 line of SQL)
- Connection stats logging (3 lines)
- Health ping (1 line)

**Action:** Inline these directly in `cmd/web/main.go` using simple `time.Ticker`.

```go
// Replace complex task manager with simple ticker
go func() {
    ticker := time.NewTicker(24 * time.Hour)
    for range ticker.C {
        repo.DB().Exec("VACUUM")
    }
}()
```

---

### 1.2 `internal/health/` - DELETE ENTIRE PACKAGE

| File | LOC | Status |
|------|-----|--------|
| `monitor.go` | ~166 | Delete |
| `checkers.go` | ~161 | Delete |
| `monitor_test.go` | ~100 | Delete |
| `checkers_test.go` | ~100 | Delete |

**Reason:** Over-engineered health check framework. Used by:
- `internal/cli/status.go` - Can inline simple checks
- `internal/cli/health.go` - Being deleted anyway
- `internal/api/server.go` - Can inline simple `/health` endpoint

**Action:** Replace with inline health checks:

```go
// Simple health check - no framework needed
func checkHealth() string {
    if err := repo.Ping(); err != nil {
        return "unhealthy"
    }
    return "healthy"
}
```

---

### 1.3 `internal/export/` - DELETE ENTIRE PACKAGE

| File | LOC | Status |
|------|-----|--------|
| `csv.go` | ~100 | Delete |
| `json.go` | ~50 | Delete |
| `csv_test.go` | ~100 | Delete |

**Reason:** Only used by `internal/cli/bulk/export.go`. The actual export logic is ~50 lines.

**Action:** Inline CSV/JSON export directly in `bulk/export.go`.

---

### 1.4 `internal/bulk/service.go` - DELETE FILE

| File | LOC | Status |
|------|-----|--------|
| `service.go` | ~186 | Delete |

**Reason:** Contains only utility functions (`FilterBySearch`, `AddUniqueStrings`, etc.) that are:
- Used only by `cli/bulk/` commands
- Simple slice operations that can be inlined

**Action:** Move needed utilities to `cli/bulk/helpers.go` or inline.

---

### 1.5 `cmd/verify-wal/` - DELETE ENTIRE COMMAND

| File | LOC | Status |
|------|-----|--------|
| `main.go` | ~50 | Delete |

**Reason:** Debug utility for SQLite WAL verification. Not needed in production.

**Action:** Delete. If needed for debugging, can be recreated.

---

## Phase 2: Delete Legacy CLI Commands

**Target:** ~1,600 LOC reduction

These commands are already hidden and deprecated. Delete after one release cycle.

### 2.1 Deprecated Analysis Commands

| File | LOC | Replacement | Status |
|------|-----|-------------|--------|
| `cli/score.go` | 161 | `tm add -n` (dry-run) | Delete |
| `cli/analyze.go` | 149 | `tm show` | Delete |
| `cli/analyze_llm.go` | 216 | `tm add --ai` | Delete |
| `cli/analyze_llm_test.go` | ~100 | - | Delete |

**Reason:** Functionality fully replaced by new commands.

---

### 2.2 Deprecated Review/Health Commands

| File | LOC | Replacement | Status |
|------|-----|-------------|--------|
| `cli/review.go` | 124 | `tm list` | Delete |
| `cli/health.go` | 198 | `tm status` | Delete |
| `cli/health/doctor.go` | ~150 | `tm status` | Delete |

**Reason:** Functionality fully replaced by new commands.

---

### 2.3 Deprecated Dump Commands

| File | LOC | Replacement | Status |
|------|-----|-------------|--------|
| `cli/dump/command.go` | 127 | `tm add` | Delete |
| `cli/dump/normal.go` | ~150 | `tm add` | Delete |
| `cli/dump/quick.go` | ~80 | `tm add -q` | Delete |
| `cli/dump/interactive.go` | ~200 | Consider keeping or delete | Delete |
| `cli/dump/llm.go` | ~100 | `tm add --ai` | Delete |
| `cli/dump/helpers.go` | ~50 | Inline in `add.go` | Delete |
| `cli/batch_dump.go` | 95 | `tm bulk import` | Delete |

**Reason:** The new `tm add` command with flags covers all use cases.

---

## Phase 3: Consolidate Duplicate Code

**Target:** ~400 LOC reduction

### 3.1 Merge Helper Files

**Current state:**
- `internal/cli/helpers.go` (79 LOC) - `displayIdeaAnalysis()`
- `internal/cli/llm_helpers.go` (12 LOC) - Single wrapper function
- `internal/cli/dump/helpers.go` (~50 LOC) - Duplicate helpers
- `internal/cli/analytics/helpers.go` (~50 LOC) - Duplicate helpers
- `internal/cli/bulk/helpers.go` (~50 LOC) - Duplicate helpers
- `internal/cliutil/helpers.go` (64 LOC) - Shared utilities

**Action:**
1. Delete `cli/llm_helpers.go` - inline the 1-line function
2. Move all shared display logic to `cliutil/`
3. Delete package-specific helper files, use `cliutil/` instead

---

### 3.2 Consolidate Color Definitions

**Current state:**
- `internal/cli/root.go` lines 57-61: Defines `successColor`, `errorColor`, etc.
- `internal/cliutil/helpers.go` lines 12-17: Defines `SuccessColor`, `ErrorColor`, etc.

**Action:** Delete duplicates from `root.go`, use `cliutil.*Color` everywhere.

---

### 3.3 Consolidate Display Functions

**Current state:** Multiple similar display functions:
- `displayIdeaAnalysis()` in `cli/helpers.go`
- `outputAddFull()` in `cli/add.go`
- `outputShowFull()` in `cli/show.go`
- `displayLLMAnalysisResult()` in `cli/analyze_llm.go`

**Action:** Create single `cliutil.DisplayIdea()` function with options.

---

### 3.4 Unify CLI Context Types

**Current state:** 4 different context types:
- `cli.CLIContext`
- `dump.CLIContext`
- `analytics.CLIContext`
- `bulk.CLIContext`

**Action:** Use single `cli.CLIContext` everywhere. Pass it directly instead of using getter functions.

---

## Phase 4: Simplify Architecture

### 4.1 Reduce Analytics Subcommands

**Current:** 7 subcommands under `tm analytics`:
- `trends`, `report`, `patterns`, `performance`, `anomaly`, `metrics`, `llm`

**Proposed:** Keep 3-4 most useful:
- `tm analytics` - Basic stats (default)
- `tm analytics trends` - Score trends
- `tm analytics report` - Full report

**Delete:**
- `analytics/anomaly.go` - Rarely used
- `analytics/performance.go` - Overlap with `metrics`
- `analytics/llm.go` - Can merge with main `llm` command

---

### 4.2 Inline Web Server Dependencies

After deleting `internal/tasks/` and `internal/health/`:

**Update `cmd/web/main.go`:**
- Inline scheduled tasks using `time.Ticker`
- Inline health check logic
- Remove imports for deleted packages

---

## Implementation Order

### Sprint 1: Safe Deletions (Low Risk)
- [ ] Delete `cmd/verify-wal/`
- [ ] Delete `internal/export/`, inline in `bulk/export.go`
- [ ] Delete `internal/bulk/service.go`, inline utilities
- [ ] Merge color definitions

### Sprint 2: Package Deletions (Medium Risk)
- [ ] Delete `internal/tasks/`, inline in web server
- [ ] Delete `internal/health/`, inline checks
- [ ] Update `cli/status.go` to not use health package

### Sprint 3: Legacy Command Cleanup
- [ ] Delete `cli/score.go`
- [ ] Delete `cli/analyze.go` and `cli/analyze_llm.go`
- [ ] Delete `cli/review.go`
- [ ] Delete `cli/health.go` and `cli/health/`
- [ ] Delete `cli/dump/` directory
- [ ] Delete `cli/batch_dump.go`
- [ ] Update `cli/root.go` to remove references

### Sprint 4: Consolidation
- [ ] Merge helper files into `cliutil/`
- [ ] Unify CLI context types
- [ ] Consolidate display functions
- [ ] Reduce analytics subcommands

---

## What We Keep

### Core Packages (Essential)
- `internal/database/` - Data persistence
- `internal/scoring/` - Both engines for dual-mode support
- `internal/profile/` - Universal mode configuration
- `internal/telos/` - Legacy mode configuration
- `internal/llm/` - AI provider abstraction
- `internal/models/` - Data structures
- `internal/patterns/` - Pattern detection
- `internal/config/` - Configuration management
- `internal/analytics/` - Analytics service (not CLI)
- `internal/logging/` - Structured logging
- `internal/metrics/` - Performance tracking

### CLI Commands (User-Facing)
- `tm add` - Add and score ideas
- `tm list` - Browse ideas
- `tm show` - View details
- `tm status` - System health
- `tm init` - Setup wizard
- `tm profile` - Profile management
- `tm prune` - Cleanup old ideas
- `tm link` - Relationship management
- `tm analytics` - Statistics
- `tm bulk` - Bulk operations
- `tm llm` - LLM management
- `tm completion` - Shell completion

### Web Server
- `cmd/web/` - Entry point
- `internal/api/` - REST API handlers

---

## Metrics

| Metric | Before | After | Change |
|--------|--------|-------|--------|
| Total Go LOC | ~22,700 | ~19,000 | -16% |
| Internal packages | 29 | 24 | -17% |
| CLI command files | 25+ | 15 | -40% |
| Test files | 51 | ~45 | -12% |
| Helper files | 6 | 1 | -83% |

---

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Breaking existing user workflows | Deprecation warnings for 1 release before deletion |
| Web server regression | Test web server after inlining tasks/health |
| Missing functionality | Comprehensive testing before each sprint |
| Git history complexity | One PR per sprint, clear commit messages |

---

## Approval

- [ ] Engineering review
- [ ] User impact assessment
- [ ] Test coverage verification
- [ ] Documentation update plan
