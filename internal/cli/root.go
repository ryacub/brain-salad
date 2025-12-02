package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/ryacub/telos-idea-matrix/internal/cli/analytics"
	"github.com/ryacub/telos-idea-matrix/internal/cli/bulk"
	clierrors "github.com/ryacub/telos-idea-matrix/internal/cli/errors"
	"github.com/ryacub/telos-idea-matrix/internal/database"
	"github.com/ryacub/telos-idea-matrix/internal/llm"
	"github.com/ryacub/telos-idea-matrix/internal/models"
	"github.com/ryacub/telos-idea-matrix/internal/patterns"
	"github.com/ryacub/telos-idea-matrix/internal/profile"
	"github.com/ryacub/telos-idea-matrix/internal/scoring"
	"github.com/ryacub/telos-idea-matrix/internal/telos"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ScoringMode indicates which scoring system is active
type ScoringMode string

const (
	// ScoringModeUniversal uses the new profile-based universal scoring
	ScoringModeUniversal ScoringMode = "universal"
	// ScoringModeLegacy uses the traditional telos.md-based scoring
	ScoringModeLegacy ScoringMode = "legacy"
)

// CLIContext holds shared dependencies for all commands
type CLIContext struct {
	Repository      *database.Repository
	Engine          *scoring.Engine          // Legacy scoring engine
	UniversalEngine *scoring.UniversalEngine // Universal scoring engine
	Detector        *patterns.Detector
	Telos           *models.Telos
	Profile         *profile.Profile
	LLMManager      *llm.Manager
	DBPath          string
	TelosPath       string
	ProfilePath     string
	ScoringMode     ScoringMode
}

var (
	ctx       *CLIContext
	dbPath    string
	telosPath string
	rootCmd   *cobra.Command

	// Color definitions
	successColor = color.New(color.FgGreen, color.Bold)
	infoColor    = color.New(color.FgCyan)
	warningColor = color.New(color.FgYellow)
)

func init() {
	rootCmd = &cobra.Command{
		Use:   "tm",
		Short: "Brain Salad - Score ideas against what matters to you",
		Long: `Brain Salad helps you decide which ideas to pursue by scoring them
against your goals and priorities.

Quick Start:
  tm init              Set up with a quick wizard
  tm add "my idea"     Add and score an idea
  tm list              Browse your ideas
  tm show <id>         View idea details

Run 'tm <command> --help' for details on any command.`,
		PersistentPreRunE: initializeCLI,
	}

	// Global flags
	homeDir, _ := os.UserHomeDir()
	defaultTelosPath := filepath.Join(homeDir, ".telos", "telos.md")
	defaultDBPath := filepath.Join(homeDir, ".telos", "ideas.db")

	rootCmd.PersistentFlags().StringVar(&dbPath, "db", defaultDBPath, "Path to ideas database")
	rootCmd.PersistentFlags().StringVar(&telosPath, "telos", defaultTelosPath, "Path to telos.md file")

	// Primary commands (new simplified UX)
	rootCmd.AddCommand(newAddCommand())
	rootCmd.AddCommand(newListCommand())
	rootCmd.AddCommand(newShowCommand())
	rootCmd.AddCommand(newStatusCommand())

	// Setup and config
	rootCmd.AddCommand(newInitCommand())
	rootCmd.AddCommand(newProfileCommand())

	// Management commands
	rootCmd.AddCommand(newPruneCommand())
	rootCmd.AddCommand(newLinkCommand())
	rootCmd.AddCommand(analytics.NewAnalyticsCommand(getAnalyticsContext))
	rootCmd.AddCommand(bulk.NewBulkCommand(getBulkContext))

	// AI/LLM management
	rootCmd.AddCommand(NewLLMCommand())

	// Shell completion
	rootCmd.AddCommand(newCompletionCommand())
}

// initializeCLI sets up the shared context for all commands
func initializeCLI(cmd *cobra.Command, args []string) error {
	// Skip initialization if context is already set (e.g., by tests)
	if ctx != nil {
		return nil
	}

	// Detect which scoring mode to use
	profilePath, _ := profile.DefaultPath()
	hasProfile := profile.Exists(profilePath)
	hasTelosFile := false
	if _, err := os.Stat(telosPath); err == nil {
		hasTelosFile = true
	}

	// Determine scoring mode and initialize accordingly
	if hasProfile {
		return initializeUniversalMode(profilePath)
	} else if hasTelosFile {
		return initializeLegacyMode()
	} else {
		// No configuration found - prompt user to run init
		_, _ = warningColor.Fprintf(os.Stderr, "⚠️  No configuration found.\n")
		fmt.Fprintf(os.Stderr, "Run 'tm init' to set up Brain Salad with a quick wizard.\n\n")
		return clierrors.WrapError(fmt.Errorf("no configuration"), "Initialization failed")
	}
}

// initializeUniversalMode sets up the context with profile-based universal scoring
func initializeUniversalMode(profilePath string) error {
	// Load profile
	p, err := profile.Load(profilePath)
	if err != nil {
		return clierrors.WrapError(err, "Failed to load profile")
	}

	// Determine database path
	profileDir, _ := profile.DefaultDir()
	actualDBPath := dbPath
	if actualDBPath == "" || actualDBPath == filepath.Join(os.Getenv("HOME"), ".telos", "ideas.db") {
		// Use brain-salad directory for new installs
		actualDBPath = filepath.Join(profileDir, "ideas.db")
	}

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(actualDBPath), 0755); err != nil {
		return clierrors.WrapError(err, "Failed to create data directory")
	}

	// Initialize database
	repo, err := database.NewRepository(actualDBPath)
	if err != nil {
		return clierrors.WrapError(err, "Failed to initialize database")
	}

	// Create universal scoring engine
	universalEngine := scoring.NewUniversalEngine(p)

	// Initialize LLM Manager
	llmConfig := llm.DefaultManagerConfig()
	llmManager := llm.NewManager(llmConfig)

	// Store in shared context
	ctx = &CLIContext{
		Repository:      repo,
		UniversalEngine: universalEngine,
		Profile:         p,
		LLMManager:      llmManager,
		DBPath:          actualDBPath,
		ProfilePath:     profilePath,
		ScoringMode:     ScoringModeUniversal,
	}

	return nil
}

// initializeLegacyMode sets up the context with traditional telos.md-based scoring
func initializeLegacyMode() error {
	// Create .telos directory if it doesn't exist
	telosDir := filepath.Dir(telosPath)
	if err := os.MkdirAll(telosDir, 0755); err != nil {
		return clierrors.WrapError(err, "Failed to create config directory")
	}

	// Parse telos.md
	parser := telos.NewParser()
	telosData, err := parser.ParseFile(telosPath)
	if err != nil {
		return clierrors.WrapError(err, "Failed to parse telos.md")
	}

	// Initialize database
	repo, err := database.NewRepository(dbPath)
	if err != nil {
		return clierrors.WrapError(err, "Failed to initialize database")
	}

	// Create scoring engine and pattern detector
	engine := scoring.NewEngine(telosData)
	detector := patterns.NewDetector(telosData)

	// Initialize LLM Manager
	llmConfig := llm.DefaultManagerConfig()
	llmManager := llm.NewManager(llmConfig)

	// Store in shared context
	ctx = &CLIContext{
		Repository:  repo,
		Engine:      engine,
		Detector:    detector,
		Telos:       telosData,
		LLMManager:  llmManager,
		DBPath:      dbPath,
		TelosPath:   telosPath,
		ScoringMode: ScoringModeLegacy,
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
