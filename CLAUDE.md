# Brain-Salad

Universal idea scoring CLI that helps anyone evaluate ideas based on their personal priorities.

## Setup

```bash
# Required: golangci-lint v2 (CI uses v2.6.2)
brew install golangci-lint
# Or upgrade: brew upgrade golangci-lint

# Verify version (must be 2.x)
golangci-lint --version
```

## Quick Start

```bash
make check   # Run before EVERY push - catches what CI catches
make test    # Run tests only
make lint    # Run linters only
make build   # Build binaries
```

## Pre-Push Checklist

**ALWAYS run `make check` before pushing.** This runs:
1. `gofmt` - Code formatting
2. `golangci-lint` - All linters (errcheck, staticcheck, ineffassign, etc.)
3. `go test -race` - Tests with race detection
4. `go build` - Compilation check

## Common Linter Issues

### errcheck: Unchecked return values

**Problem:** Color print functions from `fatih/color` return `(int, error)`.

**Fix:** Explicitly ignore with blank identifier:
```go
// Wrong - triggers errcheck
headerColor.Println("Hello")

// Correct - explicit ignore
_, _ = headerColor.Println("Hello")
```

### ineffassign: Ineffectual assignment

**Problem:** Assigning a value that's immediately overwritten.

**Fix:** Use `var` declaration instead:
```go
// Wrong - value never used
baseScore := 0.5
if condition {
    baseScore = 0.8
} else {
    baseScore = 0.3
}

// Correct
var baseScore float64
if condition {
    baseScore = 0.8
} else {
    baseScore = 0.3
}
```

### staticcheck QF1003: Use tagged switch

**Problem:** If-else chain on same variable should be switch.

**Fix:**
```go
// Wrong
if x == "a" {
    // ...
} else if x == "b" {
    // ...
}

// Correct
switch x {
case "a":
    // ...
case "b":
    // ...
}
```

## Architecture

```
internal/
├── profile/           # User preference system
│   ├── profile.go     # Profile struct and types
│   ├── loader.go      # YAML load/save
│   └── keywords.go    # Goal/avoid keyword extraction
├── scoring/
│   ├── universal_engine.go  # New: Profile-based scoring
│   ├── dimensions.go        # Universal dimension definitions
│   └── engine.go            # Legacy: telos.md-based scoring
├── cli/
│   ├── wizard/        # Interactive setup wizard
│   ├── root.go        # CLI initialization, mode detection
│   ├── score.go       # Score command
│   └── profile_cmd.go # Profile management
└── cliutil/           # CLI display helpers
```

## Dual Scoring Modes

1. **Universal Mode** (default): Uses `~/.brain-salad/profile.yaml`
   - Created via `brain-salad init` wizard
   - 6 dimensions: completion, skill fit, timeline, reward, sustainability, avoidance

2. **Legacy Mode**: Uses `~/.telos/telos.md`
   - For power users with existing telos configurations
   - 3 categories: Mission, AntiChallenge, Strategic

## Testing

```bash
make test              # All tests
make test-coverage     # With coverage report
go test ./internal/profile/...   # Specific package
```

Test files follow `*_test.go` convention alongside source files.

## File Locations

- `~/.brain-salad/profile.yaml` - User preferences (universal mode)
- `~/.brain-salad/ideas.db` - Idea database
- `~/.telos/telos.md` - Legacy configuration (if present)
