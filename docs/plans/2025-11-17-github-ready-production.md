# GitHub-Ready Production Deployment Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Transform telos-idea-matrix from a personal tool into a production-grade, generalized idea management framework that users can personalize with their own Telos/goal documents, with proper GitHub presence, Docker support, and comprehensive documentation.

**Architecture:**
- Decouple Telos-specific logic into a pluggable configuration layer
- Support multiple Telos/goal file formats and locations
- Create Docker containerization for cross-platform consistency
- Establish CI/CD pipeline with GitHub Actions
- Comprehensive testing, documentation, and release infrastructure

**Tech Stack:** Rust (Tokio async), SQLite, Ollama/LLM integration, Docker, GitHub Actions, Cargo

---

## Phase 1: Code Architecture Refactoring (Decoupling Personal Dependencies)

### Task 1: Abstract Telos Configuration System

**Objective:** Make Telos file loading generic and user-configurable instead of hardcoded to `/Users/rayyacub/Documents/CCResearch/Hanai/telos.md`

**Files:**
- Modify: `src/telos.rs` (lines 1-150)
- Create: `src/config.rs` (new file)
- Modify: `src/main.rs` (lines 1-50)
- Create: `docs/CONFIGURATION.md` (new file)

**Step 1: Create configuration loader module**

Create `src/config.rs`:

```rust
use std::path::{Path, PathBuf};
use anyhow::{Result, Context};
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ConfigPaths {
    /// Path to user's Telos/goal document
    pub telos_file: PathBuf,
    /// Path to database directory
    pub data_dir: PathBuf,
    /// Path to logs directory
    pub log_dir: PathBuf,
    /// Optional custom configuration file
    pub config_file: Option<PathBuf>,
}

impl ConfigPaths {
    /// Load from environment, config file, or use defaults
    pub fn load() -> Result<Self> {
        // 1. Check environment variable for telos file
        if let Ok(telos_path) = std::env::var("TELOS_FILE") {
            return Ok(ConfigPaths {
                telos_file: PathBuf::from(telos_path),
                data_dir: Self::default_data_dir(),
                log_dir: Self::default_log_dir(),
                config_file: None,
            });
        }

        // 2. Check for config file in XDG_CONFIG_HOME or ~/.config
        if let Some(config_paths) = Self::load_from_config_file() {
            return Ok(config_paths);
        }

        // 3. Check for telos.md in current directory
        if Path::new("telos.md").exists() {
            return Ok(ConfigPaths {
                telos_file: PathBuf::from("telos.md"),
                data_dir: Self::default_data_dir(),
                log_dir: Self::default_log_dir(),
                config_file: None,
            });
        }

        // 4. Return error with helpful instructions
        Err(anyhow::anyhow!(
            "No Telos configuration found. Please:\n\
             1. Set TELOS_FILE environment variable, OR\n\
             2. Place telos.md in current directory, OR\n\
             3. Create ~/.config/telos-matrix/config.toml with telos_file path\n\
             See docs/CONFIGURATION.md for details"
        ))
    }

    fn default_data_dir() -> PathBuf {
        dirs::data_dir()
            .unwrap_or_else(|| PathBuf::from("."))
            .join("telos-matrix")
    }

    fn default_log_dir() -> PathBuf {
        dirs::cache_dir()
            .unwrap_or_else(|| PathBuf::from("."))
            .join("telos-matrix/logs")
    }

    fn load_from_config_file() -> Option<ConfigPaths> {
        let config_dir = dirs::config_dir()?;
        let config_path = config_dir.join("telos-matrix/config.toml");

        if config_path.exists() {
            let content = std::fs::read_to_string(&config_path).ok()?;
            let config: ConfigPaths = toml::from_str(&content).ok()?;
            return Some(config);
        }
        None
    }

    pub fn ensure_directories_exist(&self) -> Result<()> {
        std::fs::create_dir_all(&self.data_dir)
            .context("Failed to create data directory")?;
        std::fs::create_dir_all(&self.log_dir)
            .context("Failed to create log directory")?;
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_default_paths_creation() {
        let paths = ConfigPaths {
            telos_file: PathBuf::from("telos.md"),
            data_dir: PathBuf::from(".telos-matrix"),
            log_dir: PathBuf::from(".telos-matrix/logs"),
            config_file: None,
        };
        assert_eq!(paths.telos_file, PathBuf::from("telos.md"));
    }
}
```

**Step 2: Update telos.rs to use configurable paths**

In `src/telos.rs`, modify the `load()` function:

```rust
pub async fn load(telos_path: &Path) -> Result<TelosConfig> {
    let content = tokio::fs::read_to_string(telos_path)
        .await
        .context(format!("Failed to read Telos file from: {:?}", telos_path))?;

    // Parse the YAML/Markdown telos structure
    parse_telos_markdown(&content)
}
```

**Step 3: Update main.rs to load configuration**

Modify `src/main.rs` top section:

```rust
mod config;
mod telos;
// ... other imports

use config::ConfigPaths;

#[tokio::main]
async fn main() -> Result<()> {
    // Load configuration
    let config_paths = ConfigPaths::load()?;
    config_paths.ensure_directories_exist()?;

    // Load Telos configuration
    let telos_config = telos::load(&config_paths.telos_file).await?;

    // Initialize database with data_dir from config
    let db = Database::new(&config_paths.data_dir).await?;

    // ... rest of main
    Ok(())
}
```

**Step 4: Create configuration documentation**

Create `docs/CONFIGURATION.md`:

```markdown
# Configuration Guide

## Overview

The Telos Idea Matrix is personalized per user via a Telos configuration file. This file describes your:
- Goals and deadlines
- Strategies and focus areas
- Technology stack preferences
- Known failure patterns

## Setting Up Your Telos File

### Option 1: Environment Variable (Recommended for Docker/CI)

```bash
export TELOS_FILE=/path/to/your/telos.md
tm dump "Your idea"
```

### Option 2: Current Directory

Place your `telos.md` in your working directory:

```bash
cd /my/project
cp /path/to/my/telos.md .
tm dump "Your idea"
```

### Option 3: Configuration File

Create `~/.config/telos-matrix/config.toml`:

```toml
telos_file = "/path/to/your/telos.md"
data_dir = "~/.local/share/telos-matrix"
log_dir = "~/.cache/telos-matrix/logs"
```

## Telos File Format

Your `telos.md` should contain sections like:

```markdown
# My Telos

## Goals
- G1: [Goal 1] (Deadline: YYYY-MM-DD)
- G2: [Goal 2] (Deadline: YYYY-MM-DD)
- G3: [Goal 3] (Deadline: YYYY-MM-DD)
- G4: [Goal 4] (Deadline: YYYY-MM-DD)

## Strategies
- S1: [Strategy 1]
- S2: [Strategy 2]
- S3: [Strategy 3]
- S4: [Strategy 4]

## Stack
- Primary: [Your main tech stack]
- Secondary: [Secondary technologies]

## Failure Patterns
- Pattern 1: [Description]
- Pattern 2: [Description]
```

## Migration from Personal Setup

If you're Ray's colleagues, copy Ray's telos.md and customize it:

```bash
cp /Users/rayyacub/Documents/CCResearch/Hanai/telos.md ./my-telos.md
# Edit my-telos.md to match your goals and patterns
export TELOS_FILE=$(pwd)/my-telos.md
tm dump "My first idea"
```
```

**Step 5: Run tests to verify configuration module**

```bash
cargo test config::tests
```

Expected output: All tests pass

**Step 6: Commit configuration changes**

```bash
git add src/config.rs src/telos.rs src/main.rs docs/CONFIGURATION.md
git commit -m "feat: abstract Telos configuration into pluggable system"
```

---

### Task 2: Remove Personal Hardcoded Paths

**Objective:** Audit codebase for hardcoded paths and replace with config-based lookups

**Files:**
- Search: `src/**/*.rs` for hardcoded paths
- Modify: Files containing `/Users/rayyacub`
- Create: `docs/HARDCODED_PATHS.md` (audit log)

**Step 1: Audit for hardcoded paths**

```bash
grep -r "/Users/rayyacub" src/ --include="*.rs"
grep -r "Hanai" src/ --include="*.rs"
grep -r "telos.md" src/ --include="*.rs"
```

**Step 2: Document findings**

Create `docs/HARDCODED_PATHS.md`:

```markdown
# Hardcoded Path Audit

## Found and Fixed
- [x] Main telos.md loading in telos.rs → Migrated to config module
- [ ] Any other personal paths

## Instructions for Complete Removal
1. Remove all references to personal paths
2. Use ConfigPaths for all file access
3. Update CLI to accept path arguments
```

**Step 3: Fix remaining hardcoded paths**

For each file found, replace with config-based path. Example:

```rust
// Before:
let telos_path = "/Users/rayyacub/Documents/CCResearch/Hanai/telos.md";

// After:
let telos_path = &config_paths.telos_file;
```

**Step 4: Commit**

```bash
git add src/
git commit -m "fix: remove hardcoded personal paths from codebase"
```

---

### Task 3: Create Abstract Scoring Interface

**Objective:** Make scoring system pluggable so it works with any goal framework (not just Telos)

**Files:**
- Create: `src/scoring/interface.rs` (new trait)
- Modify: `src/scoring.rs` (current implementation)
- Create: `src/scoring/telos_impl.rs` (Telos-specific scoring)

**Step 1: Define scoring trait**

Create `src/scoring/interface.rs`:

```rust
use async_trait::async_trait;
use crate::types::Idea;

#[async_trait]
pub trait ScoringStrategy: Send + Sync {
    /// Score an idea on 0-10 scale
    async fn score(&self, idea: &Idea) -> f32;

    /// Get detailed breakdown of scoring
    async fn score_detailed(&self, idea: &Idea) -> ScoreBreakdown;

    /// Detect patterns from idea
    async fn detect_patterns(&self, idea: &Idea) -> Vec<Pattern>;
}

#[derive(Debug, Clone)]
pub struct ScoreBreakdown {
    pub overall: f32,
    pub mission_alignment: f32,
    pub strategic_fit: f32,
    pub pattern_risks: f32,
    pub reasoning: String,
}

#[derive(Debug, Clone)]
pub struct Pattern {
    pub name: String,
    pub severity: PatternSeverity,
    pub description: String,
}

#[derive(Debug, Clone, PartialEq)]
pub enum PatternSeverity {
    High,
    Medium,
    Low,
}
```

**Step 2: Move Telos scoring to implementation**

Rename current `src/scoring.rs` → `src/scoring/telos_impl.rs` and implement the trait:

```rust
use async_trait::async_trait;
use super::interface::*;
use crate::telos::TelosConfig;

pub struct TelosScoringStrategy {
    telos: TelosConfig,
}

impl TelosScoringStrategy {
    pub fn new(telos: TelosConfig) -> Self {
        Self { telos }
    }
}

#[async_trait]
impl ScoringStrategy for TelosScoringStrategy {
    async fn score(&self, idea: &Idea) -> f32 {
        // Existing scoring logic here
        todo!()
    }

    async fn score_detailed(&self, idea: &Idea) -> ScoreBreakdown {
        // Existing detailed scoring here
        todo!()
    }

    async fn detect_patterns(&self, idea: &Idea) -> Vec<Pattern> {
        // Existing pattern detection here
        todo!()
    }
}
```

**Step 3: Update mod.rs to export trait**

Create `src/scoring/mod.rs`:

```rust
pub mod interface;
pub mod telos_impl;

pub use interface::{ScoringStrategy, ScoreBreakdown, Pattern, PatternSeverity};
pub use telos_impl::TelosScoringStrategy;
```

**Step 4: Update main.rs to use trait**

```rust
let scoring_strategy: Box<dyn ScoringStrategy> =
    Box::new(TelosScoringStrategy::new(telos_config));

// Pass strategy to commands
// Example: dump_command(idea, &scoring_strategy).await
```

**Step 5: Commit**

```bash
git add src/scoring/
git commit -m "refactor: abstract scoring into pluggable strategy interface"
```

---

## Phase 2: Testing & Quality Assurance

### Task 4: Write Integration Tests for Configuration Loading

**Objective:** Ensure configuration loading works reliably across different scenarios

**Files:**
- Create: `tests/config_integration_test.rs` (new)
- Create: `tests/fixtures/sample_telos.md` (test fixture)

**Step 1: Create test fixture**

Create `tests/fixtures/sample_telos.md`:

```markdown
# Sample Telos for Testing

## Goals
- G1: Ship product (Deadline: 2025-12-31)
- G2: Build community (Deadline: 2025-12-31)
- G3: Establish credibility (Deadline: 2025-12-31)
- G4: Create income stream (Deadline: 2025-12-31)

## Strategies
- S1: Focus on shipping
- S2: One stack rule
- S3: Build in public
- S4: MVP mindset

## Stack
- Primary: Rust
- Secondary: Python

## Failure Patterns
- Context-switching
- Perfectionism
- Procrastination
```

**Step 2: Write configuration tests**

Create `tests/config_integration_test.rs`:

```rust
use std::path::PathBuf;
use telos_idea_matrix::config::ConfigPaths;

#[test]
fn test_load_config_from_environment_variable() {
    let fixture_path = PathBuf::from("tests/fixtures/sample_telos.md");
    std::env::set_var("TELOS_FILE", &fixture_path);

    let config = ConfigPaths::load().expect("Config should load from env");
    assert_eq!(config.telos_file, fixture_path);
}

#[test]
fn test_load_config_from_current_directory() {
    // Create temporary telos.md in temp directory
    let temp_dir = tempfile::tempdir().unwrap();
    let telos_path = temp_dir.path().join("telos.md");
    std::fs::write(&telos_path, "# Test Telos").unwrap();

    // Change to temp directory
    let original_dir = std::env::current_dir().unwrap();
    std::env::set_current_dir(&temp_dir).unwrap();
    std::env::remove_var("TELOS_FILE");

    let config = ConfigPaths::load().expect("Config should load from cwd");
    assert_eq!(config.telos_file, PathBuf::from("telos.md"));

    // Restore original directory
    std::env::set_current_dir(original_dir).unwrap();
}

#[test]
fn test_ensure_directories_created() {
    let temp_dir = tempfile::tempdir().unwrap();
    let config = ConfigPaths {
        telos_file: PathBuf::from("telos.md"),
        data_dir: temp_dir.path().join("data"),
        log_dir: temp_dir.path().join("logs"),
        config_file: None,
    };

    assert!(!config.data_dir.exists());
    config.ensure_directories_exist().unwrap();
    assert!(config.data_dir.exists());
    assert!(config.log_dir.exists());
}
```

**Step 3: Run tests**

```bash
cargo test --test config_integration_test
```

Expected: All tests pass

**Step 4: Commit**

```bash
git add tests/
git commit -m "test: add configuration loading integration tests"
```

---

### Task 5: Add Unit Tests for Scoring Strategy

**Objective:** Verify scoring logic works correctly with trait abstraction

**Files:**
- Create: `tests/scoring_strategy_test.rs` (new)

**Step 1: Write scoring strategy tests**

Create `tests/scoring_strategy_test.rs`:

```rust
use telos_idea_matrix::scoring::ScoringStrategy;
use telos_idea_matrix::types::Idea;

#[tokio::test]
async fn test_telos_scoring_returns_valid_range() {
    let telos = load_test_telos().await.unwrap();
    let strategy = TelosScoringStrategy::new(telos);

    let idea = Idea {
        content: "Build a Rust project aligned with goals".to_string(),
        ..Default::default()
    };

    let score = strategy.score(&idea).await;
    assert!(score >= 0.0 && score <= 10.0);
}

#[tokio::test]
async fn test_pattern_detection_identifies_context_switching() {
    let telos = load_test_telos().await.unwrap();
    let strategy = TelosScoringStrategy::new(telos);

    let idea = Idea {
        content: "Switch to learning JavaScript and Node.js".to_string(),
        ..Default::default()
    };

    let patterns = strategy.detect_patterns(&idea).await;
    assert!(patterns.iter().any(|p| p.name.contains("context-switch")));
}

async fn load_test_telos() -> Result<TelosConfig> {
    telos::load(&PathBuf::from("tests/fixtures/sample_telos.md")).await
}
```

**Step 2: Run tests**

```bash
cargo test --test scoring_strategy_test
```

Expected: All tests pass

**Step 3: Commit**

```bash
git add tests/scoring_strategy_test.rs
git commit -m "test: add scoring strategy unit tests"
```

---

### Task 6: Add Cargo Test Coverage to CI

**Objective:** Ensure tests run in CI pipeline

**Files:**
- Modify: `Cargo.toml` (add test configuration)
- Create: `.github/workflows/test.yml` (new)

**Step 1: Verify test configuration in Cargo.toml**

```toml
[dev-dependencies]
tempfile = "3.8"
tokio = { version = "1.35", features = ["full", "testing-util"] }
```

**Step 2: Create GitHub Actions workflow**

Create `.github/workflows/test.yml`:

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: dtolnay/rust-toolchain@stable
      - uses: Swatinem/rust-cache@v2

      - name: Run tests
        run: cargo test --all-features --verbose

      - name: Run clippy
        run: cargo clippy --all-targets -- -D warnings

      - name: Check formatting
        run: cargo fmt -- --check
```

**Step 3: Commit**

```bash
git add Cargo.toml .github/workflows/test.yml
git commit -m "ci: add GitHub Actions test workflow"
```

---

## Phase 3: Docker & Deployment

### Task 7: Create Dockerfile for Cross-Platform Support

**Objective:** Enable consistent execution on any system with Docker

**Files:**
- Create: `Dockerfile` (new)
- Create: `.dockerignore` (new)
- Create: `docker-compose.yml` (new)
- Create: `docs/DOCKER_GUIDE.md` (new)

**Step 1: Create Dockerfile**

Create `Dockerfile`:

```dockerfile
# Build stage
FROM rust:1.75-slim as builder

WORKDIR /build

# Install dependencies
RUN apt-get update && apt-get install -y \
    pkg-config \
    libssl-dev \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*

# Copy source
COPY Cargo.toml Cargo.lock ./
COPY src ./src
COPY migrations ./migrations

# Build release binary
RUN cargo build --release

# Runtime stage
FROM debian:bookworm-slim

WORKDIR /app

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*

# Copy binary from builder
COPY --from=builder /build/target/release/tm /usr/local/bin/tm

# Create data directory
RUN mkdir -p /data /config /logs

# Set environment
ENV TELOS_FILE=/config/telos.md

# Default command
ENTRYPOINT ["tm"]
CMD ["--help"]
```

**Step 2: Create .dockerignore**

Create `.dockerignore`:

```
target/
.git/
.gitignore
logs/
data/
*.db
*.sqlite
*.sqlite3
.DS_Store
.cargo/
benches/
README.md
DOCUMENTATION.md
PRD.md
TODO.md
```

**Step 3: Create docker-compose.yml**

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  telos-matrix:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: telos-matrix
    volumes:
      # Mount your telos.md file
      - ./telos.md:/config/telos.md:ro
      # Mount data directory for persistence
      - telos-data:/data
      # Mount logs directory
      - telos-logs:/logs
    environment:
      - TELOS_FILE=/config/telos.md
      - RUST_LOG=info
    stdin_open: true
    tty: true

volumes:
  telos-data:
  telos-logs:
```

**Step 4: Create Docker guide**

Create `docs/DOCKER_GUIDE.md`:

```markdown
# Docker Setup Guide

## Quick Start

### 1. Build the Docker Image

```bash
docker build -t telos-matrix:latest .
```

### 2. Run with Your Telos File

```bash
# Place your telos.md in current directory
docker run -it \
  -v $(pwd)/telos.md:/config/telos.md:ro \
  -v telos-data:/data \
  telos-matrix:latest dump "Your idea"
```

### 3. Using Docker Compose (Recommended)

```bash
# Copy your telos.md to current directory
cp /path/to/your/telos.md .

# Start the container
docker-compose up -d

# Run commands
docker-compose exec telos-matrix dump "Your idea"
docker-compose exec telos-matrix review
docker-compose exec telos-matrix prune
```

## Volume Mapping

- `/config/telos.md` - Your Telos configuration (read-only)
- `/data` - Persistent database and idea storage
- `/logs` - Application logs

## Environment Variables

```bash
docker run \
  -e TELOS_FILE=/config/telos.md \
  -e RUST_LOG=debug \
  telos-matrix:latest dump "idea"
```

## Advanced: Custom Ollama Integration

If you want to use local Ollama:

```yaml
# docker-compose.yml with Ollama
version: '3.8'
services:
  ollama:
    image: ollama/ollama:latest
    ports:
      - "11434:11434"
    volumes:
      - ollama-models:/root/.ollama

  telos-matrix:
    build: .
    depends_on:
      - ollama
    environment:
      - OLLAMA_HOST=http://ollama:11434
    volumes:
      - ./telos.md:/config/telos.md:ro
      - telos-data:/data

volumes:
  ollama-models:
  telos-data:
```

## Troubleshooting

**Permission denied errors:**
```bash
docker-compose exec -u root telos-matrix chmod 777 /data
```

**Database locked:**
Ensure only one container instance is running:
```bash
docker-compose down
docker-compose up -d
```
```

**Step 5: Test Docker build**

```bash
docker build -t telos-matrix:test .
```

Expected: Build succeeds

**Step 6: Commit**

```bash
git add Dockerfile .dockerignore docker-compose.yml docs/DOCKER_GUIDE.md
git commit -m "feat: add Docker support for cross-platform deployment"
```

---

### Task 8: Create Docker CI Workflow

**Objective:** Build and test Docker image in CI

**Files:**
- Create: `.github/workflows/docker.yml` (new)

**Step 1: Create Docker workflow**

Create `.github/workflows/docker.yml`:

```yaml
name: Docker Build

on:
  push:
    branches: [ main, develop ]
    tags: [ 'v*' ]
  pull_request:
    branches: [ main ]

jobs:
  docker:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
      - uses: actions/checkout@v3

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2

      - name: Build Docker image
        uses: docker/build-push-action@v4
        with:
          context: .
          push: false
          tags: telos-matrix:${{ github.sha }}

      - name: Test Docker image
        run: |
          docker build -t telos-matrix:test .
          docker run --rm telos-matrix:test --version || docker run --rm telos-matrix:test --help
```

**Step 2: Commit**

```bash
git add .github/workflows/docker.yml
git commit -m "ci: add Docker build workflow"
```

---

## Phase 4: GitHub Repository Setup

### Task 9: Create Comprehensive README for GitHub

**Objective:** Make project immediately understandable to new users

**Files:**
- Modify: `README.md` (replace personal-focused version)
- Create: `CONTRIBUTING.md` (new)
- Create: `.github/ISSUE_TEMPLATE/bug_report.md` (new)
- Create: `.github/ISSUE_TEMPLATE/feature_request.md` (new)

**Step 1: Rewrite README.md**

Update `README.md`:

```markdown
# Telos Idea Matrix

**Smart idea capture + personalized goal alignment for busy humans**

[![Tests](https://github.com/YOUR_USERNAME/telos-idea-matrix/workflows/Tests/badge.svg)](https://github.com/YOUR_USERNAME/telos-idea-matrix/actions)
[![Docker Build](https://github.com/YOUR_USERNAME/telos-idea-matrix/workflows/Docker%20Build/badge.svg)](https://github.com/YOUR_USERNAME/telos-idea-matrix/actions)

## What Is This?

Telos Idea Matrix is a command-line tool that captures your ideas and evaluates them against *your personal goals and strategies* instead of generic priority metrics.

Unlike generic todo/idea managers, TIM:
- **Personalizes to your framework**: Configure it with your Telos file (or any goal document)
- **Detects personal failure patterns**: Identifies when ideas trigger context-switching, perfectionism, or other traps
- **Provides objective scoring**: 0-10 scale aligned with YOUR specific mission, not arbitrary metrics
- **Works offline**: Local SQLite database, optional AI enhancement with Ollama

Perfect for founders, builders, and anyone with decision paralysis.

## Quick Start

### Prerequisites
- Rust 1.75+ OR Docker
- Your personal Telos/goal document (see [Configuration](docs/CONFIGURATION.md))

### Installation

**Option 1: Cargo (from source)**
```bash
git clone https://github.com/YOUR_USERNAME/telos-idea-matrix
cd telos-idea-matrix
cargo build --release
./target/release/tm --help
```

**Option 2: Docker**
```bash
docker build -t telos-matrix .
docker run -it -v $(pwd)/telos.md:/config/telos.md:ro telos-matrix dump "Your idea"
```

### Configure Your Telos

Create `telos.md` in your project:

```markdown
# My Telos

## Goals
- G1: Ship product (Deadline: 2025-12-31)
- G2: Build community (Deadline: 2025-12-31)
- G3: Establish credibility (Deadline: 2025-12-31)
- G4: Create income stream (Deadline: 2025-12-31)

## Strategies
- S1: Focus on shipping
- S2: One stack rule
- S3: Build in public
- S4: MVP mindset

## Stack
- Primary: Rust
- Secondary: Python

## Failure Patterns
- Context-switching (I abandon projects for new shiny ideas)
- Perfectionism (I gold-plate features instead of shipping)
- Procrastination (I consume instead of build)
```

Or [use Ray's as a template](examples/ray-telos.md).

### First Capture

```bash
export TELOS_FILE=$(pwd)/telos.md
tm dump "Build a tool that helps analyze sentiment"
```

The system will:
1. Capture your idea
2. Score it against your goals (0-10)
3. Detect if it triggers your known failure patterns
4. Store it locally for future review

## Usage Examples

```bash
# Capture and analyze an idea
tm dump "Rewrite API in Go"

# Review all ideas with scoring
tm review

# Review only high-potential ideas
tm review --min-score 7.0

# Find ideas to prune (low value, old)
tm review --pruning

# Interactive idea pruning
tm prune

# Link related ideas
tm link add <idea_id> <idea_id> --type "depends-on"

# Search ideas
tm review --search "dashboard"

# Export ideas
tm export --format csv --output ideas.csv
```

## Architecture

```
telos-idea-matrix/
├── src/
│   ├── main.rs                 # CLI entry point
│   ├── config.rs               # Configuration loading
│   ├── telos.rs                # Telos parsing
│   ├── commands/               # CLI commands
│   ├── scoring/                # Pluggable scoring strategies
│   ├── database.rs             # SQLite operations
│   ├── ai/                     # Optional Ollama integration
│   └── errors.rs               # Error types
├── tests/                      # Integration tests
├── docs/                       # Documentation
├── Dockerfile                  # Container image
└── docker-compose.yml          # Local development
```

## Configuration

See [Configuration Guide](docs/CONFIGURATION.md) for:
- Setting up your Telos file
- Environment variables
- Custom config file locations
- Data directory structure

## Documentation

- [Configuration Guide](docs/CONFIGURATION.md) - Customize for your goals
- [Docker Guide](docs/DOCKER_GUIDE.md) - Run in containers
- [Architecture](docs/ARCHITECTURE.md) - Technical deep dive
- [API Guide](docs/API.md) - Programmatic usage
- [Contributing](CONTRIBUTING.md) - Help improve TIM

## Features

### ✅ Implemented
- [x] Idea capture and storage
- [x] Telos-aligned scoring
- [x] Personal failure pattern detection
- [x] Idea review and browsing
- [x] Relationship mapping (links between ideas)
- [x] Optional Ollama AI integration
- [x] Structured logging
- [x] Docker containerization
- [x] CLI with aliases

### 🚀 Planned
- [ ] Web UI for review/browsing
- [ ] Sync across devices
- [ ] Telos file generation wizard
- [ ] Advanced analytics
- [ ] Team/family shared ideas (fork)

## Why Telos?

[Telos](https://en.wikipedia.org/wiki/Telos) means "purpose" or "end goal." This system helps you:

1. **Capture**: Write ideas without friction
2. **Align**: Check against your stated purpose
3. **Decide**: Get objective scoring instead of decision paralysis
4. **Act**: Focus on ideas that truly matter

## Testing

```bash
# Run all tests
cargo test

# Run specific test category
cargo test config
cargo test scoring

# With verbose output
cargo test -- --nocapture
```

## Contributing

We welcome contributions! See [Contributing Guidelines](CONTRIBUTING.md) for:
- Development setup
- Code style
- Testing requirements
- Pull request process

## License

[Your License - e.g., MIT]

## Support

- 📖 [Documentation](docs/)
- 🐛 [Report Issues](https://github.com/YOUR_USERNAME/telos-idea-matrix/issues)
- 💬 [Discussions](https://github.com/YOUR_USERNAME/telos-idea-matrix/discussions)

## Acknowledgments

Built with:
- [Rust](https://www.rust-lang.org/) + [Tokio](https://tokio.rs/)
- [SQLx](https://github.com/launchbadge/sqlx) for database
- [Ollama](https://ollama.ai) for optional AI
- [Clap](https://docs.rs/clap/) for CLI

---

**Made with ❤️ by [Ray Yacub](https://github.com/YOUR_USERNAME)**
```

**Step 2: Create CONTRIBUTING.md**

Create `CONTRIBUTING.md`:

```markdown
# Contributing to Telos Idea Matrix

We love contributions! This document explains how to get started.

## Development Setup

```bash
git clone https://github.com/YOUR_USERNAME/telos-idea-matrix
cd telos-idea-matrix
cargo build --dev
cargo test
```

## Code Style

- Use `cargo fmt` before committing
- Follow clippy suggestions: `cargo clippy`
- Write tests for new functionality
- Document public APIs with doc comments

```bash
cargo fmt
cargo clippy --all-targets -- -D warnings
```

## Testing Requirements

- Unit tests for new functions
- Integration tests for new commands
- Run full test suite: `cargo test --all-features`

## Commit Guidelines

- Use conventional commits: `feat:`, `fix:`, `docs:`, `test:`
- Example: `feat: add export to JSON format`
- Keep commits focused and atomic

## Pull Request Process

1. Fork the repository
2. Create feature branch: `git checkout -b feat/your-feature`
3. Make changes with tests
4. Run full test suite: `cargo test --all-features`
5. Push to fork: `git push origin feat/your-feature`
6. Create Pull Request with:
   - Clear title describing change
   - Description of what and why
   - Reference any related issues

## Areas for Contribution

- Bug fixes
- Performance improvements
- Documentation enhancements
- Test coverage
- New commands or features
- Docker/deployment improvements

See [Issues](https://github.com/YOUR_USERNAME/telos-idea-matrix/issues) for ideas.

## Questions?

Open a [Discussion](https://github.com/YOUR_USERNAME/telos-idea-matrix/discussions) or join our community.
```

**Step 3: Create issue templates**

Create `.github/ISSUE_TEMPLATE/bug_report.md`:

```markdown
---
name: Bug Report
about: Report a bug or unexpected behavior
---

## Description
<!-- Clear description of the issue -->

## Steps to Reproduce
1.
2.
3.

## Expected Behavior
<!-- What should happen -->

## Actual Behavior
<!-- What actually happens -->

## Environment
- OS: [macOS/Linux/Windows]
- Rust Version: `rustc --version`
- Installation: [Cargo/Docker]

## Logs
<!-- Any error messages or logs -->
\`\`\`
Paste logs here
\`\`\`

## Workaround
<!-- If you've found a workaround, describe it -->
```

Create `.github/ISSUE_TEMPLATE/feature_request.md`:

```markdown
---
name: Feature Request
about: Suggest an enhancement
---

## Description
<!-- Clear description of the feature -->

## Problem It Solves
<!-- What problem or workflow would this improve? -->

## Proposed Solution
<!-- How should this work? -->

## Alternatives Considered
<!-- Other approaches? -->

## Example Usage
\`\`\`bash
# Show how you'd use this
\`\`\`
```

**Step 4: Commit**

```bash
git add README.md CONTRIBUTING.md .github/ISSUE_TEMPLATE/
git commit -m "docs: create comprehensive GitHub documentation"
```

---

### Task 10: Set Up GitHub Release Process

**Objective:** Automate releases and binary distribution

**Files:**
- Create: `.github/workflows/release.yml` (new)
- Create: `docs/RELEASE_PROCESS.md` (new)

**Step 1: Create release workflow**

Create `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  create-release:
    runs-on: ubuntu-latest
    outputs:
      upload_url: ${{ steps.create_release.outputs.upload_url }}
    steps:
      - uses: actions/create-release@v1
        id: create_release
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          tag_name: ${{ github.ref }}
          release_name: Release ${{ github.ref }}
          draft: false
          prerelease: false

  build-releases:
    needs: create-release
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        include:
          - os: ubuntu-latest
            target: x86_64-unknown-linux-gnu
            name: linux-x64
          - os: ubuntu-latest
            target: aarch64-unknown-linux-gnu
            name: linux-arm64
          - os: macos-latest
            target: x86_64-apple-darwin
            name: macos-x64
          - os: macos-latest
            target: aarch64-apple-darwin
            name: macos-arm64

    steps:
      - uses: actions/checkout@v3

      - uses: dtolnay/rust-toolchain@stable
        with:
          targets: ${{ matrix.target }}

      - name: Build
        run: cargo build --release --target ${{ matrix.target }}

      - name: Create archive
        run: |
          mkdir -p staging
          cp target/${{ matrix.target }}/release/tm staging/
          cp README.md docs/CONFIGURATION.md staging/
          tar czf tm-${{ matrix.name }}.tar.gz -C staging .

      - name: Upload Release Asset
        uses: actions/upload-release-asset@v1
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          upload_url: ${{ needs.create-release.outputs.upload_url }}
          asset_path: ./tm-${{ matrix.name }}.tar.gz
          asset_name: tm-${{ matrix.name }}.tar.gz
          asset_content_type: application/gzip
```

**Step 2: Create release documentation**

Create `docs/RELEASE_PROCESS.md`:

```markdown
# Release Process

## Automated Releases

Releases are automated via GitHub Actions. Simply tag a commit:

```bash
git tag v0.2.0
git push origin v0.2.0
```

This will:
1. Create GitHub release
2. Build binaries for Linux (x64, arm64) and macOS (Intel, Apple Silicon)
3. Upload archives to release page
4. Generate checksums

## Manual Release (Alternative)

```bash
# Update version in Cargo.toml
cargo build --release

# Test the binary
./target/release/tm --version

# Create tag
git tag -a v0.2.0 -m "Release version 0.2.0"

# Push
git push origin main --tags
```

## Installation from Release

```bash
# Download and extract
wget https://github.com/YOUR_USERNAME/telos-idea-matrix/releases/download/v0.2.0/tm-linux-x64.tar.gz
tar xzf tm-linux-x64.tar.gz
sudo cp tm /usr/local/bin/

# Verify
tm --version
```

## Versioning

We follow semantic versioning:
- MAJOR: Breaking changes
- MINOR: New features (backwards compatible)
- PATCH: Bug fixes

Examples:
- v0.1.0 → v0.2.0: New command or feature
- v0.2.0 → v0.2.1: Bug fix
- v0.2.0 → v1.0.0: Breaking API change
```

**Step 3: Commit**

```bash
git add .github/workflows/release.yml docs/RELEASE_PROCESS.md
git commit -m "ci: add automated release workflow"
```

---

### Task 11: Create .gitignore and Clean Repository

**Objective:** Prevent committing unnecessary files to GitHub

**Files:**
- Create: `.gitignore` (new)
- Remove: Files that should not be tracked

**Step 1: Create comprehensive .gitignore**

Create `.gitignore`:

```
# Rust
/target/
/Cargo.lock
**/*.rs.bk
*.pdb

# Database and logs
*.db
*.sqlite
*.sqlite3
/data/
/logs/

# macOS
.DS_Store
.AppleDouble
.LSOverride
._*

# IDE
.vscode/
.idea/
*.swp
*.swo
*~
.DS_Store

# Environment
.env
.env.local
.env.*.local

# Temporary files
*.tmp
tmp/
temp/

# OS
Thumbs.db
.DS_Store

# Build artifacts
dist/
build/

# Dependencies
node_modules/
```

**Step 2: Remove tracked files that should be ignored**

```bash
git rm --cached .DS_Store
git rm --cached -r target/
git rm --cached -r logs/
git rm --cached -r data/*.db
git rm --cached -r .cargo/
```

**Step 3: Commit gitignore**

```bash
git add .gitignore
git commit -m "chore: add comprehensive gitignore"
```

---

## Phase 5: Documentation & Examples

### Task 12: Create Architecture Documentation

**Objective:** Help users and contributors understand system design

**Files:**
- Create: `docs/ARCHITECTURE.md` (new)
- Create: `docs/API.md` (new)

**Step 1: Write architecture documentation**

Create `docs/ARCHITECTURE.md`:

```markdown
# Architecture Overview

## System Design

Telos Idea Matrix follows a layered architecture:

```
┌─────────────────────────────────────────┐
│         CLI Layer (Clap)                │
├─────────────────────────────────────────┤
│      Command Handlers (dump, review)    │
├─────────────────────────────────────────┤
│  Scoring Strategy  │  Database Layer    │
│  (Pluggable)       │  (SQLx + SQLite)   │
├─────────────────────────────────────────┤
│  Configuration  │  Telos  │  AI Layer   │
│  Module         │ Parsing │ (Ollama)    │
└─────────────────────────────────────────┘
```

### Modules

#### Configuration (`src/config.rs`)
- Loads user's Telos file from multiple sources (env var, file, config)
- Manages data directories (database, logs)
- Validates configuration on startup

#### Telos (`src/telos.rs`)
- Parses Telos/goal documents
- Extracts goals, strategies, and failure patterns
- Provides typed access to configuration

#### Scoring Strategy (`src/scoring/`)
- Trait-based architecture for extensibility
- `TelosScoringStrategy` implementation
- Provides both simple scores and detailed breakdowns

#### Database (`src/database.rs`)
- SQLite with async SQLx driver
- Connection pooling for concurrency
- Migrations for schema management
- Idea storage and retrieval

#### AI Layer (`src/ai/`)
- Optional Ollama integration
- Circuit breaker for resilience
- Fallback to rule-based analysis

#### Commands (`src/commands/`)
- `dump`: Capture ideas with analysis
- `review`: Browse and filter ideas
- `prune`: Remove low-value ideas
- `link`: Connect related ideas
- `export`: Output in various formats

## Data Flow

### Idea Capture
```
User Input
    ↓
CLI (clap)
    ↓
Load Config + Telos
    ↓
Scoring Strategy.score()
    ↓
Detect Patterns
    ↓
Database.insert_idea()
    ↓
Display Result
```

### Idea Review
```
Database.get_ideas()
    ↓
Apply Filters (score, date, search)
    ↓
Format for Display
    ↓
User Sees Results
```

## Extension Points

### Add Custom Scoring Strategy

Implement the `ScoringStrategy` trait:

```rust
pub struct MyCustomStrategy {
    // Your fields
}

#[async_trait]
impl ScoringStrategy for MyCustomStrategy {
    async fn score(&self, idea: &Idea) -> f32 {
        // Your logic
    }

    async fn detect_patterns(&self, idea: &Idea) -> Vec<Pattern> {
        // Pattern detection
    }
}
```

Then use in main:
```rust
let strategy: Box<dyn ScoringStrategy> = Box::new(MyCustomStrategy::new());
```

### Add New Commands

1. Create `src/commands/mycommand.rs`
2. Implement command logic
3. Register in `main.rs` CLI
4. Add tests in `tests/mycommand_test.rs`

### Add Custom AI Provider

Extend `src/ai/mod.rs` with new provider:

```rust
pub enum AiProvider {
    Ollama(OllamaClient),
    Custom(CustomClient),
}
```

## Performance Considerations

### Database Optimization
- Connection pooling (min: 2, max: 10)
- Indexes on frequently queried columns
- Async operations with Tokio

### Memory Efficiency
- Lazy loading of ideas in review
- Streaming for large exports
- Proper cleanup in error paths

### Async Pattern
- Non-blocking I/O throughout
- Proper cancellation support
- Timeout protection on external calls

## Error Handling

Comprehensive error types:
- `ConfigError`: Configuration loading issues
- `DatabaseError`: SQL and database errors
- `ScoringError`: Scoring logic errors
- `AiError`: AI integration errors
- `ValidationError`: Input validation errors

All errors implement `std::error::Error` with helpful messages.
```

**Step 2: Create API documentation**

Create `docs/API.md`:

```markdown
# API Documentation

## Command-Line Interface

### Global Options

```bash
tm [OPTIONS] <COMMAND>
```

Options:
- `--help`: Show help message
- `--version`: Show version
- `--no-ai`: Disable AI integration, use rules only
- `--debug`: Show debug logs

### Commands

#### dump

Capture and analyze a new idea.

```bash
tm dump [OPTIONS] [IDEA_TEXT]
```

Options:
- `--interactive, -i`: Open editor for longer ideas
- `--quick, -q`: Skip detailed analysis
- `--ai-only`: Use AI analysis only (no rules)

Examples:
```bash
# Inline idea
tm dump "Build a dashboard"

# Interactive mode
tm dump --interactive

# Quick capture
tm dump -q "Shopping list"
```

#### review

Browse and filter stored ideas.

```bash
tm review [OPTIONS]
```

Options:
- `--limit N`: Show top N ideas
- `--min-score N.N`: Filter by minimum score
- `--max-score N.N`: Filter by maximum score
- `--search TEXT`: Full-text search
- `--older-than DAYS`: Show ideas older than X days
- `--tags TAG1,TAG2`: Filter by tags
- `--pruning`: Show low-value candidates for deletion
- `--sort FIELD`: Sort by (score, date, etc.)

Examples:
```bash
# All ideas
tm review

# High-potential ideas
tm review --min-score 7.0 --limit 10

# Search
tm review --search "dashboard"

# Ideas to prune
tm review --pruning --older-than 30 --max-score 3.0
```

#### prune

Interactive removal of low-value ideas.

```bash
tm prune [OPTIONS]
```

Options:
- `--auto`: Auto-delete ideas matching criteria (use carefully!)
- `--max-score N.N`: Only prune below this score
- `--older-than DAYS`: Only prune older than X days

Examples:
```bash
# Interactive pruning
tm prune

# Auto-delete low-value old ideas
tm prune --auto --max-score 2.0 --older-than 60
```

#### link

Manage relationships between ideas.

```bash
tm link <SUBCOMMAND>
```

Subcommands:
- `add`: Create relationship between ideas
- `remove`: Delete relationship
- `show`: View all relationships for an idea

Examples:
```bash
# Create dependency link
tm link add idea-123 idea-456 --type "depends-on"

# See relationships
tm link show idea-123

# Remove relationship
tm link remove idea-123 idea-456
```

#### export

Export ideas in various formats.

```bash
tm export [OPTIONS]
```

Options:
- `--format FORMAT`: csv, json, markdown
- `--output FILE`: Output file path
- `--filter-score N.N`: Only export above score

Examples:
```bash
# JSON export
tm export --format json --output ideas.json

# CSV for spreadsheet
tm export --format csv --output ideas.csv --filter-score 6.0

# Markdown for sharing
tm export --format markdown --output ideas.md
```

#### health

Check system status.

```bash
tm health
```

Shows:
- Database connectivity
- AI service availability
- Configuration status
- Storage usage

### Exit Codes

- `0`: Success
- `1`: General error
- `2`: Configuration error
- `3`: Database error
- `4`: Validation error

## Configuration API

See [Configuration Guide](CONFIGURATION.md) for setting up the system.

## Programmatic Usage

While primarily a CLI tool, the core modules can be used as a library:

```rust
use telos_idea_matrix::{
    config::ConfigPaths,
    telos,
    scoring::TelosScoringStrategy,
    database::Database,
    types::Idea,
};

#[tokio::main]
async fn main() -> Result<()> {
    // Load configuration
    let config = ConfigPaths::load()?;

    // Load telos
    let telos_config = telos::load(&config.telos_file).await?;

    // Create scoring strategy
    let scorer = TelosScoringStrategy::new(telos_config);

    // Score an idea
    let idea = Idea::new("My idea");
    let score = scorer.score(&idea).await;

    println!("Score: {}", score);
    Ok(())
}
```
```

**Step 3: Commit**

```bash
git add docs/ARCHITECTURE.md docs/API.md
git commit -m "docs: add architecture and API documentation"
```

---

### Task 13: Create Example Configurations

**Objective:** Help users get started with templates

**Files:**
- Create: `examples/telos_templates.md` (new)
- Create: `examples/startup_founder_telos.md` (new)
- Create: `examples/engineer_telos.md` (new)

**Step 1: Create templates documentation**

Create `examples/telos_templates.md`:

```markdown
# Telos File Templates

Use these templates as starting points for your own Telos configuration.

## What is Telos?

Telos means "end goal" or "purpose." Your Telos defines:
- **Goals**: Your 4 main objectives with deadlines
- **Strategies**: How you'll achieve them
- **Stack**: Your technology/skill focus
- **Failure Patterns**: Behaviors that sabotage you

## How to Use Templates

1. Copy a template below
2. Modify to match your actual goals and patterns
3. Save as `telos.md`
4. Run: `export TELOS_FILE=$(pwd)/telos.md`

See [Configuration Guide](../docs/CONFIGURATION.md) for more info.
```

Create `examples/startup_founder_telos.md`:

```markdown
# Telos: Startup Founder

## Goals
- G1: Ship MVP to market (Deadline: 2025-03-31)
- G2: Acquire first 100 paying customers (Deadline: 2025-06-30)
- G3: Raise seed funding (Deadline: 2025-09-30)
- G4: Build sustainable unit economics (Deadline: 2025-12-31)

## Strategies
- S1: Fast iteration over perfection (2-week sprints)
- S2: Revenue focus from day one (no free users until $100MRR)
- S3: Customer intimacy (weekly user interviews)
- S4: Focus on one product, one market (no pivots)

## Stack
- Primary: TypeScript/React (frontend), Python (backend)
- Secondary: AWS infrastructure
- Avoid: Rust, new frameworks, native mobile

## Failure Patterns
- **Perfectionism**: I redesign UI instead of shipping features
- **Goldplating**: I add features customers don't ask for
- **Distraction**: I build integrations before core product works
- **Tool-switching**: I adopt new tools mid-project
- **Analysis paralysis**: I research markets instead of selling
```

Create `examples/engineer_telos.md`:

```markdown
# Telos: Principal Engineer

## Goals
- G1: Publish 3 technical articles (Deadline: 2025-06-30)
- G2: Contribute to high-impact OSS (Deadline: 2025-09-30)
- G3: Become expert in Rust systems (Deadline: 2025-12-31)
- G4: Build portfolio of production systems (Deadline: 2025-12-31)

## Strategies
- S1: Write-first development (doc ideas before coding)
- S2: One language focus (Rust as primary)
- S3: Public learning (share progress openly)
- S4: Production readiness always (proper error handling, tests, docs)

## Stack
- Primary: Rust, PostgreSQL
- Secondary: Python (data analysis only)
- Avoid: PHP, legacy systems, JavaScript frameworks

## Failure Patterns
- **Architecture-nerd**: I design overly complex systems
- **Yak-shaving**: I polish tooling instead of building
- **Framework-chasing**: I learn new frameworks instead of deepening expertise
- **Solo-coding**: I code alone for weeks without feedback
- **Blogging-procrastination**: I write blog posts instead of shipping code
```

**Step 2: Commit**

```bash
git add examples/
git commit -m "docs: add Telos configuration templates"
```

---

## Phase 6: Quality Gate & Release Preparation

### Task 14: Set Up .gitattributes and Line Endings

**Objective:** Ensure consistent line endings across platforms

**Files:**
- Create: `.gitattributes` (new)

**Step 1: Create git attributes**

Create `.gitattributes`:

```
# Auto detect text files and normalize line endings to LF
* text=auto

# Rust source files
*.rs text eol=lf

# Shell scripts
*.sh text eol=lf

# YAML
*.yml text eol=lf
*.yaml text eol=lf

# Markdown
*.md text eol=lf

# JSON
*.json text eol=lf
*.toml text eol=lf

# Binary files
*.db binary
*.sqlite binary
target/ export-ignore
```

**Step 2: Commit**

```bash
git add .gitattributes
git commit -m "chore: add git attributes for line endings"
```

---

### Task 15: Create LICENSE File

**Objective:** Make licensing explicit for GitHub distribution

**Files:**
- Create: `LICENSE` (new)

**Step 1: Choose and add license**

Create `LICENSE` (example with MIT):

```
MIT License

Copyright (c) 2025 Ray Yacub

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

**Step 2: Update Cargo.toml with license**

Modify `Cargo.toml`:

```toml
[package]
name = "telos-idea-matrix"
version = "0.1.0"
edition = "2021"
authors = ["Ray Yacub <ray@example.com>"]
description = "Idea capture + Telos-aligned analysis for decision paralysis"
license = "MIT"
repository = "https://github.com/YOUR_USERNAME/telos-idea-matrix"
homepage = "https://github.com/YOUR_USERNAME/telos-idea-matrix"
keywords = ["productivity", "ideas", "telos", "goals", "planning"]
categories = ["command-line-utilities"]
```

**Step 3: Commit**

```bash
git add LICENSE Cargo.toml
git commit -m "chore: add MIT license"
```

---

### Task 16: Create CHANGELOG

**Objective:** Track changes for future releases

**Files:**
- Create: `CHANGELOG.md` (new)

**Step 1: Create changelog**

Create `CHANGELOG.md`:

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Configuration abstraction for personalization
- Docker containerization and docker-compose setup
- Comprehensive GitHub documentation and templates
- Issue and PR templates
- Automated release workflows
- Integration test suite

### Changed
- Refactored Telos loading to support multiple file locations
- Scoring system now uses pluggable strategy pattern

### Fixed
- Hardcoded personal paths removed

## [0.1.0] - 2024-11-17

### Added
- Initial release
- Idea capture and storage
- Telos-aligned scoring
- Pattern detection
- Idea review and management
- Ollama AI integration
- SQLite database with connection pooling
- Structured logging
- CLI with aliases

[Unreleased]: https://github.com/YOUR_USERNAME/telos-idea-matrix/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/YOUR_USERNAME/telos-idea-matrix/releases/tag/v0.1.0
```

**Step 2: Commit**

```bash
git add CHANGELOG.md
git commit -m "docs: add changelog"
```

---

### Task 17: Final Quality Checks

**Objective:** Ensure code quality before GitHub release

**Files:**
- Run: Quality check commands

**Step 1: Run formatter**

```bash
cargo fmt --all
```

**Step 2: Run clippy**

```bash
cargo clippy --all-targets --all-features -- -D warnings
```

Expected: No warnings

**Step 3: Run all tests**

```bash
cargo test --all-features --verbose
```

Expected: All tests pass

**Step 4: Check documentation**

```bash
cargo doc --no-deps --open
```

Expected: All public items documented

**Step 5: Verify Docker build**

```bash
docker build -t telos-matrix:final .
```

Expected: Build succeeds without errors

**Step 6: Commit any formatting changes**

```bash
git add .
git commit -m "chore: format code and pass clippy checks"
```

---

### Task 18: Prepare Initial GitHub Release

**Objective:** Create clean repository for GitHub

**Files:**
- Execute: GitHub push commands

**Step 1: Initialize commit**

```bash
git add .
git commit -m "initial: clean production-ready release

- Abstracted Telos configuration for personalization
- Docker containerization for cross-platform support
- Comprehensive documentation for GitHub
- CI/CD workflows for testing and releases
- Removed personal hardcoded paths
- Ready for public distribution"
```

**Step 2: Create initial release tag**

```bash
git tag -a v0.1.0 -m "Initial production-ready release"
```

**Step 3: Configure repository metadata**

In GitHub:
1. Add repository description: "Smart idea capture + personalized goal alignment"
2. Add topics: `rust`, `productivity`, `ideas`, `goals`, `telos`
3. Enable discussions
4. Set up branch protection for `main`

**Step 4: Document next steps**

Create `docs/AFTER_GITHUB.md`:

```markdown
# Post-GitHub Setup Checklist

## Immediate Tasks
- [ ] Update README.md with actual GitHub URLs
- [ ] Update Cargo.toml with correct repository URL
- [ ] Update CI workflows with correct username
- [ ] Add example link in README
- [ ] Create first GitHub release from tag

## First Week
- [ ] Share with friends/family
- [ ] Gather initial feedback
- [ ] Document any setup issues
- [ ] Fix bugs reported

## First Month
- [ ] Monitor GitHub issues
- [ ] Improve documentation based on questions
- [ ] Add CI badge to README
- [ ] Consider draft web UI
```

**Step 5: Final commit**

```bash
git add docs/AFTER_GITHUB.md
git commit -m "docs: add post-GitHub setup checklist"
```

---

## Summary

This plan transforms telos-idea-matrix from a personal tool into a production-grade, shareable system in 18 concrete tasks across 6 phases:

### Phase 1: Architecture Refactoring
- Abstract configuration system
- Remove hardcoded paths
- Create pluggable scoring interface

### Phase 2: Testing & Quality
- Integration tests for config
- Unit tests for scoring
- CI test automation

### Phase 3: Docker & Deployment
- Create Dockerfile
- Set up docker-compose
- Add Docker CI workflow

### Phase 4: GitHub Repository
- Comprehensive README
- Contributing guidelines
- Issue templates
- Release automation

### Phase 5: Documentation
- Architecture guide
- API documentation
- Example configurations

### Phase 6: Quality & Release
- Git attributes
- License
- Changelog
- Final quality checks

### Estimated Effort
- **Quick path** (MVP for friends): 8-10 hours
- **Full path** (production-ready): 16-20 hours
- **All tasks** (comprehensive): 24-30 hours

### Trade-offs Addressed

**Generalization vs. Personalization**
- ✅ Supports both: Users customize with their Telos
- ✅ Extensible: Pluggable scoring strategies

**Complexity vs. Simplicity**
- ✅ Simple CLI interface
- ✅ Complex architecture hidden
- ✅ Docker for easy distribution

**Coverage vs. Scope**
- ✅ Core functionality stable
- ✅ Tests ensure reliability
- ✅ Documentation for contribution

---

**Ready to execute?** Choose your execution strategy:
1. **Subagent-Driven** - I dispatch fresh subagent per task, code reviews between
2. **Parallel Session** - Open new session with executing-plans skill for batch execution
```
