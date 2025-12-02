# Brain-Salad - CLI Reference

Complete command-line interface reference for the Brain-Salad tool.

## Table of Contents

- [Installation](#installation)
- [Global Flags](#global-flags)
- [Commands](#commands)
  - [analyze](#analyze)
  - [dump](#dump)
  - [list](#list)
  - [link](#link)
  - [review](#review)
  - [server](#server)
  - [version](#version)
- [Configuration](#configuration)
- [Examples](#examples)

## Installation

### From Source

```bash
cd go
make build-cli
sudo mv bin/tm /usr/local/bin/
```

### Using Docker

```bash
docker-compose run telos-cli --help
```

## Global Flags

These flags are available for all commands:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--telos` | `-t` | string | `./telos.md` | Path to telos configuration file |
| `--db` | `-d` | string | `~/.telos/ideas.db` | Path to database file |
| `--help` | `-h` | - | - | Show help for command |
| `--verbose` | `-v` | - | - | Enable verbose output |

### Example

```bash
tm --telos ~/my-telos.md --db ~/data/ideas.db analyze "My idea"
```

## Commands

### analyze

Analyze an idea against your telos without storing it.

#### Usage

```bash
tm analyze [flags] <idea>
```

#### Arguments

- `<idea>` (required): The idea text to analyze

#### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | false | Output results in JSON format |
| `--save` | bool | false | Save the idea to database after analysis |

#### Examples

```bash
# Basic analysis
tm analyze "Build a microservices platform"

# Analyze with JSON output
tm analyze --json "Implement GraphQL API"

# Analyze and save to database
tm analyze --save "Create mobile app"

# Read from file
tm analyze "$(cat idea.txt)"

# Pipe from stdin
echo "Refactor authentication module" | tm analyze -
```

#### Output

```
Idea Analysis
=============
Raw Score: 8.5/10
Final Score: 7.5/10

Goal Alignment:
  ✓ Technical Excellence: 10.0 (matched: testing, quality, scalable)
  ✓ Innovation: 5.0 (matched: innovative)
  ○ User Experience: 0.0

Detected Patterns:
  ⚠ Technical Debt (penalty: -1.0)
    Description: Quick fixes without proper design

Recommendation: PROCEED_WITH_CAUTION
Address the technical debt concerns before implementation.
```

---

### dump

Import ideas from various sources into the database.

#### Usage

```bash
tm dump [flags] <source>
```

#### Arguments

- `<source>` (required): Path to file or directory containing ideas

#### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | string | auto | Input format: `text`, `json`, `csv`, `markdown` |
| `--recursive` | bool | false | Recursively process directories |
| `--analyze` | bool | true | Analyze ideas during import |
| `--skip-duplicates` | bool | true | Skip duplicate ideas |
| `--batch-size` | int | 100 | Number of ideas to process in one batch |

#### Supported Formats

**Text File** (one idea per line):
```text
Build a testing framework
Implement CI/CD pipeline
Create documentation system
```

**JSON File**:
```json
[
  {"content": "First idea"},
  {"content": "Second idea", "status": "active"}
]
```

**CSV File**:
```csv
content,status
"Implement authentication",active
"Add logging",active
```

**Markdown File**:
```markdown
# Ideas

- Build a testing framework
- Implement CI/CD pipeline
- Create documentation system
```

#### Examples

```bash
# Import from text file
tm dump ideas.txt

# Import from JSON
tm dump --format json ideas.json

# Import directory recursively
tm dump --recursive ./idea-notes/

# Import without analysis (faster)
tm dump --analyze=false large-list.txt

# Import with custom batch size
tm dump --batch-size 50 ideas.txt
```

#### Output

```
Importing ideas from: ideas.txt
Format: text (auto-detected)

Processing: ████████████████████ 100% (250/250)

Summary:
  ✓ Imported: 245 ideas
  ⊘ Skipped (duplicates): 5 ideas
  ✗ Failed: 0 ideas

  Average Score: 7.2/10
  Processing Time: 2.3s
  Rate: 108 ideas/sec
```

---

### list

List and filter ideas from the database.

#### Usage

```bash
tm list [flags]
```

#### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--status` | string | all | Filter by status: `active`, `archived`, `deleted`, `all` |
| `--min-score` | float | 0.0 | Minimum final score |
| `--max-score` | float | 10.0 | Maximum final score |
| `--limit` | int | 100 | Maximum number of results |
| `--offset` | int | 0 | Number of results to skip |
| `--sort` | string | created | Sort by: `created`, `score`, `updated` |
| `--order` | string | desc | Sort order: `asc`, `desc` |
| `--json` | bool | false | Output in JSON format |
| `--format` | string | table | Output format: `table`, `json`, `csv` |

#### Examples

```bash
# List all active ideas
tm list --status active

# List high-scoring ideas
tm list --min-score 8.0

# List recently created ideas
tm list --sort created --limit 10

# Export to CSV
tm list --format csv > ideas.csv

# Complex filter
tm list --status active --min-score 7.0 --max-score 9.0 --limit 50
```

#### Output Formats

**Table** (default):
```
ID                                     Content                          Score  Status   Created
550e8400-e29b-41d4-a716-446655440000  Build testing framework          8.5    active   2025-11-19
7c9e6679-7425-40de-944b-e07fc1f90ae7  Implement CI/CD                  7.2    active   2025-11-18
...
```

**JSON**:
```json
{
  "ideas": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "content": "Build testing framework",
      "final_score": 8.5,
      "status": "active",
      "created_at": "2025-11-19T10:30:00Z"
    }
  ],
  "total": 42,
  "limit": 100,
  "offset": 0
}
```

**CSV**:
```csv
id,content,raw_score,final_score,status,created_at
550e8400-e29b-41d4-a716-446655440000,"Build testing framework",9.0,8.5,active,2025-11-19T10:30:00Z
```

---

### link

Manage relationships between ideas.

For complete documentation, see:
- **[Link Command User Guide](user-guide/link-command.md)** - Comprehensive guide
- **[Getting Started Tutorial](tutorials/getting-started-with-links.md)** - Beginner tutorial
- **[Cheat Sheet](quick-reference/link-cheatsheet.md)** - Quick reference
- **[FAQ](faq/link-command-faq.md)** - Common questions

#### Subcommands

**Create a relationship:**
```bash
tm link create <source-id> <target-id> <type> [--no-confirm]
```

**List relationships:**
```bash
tm link list <idea-id>
```

**Show related ideas:**
```bash
tm link show <idea-id> [--type <type>]
```

**Remove a relationship:**
```bash
tm link remove <relationship-id> [--no-confirm]
```

**Find paths:**
```bash
tm link path <from-id> <to-id> [--max-depth N]
```

#### Relationship Types

- `depends_on` - Source depends on target
- `blocked_by` - Source is blocked by target
- `blocks` - Source blocks target
- `part_of` - Source is part of target
- `parent` - Source is parent of target
- `child` - Source is child of target
- `related_to` - Ideas are related (symmetric)
- `similar_to` - Ideas are similar (symmetric)
- `duplicate` - Ideas are duplicates (symmetric)

#### Examples

```bash
# Create a dependency
tm link create api-123 db-456 depends_on

# View all relationships for an idea
tm link list api-123

# Show related ideas with details
tm link show api-123 --type depends_on

# Find path between two ideas
tm link path ui-789 db-456

# Remove a relationship
tm link remove rel-xyz

# Skip confirmation
tm link create idea1 idea2 duplicate --no-confirm
```

#### Quick Start

```bash
# Project breakdown
tm link create task project part_of

# Track dependencies
tm link create taskA taskB depends_on

# Mark duplicates
tm link create idea1 idea2 duplicate

# Track blockers
tm link create task blocker blocked_by
```

See the [Link Command User Guide](user-guide/link-command.md) for detailed workflows, best practices, and troubleshooting.

---

### review

Review and update the status of ideas.

#### Usage

```bash
tm review [flags] <id>
```

#### Arguments

- `<id>` (required): The idea ID to review

#### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--status` | string | - | Update status: `active`, `archived`, `deleted` |
| `--interactive` | bool | true | Enable interactive review mode |

#### Examples

```bash
# Review idea interactively
tm review 550e8400-e29b-41d4-a716-446655440000

# Archive an idea
tm review --status archived 550e8400-e29b-41d4-a716-446655440000

# Non-interactive status update
tm review --interactive=false --status deleted <id>
```

#### Interactive Mode

```
Reviewing Idea: 550e8400-e29b-41d4-a716-446655440000
=================================================

Content: Build a comprehensive testing framework with coverage tracking

Analysis:
  Raw Score: 9.0/10
  Final Score: 8.5/10

Goal Alignment:
  ✓ Technical Excellence: 10.0
  ✓ Quality Assurance: 8.0

Detected Patterns:
  ⚠ Scope Creep (penalty: -0.5)

Actions:
  [a] Archive
  [d] Delete
  [k] Keep active
  [e] Edit content
  [r] Re-analyze
  [q] Quit

Choice:
```

---

### server

Start the web API server.

#### Usage

```bash
tm server [flags]
```

#### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--port` | int | 8080 | Port to listen on |
| `--host` | string | 0.0.0.0 | Host address to bind to |
| `--cors` | string | * | CORS allowed origins (comma-separated) |
| `--cache-ttl` | duration | 5m | Cache TTL duration |
| `--rate-limit` | int | 100 | Rate limit (requests per minute) |

#### Examples

```bash
# Start server on default port
tm server

# Start on custom port
tm server --port 3000

# Start with specific CORS origins
tm server --cors "http://localhost:3000,http://localhost:5173"

# Start with custom cache TTL
tm server --cache-ttl 10m

# Start with higher rate limit
tm server --rate-limit 500
```

#### Output

```
Starting Brain-Salad API Server
======================================

Configuration:
  Port: 8080
  Host: 0.0.0.0
  Database: /home/user/.telos/ideas.db
  Telos: ./telos.md
  Cache TTL: 5m0s
  Rate Limit: 100 req/min

Server ready at http://0.0.0.0:8080
API endpoints:
  GET  /health
  GET  /api/v1/csrf-token
  POST /api/v1/analyze
  GET  /api/v1/ideas
  POST /api/v1/ideas
  GET  /api/v1/ideas/:id
  PUT  /api/v1/ideas/:id
  DELETE /api/v1/ideas/:id
  GET  /api/v1/analytics/stats

Press Ctrl+C to stop
```

---

### version

Display version information.

#### Usage

```bash
tm version [flags]
```

#### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | false | Output in JSON format |

#### Examples

```bash
# Display version
tm version

# JSON output
tm version --json
```

#### Output

```
Brain-Salad
Version: 1.0.0
Build: 2025-11-19T10:30:00Z
Go Version: go1.25.4
Platform: linux/amd64
```

---

## Configuration

### Telos File

The telos file defines your goals, strategies, and anti-patterns. See [Configuration Guide](CONFIGURATION.md) for details.

**Example telos.md**:

```markdown
# My Telos

## Core Goals

### Technical Excellence (10)
Build high-quality, maintainable software.

**Keywords**: quality, testing, code review, maintainable, clean code

### Innovation (8)
Drive innovation and creative solutions.

**Keywords**: innovation, creative, breakthrough, novel

## Strategies

### Test-Driven Development
Write tests before implementation.

**Keywords**: TDD, tests, testing, test coverage

### Continuous Integration
Automate building and testing.

**Keywords**: CI/CD, automation, pipeline

## Anti-Patterns

### Technical Debt (-2)
Accumulating shortcuts and quick fixes.

**Keywords**: hack, quick fix, TODO, FIXME, workaround

### Premature Optimization (-1)
Optimizing before understanding requirements.

**Keywords**: premature, over-engineering, complexity
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `TELOS_FILE` | `./telos.md` | Path to telos configuration |
| `TELOS_DB` | `~/.telos/ideas.db` | Path to database file |
| `TELOS_PORT` | `8080` | API server port |
| `TELOS_LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |

### Database Location

Default database location: `~/.telos/ideas.db`

The database directory will be created automatically if it doesn't exist.

---

## Examples

### Daily Workflow

```bash
# Morning: Review new ideas from notes
tm dump ~/notes/ideas.txt

# Check high-priority ideas
tm list --min-score 8.0 --status active

# Analyze a quick idea
tm analyze "Add metrics dashboard to monitoring system"

# Review ideas interactively
tm review $(tm list --json | jq -r '.ideas[0].id')
```

### Integration with Other Tools

**Git commit hook** (analyze commit messages):
```bash
#!/bin/bash
# .git/hooks/commit-msg

commit_msg=$(cat "$1")
score=$(tm analyze --json "$commit_msg" | jq -r '.analysis.final_score')

if (( $(echo "$score < 5.0" | bc -l) )); then
    echo "Warning: Low-quality commit message (score: $score)"
    exit 1
fi
```

**Vim integration** (analyze selected text):
```vim
" Add to .vimrc
vnoremap <leader>ta y:!tm analyze "<C-R>""<CR>
```

**Shell alias**:
```bash
# Add to .bashrc or .zshrc
alias idea='tm analyze --save'
alias ideas='tm list --status active'
```

### API Integration

```bash
# Start server in background
tm server &

# Use API with curl
curl -X POST http://localhost:8080/api/v1/analyze \
  -H "Content-Type: application/json" \
  -d '{"content": "Build a testing framework"}'

# Get statistics
curl http://localhost:8080/api/v1/analytics/stats
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments |
| 3 | File not found |
| 4 | Database error |
| 5 | Configuration error |

---

## Troubleshooting

### Database locked

```bash
# If you get "database is locked" errors
# Make sure no other tm processes are running
killall tm

# Or use a different database
tm --db /tmp/ideas.db list
```

### Telos file not found

```bash
# Specify the telos file path
tm --telos ~/my-telos.md analyze "My idea"

# Or set environment variable
export TELOS_FILE=~/my-telos.md
tm analyze "My idea"
```

### Permission denied

```bash
# Make binary executable
chmod +x /usr/local/bin/tm

# Or run with explicit path
./bin/tm --help
```

---

## See Also

- [User Guide](../README.md)
- [API Documentation](api/openapi.yaml)
- [Configuration Guide](CONFIGURATION.md)
- [Development Guide](DEVELOPMENT.md)
