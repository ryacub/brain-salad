# Brain-Salad - CLI Reference

Complete command-line interface reference for the Brain-Salad tool.

## Table of Contents

- [Installation](#installation)
- [Global Flags](#global-flags)
- [Commands](#commands)
  - [add](#add)
  - [init](#init)
  - [list](#list)
  - [show](#show)
  - [link](#link)
  - [bulk](#bulk)
  - [analytics](#analytics)
  - [profile](#profile)
  - [prune](#prune)
  - [llm](#llm)
  - [completion](#completion)
- [Examples](#examples)

## Installation

### Using Go Install

```bash
go install github.com/ryacub/telos-idea-matrix/cmd/cli@latest
```

### From Source

```bash
git clone https://github.com/ryacub/brain-salad.git
cd brain-salad
make build
sudo mv bin/tm /usr/local/bin/
```

### Quick Install Script

```bash
curl -sSL https://raw.githubusercontent.com/ryacub/brain-salad/main/scripts/install.sh | bash
```

## Global Flags

These flags are available for all commands:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--telos` | `-t` | string | `./telos.md` | Path to telos configuration file |
| `--db` | `-d` | string | `~/.telos/ideas.db` | Path to database file |
| `--help` | `-h` | - | - | Show help for command |

## Commands

### add

Add and score an idea, saving it to the database.

#### Usage
```bash
tm add <idea> [flags]
```

#### Flags
| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--ai` | | - | - | Use AI for deeper analysis |
| `--dry-run` | `-n` | - | - | Score without saving |
| `--json` | | - | - | Output as JSON |
| `--provider` | `-p` | string | - | AI provider (ollama|openai|claude) |
| `--quiet` | `-q` | - | - | Minimal output |
| `--from-clipboard` | | - | - | Read idea from clipboard |
| `--to-clipboard` | | - | - | Copy result to clipboard |

#### Examples
```bash
tm add "Build a mobile app for tracking inventory"
tm add "Start a podcast" --ai
tm add "Quick idea" --quiet
tm add "Test idea" --dry-run
```

### init

Initialize Brain Salad for first-time use with an interactive wizard.

#### Usage
```bash
tm init [flags]
```

#### Flags
| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--advanced` | | - | - | Create telos.md for advanced configuration |

#### Examples
```bash
tm init                    # Interactive wizard
tm init --advanced        # Create telos.md
```

### list

List and filter your saved ideas.

#### Usage
```bash
tm list [flags]
```

#### Flags
| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--limit` | `-l` | int | 10 | Max ideas to show |
| `--min-score` | | float | - | Minimum score |
| `--max-score` | | float | - | Maximum score |
| `--status` | | string | active | Status (active|archived|deleted) |
| `--json` | | - | - | Output as JSON |
| `--quiet` | `-q` | - | - | Compact output |

#### Examples
```bash
tm list                                    # List recent ideas
tm list --min-score 7.0                   # High-scoring ideas only
tm list --status archived                  # Archived ideas
tm list --limit 20                         # Show more ideas
tm list --json                              # JSON output
```

### show

Show detailed information about a specific idea.

#### Usage
```bash
tm show <id> [flags]
```

#### Flags
| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--json` | | - | - | Output as JSON |

#### Examples
```bash
tm show abc123-def456                    # Show idea details
tm show abc123-def456 --json              # JSON output
```

### link

Manage relationships between related ideas.

#### Usage
```bash
tm link <command> [arguments]
```

#### Subcommands
- `create <idea1> <idea2> <type>` - Create a link between ideas
- `list` - List all links
- `delete <id>` - Delete a link

#### Examples
```bash
tm link create abc123 def456 "similar"     # Link related ideas
tm link list                              # Show all links
```

### bulk

Bulk operations on multiple ideas.

#### Usage
```bash
tm bulk <command> [arguments]
```

#### Subcommands
- `analyze` - Re-score multiple ideas
- `export` - Export ideas to file
- `import` - Import ideas from file
- `delete` - Delete multiple ideas
- `archive` - Archive multiple ideas
- `tag` - Add tags to ideas

### analytics

View statistics and trends about your ideas.

#### Usage
```bash
tm analytics <command> [arguments]
```

#### Subcommands
- `trends` - Score trends over time
- `anomaly` - Detect unusual patterns
- `stats` - General statistics

### profile

View your scoring profile.

#### Usage
```bash
tm profile [flags]
```

#### Flags
| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--reset` | | - | - | Re-run the wizard |

#### Examples
```bash
tm profile                                 # View profile
tm profile --reset                         # Re-run wizard
```

### prune

Clean up old or low-scoring ideas.

#### Usage
```bash
tm prune [flags]
```

#### Flags
| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--min-score` | | float | 3.0 | Minimum score threshold |
| `--older-than` | | duration | - | Only ideas older than duration |
| `--dry-run` | | - | - | Show what would be pruned |
| `--force` | | - | - | Skip confirmation

### llm

Manage LLM providers for AI analysis.

#### Usage
```bash
tm llm <command> [arguments]
```

#### Subcommands
- `list` - List configured providers
- `test` - Test provider connectivity
- `config` - Configure provider settings
- `set-default <provider>` - Set default provider

### completion

Generate shell completion scripts.

#### Usage
```bash
tm completion <shell>
```

#### Shells
- `bash`
- `zsh`
- `fish`
- `powershell`

## Examples

### Basic Workflow
```bash
# Initialize
tm init

# Add ideas
tm add "Build a mobile app for inventory tracking"
tm add "Start a podcast about software development" --ai

# List ideas
tm list

# Show details
tm show abc123-def456
```

### Advanced Workflow
```bash
# High-scoring ideas only
tm list --min-score 8.0

# Archived ideas
tm list --status archived

# JSON output for scripting
tm list --json | jq '.[] | select(.score > 7)'

# Clean up low-scoring ideas
tm prune --min-score 3.0 --dry-run
```

### AI Analysis Workflow
```bash
# Configure OpenAI
export OPENAI_API_KEY=sk-...
tm llm config openai

# Use AI for deeper analysis
tm add "Complex AI project" --ai --provider openai

# Test provider
tm llm test openai
```

### Bulk Operations
```bash
# Re-score all ideas
tm bulk analyze

# Export to CSV
tm bulk export ideas.csv

# Import from file
tm bulk import ideas.json
```

## LLM Integration

Brain-Salad supports multiple LLM providers for deeper analysis:

### OpenAI
```bash
export OPENAI_API_KEY=sk-...
tm add "idea" --ai --provider openai
```

### Claude (Anthropic)
```bash
export ANTHROPIC_API_KEY=sk-ant-...
tm add "idea" --ai --provider claude
```

### Ollama (Local)
```bash
ollama serve
tm add "idea" --ai --provider ollama
```

## Troubleshooting

### Common Issues

**"Command not found"**
- Verify installation: `which tm`
- Check PATH: `echo $PATH`
- Reinstall: `go install github.com/ryacub/telos-idea-matrix/cmd/cli@latest`

**Database errors**
- Check permissions: `ls -la ~/.telos/`
- Verify telos file: `cat ~/.telos/telos.md`
- Reset profile: `tm profile --reset`

**AI analysis not working**
- Check API keys: `tm llm list`
- Test provider: `tm llm test <provider>`
- Verify network: `curl -I https://api.openai.com/v1/models`

### Debug Mode
```bash
# Enable verbose logging
tm add "test" --verbose

# Check configuration
tm llm config
tm profile
```