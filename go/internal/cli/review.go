package cli

import (
	"fmt"
	"strings"

	"github.com/rayyacub/telos-idea-matrix/internal/database"
	"github.com/spf13/cobra"
)

var (
	reviewMinScore float64
	reviewMaxScore float64
	reviewStatus   string
	reviewLimit    int
)

func newReviewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review",
		Short: "Browse and filter saved ideas",
		Long: `Browse and filter your saved ideas with various criteria.

Examples:
  tm review                           # List all ideas
  tm review --min-score 7.0           # Ideas with score >= 7.0
  tm review --max-score 5.0           # Ideas with score <= 5.0
  tm review --status archived         # Archived ideas only
  tm review --limit 5                 # Show only 5 ideas`,
		RunE: runReview,
	}

	cmd.Flags().Float64Var(&reviewMinScore, "min-score", 0, "Minimum score filter")
	cmd.Flags().Float64Var(&reviewMaxScore, "max-score", 0, "Maximum score filter")
	cmd.Flags().StringVar(&reviewStatus, "status", "active", "Status filter (active/archived/deleted)")
	cmd.Flags().IntVar(&reviewLimit, "limit", 10, "Maximum number of ideas to show")

	return cmd
}

func runReview(cmd *cobra.Command, args []string) error {
	// Build list options
	opts := database.ListOptions{
		Status:  reviewStatus,
		OrderBy: "final_score DESC",
	}

	if cmd.Flags().Changed("min-score") {
		opts.MinScore = &reviewMinScore
	}
	if cmd.Flags().Changed("max-score") {
		opts.MaxScore = &reviewMaxScore
	}
	if reviewLimit > 0 {
		opts.Limit = &reviewLimit
	}

	// Fetch ideas
	ideas, err := ctx.Repository.List(opts)
	if err != nil {
		return fmt.Errorf("failed to list ideas: %w", err)
	}

	if len(ideas) == 0 {
		warningColor.Println("No ideas found matching your filters.")
		return nil
	}

	// Display results
	fmt.Println(strings.Repeat("═", 80))
	successColor.Printf(" %d Ideas Found\n", len(ideas))
	fmt.Println(strings.Repeat("═", 80))
	fmt.Println()

	for i, idea := range ideas {
		// Score color
		scoreColor := getScoreColor(idea.FinalScore)

		// Header
		fmt.Printf("%d. ", i+1)
		scoreColor.Printf("%.1f/10", idea.FinalScore)
		fmt.Printf(" - ID: %s\n", idea.ID[:8])

		// Content
		fmt.Printf("   %s\n", truncateString(idea.Content, 70))

		// Recommendation
		if idea.Recommendation != "" {
			recColor := getRecommendationColor(idea.Recommendation)
			recColor.Printf("   %s\n", idea.Recommendation)
		}

		// Patterns
		if len(idea.Patterns) > 0 {
			warningColor.Printf("   Patterns: %d detected\n", len(idea.Patterns))
		}

		// Created date
		fmt.Printf("   Created: %s\n", idea.CreatedAt.Format("2006-01-02 15:04"))

		fmt.Println()
	}

	fmt.Println(strings.Repeat("═", 80))
	infoColor.Printf("Showing %d of your ideas (use --limit to see more)\n", len(ideas))
	fmt.Println(strings.Repeat("═", 80))

	return nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
