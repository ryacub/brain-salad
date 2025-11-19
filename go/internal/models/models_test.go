package models_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rayyacub/telos-idea-matrix/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// IDEA TESTS
// ============================================================================

func TestIdea_Validate_ValidIdea_ReturnsNoError(t *testing.T) {
	idea := &models.Idea{
		Title:  "Build a SaaS product",
		Status: "active",
	}

	err := idea.Validate()
	assert.NoError(t, err)
}

func TestIdea_Validate_EmptyTitle_ReturnsError(t *testing.T) {
	idea := &models.Idea{
		Title:  "",
		Status: "active",
	}

	err := idea.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "title")
}

func TestIdea_Validate_TitleTooShort_ReturnsError(t *testing.T) {
	idea := &models.Idea{
		Title:  "AB", // Less than minimum (3 chars)
		Status: "active",
	}

	err := idea.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "title")
}

func TestIdea_Validate_TitleTooLong_ReturnsError(t *testing.T) {
	longTitle := strings.Repeat("a", 201) // 201 characters

	idea := &models.Idea{
		Title:  longTitle,
		Status: "active",
	}

	err := idea.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "title")
}

func TestIdea_Validate_InvalidStatus_ReturnsError(t *testing.T) {
	idea := &models.Idea{
		Title:  "Valid title",
		Status: "invalid_status",
	}

	err := idea.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status")
}

func TestIdea_Validate_ValidStatuses(t *testing.T) {
	validStatuses := []string{"active", "archived", "deleted"}

	for _, status := range validStatuses {
		idea := &models.Idea{
			Title:  "Valid title",
			Status: status,
		}

		err := idea.Validate()
		assert.NoError(t, err, "Status %s should be valid", status)
	}
}

func TestIdea_JSONSerialization_RoundTrip(t *testing.T) {
	now := time.Now().UTC()
	original := &models.Idea{
		ID:          uuid.New().String(),
		Content:     "Build an AI automation tool",
		RawScore:    8.5,
		FinalScore:  8.2,
		CreatedAt:   now,
		ReviewedAt:  &now,
		Status:      "active",
		Patterns:    []string{"context-switching", "perfectionism"},
		Recommendation: "=% PRIORITIZE NOW",
	}

	// Serialize
	jsonBytes, err := json.Marshal(original)
	require.NoError(t, err)

	// Deserialize
	var decoded models.Idea
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err)

	// Compare
	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.Content, decoded.Content)
	assert.Equal(t, original.RawScore, decoded.RawScore)
	assert.Equal(t, original.FinalScore, decoded.FinalScore)
	assert.Equal(t, original.Status, decoded.Status)
	assert.Equal(t, original.Patterns, decoded.Patterns)
}

func TestIdea_NewIdea_GeneratesValidID(t *testing.T) {
	idea := models.NewIdea("Build a SaaS product")

	assert.NotEmpty(t, idea.ID)
	assert.Equal(t, "Build a SaaS product", idea.Content)
	assert.Equal(t, "active", idea.Status)
	assert.NotZero(t, idea.CreatedAt)
}

func TestIdeaStatus_String_ReturnsCorrectValue(t *testing.T) {
	testCases := []struct {
		status   models.IdeaStatus
		expected string
	}{
		{models.StatusActive, "active"},
		{models.StatusArchived, "archived"},
		{models.StatusDeleted, "deleted"},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, tc.status.String())
	}
}

// ============================================================================
// TELOS TESTS
// ============================================================================

func TestTelos_Validate_ValidTelos_ReturnsNoError(t *testing.T) {
	deadline := time.Now().Add(30 * 24 * time.Hour)
	telos := &models.Telos{
		Goals: []models.Goal{
			{ID: "G1", Description: "Build products", Deadline: &deadline, Priority: 1},
		},
		Strategies: []models.Strategy{
			{ID: "S1", Description: "Ship early"},
		},
		Stack: models.Stack{
			Primary:   []string{"Go", "TypeScript"},
			Secondary: []string{"Docker"},
		},
	}

	err := telos.Validate()
	assert.NoError(t, err)
}

func TestTelos_Validate_NoGoals_ReturnsError(t *testing.T) {
	telos := &models.Telos{
		Goals: []models.Goal{},
	}

	err := telos.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "goal")
}

func TestTelos_Validate_EmptyGoalDescription_ReturnsError(t *testing.T) {
	telos := &models.Telos{
		Goals: []models.Goal{
			{ID: "G1", Description: "", Priority: 1},
		},
	}

	err := telos.Validate()
	assert.Error(t, err)
}

func TestTelos_Validate_InvalidStrategy_ReturnsError(t *testing.T) {
	telos := &models.Telos{
		Goals: []models.Goal{
			{ID: "G1", Description: "Build products", Priority: 1},
		},
		Strategies: []models.Strategy{
			{ID: "", Description: "Ship early"}, // Invalid: empty ID
		},
	}

	err := telos.Validate()
	assert.Error(t, err)
}

func TestTelos_Validate_InvalidPattern_ReturnsError(t *testing.T) {
	telos := &models.Telos{
		Goals: []models.Goal{
			{ID: "G1", Description: "Build products", Priority: 1},
		},
		FailurePatterns: []models.Pattern{
			{Name: "", Description: "Something"}, // Invalid: empty name
		},
	}

	err := telos.Validate()
	assert.Error(t, err)
}

func TestTelos_JSONSerialization_RoundTrip(t *testing.T) {
	deadline := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	original := &models.Telos{
		Goals: []models.Goal{
			{
				ID:          "G1",
				Description: "Build profitable SaaS",
				Deadline:    &deadline,
				Priority:    1,
			},
		},
		Strategies: []models.Strategy{
			{ID: "S1", Description: "Ship early and often"},
		},
		Stack: models.Stack{
			Primary:   []string{"Go", "TypeScript"},
			Secondary: []string{"Docker"},
		},
		FailurePatterns: []models.Pattern{
			{
				Name:        "Context switching",
				Description: "Starting new projects before finishing current ones",
				Keywords:    []string{"new", "projects", "starting"},
			},
		},
	}

	// Serialize
	jsonBytes, err := json.Marshal(original)
	require.NoError(t, err)

	// Deserialize
	var decoded models.Telos
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err)

	// Compare
	assert.Equal(t, len(original.Goals), len(decoded.Goals))
	assert.Equal(t, original.Goals[0].ID, decoded.Goals[0].ID)
	assert.Equal(t, original.Goals[0].Description, decoded.Goals[0].Description)
	assert.Equal(t, len(original.Strategies), len(decoded.Strategies))
	assert.Equal(t, len(original.Stack.Primary), len(decoded.Stack.Primary))
}

// ============================================================================
// GOAL TESTS
// ============================================================================

func TestGoal_Validate_ValidGoal_ReturnsNoError(t *testing.T) {
	deadline := time.Now().Add(30 * 24 * time.Hour)
	goal := &models.Goal{
		ID:          "G1",
		Description: "Build a SaaS product",
		Deadline:    &deadline,
		Priority:    1,
	}

	err := goal.Validate()
	assert.NoError(t, err)
}

func TestGoal_Validate_EmptyID_ReturnsError(t *testing.T) {
	goal := &models.Goal{
		ID:          "",
		Description: "Build a SaaS product",
		Priority:    1,
	}

	err := goal.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ID")
}

func TestGoal_Validate_EmptyDescription_ReturnsError(t *testing.T) {
	goal := &models.Goal{
		ID:          "G1",
		Description: "",
		Priority:    1,
	}

	err := goal.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "description")
}

// ============================================================================
// STRATEGY TESTS
// ============================================================================

func TestStrategy_Validate_ValidStrategy_ReturnsNoError(t *testing.T) {
	strategy := &models.Strategy{
		ID:          "S1",
		Description: "Ship early and often",
	}

	err := strategy.Validate()
	assert.NoError(t, err)
}

func TestStrategy_Validate_EmptyID_ReturnsError(t *testing.T) {
	strategy := &models.Strategy{
		ID:          "",
		Description: "Ship early",
	}

	err := strategy.Validate()
	assert.Error(t, err)
}

func TestStrategy_Validate_EmptyDescription_ReturnsError(t *testing.T) {
	strategy := &models.Strategy{
		ID:          "S1",
		Description: "",
	}

	err := strategy.Validate()
	assert.Error(t, err)
}

// ============================================================================
// PATTERN TESTS
// ============================================================================

func TestPattern_Validate_ValidPattern_ReturnsNoError(t *testing.T) {
	pattern := &models.Pattern{
		Name:        "Context switching",
		Description: "Starting new projects before finishing current ones",
		Keywords:    []string{"new", "projects"},
	}

	err := pattern.Validate()
	assert.NoError(t, err)
}

func TestPattern_Validate_EmptyName_ReturnsError(t *testing.T) {
	pattern := &models.Pattern{
		Name:        "",
		Description: "Some description",
		Keywords:    []string{"keyword"},
	}

	err := pattern.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestPattern_Validate_EmptyDescription_ReturnsError(t *testing.T) {
	pattern := &models.Pattern{
		Name:        "Context switching",
		Description: "",
		Keywords:    []string{"keyword"},
	}

	err := pattern.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "description")
}

// ============================================================================
// ANALYSIS TESTS
// ============================================================================

func TestAnalysis_CalculateRecommendation_Priority(t *testing.T) {
	analysis := &models.Analysis{
		FinalScore: 8.5,
	}

	rec := analysis.GetRecommendation()
	assert.Equal(t, "\U0001F525 PRIORITIZE NOW", rec)
}

func TestAnalysis_CalculateRecommendation_Good(t *testing.T) {
	analysis := &models.Analysis{
		FinalScore: 7.5,
	}

	rec := analysis.GetRecommendation()
	assert.Equal(t, "\u2705 GOOD ALIGNMENT", rec)
}

func TestAnalysis_CalculateRecommendation_Consider(t *testing.T) {
	analysis := &models.Analysis{
		FinalScore: 6.0,
	}

	rec := analysis.GetRecommendation()
	assert.Equal(t, "\u26A0\uFE0F CONSIDER LATER", rec)
}

func TestAnalysis_CalculateRecommendation_Avoid(t *testing.T) {
	analysis := &models.Analysis{
		FinalScore: 4.0,
	}

	rec := analysis.GetRecommendation()
	assert.Equal(t, "\U0001F6AB AVOID FOR NOW", rec)
}

func TestAnalysis_JSONSerialization_RoundTrip(t *testing.T) {
	now := time.Now().UTC()
	original := &models.Analysis{
		RawScore:   9.2,
		FinalScore: 9.2,
		Mission: models.MissionScores{
			DomainExpertise:   1.1,
			AIAlignment:       1.4,
			ExecutionSupport:  0.75,
			RevenueoPotential: 0.45,
			Total:             3.7,
		},
		AntiChallenge: models.AntiChallengeScores{
			ContextSwitching:  1.15,
			RapidPrototyping:  0.95,
			Accountability:    0.7,
			IncomeAnxiety:     0.45,
			Total:             3.25,
		},
		Strategic: models.StrategicScores{
			StackCompatibility:   0.9,
			ShippingHabit:        0.7,
			PublicAccountability: 0.35,
			RevenueTesting:       0.28,
			Total:                2.23,
		},
		DetectedPatterns: []models.DetectedPattern{
			{
				Name:        "Context switching",
				Description: "Staying focused on current stack",
				Confidence:  0.9,
				Severity:    "low",
			},
		},
		Recommendations: []string{
			"Great alignment with current goals",
			"Ship MVP within 30 days",
		},
		AnalyzedAt: now,
	}

	// Serialize
	jsonBytes, err := json.Marshal(original)
	require.NoError(t, err)

	// Deserialize
	var decoded models.Analysis
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err)

	// Compare
	assert.Equal(t, original.RawScore, decoded.RawScore)
	assert.Equal(t, original.FinalScore, decoded.FinalScore)
	assert.Equal(t, original.Mission.Total, decoded.Mission.Total)
	assert.Equal(t, original.AntiChallenge.Total, decoded.AntiChallenge.Total)
	assert.Equal(t, original.Strategic.Total, decoded.Strategic.Total)
	assert.Equal(t, len(original.DetectedPatterns), len(decoded.DetectedPatterns))
}

// ============================================================================
// DETECTED PATTERN TESTS
// ============================================================================

func TestDetectedPattern_Validate_ValidPattern_ReturnsNoError(t *testing.T) {
	pattern := &models.DetectedPattern{
		Name:        "Context switching",
		Description: "Using current stack",
		Confidence:  0.9,
		Severity:    "low",
	}

	err := pattern.Validate()
	assert.NoError(t, err)
}

func TestDetectedPattern_Validate_InvalidSeverity_ReturnsError(t *testing.T) {
	pattern := &models.DetectedPattern{
		Name:        "Context switching",
		Description: "Using current stack",
		Confidence:  0.9,
		Severity:    "invalid",
	}

	err := pattern.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "severity")
}

func TestDetectedPattern_Validate_EmptyName_ReturnsError(t *testing.T) {
	pattern := &models.DetectedPattern{
		Name:        "",
		Description: "Using current stack",
		Confidence:  0.9,
		Severity:    "low",
	}

	err := pattern.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestDetectedPattern_Validate_EmptyDescription_ReturnsError(t *testing.T) {
	pattern := &models.DetectedPattern{
		Name:        "Context switching",
		Description: "",
		Confidence:  0.9,
		Severity:    "low",
	}

	err := pattern.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "description")
}

func TestDetectedPattern_Validate_ValidSeverities(t *testing.T) {
	validSeverities := []string{"low", "medium", "high", "critical"}

	for _, severity := range validSeverities {
		pattern := &models.DetectedPattern{
			Name:        "Test pattern",
			Description: "Test description",
			Confidence:  0.8,
			Severity:    severity,
		}

		err := pattern.Validate()
		assert.NoError(t, err, "Severity %s should be valid", severity)
	}
}

func TestDetectedPattern_Validate_ConfidenceOutOfRange_ReturnsError(t *testing.T) {
	testCases := []struct {
		name       string
		confidence float64
	}{
		{"negative confidence", -0.1},
		{"confidence too high", 1.1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pattern := &models.DetectedPattern{
				Name:        "Test pattern",
				Description: "Test description",
				Confidence:  tc.confidence,
				Severity:    "low",
			}

			err := pattern.Validate()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "confidence")
		})
	}
}

func TestRecommendation_String_ReturnsCorrectValue(t *testing.T) {
	testCases := []struct {
		rec      models.Recommendation
		expected string
	}{
		{models.RecommendationPriority, "\U0001F525 PRIORITIZE NOW"},
		{models.RecommendationGood, "\u2705 GOOD ALIGNMENT"},
		{models.RecommendationConsider, "\u26A0\uFE0F CONSIDER LATER"},
		{models.RecommendationAvoid, "\U0001F6AB AVOID FOR NOW"},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, tc.rec.String())
	}
}
