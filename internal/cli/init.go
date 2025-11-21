package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize Telos Idea Matrix for first-time use",
		Long: `Initialize your Telos Idea Matrix workspace.

This command will:
  - Create a data directory
  - Generate a template telos.md file
  - Initialize the database
  - Set up default configuration
`,
		RunE: runInit,
	}

	// Skip initialization for init command (doesn't need existing telos.md)
	// Override the parent's PersistentPreRunE with a no-op function
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return nil
	}

	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	fmt.Println("🚀 Initializing Telos Idea Matrix...")
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
		// Database will be created on first run
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
	fmt.Println("🎉 Initialization complete!")
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
