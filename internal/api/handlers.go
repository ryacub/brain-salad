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
	"github.com/rayyacub/telos-idea-matrix/internal/metrics"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/rayyacub/telos-idea-matrix/internal/patterns"
	"github.com/rayyacub/telos-idea-matrix/internal/scoring"
	"github.com/rs/zerolog/log"
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
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Warn().Err(err).Msg("failed to encode JSON response")
		}
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
	ctx := r.Context()
	status := s.healthMonitor.RunAllChecks(ctx)

	// Return appropriate HTTP status code based on health status
	httpStatus := http.StatusOK
	switch status.Status {
	case "unhealthy":
		httpStatus = http.StatusServiceUnavailable
	case "degraded":
		httpStatus = http.StatusOK // Still return 200 for degraded
	}

	respondJSON(w, httpStatus, status)
}

// AnalyzeHandler handles idea analysis requests
func (s *Server) AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

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
		// Log internal error details but don't expose to client
		log.Error().Err(err).Msg("Failed to analyze idea")
		respondError(w, http.StatusInternalServerError, "Failed to analyze idea")
		return
	}

	detector := patterns.NewDetector(s.telos)
	detectedPatterns := detector.DetectPatterns(req.Content)

	// Update analysis with detected patterns
	analysis.DetectedPatterns = detectedPatterns

	// Record metrics
	metrics.RecordScoringDuration(time.Since(start))

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
		// Log internal error details but don't expose to client
		log.Error().Err(err).Msg("Failed to analyze idea")
		respondError(w, http.StatusInternalServerError, "Failed to analyze idea")
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
		// Log internal error details but don't expose to client
		log.Error().Err(err).Str("idea_id", idea.ID).Msg("Failed to create idea")
		respondError(w, http.StatusInternalServerError, "Failed to create idea")
		return
	}

	// Record metrics
	metrics.RecordIdeaCreated()

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
		if database.IsNotFound(err) {
			respondError(w, http.StatusNotFound, "Idea not found")
			return
		}
		log.Error().Err(err).Str("idea_id", idStr).Msg("Failed to get idea")
		respondError(w, http.StatusInternalServerError, "Failed to get idea")
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
		// Log internal error details but don't expose to client
		log.Error().Err(err).Msg("Failed to list ideas")
		respondError(w, http.StatusInternalServerError, "Failed to list ideas")
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
		if database.IsNotFound(err) {
			respondError(w, http.StatusNotFound, "Idea not found")
			return
		}
		log.Error().Err(err).Str("idea_id", idStr).Msg("Failed to get idea")
		respondError(w, http.StatusInternalServerError, "Failed to get idea")
		return
	}

	// Update fields
	if req.Content != nil {
		idea.Content = *req.Content

		// Re-analyze if content changed
		scoringEngine := scoring.NewEngine(s.telos)
		analysis, err := scoringEngine.CalculateScore(idea.Content)
		if err != nil {
			// Log internal error details but don't expose to client
			log.Error().Err(err).Str("idea_id", idea.ID).Msg("Failed to analyze idea")
			respondError(w, http.StatusInternalServerError, "Failed to analyze idea")
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
		// Log internal error details but don't expose to client
		log.Error().Err(err).Str("idea_id", idea.ID).Msg("Failed to update idea")
		respondError(w, http.StatusInternalServerError, "Failed to update idea")
		return
	}

	// Record metrics
	metrics.RecordIdeaUpdated()

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
		if database.IsNotFound(err) {
			respondError(w, http.StatusNotFound, "Idea not found")
			return
		}
		log.Error().Err(err).Str("idea_id", idStr).Msg("Failed to get idea")
		respondError(w, http.StatusInternalServerError, "Failed to get idea")
		return
	}

	if err := s.repo.Delete(idStr); err != nil {
		// Log internal error details but don't expose to client
		log.Error().Err(err).Str("idea_id", idStr).Msg("Failed to delete idea")
		respondError(w, http.StatusInternalServerError, "Failed to delete idea")
		return
	}

	// Record metrics
	metrics.RecordIdeaDeleted()

	w.WriteHeader(http.StatusNoContent)
}

// AnalyticsStatsHandler handles requests for analytics statistics
func (s *Server) AnalyticsStatsHandler(w http.ResponseWriter, _ *http.Request) {
	// Note: For personal use with <10K ideas, loading all ideas is acceptable.
	// For larger datasets, this should use SQL aggregation (COUNT, AVG, MIN, MAX).
	allIdeas, err := s.repo.List(database.ListOptions{})
	if err != nil {
		// Log internal error details but don't expose to client
		log.Error().Err(err).Msg("Failed to get statistics (all ideas)")
		respondError(w, http.StatusInternalServerError, "Failed to get statistics")
		return
	}

	// Get active ideas
	activeIdeas, err := s.repo.List(database.ListOptions{Status: "active"})
	if err != nil {
		// Log internal error details but don't expose to client
		log.Error().Err(err).Msg("Failed to get statistics (active ideas)")
		respondError(w, http.StatusInternalServerError, "Failed to get statistics")
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

// MetricsHandler handles requests for application metrics
func (s *Server) MetricsHandler(w http.ResponseWriter, _ *http.Request) {
	snapshot := metrics.GetMetrics()

	// Convert to a more friendly format for API responses
	response := make(map[string]interface{})

	for name, metric := range snapshot {
		metricData := map[string]interface{}{
			"type":      string(metric.Type),
			"timestamp": metric.Timestamp.Format(time.RFC3339),
		}

		switch metric.Type {
		case metrics.Counter:
			metricData["value"] = metric.Value
			metricData["count"] = metric.Count
		case metrics.Gauge:
			metricData["value"] = metric.Value
		case metrics.Histogram:
			metricData["count"] = metric.Count
			if len(metric.Values) > 0 {
				// Calculate basic stats
				sum := 0.0
				minVal := metric.Values[0]
				maxVal := metric.Values[0]
				for _, v := range metric.Values {
					sum += v
					if v < minVal {
						minVal = v
					}
					if v > maxVal {
						maxVal = v
					}
				}
				metricData["stats"] = map[string]interface{}{
					"count": len(metric.Values),
					"min":   minVal,
					"max":   maxVal,
					"avg":   sum / float64(len(metric.Values)),
				}
			}
		}

		response[name] = metricData
	}

	respondJSON(w, http.StatusOK, response)
}

// OpenAPIHandler serves the OpenAPI specification
func (s *Server) OpenAPIHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "api/openapi.yaml")
}

// APIDocsHandler redirects to Swagger UI with the OpenAPI spec
func (s *Server) APIDocsHandler(w http.ResponseWriter, r *http.Request) {
	// Redirect to Swagger UI hosted online with our spec
	specURL := "http://localhost:8080/api/openapi.yaml"
	swaggerURL := fmt.Sprintf("https://petstore.swagger.io/?url=%s", specURL)
	http.Redirect(w, r, swaggerURL, http.StatusFound)
}
