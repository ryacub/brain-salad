package processing

import (
	"strings"
	"testing"
)

func TestNewValidator(t *testing.T) {
	v := NewValidator()
	if v == nil {
		t.Fatal("expected validator to be non-nil")
		return
	}

	// Verify default thresholds
	if v.maxMissionScore != 4.0 {
		t.Errorf("expected maxMissionScore 4.0, got %v", v.maxMissionScore)
	}
	if v.maxAntiChallengeScore != 3.5 {
		t.Errorf("expected maxAntiChallengeScore 3.5, got %v", v.maxAntiChallengeScore)
	}
	if v.maxStrategicScore != 2.5 {
		t.Errorf("expected maxStrategicScore 2.5, got %v", v.maxStrategicScore)
	}
	if v.maxFinalScore != 10.0 {
		t.Errorf("expected maxFinalScore 10.0, got %v", v.maxFinalScore)
	}
}

func TestValidator_Validate_ValidResult(t *testing.T) {
	v := NewValidator()

	result := &ProcessedResult{
		Scores: ScoreBreakdown{
			MissionAlignment: 3.5,
			AntiChallenge:    2.8,
			StrategicFit:     2.0,
		},
		FinalScore:     8.3,
		Recommendation: "GOOD ALIGNMENT",
		Explanations:   make(map[string]string),
	}

	err := v.Validate(result)
	if err != nil {
		t.Errorf("expected no error for valid result, got: %v", err)
	}
}

func TestValidator_Validate_InvalidMissionAlignment(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name  string
		score float64
	}{
		{"too high", 4.5},
		{"negative", -0.5},
		{"way too high", 10.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ProcessedResult{
				Scores: ScoreBreakdown{
					MissionAlignment: tt.score,
					AntiChallenge:    2.0,
					StrategicFit:     1.5,
				},
				FinalScore:     6.5,
				Recommendation: "CONSIDER LATER",
			}

			err := v.Validate(result)
			if err == nil {
				t.Error("expected validation error for invalid mission alignment")
			}
			if !strings.Contains(err.Error(), "Mission Alignment") {
				t.Errorf("expected error to mention 'Mission Alignment', got: %v", err)
			}
		})
	}
}

func TestValidator_Validate_InvalidAntiChallenge(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name  string
		score float64
	}{
		{"too high", 4.0},
		{"negative", -1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ProcessedResult{
				Scores: ScoreBreakdown{
					MissionAlignment: 3.0,
					AntiChallenge:    tt.score,
					StrategicFit:     1.5,
				},
				FinalScore:     6.5,
				Recommendation: "CONSIDER LATER",
			}

			err := v.Validate(result)
			if err == nil {
				t.Error("expected validation error for invalid anti-challenge")
			}
			if !strings.Contains(err.Error(), "Anti-Challenge") {
				t.Errorf("expected error to mention 'Anti-Challenge', got: %v", err)
			}
		})
	}
}

func TestValidator_Validate_InvalidStrategicFit(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name  string
		score float64
	}{
		{"too high", 3.0},
		{"negative", -0.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ProcessedResult{
				Scores: ScoreBreakdown{
					MissionAlignment: 3.0,
					AntiChallenge:    2.0,
					StrategicFit:     tt.score,
				},
				FinalScore:     6.5,
				Recommendation: "CONSIDER LATER",
			}

			err := v.Validate(result)
			if err == nil {
				t.Error("expected validation error for invalid strategic fit")
			}
			if !strings.Contains(err.Error(), "Strategic Fit") {
				t.Errorf("expected error to mention 'Strategic Fit', got: %v", err)
			}
		})
	}
}

func TestValidator_Validate_InvalidFinalScore(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name  string
		score float64
	}{
		{"too high", 11.0},
		{"negative", -1.0},
		{"way too high", 100.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ProcessedResult{
				Scores: ScoreBreakdown{
					MissionAlignment: 3.0,
					AntiChallenge:    2.0,
					StrategicFit:     1.5,
				},
				FinalScore:     tt.score,
				Recommendation: "CONSIDER LATER",
			}

			err := v.Validate(result)
			if err == nil {
				t.Error("expected validation error for invalid final score")
			}
			if !strings.Contains(err.Error(), "Final Score") {
				t.Errorf("expected error to mention 'Final Score', got: %v", err)
			}
		})
	}
}

func TestValidator_Validate_InvalidRecommendation(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name           string
		recommendation string
	}{
		{"empty", ""},
		{"invalid text", "MAYBE GOOD"},
		{"lowercase", "good alignment"},
		{"typo", "PRIORITIZE NOW!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ProcessedResult{
				Scores: ScoreBreakdown{
					MissionAlignment: 3.0,
					AntiChallenge:    2.0,
					StrategicFit:     1.5,
				},
				FinalScore:     6.5,
				Recommendation: tt.recommendation,
			}

			err := v.Validate(result)
			if err == nil {
				t.Errorf("expected validation error for recommendation '%s'", tt.recommendation)
			}
			if !strings.Contains(err.Error(), "recommendation") {
				t.Errorf("expected error to mention 'recommendation', got: %v", err)
			}
		})
	}
}

func TestValidator_Validate_AllValidRecommendations(t *testing.T) {
	v := NewValidator()

	validRecommendations := []string{
		"PRIORITIZE NOW",
		"GOOD ALIGNMENT",
		"CONSIDER LATER",
		"AVOID FOR NOW",
	}

	for _, rec := range validRecommendations {
		t.Run(rec, func(t *testing.T) {
			result := &ProcessedResult{
				Scores: ScoreBreakdown{
					MissionAlignment: 3.0,
					AntiChallenge:    2.0,
					StrategicFit:     1.5,
				},
				FinalScore:     6.5,
				Recommendation: rec,
			}

			err := v.Validate(result)
			if err != nil {
				t.Errorf("expected no error for valid recommendation '%s', got: %v", rec, err)
			}
		})
	}
}

func TestValidator_Validate_BoundaryValues(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name      string
		result    *ProcessedResult
		wantError bool
	}{
		{
			name: "all zeros (valid)",
			result: &ProcessedResult{
				Scores: ScoreBreakdown{
					MissionAlignment: 0.0,
					AntiChallenge:    0.0,
					StrategicFit:     0.0,
				},
				FinalScore:     0.0,
				Recommendation: "AVOID FOR NOW",
			},
			wantError: false,
		},
		{
			name: "all max values (valid)",
			result: &ProcessedResult{
				Scores: ScoreBreakdown{
					MissionAlignment: 4.0,
					AntiChallenge:    3.5,
					StrategicFit:     2.5,
				},
				FinalScore:     10.0,
				Recommendation: "PRIORITIZE NOW",
			},
			wantError: false,
		},
		{
			name: "mission alignment just over max (invalid)",
			result: &ProcessedResult{
				Scores: ScoreBreakdown{
					MissionAlignment: 4.01,
					AntiChallenge:    2.0,
					StrategicFit:     1.5,
				},
				FinalScore:     7.51,
				Recommendation: "GOOD ALIGNMENT",
			},
			wantError: true,
		},
		{
			name: "anti-challenge just over max (invalid)",
			result: &ProcessedResult{
				Scores: ScoreBreakdown{
					MissionAlignment: 3.0,
					AntiChallenge:    3.51,
					StrategicFit:     1.5,
				},
				FinalScore:     8.01,
				Recommendation: "GOOD ALIGNMENT",
			},
			wantError: true,
		},
		{
			name: "strategic fit just over max (invalid)",
			result: &ProcessedResult{
				Scores: ScoreBreakdown{
					MissionAlignment: 3.0,
					AntiChallenge:    2.0,
					StrategicFit:     2.51,
				},
				FinalScore:     7.51,
				Recommendation: "GOOD ALIGNMENT",
			},
			wantError: true,
		},
		{
			name: "final score just over max (invalid)",
			result: &ProcessedResult{
				Scores: ScoreBreakdown{
					MissionAlignment: 4.0,
					AntiChallenge:    3.5,
					StrategicFit:     2.5,
				},
				FinalScore:     10.01,
				Recommendation: "PRIORITIZE NOW",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.result)
			if tt.wantError && err == nil {
				t.Error("expected validation error")
			}
			if !tt.wantError && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestValidator_ValidateScore(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name      string
		score     float64
		min       float64
		max       float64
		scoreName string
		wantError bool
	}{
		{"valid score", 2.5, 0.0, 5.0, "Test", false},
		{"min boundary", 0.0, 0.0, 5.0, "Test", false},
		{"max boundary", 5.0, 0.0, 5.0, "Test", false},
		{"below min", -0.1, 0.0, 5.0, "Test", true},
		{"above max", 5.1, 0.0, 5.0, "Test", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.validateScore(tt.score, tt.min, tt.max, tt.scoreName)
			if tt.wantError && err == nil {
				t.Error("expected validation error")
			}
			if !tt.wantError && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
			if err != nil && !strings.Contains(err.Error(), tt.scoreName) {
				t.Errorf("expected error to contain score name '%s', got: %v", tt.scoreName, err)
			}
		})
	}
}

func TestValidator_ValidateRecommendation(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		recommendation string
		wantError      bool
	}{
		{"PRIORITIZE NOW", false},
		{"GOOD ALIGNMENT", false},
		{"CONSIDER LATER", false},
		{"AVOID FOR NOW", false},
		{"INVALID", true},
		{"", true},
		{"prioritize now", true}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.recommendation, func(t *testing.T) {
			err := v.validateRecommendation(tt.recommendation)
			if tt.wantError && err == nil {
				t.Errorf("expected error for recommendation '%s'", tt.recommendation)
			}
			if !tt.wantError && err != nil {
				t.Errorf("expected no error for recommendation '%s', got: %v", tt.recommendation, err)
			}
		})
	}
}

// BenchmarkValidator_Validate benchmarks validation performance
func BenchmarkValidator_Validate(b *testing.B) {
	v := NewValidator()
	result := &ProcessedResult{
		Scores: ScoreBreakdown{
			MissionAlignment: 3.5,
			AntiChallenge:    2.8,
			StrategicFit:     2.0,
		},
		FinalScore:     8.3,
		Recommendation: "GOOD ALIGNMENT",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.Validate(result)
	}
}
