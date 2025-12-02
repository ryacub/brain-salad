package bulk

import (
	"github.com/ryacub/telos-idea-matrix/internal/database"
	"github.com/ryacub/telos-idea-matrix/internal/llm"
	"github.com/ryacub/telos-idea-matrix/internal/models"
	"github.com/spf13/cobra"
)

const (
	// FormatJSON represents JSON format for export/import
	FormatJSON = "json"
	// FormatCSV represents CSV format for export/import
	FormatCSV = "csv"
)

// CLIContext represents the shared CLI dependencies for bulk operations
type CLIContext struct {
	Repository *database.Repository
	Telos      *models.Telos
	LLMManager *llm.Manager
}

// NewBulkCommand creates the bulk operations command
func NewBulkCommand(getContext func() *CLIContext) *cobra.Command {
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

	// Add subcommands with context getter
	cmd.AddCommand(NewAnalyzeCommand(getContext))
	cmd.AddCommand(NewUpdateCommand(getContext))
	cmd.AddCommand(NewTagCommand(getContext))
	cmd.AddCommand(NewArchiveCommand(getContext))
	cmd.AddCommand(NewDeleteCommand(getContext))
	cmd.AddCommand(NewImportCommand(getContext))
	cmd.AddCommand(NewExportCommand(getContext))

	return cmd
}
