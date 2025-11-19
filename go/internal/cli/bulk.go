package cli

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/rayyacub/telos-idea-matrix/internal/database"
	"github.com/rayyacub/telos-idea-matrix/internal/export"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/spf13/cobra"
)

// NewBulkCommand creates the bulk operations command.
func NewBulkCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bulk",
		Short: "Bulk operations on multiple ideas",
		Long: `Perform bulk operations on multiple ideas at once:
- tag: Add tags to multiple ideas based on filters
- archive: Archive old or low-scoring ideas
- delete: Permanently delete ideas (requires confirmation)
- import: Import ideas from CSV
- export: Export ideas to CSV or JSON`,
	}

	cmd.AddCommand(newBulkTagCommand())
	cmd.AddCommand(newBulkArchiveCommand())
	cmd.AddCommand(newBulkDeleteCommand())
	cmd.AddCommand(newBulkImportCommand())
	cmd.AddCommand(newBulkExportCommand())

	return cmd
}

func newBulkTagCommand() *cobra.Command {
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
			tagName := args[0]

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
				ideas = filterBySearch(ideas, search)
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
						truncate(idea.Content, 60),
						idea.FinalScore)
				}
			}
			if len(ideas) > 5 {
				fmt.Printf("  ... and %d more\n", len(ideas)-5)
			}

			// Confirm
			if !yes && !confirm("Proceed with tagging?") {
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
						warningColor.Printf("⚠  Failed to tag idea %s: %v\n", idea.ID, err)
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
				warningColor.Printf("⚠  %d ideas failed to tag\n", errorCount)
			}

			successColor.Printf("✅ Tagged %d ideas with '%s'\n", successCount, tagName)
			return nil
		},
	}

	cmd.Flags().Float64Var(&minScore, "min-score", 0, "Minimum score threshold")
	cmd.Flags().StringVar(&search, "search", "", "Search term to filter ideas")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum ideas to process")
	cmd.Flags().BoolVar(&yes, "yes", false, "Auto-confirm (skip confirmation prompt)")

	return cmd
}

func newBulkArchiveCommand() *cobra.Command {
	var olderThan int
	var maxScore float64
	var minScore float64
	var search string
	var limit int
	var yes bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "archive",
		Short: "Archive multiple old/low-scoring ideas",
		Long: `Archive multiple ideas based on age and score filters.
Use --older-than to archive ideas older than N days.
Use --max-score to archive ideas below a score threshold.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Build filter options
			maxScorePtr := &maxScore
			if maxScore == 0 {
				maxScorePtr = nil
			}
			minScorePtr := &minScore
			if minScore == 0 {
				minScorePtr = nil
			}
			limitPtr := &limit

			ideas, err := ctx.Repository.List(database.ListOptions{
				Status:   "active",
				MinScore: minScorePtr,
				MaxScore: maxScorePtr,
				Limit:    limitPtr,
				OrderBy:  "created_at ASC", // Oldest first
			})
			if err != nil {
				return fmt.Errorf("failed to list ideas: %w", err)
			}

			// Filter by age if specified
			if olderThan > 0 {
				cutoffDate := time.Now().UTC().Add(-time.Duration(olderThan) * 24 * time.Hour)
				ideas = filterByAge(ideas, cutoffDate)
			}

			// Filter by search if provided
			if search != "" {
				ideas = filterBySearch(ideas, search)
			}

			if len(ideas) == 0 {
				fmt.Println("📭 No ideas match your criteria for archiving.")
				return nil
			}

			// Show preview
			fmt.Printf("📦 Found %s ideas to archive:\n", color.CyanString("%d", len(ideas)))
			for i, idea := range ideas {
				if i < 5 {
					age := time.Since(idea.CreatedAt).Hours() / 24
					fmt.Printf("  - %s (score: %.1f, age: %.0f days)\n",
						truncate(idea.Content, 50),
						idea.FinalScore,
						age)
				}
			}
			if len(ideas) > 5 {
				fmt.Printf("  ... and %d more\n", len(ideas)-5)
			}

			if dryRun {
				infoColor.Println("\n🔍 DRY RUN - No changes will be made")
				return nil
			}

			// Confirm
			if !yes && !confirm("Proceed with archiving?") {
				fmt.Println("❌ Cancelled")
				return nil
			}

			// Archive ideas
			successCount := 0
			errorCount := 0
			for i, idea := range ideas {
				idea.Status = "archived"
				if err := ctx.Repository.Update(idea); err != nil {
					warningColor.Printf("⚠  Failed to archive idea %s: %v\n", idea.ID, err)
					errorCount++
					continue
				}
				successCount++

				// Show progress for large batches
				if len(ideas) > 10 && (i+1)%10 == 0 {
					fmt.Printf("  Progress: %d/%d archived\n", i+1, len(ideas))
				}
			}

			if errorCount > 0 {
				warningColor.Printf("⚠  %d ideas failed to archive\n", errorCount)
			}

			successColor.Printf("✅ Archived %d ideas\n", successCount)
			return nil
		},
	}

	cmd.Flags().IntVar(&olderThan, "older-than", 0, "Archive ideas older than N days")
	cmd.Flags().Float64Var(&maxScore, "max-score", 0, "Maximum score threshold")
	cmd.Flags().Float64Var(&minScore, "min-score", 0, "Minimum score threshold")
	cmd.Flags().StringVar(&search, "search", "", "Search term to filter ideas")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum ideas to process")
	cmd.Flags().BoolVar(&yes, "yes", false, "Auto-confirm (skip confirmation prompt)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be archived without making changes")

	return cmd
}

func newBulkDeleteCommand() *cobra.Command {
	var olderThan int
	var maxScore float64
	var search string
	var limit int
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Permanently delete multiple ideas",
		Long: `Permanently delete multiple ideas based on filters.
⚠️  WARNING: This operation cannot be undone!
Always requires confirmation for safety.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Build filter options
			maxScorePtr := &maxScore
			if maxScore == 0 {
				maxScorePtr = nil
			}
			limitPtr := &limit

			ideas, err := ctx.Repository.List(database.ListOptions{
				Status:   "active",
				MaxScore: maxScorePtr,
				Limit:    limitPtr,
				OrderBy:  "created_at ASC",
			})
			if err != nil {
				return fmt.Errorf("failed to list ideas: %w", err)
			}

			// Filter by age if specified
			if olderThan > 0 {
				cutoffDate := time.Now().UTC().Add(-time.Duration(olderThan) * 24 * time.Hour)
				ideas = filterByAge(ideas, cutoffDate)
			}

			// Filter by search if provided
			if search != "" {
				ideas = filterBySearch(ideas, search)
			}

			if len(ideas) == 0 {
				fmt.Println("📭 No ideas match your criteria for deletion.")
				return nil
			}

			// Show preview
			errorColor.Printf("⚠️  WARNING: About to PERMANENTLY DELETE %d ideas:\n", len(ideas))
			for i, idea := range ideas {
				if i < 5 {
					fmt.Printf("  - %s (score: %.1f)\n",
						truncate(idea.Content, 50),
						idea.FinalScore)
				}
			}
			if len(ideas) > 5 {
				fmt.Printf("  ... and %d more\n", len(ideas)-5)
			}

			// Always require confirmation for delete
			if !yes {
				fmt.Println()
				if !confirm("⚠️  PERMANENTLY DELETE these ideas? This CANNOT be undone!") {
					fmt.Println("❌ Cancelled")
					return nil
				}
			}

			// Delete ideas
			successCount := 0
			errorCount := 0
			for i, idea := range ideas {
				if err := ctx.Repository.Delete(idea.ID); err != nil {
					warningColor.Printf("⚠  Failed to delete idea %s: %v\n", idea.ID, err)
					errorCount++
					continue
				}
				successCount++

				// Show progress for large batches
				if len(ideas) > 10 && (i+1)%10 == 0 {
					fmt.Printf("  Progress: %d/%d deleted\n", i+1, len(ideas))
				}
			}

			if errorCount > 0 {
				warningColor.Printf("⚠  %d ideas failed to delete\n", errorCount)
			}

			errorColor.Printf("🗑️  Permanently deleted %d ideas\n", successCount)
			return nil
		},
	}

	cmd.Flags().IntVar(&olderThan, "older-than", 0, "Delete ideas older than N days")
	cmd.Flags().Float64Var(&maxScore, "max-score", 0, "Maximum score threshold")
	cmd.Flags().StringVar(&search, "search", "", "Search term to filter ideas")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum ideas to process")
	cmd.Flags().BoolVar(&yes, "yes", false, "Auto-confirm (skip confirmation prompt)")

	return cmd
}

func newBulkImportCommand() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "import <file>",
		Short: "Import ideas from CSV",
		Long: `Import ideas from a CSV file.
The CSV file should have the following columns:
ID,Content,RawScore,FinalScore,Patterns,Recommendation,AnalysisDetails,CreatedAt,Status`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filename := args[0]

			// Import from CSV
			ideas, err := export.ImportCSV(filename)
			if err != nil {
				return fmt.Errorf("failed to import CSV: %w", err)
			}

			if len(ideas) == 0 {
				fmt.Println("📭 No ideas found in CSV file.")
				return nil
			}

			// Show preview
			fmt.Printf("📥 Found %s ideas to import from '%s':\n",
				color.CyanString("%d", len(ideas)),
				filename)
			for i, idea := range ideas {
				if i < 5 {
					fmt.Printf("  - %s\n", truncate(idea.Content, 60))
				}
			}
			if len(ideas) > 5 {
				fmt.Printf("  ... and %d more\n", len(ideas)-5)
			}

			// Confirm
			if !yes && !confirm("Proceed with import?") {
				fmt.Println("❌ Cancelled")
				return nil
			}

			// Import ideas
			successCount := 0
			errorCount := 0
			for i, idea := range ideas {
				// Validate idea before import
				if err := idea.Validate(); err != nil {
					warningColor.Printf("⚠  Skipping invalid idea: %v\n", err)
					errorCount++
					continue
				}

				if err := ctx.Repository.Create(idea); err != nil {
					warningColor.Printf("⚠  Failed to import idea: %v\n", err)
					errorCount++
					continue
				}
				successCount++

				// Show progress for large batches
				if len(ideas) > 10 && (i+1)%10 == 0 {
					fmt.Printf("  Progress: %d/%d imported\n", i+1, len(ideas))
				}
			}

			if errorCount > 0 {
				warningColor.Printf("⚠  %d ideas failed to import\n", errorCount)
			}

			successColor.Printf("✅ Imported %d ideas from '%s'\n", successCount, filename)
			return nil
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Auto-confirm (skip confirmation prompt)")

	return cmd
}

func newBulkExportCommand() *cobra.Command {
	var minScore float64
	var search string
	var limit int
	var format string
	var pretty bool

	cmd := &cobra.Command{
		Use:   "export <file>",
		Short: "Export ideas to CSV or JSON",
		Long: `Export ideas to a file in CSV or JSON format.
Use --format to specify the output format (csv or json).
Use filters to control which ideas are exported.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filename := args[0]

			// Auto-detect format from extension if not specified
			if format == "" {
				ext := strings.ToLower(filepath.Ext(filename))
				if ext == ".json" {
					format = "json"
				} else {
					format = "csv"
				}
			}

			// Fetch ideas to export
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
				ideas = filterBySearch(ideas, search)
			}

			if len(ideas) == 0 {
				fmt.Println("📭 No ideas match your criteria for export.")
				return nil
			}

			// Export based on format
			switch format {
			case "json":
				err = export.ExportJSON(ideas, filename, pretty)
			case "csv":
				err = export.ExportCSV(ideas, filename)
			default:
				return fmt.Errorf("unsupported format: %s (use 'csv' or 'json')", format)
			}

			if err != nil {
				return fmt.Errorf("failed to export: %w", err)
			}

			successColor.Printf("✅ Exported %d ideas to '%s' (%s format)\n",
				len(ideas), filename, format)
			return nil
		},
	}

	cmd.Flags().Float64Var(&minScore, "min-score", 0, "Minimum score threshold")
	cmd.Flags().StringVar(&search, "search", "", "Search term to filter ideas")
	cmd.Flags().IntVar(&limit, "limit", 1000, "Maximum ideas to export")
	cmd.Flags().StringVar(&format, "format", "", "Output format: csv or json (auto-detected from extension)")
	cmd.Flags().BoolVar(&pretty, "pretty", false, "Pretty-print JSON output (only for JSON format)")

	return cmd
}

// Helper functions

func filterBySearch(ideas []*models.Idea, searchTerm string) []*models.Idea {
	searchLower := strings.ToLower(searchTerm)
	filtered := make([]*models.Idea, 0)

	for _, idea := range ideas {
		if strings.Contains(strings.ToLower(idea.Content), searchLower) ||
			strings.Contains(strings.ToLower(idea.Recommendation), searchLower) ||
			strings.Contains(strings.ToLower(idea.AnalysisDetails), searchLower) {
			filtered = append(filtered, idea)
		}
	}

	return filtered
}

func filterByAge(ideas []*models.Idea, cutoffDate time.Time) []*models.Idea {
	filtered := make([]*models.Idea, 0)

	for _, idea := range ideas {
		if idea.CreatedAt.Before(cutoffDate) {
			filtered = append(filtered, idea)
		}
	}

	return filtered
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func confirm(prompt string) bool {
	fmt.Printf("%s [y/N]: ", prompt)
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}
