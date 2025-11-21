package dump

import (
	"strings"
	"testing"

	"github.com/rayyacub/telos-idea-matrix/internal/cliutil"
	"github.com/rayyacub/telos-idea-matrix/internal/llm"
)

func TestGetScoreIndicator(t *testing.T) {
	tests := []struct {
		score    float64
		expected string
	}{
		{0.0, "[░░░░░░░░░░]"},
		{5.0, "[█████░░░░░]"},
		{10.0, "[██████████]"},
		{7.5, "[███████░░░]"},
		{3.2, "[███░░░░░░░]"},
		{-1.0, "[░░░░░░░░░░]"}, // negative should be 0 bars
		{11.5, "[██████████]"}, // over 10 should be 10 bars
	}

	for _, tt := range tests {
		result := getScoreIndicator(tt.score)
		if result != tt.expected {
			t.Errorf("getScoreIndicator(%.1f) = %s, want %s",
				tt.score, result, tt.expected)
		}
	}
}

func TestGetRecommendationIndicator(t *testing.T) {
	tests := []struct {
		rec      string
		contains string
	}{
		{"PURSUE", "✓"},
		{"STRONG PURSUE", "✓"},
		{"pursue", "✓"},
		{"CONSIDER", "⏸"},
		{"MODERATE", "⏸"},
		{"AVOID", "✗"},
		{"WEAK", "✗"},
		{"DEFER", "✗"},
		{"Unknown", "?"}, // Should return "?" for unknown
	}

	for _, tt := range tests {
		result := getRecommendationIndicator(tt.rec)
		if !strings.Contains(result, tt.contains) {
			t.Errorf("getRecommendationIndicator(%s) = %s, should contain %s",
				tt.rec, result, tt.contains)
		}
	}
}

func TestWrapTextSimple(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		validate func(string) bool
	}{
		{
			name:  "short text",
			text:  "Hello world",
			width: 20,
			validate: func(result string) bool {
				// Should have indent
				return strings.HasPrefix(result, "  ")
			},
		},
		{
			name:  "long text",
			text:  "This is a long piece of text that should be wrapped properly when it exceeds the specified width",
			width: 30,
			validate: func(result string) bool {
				// Should have multiple lines
				lines := strings.Split(result, "\n")
				if len(lines) < 2 {
					return false
				}
				// Each line should start with indent
				for _, line := range lines {
					if !strings.HasPrefix(line, "  ") {
						return false
					}
					// Check line length (excluding indent)
					if len(line) > 32 { // width + indent
						return false
					}
				}
				return true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrapTextSimple(tt.text, tt.width)
			if !tt.validate(result) {
				t.Errorf("wrapTextSimple validation failed for %s\nResult: %q", tt.name, result)
			}
		})
	}
}

func TestGetProviderStatus(t *testing.T) {
	// Create a mock available provider
	availableProvider := &mockProvider{
		name:      "test_available",
		available: true,
	}

	// Create a mock unavailable provider
	unavailableProvider := &mockProvider{
		name:      "test_unavailable",
		available: false,
	}

	tests := []struct {
		name     string
		provider llm.Provider
		expected string
	}{
		{
			name:     "available provider",
			provider: availableProvider,
			expected: "✓ Available and ready",
		},
		{
			name:     "unavailable provider",
			provider: unavailableProvider,
			expected: "✗ Not available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getProviderStatus(tt.provider)
			if result != tt.expected {
				t.Errorf("getProviderStatus() = %s, want %s", result, tt.expected)
			}
		})
	}
}

// mockProvider is a mock implementation of llm.Provider for testing
type mockProvider struct {
	name      string
	available bool
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) IsAvailable() bool {
	return m.available
}

func (m *mockProvider) Analyze(req llm.AnalysisRequest) (*llm.AnalysisResult, error) {
	return &llm.AnalysisResult{
		FinalScore:     7.5,
		Recommendation: "PURSUE",
		Provider:       m.name,
	}, nil
}

func TestTruncateText(t *testing.T) {
	tests := []struct {
		text     string
		maxLen   int
		expected string
	}{
		{"Hello", 10, "Hello"},
		{"Hello World", 5, "Hello..."},
		{"Test", 4, "Test"},
		{"This is a long text", 10, "This is a ..."},
	}

	for _, tt := range tests {
		result := cliutil.TruncateText(tt.text, tt.maxLen)
		if result != tt.expected {
			t.Errorf("cliutil.TruncateText(%q, %d) = %q, want %q",
				tt.text, tt.maxLen, result, tt.expected)
		}
	}
}
