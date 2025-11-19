package quality

import (
	"testing"
)

func TestCalculateCompleteness(t *testing.T) {
	tests := []struct {
		name            string
		hasScores       bool
		hasExplanations bool
		hasFinalScore   bool
		want            float64
	}{
		{
			name:            "all present",
			hasScores:       true,
			hasExplanations: true,
			hasFinalScore:   true,
			want:            1.0,
		},
		{
			name:            "only scores",
			hasScores:       true,
			hasExplanations: false,
			hasFinalScore:   false,
			want:            0.4,
		},
		{
			name:            "only explanations",
			hasScores:       false,
			hasExplanations: true,
			hasFinalScore:   false,
			want:            0.3,
		},
		{
			name:            "only final score",
			hasScores:       false,
			hasExplanations: false,
			hasFinalScore:   true,
			want:            0.3,
		},
		{
			name:            "scores and final score",
			hasScores:       true,
			hasExplanations: false,
			hasFinalScore:   true,
			want:            0.7,
		},
		{
			name:            "scores and explanations",
			hasScores:       true,
			hasExplanations: true,
			hasFinalScore:   false,
			want:            0.7,
		},
		{
			name:            "explanations and final score",
			hasScores:       false,
			hasExplanations: true,
			hasFinalScore:   true,
			want:            0.6,
		},
		{
			name:            "nothing present",
			hasScores:       false,
			hasExplanations: false,
			hasFinalScore:   false,
			want:            0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateCompleteness(tt.hasScores, tt.hasExplanations, tt.hasFinalScore)
			if got != tt.want {
				t.Errorf("CalculateCompleteness() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateConsistency(t *testing.T) {
	tests := []struct {
		name            string
		finalScore      float64
		sumOfComponents float64
		want            float64
	}{
		{
			name:            "both zero",
			finalScore:      0.0,
			sumOfComponents: 0.0,
			want:            1.0,
		},
		{
			name:            "exact match",
			finalScore:      8.0,
			sumOfComponents: 8.0,
			want:            1.0,
		},
		{
			name:            "within tolerance (0.3 diff)",
			finalScore:      8.0,
			sumOfComponents: 7.7,
			want:            1.0,
		},
		{
			name:            "within tolerance (0.5 diff)",
			finalScore:      8.0,
			sumOfComponents: 7.5,
			want:            1.0,
		},
		{
			name:            "slightly beyond tolerance (1.0 diff)",
			finalScore:      8.0,
			sumOfComponents: 7.0,
			want:            0.95,
		},
		{
			name:            "large difference (5.0 diff)",
			finalScore:      8.0,
			sumOfComponents: 3.0,
			want:            0.55,
		},
		{
			name:            "very large difference (10.0 diff)",
			finalScore:      10.0,
			sumOfComponents: 0.0,
			want:            0.05,
		},
		{
			name:            "huge difference (>10.0 diff)",
			finalScore:      15.0,
			sumOfComponents: 0.0,
			want:            0.0,
		},
		{
			name:            "reverse large difference",
			finalScore:      3.0,
			sumOfComponents: 8.0,
			want:            0.55,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateConsistency(tt.finalScore, tt.sumOfComponents)
			if !floatEquals(got, tt.want, 0.001) {
				t.Errorf("CalculateConsistency(%v, %v) = %v, want %v",
					tt.finalScore, tt.sumOfComponents, got, tt.want)
			}
		})
	}
}

func TestCalculateConfidence(t *testing.T) {
	tests := []struct {
		name              string
		explanationLength int
		hasQualifiers     bool
		want              float64
	}{
		{
			name:              "long explanation, no qualifiers",
			explanationLength: 150,
			hasQualifiers:     false,
			want:              0.8,
		},
		{
			name:              "medium explanation, no qualifiers",
			explanationLength: 75,
			hasQualifiers:     false,
			want:              0.7,
		},
		{
			name:              "short explanation, no qualifiers",
			explanationLength: 30,
			hasQualifiers:     false,
			want:              0.5,
		},
		{
			name:              "long explanation with qualifiers",
			explanationLength: 150,
			hasQualifiers:     true,
			want:              0.6,
		},
		{
			name:              "medium explanation with qualifiers",
			explanationLength: 75,
			hasQualifiers:     true,
			want:              0.5,
		},
		{
			name:              "short explanation with qualifiers",
			explanationLength: 30,
			hasQualifiers:     true,
			want:              0.3,
		},
		{
			name:              "no explanation, no qualifiers",
			explanationLength: 0,
			hasQualifiers:     false,
			want:              0.5,
		},
		{
			name:              "no explanation with qualifiers",
			explanationLength: 0,
			hasQualifiers:     true,
			want:              0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateConfidence(tt.explanationLength, tt.hasQualifiers)
			if !floatEquals(got, tt.want, 0.001) {
				t.Errorf("CalculateConfidence(%v, %v) = %v, want %v",
					tt.explanationLength, tt.hasQualifiers, got, tt.want)
			}
		})
	}
}

func TestAbs(t *testing.T) {
	tests := []struct {
		input float64
		want  float64
	}{
		{5.0, 5.0},
		{-5.0, 5.0},
		{0.0, 0.0},
		{3.14, 3.14},
		{-3.14, 3.14},
	}

	for _, tt := range tests {
		got := abs(tt.input)
		if got != tt.want {
			t.Errorf("abs(%v) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestContainsQualifiers(t *testing.T) {
	tests := []struct {
		text string
		want bool
	}{
		{"This is definitely a good idea", false},
		{"Maybe this could work", true},
		{"This might be a good approach", true},
		{"Possibly we should consider this", true},
		{"Perhaps this is worth exploring", true},
		{"This could be interesting", true},
		{"MAYBE in uppercase", true},
		{"Complex sentence, but maybe worth it", true},
		{"No uncertainty here", false},
		{"", false},
		{"Just normal text without qualifiers", false},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got := containsQualifiers(tt.text)
			if got != tt.want {
				t.Errorf("containsQualifiers(%q) = %v, want %v", tt.text, got, tt.want)
			}
		})
	}
}

func TestCalculateConfidence_BoundaryConditions(t *testing.T) {
	tests := []struct {
		name              string
		explanationLength int
		hasQualifiers     bool
		wantMin           float64
		wantMax           float64
	}{
		{
			name:              "should not exceed 1.0",
			explanationLength: 1000,
			hasQualifiers:     false,
			wantMin:           0.0,
			wantMax:           1.0,
		},
		{
			name:              "should not go below 0.0",
			explanationLength: 0,
			hasQualifiers:     true,
			wantMin:           0.0,
			wantMax:           1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateConfidence(tt.explanationLength, tt.hasQualifiers)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("CalculateConfidence() = %v, want between %v and %v",
					got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCalculateConsistency_Symmetry(t *testing.T) {
	// Test that consistency is symmetric: f(a,b) == f(b,a)
	testCases := []struct {
		a float64
		b float64
	}{
		{8.0, 7.0},
		{5.5, 6.5},
		{10.0, 9.0},
		{3.0, 8.0},
	}

	for _, tc := range testCases {
		forward := CalculateConsistency(tc.a, tc.b)
		reverse := CalculateConsistency(tc.b, tc.a)

		if forward != reverse {
			t.Errorf("Consistency not symmetric: CalculateConsistency(%v,%v) = %v, CalculateConsistency(%v,%v) = %v",
				tc.a, tc.b, forward, tc.b, tc.a, reverse)
		}
	}
}

// BenchmarkCalculateCompleteness benchmarks completeness calculation
func BenchmarkCalculateCompleteness(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CalculateCompleteness(true, true, true)
	}
}

// BenchmarkCalculateConsistency benchmarks consistency calculation
func BenchmarkCalculateConsistency(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CalculateConsistency(8.0, 7.5)
	}
}

// BenchmarkCalculateConfidence benchmarks confidence calculation
func BenchmarkCalculateConfidence(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CalculateConfidence(100, false)
	}
}

// BenchmarkContainsQualifiers benchmarks qualifier detection
func BenchmarkContainsQualifiers(b *testing.B) {
	text := "This might be a good approach that could work"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		containsQualifiers(text)
	}
}

// Helper functions

// floatEquals compares two floats with a tolerance
func floatEquals(a, b, tolerance float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff <= tolerance
}
