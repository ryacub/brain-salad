package bulk

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/rayyacub/telos-idea-matrix/internal/bulk"
	"github.com/rayyacub/telos-idea-matrix/internal/cliutil"
	"github.com/rayyacub/telos-idea-matrix/internal/database"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// NewTagCommand creates the bulk tag command
func NewTagCommand(getContext func() *CLIContext) *cobra.Command {
	var minScore float64
	var search string
	var limit int
	var yes bool

	cmd := &cobra.Command{
		Use:   "tag <tag-name>",
		Short: "Add tag to multiple ideas",
		Long: `Add a tag to multiple ideas based on filters.
Use --min-score, --search, and --limit to control which ideas are tagged.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := getContext()
			if ctx == nil {
				return fmt.Errorf("CLI context not initialized")
			}

			tagName := args[0]

			// Create service once
			service := bulk.NewService(ctx.Repository)

			// Find matching ideas
			minScorePtr := &minScore
			limitPtr := &limit
			ideas, err := ctx.Repository.List(database.ListOptions{
				Status:   "active",
				MinScore: minScorePtr,
				Limit:    limitPtr,
				OrderBy:  "final_score DESC",
			})
			if err != nil {
				return fmt.Errorf("failed to list ideas: %w", err)
			}

			// Filter by search if provided
			if search != "" {
				ideas = service.FilterBySearch(ideas, search)
			}

			if len(ideas) == 0 {
				fmt.Println("📭 No ideas match your criteria.")
				return nil
			}

			// Show preview
			fmt.Printf("🎯 Found %s ideas to tag with '%s':\n",
				color.CyanString("%d", len(ideas)),
				color.GreenString(tagName))
			for i, idea := range ideas {
				if i < 5 { // Show first 5
					fmt.Printf("  - %s (score: %.1f)\n",
						cliutil.TruncateText(idea.Content, 60),
						idea.FinalScore)
				}
			}
			if len(ideas) > 5 {
				fmt.Printf("  ... and %d more\n", len(ideas)-5)
			}

			// Confirm
			if !yes && !cliutil.Confirm("Proceed with tagging?") {
				fmt.Println("❌ Cancelled")
				return nil
			}

			// Apply tags (placeholder - would need tags table)
			successCount := 0
			errorCount := 0
			for i, idea := range ideas {
				// In a real implementation, we would add tags to a tags table
				// For now, we'll append to analysis details as a workaround
				if !strings.Contains(idea.AnalysisDetails, tagName) {
					idea.AnalysisDetails = fmt.Sprintf("%s [tag:%s]", idea.AnalysisDetails, tagName)
					if err := ctx.Repository.Update(idea); err != nil {
						if _, printErr := cliutil.WarningColor.Printf("⚠  Failed to tag idea %s: %v\n", idea.ID, err); printErr != nil {
							log.Warn().Err(printErr).Msg("failed to print error message")
						}
						errorCount++
						continue
					}
				}
				successCount++

				// Show progress for large batches
				if len(ideas) > 10 && (i+1)%10 == 0 {
					fmt.Printf("  Progress: %d/%d tagged\n", i+1, len(ideas))
				}
			}

			if errorCount > 0 {
				if _, err := cliutil.WarningColor.Printf("⚠  %d ideas failed to tag\n", errorCount); err != nil {
					log.Warn().Err(err).Msg("failed to print warning message")
				}
			}

			if _, err := cliutil.SuccessColor.Printf("✅ Tagged %d ideas with '%s'\n", successCount, tagName); err != nil {
				log.Warn().Err(err).Msg("failed to print success message")
			}
			return nil
		},
	}

	cmd.Flags().Float64Var(&minScore, "min-score", 0, "Minimum score threshold")
	cmd.Flags().StringVar(&search, "search", "", "Search term to filter ideas")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum ideas to process")
	cmd.Flags().BoolVar(&yes, "yes", false, "Auto-confirm (skip confirmation prompt)")

	return cmd
}
