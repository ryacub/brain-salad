# Telos Idea Matrix

> AI-powered personal productivity system for decision-making and idea management

**Language:** Go 1.25.4+
**Status:** Production Ready ✅
**License:** MIT

## Overview

Telos Matrix helps combat decision paralysis and context-switching by providing intelligent, goal-aligned analysis of ideas. It scores ideas against your personal "Telos" (goals, mission, strategies) and provides actionable recommendations.

## Quick Start

### Installation

#### Quick Install (Linux/macOS)

```bash
curl -sSL https://raw.githubusercontent.com/ryacub/brain-salad/main/scripts/install.sh | bash
```

#### Manual Build

```bash
# Clone repository
git clone https://github.com/ryacub/brain-salad.git
cd brain-salad

# Build binary
go build -o tm ./cmd/cli

# Install globally (optional)
go install ./cmd/cli

# Run
./tm --help
```

#### First-Time Setup

```bash
# Initialize your workspace
tm init

# This creates:
# - ~/.telos/telos.md (your goals and mission)
# - ~/.telos/.env.example (configuration template)
# - Database directory
```

### Docker

```bash
# Using Docker Compose
cd deployments/docker
docker compose up

# Or build manually from project root
docker build -f deployments/docker/Dockerfile -t telos-matrix .
docker run -v $(pwd)/telos.md:/app/telos.md telos-matrix
```

### Usage

```bash
# 1. Create your telos.md file
mkdir -p ~/.telos
cat > ~/.telos/telos.md <<'TELOS'
# My Telos

## Goals
- Build AI-powered developer tools (Deadline: 2025-12-31)
- Launch profitable SaaS product (Deadline: 2025-06-30)

## Strategies
- Ship fast, iterate based on feedback
- Build in public for accountability

## Stack
- Primary: Go, TypeScript, Python
- Secondary: React

## Failure Patterns
- Context-switching between too many technologies
- Perfectionism before validation
TELOS

# 2. Capture and analyze an idea
tm dump "Build a CLI tool for managing GitHub issues with AI"

# 3. Use LLM for deeper analysis
tm dump "Start a podcast about developer tools" --use-ai --provider ollama

# 4. Review your ideas
tm review

# 5. Link related ideas
tm link create idea1 idea2 depends_on
```

## Features

### Multi-Dimensional Scoring
- **Mission Alignment (40%)** - Domain expertise, AI focus, execution capability
- **Anti-Challenge Detection (35%)** - Avoids context-switching, perfectionism, etc.
- **Strategic Fit (25%)** - Stack compatibility, shipping habits, revenue potential

### 5 LLM Providers
- **Ollama** - Local LLM (privacy-focused)
- **OpenAI** - GPT-4, GPT-5.1 support
- **Claude** - Anthropic Claude API
- **Custom** - Any HTTP/REST LLM endpoint
- **Rule-based** - Always available fallback (no LLM needed)

### User Experience Modes
- **Normal** - Standard analysis with scoring engine
- **Interactive (--interactive)** - Step-by-step analysis with confirmations
- **Quick (--quick)** - Ultra-fast capture without LLM overhead

### Advanced Features
- **Relationship Tracking** - Link ideas with dependencies, hierarchies
- **Analytics** - Trends, anomalies, performance metrics over time
- **Pattern Detection** - Identifies failure patterns automatically
- **Bulk Operations** - Batch analyze, update, tag, archive, delete
- **Clipboard Integration** - Paste ideas from clipboard
- **Export** - CSV, JSON, Markdown formats

### Shell Completion

Enable shell completion for faster command entry:

**Bash:**
```bash
source <(tm completion bash)
# Or install permanently:
sudo tm completion bash > /etc/bash_completion.d/tm
```

**Zsh:**
```bash
tm completion zsh > "${fpath[1]}/_tm"
exec zsh
```

**Fish:**
```bash
tm completion fish > ~/.config/fish/completions/tm.fish
```

**PowerShell:**
```powershell
tm completion powershell | Out-String | Invoke-Expression
```

## Commands

```bash
tm init                     # Initialize workspace (first-time setup)
tm dump <idea>              # Capture and analyze idea
tm dump --interactive       # Interactive mode with step-by-step LLM analysis
tm dump --quick            # Quick mode for fast capture
tm dump --use-ai           # Use LLM analysis (Ollama/OpenAI/Claude)

tm analyze <idea>           # Analyze without saving
tm analyze --id <id>        # Re-analyze existing idea
tm analyze --use-ai         # Use LLM for analysis

tm score <idea>             # Quick score without saving

tm review                   # Browse saved ideas
tm review --min-score 7.0   # Filter by score

tm prune                    # Clean up low-scoring ideas
tm prune --before 30d       # Archive ideas older than 30 days

tm link create <src> <tgt> <type>  # Link ideas
tm link list <id>                  # Show relationships
tm link path <src> <tgt>           # Find dependency paths

tm bulk analyze             # Re-analyze multiple ideas
tm bulk update --set-status archived  # Batch update

tm analytics trends         # Score trends over time
tm analytics anomaly        # Detect unusual patterns
tm analytics metrics        # System metrics

tm llm                      # Manage LLM providers
tm llm-config               # Configure providers
tm llm-health               # Check provider health

tm health                   # System health check
tm completion [shell]       # Generate shell completion script
```

## Architecture

**Language:** Go 1.25.4+
**Database:** SQLite (no external DB required)
**CLI Framework:** Cobra + Viper
**LLM Integration:** Multi-provider with fallback chain
**Testing:** Standard library + Testify

```
brain-salad/
├── cmd/                  # Application entry points
│   ├── cli/              # CLI binary
│   ├── web/              # API server
│   └── verify-wal/       # Database tools
├── internal/             # Private application code
│   ├── cli/              # Command implementations
│   ├── database/         # SQLite repository
│   ├── scoring/          # Scoring engine
│   ├── telos/            # Telos.md parser
│   ├── patterns/         # Pattern detection
│   ├── llm/              # LLM provider abstraction
│   ├── models/           # Domain models
│   ├── api/              # HTTP handlers
│   └── ...
├── pkg/                  # Public libraries
├── web/                  # SvelteKit frontend
├── test/                 # Integration tests
├── docs/                 # Documentation
├── examples/             # Example files
├── scripts/              # Build and utility scripts
├── configs/              # Configuration files
└── deployments/          # Deployment configurations
    ├── docker/           # Docker files
    ├── nginx/            # Nginx configs
    └── monitoring/       # Prometheus, Grafana
```

See [Architecture Documentation](./docs/ARCHITECTURE.md) for details.

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/cli/...

# Integration tests
go test ./test/...

# Using Makefile
make test
make test-coverage
```

## Configuration

Configuration via:
1. Environment variables (`TELOS_FILE`, `DB_PATH`)
2. Config file (`~/.config/telos-matrix/config.toml`)
3. Command-line flags (`--telos`, `--db`)

### LLM Provider Setup

```bash
# Ollama (local)
# Install: https://ollama.ai
ollama serve
tm dump "idea" --use-ai --provider ollama

# OpenAI
export OPENAI_API_KEY=sk-...
tm dump "idea" --use-ai --provider openai

# Claude
export CLAUDE_API_KEY=sk-ant-...
tm dump "idea" --use-ai --provider claude
```

See [Configuration Guide](./docs/CONFIGURATION.md) for full details.

## API Documentation

Full API documentation is available in OpenAPI format:
- [OpenAPI Specification](api/openapi.yaml)
- [Interactive API Docs](http://localhost:8080/api/docs) (when server running)

## Documentation

- [CLI Reference](./docs/CLI_REFERENCE.md) - Complete command documentation
- [Configuration](./docs/CONFIGURATION.md) - Setup and config options
- [Architecture](./docs/ARCHITECTURE.md) - System design and patterns
- [Docker Guide](./docs/DOCKER_GUIDE.md) - Container deployment
- [API Documentation](./docs/API.md) - REST API reference
- [Development](./CONTRIBUTING.md) - Contributing guide

## Contributing

Contributions welcome! See [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

```bash
# Fork repository
# Create feature branch
git checkout -b feature/my-feature

# Make changes, add tests
go test ./...

# Commit with conventional commits
git commit -m "feat: add new feature"

# Push and create PR
git push origin feature/my-feature
```

## License

MIT License - see [LICENSE](./LICENSE) file for details.

## Support

- **Issues:** [GitHub Issues](https://github.com/ryacub/brain-salad/issues)
- **Discussions:** [GitHub Discussions](https://github.com/ryacub/brain-salad/discussions)
- **Documentation:** [docs/](./docs/)

## Acknowledgments

Built with:
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [SQLite](https://www.sqlite.org/) - Database
- [Ollama](https://ollama.ai/) - Local LLM support
