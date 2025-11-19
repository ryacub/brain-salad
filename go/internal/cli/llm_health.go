package cli

import (
	"fmt"
	"sort"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/llm"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	healthWatch    bool
	healthInterval int
)

func newLLMHealthCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "llm-health",
		Short: "Check health of LLM providers",
		Long: `Run health checks on all registered LLM providers.

Use --watch to continuously monitor provider health.

Examples:
  tm llm-health
  tm llm-health --watch
  tm llm-health --watch --interval 10`,
		RunE: runLLMHealth,
	}

	cmd.Flags().BoolVar(&healthWatch, "watch", false, "Continuously monitor health")
	cmd.Flags().IntVar(&healthInterval, "interval", 30, "Check interval in seconds (with --watch)")

	return cmd
}

func runLLMHealth(cmd *cobra.Command, args []string) error {
	// Create LLM manager
	manager := llm.NewManager(nil)

	if healthWatch {
		return watchHealth(manager, healthInterval)
	}

	// Single health check
	health := manager.HealthCheck()
	displayHealthStatus(health)

	return nil
}

func displayHealthStatus(health map[string]bool) {
	// Sort provider names for consistent output
	providerNames := make([]string, 0, len(health))
	for name := range health {
		providerNames = append(providerNames, name)
	}
	sort.Strings(providerNames)

	// Display header
	fmt.Println("═══════════════════════════════════════════════════════════════")
	if _, err := successColor.Println("🏥 Provider Health Status"); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	// Display each provider
	for _, name := range providerNames {
		healthy := health[name]
		statusText := "UNAVAILABLE"
		statusColor := errorColor
		statusIcon := "✗"

		if healthy {
			statusText = "HEALTHY"
			statusColor = successColor
			statusIcon = "✓"
		}

		fmt.Printf("%-30s %s %s\n", name+":", statusColor.Sprint(statusIcon), statusText)
	}

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
}

func watchHealth(manager *llm.Manager, intervalSec int) error {
	ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
	defer ticker.Stop()

	// Clear screen function
	clearScreen := func() {
		fmt.Print("\033[H\033[2J")
	}

	// Display initial health
	clearScreen()
	displayWatchHealth(manager, intervalSec)

	// Watch loop
	for {
		select {
		case <-ticker.C:
			clearScreen()
			displayWatchHealth(manager, intervalSec)
		}
	}
}

func displayWatchHealth(manager *llm.Manager, intervalSec int) {
	health := manager.HealthCheck()

	// Sort provider names for consistent output
	providerNames := make([]string, 0, len(health))
	for name := range health {
		providerNames = append(providerNames, name)
	}
	sort.Strings(providerNames)

	// Display header
	fmt.Println("═══════════════════════════════════════════════════════════════")
	if _, err := successColor.Printf("🏥 Provider Health Status (refreshing every %ds)\n", intervalSec); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	if _, err := infoColor.Printf("Last check: %s\n", time.Now().Format("15:04:05")); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}
	fmt.Println()

	// Display each provider
	for _, name := range providerNames {
		healthy := health[name]
		statusText := "UNAVAILABLE"
		statusColor := errorColor
		statusIcon := "✗"

		if healthy {
			statusText = "HEALTHY"
			statusColor = successColor
			statusIcon = "✓"
		}

		fmt.Printf("%-30s %s %s\n", name+":", statusColor.Sprint(statusIcon), statusText)
	}

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	if _, err := warningColor.Println("\nPress Ctrl+C to exit"); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}
}
