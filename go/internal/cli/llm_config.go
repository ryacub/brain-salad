package cli

import (
	"fmt"
	"os"

	"github.com/rayyacub/telos-idea-matrix/internal/llm"
	"github.com/spf13/cobra"
)

func newLLMConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "llm-config",
		Short: "Configure LLM provider settings",
		Long: `Configure LLM provider settings including default provider and environment variables.

Subcommands:
  set-default  Set the default LLM provider
  show         Show current provider configuration

Examples:
  tm llm-config set-default ollama
  tm llm-config show`,
	}

	// Add subcommands
	cmd.AddCommand(newLLMConfigSetDefaultCommand())
	cmd.AddCommand(newLLMConfigShowCommand())

	return cmd
}

func newLLMConfigSetDefaultCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set-default <provider>",
		Short: "Set the default LLM provider",
		Long: `Set the default LLM provider to use for analysis.

Available providers:
  • ollama      - Local Ollama instance
  • rule_based  - Rule-based scoring engine

Examples:
  tm llm-config set-default ollama
  tm llm-config set-default rule_based`,
		Args: cobra.ExactArgs(1),
		RunE: runLLMConfigSetDefault,
	}
}

func newLLMConfigShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current provider configuration",
		Long: `Display current provider configuration including primary provider
and environment variable settings.

Examples:
  tm llm-config show`,
		RunE: runLLMConfigShow,
	}
}

func runLLMConfigSetDefault(cmd *cobra.Command, args []string) error {
	providerName := args[0]

	// Create LLM manager
	manager := llm.NewManager(nil)

	// Set primary provider
	if err := manager.SetPrimaryProvider(providerName); err != nil {
		return fmt.Errorf("failed to set default provider: %w", err)
	}

	successColor.Printf("✓ Default provider set to: %s\n", providerName)
	fmt.Println()
	warningColor.Println("⚠️  Note: This setting is only active for the current session.")
	fmt.Println("   To persist this setting, set environment variables in your shell profile.")

	return nil
}

func runLLMConfigShow(cmd *cobra.Command, args []string) error {
	// Create LLM manager
	manager := llm.NewManager(nil)
	primary := manager.GetPrimaryProviderName()

	// Display configuration
	fmt.Println("═══════════════════════════════════════════════════════════════")
	successColor.Println("⚙️  Current LLM Configuration")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	successColor.Printf("Primary Provider: %s\n", primary)
	fmt.Println()

	// Show environment variables
	fmt.Println("Environment Variables:")
	showEnvVars()

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")

	return nil
}

func showEnvVars() {
	envVars := []struct {
		name        string
		description string
		maskValue   bool
	}{
		{"OLLAMA_BASE_URL", "Ollama server URL", false},
		{"OLLAMA_MODEL", "Ollama model name", false},
		{"OPENAI_API_KEY", "OpenAI API key", true},
		{"OPENAI_MODEL", "OpenAI model name", false},
		{"ANTHROPIC_API_KEY", "Anthropic API key", true},
		{"CLAUDE_MODEL", "Claude model name", false},
		{"CUSTOM_LLM_NAME", "Custom LLM provider name", false},
		{"CUSTOM_LLM_ENDPOINT", "Custom LLM endpoint URL", false},
	}

	fmt.Println()
	for _, env := range envVars {
		value := os.Getenv(env.name)
		displayValue := "(not set)"

		if value != "" {
			if env.maskValue {
				displayValue = maskAPIKey(value)
			} else {
				displayValue = value
			}
			infoColor.Printf("  %-25s %s\n", env.name+":", displayValue)
		} else {
			fmt.Printf("  %-25s %s\n", env.name+":", displayValue)
		}
	}
}

func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + "..." + key[len(key)-4:]
}
