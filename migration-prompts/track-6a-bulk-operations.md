# Track 6A: Bulk Operations

**Phase**: 6 - Advanced CLI Features
**Estimated Time**: 10-12 hours
**Dependencies**: None
**Can Run in Parallel**: Yes (with 5D, 6B)

---

## Mission

You are implementing bulk operations for managing multiple ideas at once in the Telos Idea Matrix Go application, following Test-Driven Development (TDD).

## Context

- The Rust implementation has bulk operations in `src/commands/bulk.rs`
- Need bulk tag, archive, delete with filters
- CSV import/export functionality
- Confirmation prompts to prevent accidental data loss

## Reference Implementation

Review `/home/user/brain-salad/src/commands/bulk.rs` for:
- Bulk tag/archive/delete logic
- CSV import/export
- Filter combinations (score, age, search)
- Confirmation UX

## Your Task

Implement bulk operations using strict TDD methodology.

## Directory Structure

Create files in `go/internal/cli/` and `go/internal/export/`:
- `cli/bulk.go` - Bulk command implementations
- `cli/bulk_test.go` - Integration tests
- `export/csv.go` - CSV import/export
- `export/json.go` - JSON export
- `export/csv_test.go` - CSV tests

## TDD Workflow (RED → GREEN → REFACTOR)

### STEP 1 - RED PHASE (Write Failing Tests)

Create `go/internal/cli/bulk_test.go`:
- `TestBulkTag_WithFilters()`
- `TestBulkArchive_WithFilters()`
- `TestBulkDelete_WithConfirmation()`
- `TestBulkImport_FromCSV()`
- `TestBulkExport_ToCSV()`

Create `go/internal/export/csv_test.go`:
- `TestCSV_Export()`
- `TestCSV_Import()`
- `TestCSV_HandleMalformed()`

Run: `go test ./internal/cli ./internal/export -v`
Expected: **ALL TESTS FAIL**

### STEP 2 - GREEN PHASE (Implement)

#### A. Implement `go/internal/cli/bulk.go`:

```go
package cli

import (
    "fmt"
    
    "github.com/fatih/color"
    "github.com/spf13/cobra"
)

func NewBulkCommand(ctx *CLIContext) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "bulk",
        Short: "Bulk operations on multiple ideas",
        Long:  "Perform bulk operations: tag, archive, delete, import, export",
    }
    
    cmd.AddCommand(newBulkTagCommand(ctx))
    cmd.AddCommand(newBulkArchiveCommand(ctx))
    cmd.AddCommand(newBulkDeleteCommand(ctx))
    cmd.AddCommand(newBulkImportCommand(ctx))
    cmd.AddCommand(newBulkExportCommand(ctx))
    
    return cmd
}

func newBulkTagCommand(ctx *CLIContext) *cobra.Command {
    var minScore float64
    var search string
    var limit int
    
    cmd := &cobra.Command{
        Use:   "tag <tag-name>",
        Short: "Add tag to multiple ideas",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            tagName := args[0]
            
            // Find matching ideas
            ideas, err := ctx.Repo.List(minScore, "", limit)
            if err != nil {
                return err
            }
            
            // Filter by search if provided
            if search != "" {
                ideas = filterBySearch(ideas, search)
            }
            
            if len(ideas) == 0 {
                fmt.Println("No ideas match criteria")
                return nil
            }
            
            // Show preview
            fmt.Printf("Will tag %d ideas with '%s':\n", len(ideas), tagName)
            for _, idea := range ideas {
                fmt.Printf("  - %s (score: %.1f)\n", idea.Title, idea.Score)
            }
            
            // Confirm
            if !confirm("Proceed?") {
                fmt.Println("Cancelled")
                return nil
            }
            
            // Apply tags
            for _, idea := range ideas {
                // Add tag logic here
            }
            
            color.Green("✓ Tagged %d ideas", len(ideas))
            return nil
        },
    }
    
    cmd.Flags().Float64Var(&minScore, "min-score", 0, "Minimum score threshold")
    cmd.Flags().StringVar(&search, "search", "", "Search term")
    cmd.Flags().IntVar(&limit, "limit", 100, "Maximum ideas to process")
    
    return cmd
}

func newBulkArchiveCommand(ctx *CLIContext) *cobra.Command {
    var olderThan int
    var maxScore float64
    var dryRun bool
    
    cmd := &cobra.Command{
        Use:   "archive",
        Short: "Archive multiple old/low-scoring ideas",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Implementation similar to bulk tag
            // Filter by age and score
            // Archive matching ideas
            return nil
        },
    }
    
    cmd.Flags().IntVar(&olderThan, "older-than", 0, "Archive ideas older than N days")
    cmd.Flags().Float64Var(&maxScore, "max-score", 10.0, "Maximum score")
    cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be archived")
    
    return cmd
}

func newBulkDeleteCommand(ctx *CLIContext) *cobra.Command {
    // Similar to archive but with stricter confirmation
    // Require explicit --confirm flag
    return &cobra.Command{
        Use:   "delete",
        Short: "Permanently delete multiple ideas",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Implementation
            return nil
        },
    }
}
```

#### B. Implement `go/internal/export/csv.go`:

```go
package export

import (
    "encoding/csv"
    "fmt"
    "os"
    "strconv"
    "time"
    
    "github.com/rayyacub/telos-idea-matrix/internal/models"
)

// ExportCSV writes ideas to CSV file
func ExportCSV(ideas []*models.Idea, filename string) error {
    file, err := os.Create(filename)
    if err != nil {
        return fmt.Errorf("create file: %w", err)
    }
    defer file.Close()
    
    writer := csv.NewWriter(file)
    defer writer.Flush()
    
    // Write header
    header := []string{"ID", "Title", "Content", "Score", "Status", "CreatedAt"}
    if err := writer.Write(header); err != nil {
        return err
    }
    
    // Write rows
    for _, idea := range ideas {
        row := []string{
            idea.ID,
            idea.Title,
            idea.Content,
            strconv.FormatFloat(idea.Score, 'f', 2, 64),
            idea.Status,
            idea.CreatedAt.Format(time.RFC3339),
        }
        if err := writer.Write(row); err != nil {
            return err
        }
    }
    
    return nil
}

// ImportCSV reads ideas from CSV file
func ImportCSV(filename string) ([]*models.Idea, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, fmt.Errorf("open file: %w", err)
    }
    defer file.Close()
    
    reader := csv.NewReader(file)
    records, err := reader.ReadAll()
    if err != nil {
        return nil, fmt.Errorf("read csv: %w", err)
    }
    
    if len(records) < 2 {
        return nil, fmt.Errorf("csv file is empty")
    }
    
    // Skip header
    ideas := make([]*models.Idea, 0, len(records)-1)
    for i, record := range records[1:] {
        if len(record) < 6 {
            return nil, fmt.Errorf("row %d: invalid format", i+2)
        }
        
        score, _ := strconv.ParseFloat(record[3], 64)
        createdAt, _ := time.Parse(time.RFC3339, record[5])
        
        idea := &models.Idea{
            ID:        record[0],
            Title:     record[1],
            Content:   record[2],
            Score:     score,
            Status:    record[4],
            CreatedAt: createdAt,
        }
        ideas = append(ideas, idea)
    }
    
    return ideas, nil
}
```

#### C. Implement `go/internal/export/json.go`:

```go
package export

import (
    "encoding/json"
    "fmt"
    "os"
    
    "github.com/rayyacub/telos-idea-matrix/internal/models"
)

// ExportJSON writes ideas to JSON file
func ExportJSON(ideas []*models.Idea, filename string, pretty bool) error {
    file, err := os.Create(filename)
    if err != nil {
        return fmt.Errorf("create file: %w", err)
    }
    defer file.Close()
    
    encoder := json.NewEncoder(file)
    if pretty {
        encoder.SetIndent("", "  ")
    }
    
    return encoder.Encode(ideas)
}
```

Run: `go test ./internal/cli ./internal/export -v`
Expected: **ALL TESTS PASS**

### STEP 3 - REFACTOR PHASE

- Add progress bars for bulk operations
- Optimize bulk database operations (batch inserts/updates)
- Add undo functionality (export backup before bulk delete)
- Extract filter logic into reusable functions

## Success Criteria

- ✅ All tests pass with >85% coverage
- ✅ Confirmation prompts prevent accidental deletion
- ✅ CSV import handles 1000+ rows
- ✅ Matches Rust `src/commands/bulk.rs` functionality

## Validation

```bash
# Bulk operations
tm bulk tag "high-priority" --min-score 8.0
tm bulk archive --older-than 90 --max-score 5.0 --dry-run
tm bulk delete --older-than 180 --confirm

# Import/Export
tm bulk export ideas.csv
tm bulk export ideas.json --pretty
tm bulk import ideas.csv
```

## Deliverables

- `go/internal/cli/bulk.go`
- `go/internal/cli/bulk_test.go`
- `go/internal/export/csv.go`
- `go/internal/export/json.go`
- `go/internal/export/csv_test.go`
