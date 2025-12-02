# Architecture Documentation

## Overview

Brain-Salad is a **Go application** providing both CLI and web API interfaces for capturing ideas and evaluating them against personalized goals and strategies. The system is designed to be modular, extensible, and privacy-focused with local-first architecture.

## System Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    CLI Interface (Cobra) / Web API (Chi)                │
│                         ┌──────────────────┐                            │
│                    ┌────┤  dump  │ review │                            │
│                    │    │ prune  │  link  │                            │
│                    │    │ score  │analyze │                            │
│                    │    └────────┬────────┘                            │
├────────────────────┼─────────────┼──────────────────────────────────────┤
│  LAYER 1:          │             │                                      │
│  Request           │             ▼                                      │
│  Processing        │    ┌─────────────────┐                            │
│                    │    │ Command Handler │ (Process user input)       │
│                    │    │ ┌─────────────┐ │                            │
│                    │    │ │  Validation │ │ (Sanitize, check bounds)  │
│                    │    │ └──────┬──────┘ │                            │
│                    │    └────────┼────────┘                            │
│                    │             │                                      │
├────────────────────┼─────────────┼──────────────────────────────────────┤
│  LAYER 2:          │             ▼                                      │
│  Business          │    ┌─────────────────────────────────────┐        │
│  Logic             │    │  Scoring Strategy (Pluggable)      │        │
│                    │    │ ┌─────────────────────────────────┐ │        │
│                    │    │ │ Interface: Provider             │ │        │
│                    │    │ │ ┌──────────────────────────────┐ │ │        │
│                    │    │ │ │ scoring.Engine.Score()       │ │ │        │
│                    │    │ │ │ - Mission alignment (40%)    │ │ │        │
│                    │    │ │ │ - Anti-challenge (35%)       │ │ │        │
│                    │    │ │ │ - Strategic fit (25%)        │ │ │        │
│                    │    │ │ └──────────────────────────────┘ │ │        │
│                    │    │ └─────────────────────────────────┘ │        │
│                    │    │                                      │        │
│                    │    │ ┌─────────────────────────────────┐ │        │
│                    │    │ │ patterns.Detector               │ │        │
│                    │    │ │ - Context switching detection   │ │        │
│                    │    │ │ - Anti-pattern identification   │ │        │
│                    │    │ └─────────────────────────────────┘ │        │
│                    │    └────────────────┬────────────────────┘        │
│                    │                     │                              │
├────────────────────┼──────────────┬──────┼──────────────────────────────┤
│  LAYER 3:          │              │      ▼                              │
│  Integration       │              │   ┌──────────────────┐             │
│                    │              │   │ LLM Integration  │             │
│                    │              │   │ (Optional)       │             │
│                    │              │   │ ┌──────────────┐ │             │
│                    │              │   │ │ Provider     │ │             │
│                    │              │   │ │ Fallback     │ │             │
│                    │              │   │ │ Chain        │ │             │
│                    │              │   │ └──────┬───────┘ │             │
│                    │              │   │        │ (fail) ▼              │
│                    │              │   │    Rule-based                  │
│                    │              │   └──────────────────┘             │
│                    │              │                                    │
│                    │    ┌─────────┴────────────────────┐              │
│                    │    ▼                              ▼              │
│                    │   ┌───────────────────────┐ ┌──────────┐        │
│                    │   │ Configuration Module  │ │ Telos    │        │
│                    │   │ ┌─────────────────────┤ │ Parser   │        │
│                    │   │ │ Config Paths        │ │          │        │
│                    │   │ │ - env vars          │ │ Extracts │        │
│                    │   │ │ - ~/.telos/         │ │ - Goals  │        │
│                    │   │ │ - ./telos.md        │ │ - Strats │        │
│                    │   │ │ - Custom paths      │ │ - Patterns│        │
│                    │   │ └─────────────────────┤ │          │        │
│                    │   └──────┬────────────────┘ └──────────┘        │
│                    │          │                    │                  │
├────────────────────┼──────────┼────────────────────┼──────────────────┤
│  LAYER 4:          │          │                    │                  │
│  Persistence       │          ▼                    │                  │
│                    │   ┌──────────────────────┐   │                  │
│                    │   │ Database Layer       │   │                  │
│                    │   │ ┌──────────────────┐ │   │                  │
│                    │   │ │ go-sqlite3       │ │   │                  │
│                    │   │ │ ┌──────────────┐ │ │   │                  │
│                    │   │ │ │ SQLite DB    │ │ │   │                  │
│                    │   │ │ │ WAL Mode     │ │ │   │                  │
│                    │   │ │ │ - Ideas      │ │ │   │                  │
│                    │   │ │ │ - Links      │ │ │   │                  │
│                    │   │ │ │ - Tags       │ │ │   │                  │
│                    │   │ │ │ - Analysis   │ │ │   │                  │
│                    │   │ │ └──────────────┘ │ │   │                  │
│                    │   │ └──────────────────┘ │   │                  │
│                    │   └──────────────────────┘   │                  │
│                    │                               │                  │
│                    │    (Reads telos.md from)     │                  │
│                    │    ◄────────────────────────┘                  │
│                    │                                                │
└────────────────────┴────────────────────────────────────────────────┘
```

## Module Structure

### Project Organization

```
go/
├── cmd/                    # Application entry points
│   ├── cli/               # CLI application (main.go)
│   ├── web/               # Web server (main.go)
│   └── verify-wal/        # Utility tools
├── internal/              # Private application code
│   ├── analytics/         # Trends, reports, visualizations
│   ├── api/              # HTTP server (chi router)
│   ├── cli/              # CLI commands (cobra framework)
│   ├── config/           # Configuration management
│   ├── database/         # Repository pattern + migrations
│   ├── export/           # CSV/JSON export
│   ├── health/           # Health checks & monitoring
│   ├── llm/              # LLM integration system
│   ├── logging/          # Structured logging (zerolog)
│   ├── metrics/          # Application metrics
│   ├── models/           # Domain models
│   ├── patterns/         # Anti-pattern detection
│   ├── scoring/          # Rule-based scoring engine
│   ├── tasks/            # Background task scheduler
│   ├── telos/            # Telos.md parser
│   └── utils/            # Utilities (clipboard, etc.)
└── pkg/                   # Public library code
    └── client/           # External client library
```

### Core Modules

#### Entry Points
- **`cmd/cli/main.go`**: CLI entry point, delegates to `internal/cli.Execute()`
- **`cmd/web/main.go`**: Web server with graceful shutdown, health checks, and middleware

#### CLI Layer
- **`internal/cli/`**: Cobra-based CLI implementation
  - `root.go`: Root command setup with global flags (`--db`, `--telos`)
  - Command files: `score.go`, `analyze.go`, `review.go`, `dump.go`, `prune.go`, etc.
  - Shared `CLIContext` for dependency injection
  - Color-coded output using `fatih/color`

#### API Layer
- **`internal/api/`**: Chi-based HTTP server
  - `server.go`: Server setup with middleware stack
  - `handlers.go`: RESTful endpoint handlers
  - `middleware.go`: Custom middleware (auth, logging, rate limiting)
  - `csrf.go`: CSRF protection
  - CORS configuration and security headers

#### Domain Models
- **`internal/models/`**: Core domain entities
  - `idea.go`: Idea entity with validation
  - `telos.go`: Telos configuration (Problems, Missions, Goals, Challenges, Strategies, Stack, Failure Patterns)
  - `analysis.go`: Complete scoring breakdown (Mission, Anti-Challenge, Strategic)
  - `relationship.go`: Idea relationships with graph support

#### Data Layer
- **`internal/database/`**: Repository pattern implementation
  - `repository.go`: Single repository with all data operations
  - `migrations/`: Embedded SQL migrations (001_initial.sql, 002_relationships.sql, 003_add_tags.sql)
  - WAL mode enabled with connection pooling
  - Graph operations (pathfinding, relationship traversal)

#### Business Logic
- **`internal/scoring/`**: Rule-based scoring engine
  - `engine.go`: Complete scoring algorithm (0-10 scale)
  - Mission Alignment (4.0 max), Anti-Challenge (3.5 max), Strategic Fit (2.5 max)
  - Pre-compiled regex patterns for performance
  - Keyword-based matching with stack compatibility

- **`internal/patterns/`**: Anti-pattern detection
  - `detector.go`: Pattern detection with confidence scores
  - Detects context switching, perfectionism, procrastination, accountability avoidance

- **`internal/telos/`**: Telos.md parsing
  - `parser.go`: Markdown parser with section-based extraction
  - Keyword extraction for failure patterns

#### LLM Integration
- **`internal/llm/`**: Comprehensive LLM system
  - `manager.go`: Provider manager with fallback chain and health checks
  - `provider.go`: Provider interface and rule-based implementation
  - `client/`: HTTP clients for Ollama, Claude, OpenAI, custom endpoints
  - `cache/`: Response caching with similarity-based matching
  - `processing/`: Response validation and parsing
  - `quality/`: Quality metrics tracking

#### Supporting Systems
- **`internal/config/`**: Environment-based configuration
- **`internal/analytics/`**: Trend analysis and reporting
- **`internal/health/`**: Health check orchestration
- **`internal/metrics/`**: In-memory metrics collection (LLM requests, tokens, errors, fallbacks)
- **`internal/logging/`**: Structured logging with zerolog and log rotation
- **`internal/tasks/`**: Background task scheduler with graceful shutdown
- **`internal/export/`**: CSV and JSON exporters
- **`internal/utils/`**: Clipboard and other utilities

### Database Schema

#### Ideas Table
```sql
CREATE TABLE ideas (
    id TEXT PRIMARY KEY,
    content TEXT NOT NULL,
    raw_score REAL,
    final_score REAL,
    patterns TEXT,              -- JSON array of strings
    tags TEXT,                  -- JSON array of strings
    recommendation TEXT,
    analysis_details TEXT,      -- JSON Analysis object
    created_at TEXT NOT NULL,   -- RFC3339 timestamp
    reviewed_at TEXT,
    status TEXT NOT NULL DEFAULT 'active'
);
```

#### Relationships Table
```sql
CREATE TABLE idea_relationships (
    id TEXT PRIMARY KEY,
    source_idea_id TEXT NOT NULL,
    target_idea_id TEXT NOT NULL,
    relationship_type TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (source_idea_id) REFERENCES ideas(id) ON DELETE CASCADE,
    FOREIGN KEY (target_idea_id) REFERENCES ideas(id) ON DELETE CASCADE,
    UNIQUE(source_idea_id, target_idea_id, relationship_type)
);
```

**Relationship Types**: `depends_on`, `related_to`, `part_of`, `parent`, `child`, `duplicate`, `blocks`, `blocked_by`, `similar_to`

## Key Design Patterns

### 1. Repository Pattern
The database layer uses the repository pattern for clean separation:
```go
type Repository struct {
    db *sql.DB
}

func (r *Repository) Create(idea *models.Idea) error
func (r *Repository) GetByID(id string) (*models.Idea, error)
func (r *Repository) List(filters ListFilters) ([]*models.Idea, error)
```

### 2. Provider Pattern with Fallback Chain
LLM integration uses a provider interface with automatic fallback:
```go
type Provider interface {
    Name() string
    IsAvailable() bool
    Analyze(req AnalysisRequest) (*AnalysisResult, error)
}
```

Providers: Ollama → Claude → OpenAI → Custom → Rule-based

### 3. Dependency Injection via Context
CLI commands share dependencies through a context struct:
```go
type CLIContext struct {
    Repository *database.Repository
    Engine     *scoring.Engine
    Detector   *patterns.Detector
    Telos      *models.Telos
    LLMManager *llm.Manager
}
```

### 4. Embedded Migrations
Database schema is managed with embedded SQL files:
```go
//go:embed migrations/*.sql
var migrationFiles embed.FS
```

Migrations run automatically on startup, ensuring schema consistency.

### 5. Explicit Error Handling
Go idiom of explicit error returns with context wrapping:
```go
if err := repo.Create(idea); err != nil {
    return fmt.Errorf("failed to create idea: %w", err)
}
```

### 6. Graceful Degradation
System continues to function when external services fail:
- LLM unavailable → Falls back to rule-based scoring
- Database error → Returns error with context
- Network timeout → Retries with next provider

### 7. Structured Logging
All logging uses zerolog for structured, parseable logs:
```go
log.Info().Str("provider", name).Msg("Provider registered")
log.Error().Err(err).Str("id", idea.ID).Msg("Failed to score idea")
```

### 8. Middleware Stack
HTTP server uses composable middleware:
- Recovery from panics
- Request logging
- CORS handling
- Rate limiting
- Security headers
- CSRF protection

## Data Flow

### CLI Flow
1. **User Input**: Command line arguments parsed by Cobra
2. **Initialization**: `PersistentPreRunE` sets up database, telos, scoring engine
3. **Configuration**: Loads from environment variables and default paths
4. **Telos Parsing**: Parses user's telos.md for goals, strategies, patterns
5. **Scoring**: Idea evaluated against telos configuration
6. **AI Enhancement**: Optional LLM analysis with fallback chain
7. **Persistence**: Results stored in SQLite database with WAL mode
8. **Output**: Color-coded results displayed to terminal

### Web API Flow
1. **HTTP Request**: Received by Chi router
2. **Middleware**: Request passes through middleware stack
3. **Handler**: Endpoint handler processes request
4. **Validation**: Input validated at boundary
5. **Business Logic**: Scoring engine or other services invoked
6. **Database**: Repository performs data operations
7. **Response**: JSON response returned with appropriate status code

### LLM Analysis Flow
1. **Manager**: Selects primary provider based on configuration
2. **Cache Check**: Checks similarity-based cache for recent analysis
3. **Provider Call**: Attempts primary provider (e.g., Ollama)
4. **Fallback**: On failure, tries next provider in chain
5. **Processing**: Response parsed and validated
6. **Quality Tracking**: Metrics recorded for monitoring
7. **Rule-based Fallback**: If all providers fail, uses deterministic scoring
8. **Metrics**: Records requests, tokens, errors, latency, fallbacks

## Scoring Algorithm

### Total: 10 Points (0-10 scale)

#### Mission Alignment (4.0 points - 40%)
- **Domain Expertise** (0-1.2): Hotel/hospitality industry keywords
- **AI Alignment** (0-1.5): AI/ML/agent keywords and patterns
- **Execution Support** (0-0.8): Rapid building, shipping, testing
- **Revenue Potential** (0-0.5): Monetization and revenue keywords

#### Anti-Challenge (3.5 points - 35%)
- **Context Switching** (0-1.2): Single focus, minimal switching
- **Rapid Prototyping** (0-1.0): Fast iteration, quick validation
- **Accountability** (0-0.8): Public building, progress sharing
- **Income Anxiety** (0-0.5): Revenue generation, income stability

#### Strategic Fit (2.5 points - 25%)
- **Stack Compatibility** (0-1.0): Matches user's tech stack
- **Shipping Habit** (0-0.8): Encourages regular shipping
- **Public Accountability** (0-0.4): Public building, sharing
- **Revenue Testing** (0-0.3): Quick revenue validation

### Recommendation Tiers
- **8.5+**: 🔥 PRIORITIZE NOW
- **7.0-8.4**: ✅ GOOD ALIGNMENT
- **5.0-6.9**: ⚠️ CONSIDER LATER
- **<5.0**: 🚫 AVOID FOR NOW

## Performance Optimizations

### Database
- **WAL Mode**: Write-Ahead Logging for concurrent reads
- **Connection Pooling**: Max 5 connections, 2 idle, 5-minute lifetime
- **Pragma Settings**: 64MB cache, synchronous=NORMAL, temp_store=MEMORY
- **Indexes**: Created on frequently queried columns

### Scoring Engine
- **Pre-compiled Regex**: Patterns compiled once at initialization
- **Case-insensitive Matching**: Lowercase conversion done once
- **Capped Scores**: Maximum values prevent recalculation

### LLM Integration
- **Similarity Cache**: Avoid redundant LLM calls for similar ideas
- **Response Caching**: TTL-based in-memory cache
- **Health Caching**: Avoid frequent health checks
- **Concurrent Provider Checks**: Parallel availability verification

### Logging
- **Log Rotation**: Lumberjack for automatic rotation (100MB max, 7-day retention)
- **Structured Logging**: Zero-allocation JSON marshaling
- **Conditional Logging**: Debug logs only when enabled

## Security Considerations

### Input Validation
- All user inputs validated at boundaries
- Idea content length limits enforced
- Relationship type validation against allowed values

### SQL Injection Prevention
- Parameterized queries using `?` placeholders
- No string concatenation for SQL
- Prepared statements for repeated queries

### Path Traversal Prevention
- Validated file path operations
- Restricted to configured directories
- Sanitized user-provided paths

### API Security
- CORS configuration with allowed origins
- CSRF token validation for state-changing operations
- Rate limiting to prevent abuse
- Security headers (X-Frame-Options, X-Content-Type-Options, etc.)

### Sensitive Data
- API keys loaded from environment variables
- No hardcoded credentials
- Structured logging sanitizes sensitive fields

### Resource Protection
- Connection pool limits prevent database exhaustion
- HTTP timeouts prevent hung connections
- Context timeouts for long-running operations
- Graceful shutdown prevents data loss

## Extension Points

The architecture supports several extension points:

### 1. Scoring Strategies
Implement custom scoring logic by:
- Creating new scoring functions in `internal/scoring/`
- Adding custom pattern detection in `internal/patterns/`
- Extending telos configuration format

### 2. LLM Providers
Add new LLM providers by:
- Implementing the `Provider` interface in `internal/llm/`
- Registering provider with manager
- Configuring fallback priority

### 3. CLI Commands
Add new commands by:
- Creating command file in `internal/cli/`
- Implementing command logic with shared `CLIContext`
- Registering command with root command

### 4. API Endpoints
Add new endpoints by:
- Adding handler in `internal/api/handlers.go`
- Registering route in `internal/api/server.go`
- Implementing business logic in appropriate package

### 5. Export Formats
Add new export formats by:
- Creating exporter in `internal/export/`
- Implementing export interface
- Wiring into dump command

### 6. Storage Backends
Swap SQLite for other databases by:
- Implementing repository interface
- Updating migration system
- Configuring connection in `internal/database/`

### 7. Analytics
Add new analytics by:
- Extending `internal/analytics/` with new calculations
- Adding visualization functions
- Exposing via CLI or API

## Dependencies

### Core Libraries
- **CLI**: `github.com/spf13/cobra` - Command-line framework
- **HTTP Router**: `github.com/go-chi/chi/v5` - Lightweight router
- **Database**: `github.com/mattn/go-sqlite3` - SQLite driver
- **Logging**: `github.com/rs/zerolog` - Structured logging
- **Log Rotation**: `gopkg.in/natefinch/lumberjack.v2` - Log rotation

### Supporting Libraries
- **Color Output**: `github.com/fatih/color` - Terminal colors
- **Clipboard**: `github.com/atotto/clipboard` - Clipboard access
- **UUID**: `github.com/google/uuid` - UUID generation
- **CORS**: `github.com/go-chi/cors` - CORS middleware
- **HTTP Middleware**: `github.com/go-chi/chi/v5/middleware` - Standard middleware

### Standard Library
Extensive use of Go standard library:
- `database/sql` - Database operations
- `net/http` - HTTP client and server
- `context` - Request context and cancellation
- `encoding/json` - JSON marshaling
- `os/signal` - Graceful shutdown
- `embed` - Embedded migrations

## Testing Strategy

The codebase includes comprehensive tests throughout:
- Unit tests for scoring engine (`scoring/engine_test.go`)
- Repository tests (`database/repository_test.go`)
- API handler tests (`api/handlers_test.go`)
- Telos parser tests (`telos/parser_test.go`)
- LLM provider tests (`llm/provider_test.go`)

Run tests with: `go test ./...`

## Development Workflow

### Local Development
1. Clone repository
2. Install Go 1.25.4+
3. Run `go mod download`
4. Create `~/.telos/telos.md` with your configuration
5. Run CLI: `go run ./cmd/cli`
6. Run web server: `go run ./cmd/web`

### Database
- Location: `~/.telos/ideas.db` (CLI) or `data/telos.db` (web)
- Migrations run automatically on startup
- WAL mode enabled for concurrent access

### Configuration
Set environment variables:
- `PORT`: Web server port (default: 8080)
- `DB_PATH`: Database location
- `TELOS_PATH`: Telos configuration file
- `ANTHROPIC_API_KEY`: Claude API key
- `OPENAI_API_KEY`: OpenAI API key
- `OLLAMA_ENDPOINT`: Ollama server URL

## Observability

### Logging
- Structured JSON logs via zerolog
- Log files: `~/.telos-idea-matrix/logs/`
- Automatic rotation (100MB max, 7-day retention)
- Configurable log level

### Metrics
In-memory metrics tracking:
- LLM requests, successes, failures
- Token usage (input/output)
- Error classification (timeout, rate_limit, auth, network)
- Cache hits/misses
- Provider fallbacks
- Response latency

### Health Checks
- Database connectivity
- Disk space
- Memory usage
- LLM provider availability
- Exposed via `/health` endpoint

## Deployment

### CLI Deployment
1. Build binary: `go build -o tm ./cmd/cli`
2. Install: `mv tm /usr/local/bin/`
3. Configure: Create `~/.telos/telos.md`
4. Run: `tm score "your idea"`

### Web Deployment
1. Build binary: `go build -o tm-web ./cmd/web`
2. Set environment variables
3. Run with process manager (systemd, supervisor)
4. Configure reverse proxy (nginx, caddy)
5. Enable HTTPS

### Docker
Dockerfile example:
```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o tm-web ./cmd/web

FROM alpine:latest
RUN apk --no-cache add ca-certificates sqlite
COPY --from=builder /app/tm-web /usr/local/bin/
CMD ["tm-web"]
```
