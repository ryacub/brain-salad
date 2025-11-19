package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rayyacub/telos-idea-matrix/internal/database"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/rayyacub/telos-idea-matrix/internal/patterns"
	"github.com/rayyacub/telos-idea-matrix/internal/scoring"
)

// Request/Response types

// AnalyzeRequest represents a request to analyze an idea
type AnalyzeRequest struct {
	Content string `json:"content"`
}

// AnalyzeResponse represents the response from analyzing an idea
type AnalyzeResponse struct {
	Analysis *models.Analysis `json:"analysis"`
}

// CreateIdeaRequest represents a request to create an idea
type CreateIdeaRequest struct {
	Content string `json:"content"`
}

// UpdateIdeaRequest represents a request to update an idea
type UpdateIdeaRequest struct {
	Content *string `json:"content,omitempty"`
	Status  *string `json:"status,omitempty"`
}

// IdeaResponse represents an idea in API responses
type IdeaResponse struct {
	ID             string           `json:"id"`
	Content        string           `json:"content"`
	RawScore       float64          `json:"raw_score"`
	FinalScore     float64          `json:"final_score"`
	Patterns       []string         `json:"patterns"`
	Recommendation string           `json:"recommendation"`
	Analysis       *models.Analysis `json:"analysis,omitempty"`
	CreatedAt      string           `json:"created_at"`
	ReviewedAt     *string          `json:"reviewed_at,omitempty"`
	Status         string           `json:"status"`
}

// ListIdeasResponse represents a paginated list of ideas
type ListIdeasResponse struct {
	Ideas  []IdeaResponse `json:"ideas"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// StatsResponse represents analytics statistics
type StatsResponse struct {
	TotalIdeas   int     `json:"total_ideas"`
	ActiveIdeas  int     `json:"active_ideas"`
	AverageScore float64 `json:"average_score"`
	HighScore    float64 `json:"high_score"`
	LowScore     float64 `json:"low_score"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// Helper functions

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, ErrorResponse{Error: message})
}

func ideaToResponse(idea *models.Idea) IdeaResponse {
	resp := IdeaResponse{
		ID:             idea.ID,
		Content:        idea.Content,
		RawScore:       idea.RawScore,
		FinalScore:     idea.FinalScore,
		Patterns:       idea.Patterns,
		Recommendation: idea.Recommendation,
		Analysis:       idea.Analysis,
		CreatedAt:      idea.CreatedAt.Format(time.RFC3339),
		Status:         idea.Status,
	}
	if idea.ReviewedAt != nil {
		reviewedAt := idea.ReviewedAt.Format(time.RFC3339)
		resp.ReviewedAt = &reviewedAt
	}
	return resp
}

// Handlers

// HealthHandler handles health check requests
func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// AnalyzeHandler handles idea analysis requests
func (s *Server) AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Content == "" {
		respondError(w, http.StatusBadRequest, "content is required")
		return
	}

	// Analyze the idea using scoring engine and pattern detector
	scoringEngine := scoring.NewEngine(s.telos)
	analysis, err := scoringEngine.CalculateScore(req.Content)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to analyze idea: %v", err))
		return
	}

	detector := patterns.NewDetector(s.telos)
	detectedPatterns := detector.DetectPatterns(req.Content)

	// Update analysis with detected patterns
	analysis.DetectedPatterns = detectedPatterns

	respondJSON(w, http.StatusOK, AnalyzeResponse{Analysis: analysis})
}

// CreateIdeaHandler handles idea creation requests
func (s *Server) CreateIdeaHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateIdeaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Content == "" {
		respondError(w, http.StatusBadRequest, "content is required")
		return
	}

	// Analyze the idea
	scoringEngine := scoring.NewEngine(s.telos)
	analysis, err := scoringEngine.CalculateScore(req.Content)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to analyze idea: %v", err))
		return
	}

	detector := patterns.NewDetector(s.telos)
	detectedPatterns := detector.DetectPatterns(req.Content)
	analysis.DetectedPatterns = detectedPatterns

	// Extract pattern names for storage
	patternNames := make([]string, len(detectedPatterns))
	for i, p := range detectedPatterns {
		patternNames[i] = p.Name
	}

	// Create idea
	idea := &models.Idea{
		ID:             uuid.New().String(),
		Content:        req.Content,
		RawScore:       analysis.RawScore,
		FinalScore:     analysis.FinalScore,
		Patterns:       patternNames,
		Recommendation: analysis.GetRecommendation(),
		Analysis:       analysis,
		Status:         "active",
		CreatedAt:      time.Now().UTC(),
	}

	if err := s.repo.Create(idea); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create idea: %v", err))
		return
	}

	respondJSON(w, http.StatusCreated, ideaToResponse(idea))
}

// GetIdeaHandler handles requests to get a single idea
func (s *Server) GetIdeaHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	// Validate UUID
	if _, err := uuid.Parse(idStr); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid idea ID format")
		return
	}

	idea, err := s.repo.GetByID(idStr)
	if err != nil {
		respondError(w, http.StatusNotFound, "Idea not found")
		return
	}

	respondJSON(w, http.StatusOK, ideaToResponse(idea))
}

// ListIdeasHandler handles requests to list ideas
func (s *Server) ListIdeasHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()

	status := query.Get("status")
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

	var limit, offset *int
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err == nil {
			limit = &l
		}
	}
	if offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err == nil {
			offset = &o
		}
	}

	// Build list options
	options := database.ListOptions{
		Status: status,
		Limit:  limit,
		Offset: offset,
	}

	ideas, err := s.repo.List(options)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list ideas: %v", err))
		return
	}

	// Convert to response format
	ideaResponses := make([]IdeaResponse, len(ideas))
	for i := range ideas {
		ideaResponses[i] = ideaToResponse(ideas[i])
	}

	// Calculate totals
	total := len(ideas)
	responseLimit := 100
	responseOffset := 0
	if limit != nil {
		responseLimit = *limit
	}
	if offset != nil {
		responseOffset = *offset
	}

	respondJSON(w, http.StatusOK, ListIdeasResponse{
		Ideas:  ideaResponses,
		Total:  total,
		Limit:  responseLimit,
		Offset: responseOffset,
	})
}

// UpdateIdeaHandler handles requests to update an idea
func (s *Server) UpdateIdeaHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	// Validate UUID
	if _, err := uuid.Parse(idStr); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid idea ID format")
		return
	}

	var req UpdateIdeaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get existing idea
	idea, err := s.repo.GetByID(idStr)
	if err != nil {
		respondError(w, http.StatusNotFound, "Idea not found")
		return
	}

	// Update fields
	if req.Content != nil {
		idea.Content = *req.Content

		// Re-analyze if content changed
		scoringEngine := scoring.NewEngine(s.telos)
		analysis, err := scoringEngine.CalculateScore(idea.Content)
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to analyze idea: %v", err))
			return
		}

		detector := patterns.NewDetector(s.telos)
		detectedPatterns := detector.DetectPatterns(idea.Content)
		analysis.DetectedPatterns = detectedPatterns

		patternNames := make([]string, len(detectedPatterns))
		for i, p := range detectedPatterns {
			patternNames[i] = p.Name
		}

		idea.RawScore = analysis.RawScore
		idea.FinalScore = analysis.FinalScore
		idea.Patterns = patternNames
		idea.Recommendation = analysis.GetRecommendation()
		idea.Analysis = analysis
	}

	if req.Status != nil {
		idea.Status = *req.Status
	}

	if err := s.repo.Update(idea); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update idea: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, ideaToResponse(idea))
}

// DeleteIdeaHandler handles requests to delete an idea
func (s *Server) DeleteIdeaHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	// Validate UUID
	if _, err := uuid.Parse(idStr); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid idea ID format")
		return
	}

	// Check if idea exists
	_, err := s.repo.GetByID(idStr)
	if err != nil {
		respondError(w, http.StatusNotFound, "Idea not found")
		return
	}

	if err := s.repo.Delete(idStr); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete idea: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AnalyticsStatsHandler handles requests for analytics statistics
func (s *Server) AnalyticsStatsHandler(w http.ResponseWriter, r *http.Request) {
	// Get all ideas
	allIdeas, err := s.repo.List(database.ListOptions{})
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get statistics: %v", err))
		return
	}

	// Get active ideas
	activeIdeas, err := s.repo.List(database.ListOptions{Status: "active"})
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get statistics: %v", err))
		return
	}

	// Calculate statistics
	stats := StatsResponse{
		TotalIdeas:  len(allIdeas),
		ActiveIdeas: len(activeIdeas),
	}

	if len(allIdeas) > 0 {
		var sum, high, low float64
		high = allIdeas[0].FinalScore
		low = allIdeas[0].FinalScore

		for _, idea := range allIdeas {
			sum += idea.FinalScore
			if idea.FinalScore > high {
				high = idea.FinalScore
			}
			if idea.FinalScore < low {
				low = idea.FinalScore
			}
		}

		stats.AverageScore = sum / float64(len(allIdeas))
		stats.HighScore = high
		stats.LowScore = low
	}

	respondJSON(w, http.StatusOK, stats)
}
