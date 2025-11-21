package analytics

import (
	"fmt"

	"github.com/rayyacub/telos-idea-matrix/internal/analytics"
	"github.com/rayyacub/telos-idea-matrix/internal/database"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// NewTrendsCommand creates the analytics trends subcommand
func NewTrendsCommand(getContext func() *CLIContext) *cobra.Command {
	var days int
	var groupBy string

	cmd := &cobra.Command{
		Use:   "trends",
		Short: "Show score trends over time",
		Long: `Display score trends grouped by time period.

The trends command shows how your idea scores have evolved over time,
helping you identify patterns in your ideation process.

Examples:
  tm analytics trends                    # Weekly trends for last 30 days
  tm analytics trends --days 90          # Weekly trends for last 90 days
  tm analytics trends --group-by month   # Monthly trends
  tm analytics trends --group-by day     # Daily trends`,
		RunE: func(cmd *cobra.Command, args []string) error {
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
				warningColor := getScoreColor(5.0)
				if _, err := warningColor.Println("No ideas found. Use 'tm dump' to capture your first idea!"); err != nil {
					log.Warn().Err(err).Msg("failed to print warning message")
				}
				return nil
			}

			// Calculate trends
			trends := analytics.CalculateScoreTrends(ideas, groupBy)

			// Display header
			fmt.Println("📈 Score Trends")
			fmt.Println("═════════════════════════════════════════════")
			fmt.Println()

			if len(trends) == 0 {
				fmt.Println("No trend data available yet.")
				return nil
			}

			// Display trends
			fmt.Printf("Grouping: %s\n\n", groupBy)

			// Generate sparkline
			values := make([]float64, len(trends))
			for i, trend := range trends {
				values[i] = trend.AvgScore
			}
			sparkline := analytics.RenderSparkline(values)

			fmt.Printf("Trend: %s\n\n", sparkline)

			successColor := getScoreColor(10.0)
			warningColor := getScoreColor(5.0)
			errorColor := getScoreColor(0.0)

			for _, trend := range trends {
				// Color code based on score
				var scoreColor func(string, ...interface{}) string
				if trend.AvgScore >= 7.0 {
					scoreColor = successColor.Sprintf
				} else if trend.AvgScore >= 5.0 {
					scoreColor = warningColor.Sprintf
				} else {
					scoreColor = errorColor.Sprintf
				}

				fmt.Printf("  %s: %s (%d ideas)\n",
					trend.Period,
					scoreColor("%.1f avg", trend.AvgScore),
					trend.IdeaCount,
				)
			}

			// Show trend direction
			direction := analytics.CalculateTrendDirection(trends)
			fmt.Println()
			switch direction {
			case "up":
				if _, err := successColor.Println("📈 Trend: Your idea quality is improving over time!"); err != nil {
					log.Warn().Err(err).Msg("failed to print trend message")
				}
			case "down":
				if _, err := warningColor.Println("📉 Trend: Consider refining your idea capture process."); err != nil {
					log.Warn().Err(err).Msg("failed to print trend message")
				}
			default:
				fmt.Println("➡️  Trend: Your idea quality is stable.")
			}

			fmt.Println("═════════════════════════════════════════════")

			return nil
		},
	}

	cmd.Flags().IntVar(&days, "days", 30, "Number of days to analyze")
	cmd.Flags().StringVar(&groupBy, "group-by", "week", "Group by: day, week, or month")

	return cmd
}
