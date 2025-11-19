// Package main demonstrates the usage of the LLM provider manager with various examples.
package main

import (
	"fmt"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/llm"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
)

func main() {
	fmt.Println("=== Provider Manager Example ===")

	// Example 1: Basic Usage with Default Configuration
	basicExample()

	// Example 2: Custom Configuration
	customConfigExample()

	// Example 3: Health Monitoring
	healthMonitoringExample()

	// Example 4: Statistics Tracking
	statisticsExample()

	// Example 5: Fallback Behavior
	fallbackExample()

	// Example 6: Priority Management
	priorityExample()
}

func basicExample() {
	fmt.Println("--- Example 1: Basic Usage ---")

	// Create manager with default configuration
	config := llm.DefaultManagerConfig()
	manager := llm.NewManager(config)

	// Create a simple telos
	telos := createSampleTelos()

	// Perform analysis
	result, err := manager.Analyze(llm.AnalysisRequest{
		IdeaContent: "Build an AI-powered task manager that learns from user behavior",
		Telos:       telos,
	})

	if err != nil {
		fmt.Printf("❌ Analysis failed: %v\n", err)
	} else {
		fmt.Printf("✓ Analysis complete!\n")
		fmt.Printf("  Provider: %s\n", result.Provider)
		fmt.Printf("  Score: %.2f/10\n", result.FinalScore)
		fmt.Printf("  Duration: %s\n", result.Duration)
	}

	fmt.Println()
}

func customConfigExample() {
	fmt.Println("--- Example 2: Custom Configuration ---")

	// Create custom configuration
	config := &llm.ManagerConfig{
		DefaultProvider:     "ollama",
		FallbackEnabled:     true,
		HealthCheckInterval: 30 * time.Second,
		Priority:            []string{"ollama", "rule_based"},
		ProviderConfig: llm.ProviderConfig{
			OllamaBaseURL: "http://localhost:11434",
			OllamaModel:   "llama2",
			OllamaTimeout: 30,
		},
	}

	manager := llm.NewManager(config)

	// Get current configuration
	primary := manager.GetPrimaryProvider()
	fmt.Printf("Primary provider: %s\n", primary.Name())
	fmt.Printf("Fallback enabled: %v\n", manager.IsFallbackEnabled())

	// Get all providers
	providers := manager.GetProviders()
	fmt.Printf("Registered providers: %d\n", len(providers))
	for _, p := range providers {
		fmt.Printf("  - %s (available: %v)\n", p.Name(), p.IsAvailable())
	}

	fmt.Println()
}

func healthMonitoringExample() {
	fmt.Println("--- Example 3: Health Monitoring ---")

	config := llm.DefaultManagerConfig()
	manager := llm.NewManager(config)

	// Perform manual health check
	fmt.Println("Performing health check...")
	status := manager.HealthCheck()

	for name, available := range status {
		if available {
			fmt.Printf("✓ %s: healthy\n", name)
		} else {
			fmt.Printf("✗ %s: unavailable\n", name)
		}
	}

	// Get detailed health status for a specific provider
	available, lastCheck, err := manager.GetHealthStatus("ollama")
	if err == nil {
		fmt.Printf("\nOllama detailed status:\n")
		fmt.Printf("  Available: %v\n", available)
		fmt.Printf("  Last checked: %s ago\n", time.Since(lastCheck).Round(time.Second))
	}

	// Start periodic health checks (would run in background)
	// stopCh := make(chan struct{})
	// go manager.StartPeriodicHealthCheck(stopCh)
	// defer close(stopCh)

	fmt.Println()
}

func statisticsExample() {
	fmt.Println("--- Example 4: Statistics Tracking ---")

	config := llm.DefaultManagerConfig()
	manager := llm.NewManager(config)
	telos := createSampleTelos()

	// Perform multiple analyses
	ideas := []string{
		"Build a code review automation tool",
		"Create a developer productivity dashboard",
		"Implement an AI pair programming assistant",
	}

	fmt.Printf("Analyzing %d ideas...\n", len(ideas))
	for _, idea := range ideas {
		_, err := manager.Analyze(llm.AnalysisRequest{
			IdeaContent: idea,
			Telos:       telos,
		})
		if err != nil {
			fmt.Printf("  ✗ Failed: %s\n", idea[:30]+"...")
		}
	}

	// Get statistics
	fmt.Println("\nProvider Statistics:")
	stats := manager.GetStats()
	for _, s := range stats {
		if s.TotalRequests > 0 {
			successRate := float64(s.SuccessCount) / float64(s.TotalRequests) * 100
			fmt.Printf("\n%s:\n", s.Name)
			fmt.Printf("  Total Requests: %d\n", s.TotalRequests)
			fmt.Printf("  Success Rate: %.1f%%\n", successRate)
			fmt.Printf("  Avg Latency: %s\n", s.AverageLatency.Round(time.Millisecond))
			if !s.LastUsed.IsZero() {
				fmt.Printf("  Last Used: %s ago\n", time.Since(s.LastUsed).Round(time.Second))
			}
		}
	}

	fmt.Println()
}

func fallbackExample() {
	fmt.Println("--- Example 5: Fallback Behavior ---")

	// Create manager with fallback enabled
	config := &llm.ManagerConfig{
		FallbackEnabled: true,
		Priority:        []string{"ollama", "rule_based"},
		ProviderConfig:  llm.DefaultProviderConfig(),
	}
	manager := llm.NewManager(config)

	telos := createSampleTelos()

	fmt.Println("Attempting analysis with fallback enabled...")
	result, err := manager.Analyze(llm.AnalysisRequest{
		IdeaContent: "Build a real-time collaboration platform",
		Telos:       telos,
	})

	if err != nil {
		fmt.Printf("❌ All providers failed: %v\n", err)
	} else {
		fmt.Printf("✓ Analysis successful via %s\n", result.Provider)
	}

	// Now disable fallback
	fmt.Println("\nDisabling fallback and trying again...")
	manager.EnableFallback(false)

	// If primary is unavailable, this will fail
	_, err = manager.Analyze(llm.AnalysisRequest{
		IdeaContent: "Build a mobile app for project management",
		Telos:       telos,
	})

	if err != nil {
		fmt.Printf("Expected failure with fallback disabled: %v\n", err)
	}

	// Re-enable fallback
	manager.EnableFallback(true)
	fmt.Println("Fallback re-enabled")

	fmt.Println()
}

func priorityExample() {
	fmt.Println("--- Example 6: Priority Management ---")

	// Initial priority order
	config := &llm.ManagerConfig{
		FallbackEnabled: true,
		Priority:        []string{"ollama", "rule_based"},
		ProviderConfig:  llm.DefaultProviderConfig(),
	}
	manager := llm.NewManager(config)

	fmt.Println("Initial priority order:")
	printProviderOrder(manager)

	// Update priority order
	fmt.Println("\nUpdating priority order...")
	newConfig := &llm.ManagerConfig{
		DefaultProvider: "rule_based",
		FallbackEnabled: true,
		Priority:        []string{"rule_based", "ollama"},
		ProviderConfig:  llm.DefaultProviderConfig(),
	}

	err := manager.LoadConfig(newConfig)
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
	} else {
		fmt.Println("✓ Configuration updated")
		printProviderOrder(manager)
	}

	fmt.Println()
}

// Helper functions

func createSampleTelos() *models.Telos {
	return &models.Telos{
		Goals: []models.Goal{
			{
				ID:          "g1",
				Description: "Build products that solve real problems",
				Priority:    1,
			},
			{
				ID:          "g2",
				Description: "Maintain work-life balance",
				Priority:    2,
			},
		},
		Missions: []models.Mission{
			{
				ID:          "m1",
				Description: "Create AI-powered developer tools",
			},
		},
		Challenges: []models.Challenge{
			{
				ID:          "c1",
				Description: "Avoid context switching",
			},
			{
				ID:          "c2",
				Description: "Ship frequently",
			},
		},
		Strategies: []models.Strategy{
			{
				ID:          "s1",
				Description: "Use Go and modern AI APIs",
			},
		},
	}
}

func printProviderOrder(manager *llm.Manager) {
	providers := manager.GetProviders()
	primary := manager.GetPrimaryProvider()

	for i, p := range providers {
		marker := "  "
		if p.Name() == primary.Name() {
			marker = "→ "
		}
		fmt.Printf("%s%d. %s\n", marker, i+1, p.Name())
	}
}
