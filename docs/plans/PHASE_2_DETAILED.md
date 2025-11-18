# Phase 2: Testing & Quality Assurance - Detailed Breakdown

> **For Subagent Execution**: Each task is 1-2 hours. Build on Phase 1 completion.

**Goal**: Ensure code quality, test critical paths, and establish CI/CD pipeline.

**Effort**: 3-4 hours total (3 main tasks × 60-75 min each)

**Blocker**: Phase 1 must be complete first (config module)

**Why Phase 2 After Phase 1?** Configuration changes are the foundation; tests verify they work reliably across scenarios.

---

## Task 2.1: Write Scoring Strategy Unit Tests

**Subagent: Create tests for scoring logic in `tests/scoring_strategy_test.rs`**

### What We're Testing

The `TelosScoringStrategy` implementation validates:
1. Score returns value 0-10
2. Pattern detection identifies known patterns
3. Goal alignment calculation works
4. Different idea types score correctly
5. Edge cases (empty content, null values) handled

### Why This Task?

Scoring is core to the system. Users depend on it. Tests prevent regressions.

### Requirements

**Input**: Existing `src/scoring.rs` and related modules

**Output**:
- Comprehensive test suite
- Tests cover happy path + edge cases
- All tests pass in CI environment

**Exit Criteria**:
- [ ] `tests/scoring_strategy_test.rs` created (~300 lines)
- [ ] `tests/fixtures/` expanded with scoring test data
- [ ] All tests pass: `cargo test --test scoring_strategy_test`
- [ ] No flaky tests (use same seed for randomness)
- [ ] Tests run in < 5 seconds

### Test Cases (Minimum 12)

#### Test 1: Score returns valid range
```rust
#[tokio::test]
async fn test_score_returns_valid_range() {
    let telos = load_test_telos().await.unwrap();
    let strategy = TelosScoringStrategy::new(telos);

    let idea = Idea {
        content: "Build a Rust project aligned with shipping goals".to_string(),
        ..Default::default()
    };

    let score = strategy.score(&idea).await;
    assert!(score >= 0.0 && score <= 10.0, "Score {} out of range", score);
}
```

#### Test 2: High-alignment ideas score high
```rust
#[tokio::test]
async fn test_high_alignment_idea_scores_high() {
    let telos = load_test_telos().await.unwrap();
    let strategy = TelosScoringStrategy::new(telos);

    let idea = Idea {
        content: "Ship MVP of AI product using Rust".to_string(),
        ..Default::default()
    };

    let score = strategy.score(&idea).await;
    assert!(score > 6.0, "High-alignment idea should score > 6.0, got {}", score);
}
```

#### Test 3: Low-alignment ideas score low
```rust
#[tokio::test]
async fn test_low_alignment_idea_scores_low() {
    let telos = load_test_telos().await.unwrap();
    let strategy = TelosScoringStrategy::new(telos);

    let idea = Idea {
        content: "Learn PHP for freelance work".to_string(),
        ..Default::default()
    };

    let score = strategy.score(&idea).await;
    assert!(score < 4.0, "Low-alignment idea should score < 4.0, got {}", score);
}
```

#### Test 4: Context-switching pattern detected
```rust
#[tokio::test]
async fn test_context_switching_pattern_detected() {
    let telos = load_test_telos().await.unwrap();
    let strategy = TelosScoringStrategy::new(telos);

    let idea = Idea {
        content: "Switch to learning JavaScript and Node.js".to_string(),
        ..Default::default()
    };

    let patterns = strategy.detect_patterns(&idea).await;
    assert!(
        patterns.iter().any(|p| p.name.contains("context") || p.name.contains("switch")),
        "Should detect context-switching pattern"
    );
}
```

#### Test 5: Perfectionism pattern detected
```rust
#[tokio::test]
async fn test_perfectionism_pattern_detected() {
    let telos = load_test_telos().await.unwrap();
    let strategy = TelosScoringStrategy::new(telos);

    let idea = Idea {
        content: "Redesign entire UI with perfect animations and interactions".to_string(),
        ..Default::default()
    };

    let patterns = strategy.detect_patterns(&idea).await;
    assert!(
        patterns.iter().any(|p| p.name.contains("perfect")),
        "Should detect perfectionism pattern"
    );
}
```

#### Test 6: Procrastination pattern detected
```rust
#[tokio::test]
async fn test_procrastination_pattern_detected() {
    let telos = load_test_telos().await.unwrap();
    let strategy = TelosScoringStrategy::new(telos);

    let idea = Idea {
        content: "Spend 2 weeks researching the perfect tech stack".to_string(),
        ..Default::default()
    };

    let patterns = strategy.detect_patterns(&idea).await;
    assert!(
        patterns.iter().any(|p| p.name.contains("procrastination") || p.name.contains("research")),
        "Should detect procrastination pattern"
    );
}
```

#### Test 7: Multiple patterns in one idea
```rust
#[tokio::test]
async fn test_multiple_patterns_in_one_idea() {
    let telos = load_test_telos().await.unwrap();
    let strategy = TelosScoringStrategy::new(telos);

    let idea = Idea {
        content: "Switch to Ruby, build perfect Rails app, spend weeks researching".to_string(),
        ..Default::default()
    };

    let patterns = strategy.detect_patterns(&idea).await;
    assert!(patterns.len() > 1, "Should detect multiple patterns");
}
```

#### Test 8: Empty idea handled gracefully
```rust
#[tokio::test]
async fn test_empty_idea_content() {
    let telos = load_test_telos().await.unwrap();
    let strategy = TelosScoringStrategy::new(telos);

    let idea = Idea {
        content: String::new(),
        ..Default::default()
    };

    // Should not panic
    let score = strategy.score(&idea).await;
    assert!(score >= 0.0 && score <= 10.0);

    let patterns = strategy.detect_patterns(&idea).await;
    // Should return empty patterns, not crash
    assert!(patterns.len() >= 0);
}
```

#### Test 9: Very long idea content
```rust
#[tokio::test]
async fn test_very_long_idea_content() {
    let telos = load_test_telos().await.unwrap();
    let strategy = TelosScoringStrategy::new(telos);

    let long_content = "Build a product ".repeat(1000); // 16KB string
    let idea = Idea {
        content: long_content,
        ..Default::default()
    };

    // Should handle without performance issues
    let score = strategy.score(&idea).await;
    assert!(score >= 0.0 && score <= 10.0);
}
```

#### Test 10: Detailed score breakdown
```rust
#[tokio::test]
async fn test_score_breakdown_has_components() {
    let telos = load_test_telos().await.unwrap();
    let strategy = TelosScoringStrategy::new(telos);

    let idea = Idea {
        content: "Ship product using Rust".to_string(),
        ..Default::default()
    };

    let breakdown = strategy.score_detailed(&idea).await;

    // Should have all components
    assert!(breakdown.overall >= 0.0 && breakdown.overall <= 10.0);
    assert!(breakdown.mission_alignment >= 0.0);
    assert!(breakdown.strategic_fit >= 0.0);
    assert!(breakdown.pattern_risks >= 0.0);
    assert!(!breakdown.reasoning.is_empty());
}
```

#### Test 11: Stack compliance detection
```rust
#[tokio::test]
async fn test_stack_compliance_affects_score() {
    let telos = load_test_telos().await.unwrap();
    let strategy = TelosScoringStrategy::new(telos);

    // Within stack
    let rust_idea = Idea {
        content: "Build with Rust and Tokio".to_string(),
        ..Default::default()
    };

    // Outside stack
    let php_idea = Idea {
        content: "Build with PHP and Laravel".to_string(),
        ..Default::default()
    };

    let rust_score = strategy.score(&rust_idea).await;
    let php_score = strategy.score(&php_idea).await;

    // Rust should score higher (within stack)
    assert!(rust_score > php_score, "In-stack should score higher");
}
```

#### Test 12: Consistent scores for same input
```rust
#[tokio::test]
async fn test_consistent_scoring() {
    let telos = load_test_telos().await.unwrap();
    let strategy = TelosScoringStrategy::new(telos);

    let idea = Idea {
        content: "Build AI product with Rust".to_string(),
        ..Default::default()
    };

    let score1 = strategy.score(&idea).await;
    let score2 = strategy.score(&idea).await;

    assert_eq!(score1, score2, "Same input should produce same score");
}
```

### Test Fixtures

**Create `tests/fixtures/telos_for_scoring.yaml`**:
```yaml
goals:
  - name: "Ship product"
    deadline: "2025-12-31"
    priority: 1
  - name: "Build community"
    deadline: "2025-12-31"
    priority: 2
  - name: "Establish credibility"
    deadline: "2025-12-31"
    priority: 3
  - name: "Create income"
    deadline: "2025-12-31"
    priority: 4

strategies:
  - name: "Focus on shipping"
    weight: 1.2
  - name: "One stack rule"
    weight: 1.0
  - name: "Build in public"
    weight: 0.8
  - name: "MVP mindset"
    weight: 1.0

stack:
  primary: "Rust"
  secondary: "Python"

failure_patterns:
  - "context-switching"
  - "perfectionism"
  - "procrastination"
  - "analysis-paralysis"
```

### Helper Functions for Tests

```rust
async fn load_test_telos() -> Result<TelosConfig> {
    // Load from fixture
    telos::load(&PathBuf::from("tests/fixtures/telos_for_scoring.yaml")).await
}

fn create_idea(content: &str) -> Idea {
    Idea {
        content: content.to_string(),
        ..Default::default()
    }
}
```

### Deliverables

1. `tests/scoring_strategy_test.rs` - ~300 lines with 12+ tests
2. `tests/fixtures/telos_for_scoring.yaml` - Test data
3. All tests pass: `cargo test --test scoring_strategy_test`
4. No test timeouts (< 5 sec total)

### Subagent Notes

- Use `#[tokio::test]` for async tests
- Load test telos once, reuse across tests
- Use descriptive test names: `test_X_produces_Y`
- Test both happy path AND edge cases
- Don't test external dependencies (AI integration) here

---

## Task 2.2: Create GitHub Actions Test Workflow

**Subagent: Set up `.github/workflows/test.yml`**

### What We're Building

Automated testing that runs on:
- Every push to `main` or `develop`
- Every pull request
- Manual trigger (optional)

Tests run on Linux (Ubuntu) and optionally macOS/Windows.

### Requirements

**Output**:
- `.github/workflows/test.yml` created
- Workflow runs on push/PR
- All tests execute
- Results visible in GitHub UI

**Exit Criteria**:
- [ ] Workflow file created
- [ ] Syntax is valid (GitHub validates)
- [ ] Runs on `push` and `pull_request`
- [ ] Covers: cargo test, clippy, fmt, build
- [ ] Results show pass/fail clearly

### Workflow Definition

Create `.github/workflows/test.yml`:

```yaml
name: Tests

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    name: Test Suite
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Install Rust toolchain
        uses: dtolnay/rust-toolchain@stable

      - name: Cache Rust build artifacts
        uses: Swatinem/rust-cache@v2

      - name: Run tests
        run: cargo test --all-features --verbose

      - name: Run doc tests
        run: cargo test --doc

  clippy:
    name: Clippy Linting
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Install Rust toolchain
        uses: dtolnay/rust-toolchain@stable
        with:
          components: clippy

      - name: Cache Rust build artifacts
        uses: Swatinem/rust-cache@v2

      - name: Run clippy
        run: cargo clippy --all-targets --all-features -- -D warnings

  fmt:
    name: Code Formatting
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Install Rust toolchain
        uses: dtolnay/rust-toolchain@stable
        with:
          components: rustfmt

      - name: Check formatting
        run: cargo fmt -- --check

  build:
    name: Build Release
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Install Rust toolchain
        uses: dtolnay/rust-toolchain@stable

      - name: Cache Rust build artifacts
        uses: Swatinem/rust-cache@v2

      - name: Build release binary
        run: cargo build --release --verbose
```

### Workflow Features

**Parallel Jobs**: 4 independent jobs run simultaneously
- `test`: Runs all tests
- `clippy`: Code quality checks
- `fmt`: Code formatting
- `build`: Release build verification

**Caching**: Uses Swatinem/rust-cache to speed up builds

**Branch Filters**: Only runs on main/develop branches (not feature branches by default)

**Status Checks**: GitHub shows pass/fail badge on PRs

### Testing the Workflow

After committing:
1. Push to GitHub
2. Go to Actions tab
3. See workflow run
4. Click to see logs
5. Green checkmark = success

### Deliverables

1. `.github/workflows/test.yml` created
2. Valid YAML syntax
3. All 4 jobs defined
4. Caching configured
5. Clear status messages

### Subagent Notes

- Use standard actions (checkout, rust-toolchain, cache)
- `-D warnings` fails on any clippy warning (good for CI)
- All features: `--all-features` ensures full test coverage
- Run on ubuntu-latest for consistency
- Can add macOS/Windows later

---

## Task 2.3: Add clippy and fmt to Development Workflow

**Subagent: Create helper script for local quality checks**

### What We're Building

A script developers can run locally before committing to catch issues early.

### Requirements

**Output**:
- Script: `scripts/check-quality.sh`
- Or: Makefile target (optional)
- Runs all checks in order
- Reports pass/fail

### Implementation Option A: Shell Script

Create `scripts/check-quality.sh`:

```bash
#!/bin/bash

set -e

echo "🔍 Running quality checks..."
echo

echo "1️⃣  Checking formatting..."
cargo fmt --all -- --check
echo "✅ Formatting check passed"
echo

echo "2️⃣  Running clippy..."
cargo clippy --all-targets --all-features -- -D warnings
echo "✅ Clippy check passed"
echo

echo "3️⃣  Running tests..."
cargo test --all-features
echo "✅ Tests passed"
echo

echo "4️⃣  Building release..."
cargo build --release
echo "✅ Release build passed"
echo

echo "✨ All checks passed! Ready to commit."
```

Make it executable:
```bash
chmod +x scripts/check-quality.sh
```

### Implementation Option B: Makefile Target

Add to Makefile (if exists) or create one:

```makefile
.PHONY: check quality

check: fmt clippy test build
	@echo "✨ All checks passed!"

fmt:
	cargo fmt --all

clippy:
	cargo clippy --all-targets --all-features -- -D warnings

test:
	cargo test --all-features

build:
	cargo build --release
```

Then: `make check`

### Pre-commit Hook Setup (Optional)

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash

# Run quality checks before allowing commit
./scripts/check-quality.sh

if [ $? -ne 0 ]; then
    echo "❌ Quality checks failed. Commit aborted."
    exit 1
fi
```

### Deliverables

1. `scripts/check-quality.sh` created and executable
2. Or: Makefile with quality targets
3. Clear output showing which checks pass/fail
4. Takes < 60 seconds to run

### Subagent Notes

- Make script fail fast (stop on first error)
- Show which check is running
- Show success/failure clearly
- Can be run anytime: `./scripts/check-quality.sh`

---

## Task 2.4: Verify All Tests Pass Locally

**Subagent: Run comprehensive test suite and report results**

### What We're Checking

1. Existing tests still pass (no regressions)
2. New config tests pass
3. New scoring tests pass
4. Build succeeds
5. No clippy warnings
6. Code is formatted

### Steps

```bash
# 1. Clean build
cargo clean
cargo build

# 2. Run all tests
cargo test --all-features --verbose

# 3. Run specific test suites
cargo test config::
cargo test --test config_integration_test
cargo test --test scoring_strategy_test

# 4. Check quality
cargo clippy --all-targets --all-features
cargo fmt --check

# 5. Build release
cargo build --release
```

### Expected Results

All should show success. If any fail:
- Subagent identifies issue
- Fixes code or reports blockers
- Retests until all pass

### Deliverables

1. All tests pass locally
2. No clippy warnings
3. Code is formatted
4. Release build succeeds
5. Report showing results

---

## Phase 2 Completion Checklist

**For Subagent to Verify:**

### Testing
- [ ] `tests/scoring_strategy_test.rs` created (12+ tests)
- [ ] `tests/fixtures/telos_for_scoring.yaml` created
- [ ] All scoring tests pass
- [ ] `tests/config_integration_test.rs` passes (from Phase 1)
- [ ] No flaky tests
- [ ] Tests run in < 10 seconds total

### CI/CD
- [ ] `.github/workflows/test.yml` created
- [ ] Workflow syntax is valid
- [ ] All 4 jobs defined (test, clippy, fmt, build)
- [ ] Caching configured
- [ ] Branch filters set correctly

### Quality
- [ ] `cargo build` succeeds
- [ ] `cargo test --all-features` passes
- [ ] `cargo clippy` produces no new warnings
- [ ] `cargo fmt --check` passes
- [ ] `cargo build --release` succeeds

### Git
- [ ] `git status` shows changes ready to commit
- [ ] Ready for 3-4 commits

---

## Commits for Phase 2

### Commit 1: Scoring tests
```bash
git add tests/scoring_strategy_test.rs tests/fixtures/telos_for_scoring.yaml
git commit -m "test: add comprehensive scoring strategy tests

- Test score range validation
- Test pattern detection
- Test multiple patterns
- Test edge cases (empty, long content)
- Test consistent scoring
- Test stack compliance"
```

### Commit 2: GitHub Actions workflow
```bash
git add .github/workflows/test.yml
git commit -m "ci: add GitHub Actions test workflow

- Run tests on push and PR
- Parallel jobs: test, clippy, fmt, build
- Cache Rust artifacts
- Clear status reporting"
```

### Commit 3: Quality check scripts
```bash
git add scripts/check-quality.sh
git commit -m "chore: add local quality check script

- Run all quality checks locally
- Fail fast on errors
- Clear progress reporting"
```

---

## Expected Output for Subagent

```
PHASE 2 COMPLETION REPORT
========================

✅ Task 2.1: Scoring Tests
- tests/scoring_strategy_test.rs created (340 lines)
- 12 test cases covering all scenarios
- All tests pass
- Test execution time: 3.2 seconds

✅ Task 2.2: GitHub Actions Workflow
- .github/workflows/test.yml created
- 4 parallel jobs: test, clippy, fmt, build
- Caching configured
- Valid YAML syntax

✅ Task 2.3: Quality Check Script
- scripts/check-quality.sh created
- Executable permissions set
- Runs all 4 quality checks
- Total execution time: 45 seconds

✅ Task 2.4: Local Verification
- cargo build ✓
- cargo test ✓ (all 50+ tests pass)
- cargo clippy ✓ (no warnings)
- cargo fmt ✓ (code formatted)
- cargo build --release ✓

Summary: Testing infrastructure complete.
GitHub Actions CI/CD ready.
Code quality verified.
```

---

**Phase 2 is ready for subagent execution. Build on Phase 1 completion.**
