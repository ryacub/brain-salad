package scoring

import (
	"strings"
	"testing"
)

func TestRuleBasedScorer_Score(t *testing.T) {
	scorer := NewRuleBasedScorer()

	tests := []struct {
		name     string
		content  string
		telos    string
		minScore float64
		maxScore float64
	}{
		{
			name:     "high quality idea",
			content:  "Build an innovative mobile app to improve productivity through automation and smart scheduling",
			telos:    "Focus on productivity and innovation",
			minScore: 6.0,
			maxScore: 10.0,
		},
		{
			name:     "low quality idea",
			content:  "Maybe do something",
			telos:    "",
			minScore: 0.0,
			maxScore: 3.0,
		},
		{
			name:     "medium quality idea",
			content:  "Create a website for local businesses",
			telos:    "Support local communities",
			minScore: 4.0,
			maxScore: 7.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.Score(tt.content, tt.telos)

			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("Score %.1f outside expected range [%.1f, %.1f]",
					score, tt.minScore, tt.maxScore)
			}
		})
	}
}

func TestScoreKeywords(t *testing.T) {
	scorer := NewRuleBasedScorer()

	tests := []struct {
		content  string
		minScore float64
	}{
		{"innovation and growth", 2.0},
		{"maybe possibly unclear", 0.0},
		{"build create improve", 3.0},
		{"random text here", 0.0},
	}

	for _, tt := range tests {
		score := scorer.scoreKeywords(tt.content)
		if score < tt.minScore {
			t.Errorf("scoreKeywords(%s) = %.1f, want >= %.1f",
				tt.content, score, tt.minScore)
		}
	}
}

func TestScoreLength(t *testing.T) {
	scorer := NewRuleBasedScorer()

	tests := []struct {
		content  string
		expected float64
	}{
		{"short", 0.5},
		{"This is a good length idea with sufficient detail to be meaningful", 2.0},
		{strings.Repeat("word ", 200), 1.5}, // 1000 chars = medium-long length
	}

	for _, tt := range tests {
		score := scorer.scoreLength(tt.content)
		if score != tt.expected {
			t.Errorf("scoreLength() = %.1f, want %.1f", score, tt.expected)
		}
	}
}

func TestExtractImportantWords(t *testing.T) {
	text := "innovation sustainability productivity the and for"

	words := extractImportantWords(text)

	if len(words) != 3 {
		t.Errorf("Expected 3 important words, got %d: %v", len(words), words)
	}

	// Should not include common words
	for _, word := range words {
		if word == "the" || word == "and" || word == "for" {
			t.Errorf("Common word should not be included: %s", word)
		}
	}
}

func BenchmarkRuleBasedScore(b *testing.B) {
	scorer := NewRuleBasedScorer()
	content := "Build an innovative platform to improve sustainability through automation"
	telos := "Focus on innovation and sustainability"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scorer.Score(content, telos)
	}
}
