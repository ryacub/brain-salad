package quality

import "strings"

// QualityMetrics represents the quality assessment of an LLM analysis
type QualityMetrics struct {
	Completeness float64 // 0.0-1.0: Are all scoring dimensions present?
	Consistency  float64 // 0.0-1.0: Do scores match expectations?
	Confidence   float64 // 0.0-1.0: Is LLM confident in analysis?
}

// CalculateCompleteness checks if all required fields are present
// Returns a score from 0.0 to 1.0
func CalculateCompleteness(hasScores, hasExplanations, hasFinalScore bool) float64 {
	score := 0.0
	if hasScores {
		score += 0.4
	}
	if hasExplanations {
		score += 0.3
	}
	if hasFinalScore {
		score += 0.3
	}
	return score
}

// CalculateConsistency checks if scores align with expectations
// Returns a score from 0.0 to 1.0
func CalculateConsistency(finalScore float64, sumOfComponents float64) float64 {
	// Both zero is considered consistent
	if finalScore == 0 && sumOfComponents == 0 {
		return 1.0
	}

	diff := abs(finalScore - sumOfComponents)
	tolerance := 0.5 // Allow 0.5 point difference

	if diff <= tolerance {
		return 1.0
	}

	// Linear decay beyond tolerance
	consistency := 1.0 - (diff-tolerance)/10.0
	if consistency < 0 {
		return 0.0
	}
	return consistency
}

// CalculateConfidence based on explanation length and clarity
// Returns a score from 0.0 to 1.0
func CalculateConfidence(explanationLength int, hasQualifiers bool) float64 {
	confidence := 0.5 // Base confidence

	// Longer explanations = more confidence
	if explanationLength > 100 {
		confidence += 0.3
	} else if explanationLength > 50 {
		confidence += 0.2
	}

	// Qualifiers ("maybe", "possibly") reduce confidence
	if hasQualifiers {
		confidence -= 0.2
	}

	if confidence < 0 {
		return 0.0
	}
	if confidence > 1.0 {
		return 1.0
	}
	return confidence
}

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// containsQualifiers checks if text contains uncertainty qualifiers
func containsQualifiers(text string) bool {
	qualifiers := []string{"maybe", "possibly", "perhaps", "might", "could be"}
	textLower := strings.ToLower(text)
	for _, q := range qualifiers {
		if strings.Contains(textLower, q) {
			return true
		}
	}
	return false
}
