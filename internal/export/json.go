package export

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/ryacub/telos-idea-matrix/internal/models"
)

// ExportJSON writes ideas to a JSON file.
func ExportJSON(ideas []*models.Idea, filename string, pretty bool) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Warn().Err(err).Msg("failed to close file")
		}
	}()

	encoder := json.NewEncoder(file)
	if pretty {
		encoder.SetIndent("", "  ")
	}

	if err := encoder.Encode(ideas); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}

	return nil
}
