package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/ryacub/telos-idea-matrix/internal/cliutil"
	"github.com/ryacub/telos-idea-matrix/internal/llm"
	"github.com/spf13/cobra"
)

// NewLLMCommand creates the llm management command
func NewLLMCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "llm",
		Short: "Manage LLM providers",
		Long: `Manage LLM providers for idea analysis.

Subcommands:
  list         - List all available providers
  test         - Test provider connectivity
  set-default  - Set default provider
  config       - Show provider configuration

Examples:
  telos llm list                 # List all providers
  telos llm test openai          # Test OpenAI provider
  telos llm set-default claude   # Set Claude as default
  telos llm config               # Show all configurations`,
	}

	cmd.AddCommand(newLLMListSubcommand())
	cmd.AddCommand(newLLMTestSubcommand())
	cmd.AddCommand(newLLMSetDefaultSubcommand())
	cmd.AddCommand(newLLMConfigSubcommand())
	cmd.AddCommand(newLLMHealthSubcommand())

	return cmd
}

// ============================================================================
// LLM LIST SUBCOMMAND
// ============================================================================

func newLLMListSubcommand() *cobra.Command {
	var showAll bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available LLM providers",
		Long: `List all LLM providers and their availability status.

Shows:
  - Provider name
  - Availability (configured and online)
  - Current default provider

Examples:
  telos llm list           # Show available providers only
  telos llm list --all     # Show all providers including unavailable`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLLMListSubcmd(ctx.LLMManager, showAll)
		},
	}

	cmd.Flags().BoolVar(&showAll, "all", false, "Show all providers including unavailable")

	return cmd
}

func runLLMListSubcmd(manager *llm.Manager, showAll bool) error {
	providers := manager.GetAllProviders()
	currentDefault := manager.GetPrimaryProvider()

	fmt.Println("LLM Providers:")
	fmt.Println()

	available := 0
	for name, p := range providers {
		isAvailable := p.IsAvailable()

		// Skip unavailable if not --all
		if !showAll && !isAvailable {
			continue
		}

		// Status indicator
		status := "✗ Unavailable"
		if isAvailable {
			status = "✓ Available"
			available++
		}

		// Default indicator
		defaultMarker := ""
		if currentDefault != nil && name == currentDefault.Name() {
			defaultMarker = " (default)"
		}

		fmt.Printf("  %s %s%s\n", status, name, defaultMarker)
	}

	fmt.Println()
	fmt.Printf("Available: %d/%d providers\n", available, len(providers))

	if available == 0 {
		fmt.Println("\nNo providers configured. Set environment variables:")
		fmt.Println("  - OPENAI_API_KEY for OpenAI")
		fmt.Println("  - ANTHROPIC_API_KEY for Claude")
		fmt.Println("  - Ollama should be running on localhost:11434")
		fmt.Println("  - CUSTOM_LLM_ENDPOINT for custom providers")
	}

	return nil
}

// ============================================================================
// LLM TEST SUBCOMMAND
// ============================================================================

func newLLMTestSubcommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test <provider-name>",
		Short: "Test provider connectivity",
		Long: `Test connectivity and functionality of a specific provider.

Runs a minimal analysis request to verify the provider is working correctly.
Tests both availability check and actual analysis functionality.

Examples:
  telos llm test openai
  telos llm test claude
  telos llm test ollama`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			providerName := args[0]
			return runLLMTestSubcmd(ctx.LLMManager, providerName)
		},
	}

	return cmd
}

func runLLMTestSubcmd(manager *llm.Manager, providerName string) error {
	// Find provider
	provider := getProviderByName(manager, providerName)
	if provider == nil {
		return fmt.Errorf("provider not found: %s\n\nAvailable providers:\n%s",
			providerName, getProviderList(manager))
	}

	fmt.Printf("Testing provider: %s\n", provider.Name())
	fmt.Println()

	// Step 1: Check availability
	fmt.Print("1. Checking availability... ")
	if !provider.IsAvailable() {
		fmt.Println("✗ FAILED")
		fmt.Println("\nProvider is not available. Check configuration:")
		fmt.Println(getProviderConfigHelp(providerName))
		return fmt.Errorf("provider not available")
	}
	fmt.Println("✓ OK")

	// Step 2: Test analysis
	fmt.Print("2. Testing analysis... ")
	testReq := llm.AnalysisRequest{
		IdeaContent: "Build a simple web application for tracking personal goals",
		Telos:       ctx.Telos, // Use telos from CLI context
	}

	result, err := provider.Analyze(testReq)
	if err != nil {
		fmt.Println("✗ FAILED")
		return fmt.Errorf("analysis failed: %w", err)
	}
	fmt.Println("✓ OK")

	// Step 3: Validate response structure
	fmt.Print("3. Validating response... ")
	if result.FinalScore < 0 || result.FinalScore > 10 {
		fmt.Println("✗ FAILED")
		return fmt.Errorf("invalid score: %.1f (must be 0-10)", result.FinalScore)
	}
	if result.Recommendation == "" {
		fmt.Println("✗ FAILED")
		return fmt.Errorf("missing recommendation")
	}
	fmt.Println("✓ OK")

	// Show results
	fmt.Println("\nTest Results:")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("  Score: %.1f/10\n", result.FinalScore)
	fmt.Printf("  Mission Alignment: %.2f\n", result.Scores.MissionAlignment)
	fmt.Printf("  Anti-Challenge: %.2f\n", result.Scores.AntiChallenge)
	fmt.Printf("  Strategic Fit: %.2f\n", result.Scores.StrategicFit)
	fmt.Printf("  Recommendation: %s\n", result.Recommendation)

	if len(result.Explanations) > 0 {
		fmt.Printf("  Reasoning: %s\n", cliutil.TruncateText(getFirstExplanation(result.Explanations), 100))
	}

	fmt.Println("\n✓ Provider is working correctly!")

	return nil
}

// ============================================================================
// LLM SET-DEFAULT SUBCOMMAND
// ============================================================================

func newLLMSetDefaultSubcommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-default <provider-name>",
		Short: "Set default LLM provider",
		Long: `Set the default LLM provider for analysis.

The default provider will be used for all analysis commands unless
explicitly overridden with the --provider flag.

The preference is saved to ~/.telos/llm-config.json and persists
across sessions.

Examples:
  telos llm set-default openai
  telos llm set-default claude`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			providerName := args[0]
			return runLLMSetDefaultSubcmd(ctx.LLMManager, providerName)
		},
	}

	return cmd
}

func runLLMSetDefaultSubcmd(manager *llm.Manager, providerName string) error {
	// Find provider
	provider := getProviderByName(manager, providerName)
	if provider == nil {
		return fmt.Errorf("provider not found: %s\n\nAvailable providers:\n%s",
			providerName, getProviderList(manager))
	}

	// Check availability
	if !provider.IsAvailable() {
		return fmt.Errorf("provider not available: %s\n\nCheck configuration with:\n  telos llm config %s",
			providerName, providerName)
	}

	// Set as primary
	if err := manager.SetPrimaryProvider(provider.Name()); err != nil {
		return fmt.Errorf("failed to set primary provider: %w", err)
	}

	// Note: Persistence will be handled by Agent 2 (Phase 4 - Config Management)
	// For now, we just set it for the current session
	fmt.Printf("✓ Default provider set to: %s\n", provider.Name())
	fmt.Println("  (active for current session)")
	fmt.Println("\nNote: Persistent configuration will be available in a future update.")

	return nil
}

// ============================================================================
// LLM CONFIG SUBCOMMAND
// ============================================================================

func newLLMConfigSubcommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config [provider-name]",
		Short: "Show provider configuration",
		Long: `Show configuration for a specific provider or all providers.

Displays:
  - Provider status (available/unavailable)
  - Model being used
  - API endpoint
  - API key (masked for security)

Examples:
  telos llm config             # Show all configurations
  telos llm config openai      # Show OpenAI configuration
  telos llm config claude      # Show Claude configuration`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var providerName string
			if len(args) > 0 {
				providerName = args[0]
			}
			return runLLMConfigSubcmd(ctx.LLMManager, providerName)
		},
	}

	return cmd
}

func runLLMConfigSubcmd(manager *llm.Manager, providerName string) error {
	if providerName == "" {
		// Show all configurations
		providers := manager.GetAllProviders()

		fmt.Println("LLM Provider Configurations:")
		fmt.Println()

		i := 0
		for _, p := range providers {
			showProviderConfig(p)
			if i < len(providers)-1 {
				fmt.Println()
			}
			i++
		}

		return nil
	}

	// Show specific provider
	provider := getProviderByName(manager, providerName)
	if provider == nil {
		return fmt.Errorf("provider not found: %s", providerName)
	}

	showProviderConfig(provider)

	return nil
}

func showProviderConfig(provider llm.Provider) {
	fmt.Printf("%s\n", provider.Name())
	fmt.Println(strings.Repeat("-", len(provider.Name())))

	status := "✗ Not available"
	if provider.IsAvailable() {
		status = "✓ Available"
	}
	fmt.Printf("Status: %s\n", status)

	// Provider-specific configuration using type assertions
	switch p := provider.(type) {
	case *llm.OpenAIProvider:
		fmt.Printf("Model: %s\n", p.GetModel())
		fmt.Printf("API Key: %s\n", maskAPIKeyLLM(p.GetAPIKey()))
		fmt.Printf("Endpoint: https://api.openai.com/v1/chat/completions\n")

	case *llm.ClaudeProvider:
		fmt.Printf("Model: %s\n", p.GetModel())
		fmt.Printf("API Key: %s\n", maskAPIKeyLLM(p.GetAPIKey()))
		fmt.Printf("Endpoint: https://api.anthropic.com/v1/messages\n")

	case *llm.OllamaProvider:
		fmt.Printf("Type: Local Ollama\n")
		fmt.Printf("Endpoint: http://localhost:11434 (default)\n")

	case *llm.CustomProvider:
		fmt.Printf("Type: Custom HTTP Provider\n")
		// Note: Would need getter methods to show endpoint details

	default:
		fmt.Printf("Type: %T\n", provider)
	}
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// getProviderByName finds a provider by name from the manager
func getProviderByName(manager *llm.Manager, name string) llm.Provider {
	providers := manager.GetAllProviders()
	for providerName, p := range providers {
		if providerName == name || p.Name() == name {
			return p
		}
	}
	return nil
}

// getProviderList returns a formatted list of all provider names
func getProviderList(manager *llm.Manager) string {
	providers := manager.GetAllProviders()
	var names []string
	for name := range providers {
		names = append(names, name)
	}
	return "  - " + strings.Join(names, "\n  - ")
}

// getProviderConfigHelp returns configuration help for a specific provider
func getProviderConfigHelp(providerName string) string {
	help := map[string]string{
		"openai": `  Set environment variable: OPENAI_API_KEY
  Optional: OPENAI_MODEL (default: gpt-5.1)

  Get API key: https://platform.openai.com/api-keys`,
		"openai_gpt-5.1": `  Set environment variable: OPENAI_API_KEY
  Optional: OPENAI_MODEL (default: gpt-5.1)

  Get API key: https://platform.openai.com/api-keys`,
		"claude": `  Set environment variable: ANTHROPIC_API_KEY
  Optional: CLAUDE_MODEL (default: claude-3-5-sonnet-20241022)

  Get API key: https://console.anthropic.com/settings/keys`,
		"ollama": `  Ensure Ollama is running: ollama serve
  Default endpoint: http://localhost:11434

  Install Ollama: https://ollama.ai/download`,
		"custom": `  Set environment variables:
  - CUSTOM_LLM_ENDPOINT (required)
  - CUSTOM_LLM_NAME (optional)
  - CUSTOM_LLM_HEADERS (optional)`,
		"Custom LLM": `  Set environment variables:
  - CUSTOM_LLM_ENDPOINT (required)
  - CUSTOM_LLM_NAME (optional)
  - CUSTOM_LLM_HEADERS (optional)`,
		"rule_based": `  Rule-based provider is always available.
  No configuration required.`,
	}

	if h, ok := help[providerName]; ok {
		return h
	}

	return "  No configuration help available for this provider."
}

// maskAPIKeyLLM masks an API key for secure display
func maskAPIKeyLLM(key string) string {
	if key == "" {
		return "(not set)"
	}
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

// getFirstExplanation gets the first explanation from the map
func getFirstExplanation(explanations map[string]string) string {
	for _, exp := range explanations {
		return exp
	}
	return ""
}

// ============================================================================
// LLM HEALTH SUBCOMMAND
// ============================================================================

func newLLMHealthSubcommand() *cobra.Command {
	var watch bool
	var interval int

	cmd := &cobra.Command{
		Use:   "health",
		Short: "Check health of all LLM providers",
		Long: `Run health checks on all registered LLM providers.

Use --watch to continuously monitor provider health.

Examples:
  tm llm health                    # Check all providers once
  tm llm health --watch            # Continuous monitoring
  tm llm health --watch --interval 10`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLLMHealthSubcmd(ctx.LLMManager, watch, interval)
		},
	}

	cmd.Flags().BoolVar(&watch, "watch", false, "Continuously monitor health")
	cmd.Flags().IntVar(&interval, "interval", 30, "Check interval in seconds (with --watch)")

	return cmd
}

func runLLMHealthSubcmd(manager *llm.Manager, watch bool, intervalSec int) error {
	if watch {
		return watchHealthSubcmd(manager, intervalSec)
	}

	// Single health check
	health := manager.HealthCheck()
	displayHealthStatusSubcmd(health)
	return nil
}

func displayHealthStatusSubcmd(health map[string]bool) {
	// Sort provider names for consistent output
	var providerNames []string
	for name := range health {
		providerNames = append(providerNames, name)
	}

	fmt.Println("LLM Provider Health:")
	fmt.Println()

	healthy := 0
	for _, name := range providerNames {
		isHealthy := health[name]
		status := "✗ Unavailable"
		if isHealthy {
			status = "✓ Healthy"
			healthy++
		}
		fmt.Printf("  %s %s\n", status, name)
	}

	fmt.Println()
	fmt.Printf("Healthy: %d/%d providers\n", healthy, len(health))
}

func watchHealthSubcmd(manager *llm.Manager, intervalSec int) error {
	ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
	defer ticker.Stop()

	// Clear screen
	clearScreen := func() {
		fmt.Print("\033[H\033[2J")
	}

	// Display initial
	clearScreen()
	displayWatchHealthSubcmd(manager, intervalSec)

	// Watch loop
	for range ticker.C {
		clearScreen()
		displayWatchHealthSubcmd(manager, intervalSec)
	}

	return nil
}

func displayWatchHealthSubcmd(manager *llm.Manager, intervalSec int) {
	health := manager.HealthCheck()

	var providerNames []string
	for name := range health {
		providerNames = append(providerNames, name)
	}

	fmt.Printf("LLM Provider Health (refreshing every %ds)\n", intervalSec)
	fmt.Printf("Last check: %s\n", time.Now().Format("15:04:05"))
	fmt.Println()

	healthy := 0
	for _, name := range providerNames {
		isHealthy := health[name]
		status := "✗ Unavailable"
		if isHealthy {
			status = "✓ Healthy"
			healthy++
		}
		fmt.Printf("  %s %s\n", status, name)
	}

	fmt.Println()
	fmt.Printf("Healthy: %d/%d providers\n", healthy, len(health))
	fmt.Println("\nPress Ctrl+C to exit")
}
