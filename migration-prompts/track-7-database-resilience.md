# Track 7: Database Resilience

**Phase**: 7 - Database Enhancement
**Estimated Time**: 12-16 hours
**Dependencies**: None
**Can Run in Parallel**: Yes (with 8A, 8B)

---

## Mission

You are implementing production-ready database resilience features for the Telos Idea Matrix Go application, following Test-Driven Development (TDD).

## Context

- The Rust implementation has resilient database layer in `src/database_simple.rs`
- Need connection pooling, retry logic, migrations, and relationship support
- Must handle transient errors (locked database, connection failures)
- Support for idea relationships (depends-on, related-to, blocks)

## Reference Implementation

Review `/home/user/brain-salad/src/database_simple.rs` for:
- Connection pool configuration
- Retry logic with exponential backoff
- Transaction handling
- Relationship tables

## Your Task

Implement database resilience features using strict TDD methodology.

## Directory Structure

Enhance `go/internal/database/`:
- `repository.go` - Add connection pooling, retries
- `migrations.go` - Migration system
- `relationships.go` - Idea relationships
- `health.go` - Database health checks
- `repository_test.go` - Expand tests
- `migrations_test.go` - Migration tests

## TDD Workflow (RED → GREEN → REFACTOR)

### STEP 1 - RED PHASE (Write Failing Tests)

Expand `go/internal/database/repository_test.go`:
- `TestRepository_ConnectionPool_Config()`
- `TestRepository_Retry_OnTransientError()`
- `TestRepository_Metrics_QueryDuration()`
- `TestRepository_HealthCheck()`
- `TestRepository_ConcurrentAccess()` (1000 goroutines)

Create `go/internal/database/migrations_test.go`:
- `TestMigrations_Apply()`
- `TestMigrations_Rollback()`
- `TestMigrations_Version()`

Run: `go test ./internal/database -v`
Expected: **ALL TESTS FAIL**

### STEP 2 - GREEN PHASE (Implement)

#### A. Enhance `go/internal/database/repository.go`:

```go
package database

import (
    "context"
    "database/sql"
    "fmt"
    "time"
    
    _ "github.com/mattn/go-sqlite3"
)

type Config struct {
    Path            string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
    RetryAttempts   int
    RetryDelay      time.Duration
}

type Repository struct {
    db     *sql.DB
    config Config
}

func NewRepository(config Config) (*Repository, error) {
    db, err := sql.Open("sqlite3", config.Path)
    if err != nil {
        return nil, fmt.Errorf("open database: %w", err)
    }
    
    // Configure connection pool
    db.SetMaxOpenConns(config.MaxOpenConns)
    db.SetMaxIdleConns(config.MaxIdleConns)
    db.SetConnMaxLifetime(config.ConnMaxLifetime)
    
    // Test connection
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("ping database: %w", err)
    }
    
    repo := &Repository{
        db:     db,
        config: config,
    }
    
    // Run migrations
    if err := repo.runMigrations(); err != nil {
        return nil, fmt.Errorf("run migrations: %w", err)
    }
    
    return repo, nil
}

// ExecuteWithRetry executes a database operation with retry logic
func (r *Repository) ExecuteWithRetry(ctx context.Context, fn func() error) error {
    var lastErr error
    
    for attempt := 0; attempt < r.config.RetryAttempts; attempt++ {
        err := fn()
        if err == nil {
            return nil
        }
        
        // Check if error is retryable
        if !isRetryableError(err) {
            return err
        }
        
        lastErr = err
        
        // Exponential backoff
        if attempt < r.config.RetryAttempts-1 {
            delay := r.config.RetryDelay * time.Duration(1<<uint(attempt))
            select {
            case <-time.After(delay):
            case <-ctx.Done():
                return ctx.Err()
            }
        }
    }
    
    return fmt.Errorf("max retries exceeded: %w", lastErr)
}

func isRetryableError(err error) bool {
    // SQLite "database is locked" error
    errStr := err.Error()
    return strings.Contains(errStr, "database is locked") ||
           strings.Contains(errStr, "database table is locked")
}

// Transaction executes function within a transaction
func (r *Repository) Transaction(ctx context.Context, fn func(*sql.Tx) error) error {
    return r.ExecuteWithRetry(ctx, func() error {
        tx, err := r.db.BeginTx(ctx, nil)
        if err != nil {
            return err
        }
        
        defer func() {
            if p := recover(); p != nil {
                tx.Rollback()
                panic(p)
            }
        }()
        
        if err := fn(tx); err != nil {
            tx.Rollback()
            return err
        }
        
        return tx.Commit()
    })
}

// HealthCheck verifies database connectivity
func (r *Repository) HealthCheck(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()
    
    return r.db.PingContext(ctx)
}

// Stats returns database statistics
func (r *Repository) Stats() sql.DBStats {
    return r.db.Stats()
}
```

#### B. Implement `go/internal/database/migrations.go`:

```go
package database

import (
    "database/sql"
    "fmt"
)

type Migration struct {
    Version int
    Name    string
    Up      string
    Down    string
}

var migrations = []Migration{
    {
        Version: 1,
        Name:    "create_ideas_table",
        Up: `
            CREATE TABLE IF NOT EXISTS ideas (
                id TEXT PRIMARY KEY,
                title TEXT NOT NULL,
                content TEXT NOT NULL,
                score REAL NOT NULL,
                status TEXT NOT NULL DEFAULT 'active',
                created_at DATETIME NOT NULL,
                updated_at DATETIME NOT NULL
            );
            CREATE INDEX idx_ideas_score ON ideas(score);
            CREATE INDEX idx_ideas_status ON ideas(status);
            CREATE INDEX idx_ideas_created_at ON ideas(created_at);
        `,
        Down: `DROP TABLE IF EXISTS ideas;`,
    },
    {
        Version: 2,
        Name:    "create_relationships_table",
        Up: `
            CREATE TABLE IF NOT EXISTS idea_relationships (
                id TEXT PRIMARY KEY,
                from_idea_id TEXT NOT NULL,
                to_idea_id TEXT NOT NULL,
                relationship_type TEXT NOT NULL,
                created_at DATETIME NOT NULL,
                FOREIGN KEY (from_idea_id) REFERENCES ideas(id) ON DELETE CASCADE,
                FOREIGN KEY (to_idea_id) REFERENCES ideas(id) ON DELETE CASCADE
            );
            CREATE INDEX idx_relationships_from ON idea_relationships(from_idea_id);
            CREATE INDEX idx_relationships_to ON idea_relationships(to_idea_id);
        `,
        Down: `DROP TABLE IF EXISTS idea_relationships;`,
    },
}

func (r *Repository) runMigrations() error {
    // Create migrations table
    _, err := r.db.Exec(`
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version INTEGER PRIMARY KEY,
            name TEXT NOT NULL,
            applied_at DATETIME NOT NULL
        )
    `)
    if err != nil {
        return fmt.Errorf("create migrations table: %w", err)
    }
    
    // Get current version
    var currentVersion int
    err = r.db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&currentVersion)
    if err != nil {
        return fmt.Errorf("get current version: %w", err)
    }
    
    // Apply pending migrations
    for _, migration := range migrations {
        if migration.Version <= currentVersion {
            continue
        }
        
        // Execute migration
        if _, err := r.db.Exec(migration.Up); err != nil {
            return fmt.Errorf("apply migration %d: %w", migration.Version, err)
        }
        
        // Record migration
        _, err := r.db.Exec(
            "INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)",
            migration.Version, migration.Name, time.Now(),
        )
        if err != nil {
            return fmt.Errorf("record migration %d: %w", migration.Version, err)
        }
    }
    
    return nil
}

func (r *Repository) GetMigrationVersion() (int, error) {
    var version int
    err := r.db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&version)
    return version, err
}
```

#### C. Implement `go/internal/database/relationships.go`:

```go
package database

import (
    "context"
    "time"
    
    "github.com/google/uuid"
)

type RelationshipType string

const (
    DependsOn  RelationshipType = "depends-on"
    RelatedTo  RelationshipType = "related-to"
    Blocks     RelationshipType = "blocks"
    Supersedes RelationshipType = "supersedes"
)

type IdeaRelationship struct {
    ID         string
    FromIdeaID string
    ToIdeaID   string
    Type       RelationshipType
    CreatedAt  time.Time
}

func (r *Repository) CreateRelationship(ctx context.Context, fromID, toID string, relType RelationshipType) (*IdeaRelationship, error) {
    rel := &IdeaRelationship{
        ID:         uuid.New().String(),
        FromIdeaID: fromID,
        ToIdeaID:   toID,
        Type:       relType,
        CreatedAt:  time.Now(),
    }
    
    err := r.ExecuteWithRetry(ctx, func() error {
        _, err := r.db.ExecContext(ctx,
            `INSERT INTO idea_relationships (id, from_idea_id, to_idea_id, relationship_type, created_at)
             VALUES (?, ?, ?, ?, ?)`,
            rel.ID, rel.FromIdeaID, rel.ToIdeaID, rel.Type, rel.CreatedAt,
        )
        return err
    })
    
    if err != nil {
        return nil, err
    }
    
    return rel, nil
}

func (r *Repository) GetRelationships(ctx context.Context, ideaID string) ([]*IdeaRelationship, error) {
    rows, err := r.db.QueryContext(ctx,
        `SELECT id, from_idea_id, to_idea_id, relationship_type, created_at
         FROM idea_relationships
         WHERE from_idea_id = ? OR to_idea_id = ?`,
        ideaID, ideaID,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    relationships := make([]*IdeaRelationship, 0)
    for rows.Next() {
        var rel IdeaRelationship
        if err := rows.Scan(&rel.ID, &rel.FromIdeaID, &rel.ToIdeaID, &rel.Type, &rel.CreatedAt); err != nil {
            return nil, err
        }
        relationships = append(relationships, &rel)
    }
    
    return relationships, rows.Err()
}
```

Run: `go test ./internal/database -v`
Expected: **ALL TESTS PASS**

### STEP 3 - REFACTOR PHASE

- Extract connection pool configuration
- Optimize prepared statement usage
- Add connection pool monitoring
- Performance testing with 10,000+ ideas

## Success Criteria

- ✅ All tests pass with >90% coverage
- ✅ Handles 1000 concurrent connections
- ✅ Retries work on transient errors (locked database)
- ✅ Migrations reversible
- ✅ Matches Rust `src/database_simple.rs` resilience

## Validation

```bash
# Unit tests
go test ./internal/database -v -cover -race

# Concurrent access test
go test ./internal/database -v -run TestRepository_ConcurrentAccess

# Migration test
go test ./internal/database -v -run TestMigrations
```

## Deliverables

- Enhanced `go/internal/database/repository.go`
- `go/internal/database/migrations.go`
- `go/internal/database/relationships.go`
- `go/internal/database/health.go`
- Comprehensive tests
