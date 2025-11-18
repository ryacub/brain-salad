# API Reference

## Command Line Interface

### Main Commands

#### `tm dump`
Capture and analyze an idea against your personal goals.

**Usage:**
```
tm dump [OPTIONS] [IDEA]
```

**Arguments:**
- `IDEA`: The idea text to analyze (omit for interactive input)

**Options:**
- `-i, --interactive`: Open editor for multi-line input
- `-q, --quick`: Save without analysis
- `--claude`: Force use of Claude CLI instead of Ollama
- `--no-ai`: Disable AI analysis, use rule-based only

**Examples:**
```bash
tm dump "Build a Rust tool for productivity"
tm dump --interactive  # Opens editor
tm dump --quick "Note for later"
```

#### `tm analyze`
Analyze an idea (from text or last captured).

**Usage:**
```
tm analyze [OPTIONS] [IDEA]
```

**Options:**
- `--last`: Analyze the most recently captured idea
- `--no-ai`: Disable AI analysis, use rule-based only

**Examples:**
```bash
tm analyze "Improve the scoring algorithm"
tm analyze --last
```

#### `tm analyze-llm`
Analyze an idea using LLM with complete prompt template.

**Usage:**
```
tm analyze-llm [OPTIONS] <IDEA>
```

**Options:**
- `--provider <PROVIDER>`: LLM provider (ollama, openai, claude) [default: ollama]
- `--model <MODEL>`: Model name to use [default: mistral]
- `--api-key <API_KEY>`: API key for the LLM provider
- `--base-url <BASE_URL>`: Base URL for custom LLM API
- `--save`: Save the analysis result to the database
- `--temperature <TEMPERATURE>`: Temperature for the LLM (0.0 to 1.0) [default: 0.3]
- `--max-tokens <MAX_TOKENS>`: Maximum tokens for the response [default: 4096]
- `--no-ai`: Disable AI analysis, use rule-based only

#### `tm score`
Quick score an idea without saving.

**Usage:**
```
tm score <IDEA>
```

**Example:**
```bash
tm score "Learn a new programming language"
```

#### `tm review`
Review and browse captured ideas.

**Usage:**
```
tm review [OPTIONS]
```

**Options:**
- `-l, --limit <LIMIT>`: Limit number of ideas to show [default: 10]
- `--min-score <MIN_SCORE>`: Filter by minimum score [default: 0.0]
- `--pruning`: Show ideas needing pruning review

**Examples:**
```bash
tm review --limit 20
tm review --min-score 7.0
tm review --pruning
```

#### `tm prune`
Manage old ideas (archive/delete).

**Usage:**
```
tm prune [OPTIONS]
```

**Options:**
- `--auto`: Auto-prune without confirmation
- `--dry-run`: Show what would be pruned (dry run)

**Examples:**
```bash
tm prune  # Interactive pruning
tm prune --auto  # Auto prune
tm prune --dry-run  # See what would be pruned
```

#### `tm health`
Health check and monitoring.

**Usage:**
```
tm health [OPTIONS]
```

**Options:**
- `--format <FORMAT>`: Output format (json or text) [default: text]

**Examples:**
```bash
tm health
tm health --format json
```

#### `tm llm`
Manage Ollama LLM service.

**Usage:**
```
tm llm <ACTION>
```

**Actions:**
- `status`: Show Ollama status and available models
- `start`: Start Ollama service
- `stop`: Stop Ollama service

**Examples:**
```bash
tm llm status
tm llm start
tm llm stop
```

### Bulk Operations

#### `tm bulk`
Perform bulk operations on multiple ideas.

**Subcommands:**
- `tag`: Add tags to ideas based on criteria
- `archive`: Archive ideas based on criteria
- `delete`: Delete ideas based on criteria
- `export`: Export ideas to various formats

**Examples:**
```bash
tm bulk tag "important" --min-score 7.0 --limit 10
tm bulk archive --older-than 30 --max-score 4.0
```

### Analytics Commands

#### `tm analytics`
Analytics and reporting.

**Subcommands:**
- `trends`: Show idea trends over time
- `performance`: Show scoring performance metrics
- `patterns`: Show detected patterns in ideas

**Examples:**
```bash
tm analytics trends
tm analytics performance
tm analytics patterns
```

### Link Commands

#### `tm link`
Link and manage idea relationships.

**Subcommands:**
- `add <SOURCE_ID> <TARGET_ID> --type <TYPE>`: Link two ideas
- `show <IDEA_ID>`: Show relationships for an idea
- `remove <SOURCE_ID> <TARGET_ID>`: Remove a link

**Link Types:**
- `depends-on`: One idea depends on another
- `related-to`: Ideas are related
- `blocks`: One idea blocks another
- `part-of`: One idea is part of a larger idea

**Examples:**
```bash
tm link add idea1_id idea2_id --type "depends-on"
tm link show idea1_id
tm link remove idea1_id idea2_id
```

## Configuration Environment Variables

- `TELOS_FILE`: Path to your telos.md file
- `TELOS_LOG_JSON`: Set to any value to enable JSON logging
- `TELOS_LOG_DIR`: Directory for log files
- `RUST_LOG`: Standard Rust logging level (info, debug, warn, error)

## Exit Codes

- `0`: Success
- `1`: General error
- `2`: Command line argument error
- `3`: Configuration error
- `4`: Database error
- `5`: Scoring error
- `6`: AI service error

## Global Options

All commands support these global options:
- `--no-ai`: Disable AI analysis, use rule-based only
- `--help`: Show help information
- `--version`: Show version information