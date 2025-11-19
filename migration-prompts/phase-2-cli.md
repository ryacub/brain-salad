# Phase 2: CLI Implementation

**Duration:** 5-7 days
**Goal:** Build complete CLI using Cobra with feature parity to Rust version
**Dependencies:** Phase 1 complete

## Context

Implement all CLI commands using Cobra framework. Wire commands to core domain logic from Phase 1.

## Commands to Implement

1. `tm dump` - Capture idea with immediate analysis
2. `tm analyze` - Analyze existing or new idea  
3. `tm score` - Quick score without saving
4. `tm review` - Browse/filter ideas
5. `tm prune` - Clean up old ideas
6. `tm analytics` - View statistics
7. `tm link` - Manage idea relationships

## TDD Approach

For each command:
1. Write command tests (using exec or direct function calls)
2. Implement command handler
3. Wire to core logic
4. Add UX polish (colors, progress indicators)

## Example Test

```go
func TestDumpCommand_WithIdeaText_SavesAndDisplays(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	cmd := newDumpCommand(repo, scorer, detector)
	cmd.SetArgs([]string{"Build a SaaS product"})

	err := cmd.Execute()
	
	assert.NoError(t, err)
	
	// Verify saved
	ideas, _ := repo.ListIdeas(ctx, ListOptions{Limit: 1})
	assert.Len(t, ideas, 1)
	assert.Contains(t, ideas[0].Title, "SaaS")
}
```

## Deliverables

- [ ] `internal/cli/root.go` - Root command setup
- [ ] `internal/cli/dump.go` - Dump command  
- [ ] `internal/cli/analyze.go` - Analyze command
- [ ] `internal/cli/review.go` - Review command
- [ ] `internal/cli/score.go` - Score command
- [ ] `internal/cli/prune.go` - Prune command
- [ ] `internal/cli/analytics.go` - Analytics commands
- [ ] `internal/cli/link.go` - Link commands
- [ ] `cmd/cli/main.go` - Entry point
- [ ] Tests for all commands (>80% coverage)
- [ ] Colored output (github.com/fatih/color)
- [ ] Feature parity with Rust CLI

## Validation

```bash
go build -o bin/tm ./cmd/cli

# Test each command
./bin/tm dump "Build a Go CLI tool"
./bin/tm review --min-score 7.0
./bin/tm analyze --last
./bin/tm score "Test idea"
./bin/tm prune --dry-run

# Compare with Rust CLI - ensure same behavior
```

## Success Criteria

✅ All commands implemented and tested
✅ Feature parity with Rust version
✅ Help text complete
✅ Colored output working
✅ Tests passing (>80% coverage)
✅ User experience smooth

## Handoff

Ready for Phase 3 (API Server) or proceed to Phase 4 if API not needed immediately.
