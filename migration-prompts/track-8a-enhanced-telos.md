# Track 8A: Enhanced Telos Parsing

**Phase**: 8 - Polish & Documentation
**Estimated Time**: 4-5 hours
**Dependencies**: None
**Can Run in Parallel**: Yes (with 7, 8B)

---

## Mission

You are enhancing the Telos parser to support the full telos.md specification for the Telos Idea Matrix Go application, following Test-Driven Development (TDD).

## Context

- Current Go parser only handles: Goals, Strategies, Stack, Failure Patterns
- Rust parser supports full spec: Problems, Missions, Goals, Challenges, Strategies, Stack
- Need to parse all sections for complete telos support

## Reference Implementation

Review `/home/user/brain-salad/src/telos.rs` for full parsing logic

## Your Task

Enhance telos parser to support full specification using strict TDD methodology.

## Directory Structure

Enhance `go/internal/telos/`:
- `parser.go` - Add Problems, Missions, Challenges sections
- `parser_test.go` - Expand tests
- `testdata/full_telos.md` - Complete example file

## TDD Workflow (RED → GREEN → REFACTOR)

### STEP 1 - RED PHASE (Write Failing Tests)

Expand `go/internal/telos/parser_test.go`:
- `TestParser_ProblemsSection()`
- `TestParser_MissionsSection()`
- `TestParser_ChallengesSection()`
- `TestParser_FullTelosFile()`
- `TestParser_AllSectionsPresent()`

Run: `go test ./internal/telos -v`
Expected: **SOME TESTS FAIL** (new sections not parsed)

### STEP 2 - GREEN PHASE (Implement)

#### A. Update `go/internal/telos/parser.go`:

```go
package telos

// Add new fields to Telos struct
type Telos struct {
    Problems   []Problem
    Missions   []Mission
    Goals      []Goal
    Challenges []Challenge
    Strategies []Strategy
    Stack      Stack
    Patterns   []FailurePattern
}

type Problem struct {
    ID          string // P1, P2, etc.
    Description string
}

type Mission struct {
    ID          string // M1, M2, etc.
    Description string
}

type Challenge struct {
    ID          string // C1, C2, etc.
    Description string
}

// Enhance ParseFile function
func ParseFile(path string) (*Telos, error) {
    content, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read file: %w", err)
    }
    
    telos := &Telos{}
    
    // Parse all sections
    telos.Problems = parseProblems(string(content))
    telos.Missions = parseMissions(string(content))
    telos.Goals = parseGoals(string(content))
    telos.Challenges = parseChallenges(string(content))
    telos.Strategies = parseStrategies(string(content))
    telos.Stack = parseStack(string(content))
    telos.Patterns = parseFailurePatterns(string(content))
    
    return telos, nil
}

func parseProblems(content string) []Problem {
    section := extractSection(content, "## Problems")
    if section == "" {
        return []Problem{}
    }
    
    // Match lines like: - P1: Description here
    re := regexp.MustCompile(`(?m)^\s*-\s*(P\d+):\s*(.+)$`)
    matches := re.FindAllStringSubmatch(section, -1)
    
    problems := make([]Problem, 0, len(matches))
    for _, match := range matches {
        if len(match) >= 3 {
            problems = append(problems, Problem{
                ID:          match[1],
                Description: strings.TrimSpace(match[2]),
            })
        }
    }
    
    return problems
}

func parseMissions(content string) []Mission {
    section := extractSection(content, "## Missions")
    if section == "" {
        return []Mission{}
    }
    
    re := regexp.MustCompile(`(?m)^\s*-\s*(M\d+):\s*(.+)$`)
    matches := re.FindAllStringSubmatch(section, -1)
    
    missions := make([]Mission, 0, len(matches))
    for _, match := range matches {
        if len(match) >= 3 {
            missions = append(missions, Mission{
                ID:          match[1],
                Description: strings.TrimSpace(match[2]),
            })
        }
    }
    
    return missions
}

func parseChallenges(content string) []Challenge {
    section := extractSection(content, "## Challenges")
    if section == "" {
        return []Challenge{}
    }
    
    re := regexp.MustCompile(`(?m)^\s*-\s*(C\d+):\s*(.+)$`)
    matches := re.FindAllStringSubmatch(section, -1)
    
    challenges := make([]Challenge, 0, len(matches))
    for _, match := range matches {
        if len(match) >= 3 {
            challenges = append(challenges, Challenge{
                ID:          match[1],
                Description: strings.TrimSpace(match[2]),
            })
        }
    }
    
    return challenges
}

// Helper to extract section between headers
func extractSection(content, header string) string {
    // Find section start
    headerIndex := strings.Index(content, header)
    if headerIndex == -1 {
        return ""
    }
    
    // Find next header (## Something)
    nextHeaderRe := regexp.MustCompile(`(?m)^##\s+`)
    matches := nextHeaderRe.FindAllStringIndex(content[headerIndex+len(header):], -1)
    
    if len(matches) > 0 {
        endIndex := headerIndex + len(header) + matches[0][0]
        return content[headerIndex+len(header) : endIndex]
    }
    
    // No next header, take rest of file
    return content[headerIndex+len(header):]
}
```

#### B. Create `go/internal/telos/testdata/full_telos.md`:

```markdown
# My Telos

## Problems
- P1: Too many ideas, decision paralysis
- P2: Context switching between projects
- P3: Perfectionism delays shipping

## Missions
- M1: Ship profitable SaaS products
- M2: Build personal brand through open source
- M3: Achieve financial independence

## Goals
- G1: Launch profitable SaaS (Deadline: 2025-12-31)
- G2: $10K MRR from side projects (Deadline: 2026-06-30)
- G3: 1000 GitHub stars on open source (Deadline: 2025-12-31)

## Challenges
- C1: Limited time (full-time job)
- C2: Small audience/network
- C3: Technical depth vs breadth tradeoff

## Strategies
- S1: Ship early and often, iterate based on feedback
- S2: Focus on one technology stack
- S3: Build in public for accountability

## Stack
- Primary: Go, Python, TypeScript
- Secondary: Docker, PostgreSQL, Redis

## Failure Patterns
- Context switching: Starting new projects before finishing current ones
- Perfectionism: Over-engineering solutions before validating
- Tutorial hell: Consuming content instead of building
```

Run: `go test ./internal/telos -v`
Expected: **ALL TESTS PASS**

### STEP 3 - REFACTOR PHASE

- Extract common parsing patterns
- Optimize regex compilation (compile once)
- Add validation for cross-references (M1 in goals, etc.)
- Better error messages for malformed sections

## Success Criteria

- ✅ All tests pass with >95% coverage
- ✅ Parses all 7 sections correctly
- ✅ Handles missing sections gracefully
- ✅ Matches Rust `src/telos.rs` parsing

## Validation

```bash
# Unit tests
go test ./internal/telos -v -cover

# Test with full telos file
go run ./cmd/cli/main.go dump "Test idea" --telos testdata/full_telos.md
```

## Deliverables

- Enhanced `go/internal/telos/parser.go`
- Expanded `go/internal/telos/parser_test.go`
- `go/internal/telos/testdata/full_telos.md`
