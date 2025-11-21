package bulk

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/rayyacub/telos-idea-matrix/internal/bulk"
	"github.com/rayyacub/telos-idea-matrix/internal/cliutil"
	"github.com/rayyacub/telos-idea-matrix/internal/database"
	"github.com/rayyacub/telos-idea-matrix/internal/export"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// NewExportCommand creates the bulk export command
func NewExportCommand(getContext func() *CLIContext) *cobra.Command {
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
			ctx := getContext()
			if ctx == nil {
				return fmt.Errorf("CLI context not initialized")
			}

			filename := args[0]

			// Create service once
			service := bulk.NewService(ctx.Repository)

			// Auto-detect format from extension if not specified
			if format == "" {
				ext := strings.ToLower(filepath.Ext(filename))
				if ext == ".json" {
					format = FormatJSON
				} else {
					format = FormatCSV
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
				ideas = service.FilterBySearch(ideas, search)
			}

			if len(ideas) == 0 {
				fmt.Println("📭 No ideas match your criteria for export.")
				return nil
			}

			// Export based on format
			switch format {
			case FormatJSON:
				err = export.ExportJSON(ideas, filename, pretty)
			case FormatCSV:
				err = export.ExportCSV(ideas, filename)
			default:
				return fmt.Errorf("unsupported format: %s (use 'csv' or 'json')", format)
			}

			if err != nil {
				return fmt.Errorf("failed to export: %w", err)
			}

			if _, err := cliutil.SuccessColor.Printf("✅ Exported %d ideas to '%s' (%s format)\n",
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
