# Brain-Salad

> Universal idea scoring CLI that helps anyone evaluate ideas based on their personal priorities

**Language:** Go 1.25.4+
**Status:** Production Ready
**License:** MIT

## Overview

Brain-Salad helps you decide which ideas to pursue by scoring them against what actually matters to you. Whether you're a potter, developer, event planner, or anyone with too many ideas — Brain-Salad learns your priorities through a quick interactive wizard and scores your ideas accordingly.

**No configuration files to write. No AI/tech jargon. Just answer a few questions and start scoring.**

## Quick Start

```bash
# Install
go install github.com/ryacub/brain-salad/cmd/cli@latest

# Run the setup wizard (2 minutes)
tm init

# Score an idea
tm score "Sell pottery at the farmer's market"

# Save and analyze an idea
tm dump "Build a mobile app for tracking inventory"
```

### What the wizard asks

```
Q1: You have two project ideas. Which do you start?
    [A] Simple and done by next week
    [B] Ambitious and done in a few months

Q2: Which feels worse?
    [A] Finishing something nobody cared about
    [B] Never finishing at all

... (5 questions total, plus your goals)
```

Your answers determine how ideas are scored. Value finishing things? Completion likelihood weighs more. Care about money? Revenue alignment matters more.

## Scoring Dimensions

Brain-Salad evaluates ideas across 6 universal dimensions:

| Dimension | Question | Weight |
|-----------|----------|--------|
| **Completion** | Will I actually finish this? | Based on your answers |
| **Skill Fit** | Can I do this with what I know? | Based on your answers |
| **Timeline** | How long until it's real? | Based on your answers |
| **Reward** | Does this give me what I want? | Based on your answers |
| **Sustainability** | Will I stay motivated? | Based on your answers |
| **Avoidance** | Does this dodge my pitfalls? | Based on your answers |

### Example Output

```
────────────────────────────────────────────────────────────────
Sell pottery at the farmer's market

Score: 8.2/10.0 — STRONG FIT

  Completion   ████████░░  1.7/2.0  Will I finish this?
  Skill Fit    ████████░░  1.6/2.0  Can I do this?
  Timeline     █████████░  1.8/2.0  How long?
  Reward       ███████░░░  1.4/2.0  What I want?
  Sustainability████████░░  0.8/1.0  Stay motivated?
  Avoidance    █████████░  0.9/1.0  Dodges pitfalls?

Insights:
  • This looks achievable and well-scoped
  • Strong alignment with your stated goals
────────────────────────────────────────────────────────────────
```

## Installation

### Quick Install (Linux/macOS)

```bash
curl -sSL https://raw.githubusercontent.com/ryacub/brain-salad/main/scripts/install.sh | bash
```

### Manual Build

```bash
git clone https://github.com/ryacub/brain-salad.git
cd brain-salad
make build
./bin/tm --help
```

### Docker

```bash
cd deployments/docker
docker compose up
```

## Commands

```bash
# Setup
tm init                     # Run the discovery wizard (creates ~/.brain-salad/profile.yaml)
tm profile                  # View your scoring profile
tm profile reset            # Re-run the wizard

# Scoring
tm score <idea>             # Score without saving
tm dump <idea>              # Score and save to database
tm dump --quick             # Fast capture, minimal output
tm dump --use-ai            # Use LLM for deeper analysis

# Review
tm review                   # Browse saved ideas
tm review --min-score 7.0   # Filter by score

# Management
tm prune                    # Clean up low-scoring ideas
tm link create <a> <b> <type>  # Link related ideas
tm bulk analyze             # Re-score multiple ideas

# Analysis
tm analytics trends         # Score trends over time
tm analytics anomaly        # Detect unusual patterns
```

## Two Scoring Modes

### Universal Mode (Default)

Uses your profile from `~/.brain-salad/profile.yaml`. Created via `tm init`.

- Works for anyone (potters, developers, writers, anyone)
- 6 universal dimensions
- Weights based on your wizard answers

### Legacy Mode (Power Users)

Uses `~/.telos/telos.md` for detailed goal configuration.

- AI/tech-focused dimensions
- Manual weight configuration
- More granular control

Brain-Salad automatically detects which mode to use based on which config exists.

## LLM Integration (Optional)

For deeper analysis, Brain-Salad supports multiple LLM providers:

```bash
# Ollama (local, private)
ollama serve
tm dump "idea" --use-ai --provider ollama

# OpenAI
export OPENAI_API_KEY=sk-...
tm dump "idea" --use-ai --provider openai

# Claude
export CLAUDE_API_KEY=sk-ant-...
tm dump "idea" --use-ai --provider claude
```

LLM is optional — the rule-based scoring works great without it.

## Development

```bash
# Run all CI checks before pushing
make check

# Individual commands
make test      # Run tests
make lint      # Run linters
make build     # Build binaries
make fmt       # Format code
```

See [CLAUDE.md](./CLAUDE.md) for development guidelines and common linter fixes.

## Project Structure

```
brain-salad/
├── cmd/cli/              # CLI entry point
├── internal/
│   ├── profile/          # User preference system
│   ├── scoring/          # Scoring engines (universal + legacy)
│   ├── cli/wizard/       # Interactive setup wizard
│   ├── cli/              # Command implementations
│   ├── database/         # SQLite repository
│   ├── llm/              # LLM provider abstraction
│   └── ...
├── test/                 # Integration tests
└── deployments/docker/   # Docker configuration
```

## Documentation

- [CLI Reference](./docs/CLI_REFERENCE.md) - Complete command documentation
- [Configuration](./docs/CONFIGURATION.md) - Setup and config options
- [Architecture](./docs/ARCHITECTURE.md) - System design
- [API Documentation](./docs/API.md) - REST API reference

## Contributing

```bash
# Fork and clone
git checkout -b feature/my-feature

# Make changes
make check  # Must pass before PR

# Commit with conventional commits
git commit -m "feat: add new feature"
git push origin feature/my-feature
```

See [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](./LICENSE)

## Support

- [GitHub Issues](https://github.com/ryacub/brain-salad/issues)
- [GitHub Discussions](https://github.com/ryacub/brain-salad/discussions)
