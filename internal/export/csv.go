// Package export provides functionality for exporting ideas to various formats including CSV.
package export

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/ryacub/telos-idea-matrix/internal/models"
)

// ExportCSV writes ideas to a CSV file.
func ExportCSV(ideas []*models.Idea, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Warn().Err(err).Msg("failed to close file")
		}
	}()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"ID",
		"Content",
		"RawScore",
		"FinalScore",
		"Patterns",
		"Recommendation",
		"AnalysisDetails",
		"CreatedAt",
		"Status",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	// Write rows
	for _, idea := range ideas {
		// Join patterns with semicolon
		patterns := strings.Join(idea.Patterns, ",")

		row := []string{
			idea.ID,
			idea.Content,
			strconv.FormatFloat(idea.RawScore, 'f', 2, 64),
			strconv.FormatFloat(idea.FinalScore, 'f', 2, 64),
			patterns,
			idea.Recommendation,
			idea.AnalysisDetails,
			idea.CreatedAt.Format(time.RFC3339),
			idea.Status,
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("write row: %w", err)
		}
	}

	return writer.Error()
}

// ImportCSV reads ideas from a CSV file.
func ImportCSV(filename string) ([]*models.Idea, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Warn().Err(err).Msg("failed to close file")
		}
	}()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read csv: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("csv file is empty")
	}

	// Skip header row
	if len(records) == 1 {
		// Only header, return empty slice
		return []*models.Idea{}, nil
	}

	ideas := make([]*models.Idea, 0, len(records)-1)

	for i, record := range records[1:] {
		if len(record) < 9 {
			return nil, fmt.Errorf("row %d: invalid format, expected 9 columns, got %d", i+2, len(record))
		}

		// Parse scores (with default 0.0 on error)
		rawScore, _ := strconv.ParseFloat(record[2], 64)
		finalScore, _ := strconv.ParseFloat(record[3], 64)

		// Parse patterns (split by comma)
		var patterns []string
		if record[4] != "" {
			patterns = strings.Split(record[4], ",")
		}

		// Parse timestamp
		createdAt, err := time.Parse(time.RFC3339, record[7])
		if err != nil {
			// Default to current time if parsing fails
			createdAt = time.Now().UTC()
		}

		idea := &models.Idea{
			ID:              record[0],
			Content:         record[1],
			RawScore:        rawScore,
			FinalScore:      finalScore,
			Patterns:        patterns,
			Recommendation:  record[5],
			AnalysisDetails: record[6],
			CreatedAt:       createdAt,
			Status:          record[8],
		}

		ideas = append(ideas, idea)
	}

	return ideas, nil
}
