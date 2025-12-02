package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ryacub/telos-idea-matrix/internal/cli/wizard"
	"github.com/ryacub/telos-idea-matrix/internal/profile"
)

var initAdvanced bool

func newInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Set up Brain Salad with an interactive discovery wizard",
		Long: `Initialize Brain Salad for first-time use.

This command will:
  - Run an interactive discovery wizard to learn your preferences
  - Create a personalized scoring profile
  - Set up the database for storing ideas

Use --advanced to create a telos.md file for power-user configuration.
`,
		RunE: runInit,
	}

	cmd.Flags().BoolVar(&initAdvanced, "advanced", false, "Create telos.md for advanced configuration")

	// Skip initialization for init command (doesn't need existing telos.md)
	// Override the parent's PersistentPreRunE with a no-op function
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return nil
	}

	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	// Check for existing profile
	profilePath, err := profile.DefaultPath()
	if err != nil {
		return fmt.Errorf("failed to determine profile path: %w", err)
	}

	// Check for existing telos.md (legacy)
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	legacyTelosPath := filepath.Join(home, ".telos", "telos.md")

	// Handle existing configurations
	if profile.Exists(profilePath) {
		runner := wizard.NewRunner()
		if !runner.Confirm("A profile already exists. Do you want to start over?") {
			fmt.Println("Keeping existing profile.")
			return nil
		}
	}

	// If advanced mode, use legacy telos.md setup
	if initAdvanced {
		return runAdvancedInit()
	}

	// Check for legacy telos.md and offer migration
	if _, err := os.Stat(legacyTelosPath); err == nil {
		fmt.Println()
		fmt.Println("Found existing telos.md configuration.")
		fmt.Println()
		runner := wizard.NewRunner()
		if runner.Confirm("Would you like to create a new profile instead? (recommended)") {
			return runWizardInit(profilePath)
		}
		fmt.Println("Keeping telos.md for advanced mode. Run 'tm init --advanced' to modify it.")
		return nil
	}

	// Run the discovery wizard
	return runWizardInit(profilePath)
}

// runWizardInit runs the interactive discovery wizard.
func runWizardInit(profilePath string) error {
	runner := wizard.NewRunner()

	// Run the wizard
	answers, err := runner.Run()
	if err != nil {
		runner.PrintError(fmt.Sprintf("Wizard failed: %v", err))
		return err
	}

	// Map answers to profile
	p := wizard.MapAnswersToProfile(answers)

	// Generate and display what we learned
	summary := wizard.GenerateSummary(p)
	runner.PrintSummary(summary)

	// Show profile preview with visual weights
	runner.PrintProfilePreview(p.Priorities, p.Goals, p.Avoid)

	// Confirm before saving
	if !runner.ConfirmSave() {
		fmt.Println()
		fmt.Println("Profile not saved. Run 'tm init' to try again.")
		return nil
	}

	// Save the profile
	if err := profile.Save(p, profilePath); err != nil {
		runner.PrintError(fmt.Sprintf("Failed to save profile: %v", err))
		return err
	}

	runner.PrintSuccess(profilePath)

	// Create data directory for database
	dataDir, err := profile.DefaultDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Show next steps
	fmt.Println("Next steps:")
	fmt.Println("  1. Score your first idea: tm score \"Your idea here\"")
	fmt.Println("  2. Save an idea: tm dump \"Your idea here\"")
	fmt.Println("  3. Review saved ideas: tm review")
	fmt.Println()

	return nil
}

// runAdvancedInit creates the legacy telos.md setup for power users.
func runAdvancedInit() error {
	fmt.Println("Setting up advanced mode with telos.md...")
	fmt.Println()

	// 1. Create data directory
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		dataDir = filepath.Join(home, ".telos")
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}
	fmt.Printf("✓ Created data directory: %s\n", dataDir)

	// 2. Create telos.md template
	telosPath := os.Getenv("TELOS_PATH")
	if telosPath == "" {
		telosPath = filepath.Join(dataDir, "telos.md")
	}

	if _, err := os.Stat(telosPath); os.IsNotExist(err) {
		if err := createTelosTemplate(telosPath); err != nil {
			return fmt.Errorf("failed to create telos.md: %w", err)
		}
		fmt.Printf("✓ Created telos template: %s\n", telosPath)
	} else {
		fmt.Printf("⚠ Telos file already exists: %s\n", telosPath)
	}

	// 3. Initialize database
	dbPath := filepath.Join(dataDir, "ideas.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Printf("✓ Database will be created at: %s\n", dbPath)
	} else {
		fmt.Printf("⚠ Database already exists: %s\n", dbPath)
	}

	// 4. Create .env template
	envPath := filepath.Join(dataDir, ".env.example")
	if err := createEnvTemplate(envPath); err != nil {
		return fmt.Errorf("failed to create .env template: %w", err)
	}
	fmt.Printf("✓ Created environment template: %s\n", envPath)

	// 5. Show next steps
	fmt.Println()
	fmt.Println("Advanced mode initialized!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. Edit your mission and goals: %s\n", telosPath)
	fmt.Printf("  2. (Optional) Configure LLM providers: %s → .env\n", envPath)
	fmt.Println("  3. Capture your first idea: tm dump \"Your first idea\"")
	fmt.Println("  4. Review your ideas: tm review")
	fmt.Println()
	fmt.Println("Run 'tm --help' to see all available commands")

	return nil
}

func createTelosTemplate(path string) error {
	template := `# My Telos (Ultimate Purpose)

## Mission
What is your overarching mission or purpose? What drives you?

Example: "Build innovative AI tools that help people overcome decision paralysis"

## Core Challenges (What You're Fighting Against)
List the main obstacles or anti-patterns you want to avoid:

- Perfectionism that leads to inaction
- Scope creep that derails projects
- Analysis paralysis
- Shiny object syndrome

## Strategic Goals
What are your current strategic priorities?

1. **Goal 1**: Launch MVP by Q2
2. **Goal 2**: Build sustainable audience
3. **Goal 3**: Achieve work-life balance

---

Edit this file to reflect your true north. The scoring engine will use this to evaluate your ideas.
`

	return os.WriteFile(path, []byte(template), 0644)
}

func createEnvTemplate(path string) error {
	template := `# Telos Idea Matrix Configuration

# Database
DB_PATH=$HOME/.telos/ideas.db

# Telos definition file
TELOS_PATH=$HOME/.telos/telos.md

# LLM Providers (optional)
# Uncomment and configure the provider you want to use:

# Ollama (local)
# OLLAMA_BASE_URL=http://localhost:11434

# OpenAI
# OPENAI_API_KEY=sk-...
# OPENAI_MODEL=gpt-4

# Anthropic Claude
# ANTHROPIC_API_KEY=sk-ant-...
# ANTHROPIC_MODEL=claude-3-sonnet-20240229

# Custom HTTP endpoint
# CUSTOM_LLM_URL=https://your-llm-api.com/analyze
# CUSTOM_LLM_API_KEY=your-key

# Web Server (if using API)
# PORT=8080
# HOST=0.0.0.0

# Authentication (if using API)
# AUTH_ENABLED=true
# AUTH_API_KEYS="key1:description1,key2:description2"

# Feature Flags
# FEATURE_LLM_ANALYSIS=true
# FEATURE_WEB_UI=true
# FEATURE_METRICS=false

# Logging
# DEBUG=false
`

	return os.WriteFile(path, []byte(template), 0644)
}
