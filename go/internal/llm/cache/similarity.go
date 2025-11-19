package cache

import (
	"regexp"
	"strings"
)

var (
	// Common English stopwords to filter out
	stopwords = map[string]bool{
		"a": true, "an": true, "and": true, "are": true, "as": true,
		"at": true, "be": true, "by": true, "for": true, "from": true,
		"has": true, "he": true, "in": true, "is": true, "it": true,
		"its": true, "of": true, "on": true, "or": true, "that": true,
		"the": true, "to": true, "was": true, "will": true, "with": true,
	}

	nonAlphanumeric = regexp.MustCompile(`[^a-z0-9\s]+`)
	multipleSpaces  = regexp.MustCompile(`\s+`)
)

// NormalizeText canonicalizes text for similarity comparison
func NormalizeText(text string) string {
	text = strings.ToLower(text)
	text = nonAlphanumeric.ReplaceAllString(text, " ")
	text = multipleSpaces.ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)
	return text
}

// Tokenize splits text into words and removes stopwords
func Tokenize(text string) []string {
	normalized := NormalizeText(text)
	words := strings.Fields(normalized)

	filtered := make([]string, 0, len(words))
	for _, word := range words {
		if !stopwords[word] && len(word) > 1 {
			filtered = append(filtered, word)
		}
	}
	return filtered
}

// JaccardSimilarity computes Jaccard similarity between two texts
func JaccardSimilarity(text1, text2 string) float64 {
	tokens1 := Tokenize(text1)
	tokens2 := Tokenize(text2)

	if len(tokens1) == 0 && len(tokens2) == 0 {
		return 1.0
	}
	if len(tokens1) == 0 || len(tokens2) == 0 {
		return 0.0
	}

	set1 := make(map[string]bool)
	for _, token := range tokens1 {
		set1[token] = true
	}

	set2 := make(map[string]bool)
	for _, token := range tokens2 {
		set2[token] = true
	}

	intersection := 0
	for token := range set1 {
		if set2[token] {
			intersection++
		}
	}

	union := len(set1) + len(set2) - intersection
	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}
