# Provider Manager Documentation

## Overview

The Provider Manager is a robust system for managing multiple LLM providers with intelligent fallback, health checks, and dynamic provider selection. It ensures high availability by automatically switching to backup providers when the primary provider fails.

## Features

- **Multi-Provider Support**: Manage multiple LLM providers (Ollama, Claude, OpenAI, Custom, Rule-based)
- **Intelligent Fallback**: Automatic failover to backup providers when primary fails
- **Health Monitoring**: Periodic health checks with caching
- **Statistics Tracking**: Request counts, success/failure rates, and latency metrics
- **Thread-Safe**: Concurrent access with proper synchronization
- **Configurable Priority**: Define provider priority order
- **Dynamic Configuration**: Hot-reload configuration without restart

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/rayyacub/telos-idea-matrix/internal/llm"
    "github.com/rayyacub/telos-idea-matrix/internal/models"
)

func main() {
    // Create manager with default configuration
    config := llm.DefaultManagerConfig()
    manager := llm.NewManager(config)

    // Create a telos for analysis
    telos := &models.Telos{
        Goals: []models.Goal{
            {ID: "g1", Description: "Build great products", Priority: 1},
        },
    }

    // Perform analysis
    result, err := manager.Analyze(llm.AnalysisRequest{
        IdeaContent: "Build an AI-powered task manager",
        Telos:       telos,
    })

    if err != nil {
        fmt.Printf("Analysis failed: %v\n", err)
        return
    }

    fmt.Printf("Score: %.2f\n", result.FinalScore)
    fmt.Printf("Provider: %s\n", result.Provider)
}
```

## Configuration

### Manager Configuration

```go
config := &llm.ManagerConfig{
    // Primary provider to use (if available)
    DefaultProvider: "ollama",

    // Enable automatic fallback to other providers
    FallbackEnabled: true,

    // How often to check provider health
    HealthCheckInterval: 30 * time.Second,

    // Provider priority order (highest to lowest)
    Priority: []string{"ollama", "claude", "openai", "rule_based"},

    // Provider-specific configuration
    ProviderConfig: llm.ProviderConfig{
        OllamaBaseURL: "http://localhost:11434",
        OllamaModel:   "llama2",
        ClaudeAPIKey:  "your-api-key",
        ClaudeModel:   "claude-3-5-sonnet-20241022",
    },
}

manager := llm.NewManager(config)
```

### Environment Variables

You can configure providers using environment variables:

```bash
# Ollama configuration
export OLLAMA_BASE_URL="http://localhost:11434"
export OLLAMA_MODEL="llama2"

# Claude configuration
export CLAUDE_API_KEY="your-api-key"
export CLAUDE_MODEL="claude-3-5-sonnet-20241022"

# OpenAI configuration (when implemented)
export OPENAI_API_KEY="your-api-key"
export OPENAI_MODEL="gpt-4"
```

## Provider Management

### Registering Providers

```go
manager := llm.NewManager(config)

// Providers are automatically registered based on configuration
// You can also register custom providers manually:
customProvider := NewMyCustomProvider()
manager.RegisterProvider(customProvider)
```

### Setting Primary Provider

```go
// Set primary provider by name
err := manager.SetPrimaryProvider("ollama")
if err != nil {
    fmt.Printf("Failed to set primary: %v\n", err)
}

// Get current primary provider
primary := manager.GetPrimaryProvider()
fmt.Printf("Primary provider: %s\n", primary.Name())
```

### Getting Available Providers

```go
// Get all currently available providers
available := manager.GetAvailableProviders()
for _, provider := range available {
    fmt.Printf("Available: %s\n", provider.Name())
}
```

## Health Monitoring

### Manual Health Check

```go
// Perform one-time health check on all providers
status := manager.HealthCheck()
for name, available := range status {
    fmt.Printf("%s: %v\n", name, available)
}
```

### Periodic Health Checks

```go
// Start background health checks
stopCh := make(chan struct{})
go manager.StartPeriodicHealthCheck(stopCh)

// Later, stop health checks
close(stopCh)
```

### Getting Health Status

```go
// Get health status for specific provider
available, lastCheck, err := manager.GetHealthStatus("ollama")
if err != nil {
    fmt.Printf("Error: %v\n", err)
    return
}

fmt.Printf("Ollama available: %v (last checked: %s)\n", available, lastCheck)
```

## Statistics

### Provider Statistics

```go
// Get statistics for all providers
allStats := manager.GetStats()
for _, stats := range allStats {
    fmt.Printf("Provider: %s\n", stats.Name)
    fmt.Printf("  Total Requests: %d\n", stats.TotalRequests)
    fmt.Printf("  Success: %d\n", stats.SuccessCount)
    fmt.Printf("  Failures: %d\n", stats.FailureCount)
    fmt.Printf("  Avg Latency: %s\n", stats.AverageLatency)
    fmt.Printf("  Last Used: %s\n", stats.LastUsed)
}
```

### Getting Stats for Specific Provider

```go
stats, err := manager.GetProviderStats("ollama")
if err != nil {
    fmt.Printf("Error: %v\n", err)
    return
}

successRate := float64(stats.SuccessCount) / float64(stats.TotalRequests) * 100
fmt.Printf("Success Rate: %.2f%%\n", successRate)
```

### Resetting Statistics

```go
// Reset all statistics
manager.ResetStats()
```

## Fallback Behavior

### How Fallback Works

1. Manager tries the primary provider first
2. If primary fails and fallback is enabled:
   - Tries each provider in priority order
   - Skips unavailable providers
   - Returns result from first successful provider
3. If all providers fail, returns error with details

### Enabling/Disabling Fallback

```go
// Disable fallback (only use primary)
manager.EnableFallback(false)

// Enable fallback
manager.EnableFallback(true)

// Check current setting
enabled := manager.IsFallbackEnabled()
```

## Priority Management

### Setting Provider Priority

```go
// Define priority order (highest to lowest)
config := &llm.ManagerConfig{
    Priority: []string{
        "ollama",      // Try first
        "claude",      // Then Claude
        "openai",      // Then OpenAI
        "rule_based",  // Finally rule-based (always works)
    },
}

manager := llm.NewManager(config)
```

### Dynamic Priority Updates

```go
// Update configuration with new priority
newConfig := &llm.ManagerConfig{
    DefaultProvider: "claude",
    FallbackEnabled: true,
    Priority: []string{"claude", "ollama", "rule_based"},
}

err := manager.LoadConfig(newConfig)
if err != nil {
    fmt.Printf("Failed to load config: %v\n", err)
}
```

## Best Practices

### 1. Always Include Rule-Based Provider

The rule-based provider is deterministic and always available. Include it as the last fallback:

```go
Priority: []string{"ollama", "claude", "rule_based"}
```

### 2. Enable Health Checks for Production

```go
config := &llm.ManagerConfig{
    HealthCheckInterval: 30 * time.Second,
    // ... other config
}

manager := llm.NewManager(config)

stopCh := make(chan struct{})
go manager.StartPeriodicHealthCheck(stopCh)

// Cleanup on shutdown
defer close(stopCh)
```

### 3. Monitor Statistics

Regularly check provider statistics to identify issues:

```go
// Log stats periodically
ticker := time.NewTicker(5 * time.Minute)
go func() {
    for range ticker.C {
        stats := manager.GetStats()
        for _, s := range stats {
            if s.FailureCount > s.SuccessCount {
                log.Printf("WARNING: %s has high failure rate", s.Name)
            }
        }
    }
}()
```

### 4. Handle Errors Gracefully

```go
result, err := manager.Analyze(req)
if err != nil {
    // Log the error
    log.Printf("Analysis failed: %v", err)

    // Check provider status
    status := manager.HealthCheck()
    log.Printf("Provider status: %v", status)

    // Optionally fall back to default values
    return defaultResponse()
}
```

### 5. Use Appropriate Timeouts

Configure reasonable timeouts for your use case:

```go
config := llm.ProviderConfig{
    OllamaTimeout: 30, // seconds
    ClaudeTimeout: 30,
}
```

## Thread Safety

The Manager is fully thread-safe and can be used concurrently:

```go
manager := llm.NewManager(config)

// Safe to call from multiple goroutines
for i := 0; i < 10; i++ {
    go func(id int) {
        result, err := manager.Analyze(req)
        // ... handle result
    }(i)
}
```

## Error Handling

### Common Errors

```go
result, err := manager.Analyze(req)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "all providers failed"):
        // All providers are down or failed
        log.Error("System-wide provider failure")

    case strings.Contains(err.Error(), "provider not found"):
        // Invalid provider name
        log.Error("Configuration error")

    case strings.Contains(err.Error(), "fallback disabled"):
        // Primary failed with no fallback
        log.Error("Consider enabling fallback")
    }
}
```

## Performance Considerations

### 1. Provider Selection Impact

- **Ollama (local)**: ~500ms - 2s (no network latency)
- **Claude API**: ~1-3s (network + processing)
- **OpenAI API**: ~1-3s (network + processing)
- **Rule-based**: <10ms (pure computation)

### 2. Fallback Chain Latency

With fallback enabled, worst-case latency is the sum of all provider timeouts. Configure appropriately:

```go
// Fast fail configuration
config := llm.ProviderConfig{
    OllamaTimeout: 5,  // Fail fast if Ollama is down
    ClaudeTimeout: 30, // Allow more time for API
}
```

### 3. Concurrent Requests

The Manager handles concurrent requests efficiently with atomic operations and minimal locking:

```go
// Efficient concurrent processing
var wg sync.WaitGroup
for _, idea := range ideas {
    wg.Add(1)
    go func(content string) {
        defer wg.Done()
        result, _ := manager.Analyze(llm.AnalysisRequest{
            IdeaContent: content,
            Telos:       telos,
        })
        // ... process result
    }(idea)
}
wg.Wait()
```

## Troubleshooting

### Provider Not Available

```go
available := manager.GetAvailableProviders()
if len(available) == 0 {
    log.Error("No providers available!")

    // Check individual provider status
    status := manager.HealthCheck()
    for name, ok := range status {
        if !ok {
            log.Printf("Provider %s is down", name)
        }
    }
}
```

### High Failure Rate

```go
stats, _ := manager.GetProviderStats("ollama")
failureRate := float64(stats.FailureCount) / float64(stats.TotalRequests)

if failureRate > 0.5 {
    // More than 50% failures - investigate
    log.Printf("High failure rate for ollama: %.2f%%", failureRate*100)

    // Consider switching primary provider
    manager.SetPrimaryProvider("claude")
}
```

### Slow Response Times

```go
stats := manager.GetStats()
for _, s := range stats {
    if s.AverageLatency > 5*time.Second {
        log.Printf("WARNING: %s average latency is %s", s.Name, s.AverageLatency)
    }
}
```

## API Reference

See [GoDoc](https://pkg.go.dev/github.com/rayyacub/telos-idea-matrix/internal/llm) for complete API documentation.

## Examples

See the `/examples` directory for complete working examples:
- `examples/basic-manager/` - Basic manager usage
- `examples/fallback-demo/` - Fallback behavior demonstration
- `examples/health-monitoring/` - Health check implementation
- `examples/statistics/` - Statistics tracking and reporting
