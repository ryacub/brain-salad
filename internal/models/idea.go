package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Idea represents a captured idea with analysis.
// Maps to StoredIdea in Rust implementation.
type Idea struct {
	ID               string    `json:"id" db:"id"`
	Content          string    `json:"content" db:"content"`
	RawScore         float64   `json:"raw_score,omitempty" db:"raw_score"`
	FinalScore       float64   `json:"final_score,omitempty" db:"final_score"`
	Patterns         []string  `json:"patterns,omitempty" db:"patterns"`
	Tags             []string  `json:"tags,omitempty" db:"tags"`
	Recommendation   string    `json:"recommendation,omitempty" db:"recommendation"`
	AnalysisDetails  string    `json:"analysis_details,omitempty" db:"analysis_details"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	ReviewedAt       *time.Time `json:"reviewed_at,omitempty" db:"reviewed_at"`
	Status           string    `json:"status" db:"status"`
	Title            string    `json:"title,omitempty"` // For compatibility
	Analysis         *Analysis `json:"analysis,omitempty"` // Full analysis object (not stored in DB)
}

// NewIdea creates a new Idea with generated ID and current timestamp.
func NewIdea(content string) *Idea {
	return &Idea{
		ID:        uuid.New().String(),
		Content:   content,
		Status:    "active",
		CreatedAt: time.Now().UTC(),
	}
}

// Validate validates the idea.
func (i *Idea) Validate() error {
	// Validate title if present (used in some contexts)
	if i.Title != "" {
		if len(i.Title) < 3 {
			return errors.New("title must be at least 3 characters")
		}
		if len(i.Title) > 200 {
			return errors.New("title must be at most 200 characters")
		}
	}

	// Validate content if title is empty
	if i.Title == "" && i.Content == "" {
		return errors.New("title or content is required")
	}

	// Validate status
	validStatuses := map[string]bool{
		"active":   true,
		"archived": true,
		"deleted":  true,
	}

	if i.Status != "" && !validStatuses[i.Status] {
		return errors.New("invalid status: must be one of 'active', 'archived', 'deleted'")
	}

	return nil
}

// IdeaStatus represents the status of an idea.
type IdeaStatus string

const (
	// StatusActive indicates an active idea.
	StatusActive IdeaStatus = "active"
	// StatusArchived indicates an archived idea.
	StatusArchived IdeaStatus = "archived"
	// StatusDeleted indicates a deleted idea.
	StatusDeleted IdeaStatus = "deleted"
)

// String returns the string representation of the status.
func (s IdeaStatus) String() string {
	return string(s)
}
