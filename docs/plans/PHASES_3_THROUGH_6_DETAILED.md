# Phases 3-6: Docker, GitHub, Documentation & Release - Detailed Breakdown

> **For Subagent Execution**: Each phase is divisible into independent tasks. Complete Phase 1 & 2 first.

---

# Phase 3: Docker & Cross-Platform Distribution (2-3 hours)

**Goal**: Enable any system to run the tool without installing Rust.

## Task 3.1: Create Dockerfile

**Subagent: Build `Dockerfile` with multi-stage approach**

### What We're Building

A Docker image that:
1. Builds the binary in a heavyweight build stage
2. Copies only the binary to a lightweight runtime stage
3. Includes runtime dependencies (SQLite, CA certs)
4. Exposes environment variables for config
5. Results in ~150MB final image

### Requirements

**Output**:
- `Dockerfile` created
- Multi-stage build (builder + runtime)
- ~150MB final image size
- Builds successfully: `docker build -t telos-matrix:test .`

**Exit Criteria**:
- [ ] Dockerfile created
- [ ] `docker build` succeeds
- [ ] Image runs: `docker run telos-matrix:test --help`
- [ ] Image size is reasonable (~150-200MB)

### Implementation

Create `Dockerfile`:

```dockerfile
# Stage 1: Build
FROM rust:1.75-slim as builder

WORKDIR /build

# Install build dependencies
RUN apt-get update && apt-get install -y \
    pkg-config \
    libssl-dev \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*

# Copy Cargo files
COPY Cargo.toml Cargo.lock ./

# Copy source code
COPY src ./src
COPY migrations ./migrations

# Build release binary
RUN cargo build --release 2>&1 | head -100

# Stage 2: Runtime
FROM debian:bookworm-slim

WORKDIR /app

# Install runtime dependencies only
RUN apt-get update && apt-get install -y \
    ca-certificates \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*

# Copy binary from builder
COPY --from=builder /build/target/release/tm /usr/local/bin/tm

# Create data and config directories
RUN mkdir -p /data /config /logs

# Set environment
ENV TELOS_FILE=/config/telos.md
ENV RUST_LOG=info

# Default command
ENTRYPOINT ["tm"]
CMD ["--help"]
```

### Build & Test Locally

```bash
# Build
docker build -t telos-matrix:test .

# Test
docker run telos-matrix:test --help
docker run telos-matrix:test --version
```

### Deliverables

1. `Dockerfile` created
2. Builds without errors
3. Image runs successfully
4. Image size ~150-200MB

---

## Task 3.2: Create .dockerignore

**Subagent: Create `.dockerignore` to exclude unnecessary files**

### What We're Doing

Prevent large/unnecessary files from being copied into Docker context.

### Implementation

Create `.dockerignore`:

```
# Version control
.git
.gitignore
.github

# Build artifacts
/target
*.lock

# IDE/Editor
.vscode
.idea
*.swp
*.swo

# Data
*.db
*.sqlite
*.sqlite3
/data
/logs

# Documentation
*.md
docs/plans
DOCUMENTATION.md
README.md
TODO.md

# OS
.DS_Store
Thumbs.db

# Temporary
tmp/
temp/
*.tmp

# Build tools
scripts/
Makefile

# Tests (optional, can include)
# tests/

# Benchmarks
benches/
```

### Deliverables

1. `.dockerignore` created
2. Properly formatted
3. Reduces Docker context size

---

## Task 3.3: Create docker-compose.yml

**Subagent: Create `docker-compose.yml` for local development**

### What We're Building

A docker-compose file that makes running the Docker image easy:
- Mounts user's telos.md file
- Persists data across container restarts
- Handles environment variables
- Can run multiple commands easily

### Requirements

**Output**:
- `docker-compose.yml` created
- Version 3.8+
- Services defined correctly
- Volumes configured

**Exit Criteria**:
- [ ] File created
- [ ] `docker-compose up` works
- [ ] `docker-compose exec telos-matrix dump "test"` works
- [ ] Data persists after container restart

### Implementation

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  telos-matrix:
    build:
      context: .
      dockerfile: Dockerfile

    container_name: telos-matrix

    image: telos-matrix:latest

    # Mount volumes
    volumes:
      # Mount telos.md from current directory (read-only)
      - ./telos.md:/config/telos.md:ro

      # Persist data between runs
      - telos-data:/data

      # Persist logs
      - telos-logs:/logs

    # Environment variables
    environment:
      - TELOS_FILE=/config/telos.md
      - RUST_LOG=info

    # Allow interactive input
    stdin_open: true
    tty: true

    # Restart policy
    restart: unless-stopped

volumes:
  telos-data:
    driver: local

  telos-logs:
    driver: local
```

### Usage Examples

Include in comments at top:

```yaml
# Usage:
#
# 1. Place your telos.md in current directory
# 2. docker-compose up -d
# 3. docker-compose exec telos-matrix dump "your idea"
# 4. docker-compose logs -f  (view logs)
# 5. docker-compose down     (stop)
#
# Example:
# $ cp ~/my-telos.md ./telos.md
# $ docker-compose up -d
# $ docker-compose exec telos-matrix dump "Build a SaaS"
```

### Deliverables

1. `docker-compose.yml` created
2. All services defined
3. Volumes configured
4. Environment variables set
5. Usage examples in comments

---

## Task 3.4: Create GitHub Actions Docker Build Workflow

**Subagent: Create `.github/workflows/docker.yml`**

### What We're Building

Automated Docker image builds on:
- Push to main (creates `latest` tag)
- Push of version tags (creates versioned tags)
- Pull requests (builds but doesn't push)

### Requirements

**Output**:
- `.github/workflows/docker.yml` created
- Builds on push/PR
- Tests Docker image
- Optional: Push to registry (can skip for now)

**Exit Criteria**:
- [ ] Workflow created
- [ ] Builds Docker image
- [ ] Tests `docker run` works
- [ ] No secrets required

### Implementation

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

    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2

      - name: Build Docker image
        uses: docker/build-push-action@v4
        with:
          context: .
          push: false
          tags: telos-matrix:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

      - name: Test Docker image
        run: |
          docker build -t telos-matrix:test .
          echo "Testing: --help"
          docker run --rm telos-matrix:test --help
          echo "✅ Docker image works"
```

### Deliverables

1. `.github/workflows/docker.yml` created
2. Builds on push/PR
3. Tests image runs
4. Caching configured

---

## Task 3.5: Create Docker Setup Guide

**Subagent: Create `docs/DOCKER_GUIDE.md`**

### What We're Writing

Clear instructions for:
- Building Docker image
- Running with docker-compose
- Custom configuration
- Troubleshooting

### Requirements

**Output**:
- `docs/DOCKER_GUIDE.md` created
- ~300 lines
- Quick start section
- Advanced section
- Troubleshooting section

### Structure

```markdown
# Docker Setup Guide

## Quick Start
[Copy-paste example]

## Using docker-compose
[copy-compose example]

## Custom Configuration
[env vars, mounts]

## Troubleshooting
[common issues]

## Advanced: Custom Ollama
[optional AI setup]
```

### Deliverables

1. `docs/DOCKER_GUIDE.md` created
2. Quick start examples
3. Troubleshooting section
4. References main README

---

## Phase 3 Completion Checklist

- [ ] Dockerfile created and builds successfully
- [ ] `.dockerignore` created
- [ ] `docker-compose.yml` created and tested
- [ ] `.github/workflows/docker.yml` created
- [ ] `docs/DOCKER_GUIDE.md` created
- [ ] Local test: `docker build -t tm . && docker run tm --help` works
- [ ] Local test: `docker-compose up && docker-compose exec telos-matrix dump "test"`

### Phase 3 Commits

```bash
# Commit 1: Docker setup files
git add Dockerfile .dockerignore docker-compose.yml
git commit -m "feat: add Docker containerization

- Multi-stage build reduces image size to ~150MB
- docker-compose for easy local development
- Environment variable configuration
- Persistent data volumes"

# Commit 2: Docker CI/CD
git add .github/workflows/docker.yml
git commit -m "ci: add Docker build workflow

- Builds on push and PR
- Tests Docker image runs
- Caching for faster builds"

# Commit 3: Docker documentation
git add docs/DOCKER_GUIDE.md
git commit -m "docs: add Docker setup and usage guide

- Quick start examples
- docker-compose usage
- Troubleshooting
- Custom configuration"
```

---

# Phase 4: GitHub Infrastructure & Professional Presence (2-3 hours)

**Goal**: Create professional GitHub project with clear contribution path.

## Task 4.1: Rewrite README.md for GitHub Audience

**Subagent: Update `README.md` for new user audience**

### What We're Changing

Current README assumes personal use. New README should:
- Target new users (not Ray)
- Explain what the tool does
- Show it works for anyone
- Multiple installation options
- Clear getting started

### Requirements

**Output**:
- `README.md` rewritten (~500 lines)
- Audience: potential users, not just Ray
- Installation options: source, binary, Docker
- Example usage with different Telos files
- Links to other docs

**Exit Criteria**:
- [ ] README rewrites for general audience
- [ ] Includes: what, why, how, examples
- [ ] 3 installation methods
- [ ] Links to Configuration, Docker guides
- [ ] License badge, CI badge

### Structure

```markdown
# Telos Idea Matrix

[badges]

## What Is This?

[2-3 paragraphs explaining the tool and who it's for]

## Quick Start

[Copy-paste example]

## Installation

### Option 1: Cargo
[Steps]

### Option 2: Pre-built Binary
[Steps]

### Option 3: Docker
[Steps]

## Configuration

[Link to CONFIGURATION.md]

## Usage Examples

[10+ examples]

## Features

[✅ Done, 🚀 Planned]

## Contributing

[Link to CONTRIBUTING.md]

## License

[License info]
```

### Key Changes

- Remove personal references
- Add feature list
- Show Docker option
- Multiple installation methods
- Clear examples
- Links to detailed guides

### Deliverables

1. `README.md` rewritten
2. ~500 lines
3. Audience-appropriate
4. All sections complete
5. Links to other docs

---

## Task 4.2: Create CONTRIBUTING.md

**Subagent: Create `CONTRIBUTING.md`**

### What We're Writing

Guidelines for developers who want to:
- Build from source
- Run tests
- Submit changes
- Code style expectations

### Requirements

**Output**:
- `CONTRIBUTING.md` created
- ~300 lines
- Clear setup steps
- Code style guide
- Testing requirements
- PR process

**Exit Criteria**:
- [ ] File created
- [ ] Development setup documented
- [ ] Code style requirements clear
- [ ] Testing requirements clear
- [ ] PR process documented

### Structure

```markdown
# Contributing

## Development Setup

[Clone, build, test]

## Code Style

[Format, clippy, idioms]

## Testing Requirements

[Unit, integration, coverage]

## Pull Request Process

[Steps from branch to merge]

## Areas for Contribution

[What we need help with]
```

### Deliverables

1. `CONTRIBUTING.md` created
2. Complete development setup
3. Code style requirements
4. Testing expectations
5. PR process explained

---

## Task 4.3: Create Issue Templates

**Subagent: Create `.github/ISSUE_TEMPLATE/` directory with templates**

### What We're Creating

Two templates:
- `bug_report.md` - for bug reports
- `feature_request.md` - for feature requests

### Requirements

**Output**:
- Both templates created
- Clear sections
- Help users provide good info
- Pre-filled content

**Exit Criteria**:
- [ ] Directory created: `.github/ISSUE_TEMPLATE/`
- [ ] `bug_report.md` created
- [ ] `feature_request.md` created
- [ ] Both have clear sections
- [ ] Instructions for reporter

### Implementation

`.github/ISSUE_TEMPLATE/bug_report.md`:
```markdown
---
name: Bug Report
about: Report a bug or unexpected behavior
title: "[BUG] Brief description"
labels: bug
---

## Description
<!-- Clear, concise description of the issue -->

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
- Installation method: [Cargo/Docker/Binary]
- Rust version: `rustc --version`

## Logs/Error Output
<!-- Any error messages or logs -->
```

`.github/ISSUE_TEMPLATE/feature_request.md`:
```markdown
---
name: Feature Request
about: Suggest an enhancement
title: "[FEATURE] Brief idea"
labels: enhancement
---

## Description
<!-- Clear description of the feature -->

## Use Case
<!-- Why is this needed? -->

## Proposed Solution
<!-- How should it work? -->

## Example Usage
<!-- Show how you'd use this -->
```

### Deliverables

1. `.github/ISSUE_TEMPLATE/` directory created
2. `bug_report.md` created
3. `feature_request.md` created
4. Both templates properly formatted

---

## Task 4.4: Create Release Automation Workflow

**Subagent: Create `.github/workflows/release.yml`**

### What We're Building

Automated binary releases:
- Build on version tags (v0.1.0, etc.)
- Create release on GitHub
- Attach binaries
- Support multiple platforms

### Requirements

**Output**:
- `.github/workflows/release.yml` created
- Builds on version tags
- Creates GitHub release
- Attaches binaries

**Exit Criteria**:
- [ ] Workflow created
- [ ] Triggers on `v*` tags
- [ ] Builds multiple targets
- [ ] Creates release
- [ ] Attaches binaries

### Implementation

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
      - name: Create Release
        id: create_release
        uses: actions/create-release@v1
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          tag_name: ${{ github.ref }}
          release_name: Release ${{ github.ref }}
          draft: false
          prerelease: false

  build-release:
    needs: create-release
    runs-on: ${{ matrix.os }}

    strategy:
      matrix:
        include:
          - os: ubuntu-latest
            target: x86_64-unknown-linux-gnu
            name: linux-x64
            artifact: tm

          - os: macos-latest
            target: x86_64-apple-darwin
            name: macos-x64
            artifact: tm

          - os: macos-latest
            target: aarch64-apple-darwin
            name: macos-arm64
            artifact: tm

    steps:
      - name: Checkout
        uses: actions/checkout@v3

      - name: Install Rust
        uses: dtolnay/rust-toolchain@stable
        with:
          targets: ${{ matrix.target }}

      - name: Build
        run: cargo build --release --target ${{ matrix.target }}

      - name: Create archive
        run: |
          mkdir staging
          cp target/${{ matrix.target }}/release/${{ matrix.artifact }} staging/
          cp README.md LICENSE staging/
          tar czf tm-${{ matrix.name }}.tar.gz -C staging .

      - name: Upload asset
        uses: actions/upload-release-asset@v1
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          upload_url: ${{ needs.create-release.outputs.upload_url }}
          asset_path: ./tm-${{ matrix.name }}.tar.gz
          asset_name: tm-${{ matrix.name }}.tar.gz
          asset_content_type: application/gzip
```

### Deliverables

1. `.github/workflows/release.yml` created
2. Builds on version tags
3. Supports multiple platforms
4. Creates release
5. Attaches binaries

---

## Phase 4 Completion Checklist

- [ ] `README.md` rewritten for general audience
- [ ] `CONTRIBUTING.md` created
- [ ] `.github/ISSUE_TEMPLATE/bug_report.md` created
- [ ] `.github/ISSUE_TEMPLATE/feature_request.md` created
- [ ] `.github/workflows/release.yml` created
- [ ] All files properly formatted
- [ ] No broken links

### Phase 4 Commits

```bash
# Commit 1: GitHub documentation
git add README.md CONTRIBUTING.md
git commit -m "docs: rewrite README and add contributing guidelines

- README rewritten for general audience
- 3 installation methods documented
- Contributing guidelines for developers
- Code style and testing requirements"

# Commit 2: Issue templates
git add .github/ISSUE_TEMPLATE/
git commit -m "chore: add GitHub issue templates

- Bug report template
- Feature request template
- Clear sections for reporters"

# Commit 3: Release automation
git add .github/workflows/release.yml
git commit -m "ci: add automated release workflow

- Build binaries on version tags
- Create GitHub releases
- Attach binaries for multiple platforms"
```

---

# Phase 5: Documentation & Examples (3-4 hours)

**Goal**: Help users understand and extend the system.

## Task 5.1: Create Configuration Guide

**Subagent: Create `docs/CONFIGURATION.md`**

### What We're Writing

Step-by-step guide for users to:
- Create their own Telos file
- Set up telos-idea-matrix
- Understand default locations
- Customize configuration

### Requirements

**Output**:
- `docs/CONFIGURATION.md` created
- ~400 lines
- Multiple setup options
- Telos file format explained
- Examples provided

**Exit Criteria**:
- [ ] File created
- [ ] All setup methods documented
- [ ] Telos format explained
- [ ] Examples provided
- [ ] References example files

---

## Task 5.2: Create Architecture Documentation

**Subagent: Create `docs/ARCHITECTURE.md`**

### What We're Writing

Technical explanation of:
- System design
- Component interactions
- Data flow
- Extension points

### Requirements

**Output**:
- `docs/ARCHITECTURE.md` created
- ~400 lines
- Diagrams (ASCII art or descriptions)
- Component descriptions
- Extension examples

---

## Task 5.3: Create API Reference

**Subagent: Create `docs/API.md`**

### What We're Writing

Complete command reference:
- All commands
- All options
- Examples for each
- Exit codes

### Requirements

**Output**:
- `docs/API.md` created
- All commands documented
- Examples for each
- Exit codes explained

---

## Task 5.4: Create Example Telos Files

**Subagent: Create `examples/` directory with templates**

### What We're Creating

Multiple example Telos files:
- Generic template
- Startup founder example
- Engineer example
- Designer example (optional)

### Requirements

**Output**:
- `examples/` directory created
- 3-4 example Telos files
- Clear, customizable
- Comments explaining each section

---

## Phase 5 Completion Checklist

- [ ] `docs/CONFIGURATION.md` created
- [ ] `docs/ARCHITECTURE.md` created
- [ ] `docs/API.md` created
- [ ] `examples/` directory with 3+ files
- [ ] All examples well-commented
- [ ] Links between docs work

### Phase 5 Commits

```bash
git add docs/ examples/
git commit -m "docs: add comprehensive documentation

- Configuration guide
- Architecture documentation
- API reference
- Example Telos files"
```

---

# Phase 6: Polish & Release (2-3 hours)

**Goal**: Final quality checks and v0.1.0 release.

## Task 6.1: Add License File

**Subagent: Create `LICENSE` file**

### What We're Adding

MIT License (or your choice) with:
- Copyright notice
- License text
- Usage terms

### Implementation

Create `LICENSE` with MIT license text.

Update `Cargo.toml`:
```toml
[package]
license = "MIT"
repository = "https://github.com/YOUR_USERNAME/telos-idea-matrix"
homepage = "https://github.com/YOUR_USERNAME/telos-idea-matrix"
keywords = ["productivity", "ideas", "telos", "goals"]
categories = ["command-line-utilities"]
```

---

## Task 6.2: Create Changelog

**Subagent: Create `CHANGELOG.md`**

### What We're Writing

Changes from start to v0.1.0:

```markdown
# Changelog

## [0.1.0] - 2024-11-17

### Added
- Configuration abstraction for personalization
- Docker containerization
- GitHub CI/CD workflows
- Comprehensive documentation
- Example Telos files
- Integration tests

### Changed
- Telos loading now accepts configurable paths
- Scoring uses pluggable strategy pattern

### Fixed
- Removed hardcoded personal paths

[0.1.0]: https://github.com/YOUR_USERNAME/telos-idea-matrix/releases/tag/v0.1.0
```

---

## Task 6.3: Create .gitignore (Final Review)

**Subagent: Create/update `.gitignore`**

### What We're Ignoring

```
# Rust
/target/
Cargo.lock
*.rs.bk

# Data
*.db
*.sqlite
/data/
/logs/

# macOS
.DS_Store

# IDE
.vscode/
.idea/

# Temp
*.tmp
tmp/
```

---

## Task 6.4: Final Quality Gate

**Subagent: Run all checks and report**

### Steps

```bash
# 1. Clean build
cargo clean && cargo build

# 2. All tests
cargo test --all-features

# 3. Quality
cargo clippy --all-targets
cargo fmt --check

# 4. Release build
cargo build --release

# 5. Docker
docker build -t tm .

# 6. Git status
git status
```

---

## Task 6.5: Create Initial Release

**Subagent: Tag and push**

### Steps

```bash
# Create tag
git tag -a v0.1.0 -m "Initial production-ready release"

# Show what's included
git log --oneline v0.1.0...

# Ready for push
git push origin v0.1.0
```

---

## Phase 6 Completion Checklist

- [ ] `LICENSE` created
- [ ] `CHANGELOG.md` created
- [ ] `.gitignore` updated
- [ ] `Cargo.toml` has license metadata
- [ ] All quality checks pass
- [ ] Docker builds successfully
- [ ] Git ready to push
- [ ] v0.1.0 tag created

### Phase 6 Commits

```bash
git add LICENSE CHANGELOG.md .gitignore Cargo.toml
git commit -m "chore: add license and release metadata

- MIT License added
- Changelog created
- .gitignore finalized
- Cargo.toml updated with metadata"
```

---

## Overall Completion

After all 6 phases:

```
✅ Phase 1: Config abstraction
✅ Phase 2: Testing & CI/CD
✅ Phase 3: Docker
✅ Phase 4: GitHub infrastructure
✅ Phase 5: Documentation
✅ Phase 6: Polish & release

Total: 16-20 hours
Result: Production-ready v0.1.0
GitHub: Ready for public distribution
Users: Can now use with their own Telos file
```

---

## What Users Can Do After Release

1. Clone repository
2. Create their telos.md
3. Run: `tm dump "my idea"`
4. Get: Personalized scoring against THEIR goals
5. Extend: Add custom scoring strategies

**No code changes required.**

---

## Post-Release: First Week

Monitor:
- GitHub issues and discussions
- Setup problems (improve docs)
- Feature requests
- Bug reports

Document:
- Any new setup issues
- Common misunderstandings
- Popular feature requests

Plan:
- v0.1.1 (bug fix release)
- v0.2.0 (feature release with high demand)

---

**Phases 3-6 are subagent-ready. Complete in order after Phase 2 succeeds.**
