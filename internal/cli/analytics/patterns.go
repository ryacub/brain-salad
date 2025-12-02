package analytics

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/ryacub/telos-idea-matrix/internal/analytics"
	"github.com/ryacub/telos-idea-matrix/internal/cliutil"
	"github.com/ryacub/telos-idea-matrix/internal/database"
	"github.com/spf13/cobra"
)

// NewPatternsCommand creates the analytics patterns subcommand
func NewPatternsCommand(getContext func() *CLIContext) *cobra.Command {
	var topN int

	cmd := &cobra.Command{
		Use:   "patterns",
		Short: "Show pattern frequency analysis",
		Long: `Display the most frequently occurring patterns in your ideas.

This helps identify recurring anti-patterns or common themes
across your idea collection.

Examples:
  tm analytics patterns           # Show top 10 patterns
  tm analytics patterns --top 5   # Show top 5 patterns`,
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
				warningColor := cliutil.GetScoreColor(5.0)
				if _, err := warningColor.Println("No ideas found. Use 'tm dump' to capture your first idea!"); err != nil {
					log.Warn().Err(err).Msg("failed to print warning message")
				}
				return nil
			}

			// Calculate pattern frequency
			freq := analytics.CalculatePatternFrequency(ideas)

			if len(freq) == 0 {
				fmt.Println("No patterns detected in your ideas yet.")
				return nil
			}

			// Get top patterns
			topPatterns := analytics.GetTopPatterns(ideas, topN)

			// Display header
			fmt.Println("🔍 Pattern Frequency Analysis")
			fmt.Println("═════════════════════════════════════════════")
			fmt.Println()

			// Prepare data for bar chart
			labels := make([]string, len(topPatterns))
			values := make([]float64, len(topPatterns))
			for i, pattern := range topPatterns {
				labels[i] = pattern
				values[i] = float64(freq[pattern])
			}

			// Display bar chart
			chart := analytics.RenderBarChart(labels, values, 40)
			fmt.Println(chart)

			// Display patterns with percentages
			warningColor := cliutil.GetScoreColor(5.0)
			for i, pattern := range topPatterns {
				count := freq[pattern]
				percentage := (count * 100) / len(ideas)

				// Highlight high-frequency patterns
				if percentage > 40 {
					if _, err := warningColor.Printf("%d. %s: %d occurrences (%d%% of ideas) ⚠️\n",
						i+1, pattern, count, percentage); err != nil {
						log.Warn().Err(err).Msg("failed to print pattern")
					}
				} else {
					fmt.Printf("%d. %s: %d occurrences (%d%% of ideas)\n",
						i+1, pattern, count, percentage)
				}
			}

			fmt.Println()
			fmt.Printf("Total unique patterns detected: %d\n", len(freq))

			// Show warning if high frequency patterns exist
			hasHighFreq := false
			for _, pattern := range topPatterns {
				percentage := (freq[pattern] * 100) / len(ideas)
				if percentage > 40 {
					hasHighFreq = true
					break
				}
			}

			if hasHighFreq {
				fmt.Println()
				if _, err := warningColor.Println("⚠️  Warning: Some patterns appear very frequently."); err != nil {
					log.Warn().Err(err).Msg("failed to print warning message")
				}
				fmt.Println("   Consider addressing these recurring anti-patterns in your ideation process.")
			}

			fmt.Println("═════════════════════════════════════════════")

			return nil
		},
	}

	cmd.Flags().IntVar(&topN, "top", 10, "Number of top patterns to display")

	return cmd
}
