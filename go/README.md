# Telos Idea Matrix (Go)

A command-line tool and web application for evaluating ideas against your personal goals, values, and constraints.

**Status:** 🚧 In Development (Migration from Rust)

[![CI](https://github.com/rayyacub/telos-idea-matrix/actions/workflows/ci.yml/badge.svg)](https://github.com/rayyacub/telos-idea-matrix/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/dl/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

---

## What is Telos Idea Matrix?

Telos Idea Matrix helps you **quickly evaluate ideas** against your personal **mission, goals, and failure patterns** to determine which ideas you should pursue now vs. later.

Instead of chasing every shiny idea, Telos gives you an objective score (0-10) that tells you:
- ✅ Does this align with my current goals?
- ⚠️ Does this trigger my known failure patterns?
- 🔥 Should I prioritize this now or save it for later?

### Example

```bash
$ tm dump "Build an AI automation tool using Python and LangChain
to help hotel staff route guest requests. Can ship MVP in 30 days."

🎯 Final Score: 8.7/10
🔥 PRIORITIZE NOW

📊 Breakdown:
  Mission Alignment: 3.50/4.00
  Anti-Challenge: 3.20/3.50
  Strategic Fit: 2.00/2.50

✅ Idea saved (ID: a1b2c3d4)
```

---

## Features

### Current (Phase 0 - Foundation)

- ✅ Project structure and build system
- ✅ CI/CD pipeline (GitHub Actions)
- ✅ Development tooling (linters, formatters, tests)
- ✅ Comprehensive Rust behavior documentation

### Coming Soon

#### Phase 1: Core Domain (In Progress)
- 🚧 Data models (Idea, Score, Telos, Analysis)
- 🚧 Scoring engine (rule-based algorithm)
- 🚧 Telos parser (parse telos.md files)
- 🚧 Pattern detector (identify failure patterns)

#### Phase 2: CLI
- ⏳ `tm dump` - Quick-capture and analyze ideas
- ⏳ `tm analyze` - Analyze without saving
- ⏳ `tm review` - Review recent ideas
- ⏳ SQLite database integration

#### Phase 3: API Server
- ⏳ RESTful API (Go + Chi)
- ⏳ API documentation
- ⏳ Docker support

#### Phase 4: Web UI
- ⏳ SvelteKit frontend
- ⏳ Interactive scoring dashboard
- ⏳ Idea management interface

---

## Quick Start

### Prerequisites

- **Go 1.21+** - [Download](https://go.dev/dl/)
- **SQLite3** - Usually pre-installed
- **make** - For build commands

### Installation

```bash
# Clone the repository
git clone https://github.com/rayyacub/telos-idea-matrix.git
cd telos-idea-matrix

# Install dependencies
go mod download

# Build
make build

# Run tests
make test
```

### CLI Usage (Coming in Phase 2)

```bash
# Quick-capture an idea
tm dump "Your idea here"

# Analyze without saving
tm analyze "Your idea here"

# Review recent ideas
tm review --limit 10 --min-score 7.0
```

### Web UI (Coming in Phase 4)

```bash
# Start API server
make build-api
./bin/tm-web

# Open http://localhost:8080
```

---

## How It Works

### 1. Define Your Telos

Create a `telos.md` file describing your:
- **Problems** - What you're trying to solve
- **Missions** - Your overarching goals
- **Goals** - Specific, measurable targets
- **Challenges** - Your known failure patterns
- **Strategies** - Your anti-failure strategies

### 2. Capture Ideas

Use the CLI to quickly dump ideas:

```bash
tm dump "Build a SaaS tool for..."
```

### 3. Get Scored

Telos analyzes your idea across 3 dimensions:

#### Mission Alignment (40%)
- Does it use your existing skills?
- Is it AI/automation-related?
- Can you ship quickly?
- Does it have revenue potential?

#### Anti-Challenge (35%)
- Does it avoid context-switching?
- Can you prototype rapidly?
- Does it have built-in accountability?
- Will it reduce income anxiety?

#### Strategic Fit (25%)
- Does it enable flow sessions?
- Will it create reusable assets?
- Can you validate quickly?
- Is it scalable?

### 4. Make Decisions

Based on your score:
- **8.5+** 🔥 PRIORITIZE NOW
- **7.0+** ✅ GOOD ALIGNMENT
- **5.0+** ⚠️ CONSIDER LATER
- **<5.0** 🚫 AVOID FOR NOW

---

## Architecture

### Technology Stack

- **Language:** Go 1.21+
- **CLI:** Cobra + Viper
- **API:** Chi (HTTP router)
- **Database:** SQLite
- **Frontend:** SvelteKit + TypeScript (Phase 4)
- **Testing:** Go testing + Testify

### Project Structure

```
telos-idea-matrix-go/
├── cmd/                  # Application entry points
│   ├── cli/             # CLI tool
│   └── web/             # API server
├── internal/            # Private application code
│   ├── models/          # Domain models
│   ├── scoring/         # Scoring engine
│   ├── telos/           # Telos parser
│   ├── patterns/        # Pattern detection
│   ├── database/        # Database layer
│   ├── cli/             # CLI handlers
│   └── api/             # API handlers
├── pkg/                 # Public libraries
├── web/                 # SvelteKit frontend
├── test/                # Integration tests
└── docs/                # Documentation
```

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for detailed architecture.

---

## Development

### Setup Development Environment

```bash
# Install dependencies
go mod download

# Install development tools
brew install golangci-lint  # macOS
# or
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run tests
make test

# Run linters
make lint

# Format code
make fmt
```

### Development Workflow

We follow **Test-Driven Development (TDD)**:

1. **RED** - Write failing test
2. **GREEN** - Write minimal code to pass
3. **REFACTOR** - Improve code quality

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for complete workflow.

### Available Make Targets

```bash
make help           # Show all targets
make build          # Build all binaries
make build-cli      # Build CLI only
make build-api      # Build API server only
make test           # Run tests
make test-coverage  # Generate coverage report
make lint           # Run linters
make fmt            # Format code
make clean          # Remove build artifacts
```

---

## Documentation

- **[DEVELOPMENT.md](docs/DEVELOPMENT.md)** - Development guide
- **[RUST_REFERENCE.md](RUST_REFERENCE.md)** - Original Rust implementation reference
- **API.md** (Coming in Phase 3) - API documentation
- **CLI.md** (Coming in Phase 2) - CLI reference

---

## Migration Progress

This project is a **Rust → Go migration**. Progress:

- [x] **Phase 0:** Foundation & Documentation
  - [x] Project structure
  - [x] Build system (Makefile)
  - [x] CI/CD (GitHub Actions)
  - [x] Development docs
  - [x] Rust behavior documentation

- [ ] **Phase 1:** Core Domain Migration
  - [ ] Data models
  - [ ] Scoring engine
  - [ ] Telos parser
  - [ ] Pattern detector
  - [ ] Unit tests (>85% coverage)

- [ ] **Phase 2:** CLI Migration
  - [ ] `tm dump` command
  - [ ] `tm analyze` command
  - [ ] `tm review` command
  - [ ] Database integration
  - [ ] Integration tests

- [ ] **Phase 3:** API Server
  - [ ] RESTful API
  - [ ] API documentation
  - [ ] Docker support

- [ ] **Phase 4:** Web UI
  - [ ] SvelteKit frontend
  - [ ] Interactive dashboard
  - [ ] Deployment

---

## Why Go?

We're migrating from Rust to Go for:

1. **Faster iteration** - Simpler syntax, faster compile times
2. **Better ecosystem** - Rich web/CLI libraries
3. **Team readiness** - Easier to onboard contributors
4. **Deployment** - Single binary, no runtime dependencies

The Rust version is preserved in [RUST_REFERENCE.md](RUST_REFERENCE.md) as the behavioral specification.

---

## Contributing

Contributions welcome! Please:

1. Read [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md)
2. Create a feature branch
3. Write tests first (TDD)
4. Ensure tests pass: `make test`
5. Run linters: `make lint`
6. Submit a Pull Request

### Code of Conduct

Be respectful, constructive, and professional.

---

## License

MIT License - see [LICENSE](LICENSE) file for details.

---

## Roadmap

### Near-term (Q1 2025)
- ✅ Phase 0: Foundation complete
- 🚧 Phase 1: Core domain migration
- ⏳ Phase 2: CLI implementation

### Mid-term (Q2 2025)
- ⏳ Phase 3: API server
- ⏳ Phase 4: Web UI
- ⏳ Docker deployment

### Long-term (Q3+ 2025)
- LLM integration (Claude/GPT for enhanced analysis)
- Idea relationships and dependencies
- Collaborative telos (team goals)
- Mobile app

---

## Acknowledgments

- Original Rust implementation by Ray Acub
- Inspired by personal productivity systems and goal-setting frameworks
- Built with ❤️ and Go

---

## Links

- **GitHub**: https://github.com/rayyacub/telos-idea-matrix
- **Issues**: https://github.com/rayyacub/telos-idea-matrix/issues
- **Discussions**: https://github.com/rayyacub/telos-idea-matrix/discussions

---

**Built to help you ship what matters.** 🚀
