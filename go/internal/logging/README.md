# Logging Package

Structured logging using zerolog for the Telos Idea Matrix application.

## Features

- **JSON structured logging** for easy parsing and analysis
- **Multiple log levels**: debug, info, warn, error
- **File rotation**: automatic rotation based on size, age, and backup count
- **Console output option** for development
- **HTTP request/response middleware** with timing and status tracking
- **Caller information** included in logs

## Usage

### Initialize Logger

```go
import "github.com/rayyacub/telos-idea-matrix/internal/logging"

cfg := logging.Config{
    Level:      "info",
    Format:     "json",        // "json" or "console"
    OutputPath: "/var/log/telos-matrix.log", // or "stdout"
    MaxSizeMB:  10,            // Max size per file in MB
    MaxBackups: 7,             // Number of backups to keep
    MaxAgeDays: 7,             // Max age in days
}

logger := logging.NewLogger(cfg)
```

### Use in HTTP Server

```go
import (
    "github.com/go-chi/chi/v5"
    "github.com/rayyacub/telos-idea-matrix/internal/logging"
)

r := chi.NewRouter()
r.Use(logging.Middleware)
```

### Log Messages

```go
import "github.com/rs/zerolog/log"

log.Info().Str("user_id", "123").Msg("User logged in")
log.Warn().Err(err).Msg("Failed to process request")
log.Error().Str("path", path).Int("status", 500).Msg("Server error")
```

## Log Format

JSON format example:
```json
{
  "level": "info",
  "time": "2025-01-19T10:30:00Z",
  "caller": "/app/handlers.go:42",
  "method": "POST",
  "path": "/api/v1/ideas",
  "status": 201,
  "duration_ms": 45,
  "message": "request completed"
}
```

## Performance

- < 1ms overhead per log statement
- Asynchronous file I/O
- Minimal memory allocation
