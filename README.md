# Telos Idea Matrix

[![CI Status](https://github.com/rayyacub/telos-idea-matrix/workflows/CI/badge.svg)](https://github.com/rayyacub/telos-idea-matrix/actions)
[![License](https://img.shields.io/github/license/rayyacub/telos-idea-matrix)](./LICENSE)
[![Rust](https://img.shields.io/badge/rust-1.75%2B-orange.svg)](https://www.rust-lang.org/)
[![Docker](https://img.shields.io/badge/docker-ready-blue.svg)](https://github.com/rayyacub/telos-idea-matrix/pkgs/container/telos-idea-matrix)

> **A command-line tool for evaluating ideas against your personal goals and values**

Stop drowning in a sea of brilliant ideas. Start shipping the ones that actually matter.

---

## What Is This?

**Telos Idea Matrix** is a CLI tool that helps you escape decision paralysis by providing instant, objective analysis of your ideas against your personal "Telos" — your life goals, mission, strategies, and known failure patterns.

Instead of letting ideas pile up in notebooks, scattered notes, or forgotten browser tabs, the Telos Idea Matrix captures them instantly and scores them on what matters to *you*. No more analysis paralysis. No more context switching. No more picking projects that don't align with where you want to go.

### The Problem It Solves

If you've ever experienced:
- **Decision paralysis** — Too many ideas, unable to choose which to pursue
- **Context switching** — Starting new projects before finishing existing ones
- **Misalignment** — Working on ideas that don't advance your actual goals
- **Perfectionism** — Getting stuck in planning instead of building
- **Pattern blindness** — Repeating the same mistakes without realizing it

...then this tool is for you.

### How It Works

1. **Define your Telos** — Write down your goals, strategies, tech stack, and failure patterns in a simple `telos.md` file
2. **Capture ideas instantly** — Run `tm dump "Your idea"` to get immediate scoring and analysis
3. **Get objective feedback** — See alignment scores, pattern warnings, and actionable insights
4. **Make informed decisions** — Review, filter, and prioritize ideas based on data, not gut feeling

The tool uses a multi-dimensional scoring system that weighs:
- **Mission Alignment (40%)** — Does this advance your core mission?
- **Anti-Patterns (35%)** — Does this trigger known failure patterns?
- **Strategic Fit (25%)** — Does this align with your current strategies?

---

## Quick Start

The fastest way to get started is with Docker (no Rust installation required):

```bash
# Pull the Docker image
docker pull ghcr.io/rayyacub/telos-idea-matrix:latest

# Create a sample telos.md file
cat > telos.md << 'EOF'
# My Telos

## Goals
- G1: Launch a profitable SaaS product (Deadline: 2025-12-31)
- G2: Build a personal brand through open source (Deadline: 2025-06-30)

## Strategies
- S1: Ship early and often, iterate based on feedback
- S2: Focus on one technology stack to maximize depth
- S3: Build in public to maintain accountability

## Stack
- Primary: Rust, TypeScript, PostgreSQL
- Secondary: Docker, GitHub Actions

## Failure Patterns
- Context switching: Starting new projects before finishing current ones
- Perfectionism: Over-engineering solutions before validating market fit
- Tutorial hell: Watching tutorials instead of building
EOF

# Analyze your first idea
docker run --rm -v $(pwd):/workspace ghcr.io/rayyacub/telos-idea-matrix:latest dump "Build a Rust CLI tool for personal productivity"
```

You should see output with alignment scores, detected patterns, and recommendations.

---

## Installation

### Option 1: Docker (Recommended for Quick Start)

Docker is the easiest way to get started without installing Rust or managing dependencies.

```bash
# Pull the latest image
docker pull ghcr.io/rayyacub/telos-idea-matrix:latest

# Create an alias for convenience (add to ~/.bashrc or ~/.zshrc)
alias tm='docker run --rm -v $(pwd):/workspace ghcr.io/rayyacub/telos-idea-matrix:latest'

# Now you can use tm directly
tm dump "Your idea here"
```

**Using Docker Compose:**

Create a `docker-compose.yml`:

```yaml
version: '3.8'
services:
  telos-matrix:
    image: ghcr.io/rayyacub/telos-idea-matrix:latest
    volumes:
      - .:/workspace
    working_dir: /workspace
```

Then run:

```bash
docker-compose run --rm telos-matrix dump "Your idea"
```

See [docs/DOCKER_GUIDE.md](./docs/DOCKER_GUIDE.md) for advanced Docker usage.

### Option 2: Cargo (For Rust Developers)

If you already have Rust installed:

```bash
# Install from crates.io (once published)
cargo install telos-idea-matrix

# Or install from source
cargo install --git https://github.com/rayyacub/telos-idea-matrix
```

### Option 3: Pre-built Binaries

Download platform-specific binaries from the [Releases](https://github.com/rayyacub/telos-idea-matrix/releases) page:

```bash
# Linux/macOS
wget https://github.com/rayyacub/telos-idea-matrix/releases/latest/download/tm-$(uname -s)-$(uname -m).tar.gz
tar -xzf tm-$(uname -s)-$(uname -m).tar.gz
sudo mv tm /usr/local/bin/

# Verify installation
tm --version
```

### Option 4: Build from Source

```bash
# Clone the repository
git clone https://github.com/rayyacub/telos-idea-matrix.git
cd telos-idea-matrix

# Build release binary
cargo build --release

# Binary will be at ./target/release/tm
./target/release/tm --version

# Optional: Install to system
cargo install --path .
```

---

## Configuration

The Telos Idea Matrix requires a `telos.md` file that defines your personal framework for evaluating ideas.

### Quick Configuration

Create a `telos.md` file in your project directory:

```markdown
# My Telos

## Goals
- G1: [Your first major goal] (Deadline: YYYY-MM-DD)
- G2: [Your second major goal] (Deadline: YYYY-MM-DD)
- G3: [Your third major goal] (Deadline: YYYY-MM-DD)

## Strategies
- S1: [Strategy for achieving goals]
- S2: [Another strategy]
- S3: [Another strategy]

## Stack
- Primary: [Your primary tech stack]
- Secondary: [Technologies you're willing to use]

## Failure Patterns
- [Pattern name]: [Description of pattern to avoid]
- [Another pattern]: [Description]
```

### Configuration Methods

The tool looks for configuration in this priority order:

1. **Environment variable** (recommended for Docker/CI):
   ```bash
   export TELOS_FILE=/path/to/your/telos.md
   tm dump "Your idea"
   ```

2. **Current directory**:
   ```bash
   cd /my/project
   tm dump "Your idea"  # Uses ./telos.md
   ```

3. **Config file** at `~/.config/telos-matrix/config.toml`:
   ```toml
   telos_file = "/path/to/your/telos.md"
   data_dir = "~/.local/share/telos-matrix"
   log_dir = "~/.cache/telos-matrix/logs"
   ```

4. **Interactive setup** — If no configuration is found, the tool will guide you through creating one

For detailed configuration options, file format specifications, and examples, see [docs/CONFIGURATION.md](./docs/CONFIGURATION.md).

---

## Usage Examples

### Basic Idea Capture

```bash
# Capture and analyze an idea
tm dump "Create a mobile app for tracking daily habits"

# Capture with additional context
tm dump "Build a Rust-based web scraper" --tags "automation,data"

# Quick scoring without storing
tm score "Learn Golang for backend development"
```

### Analyzing Ideas

```bash
# Analyze an idea without storing it
tm analyze "Should I start a YouTube channel about Rust?"

# Get detailed explanation of a stored idea
tm explain <idea-id>

# Compare two ideas side-by-side
tm compare <idea-id-1> <idea-id-2>
```

### Reviewing Ideas

```bash
# Review all captured ideas
tm review

# Review with score filtering
tm review --min-score 7.0

# Review ideas from the last 7 days
tm review --since "7 days ago"

# Review top 10 ideas
tm review --limit 10 --sort-by score

# Review ideas with specific tags
tm review --tags "rust,cli"

# Review pending ideas only
tm review --status pending
```

### Managing Ideas

```bash
# Update idea status
tm update <idea-id> --status "in-progress"

# Add tags to an idea
tm tag add <idea-id> "important" "urgent"

# Remove tags
tm tag remove <idea-id> "urgent"

# Archive an idea
tm archive <idea-id>

# Delete an idea permanently
tm delete <idea-id>
```

### Linking Related Ideas

```bash
# Create a relationship between two ideas
tm link create <source-id> <target-id> <relationship-type>

# View all relationships for an idea
tm link list <idea-id>

# Show related ideas with full details
tm link show <idea-id> [--type <relationship-type>]

# Find dependency paths between ideas
tm link path <from-id> <to-id>

# Remove a relationship
tm link remove <relationship-id>
```

**Relationship types:** `depends_on`, `blocked_by`, `blocks`, `part_of`, `parent`, `child`, `related_to`, `similar_to`, `duplicate`

**Learn more:**
- [Link Command User Guide](./docs/user-guide/link-command.md) - Complete documentation
- [Getting Started Tutorial](./docs/tutorials/getting-started-with-links.md) - Step-by-step guide
- [Quick Reference](./docs/quick-reference/link-cheatsheet.md) - Cheat sheet

### Pruning and Maintenance

```bash
# Automatically prune low-scoring ideas (score < 4.0)
tm prune --auto

# Prune with custom threshold
tm prune --threshold 5.0

# Dry-run to see what would be pruned
tm prune --dry-run

# Prune ideas older than 90 days with low scores
tm prune --older-than "90 days" --threshold 6.0
```

### Analytics and Insights

```bash
# View statistics about captured ideas
tm stats

# Show trend of idea scores over time
tm trends

# Export ideas to JSON
tm export --format json > ideas.json

# Export to CSV
tm export --format csv > ideas.csv
```

### AI-Powered Analysis

```bash
# Use AI for deeper analysis (requires Ollama)
tm dump "Build a personal finance tracker" --ai

# Use specific AI model
tm analyze "Create a blog about functional programming" --ai --model "llama2"

# Batch analyze all pending ideas with AI
tm batch analyze --status pending --ai
```

---

## Features

### Core Features ✅

- **Idea Capture & Storage** — Save ideas instantly with automatic timestamping and organization
- **Telos-Aligned Scoring** — Multi-dimensional scoring based on your personal goals and values
- **Pattern Detection** — Automatically identifies context-switching, perfectionism, and other anti-patterns
- **Flexible Configuration** — Multiple configuration methods (env, file, interactive)
- **Fast & Lightweight** — Built in Rust for performance and minimal resource usage
- **Cross-Platform** — Works on Linux, macOS, and Windows
- **Docker Support** — Pre-built Docker images for easy deployment
- **SQLite Storage** — Reliable, portable data storage with no external database required
- **Rich CLI** — Intuitive commands with helpful error messages and colored output
- **Data Export** — Export ideas to JSON, CSV, or Markdown formats

### Scoring System

The tool evaluates ideas across three dimensions:

1. **Mission Alignment (40% weight)**
   - How well does this idea advance your stated goals?
   - Does it contribute to your core mission?
   - Is it aligned with your long-term vision?

2. **Anti-Pattern Detection (35% weight)**
   - Does this trigger context-switching patterns?
   - Does this indicate perfectionism or over-engineering?
   - Does this align with procrastination patterns (e.g., tutorial consumption)?
   - Does it lack accountability mechanisms?

3. **Strategic Fit (25% weight)**
   - Does this align with your current strategies?
   - Does it use your committed technology stack?
   - Is the timing right given your deadlines?

**Final Score**: 0-10 scale
- **8-10**: Highly aligned, strong candidate for immediate action
- **6-7.9**: Good alignment, worth considering
- **4-5.9**: Weak alignment, proceed with caution
- **0-3.9**: Misaligned, likely a distraction

### Advanced Features 🚀

- **AI Integration** — Optional Ollama integration for LLM-powered analysis and insights
- **Relationship Mapping** — Connect related ideas with typed relationships (depends-on, related-to, etc.)
- **Batch Operations** — Perform bulk operations on multiple ideas at once
- **Tag System** — Organize ideas with custom tags and filters
- **Status Tracking** — Track idea lifecycle from pending → in-progress → completed → archived
- **Trend Analysis** — Visualize patterns in your ideation over time
- **Circuit Breaker** — Resilient AI integration with automatic fallback
- **Logging & Debugging** — Comprehensive logging with configurable verbosity

---

## Architecture

The Telos Idea Matrix is built with modern Rust practices and production-ready patterns:

### Technology Stack

- **Language**: Rust 1.75+
- **Async Runtime**: Tokio for concurrent operations
- **CLI Framework**: Clap v4 for command parsing and validation
- **Database**: SQLite with SQLx for type-safe queries
- **AI Integration**: Ollama client with circuit breaker pattern
- **Serialization**: Serde for JSON/TOML handling
- **Logging**: Tracing + tracing-subscriber for structured logging

### Design Patterns

- **Repository Pattern** — Clean abstraction over data storage
- **Builder Pattern** — Flexible object construction for complex types
- **Strategy Pattern** — Pluggable scoring algorithms
- **Command Pattern** — Modular CLI command structure
- **Circuit Breaker** — Resilient external service integration

### Project Structure

```
src/
├── main.rs              # Entry point and CLI routing
├── commands/            # CLI command implementations
│   ├── dump.rs         # Idea capture and analysis
│   ├── review.rs       # Idea browsing and filtering
│   ├── analyze.rs      # Analysis without storage
│   ├── prune.rs        # Cleanup and maintenance
│   └── ...
├── config/             # Configuration management
│   ├── loader.rs       # Multi-source config loading
│   └── paths.rs        # Path resolution
├── scoring/            # Scoring engine
│   ├── engine.rs       # Core scoring logic
│   └── weights.rs      # Scoring weights and thresholds
├── telos/              # Telos parsing and processing
│   ├── parser.rs       # Markdown parsing
│   └── validator.rs    # Configuration validation
├── database/           # Data layer
│   ├── repository.rs   # Database operations
│   └── migrations.rs   # Schema migrations
├── ai/                 # AI integration
│   ├── client.rs       # Ollama client
│   └── circuit.rs      # Circuit breaker
└── types/              # Shared data structures
```

---

## Roadmap

### Completed ✅

- [x] Core idea capture and storage
- [x] Multi-dimensional scoring system
- [x] Pattern detection (context-switching, perfectionism)
- [x] SQLite database with migrations
- [x] Docker support with pre-built images
- [x] Comprehensive test suite
- [x] CI/CD pipeline
- [x] Multiple installation methods
- [x] Configuration flexibility
- [x] AI integration (Ollama)

### In Progress 🚧

- [ ] Web UI for idea visualization
- [ ] Enhanced analytics dashboard
- [ ] Import/export for additional formats (YAML, XML)

### Planned 🚀

- [ ] Team collaboration features (shared Telos, idea voting)
- [ ] Integration with task management systems (Todoist, Notion, etc.)
- [ ] Browser extension for capturing web-based ideas
- [ ] Mobile companion app (iOS/Android)
- [ ] Plugin system for custom scoring algorithms
- [ ] Advanced analytics and trend visualization
- [ ] Multi-user support with role-based permissions
- [ ] API server mode for programmatic access
- [ ] GitHub/GitLab integration for tracking issues and PRs
- [ ] Slack/Discord bot for team idea capture

Want to help with any of these? Check out our [Contributing Guide](./CONTRIBUTING.md)!

---

## Contributing

We welcome contributions of all kinds! Whether you're fixing bugs, adding features, improving documentation, or sharing feedback, your input is valuable.

### Ways to Contribute

- **Report bugs** — Found an issue? [Open a bug report](https://github.com/rayyacub/telos-idea-matrix/issues/new)
- **Request features** — Have an idea? [Share it with us](https://github.com/rayyacub/telos-idea-matrix/issues/new)
- **Improve documentation** — Fix typos, clarify instructions, add examples
- **Write code** — Pick up an issue labeled `good-first-issue` or `help-wanted`
- **Share feedback** — Tell us how you're using the tool and what could be better

### Getting Started

1. Read our [Contributing Guide](./CONTRIBUTING.md) for setup instructions
2. Check out the [open issues](https://github.com/rayyacub/telos-idea-matrix/issues)
3. Join the discussion in [GitHub Discussions](https://github.com/rayyacub/telos-idea-matrix/discussions)

### Development Quick Start

```bash
# Clone and build
git clone https://github.com/rayyacub/telos-idea-matrix.git
cd telos-idea-matrix
cargo build

# Run tests
cargo test

# Run with logging
RUST_LOG=debug cargo run -- dump "Test idea"
```

See [CONTRIBUTING.md](./CONTRIBUTING.md) for detailed guidelines on code style, testing, and pull requests.

---

## License

This project is licensed under the **MIT License** — see the [LICENSE](./LICENSE) file for details.

Copyright (c) 2025 Telos Idea Matrix Contributors

---

## Support & Community

- **Documentation**: [docs/](./docs/)
- **Issues**: [GitHub Issues](https://github.com/rayyacub/telos-idea-matrix/issues)
- **Discussions**: [GitHub Discussions](https://github.com/rayyacub/telos-idea-matrix/discussions)
- **Releases**: [GitHub Releases](https://github.com/rayyacub/telos-idea-matrix/releases)

---

## Acknowledgments

This project was inspired by the need to combat decision paralysis and context-switching in personal productivity. Special thanks to:

- The Rust community for excellent tooling and libraries
- Early beta testers who provided invaluable feedback
- Contributors who helped shape the Telos framework
- Everyone who has shared their ideas and use cases

---

**Stop collecting ideas. Start shipping the ones that matter.**

Try it now: `docker run --rm ghcr.io/rayyacub/telos-idea-matrix:latest --help`