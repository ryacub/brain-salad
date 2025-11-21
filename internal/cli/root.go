package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/rayyacub/telos-idea-matrix/internal/cli/analytics"
	"github.com/rayyacub/telos-idea-matrix/internal/cli/bulk"
	"github.com/rayyacub/telos-idea-matrix/internal/cli/dump"
	"github.com/rayyacub/telos-idea-matrix/internal/database"
	"github.com/rayyacub/telos-idea-matrix/internal/llm"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/rayyacub/telos-idea-matrix/internal/patterns"
	"github.com/rayyacub/telos-idea-matrix/internal/scoring"
	"github.com/rayyacub/telos-idea-matrix/internal/telos"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CLIContext holds shared dependencies for all commands
type CLIContext struct {
	Repository *database.Repository
	Engine     *scoring.Engine
	Detector   *patterns.Detector
	Telos      *models.Telos
	LLMManager *llm.Manager
	DBPath     string
	TelosPath  string
}

var (
	ctx       *CLIContext
	dbPath    string
	telosPath string
	rootCmd   *cobra.Command

	// Color definitions
	successColor = color.New(color.FgGreen, color.Bold)
	errorColor   = color.New(color.FgRed, color.Bold)
	infoColor    = color.New(color.FgCyan)
	warningColor = color.New(color.FgYellow)
)

func init() {
	rootCmd = &cobra.Command{
		Use:   "tm",
		Short: "Telos Matrix - An AI-powered idea management system",
		Long: `Telos Matrix helps you capture, analyze, and manage ideas aligned with your goals.
It scores ideas based on your mission, anti-patterns, and strategic fit to help
you focus on what truly matters.`,
		PersistentPreRunE: initializeCLI,
	}

	// Global flags
	homeDir, _ := os.UserHomeDir()
	defaultTelosPath := filepath.Join(homeDir, ".telos", "telos.md")
	defaultDBPath := filepath.Join(homeDir, ".telos", "ideas.db")

	rootCmd.PersistentFlags().StringVar(&dbPath, "db", defaultDBPath, "Path to ideas database")
	rootCmd.PersistentFlags().StringVar(&telosPath, "telos", defaultTelosPath, "Path to telos.md file")

	// Add subcommands
	dumpCmd := dump.NewDumpCommand(getDumpContext)
	dumpCmd.AddCommand(newBatchDumpCommand())
	rootCmd.AddCommand(dumpCmd)
	rootCmd.AddCommand(newScoreCommand())
	rootCmd.AddCommand(newAnalyzeCommand())
	rootCmd.AddCommand(newReviewCommand())
	rootCmd.AddCommand(newPruneCommand())
	rootCmd.AddCommand(analytics.NewAnalyticsCommand(getAnalyticsContext))
	rootCmd.AddCommand(newLinkCommand())
	rootCmd.AddCommand(newHealthCommand())
	rootCmd.AddCommand(bulk.NewBulkCommand(getBulkContext))

	// LLM commands (new hierarchical structure)
	rootCmd.AddCommand(NewLLMCommand())

	// Legacy LLM commands (flat structure - may be deprecated)
	rootCmd.AddCommand(newAnalyzeLLMCommand())
	rootCmd.AddCommand(newLLMListCommand())
	rootCmd.AddCommand(newLLMConfigCommand())
	rootCmd.AddCommand(newLLMHealthCommand())

	// Utility commands
	rootCmd.AddCommand(newInitCommand())
	rootCmd.AddCommand(newCompletionCommand())
}

// initializeCLI sets up the shared context for all commands
func initializeCLI(cmd *cobra.Command, args []string) error {
	// Skip initialization if context is already set (e.g., by tests)
	if ctx != nil {
		return nil
	}

	// Create .telos directory if it doesn't exist
	telosDir := filepath.Dir(telosPath)
	if err := os.MkdirAll(telosDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check if telos.md exists
	if _, err := os.Stat(telosPath); os.IsNotExist(err) {
		if _, printErr := warningColor.Fprintf(os.Stderr, "⚠️  Telos file not found at %s\n", telosPath); printErr != nil {
			log.Warn().Err(printErr).Msg("failed to print warning")
		}
		fmt.Fprintf(os.Stderr, "Please create a telos.md file with your goals, strategies, and stack.\n")
		fmt.Fprintf(os.Stderr, "Run 'tm init' to create a template or see documentation for format.\n\n")
		return fmt.Errorf("telos file not found")
	}

	// Parse telos.md
	parser := telos.NewParser()
	telosData, err := parser.ParseFile(telosPath)
	if err != nil {
		return fmt.Errorf("failed to parse telos.md: %w", err)
	}

	// Initialize database
	repo, err := database.NewRepository(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Create scoring engine and pattern detector
	engine := scoring.NewEngine(telosData)
	detector := patterns.NewDetector(telosData)

	// Initialize LLM Manager
	llmConfig := llm.DefaultManagerConfig()
	llmManager := llm.NewManager(llmConfig)

	// Store in shared context
	ctx = &CLIContext{
		Repository: repo,
		Engine:     engine,
		Detector:   detector,
		Telos:      telosData,
		LLMManager: llmManager,
		DBPath:     dbPath,
		TelosPath:  telosPath,
	}

	return nil
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// resetCommandFlags recursively resets all flags for a command and its subcommands
func resetCommandFlags(cmd *cobra.Command) {
	// Reset local flags
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		flag.Changed = false
		_ = flag.Value.Set(flag.DefValue)
	})

	// Reset persistent flags
	cmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		flag.Changed = false
		_ = flag.Value.Set(flag.DefValue)
	})

	// Reset all subcommands recursively
	for _, subCmd := range cmd.Commands() {
		resetCommandFlags(subCmd)
	}
}

// GetRootCmd returns the root command for testing
func GetRootCmd() *cobra.Command {
	// Reset command state for testing
	// This ensures flags are re-parsed fresh for each test
	if rootCmd != nil {
		rootCmd.SilenceUsage = false
		rootCmd.SilenceErrors = false

		// Reset all flags recursively for root and all subcommands
		resetCommandFlags(rootCmd)
	}
	return rootCmd
}

// SetContext allows setting a custom context for testing
func SetContext(c *CLIContext) {
	ctx = c
}

// ClearContext clears the global context (used for test cleanup)
func ClearContext() {
	ctx = nil
	// Also reset the global flag variables
	dbPath = ""
	telosPath = ""
}

// getDumpContext converts CLIContext to dump.CLIContext
func getDumpContext() *dump.CLIContext {
	if ctx == nil {
		return nil
	}
	return &dump.CLIContext{
		Repository: ctx.Repository,
		Engine:     ctx.Engine,
		Detector:   ctx.Detector,
		Telos:      ctx.Telos,
		LLMManager: ctx.LLMManager,
	}
}

// getAnalyticsContext converts CLIContext to analytics.CLIContext
func getAnalyticsContext() *analytics.CLIContext {
	if ctx == nil {
		return nil
	}
	return &analytics.CLIContext{
		Repository: ctx.Repository,
		DBPath:     ctx.DBPath,
	}
}

// getBulkContext converts CLIContext to bulk.CLIContext
func getBulkContext() *bulk.CLIContext {
	if ctx == nil {
		return nil
	}
	return &bulk.CLIContext{
		Repository: ctx.Repository,
		Telos:      ctx.Telos,
		LLMManager: ctx.LLMManager,
	}
}
