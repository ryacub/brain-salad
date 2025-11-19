# Track 8C: Testing & Documentation

**Phase**: 8 - Polish & Documentation
**Estimated Time**: 4-5 hours
**Dependencies**: 7, 8A, 8B (waits for other Sprint 4 tracks)
**Can Run in Parallel**: Partial (start after others complete)

---

## Mission

You are completing integration tests and documentation for the Telos Idea Matrix Go migration.

## Context

- Currently 0% test coverage on CLI and config layers
- Need integration tests for all commands
- Documentation needs updating for Go implementation
- Migration guide for users switching from Rust

## Your Task

Complete testing and documentation to achieve production readiness.

## Directory Structure

Create/update:
- `go/internal/cli/integration_test.go` - CLI integration tests
- `go/internal/config/config_test.go` - Config tests
- `README.md` - Update for Go implementation
- `MIGRATION.md` - Migration guide from Rust
- `docs/API.md` - Update API documentation

## Testing Tasks

### A. CLI Integration Tests

Create `go/internal/cli/integration_test.go`:

```go
// +build integration

package cli

import (
    "os"
    "testing"
    
    "github.com/stretchr/testify/assert"
)

func TestDumpCommand_Integration(t *testing.T) {
    // Setup test database
    tmpDB := setupTestDB(t)
    defer os.Remove(tmpDB)
    
    // Setup test telos file
    tmpTelos := setupTestTelos(t)
    defer os.Remove(tmpTelos)
    
    ctx := &CLIContext{
        DBPath:    tmpDB,
        TelosPath: tmpTelos,
    }
    
    cmd := NewDumpCommand(ctx)
    cmd.SetArgs([]string{"Build a Python automation tool"})
    
    err := cmd.Execute()
    assert.NoError(t, err)
    
    // Verify idea was saved
    ideas, err := ctx.Repo.List(0, "", 10)
    assert.NoError(t, err)
    assert.Len(t, ideas, 1)
    assert.Contains(t, ideas[0].Content, "Python")
}

func TestReviewCommand_Integration(t *testing.T) {
    // Test review with filters
}

func TestPruneCommand_Integration(t *testing.T) {
    // Test prune dry-run and actual execution
}
```

Run: `go test -tags=integration ./internal/cli -v`

### B. Config Tests

Create `go/internal/config/config_test.go`:

```go
package config

import (
    "os"
    "testing"
    
    "github.com/stretchr/testify/assert"
)

func TestConfig_LoadFromFile(t *testing.T) {
    tmpFile := createTestConfigFile(t)
    defer os.Remove(tmpFile)
    
    cfg, err := LoadConfig(tmpFile)
    assert.NoError(t, err)
    assert.NotNil(t, cfg)
    assert.Equal(t, "test_telos.md", cfg.TelosFile)
}

func TestConfig_LoadFromEnv(t *testing.T) {
    os.Setenv("TELOS_FILE", "/tmp/telos.md")
    defer os.Unsetenv("TELOS_FILE")
    
    cfg, err := LoadConfig("")
    assert.NoError(t, err)
    assert.Equal(t, "/tmp/telos.md", cfg.TelosFile)
}

func TestConfig_Defaults(t *testing.T) {
    cfg := DefaultConfig()
    assert.NotEmpty(t, cfg.DataDir)
    assert.NotEmpty(t, cfg.LogDir)
}
```

Run: `go test ./internal/config -v -cover`

## Documentation Tasks

### C. Update README.md

Add sections:
```markdown
## Go Implementation

The Telos Idea Matrix has been migrated to Go for improved performance and deployment simplicity.

### Features
- ✅ All CLI commands (dump, score, review, analyze, prune, analytics)
- ✅ RESTful API server
- ✅ LLM integration (Ollama)
- ✅ Semantic caching
- ✅ Production-ready (health checks, logging, metrics)

### Installation

**Option 1: Pre-built Binary**
```bash
# Download from releases
wget https://github.com/ryacub/brain-salad/releases/latest/download/tm-linux-amd64
chmod +x tm-linux-amd64
sudo mv tm-linux-amd64 /usr/local/bin/tm
```

**Option 2: Build from Source**
```bash
git clone https://github.com/ryacub/brain-salad
cd brain-salad/go
make build
sudo mv bin/tm /usr/local/bin/
```

### Migration from Rust

See [MIGRATION.md](./MIGRATION.md) for detailed migration guide.
```

### D. Create MIGRATION.md

```markdown
# Migrating from Rust to Go Implementation

## Database Compatibility

The Go implementation uses the **same SQLite schema** as Rust, so your existing database will work without modification.

## Configuration

### Rust Config
```toml
# ~/.config/telos-matrix/config.toml
telos_file = "/path/to/telos.md"
data_dir = "~/.local/share/telos-matrix"
```

### Go Config
Same locations supported:
- Environment variable: `TELOS_FILE=/path/to/telos.md`
- Config file: `~/.config/telos-matrix/config.toml` (same format)
- Current directory: `./telos.md`

## Command Mapping

| Rust | Go | Notes |
|------|-----|-------|
| `tm dump` | `tm dump` | Same |
| `tm analyze` | `tm analyze` | Add `--ai` for LLM |
| `tm review` | `tm review` | Same |
| `tm prune` | `tm prune` | Same |
| `tm analytics` | `tm analytics` | Enhanced with trends |
| N/A | `tm llm` | New: service management |
| N/A | `tm bulk` | New: bulk operations |

## Feature Parity

### ✅ Fully Migrated
- Core scoring engine (exact algorithm match)
- Pattern detection
- Database operations
- All CLI commands
- Telos parsing

### ✨ Enhanced in Go
- RESTful API server
- Bulk operations
- Enhanced analytics
- Better error handling

### ⚠️ Rust-Only Features (Optional)
If you need these, keep Rust installed:
- Advanced LLM prompt management
- Custom quality metrics beyond basic
- Circuit breaker advanced configuration

## Migration Steps

1. **Backup your data**:
   ```bash
   cp ~/.local/share/telos-matrix/ideas.db ~/.local/share/telos-matrix/ideas.db.backup
   ```

2. **Install Go version**:
   ```bash
   # See README.md for installation options
   ```

3. **Test with existing database**:
   ```bash
   tm review --limit 5
   # Should show your existing ideas
   ```

4. **Optional: Remove Rust version**:
   ```bash
   cargo uninstall telos-idea-matrix
   ```

## Troubleshooting

### Issue: "Database locked"
**Solution**: Make sure no Rust version is running simultaneously

### Issue: "Telos file not found"
**Solution**: Set `TELOS_FILE` environment variable or use `--telos` flag

### Issue: "Scores don't match Rust"
**Solution**: Scoring algorithm is identical; minor differences (<0.1) are due to floating-point precision

## Performance Comparison

| Operation | Rust | Go | Improvement |
|-----------|------|-----|-------------|
| Startup time | 50ms | 30ms | 40% faster |
| Score 1 idea | 5ms | 3ms | 40% faster |
| List 1000 ideas | 100ms | 60ms | 40% faster |
| API response | N/A | <50ms | New feature |

## Support

- Issues: https://github.com/ryacub/brain-salad/issues
- Discussions: https://github.com/ryacub/brain-salad/discussions
```

### E. Update Docker Documentation

Update `docs/DOCKER_GUIDE.md` for Go implementation.

## Success Criteria

- ✅ CLI integration tests achieve >70% coverage
- ✅ Config tests achieve >85% coverage
- ✅ Documentation accurate and complete
- ✅ Migration guide tested by fresh user
- ✅ Overall Go test coverage >85%

## Validation

```bash
# Run all tests with coverage
go test ./... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out

# Integration tests
go test -tags=integration ./... -v

# Build and smoke test
make build
./bin/tm --version
./bin/tm dump "Test idea"
./bin/tm review
```

## Deliverables

- `go/internal/cli/integration_test.go`
- `go/internal/config/config_test.go`
- Updated `README.md`
- New `MIGRATION.md`
- Updated `docs/API.md`
- Updated `docs/DOCKER_GUIDE.md`
