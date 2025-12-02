package analytics

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/ryacub/telos-idea-matrix/internal/analytics"
	"github.com/ryacub/telos-idea-matrix/internal/cliutil"
	"github.com/ryacub/telos-idea-matrix/internal/database"
	"github.com/spf13/cobra"
)

// CLIContext represents the shared CLI dependencies
type CLIContext struct {
	Repository *database.Repository
	DBPath     string
}

// NewAnalyticsCommand creates the analytics command with all subcommands
func NewAnalyticsCommand(getContext func() *CLIContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analytics",
		Short: "View statistics about your ideas",
		Long: `Display statistics and insights about your captured ideas.

Examples:
  tm analytics              # Show basic statistics
  tm analytics trends       # Show score trends over time
  tm analytics report       # Generate comprehensive report
  tm analytics patterns     # Show pattern frequency`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAnalytics(getContext)
		},
	}

	// Add subcommands
	cmd.AddCommand(NewTrendsCommand(getContext))
	cmd.AddCommand(NewReportCommand(getContext))
	cmd.AddCommand(NewPatternsCommand(getContext))
	cmd.AddCommand(NewMetricsCommand(getContext))

	return cmd
}

func runAnalytics(getContext func() *CLIContext) error {
	ctx := getContext()
	if ctx == nil {
		return fmt.Errorf("CLI context not initialized")
	}

	// Fetch all active ideas
	ideas, err := ctx.Repository.List(database.ListOptions{
		Status: "active",
	})
	if err != nil {
		return fmt.Errorf("failed to list ideas: %w", err)
	}

	if len(ideas) == 0 {
		warningColor := cliutil.GetScoreColor(5.0)
		if _, err := warningColor.Println("No ideas found. Use 'tm dump' to capture your first idea!"); err != nil {
			log.Warn().Err(err).Msg("failed to print warning message")
		}
		return nil
	}

	// Calculate statistics using service
	service := analytics.NewService(ctx.Repository)
	stats := service.GetBasicStats(ideas)
	highPct, mediumPct, lowPct := service.ScoreDistribution(stats)

	// Display statistics
	fmt.Println("📊 Idea Analytics")
	fmt.Println("═════════════════════════════════════════════")
	fmt.Println()

	successColor := cliutil.GetScoreColor(10.0)
	if _, err := successColor.Printf("Total Ideas: %d\n", stats.TotalIdeas); err != nil {
		log.Warn().Err(err).Msg("failed to print total ideas")
	}
	fmt.Printf("Average Score: %.1f/10.0\n", stats.AverageScore)
	fmt.Printf("Highest Score: %.1f/10.0\n", stats.HighScore)
	fmt.Printf("Lowest Score:  %.1f/10.0\n\n", stats.LowScore)

	fmt.Println("Score Distribution:")

	// Visual distribution bar
	distBar := analytics.RenderDistribution(stats.HighCount, stats.MediumCount, stats.LowCount, 50)
	fmt.Printf("%s\n\n", distBar)

	if _, err := successColor.Printf("  🔥 High (>= 7.0):   %d ideas (%.0f%%)\n",
		stats.HighCount, highPct); err != nil {
		log.Warn().Err(err).Msg("failed to print high count")
	}

	warningColor := cliutil.GetScoreColor(5.0)
	if _, err := warningColor.Printf("  ⚠️  Medium (5-7):   %d ideas (%.0f%%)\n",
		stats.MediumCount, mediumPct); err != nil {
		log.Warn().Err(err).Msg("failed to print medium count")
	}

	errorColor := cliutil.GetScoreColor(0.0)
	if _, err := errorColor.Printf("  🚫 Low (< 5.0):     %d ideas (%.0f%%)\n",
		stats.LowCount, lowPct); err != nil {
		log.Warn().Err(err).Msg("failed to print low count")
	}
	fmt.Println()

	// Recommendations
	if stats.HighCount > 0 {
		if _, err := successColor.Printf("✨ You have %d high-scoring ideas to prioritize!\n", stats.HighCount); err != nil {
			log.Warn().Err(err).Msg("failed to print recommendation")
		}
	}
	if stats.LowCount > stats.TotalIdeas/2 {
		if _, err := warningColor.Println("💡 Tip: Many ideas are low-scoring. Consider aligning more with your telos."); err != nil {
			log.Warn().Err(err).Msg("failed to print tip")
		}
	}

	fmt.Println("═════════════════════════════════════════════")

	return nil
}
