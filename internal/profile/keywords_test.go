package profile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractKeywords_RemovesStopwords(t *testing.T) {
	phrases := []string{"I want to sell pottery at the market"}
	keywords := ExtractKeywords(phrases)

	assert.NotContains(t, keywords, "i")
	assert.NotContains(t, keywords, "to")
	assert.NotContains(t, keywords, "the")
	assert.NotContains(t, keywords, "at")
}

func TestExtractKeywords_RemovesShortWords(t *testing.T) {
	phrases := []string{"go do it now"}
	keywords := ExtractKeywords(phrases)

	assert.NotContains(t, keywords, "go")
	assert.NotContains(t, keywords, "do")
	assert.NotContains(t, keywords, "it")
}

func TestExtractKeywords_ExtractsMeaningfulWords(t *testing.T) {
	phrases := []string{"sell pottery at the farmer's market"}
	keywords := ExtractKeywords(phrases)

	assert.Contains(t, keywords, "sell")
	assert.Contains(t, keywords, "pottery")
	assert.Contains(t, keywords, "farmer")
	assert.Contains(t, keywords, "market")
}

func TestExtractKeywords_LowercasesWords(t *testing.T) {
	phrases := []string{"Build POTTERY Business"}
	keywords := ExtractKeywords(phrases)

	assert.Contains(t, keywords, "build")
	assert.Contains(t, keywords, "pottery")
	assert.Contains(t, keywords, "business")
	assert.NotContains(t, keywords, "Build")
	assert.NotContains(t, keywords, "POTTERY")
}

func TestExtractKeywords_HandlesMultiplePhrases(t *testing.T) {
	phrases := []string{"sell pottery", "finish projects"}
	keywords := ExtractKeywords(phrases)

	assert.Contains(t, keywords, "sell")
	assert.Contains(t, keywords, "pottery")
	assert.Contains(t, keywords, "finish")
	assert.Contains(t, keywords, "projects")
}

func TestExtractKeywords_DeduplicatesWords(t *testing.T) {
	phrases := []string{"pottery pottery pottery"}
	keywords := ExtractKeywords(phrases)

	count := 0
	for _, kw := range keywords {
		if kw == "pottery" {
			count++
		}
	}
	assert.Equal(t, 1, count)
}

func TestExtractKeywords_EmptyInput_ReturnsEmpty(t *testing.T) {
	keywords := ExtractKeywords([]string{})
	assert.Empty(t, keywords)
}

func TestMatchScore_NoKeywords_ReturnsNeutral(t *testing.T) {
	score := MatchScore("some idea text", []string{})
	assert.Equal(t, 0.5, score)
}

func TestMatchScore_FullMatch_ReturnsHigh(t *testing.T) {
	keywords := []string{"pottery", "market", "sell"}
	text := "I want to sell pottery at the market"

	score := MatchScore(text, keywords)
	assert.Greater(t, score, 0.5)
}

func TestMatchScore_NoMatch_ReturnsLow(t *testing.T) {
	keywords := []string{"pottery", "market", "sell"}
	text := "build a mobile app for tracking expenses"

	score := MatchScore(text, keywords)
	assert.Less(t, score, 0.5)
}

func TestMatchScore_PartialMatch_ReturnsMedium(t *testing.T) {
	keywords := []string{"pottery", "market", "sell", "handmade"}
	text := "sell things at market" // matches 2 of 4

	score := MatchScore(text, keywords)
	assert.Greater(t, score, 0.0)
	assert.Less(t, score, 1.0)
}

func TestMatchScore_CaseInsensitive(t *testing.T) {
	keywords := []string{"pottery"}
	text := "POTTERY is great"

	score := MatchScore(text, keywords)
	assert.Greater(t, score, 0.5)
}

func TestAvoidanceScore_NoAvoidWords_ReturnsPerfect(t *testing.T) {
	score := AvoidanceScore("build pottery", []string{})
	assert.Equal(t, 1.0, score)
}

func TestAvoidanceScore_NoMatches_ReturnsPerfect(t *testing.T) {
	avoidWords := []string{"wholesale", "inventory"}
	text := "sell pottery at farmer's market"

	score := AvoidanceScore(text, avoidWords)
	assert.Equal(t, 1.0, score)
}

func TestAvoidanceScore_AllMatches_ReturnsZero(t *testing.T) {
	avoidWords := []string{"wholesale", "inventory"}
	text := "wholesale pottery with large inventory"

	score := AvoidanceScore(text, avoidWords)
	assert.Equal(t, 0.0, score)
}

func TestAvoidanceScore_PartialMatch_ReturnsMedium(t *testing.T) {
	avoidWords := []string{"wholesale", "inventory", "mass production"}
	text := "small batch pottery, no wholesale" // matches 1 of 3

	score := AvoidanceScore(text, avoidWords)
	assert.Greater(t, score, 0.0)
	assert.Less(t, score, 1.0)
}

func TestAvoidanceScore_CaseInsensitive(t *testing.T) {
	avoidWords := []string{"wholesale"}
	text := "WHOLESALE is bad"

	score := AvoidanceScore(text, avoidWords)
	assert.Equal(t, 0.0, score)
}

func TestContainsWord_ExactMatch_ReturnsTrue(t *testing.T) {
	assert.True(t, containsWord("sell pottery", "sell"))
	assert.True(t, containsWord("pottery", "pottery"))
	assert.True(t, containsWord("sell pottery now", "pottery"))
}

func TestContainsWord_SubstringOnly_ReturnsFalse(t *testing.T) {
	// "pot" is substring of "pottery" but not a whole word
	assert.False(t, containsWord("pottery", "pot"))
	assert.False(t, containsWord("selling", "sell"))
}

func TestContainsWord_NotPresent_ReturnsFalse(t *testing.T) {
	assert.False(t, containsWord("pottery", "market"))
}

func TestSimilarityScore_IdenticalText_ReturnsHigh(t *testing.T) {
	score := SimilarityScore("sell pottery market", "sell pottery market")
	assert.Greater(t, score, 0.8)
}

func TestSimilarityScore_NoOverlap_ReturnsZero(t *testing.T) {
	score := SimilarityScore("sell pottery", "build software")
	assert.Equal(t, 0.0, score)
}

func TestSimilarityScore_PartialOverlap_ReturnsMedium(t *testing.T) {
	score := SimilarityScore("sell pottery online", "pottery business plan")
	assert.Greater(t, score, 0.0)
	assert.Less(t, score, 1.0)
}

func TestSimilarityScore_EmptyText_ReturnsZero(t *testing.T) {
	score := SimilarityScore("", "pottery")
	assert.Equal(t, 0.0, score)

	score = SimilarityScore("pottery", "")
	assert.Equal(t, 0.0, score)
}

func TestExtractKeywordsFromProfile_ExtractsBothLists(t *testing.T) {
	p := &Profile{
		Goals: []string{"sell pottery at market"},
		Avoid: []string{"wholesale accounts"},
	}

	goalKw, avoidKw := ExtractKeywordsFromProfile(p)

	assert.Contains(t, goalKw, "pottery")
	assert.Contains(t, goalKw, "market")
	assert.Contains(t, avoidKw, "wholesale")
	assert.Contains(t, avoidKw, "accounts")
}
