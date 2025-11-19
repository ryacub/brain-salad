package cli

import (
	"fmt"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/database"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/spf13/cobra"
)

var (
	pruneDays   int
	pruneScore  float64
	pruneDryRun bool
)

func newPruneCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Clean up old or low-scoring ideas",
		Long: `Archive or delete old ideas based on age or score.

Examples:
  tm prune --days 90 --dry-run      # Show ideas older than 90 days
  tm prune --score 3.0 --dry-run    # Show ideas with score < 3.0
  tm prune --days 90                # Archive ideas older than 90 days`,
		RunE: runPrune,
	}

	cmd.Flags().IntVar(&pruneDays, "days", 0, "Archive ideas older than N days")
	cmd.Flags().Float64Var(&pruneScore, "score", 0, "Archive ideas with score below N")
	cmd.Flags().BoolVar(&pruneDryRun, "dry-run", false, "Show what would be pruned without doing it")

	return cmd
}

func runPrune(cmd *cobra.Command, args []string) error {
	if pruneDays == 0 && pruneScore == 0 {
		return fmt.Errorf("specify --days or --score")
	}

	// Build filter options
	opts := database.ListOptions{
		Status: "active",
	}

	if pruneScore > 0 {
		opts.MaxScore = &pruneScore
	}

	// Fetch ideas
	ideas, err := ctx.Repository.List(opts)
	if err != nil {
		return fmt.Errorf("failed to list ideas: %w", err)
	}

	// Filter by age if specified
	var toPrune []*models.Idea
	cutoffDate := time.Now().AddDate(0, 0, -pruneDays)

	for _, idea := range ideas {
		if pruneDays > 0 && idea.CreatedAt.Before(cutoffDate) {
			toPrune = append(toPrune, idea)
		} else if pruneDays == 0 && pruneScore > 0 {
			toPrune = append(toPrune, idea)
		}
	}

	if len(toPrune) == 0 {
		successColor.Println("✅ No ideas to prune!")
		return nil
	}

	// Display what would be pruned
	fmt.Printf("Found %d ideas to prune:\n\n", len(toPrune))
	for i, idea := range toPrune {
		fmt.Printf("%d. [%.1f] %s\n", i+1, idea.FinalScore, truncateString(idea.Content, 60))
		fmt.Printf("   Created: %s\n", idea.CreatedAt.Format("2006-01-02"))
	}
	fmt.Println()

	if pruneDryRun {
		infoColor.Println("🔍 Dry run - nothing was changed")
		return nil
	}

	// Archive ideas
	archived := 0
	for _, idea := range toPrune {
		idea.Status = "archived"
		if err := ctx.Repository.Update(idea); err != nil {
			warningColor.Printf("Failed to archive idea %s: %v\n", idea.ID[:8], err)
			continue
		}
		archived++
	}

	successColor.Printf("✅ Archived %d ideas\n", archived)
	return nil
}
