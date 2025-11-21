package cli

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/rayyacub/telos-idea-matrix/internal/bulk"
	"github.com/rayyacub/telos-idea-matrix/internal/database"
	"github.com/rayyacub/telos-idea-matrix/internal/export"
	"github.com/rayyacub/telos-idea-matrix/internal/llm"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const (
	formatJSON = "json"
	formatCSV  = "csv"
)

// NewBulkCommand creates the bulk operations command.
func NewBulkCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bulk",
		Short: "Bulk operations on multiple ideas",
		Long: `Perform bulk operations on multiple ideas at once:
- analyze: Re-analyze multiple ideas with updated criteria
- update: Update multiple ideas in batch
- tag: Add tags to multiple ideas based on filters
- archive: Archive old or low-scoring ideas
- delete: Permanently delete ideas (requires confirmation)
- import: Import ideas from CSV
- export: Export ideas to CSV or JSON`,
	}

	cmd.AddCommand(newBulkAnalyzeCommand())
	cmd.AddCommand(newBulkUpdateCommand())
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
				service := bulk.NewService(ctx.Repository)
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
						if _, printErr := warningColor.Printf("⚠  Failed to tag idea %s: %v\n", idea.ID, err); printErr != nil {
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
				if _, err := warningColor.Printf("⚠  %d ideas failed to tag\n", errorCount); err != nil {
					log.Warn().Err(err).Msg("failed to print warning message")
				}
			}

			if _, err := successColor.Printf("✅ Tagged %d ideas with '%s'\n", successCount, tagName); err != nil {
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
				service := bulk.NewService(ctx.Repository)
				ideas = service.FilterByAge(ideas, cutoffDate)
			}

			// Filter by search if provided
			if search != "" {
				service := bulk.NewService(ctx.Repository)
				ideas = service.FilterBySearch(ideas, search)
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
				if _, err := infoColor.Println("\n🔍 DRY RUN - No changes will be made"); err != nil {
					log.Warn().Err(err).Msg("failed to print message")
				}
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
					if _, printErr := warningColor.Printf("⚠  Failed to archive idea %s: %v\n", idea.ID, err); printErr != nil {
						log.Warn().Err(printErr).Msg("failed to print error message")
					}
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
				if _, err := warningColor.Printf("⚠  %d ideas failed to archive\n", errorCount); err != nil {
					log.Warn().Err(err).Msg("failed to print warning message")
				}
			}

			if _, err := successColor.Printf("✅ Archived %d ideas\n", successCount); err != nil {
				log.Warn().Err(err).Msg("failed to print success message")
			}
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
				service := bulk.NewService(ctx.Repository)
				ideas = service.FilterByAge(ideas, cutoffDate)
			}

			// Filter by search if provided
			if search != "" {
				service := bulk.NewService(ctx.Repository)
				ideas = service.FilterBySearch(ideas, search)
			}

			if len(ideas) == 0 {
				fmt.Println("📭 No ideas match your criteria for deletion.")
				return nil
			}

			// Show preview
			if _, err := errorColor.Printf("⚠️  WARNING: About to PERMANENTLY DELETE %d ideas:\n", len(ideas)); err != nil {
				log.Warn().Err(err).Msg("failed to print warning message")
			}
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
					if _, printErr := warningColor.Printf("⚠  Failed to delete idea %s: %v\n", idea.ID, err); printErr != nil {
						log.Warn().Err(printErr).Msg("failed to print error message")
					}
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
				if _, err := warningColor.Printf("⚠  %d ideas failed to delete\n", errorCount); err != nil {
					log.Warn().Err(err).Msg("failed to print warning message")
				}
			}

			if _, err := errorColor.Printf("🗑️  Permanently deleted %d ideas\n", successCount); err != nil {
				log.Warn().Err(err).Msg("failed to print message")
			}
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
					if _, printErr := warningColor.Printf("⚠  Skipping invalid idea: %v\n", err); printErr != nil {
						log.Warn().Err(printErr).Msg("failed to print warning")
					}
					errorCount++
					continue
				}

				if err := ctx.Repository.Create(idea); err != nil {
					if _, printErr := warningColor.Printf("⚠  Failed to import idea: %v\n", err); printErr != nil {
						log.Warn().Err(printErr).Msg("failed to print error message")
					}
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
				if _, err := warningColor.Printf("⚠  %d ideas failed to import\n", errorCount); err != nil {
					log.Warn().Err(err).Msg("failed to print warning message")
				}
			}

			if _, err := successColor.Printf("✅ Imported %d ideas from '%s'\n", successCount, filename); err != nil {
				log.Warn().Err(err).Msg("failed to print success message")
			}
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
					format = formatJSON
				} else {
					format = formatCSV
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
				service := bulk.NewService(ctx.Repository)
				ideas = service.FilterBySearch(ideas, search)
			}

			if len(ideas) == 0 {
				fmt.Println("📭 No ideas match your criteria for export.")
				return nil
			}

			// Export based on format
			switch format {
			case formatJSON:
				err = export.ExportJSON(ideas, filename, pretty)
			case formatCSV:
				err = export.ExportCSV(ideas, filename)
			default:
				return fmt.Errorf("unsupported format: %s (use 'csv' or 'json')", format)
			}

			if err != nil {
				return fmt.Errorf("failed to export: %w", err)
			}

			if _, err := successColor.Printf("✅ Exported %d ideas to '%s' (%s format)\n",
				len(ideas), filename, format); err != nil {
				log.Warn().Err(err).Msg("failed to print success message")
			}
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

func newBulkUpdateCommand() *cobra.Command {
	var (
		setStatus      string
		addPatterns    string
		removePatterns string
		addTags        string
		removeTags     string
		scoreMin       float64
		scoreMax       float64
		statusFilter   string
		dryRun         bool
		yes            bool
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update multiple ideas in batch",
		Long: `Update fields for multiple ideas based on filtering criteria.

Supports updating:
- Status (active/archived/deleted)
- Patterns (add/remove)
- Tags (add/remove)

Examples:
  # Archive all low-scoring ideas
  telos bulk update --score-max 3.0 --set-status archived

  # Add pattern to all ideas in score range
  telos bulk update --score-min 7.0 --add-patterns "high-value"

  # Remove obsolete pattern from all ideas
  telos bulk update --remove-patterns "old-pattern"

  # Dry-run to preview changes
  telos bulk update --score-max 3.0 --set-status archived --dry-run

  # Add tag to archived ideas
  telos bulk update --status archived --add-tags "reviewed"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBulkUpdate(bulkUpdateOptions{
				setStatus:      setStatus,
				addPatterns:    bulk.SplitCommaSeparated(addPatterns),
				removePatterns: bulk.SplitCommaSeparated(removePatterns),
				addTags:        bulk.SplitCommaSeparated(addTags),
				removeTags:     bulk.SplitCommaSeparated(removeTags),
				scoreMin:       scoreMin,
				scoreMax:       scoreMax,
				statusFilter:   statusFilter,
				dryRun:         dryRun,
				yes:            yes,
			})
		},
	}

	// Update operations
	cmd.Flags().StringVar(&setStatus, "set-status", "", "Set status (active|archived|deleted)")
	cmd.Flags().StringVar(&addPatterns, "add-patterns", "", "Add patterns (comma-separated)")
	cmd.Flags().StringVar(&removePatterns, "remove-patterns", "", "Remove patterns (comma-separated)")
	cmd.Flags().StringVar(&addTags, "add-tags", "", "Add tags (comma-separated)")
	cmd.Flags().StringVar(&removeTags, "remove-tags", "", "Remove tags (comma-separated)")

	// Filters
	cmd.Flags().Float64Var(&scoreMin, "score-min", 0, "Minimum score filter")
	cmd.Flags().Float64Var(&scoreMax, "score-max", 10, "Maximum score filter")
	cmd.Flags().StringVar(&statusFilter, "status", "", "Filter by status")

	// Options
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without applying")
	cmd.Flags().BoolVar(&yes, "yes", false, "Auto-confirm (skip confirmation prompt)")

	return cmd
}

type bulkUpdateOptions struct {
	setStatus      string
	addPatterns    []string
	removePatterns []string
	addTags        []string
	removeTags     []string
	scoreMin       float64
	scoreMax       float64
	statusFilter   string
	dryRun         bool
	yes            bool
}

func runBulkUpdate(opts bulkUpdateOptions) error {
	// Validate that at least one update operation is specified
	if opts.setStatus == "" &&
		len(opts.addPatterns) == 0 &&
		len(opts.removePatterns) == 0 &&
		len(opts.addTags) == 0 &&
		len(opts.removeTags) == 0 {
		return fmt.Errorf("no updates specified (use --set-status, --add-patterns, --remove-patterns, --add-tags, or --remove-tags)")
	}

	// Validate status value if provided
	if opts.setStatus != "" {
		validStatuses := []string{"active", "archived", "deleted"}
		if !bulk.Contains(validStatuses, opts.setStatus) {
			return fmt.Errorf("invalid status: %s (must be one of: %s)",
				opts.setStatus, strings.Join(validStatuses, ", "))
		}
	}

	// Safety check for tests
	if ctx == nil || ctx.Repository == nil {
		return fmt.Errorf("CLI context not initialized")
	}

	// Build filter
	minScorePtr := &opts.scoreMin
	maxScorePtr := &opts.scoreMax
	limitPtr := new(int)
	*limitPtr = 10000 // High limit for bulk operations

	ideas, err := ctx.Repository.List(database.ListOptions{
		Status:   opts.statusFilter,
		MinScore: minScorePtr,
		MaxScore: maxScorePtr,
		Limit:    limitPtr,
		OrderBy:  "final_score DESC",
	})
	if err != nil {
		return fmt.Errorf("failed to find ideas: %w", err)
	}

	if len(ideas) == 0 {
		fmt.Println("📭 No ideas match the criteria.")
		return nil
	}

	// Show preview
	fmt.Printf("🎯 Found %s ideas to update:\n\n",
		color.CyanString("%d", len(ideas)))
	fmt.Println("Filters applied:")
	fmt.Printf("  Score range: %.1f - %.1f\n", opts.scoreMin, opts.scoreMax)
	if opts.statusFilter != "" {
		fmt.Printf("  Status: %s\n", opts.statusFilter)
	}
	fmt.Println()

	fmt.Println("Updates to apply:")
	if opts.setStatus != "" {
		fmt.Printf("  - Set status: %s\n", color.GreenString(opts.setStatus))
	}
	if len(opts.addPatterns) > 0 {
		fmt.Printf("  - Add patterns: %s\n", color.GreenString(strings.Join(opts.addPatterns, ", ")))
	}
	if len(opts.removePatterns) > 0 {
		fmt.Printf("  - Remove patterns: %s\n", color.YellowString(strings.Join(opts.removePatterns, ", ")))
	}
	if len(opts.addTags) > 0 {
		fmt.Printf("  - Add tags: %s\n", color.GreenString(strings.Join(opts.addTags, ", ")))
	}
	if len(opts.removeTags) > 0 {
		fmt.Printf("  - Remove tags: %s\n", color.YellowString(strings.Join(opts.removeTags, ", ")))
	}
	fmt.Println()

	if opts.dryRun {
		if _, err := infoColor.Println("🔍 DRY RUN - Showing affected ideas and changes:"); err != nil {
			log.Warn().Err(err).Msg("failed to print message")
		}
		for i, idea := range ideas {
			if i >= 10 {
				fmt.Printf("\n... and %d more ideas\n", len(ideas)-10)
				break
			}
			fmt.Printf("\n%d. [%s] %s\n", i+1, idea.ID[:8], truncate(idea.Content, 60))
			fmt.Printf("   Current - Score: %.1f, Status: %s\n", idea.FinalScore, idea.Status)

			if opts.setStatus != "" && idea.Status != opts.setStatus {
				fmt.Printf("   %s Status change: %s → %s\n",
					color.CyanString("→"), idea.Status, opts.setStatus)
			}
			if len(opts.addPatterns) > 0 {
				newPatterns := bulk.AddUniqueStrings(idea.Patterns, opts.addPatterns)
				if len(newPatterns) > len(idea.Patterns) {
					fmt.Printf("   %s Patterns: %v → %v\n",
						color.CyanString("→"), idea.Patterns, newPatterns)
				}
			}
			if len(opts.removePatterns) > 0 {
				newPatterns := bulk.RemoveStrings(idea.Patterns, opts.removePatterns)
				if len(newPatterns) < len(idea.Patterns) {
					fmt.Printf("   %s Patterns: %v → %v\n",
						color.CyanString("→"), idea.Patterns, newPatterns)
				}
			}
			if len(opts.addTags) > 0 {
				newTags := bulk.AddUniqueStrings(idea.Tags, opts.addTags)
				if len(newTags) > len(idea.Tags) {
					fmt.Printf("   %s Tags: %v → %v\n",
						color.CyanString("→"), idea.Tags, newTags)
				}
			}
			if len(opts.removeTags) > 0 {
				newTags := bulk.RemoveStrings(idea.Tags, opts.removeTags)
				if len(newTags) < len(idea.Tags) {
					fmt.Printf("   %s Tags: %v → %v\n",
						color.CyanString("→"), idea.Tags, newTags)
				}
			}
		}
		return nil
	}

	// Confirm
	if !opts.yes && !confirm(fmt.Sprintf("Update %d ideas?", len(ideas))) {
		fmt.Println("❌ Cancelled")
		return nil
	}

	// Apply updates
	updated := 0
	unchanged := 0
	failed := 0
	errors := make([]string, 0)

	service := bulk.NewService(ctx.Repository)
	updateOpts := bulk.UpdateOptions{
		SetStatus:      opts.setStatus,
		AddPatterns:    opts.addPatterns,
		RemovePatterns: opts.removePatterns,
		AddTags:        opts.addTags,
		RemoveTags:     opts.removeTags,
	}

	for i, idea := range ideas {
		// Apply updates using service
		modified := service.ApplyUpdates(idea, updateOpts)

		// Only save if something actually changed
		if modified {
			if err := ctx.Repository.Update(idea); err != nil {
				failed++
				errors = append(errors, fmt.Sprintf("%s: %v", idea.ID[:8], err))
				continue
			}
			updated++
		} else {
			unchanged++
		}

		// Show progress for large batches
		if len(ideas) > 10 && (i+1)%10 == 0 {
			fmt.Printf("  Progress: %d/%d processed\n", i+1, len(ideas))
		}
	}

	fmt.Printf("\n%s Update complete:\n", successColor.Sprint("✅"))
	fmt.Printf("  ✓ Updated: %s\n", color.GreenString("%d", updated))
	if unchanged > 0 {
		fmt.Printf("  - Unchanged: %s (no modifications needed)\n", color.CyanString("%d", unchanged))
	}
	if failed > 0 {
		fmt.Printf("  ✗ Failed: %s\n", errorColor.Sprint(failed))
		if len(errors) > 0 && len(errors) <= 10 {
			fmt.Println("\nErrors:")
			for _, errMsg := range errors {
				fmt.Printf("  - %s\n", errMsg)
			}
		}
	}

	return nil
}

// Helper functions

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func confirm(prompt string) bool {
	fmt.Printf("%s [y/N]: ", prompt)
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		log.Warn().Err(err).Msg("failed to read user input")
	}
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

func newBulkAnalyzeCommand() *cobra.Command {
	var (
		scoreMin  float64
		scoreMax  float64
		status    string
		olderThan string
		dryRun    bool
		provider  string
		yes       bool
	)

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Re-analyze multiple ideas with updated criteria",
		Long: `Re-analyze multiple ideas using current telos and LLM provider.

This is useful when:
- Your telos has changed
- You want to use a different LLM provider
- You've improved the analysis algorithm
- You want to refresh old analyses

Examples:
  # Re-analyze all low-scoring ideas
  telos bulk analyze --score-max 5.0

  # Re-analyze ideas from last month
  telos bulk analyze --older-than 30d

  # Re-analyze with specific provider
  telos bulk analyze --provider ollama

  # Dry-run to see what would be analyzed
  telos bulk analyze --score-max 5.0 --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBulkAnalyze(bulkAnalyzeOptions{
				scoreMin:  scoreMin,
				scoreMax:  scoreMax,
				status:    status,
				olderThan: olderThan,
				dryRun:    dryRun,
				provider:  provider,
				yes:       yes,
			})
		},
	}

	cmd.Flags().Float64Var(&scoreMin, "score-min", 0, "Minimum score (inclusive)")
	cmd.Flags().Float64Var(&scoreMax, "score-max", 10, "Maximum score (inclusive)")
	cmd.Flags().StringVar(&status, "status", "active", "Filter by status (active|archived|deleted)")
	cmd.Flags().StringVar(&olderThan, "older-than", "", "Re-analyze ideas older than duration (e.g., 30d, 6h)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be analyzed without making changes")
	cmd.Flags().StringVar(&provider, "provider", "", "LLM provider to use (ollama|claude|openai|rule_based)")
	cmd.Flags().BoolVar(&yes, "yes", false, "Auto-confirm (skip confirmation prompt)")

	return cmd
}

// bulkAnalyzeOptions contains options for bulk analysis
type bulkAnalyzeOptions struct {
	scoreMin  float64
	scoreMax  float64
	status    string
	olderThan string
	dryRun    bool
	provider  string
	yes       bool
}

// runBulkAnalyze performs bulk re-analysis of ideas
func runBulkAnalyze(opts bulkAnalyzeOptions) error {
	// Parse olderThan duration if specified
	var cutoffTime time.Time
	if opts.olderThan != "" {
		duration, err := parseDuration(opts.olderThan)
		if err != nil {
			return fmt.Errorf("invalid duration: %w", err)
		}
		cutoffTime = time.Now().UTC().Add(-duration)
	}

	// Build filter criteria
	minScorePtr := &opts.scoreMin
	maxScorePtr := &opts.scoreMax
	limit := 1000 // Safety limit

	ideas, err := ctx.Repository.List(database.ListOptions{
		Status:   opts.status,
		MinScore: minScorePtr,
		MaxScore: maxScorePtr,
		Limit:    &limit,
		OrderBy:  "created_at ASC",
	})
	if err != nil {
		return fmt.Errorf("failed to find ideas: %w", err)
	}

	// Filter by age if specified
	if !cutoffTime.IsZero() {
		ideas = filterByAge(ideas, cutoffTime)
	}

	if len(ideas) == 0 {
		fmt.Println("📭 No ideas match the criteria.")
		return nil
	}

	// Show summary
	fmt.Printf("🔍 Found %s ideas matching criteria:\n",
		color.CyanString("%d", len(ideas)))
	fmt.Printf("  Score range: %.1f - %.1f\n", opts.scoreMin, opts.scoreMax)
	if opts.status != "" {
		fmt.Printf("  Status: %s\n", opts.status)
	}
	if opts.olderThan != "" {
		fmt.Printf("  Older than: %s\n", opts.olderThan)
	}
	fmt.Println()

	if opts.dryRun {
		if _, err := infoColor.Println("🔍 DRY RUN - No changes will be made"); err != nil {
			log.Warn().Err(err).Msg("failed to print message")
		}
		fmt.Println()
		for i, idea := range ideas {
			if i < 10 { // Show first 10
				age := time.Since(idea.CreatedAt).Hours() / 24
				fmt.Printf("%d. [%s] %s (score: %.1f, age: %.0fd)\n",
					i+1, idea.ID[:8], truncate(idea.Content, 60), idea.FinalScore, age)
			}
		}
		if len(ideas) > 10 {
			fmt.Printf("... and %d more\n", len(ideas)-10)
		}
		return nil
	}

	// Confirm with user
	if !opts.yes && !confirm(fmt.Sprintf("Re-analyze %d ideas?", len(ideas))) {
		fmt.Println("❌ Cancelled")
		return nil
	}

	// Create LLM manager
	llmManager := createLLMManager()

	// Set provider if specified
	if opts.provider != "" {
		if err := llmManager.SetPrimaryProvider(opts.provider); err != nil {
			return fmt.Errorf("failed to set provider: %w", err)
		}
		if _, err := infoColor.Printf("🤖 Using provider: %s\n", opts.provider); err != nil {
			log.Warn().Err(err).Msg("failed to print message")
		}
	} else {
		primaryProvider := llmManager.GetPrimaryProvider()
		if primaryProvider != nil {
			if _, err := infoColor.Printf("🤖 Using provider: %s\n", primaryProvider.Name()); err != nil {
				log.Warn().Err(err).Msg("failed to print message")
			}
		}
	}
	fmt.Println()

	// Analyze ideas with progress tracking
	successful := 0
	failed := 0
	errors := make([]string, 0)

	for i, idea := range ideas {
		// Show progress
		progress := float64(i+1) / float64(len(ideas)) * 100
		fmt.Printf("\r[%d/%d] 🔄 Analyzing ideas... %.1f%%",
			i+1, len(ideas), progress)

		// Re-analyze using LLM
		result, err := llmManager.AnalyzeWithTelos(idea.Content, ctx.Telos)
		if err != nil {
			failed++
			errors = append(errors, fmt.Sprintf("%s: %v", idea.ID[:8], err))
			continue
		}

		// Detect patterns
		detectedPatterns := ctx.Detector.DetectPatterns(idea.Content)
		patternStrings := make([]string, len(detectedPatterns))
		for j, p := range detectedPatterns {
			patternStrings[j] = fmt.Sprintf("%s: %s", p.Name, p.Description)
		}

		// Format explanations as JSON for storage
		analysisDetails := ""
		if len(result.Explanations) > 0 {
			detailsMap := map[string]interface{}{
				"explanations": result.Explanations,
				"provider":     result.Provider,
				"scores": map[string]float64{
					"mission_alignment": result.Scores.MissionAlignment,
					"anti_challenge":    result.Scores.AntiChallenge,
					"strategic_fit":     result.Scores.StrategicFit,
				},
			}
			detailsBytes, _ := json.Marshal(detailsMap)
			analysisDetails = string(detailsBytes)
		} else {
			analysisDetails = result.Recommendation
		}

		// Update idea
		idea.FinalScore = result.FinalScore
		idea.Patterns = patternStrings
		idea.Recommendation = result.Recommendation
		idea.AnalysisDetails = analysisDetails

		if err := ctx.Repository.Update(idea); err != nil {
			failed++
			errors = append(errors, fmt.Sprintf("%s: failed to save: %v", idea.ID[:8], err))
			continue
		}

		successful++
	}

	fmt.Println() // New line after progress
	fmt.Println()

	// Show summary
	if _, err := successColor.Printf("✅ Re-analysis complete:\n"); err != nil {
		log.Warn().Err(err).Msg("failed to print success message")
	}
	fmt.Printf("  ✓ Successful: %d\n", successful)
	if failed > 0 {
		if _, err := warningColor.Printf("  ✗ Failed: %d\n", failed); err != nil {
			log.Warn().Err(err).Msg("failed to print failed count")
		}
		if len(errors) > 0 && len(errors) <= 10 {
			fmt.Println("\nErrors:")
			for _, errMsg := range errors {
				fmt.Printf("  - %s\n", errMsg)
			}
		} else if len(errors) > 10 {
			fmt.Printf("\n  (Showing first 10 of %d errors)\n", len(errors))
			for i := 0; i < 10; i++ {
				fmt.Printf("  - %s\n", errors[i])
			}
		}
	}

	return nil
}

// parseDuration parses duration strings like "30d", "6h", "45m"
func parseDuration(s string) (time.Duration, error) {
	if len(s) < 2 {
		return 0, fmt.Errorf("invalid duration format")
	}

	value := s[:len(s)-1]
	unit := s[len(s)-1:]

	var multiplier time.Duration
	switch unit {
	case "d":
		multiplier = 24 * time.Hour
	case "h":
		multiplier = time.Hour
	case "m":
		multiplier = time.Minute
	case "s":
		multiplier = time.Second
	default:
		// Fallback to standard Go duration parsing
		return time.ParseDuration(s)
	}

	var numValue int
	n, err := fmt.Sscanf(value, "%d", &numValue)
	if err != nil || n != 1 {
		return 0, fmt.Errorf("invalid duration value: %w", err)
	}

	return time.Duration(numValue) * multiplier, nil
}

// createLLMManager creates and configures an LLM manager
func createLLMManager() *llm.Manager {
	return llm.NewManager(nil)
}
