package cli

import (
	"fmt"

	"github.com/rayyacub/telos-idea-matrix/internal/database"
	"github.com/spf13/cobra"
)

func newAnalyticsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "analytics",
		Short: "View statistics about your ideas",
		Long: `Display statistics and insights about your captured ideas.

Examples:
  tm analytics`,
		RunE: runAnalytics,
	}
}

func runAnalytics(cmd *cobra.Command, args []string) error {
	// Fetch all active ideas
	ideas, err := ctx.Repository.List(database.ListOptions{
		Status: "active",
	})
	if err != nil {
		return fmt.Errorf("failed to list ideas: %w", err)
	}

	if len(ideas) == 0 {
		warningColor.Println("No ideas found. Use 'tm dump' to capture your first idea!")
		return nil
	}

	// Calculate statistics
	var totalScore float64
	var highScore, lowScore float64 = 0, 10
	highCount := 0  // >= 7.0
	mediumCount := 0 // 5.0-7.0
	lowCount := 0    // < 5.0

	for _, idea := range ideas {
		totalScore += idea.FinalScore

		if idea.FinalScore > highScore {
			highScore = idea.FinalScore
		}
		if idea.FinalScore < lowScore {
			lowScore = idea.FinalScore
		}

		switch {
		case idea.FinalScore >= 7.0:
			highCount++
		case idea.FinalScore >= 5.0:
			mediumCount++
		default:
			lowCount++
		}
	}

	avgScore := totalScore / float64(len(ideas))

	// Display statistics
	fmt.Println("📊 Idea Analytics")
	fmt.Println("═════════════════════════════════════════════")
	fmt.Println()

	successColor.Printf("Total Ideas: %d\n", len(ideas))
	fmt.Printf("Average Score: %.1f/10.0\n", avgScore)
	fmt.Printf("Highest Score: %.1f/10.0\n", highScore)
	fmt.Printf("Lowest Score:  %.1f/10.0\n\n", lowScore)

	fmt.Println("Score Distribution:")
	successColor.Printf("  🔥 High (>= 7.0):   %d ideas (%.0f%%)\n",
		highCount, float64(highCount)/float64(len(ideas))*100)
	warningColor.Printf("  ⚠️  Medium (5-7):   %d ideas (%.0f%%)\n",
		mediumCount, float64(mediumCount)/float64(len(ideas))*100)
	errorColor.Printf("  🚫 Low (< 5.0):     %d ideas (%.0f%%)\n",
		lowCount, float64(lowCount)/float64(len(ideas))*100)
	fmt.Println()

	// Recommendations
	if highCount > 0 {
		successColor.Printf("✨ You have %d high-scoring ideas to prioritize!\n", highCount)
	}
	if lowCount > len(ideas)/2 {
		warningColor.Println("💡 Tip: Many ideas are low-scoring. Consider aligning more with your telos.")
	}

	fmt.Println("═════════════════════════════════════════════")

	return nil
}
