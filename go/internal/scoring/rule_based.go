package scoring

import (
	"strings"
	"unicode"
)

// RuleBasedScorer provides fast, heuristic-based scoring
type RuleBasedScorer struct {
	weights map[string]float64
}

// NewRuleBasedScorer creates a new rule-based scorer
func NewRuleBasedScorer() *RuleBasedScorer {
	return &RuleBasedScorer{
		weights: map[string]float64{
			"keyword_match":   3.0, // Max 3 points for keyword matches
			"length":          2.0, // Max 2 points for appropriate length
			"telos_alignment": 3.0, // Max 3 points for telos keywords
			"complexity":      2.0, // Max 2 points for idea complexity
		},
	}
}

// Score calculates a rule-based score (0-10)
func (s *RuleBasedScorer) Score(content, telos string) float64 {
	var totalScore float64

	// 1. Keyword matching (0-3 points)
	keywordScore := s.scoreKeywords(content)
	totalScore += keywordScore * s.weights["keyword_match"] / 3.0

	// 2. Length appropriateness (0-2 points)
	lengthScore := s.scoreLength(content)
	totalScore += lengthScore * s.weights["length"] / 2.0

	// 3. Telos alignment (0-3 points)
	if telos != "" {
		telosScore := s.scoreTelosAlignment(content, telos)
		totalScore += telosScore * s.weights["telos_alignment"] / 3.0
	} else {
		// If no telos, redistribute weight to other factors
		totalScore += 1.5 // Neutral score
	}

	// 4. Complexity/detail (0-2 points)
	complexityScore := s.scoreComplexity(content)
	totalScore += complexityScore * s.weights["complexity"] / 2.0

	// Ensure score is between 0 and 10
	if totalScore < 0 {
		totalScore = 0
	}
	if totalScore > 10 {
		totalScore = 10
	}

	return totalScore
}

// scoreKeywords scores based on positive and negative keywords
func (s *RuleBasedScorer) scoreKeywords(content string) float64 {
	contentLower := strings.ToLower(content)

	positiveKeywords := []string{
		"innovation", "improve", "solve", "build", "create",
		"impact", "sustainable", "efficient", "scale", "growth",
		"productivity", "automate", "optimize", "enhance", "transform",
	}

	negativeKeywords := []string{
		"maybe", "might", "possibly", "unclear", "vague",
		"eventually", "someday", "unsure",
	}

	score := 0.0

	// Add points for positive keywords (max 3 points)
	matches := 0
	for _, keyword := range positiveKeywords {
		if strings.Contains(contentLower, keyword) {
			matches++
			if matches >= 3 {
				break
			}
		}
	}
	score += float64(matches)

	// Subtract points for negative keywords
	for _, keyword := range negativeKeywords {
		if strings.Contains(contentLower, keyword) {
			score -= 0.5
		}
	}

	if score < 0 {
		score = 0
	}
	if score > 3 {
		score = 3
	}

	return score
}

// scoreLength scores based on content length
func (s *RuleBasedScorer) scoreLength(content string) float64 {
	length := len(strings.TrimSpace(content))

	// Too short
	if length < 10 {
		return 0.5
	}

	// Good length (20-500 characters)
	if length >= 20 && length <= 500 {
		return 2.0
	}

	// Medium length (500-1000)
	if length > 500 && length <= 1000 {
		return 1.5
	}

	// Very long (might be rambling)
	return 1.0
}

// scoreTelosAlignment scores alignment with telos keywords
func (s *RuleBasedScorer) scoreTelosAlignment(content, telos string) float64 {
	contentLower := strings.ToLower(content)
	telosLower := strings.ToLower(telos)

	// Extract important words from telos (length > 4, not common words)
	telosWords := extractImportantWords(telosLower)
	if len(telosWords) == 0 {
		return 1.5 // Neutral if no keywords
	}

	// Count matches
	matches := 0
	for _, word := range telosWords {
		if strings.Contains(contentLower, word) {
			matches++
		}
	}

	// Calculate score (0-3)
	matchRatio := float64(matches) / float64(len(telosWords))
	score := matchRatio * 3.0

	if score > 3 {
		score = 3
	}

	return score
}

// scoreComplexity scores based on sentence structure and detail
func (s *RuleBasedScorer) scoreComplexity(content string) float64 {
	// Count sentences
	sentences := strings.FieldsFunc(content, func(r rune) bool {
		return r == '.' || r == '!' || r == '?'
	})

	// Count words
	words := strings.Fields(content)

	// Single sentence or very few words
	if len(sentences) <= 1 || len(words) < 10 {
		return 0.5
	}

	// Good detail (2-5 sentences, 20-100 words)
	if len(sentences) >= 2 && len(sentences) <= 5 && len(words) >= 20 && len(words) <= 100 {
		return 2.0
	}

	// Some detail
	if len(words) >= 10 && len(words) < 200 {
		return 1.5
	}

	// Very detailed (might be too much)
	return 1.0
}

// extractImportantWords extracts significant words from text
func extractImportantWords(text string) []string {
	commonWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "from": true,
		"as": true, "is": true, "was": true, "are": true, "were": true,
		"be": true, "been": true, "being": true, "have": true, "has": true,
		"had": true, "do": true, "does": true, "did": true, "will": true,
		"would": true, "could": true, "should": true, "may": true, "might": true,
		"must": true, "can": true, "this": true, "that": true, "these": true,
		"those": true, "i": true, "you": true, "he": true, "she": true,
		"it": true, "we": true, "they": true, "my": true, "your": true,
	}

	words := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	var important []string
	for _, word := range words {
		word = strings.ToLower(strings.TrimSpace(word))
		// Include if: length > 4 and not a common word
		if len(word) > 4 && !commonWords[word] {
			important = append(important, word)
		}
	}

	return important
}
