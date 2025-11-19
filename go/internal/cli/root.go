package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/rayyacub/telos-idea-matrix/internal/database"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/rayyacub/telos-idea-matrix/internal/patterns"
	"github.com/rayyacub/telos-idea-matrix/internal/scoring"
	"github.com/rayyacub/telos-idea-matrix/internal/telos"
	"github.com/spf13/cobra"
)

// CLIContext holds shared dependencies for all commands
type CLIContext struct {
	Repository *database.Repository
	Engine     *scoring.Engine
	Detector   *patterns.Detector
	Telos      *models.Telos
	DBPath     string
	TelosPath  string
}

var (
	ctx        *CLIContext
	dbPath     string
	telosPath  string
	rootCmd    *cobra.Command

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
	rootCmd.AddCommand(newDumpCommand())
	rootCmd.AddCommand(newScoreCommand())
	rootCmd.AddCommand(newAnalyzeCommand())
	rootCmd.AddCommand(newReviewCommand())
	rootCmd.AddCommand(newPruneCommand())
	rootCmd.AddCommand(newAnalyticsCommand())
	rootCmd.AddCommand(newLinkCommand())
	rootCmd.AddCommand(newHealthCommand())
	rootCmd.AddCommand(NewBulkCommand())
}

// initializeCLI sets up the shared context for all commands
func initializeCLI(cmd *cobra.Command, args []string) error {
	// Create .telos directory if it doesn't exist
	telosDir := filepath.Dir(telosPath)
	if err := os.MkdirAll(telosDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check if telos.md exists
	if _, err := os.Stat(telosPath); os.IsNotExist(err) {
		warningColor.Fprintf(os.Stderr, "�  Telos file not found at %s\n", telosPath)
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

	// Store in shared context
	ctx = &CLIContext{
		Repository: repo,
		Engine:     engine,
		Detector:   detector,
		Telos:      telosData,
		DBPath:     dbPath,
		TelosPath:  telosPath,
	}

	return nil
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// GetRootCmd returns the root command for testing
func GetRootCmd() *cobra.Command {
	return rootCmd
}

// SetContext allows setting a custom context for testing
func SetContext(c *CLIContext) {
	ctx = c
}
