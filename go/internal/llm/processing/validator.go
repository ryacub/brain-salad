package processing

import (
	"fmt"
)

// Validator validates processed results
type Validator struct {
	maxMissionScore      float64
	maxAntiChallengeScore float64
	maxStrategicScore    float64
	maxFinalScore        float64
}

// NewValidator creates a new validator with default thresholds
func NewValidator() *Validator {
	return &Validator{
		maxMissionScore:      4.0,
		maxAntiChallengeScore: 3.5,
		maxStrategicScore:    2.5,
		maxFinalScore:        10.0,
	}
}

// Validate performs comprehensive validation on a processed result
func (v *Validator) Validate(result *ProcessedResult) error {
	// Validate mission alignment score
	if err := v.validateScore(result.Scores.MissionAlignment, 0, v.maxMissionScore, "Mission Alignment"); err != nil {
		return err
	}

	// Validate anti-challenge score
	if err := v.validateScore(result.Scores.AntiChallenge, 0, v.maxAntiChallengeScore, "Anti-Challenge"); err != nil {
		return err
	}

	// Validate strategic fit score
	if err := v.validateScore(result.Scores.StrategicFit, 0, v.maxStrategicScore, "Strategic Fit"); err != nil {
		return err
	}

	// Validate final score
	if err := v.validateScore(result.FinalScore, 0, v.maxFinalScore, "Final Score"); err != nil {
		return err
	}

	// Validate recommendation
	if err := v.validateRecommendation(result.Recommendation); err != nil {
		return err
	}

	return nil
}

// validateScore checks if a score is within valid range
func (v *Validator) validateScore(score, min, max float64, name string) error {
	if score < min || score > max {
		return fmt.Errorf("%s score %v is out of valid range [%v, %v]", name, score, min, max)
	}
	return nil
}

// validateRecommendation checks if recommendation is valid
func (v *Validator) validateRecommendation(recommendation string) error {
	validRecommendations := map[string]bool{
		"PRIORITIZE NOW":  true,
		"GOOD ALIGNMENT":  true,
		"CONSIDER LATER":  true,
		"AVOID FOR NOW":   true,
	}

	if !validRecommendations[recommendation] {
		return fmt.Errorf("invalid recommendation: %s", recommendation)
	}

	return nil
}
