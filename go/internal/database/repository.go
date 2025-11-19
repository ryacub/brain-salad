package database

import (
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Repository handles database operations for ideas.
type Repository struct {
	db *sql.DB
}

// ListOptions defines options for listing ideas.
type ListOptions struct {
	Status   string   // Filter by status (e.g., "active", "archived")
	MinScore *float64 // Filter by minimum score
	MaxScore *float64 // Filter by maximum score
	OrderBy  string   // Order by clause (e.g., "final_score DESC")
	Limit    *int     // Limit number of results
	Offset   *int     // Offset for pagination
}

// NewRepository creates a new database repository and runs migrations.
func NewRepository(dbPath string) (*Repository, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	// Open database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	repo := &Repository{db: db}

	// Run migrations
	if err := repo.runMigrations(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return repo, nil
}

// runMigrations applies all migration files.
func (r *Repository) runMigrations() error {
	// Read all migration files
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		// Read migration file
		content, err := migrationsFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", entry.Name(), err)
		}

		// Execute migration
		if _, err := r.db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", entry.Name(), err)
		}
	}

	return nil
}

// Create saves a new idea to the database.
func (r *Repository) Create(idea *models.Idea) error {
	if idea == nil {
		return errors.New("idea cannot be nil")
	}

	// Validate idea
	if err := idea.Validate(); err != nil {
		return fmt.Errorf("invalid idea: %w", err)
	}

	// Serialize patterns to JSON
	patternsJSON, err := json.Marshal(idea.Patterns)
	if err != nil {
		return fmt.Errorf("failed to serialize patterns: %w", err)
	}

	// Format timestamps as RFC3339
	createdAt := idea.CreatedAt.Format(time.RFC3339)
	var reviewedAt *string
	if idea.ReviewedAt != nil {
		t := idea.ReviewedAt.Format(time.RFC3339)
		reviewedAt = &t
	}

	query := `
		INSERT INTO ideas (
			id, content, raw_score, final_score, patterns,
			recommendation, analysis_details, created_at, reviewed_at, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.Exec(
		query,
		idea.ID,
		idea.Content,
		idea.RawScore,
		idea.FinalScore,
		string(patternsJSON),
		idea.Recommendation,
		idea.AnalysisDetails,
		createdAt,
		reviewedAt,
		idea.Status,
	)

	if err != nil {
		return fmt.Errorf("failed to insert idea: %w", err)
	}

	return nil
}

// GetByID retrieves an idea by its ID.
func (r *Repository) GetByID(id string) (*models.Idea, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	query := `
		SELECT id, content, raw_score, final_score, patterns,
		       recommendation, analysis_details, created_at, reviewed_at, status
		FROM ideas
		WHERE id = ?
	`

	var idea models.Idea
	var patternsJSON string
	var createdAt string
	var reviewedAt sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&idea.ID,
		&idea.Content,
		&idea.RawScore,
		&idea.FinalScore,
		&patternsJSON,
		&idea.Recommendation,
		&idea.AnalysisDetails,
		&createdAt,
		&reviewedAt,
		&idea.Status,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("idea not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query idea: %w", err)
	}

	// Parse patterns JSON
	if patternsJSON != "" && patternsJSON != "null" {
		if err := json.Unmarshal([]byte(patternsJSON), &idea.Patterns); err != nil {
			return nil, fmt.Errorf("failed to parse patterns: %w", err)
		}
	}

	// Parse timestamps
	if parsedTime, err := time.Parse(time.RFC3339, createdAt); err == nil {
		idea.CreatedAt = parsedTime
	}

	if reviewedAt.Valid {
		if parsedTime, err := time.Parse(time.RFC3339, reviewedAt.String); err == nil {
			idea.ReviewedAt = &parsedTime
		}
	}

	return &idea, nil
}

// Update updates an existing idea in the database.
func (r *Repository) Update(idea *models.Idea) error {
	if idea == nil {
		return errors.New("idea cannot be nil")
	}

	// Validate idea
	if err := idea.Validate(); err != nil {
		return fmt.Errorf("invalid idea: %w", err)
	}

	// Serialize patterns to JSON
	patternsJSON, err := json.Marshal(idea.Patterns)
	if err != nil {
		return fmt.Errorf("failed to serialize patterns: %w", err)
	}

	// Format timestamps
	var reviewedAt *string
	if idea.ReviewedAt != nil {
		t := idea.ReviewedAt.Format(time.RFC3339)
		reviewedAt = &t
	}

	query := `
		UPDATE ideas
		SET content = ?, raw_score = ?, final_score = ?, patterns = ?,
		    recommendation = ?, analysis_details = ?, reviewed_at = ?, status = ?
		WHERE id = ?
	`

	result, err := r.db.Exec(
		query,
		idea.Content,
		idea.RawScore,
		idea.FinalScore,
		string(patternsJSON),
		idea.Recommendation,
		idea.AnalysisDetails,
		reviewedAt,
		idea.Status,
		idea.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update idea: %w", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("idea not found: %s", idea.ID)
	}

	return nil
}

// Delete deletes an idea from the database.
func (r *Repository) Delete(id string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}

	query := "DELETE FROM ideas WHERE id = ?"

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete idea: %w", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("idea not found: %s", id)
	}

	return nil
}

// List retrieves ideas based on the provided options.
func (r *Repository) List(options ListOptions) ([]*models.Idea, error) {
	query := `
		SELECT id, content, raw_score, final_score, patterns,
		       recommendation, analysis_details, created_at, reviewed_at, status
		FROM ideas
		WHERE 1=1
	`
	args := []interface{}{}

	// Add filters
	if options.Status != "" {
		query += " AND status = ?"
		args = append(args, options.Status)
	}

	if options.MinScore != nil {
		query += " AND final_score >= ?"
		args = append(args, *options.MinScore)
	}

	if options.MaxScore != nil {
		query += " AND final_score <= ?"
		args = append(args, *options.MaxScore)
	}

	// Add ordering
	if options.OrderBy != "" {
		query += " ORDER BY " + options.OrderBy
	} else {
		query += " ORDER BY created_at DESC"
	}

	// Add limit and offset
	if options.Limit != nil {
		query += " LIMIT ?"
		args = append(args, *options.Limit)
	}

	if options.Offset != nil {
		query += " OFFSET ?"
		args = append(args, *options.Offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query ideas: %w", err)
	}
	defer rows.Close()

	var ideas []*models.Idea

	for rows.Next() {
		var idea models.Idea
		var patternsJSON string
		var createdAt string
		var reviewedAt sql.NullString

		err := rows.Scan(
			&idea.ID,
			&idea.Content,
			&idea.RawScore,
			&idea.FinalScore,
			&patternsJSON,
			&idea.Recommendation,
			&idea.AnalysisDetails,
			&createdAt,
			&reviewedAt,
			&idea.Status,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Parse patterns JSON
		if patternsJSON != "" && patternsJSON != "null" {
			if err := json.Unmarshal([]byte(patternsJSON), &idea.Patterns); err != nil {
				return nil, fmt.Errorf("failed to parse patterns: %w", err)
			}
		}

		// Parse timestamps
		if parsedTime, err := time.Parse(time.RFC3339, createdAt); err == nil {
			idea.CreatedAt = parsedTime
		}

		if reviewedAt.Valid {
			if parsedTime, err := time.Parse(time.RFC3339, reviewedAt.String); err == nil {
				idea.ReviewedAt = &parsedTime
			}
		}

		ideas = append(ideas, &idea)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return ideas, nil
}

// Close closes the database connection.
func (r *Repository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}
