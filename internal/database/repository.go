// Package database provides SQLite database operations and migration management for idea storage.
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
	"github.com/rs/zerolog/log"
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

// validOrderByColumns defines the whitelist of allowed ORDER BY columns
var validOrderByColumns = map[string]bool{
	"id":               true,
	"content":          true,
	"raw_score":        true,
	"final_score":      true,
	"created_at":       true,
	"reviewed_at":      true,
	"status":           true,
	"final_score DESC": true,
	"final_score ASC":  true,
	"created_at DESC":  true,
	"created_at ASC":   true,
	"raw_score DESC":   true,
	"raw_score ASC":    true,
}

// validateOrderBy validates and sanitizes the ORDER BY clause against a whitelist
func validateOrderBy(orderBy string) (string, error) {
	if orderBy == "" {
		return "", nil
	}

	// Check if the exact string is in the whitelist
	if validOrderByColumns[orderBy] {
		return orderBy, nil
	}

	return "", fmt.Errorf("invalid ORDER BY clause: %s", orderBy)
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

	// Enable WAL mode and other optimizations via connection string
	dsn := dbPath + "?_journal_mode=WAL&_synchronous=NORMAL&_busy_timeout=5000"

	// Open database connection
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pooling
	// For single-user local usage, conservative settings are appropriate
	db.SetMaxOpenConns(5)                  // Max 5 concurrent connections
	db.SetMaxIdleConns(2)                  // Keep 2 connections ready for reuse
	db.SetConnMaxLifetime(5 * time.Minute) // Refresh connections every 5 minutes

	// Test connection
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Apply additional WAL optimizations
	// These ensure WAL mode is active even if connection string params don't work
	pragmas := []string{
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA busy_timeout = 5000",
		"PRAGMA cache_size = -64000", // 64MB cache
		"PRAGMA temp_store = MEMORY", // Keep temp tables in memory
		"PRAGMA foreign_keys = ON",   // Enable foreign keys
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("failed to execute %s: %w", pragma, err)
		}
	}

	repo := &Repository{db: db}

	// Run migrations
	if err := repo.runMigrations(); err != nil {
		_ = db.Close()
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
			// For idempotency: ignore "duplicate column" errors from ALTER TABLE ADD COLUMN
			// This allows migrations to be run multiple times safely (e.g., in tests)
			if strings.Contains(err.Error(), "duplicate column name") {
				// Migration already applied, skip silently
				continue
			}
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

	// Serialize tags to JSON
	tagsJSON, err := json.Marshal(idea.Tags)
	if err != nil {
		return fmt.Errorf("failed to serialize tags: %w", err)
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
			id, content, raw_score, final_score, patterns, tags,
			recommendation, analysis_details, created_at, reviewed_at, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.Exec(
		query,
		idea.ID,
		idea.Content,
		idea.RawScore,
		idea.FinalScore,
		string(patternsJSON),
		string(tagsJSON),
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
		SELECT id, content, raw_score, final_score, patterns, tags,
		       recommendation, analysis_details, created_at, reviewed_at, status
		FROM ideas
		WHERE id = ?
	`

	var idea models.Idea
	var patternsJSON string
	var tagsJSON string
	var createdAt string
	var reviewedAt sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&idea.ID,
		&idea.Content,
		&idea.RawScore,
		&idea.FinalScore,
		&patternsJSON,
		&tagsJSON,
		&idea.Recommendation,
		&idea.AnalysisDetails,
		&createdAt,
		&reviewedAt,
		&idea.Status,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, id)
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

	// Parse tags JSON
	if tagsJSON != "" && tagsJSON != "null" {
		if err := json.Unmarshal([]byte(tagsJSON), &idea.Tags); err != nil {
			return nil, fmt.Errorf("failed to parse tags: %w", err)
		}
	}

	// Parse timestamps
	if createdAt != "" {
		parsedTime, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, fmt.Errorf("corrupted created_at timestamp in database: %w", err)
		}
		idea.CreatedAt = parsedTime
	}

	if reviewedAt.Valid {
		parsedTime, err := time.Parse(time.RFC3339, reviewedAt.String)
		if err != nil {
			return nil, fmt.Errorf("corrupted reviewed_at timestamp in database: %w", err)
		}
		idea.ReviewedAt = &parsedTime
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

	// Serialize tags to JSON
	tagsJSON, err := json.Marshal(idea.Tags)
	if err != nil {
		return fmt.Errorf("failed to serialize tags: %w", err)
	}

	// Format timestamps
	var reviewedAt *string
	if idea.ReviewedAt != nil {
		t := idea.ReviewedAt.Format(time.RFC3339)
		reviewedAt = &t
	}

	query := `
		UPDATE ideas
		SET content = ?, raw_score = ?, final_score = ?, patterns = ?, tags = ?,
		    recommendation = ?, analysis_details = ?, reviewed_at = ?, status = ?
		WHERE id = ?
	`

	result, err := r.db.Exec(
		query,
		idea.Content,
		idea.RawScore,
		idea.FinalScore,
		string(patternsJSON),
		string(tagsJSON),
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
		return fmt.Errorf("%w: %s", ErrNotFound, idea.ID)
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
		return fmt.Errorf("%w: %s", ErrNotFound, id)
	}

	return nil
}

// scanIdeaRow scans a single database row into an Idea struct
func scanIdeaRow(rows *sql.Rows) (*models.Idea, error) {
	var idea models.Idea
	var patternsJSON string
	var tagsJSON string
	var createdAt string
	var reviewedAt sql.NullString

	err := rows.Scan(
		&idea.ID,
		&idea.Content,
		&idea.RawScore,
		&idea.FinalScore,
		&patternsJSON,
		&tagsJSON,
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

	// Parse tags JSON
	if tagsJSON != "" && tagsJSON != "null" {
		if err := json.Unmarshal([]byte(tagsJSON), &idea.Tags); err != nil {
			return nil, fmt.Errorf("failed to parse tags: %w", err)
		}
	}

	// Parse timestamps
	if createdAt != "" {
		parsedTime, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, fmt.Errorf("corrupted created_at timestamp in database: %w", err)
		}
		idea.CreatedAt = parsedTime
	}

	if reviewedAt.Valid {
		parsedTime, err := time.Parse(time.RFC3339, reviewedAt.String)
		if err != nil {
			return nil, fmt.Errorf("corrupted reviewed_at timestamp in database: %w", err)
		}
		idea.ReviewedAt = &parsedTime
	}

	return &idea, nil
}

// List retrieves ideas based on the provided options.
func (r *Repository) List(options ListOptions) ([]*models.Idea, error) {
	query := `
		SELECT id, content, raw_score, final_score, patterns, tags,
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

	// Add ordering with validation to prevent SQL injection
	if options.OrderBy != "" {
		validatedOrderBy, err := validateOrderBy(options.OrderBy)
		if err != nil {
			return nil, fmt.Errorf("invalid order by clause: %w", err)
		}
		query += " ORDER BY " + validatedOrderBy
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
	defer func() {
		if err := rows.Close(); err != nil {
			log.Warn().Err(err).Msg("failed to close rows")
		}
	}()

	var ideas []*models.Idea

	for rows.Next() {
		idea, err := scanIdeaRow(rows)
		if err != nil {
			return nil, err
		}
		ideas = append(ideas, idea)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return ideas, nil
}

// DB returns the underlying database connection for health checks and other purposes.
func (r *Repository) DB() *sql.DB {
	return r.db
}

// Ping verifies a connection to the database is still alive.
func (r *Repository) Ping() error {
	if r.db != nil {
		return r.db.Ping()
	}
	return errors.New("database connection is nil")
}

// Close closes the database connection.
func (r *Repository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

// --- Relationship Methods ---

// CreateRelationship creates a new relationship between two ideas
func (r *Repository) CreateRelationship(relationship *models.IdeaRelationship) error {
	if relationship == nil {
		return errors.New("relationship cannot be nil")
	}

	// Validate the relationship
	if err := relationship.Validate(); err != nil {
		return fmt.Errorf("invalid relationship: %w", err)
	}

	// Check that both ideas exist
	if _, err := r.GetByID(relationship.SourceIdeaID); err != nil {
		return fmt.Errorf("source idea not found: %w", err)
	}
	if _, err := r.GetByID(relationship.TargetIdeaID); err != nil {
		return fmt.Errorf("target idea not found: %w", err)
	}

	// Check for duplicate relationship
	checkQuery := `
		SELECT COUNT(*) FROM idea_relationships
		WHERE source_idea_id = ? AND target_idea_id = ? AND relationship_type = ?
	`
	var count int
	err := r.db.QueryRow(checkQuery, relationship.SourceIdeaID, relationship.TargetIdeaID, relationship.RelationshipType.String()).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for duplicate relationship: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("%w: relationship already exists between %s and %s with type %s", ErrAlreadyExists, relationship.SourceIdeaID, relationship.TargetIdeaID, relationship.RelationshipType)
	}

	// Insert the relationship
	query := `
		INSERT INTO idea_relationships (id, source_idea_id, target_idea_id, relationship_type, created_at)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err = r.db.Exec(
		query,
		relationship.ID,
		relationship.SourceIdeaID,
		relationship.TargetIdeaID,
		relationship.RelationshipType.String(),
		relationship.CreatedAt.Format(time.RFC3339),
	)

	if err != nil {
		return fmt.Errorf("failed to create relationship: %w", err)
	}

	return nil
}

// GetRelationship retrieves a relationship by its ID
func (r *Repository) GetRelationship(id string) (*models.IdeaRelationship, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	query := `
		SELECT id, source_idea_id, target_idea_id, relationship_type, created_at
		FROM idea_relationships
		WHERE id = ?
	`

	var rel models.IdeaRelationship
	var createdAt string
	var relTypeStr string

	err := r.db.QueryRow(query, id).Scan(
		&rel.ID,
		&rel.SourceIdeaID,
		&rel.TargetIdeaID,
		&relTypeStr,
		&createdAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query relationship: %w", err)
	}

	// Parse relationship type
	relType, err := models.ParseRelationshipType(relTypeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid relationship type in database: %w", err)
	}
	rel.RelationshipType = relType

	// Parse timestamp
	if createdAt != "" {
		parsedTime, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, fmt.Errorf("corrupted created_at timestamp in database: %w", err)
		}
		rel.CreatedAt = parsedTime
	}

	return &rel, nil
}

// GetRelationshipsForIdea retrieves all relationships for a given idea
func (r *Repository) GetRelationshipsForIdea(ideaID string) ([]*models.IdeaRelationship, error) {
	if ideaID == "" {
		return nil, errors.New("ideaID cannot be empty")
	}

	query := `
		SELECT id, source_idea_id, target_idea_id, relationship_type, created_at
		FROM idea_relationships
		WHERE source_idea_id = ? OR target_idea_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, ideaID, ideaID)
	if err != nil {
		return nil, fmt.Errorf("failed to query relationships: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Warn().Err(err).Msg("failed to close rows")
		}
	}()

	var relationships []*models.IdeaRelationship

	for rows.Next() {
		var rel models.IdeaRelationship
		var createdAt string
		var relTypeStr string

		err := rows.Scan(
			&rel.ID,
			&rel.SourceIdeaID,
			&rel.TargetIdeaID,
			&relTypeStr,
			&createdAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan relationship: %w", err)
		}

		// Parse relationship type
		relType, err := models.ParseRelationshipType(relTypeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid relationship type in database: %w", err)
		}
		rel.RelationshipType = relType

		// Parse timestamp
		if createdAt != "" {
			parsedTime, err := time.Parse(time.RFC3339, createdAt)
			if err != nil {
				return nil, fmt.Errorf("corrupted created_at timestamp in database: %w", err)
			}
			rel.CreatedAt = parsedTime
		}

		relationships = append(relationships, &rel)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return relationships, nil
}

// GetRelatedIdeas retrieves ideas related to a given idea, optionally filtered by relationship type
func (r *Repository) GetRelatedIdeas(ideaID string, relType *models.RelationshipType) ([]*models.Idea, error) {
	if ideaID == "" {
		return nil, errors.New("ideaID cannot be empty")
	}

	baseQuery := `
		SELECT DISTINCT i.id, i.content, i.raw_score, i.final_score, i.patterns, i.tags,
		       i.recommendation, i.analysis_details, i.created_at, i.reviewed_at, i.status
		FROM ideas i
		INNER JOIN idea_relationships r ON (i.id = r.target_idea_id OR i.id = r.source_idea_id)
		WHERE (r.source_idea_id = ? OR r.target_idea_id = ?)
		AND i.id != ?
	`

	args := []interface{}{ideaID, ideaID, ideaID}

	if relType != nil {
		baseQuery += " AND r.relationship_type = ?"
		args = append(args, relType.String())
	}

	rows, err := r.db.Query(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query related ideas: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Warn().Err(err).Msg("failed to close rows")
		}
	}()

	var ideas []*models.Idea

	for rows.Next() {
		idea, err := scanIdeaRow(rows)
		if err != nil {
			return nil, err
		}
		ideas = append(ideas, idea)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return ideas, nil
}

// DeleteRelationship deletes a relationship by its ID
func (r *Repository) DeleteRelationship(id string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}

	query := "DELETE FROM idea_relationships WHERE id = ?"

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete relationship: %w", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%w: %s", ErrNotFound, id)
	}

	return nil
}

// pathState represents a state in the BFS path finding algorithm
type pathState struct {
	currentID string
	path      []*models.IdeaRelationship
	visited   map[string]bool
}

// FindRelationshipPath finds all paths between two ideas using BFS
func (r *Repository) FindRelationshipPath(sourceID, targetID string, maxDepth int) ([][]*models.IdeaRelationship, error) {
	if sourceID == "" || targetID == "" {
		return nil, errors.New("sourceID and targetID cannot be empty")
	}

	if maxDepth <= 0 {
		maxDepth = 3 // Default max depth
	}

	// Check if both ideas exist
	if _, err := r.GetByID(sourceID); err != nil {
		return nil, fmt.Errorf("source idea not found: %w", err)
	}
	if _, err := r.GetByID(targetID); err != nil {
		return nil, fmt.Errorf("target idea not found: %w", err)
	}

	// Initialize BFS queue
	queue := []pathState{{
		currentID: sourceID,
		path:      []*models.IdeaRelationship{},
		visited:   map[string]bool{sourceID: true},
	}}

	var foundPaths [][]*models.IdeaRelationship

	for len(queue) > 0 {
		state := queue[0]
		queue = queue[1:]

		// Check if we've reached max depth
		if len(state.path) >= maxDepth {
			continue
		}

		// Get all relationships for current idea
		relationships, err := r.GetRelationshipsForIdea(state.currentID)
		if err != nil {
			return nil, fmt.Errorf("failed to get relationships: %w", err)
		}

		for _, rel := range relationships {
			// Determine next node
			var nextID string
			if rel.SourceIdeaID == state.currentID {
				nextID = rel.TargetIdeaID
			} else {
				nextID = rel.SourceIdeaID
			}

			// Skip if already visited in this path
			if state.visited[nextID] {
				continue
			}

			// Create new path
			newPath := make([]*models.IdeaRelationship, len(state.path)+1)
			copy(newPath, state.path)
			newPath[len(state.path)] = rel

			// Check if we've found the target
			if nextID == targetID {
				foundPaths = append(foundPaths, newPath)
				continue
			}

			// Add to queue for further exploration
			newVisited := make(map[string]bool)
			for k, v := range state.visited {
				newVisited[k] = v
			}
			newVisited[nextID] = true

			queue = append(queue, pathState{
				currentID: nextID,
				path:      newPath,
				visited:   newVisited,
			})
		}
	}

	return foundPaths, nil
}
