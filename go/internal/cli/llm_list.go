package cli

import (
	"fmt"
	"sort"

	"github.com/rayyacub/telos-idea-matrix/internal/llm"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	listShowHealth bool
)

func newLLMListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "llm-list",
		Short: "List available LLM providers",
		Long: `Display all registered LLM providers and their availability status.

Use --health to run a health check on all providers.

Examples:
  tm llm-list
  tm llm-list --health`,
		RunE: runLLMList,
	}

	cmd.Flags().BoolVar(&listShowHealth, "health", false, "Run health check on all providers")

	return cmd
}

func runLLMList(cmd *cobra.Command, args []string) error {
	// Create LLM manager
	manager := llm.NewManager(nil)

	// Get health status if requested
	var health map[string]bool
	if listShowHealth {
		if _, err := infoColor.Println("🔍 Running health checks..."); err != nil {
			log.Warn().Err(err).Msg("failed to print message")
		}
		fmt.Println()
		health = manager.HealthCheck()
	}

	// Get all providers
	allProviders := manager.GetAllProviders()
	primaryName := manager.GetPrimaryProviderName()

	// Display header
	fmt.Println("═══════════════════════════════════════════════════════════════")
	if _, err := successColor.Println("🤖 LLM Providers"); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	// Display primary
	if _, err := successColor.Printf("Primary Provider: %s\n", primaryName); err != nil {
		log.Warn().Err(err).Msg("failed to print message")
	}
	fmt.Println()

	// Sort provider names for consistent output
	providerNames := make([]string, 0, len(allProviders))
	for name := range allProviders {
		providerNames = append(providerNames, name)
	}
	sort.Strings(providerNames)

	// Display all providers
	fmt.Println("Available Providers:")
	for _, name := range providerNames {
		provider := allProviders[name]
		marker := "  "
		if name == primaryName {
			marker = "* "
		}

		isAvailable := provider.IsAvailable()
		status := ""
		if listShowHealth {
			if isAvailable {
				status = " " + successColor.Sprint("✓")
			} else {
				status = " " + errorColor.Sprint("✗")
			}
		}

		fmt.Printf("%s%s%s\n", marker, name, status)
	}

	if listShowHealth {
		fmt.Println()
		fmt.Println("───────────────────────────────────────────────────────────────")
		fmt.Println("Health Status:")
		fmt.Println()

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

			fmt.Printf("  %-20s %s %s\n", name+":", statusColor.Sprint(statusIcon), statusText)
		}
	}

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")

	return nil
}
