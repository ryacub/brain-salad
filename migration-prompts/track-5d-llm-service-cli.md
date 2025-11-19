# Track 5D: LLM Service Management & CLI Commands

**Phase**: 5 - LLM Integration
**Estimated Time**: 6-8 hours
**Dependencies**: 5A, 5B, 5C (needs all LLM components)
**Can Run in Parallel**: No (must wait for Sprint 2 completion)

---

## Mission

You are implementing Ollama service management and LLM-powered CLI commands for the Telos Idea Matrix Go application, following Test-Driven Development (TDD).

## Context

- The Rust implementation has Ollama service management in `src/llm_fallback.rs`
- CLI commands in `src/commands/llm.rs` and `src/commands/analyze_llm.rs`
- Need to start/stop/status Ollama service
- Add `--ai` flag to existing commands for LLM-powered analysis

## Reference Implementation

Review:
- `/home/user/brain-salad/src/llm_fallback.rs` - Service management
- `/home/user/brain-salad/src/commands/llm.rs` - LLM CLI commands
- `/home/user/brain-salad/src/commands/analyze_llm.rs` - AI-powered analysis

## Your Task

Implement LLM service management and CLI integration using strict TDD methodology.

**IMPORTANT**: This track requires 5A, 5B, and 5C to be complete. Do not start until Sprint 2 is finished.

## Directory Structure

Create files in `go/internal/llm/service/` and update `go/internal/cli/`:
- `service/manager.go` - Ollama service management
- `service/manager_test.go` - Service tests
- `cli/llm.go` - LLM service CLI commands
- `cli/analyze_llm.go` - AI-powered analysis command
- Update `cli/dump.go` - Add `--ai` flag
- Update `cli/analyze.go` - Add `--ai` flag

## TDD Workflow (RED → GREEN → REFACTOR)

### STEP 1 - RED PHASE (Write Failing Tests)

Create `go/internal/llm/service/manager_test.go`:
- `TestServiceManager_CheckStatus()`
- `TestServiceManager_ListModels()`
- `TestServiceManager_StartOllama()` (if not running)
- `TestServiceManager_StopOllama()` (gracefully)

Create `go/internal/cli/llm_test.go`:
- `TestLlmCommand_Status()`
- `TestLlmCommand_Models()`

Run: `go test ./internal/llm/service ./internal/cli -v`
Expected: **ALL TESTS FAIL**

### STEP 2 - GREEN PHASE (Implement)

#### A. Implement `go/internal/llm/service/manager.go`:

```go
package service

import (
    "context"
    "fmt"
    "os/exec"
    "time"
    
    "github.com/rayyacub/telos-idea-matrix/internal/llm/client"
)

type ServiceManager struct {
    client *client.OllamaClient
}

func NewServiceManager() *ServiceManager {
    return &ServiceManager{
        client: client.NewOllamaClient("", 5*time.Second),
    }
}

// CheckStatus verifies if Ollama is running
func (sm *ServiceManager) CheckStatus() (bool, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    
    err := sm.client.HealthCheck(ctx)
    return err == nil, err
}

// ListModels returns available Ollama models
func (sm *ServiceManager) ListModels() ([]string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    return sm.client.ListModels(ctx)
}

// StartOllama attempts to start Ollama service
func (sm *ServiceManager) StartOllama() error {
    // Check if already running
    if running, _ := sm.CheckStatus(); running {
        return fmt.Errorf("ollama is already running")
    }
    
    // Try to start via `ollama serve` in background
    cmd := exec.Command("ollama", "serve")
    if err := cmd.Start(); err != nil {
        return fmt.Errorf("failed to start ollama: %w", err)
    }
    
    // Wait for service to be ready (up to 10 seconds)
    for i := 0; i < 20; i++ {
        time.Sleep(500 * time.Millisecond)
        if running, _ := sm.CheckStatus(); running {
            return nil
        }
    }
    
    return fmt.Errorf("ollama started but not responding")
}

// StopOllama attempts to stop Ollama service gracefully
func (sm *ServiceManager) StopOllama() error {
    // Check if running
    if running, _ := sm.CheckStatus(); !running {
        return fmt.Errorf("ollama is not running")
    }
    
    // Find and kill ollama process
    cmd := exec.Command("pkill", "-TERM", "ollama")
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to stop ollama: %w", err)
    }
    
    // Wait for shutdown (up to 5 seconds)
    for i := 0; i < 10; i++ {
        time.Sleep(500 * time.Millisecond)
        if running, _ := sm.CheckStatus(); !running {
            return nil
        }
    }
    
    // Force kill if still running
    exec.Command("pkill", "-KILL", "ollama").Run()
    return nil
}
```

#### B. Implement `go/internal/cli/llm.go`:

```go
package cli

import (
    "fmt"
    
    "github.com/fatih/color"
    "github.com/spf13/cobra"
    "github.com/rayyacub/telos-idea-matrix/internal/llm/service"
)

func NewLlmCommand(ctx *CLIContext) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "llm",
        Short: "Manage LLM service (Ollama)",
        Long:  "Manage Ollama LLM service: check status, list models, start/stop service",
    }
    
    cmd.AddCommand(newLlmStatusCommand(ctx))
    cmd.AddCommand(newLlmModelsCommand(ctx))
    cmd.AddCommand(newLlmStartCommand(ctx))
    cmd.AddCommand(newLlmStopCommand(ctx))
    
    return cmd
}

func newLlmStatusCommand(ctx *CLIContext) *cobra.Command {
    return &cobra.Command{
        Use:   "status",
        Short: "Check Ollama service status",
        RunE: func(cmd *cobra.Command, args []string) error {
            sm := service.NewServiceManager()
            
            running, err := sm.CheckStatus()
            if running {
                color.Green("✓ Ollama is running")
                
                // List models
                models, err := sm.ListModels()
                if err == nil && len(models) > 0 {
                    fmt.Println("\nAvailable models:")
                    for _, model := range models {
                        fmt.Printf("  - %s\n", model)
                    }
                }
                return nil
            }
            
            color.Red("✗ Ollama is not running")
            if err != nil {
                fmt.Printf("Error: %v\n", err)
            }
            fmt.Println("\nStart Ollama with: tm llm start")
            return nil
        },
    }
}

func newLlmModelsCommand(ctx *CLIContext) *cobra.Command {
    return &cobra.Command{
        Use:   "models",
        Short: "List available Ollama models",
        RunE: func(cmd *cobra.Command, args []string) error {
            sm := service.NewServiceManager()
            
            models, err := sm.ListModels()
            if err != nil {
                return fmt.Errorf("failed to list models: %w", err)
            }
            
            if len(models) == 0 {
                fmt.Println("No models installed")
                fmt.Println("\nInstall a model with: ollama pull llama2")
                return nil
            }
            
            fmt.Println("Available models:")
            for _, model := range models {
                fmt.Printf("  - %s\n", model)
            }
            return nil
        },
    }
}

func newLlmStartCommand(ctx *CLIContext) *cobra.Command {
    return &cobra.Command{
        Use:   "start",
        Short: "Start Ollama service",
        RunE: func(cmd *cobra.Command, args []string) error {
            sm := service.NewServiceManager()
            
            fmt.Println("Starting Ollama service...")
            if err := sm.StartOllama(); err != nil {
                return fmt.Errorf("failed to start: %w", err)
            }
            
            color.Green("✓ Ollama started successfully")
            return nil
        },
    }
}

func newLlmStopCommand(ctx *CLIContext) *cobra.Command {
    return &cobra.Command{
        Use:   "stop",
        Short: "Stop Ollama service",
        RunE: func(cmd *cobra.Command, args []string) error {
            sm := service.NewServiceManager()
            
            fmt.Println("Stopping Ollama service...")
            if err := sm.StopOllama(); err != nil {
                return fmt.Errorf("failed to stop: %w", err)
            }
            
            color.Green("✓ Ollama stopped successfully")
            return nil
        },
    }
}
```

#### C. Update `go/internal/cli/analyze.go` to add `--ai` flag:

```go
// Add to analyze command
var useAI bool
analyzeCmd.Flags().BoolVar(&useAI, "ai", false, "Use AI-powered analysis (requires Ollama)")

// In RunE function:
if useAI {
    // Use LLM provider with cache
    cache := llmcache.NewCache()
    provider := llm.NewOllamaProvider("", "llama2")
    cachedProvider := &CachedProvider{provider: provider, cache: cache}
    
    result, err := cachedProvider.Analyze(llm.AnalysisRequest{
        IdeaContent: ideaContent,
        TelosPath:   telosPath,
    })
    // Display result with cache hit indicator
} else {
    // Use rule-based scoring (existing code)
}
```

Run: `go test ./internal/llm/service ./internal/cli -v`
Expected: **ALL TESTS PASS**

### STEP 3 - REFACTOR PHASE

- Add Ollama auto-start on first use
- Optimize model download/pull
- Add model recommendation based on hardware
- Extract CLI display formatting

## Integration

1. Wire cache into all `--ai` commands
2. Add quality tracking for all LLM analyses
3. Display cache hit/miss stats in verbose mode
4. Add `tm llm install <model>` command (optional)

## Success Criteria

- ✅ All tests pass with >75% coverage (lower due to exec/system calls)
- ✅ Service management works on Linux/macOS
- ✅ Graceful handling of missing Ollama
- ✅ Cache integration reduces API calls by >60%
- ✅ `--ai` flag works on dump and analyze commands

## Validation

```bash
# Service management
tm llm status
tm llm start
tm llm models
tm llm stop

# AI-powered analysis
tm analyze --ai "Build a Python automation tool"
tm dump --ai "Create a hotel booking system"

# Verify caching
tm analyze --ai "Build automation tool"
tm analyze --ai "Create automation tool"
# Second should be instant (cache hit)
```

## Deliverables

- `go/internal/llm/service/manager.go`
- `go/internal/llm/service/manager_test.go`
- `go/internal/cli/llm.go`
- Updated `go/internal/cli/analyze.go` (add --ai flag)
- Updated `go/internal/cli/dump.go` (add --ai flag)
- Integration with cache and quality tracking
