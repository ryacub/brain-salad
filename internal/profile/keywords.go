package profile

import (
	"regexp"
	"strings"
	"unicode"
)

// Common stopwords to filter out when extracting keywords
var stopwords = map[string]bool{
	"a": true, "an": true, "the": true, "and": true, "or": true, "but": true,
	"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
	"with": true, "by": true, "from": true, "as": true, "is": true, "was": true,
	"are": true, "were": true, "been": true, "be": true, "have": true, "has": true,
	"had": true, "do": true, "does": true, "did": true, "will": true, "would": true,
	"could": true, "should": true, "may": true, "might": true, "must": true,
	"shall": true, "can": true, "need": true, "dare": true, "ought": true,
	"used": true, "i": true, "me": true, "my": true, "myself": true, "we": true,
	"our": true, "ours": true, "ourselves": true, "you": true, "your": true,
	"yours": true, "yourself": true, "yourselves": true, "he": true, "him": true,
	"his": true, "himself": true, "she": true, "her": true, "hers": true,
	"herself": true, "it": true, "its": true, "itself": true, "they": true,
	"them": true, "their": true, "theirs": true, "themselves": true, "what": true,
	"which": true, "who": true, "whom": true, "this": true, "that": true,
	"these": true, "those": true, "am": true, "being": true, "here": true,
	"there": true, "when": true, "where": true, "why": true, "how": true,
	"all": true, "each": true, "few": true, "more": true, "most": true,
	"other": true, "some": true, "such": true, "no": true, "nor": true,
	"not": true, "only": true, "own": true, "same": true, "so": true,
	"than": true, "too": true, "very": true, "just": true, "also": true,
	"now": true, "want": true, "make": true, "get": true, "go": true,
	"going": true, "start": true, "stop": true, "keep": true,
	"try": true, "trying": true, "thing": true, "things": true, "something": true,
}

// ExtractKeywords extracts meaningful keywords from a list of goal/avoid strings.
// Filters out stopwords, short words, and normalizes to lowercase.
func ExtractKeywords(phrases []string) []string {
	keywordSet := make(map[string]bool)
	wordRegex := regexp.MustCompile(`[a-zA-Z]+`)

	for _, phrase := range phrases {
		words := wordRegex.FindAllString(phrase, -1)
		for _, word := range words {
			lower := strings.ToLower(word)
			// Skip stopwords and very short words
			if len(lower) < 3 || stopwords[lower] {
				continue
			}
			keywordSet[lower] = true
		}
	}

	// Convert set to slice
	keywords := make([]string, 0, len(keywordSet))
	for kw := range keywordSet {
		keywords = append(keywords, kw)
	}

	return keywords
}

// MatchScore calculates how well text matches a set of keywords.
// Returns a score from 0.0 to 1.0 based on keyword presence and density.
func MatchScore(text string, keywords []string) float64 {
	if len(keywords) == 0 {
		return 0.5 // Neutral score when no keywords to match
	}

	textLower := strings.ToLower(text)
	matchCount := 0
	totalWeight := 0.0

	for _, keyword := range keywords {
		weight := keywordWeight(keyword)
		totalWeight += weight

		if strings.Contains(textLower, keyword) {
			matchCount++
			// Bonus for exact word match (not just substring)
			if containsWord(textLower, keyword) {
				matchCount++ // Double weight for exact match
			}
		}
	}

	if totalWeight == 0 {
		return 0.5
	}

	// Calculate match ratio with diminishing returns
	ratio := float64(matchCount) / float64(len(keywords)*2) // *2 because exact matches count double
	if ratio > 1.0 {
		ratio = 1.0
	}

	return ratio
}

// AvoidanceScore calculates how well text avoids a set of keywords.
// Returns a score from 0.0 to 1.0 where higher means better avoidance.
func AvoidanceScore(text string, avoidKeywords []string) float64 {
	if len(avoidKeywords) == 0 {
		return 1.0 // Perfect avoidance when nothing to avoid
	}

	textLower := strings.ToLower(text)
	hitCount := 0

	for _, keyword := range avoidKeywords {
		keywordLower := strings.ToLower(keyword)
		// Check for substring match
		if strings.Contains(textLower, keywordLower) {
			hitCount++
		}
	}

	// More hits = lower score
	avoidRatio := 1.0 - (float64(hitCount) / float64(len(avoidKeywords)))
	if avoidRatio < 0 {
		avoidRatio = 0
	}

	return avoidRatio
}

// keywordWeight returns a weight for a keyword based on its characteristics.
// Longer, more specific words get higher weights.
func keywordWeight(keyword string) float64 {
	length := len(keyword)
	switch {
	case length >= 8:
		return 1.5 // Long words are more specific
	case length >= 5:
		return 1.0 // Normal weight
	default:
		return 0.7 // Short words are often less specific
	}
}

// containsWord checks if text contains keyword as a whole word.
func containsWord(text, keyword string) bool {
	// Simple word boundary check
	idx := strings.Index(text, keyword)
	if idx == -1 {
		return false
	}

	// Check character before
	if idx > 0 {
		before := rune(text[idx-1])
		if unicode.IsLetter(before) || unicode.IsDigit(before) {
			return false
		}
	}

	// Check character after
	endIdx := idx + len(keyword)
	if endIdx < len(text) {
		after := rune(text[endIdx])
		if unicode.IsLetter(after) || unicode.IsDigit(after) {
			return false
		}
	}

	return true
}

// ExtractKeywordsFromProfile extracts all relevant keywords from a profile.
// Combines goals and avoid lists into a single keyword set.
func ExtractKeywordsFromProfile(p *Profile) (goalKeywords, avoidKeywords []string) {
	goalKeywords = ExtractKeywords(p.Goals)
	avoidKeywords = ExtractKeywords(p.Avoid)
	return
}

// SimilarityScore calculates semantic similarity between two texts.
// Uses a simple keyword overlap approach.
func SimilarityScore(text1, text2 string) float64 {
	keywords1 := ExtractKeywords([]string{text1})
	keywords2 := ExtractKeywords([]string{text2})

	if len(keywords1) == 0 || len(keywords2) == 0 {
		return 0.0
	}

	// Count overlapping keywords
	set1 := make(map[string]bool)
	for _, kw := range keywords1 {
		set1[kw] = true
	}

	overlap := 0
	for _, kw := range keywords2 {
		if set1[kw] {
			overlap++
		}
	}

	// Jaccard similarity
	union := len(keywords1) + len(keywords2) - overlap
	if union == 0 {
		return 0.0
	}

	return float64(overlap) / float64(union)
}
