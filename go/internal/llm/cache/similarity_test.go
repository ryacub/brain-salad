package cache

import (
	"testing"
)

func TestNormalizeText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase conversion",
			input:    "HELLO WORLD",
			expected: "hello world",
		},
		{
			name:     "remove punctuation",
			input:    "Hello, World! How are you?",
			expected: "hello world how are you",
		},
		{
			name:     "remove special characters",
			input:    "test@#$%ing & stuff",
			expected: "test ing stuff",
		},
		{
			name:     "collapse multiple spaces",
			input:    "hello    world   test",
			expected: "hello world test",
		},
		{
			name:     "trim whitespace",
			input:    "  hello world  ",
			expected: "hello world",
		},
		{
			name:     "preserve numbers",
			input:    "Test 123 with numbers",
			expected: "test 123 with numbers",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeText(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeText(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "basic tokenization",
			input:    "hello world test",
			expected: []string{"hello", "world", "test"},
		},
		{
			name:     "remove stopwords",
			input:    "the quick brown fox jumps over the lazy dog",
			expected: []string{"quick", "brown", "fox", "jumps", "over", "lazy", "dog"},
		},
		{
			name:     "remove single characters",
			input:    "a b test c d word",
			expected: []string{"test", "word"},
		},
		{
			name:     "with punctuation",
			input:    "Hello, World! This is a test.",
			expected: []string{"hello", "world", "this", "test"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "only stopwords",
			input:    "the and or is",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Tokenize(tt.input)
			if len(result) != len(tt.expected) {
				t.Fatalf("Tokenize(%q) returned %d tokens, want %d\nGot: %v\nWant: %v",
					tt.input, len(result), len(tt.expected), result, tt.expected)
			}
			for i, token := range result {
				if token != tt.expected[i] {
					t.Errorf("Tokenize(%q)[%d] = %q, want %q", tt.input, i, token, tt.expected[i])
				}
			}
		})
	}
}

func TestJaccardSimilarity(t *testing.T) {
	tests := []struct {
		name      string
		text1     string
		text2     string
		threshold float64
		expected  bool // true if similarity should be >= threshold
	}{
		{
			name:      "identical texts",
			text1:     "build automation tool",
			text2:     "build automation tool",
			threshold: 1.0,
			expected:  true,
		},
		{
			name:      "very similar texts",
			text1:     "build automation tool",
			text2:     "create automation tool",
			threshold: 0.85,
			expected:  false, // only 2/3 tokens match (automation, tool)
		},
		{
			name:      "similar with different wording",
			text1:     "implement user authentication",
			text2:     "implement user authentication system",
			threshold: 0.85,
			expected:  false, // 3/4 = 0.75, below 0.85 threshold
		},
		{
			name:      "different texts",
			text1:     "build automation tool",
			text2:     "create web dashboard",
			threshold: 0.85,
			expected:  false,
		},
		{
			name:      "empty strings",
			text1:     "",
			text2:     "",
			threshold: 1.0,
			expected:  true,
		},
		{
			name:      "one empty string",
			text1:     "test",
			text2:     "",
			threshold: 0.0,
			expected:  true, // similarity should be 0.0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			similarity := JaccardSimilarity(tt.text1, tt.text2)
			if tt.expected {
				if similarity < tt.threshold {
					t.Errorf("JaccardSimilarity(%q, %q) = %.2f, want >= %.2f",
						tt.text1, tt.text2, similarity, tt.threshold)
				}
			} else {
				if similarity >= tt.threshold {
					t.Errorf("JaccardSimilarity(%q, %q) = %.2f, want < %.2f",
						tt.text1, tt.text2, similarity, tt.threshold)
				}
			}
		})
	}
}

func TestJaccardSimilarity_ExactValues(t *testing.T) {
	tests := []struct {
		name     string
		text1    string
		text2    string
		expected float64
		delta    float64
	}{
		{
			name:     "identical",
			text1:    "test word",
			text2:    "test word",
			expected: 1.0,
			delta:    0.01,
		},
		{
			name:     "no overlap",
			text1:    "hello world",
			text2:    "foo bar",
			expected: 0.0,
			delta:    0.01,
		},
		{
			name:     "50% overlap",
			text1:    "hello world",
			text2:    "hello foo",
			expected: 0.33, // 1 shared / 3 total = 0.33
			delta:    0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			similarity := JaccardSimilarity(tt.text1, tt.text2)
			if similarity < tt.expected-tt.delta || similarity > tt.expected+tt.delta {
				t.Errorf("JaccardSimilarity(%q, %q) = %.2f, want %.2f ± %.2f",
					tt.text1, tt.text2, similarity, tt.expected, tt.delta)
			}
		})
	}
}
