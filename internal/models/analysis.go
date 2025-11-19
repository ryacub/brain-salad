package models

import (
	"errors"
	"time"
)

// Analysis represents the complete scoring breakdown and analysis of an idea.
// This matches the Score struct in the Rust implementation.
type Analysis struct {
	RawScore         float64               `json:"raw_score"`
	FinalScore       float64               `json:"final_score"`
	Mission          MissionScores         `json:"mission"`
	AntiChallenge    AntiChallengeScores   `json:"anti_challenge"`
	Strategic        StrategicScores       `json:"strategic"`
	DetectedPatterns []DetectedPattern     `json:"detected_patterns"`
	Recommendations  []string              `json:"recommendations"`
	ScoringDetails   []string              `json:"scoring_details,omitempty"`
	Explanations     map[string]string     `json:"explanations,omitempty"`
	AnalyzedAt       time.Time             `json:"analyzed_at"`
}

// GetRecommendation returns the recommendation based on the final score.
// Matches Rust Recommendation enum and thresholds.
func (a *Analysis) GetRecommendation() string {
	switch {
	case a.FinalScore >= 8.5:
		return "\U0001F525 PRIORITIZE NOW" // 🔥
	case a.FinalScore >= 7.0:
		return "\u2705 GOOD ALIGNMENT" // ✅
	case a.FinalScore >= 5.0:
		return "\u26A0\uFE0F CONSIDER LATER" // ⚠️
	default:
		return "\U0001F6AB AVOID FOR NOW" // 🚫
	}
}

// MissionScores represents the mission alignment scoring breakdown.
// Max total: 4.0 points (40% of total score)
type MissionScores struct {
	DomainExpertise   float64 `json:"domain_expertise"`   // 0-1.2 points max
	AIAlignment       float64 `json:"ai_alignment"`       // 0-1.5 points max
	ExecutionSupport  float64 `json:"execution_support"`  // 0-0.8 points max
	RevenuePotential  float64 `json:"revenue_potential"`  // 0-0.5 points max
	Total             float64 `json:"total"`              // max 4.0 points
}

// AntiChallengeScores represents the anti-challenge scoring breakdown.
// Max total: 3.5 points (35% of total score)
type AntiChallengeScores struct {
	ContextSwitching  float64 `json:"context_switching"`  // 0-1.2 points max
	RapidPrototyping  float64 `json:"rapid_prototyping"`  // 0-1.0 points max
	Accountability    float64 `json:"accountability"`     // 0-0.8 points max
	IncomeAnxiety     float64 `json:"income_anxiety"`     // 0-0.5 points max
	Total             float64 `json:"total"`              // max 3.5 points
}

// StrategicScores represents the strategic fit scoring breakdown.
// Max total: 2.5 points (25% of total score)
type StrategicScores struct {
	StackCompatibility   float64 `json:"stack_compatibility"`    // 0-1.0 points max
	ShippingHabit        float64 `json:"shipping_habit"`         // 0-0.8 points max
	PublicAccountability float64 `json:"public_accountability"`  // 0-0.4 points max
	RevenueTesting       float64 `json:"revenue_testing"`        // 0-0.3 points max
	Total                float64 `json:"total"`                  // max 2.5 points
}

// DetectedPattern represents a detected anti-pattern or positive pattern.
type DetectedPattern struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"` // 0.0 to 1.0
	Severity    string  `json:"severity"`   // "low", "medium", "high", "critical"
}

// Validate validates the detected pattern.
func (p *DetectedPattern) Validate() error {
	if p.Name == "" {
		return errors.New("pattern name is required")
	}
	if p.Description == "" {
		return errors.New("pattern description is required")
	}
	if p.Confidence < 0.0 || p.Confidence > 1.0 {
		return errors.New("confidence must be between 0.0 and 1.0")
	}

	validSeverities := map[string]bool{
		"low":      true,
		"medium":   true,
		"high":     true,
		"critical": true,
	}

	if p.Severity != "" && !validSeverities[p.Severity] {
		return errors.New("invalid severity: must be one of 'low', 'medium', 'high', 'critical'")
	}

	return nil
}

// Recommendation represents the recommendation level based on score.
type Recommendation string

const (
	// RecommendationPriority indicates a high-priority idea (>= 8.5).
	RecommendationPriority Recommendation = "\U0001F525 PRIORITIZE NOW" // 🔥
	// RecommendationGood indicates good alignment (>= 7.0).
	RecommendationGood Recommendation = "\u2705 GOOD ALIGNMENT" // ✅
	// RecommendationConsider indicates worth considering later (>= 5.0).
	RecommendationConsider Recommendation = "\u26A0\uFE0F CONSIDER LATER" // ⚠️
	// RecommendationAvoid indicates should avoid for now (< 5.0).
	RecommendationAvoid Recommendation = "\U0001F6AB AVOID FOR NOW" // 🚫
)

// String returns the string representation of the recommendation.
func (r Recommendation) String() string {
	return string(r)
}
