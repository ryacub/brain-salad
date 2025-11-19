package processing

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestNewSimpleProcessor(t *testing.T) {
	processor := NewSimpleProcessor(nil)
	if processor == nil {
		t.Fatal("expected processor to be non-nil")
	}
}

func TestSimpleProcessor_Process_ValidJSON(t *testing.T) {
	processor := NewSimpleProcessor(nil)

	validResponse := `{
		"scores": {
			"mission_alignment": 3.5,
			"anti_challenge": 2.8,
			"strategic_fit": 2.0
		},
		"final_score": 8.3,
		"recommendation": "GOOD ALIGNMENT",
		"explanations": {
			"mission": "High alignment with AI goals",
			"anti": "Good execution speed",
			"strategic": "Fits current tech stack"
		}
	}`

	result, err := processor.Process(validResponse, "test idea")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result.Scores.MissionAlignment != 3.5 {
		t.Errorf("expected MissionAlignment 3.5, got %v", result.Scores.MissionAlignment)
	}
	if result.Scores.AntiChallenge != 2.8 {
		t.Errorf("expected AntiChallenge 2.8, got %v", result.Scores.AntiChallenge)
	}
	if result.Scores.StrategicFit != 2.0 {
		t.Errorf("expected StrategicFit 2.0, got %v", result.Scores.StrategicFit)
	}
	if result.FinalScore != 8.3 {
		t.Errorf("expected FinalScore 8.3, got %v", result.FinalScore)
	}
	if result.Recommendation != "GOOD ALIGNMENT" {
		t.Errorf("expected recommendation 'GOOD ALIGNMENT', got %s", result.Recommendation)
	}
	if len(result.Explanations) != 3 {
		t.Errorf("expected 3 explanations, got %d", len(result.Explanations))
	}
	if result.UsedFallback {
		t.Error("expected UsedFallback to be false")
	}
}

func TestSimpleProcessor_Process_InvalidScores(t *testing.T) {
	fallbackCalled := false
	fallbackFunc := func(ideaContent string) (*ProcessedResult, error) {
		fallbackCalled = true
		return &ProcessedResult{
			Scores: ScoreBreakdown{
				MissionAlignment: 2.0,
				AntiChallenge:    1.5,
				StrategicFit:     1.0,
			},
			FinalScore:     4.5,
			Recommendation: "CONSIDER LATER",
			Explanations:   make(map[string]string),
			UsedFallback:   true,
		}, nil
	}

	processor := NewSimpleProcessor(fallbackFunc)

	// Invalid score (mission_alignment > 4.0)
	invalidResponse := `{
		"scores": {
			"mission_alignment": 5.5,
			"anti_challenge": 2.8,
			"strategic_fit": 2.0
		},
		"final_score": 10.3,
		"recommendation": "GOOD ALIGNMENT",
		"explanations": {}
	}`

	result, err := processor.Process(invalidResponse, "test idea")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !fallbackCalled {
		t.Error("expected fallback to be called for invalid scores")
	}

	if !result.UsedFallback {
		t.Error("expected UsedFallback to be true")
	}

	if result.FinalScore != 4.5 {
		t.Errorf("expected fallback score 4.5, got %v", result.FinalScore)
	}
}

func TestSimpleProcessor_Process_RegexExtraction(t *testing.T) {
	processor := NewSimpleProcessor(nil)

	// Response with embedded JSON-like structure (not pure JSON)
	regexResponse := `Here's my analysis:
	"mission_alignment": 3.2,
	"anti_challenge": 2.5,
	"strategic_fit": 1.8,
	"final_score": 7.5
	Therefore, this is a good idea.`

	result, err := processor.Process(regexResponse, "test idea")
	if err != nil {
		t.Fatalf("expected regex extraction to succeed, got error: %v", err)
	}

	if result.Scores.MissionAlignment != 3.2 {
		t.Errorf("expected MissionAlignment 3.2, got %v", result.Scores.MissionAlignment)
	}
	if result.Scores.AntiChallenge != 2.5 {
		t.Errorf("expected AntiChallenge 2.5, got %v", result.Scores.AntiChallenge)
	}
	if result.Scores.StrategicFit != 1.8 {
		t.Errorf("expected StrategicFit 1.8, got %v", result.Scores.StrategicFit)
	}
	if result.FinalScore != 7.5 {
		t.Errorf("expected FinalScore 7.5, got %v", result.FinalScore)
	}
}

func TestSimpleProcessor_Process_FallbackOnJSONError(t *testing.T) {
	fallbackCalled := false
	fallbackFunc := func(ideaContent string) (*ProcessedResult, error) {
		fallbackCalled = true
		if ideaContent != "test idea" {
			t.Errorf("expected ideaContent 'test idea', got %s", ideaContent)
		}
		return &ProcessedResult{
			Scores: ScoreBreakdown{
				MissionAlignment: 2.0,
				AntiChallenge:    1.5,
				StrategicFit:     1.0,
			},
			FinalScore:     4.5,
			Recommendation: "CONSIDER LATER",
			Explanations:   make(map[string]string),
			UsedFallback:   true,
		}, nil
	}

	processor := NewSimpleProcessor(fallbackFunc)

	// Invalid response that can't be parsed
	invalidResponse := "This is completely unparseable text"

	result, err := processor.Process(invalidResponse, "test idea")
	if err != nil {
		t.Fatalf("expected no error with fallback, got: %v", err)
	}

	if !fallbackCalled {
		t.Error("expected fallback to be called")
	}
	if !result.UsedFallback {
		t.Error("expected UsedFallback to be true")
	}
}

func TestSimpleProcessor_Process_FallbackError(t *testing.T) {
	fallbackFunc := func(ideaContent string) (*ProcessedResult, error) {
		return nil, errors.New("fallback failed")
	}

	processor := NewSimpleProcessor(fallbackFunc)

	invalidResponse := "unparseable"

	_, err := processor.Process(invalidResponse, "test idea")
	if err == nil {
		t.Fatal("expected error when fallback fails")
	}
}

func TestSimpleProcessor_Process_NoFallback(t *testing.T) {
	processor := NewSimpleProcessor(nil)

	invalidResponse := "completely invalid"

	_, err := processor.Process(invalidResponse, "test idea")
	if err == nil {
		t.Fatal("expected error without fallback")
	}
}

func TestSimpleProcessor_Validate(t *testing.T) {
	processor := NewSimpleProcessor(nil)

	tests := []struct {
		name   string
		result *ProcessedResult
		want   bool
	}{
		{
			name: "valid scores",
			result: &ProcessedResult{
				Scores: ScoreBreakdown{
					MissionAlignment: 3.5,
					AntiChallenge:    2.8,
					StrategicFit:     2.0,
				},
				FinalScore: 8.3,
			},
			want: true,
		},
		{
			name: "mission alignment too high",
			result: &ProcessedResult{
				Scores: ScoreBreakdown{
					MissionAlignment: 4.5,
					AntiChallenge:    2.8,
					StrategicFit:     2.0,
				},
				FinalScore: 9.3,
			},
			want: false,
		},
		{
			name: "mission alignment negative",
			result: &ProcessedResult{
				Scores: ScoreBreakdown{
					MissionAlignment: -0.5,
					AntiChallenge:    2.8,
					StrategicFit:     2.0,
				},
				FinalScore: 4.3,
			},
			want: false,
		},
		{
			name: "anti-challenge too high",
			result: &ProcessedResult{
				Scores: ScoreBreakdown{
					MissionAlignment: 3.0,
					AntiChallenge:    4.0,
					StrategicFit:     2.0,
				},
				FinalScore: 9.0,
			},
			want: false,
		},
		{
			name: "strategic fit too high",
			result: &ProcessedResult{
				Scores: ScoreBreakdown{
					MissionAlignment: 3.0,
					AntiChallenge:    2.0,
					StrategicFit:     3.0,
				},
				FinalScore: 8.0,
			},
			want: false,
		},
		{
			name: "final score too high",
			result: &ProcessedResult{
				Scores: ScoreBreakdown{
					MissionAlignment: 3.0,
					AntiChallenge:    2.0,
					StrategicFit:     2.0,
				},
				FinalScore: 11.0,
			},
			want: false,
		},
		{
			name: "final score negative",
			result: &ProcessedResult{
				Scores: ScoreBreakdown{
					MissionAlignment: 0,
					AntiChallenge:    0,
					StrategicFit:     0,
				},
				FinalScore: -1.0,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := processor.validate(tt.result)
			if got != tt.want {
				t.Errorf("validate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetermineRecommendation(t *testing.T) {
	tests := []struct {
		score float64
		want  string
	}{
		{9.5, "PRIORITIZE NOW"},
		{8.5, "PRIORITIZE NOW"},
		{8.0, "GOOD ALIGNMENT"},
		{7.5, "GOOD ALIGNMENT"},
		{7.0, "GOOD ALIGNMENT"},
		{6.5, "CONSIDER LATER"},
		{5.5, "CONSIDER LATER"},
		{5.0, "CONSIDER LATER"},
		{4.5, "AVOID FOR NOW"},
		{3.0, "AVOID FOR NOW"},
		{1.0, "AVOID FOR NOW"},
		{0.0, "AVOID FOR NOW"},
	}

	for _, tt := range tests {
		t.Run("", func(_ *testing.T) {
			got := DetermineRecommendation(tt.score)
			if got != tt.want {
				t.Errorf("DetermineRecommendation(%v) = %v, want %v", tt.score, got, tt.want)
			}
		})
	}
}

func TestSimpleProcessor_ExtractWithRegex_PartialMatch(t *testing.T) {
	processor := NewSimpleProcessor(nil)

	// Missing strategic_fit
	partialResponse := `"mission_alignment": 3.0, "anti_challenge": 2.0`

	result := processor.extractWithRegex(partialResponse)
	if result != nil {
		t.Error("expected nil for partial regex match")
	}
}

func TestSimpleProcessor_Process_EmptyExplanations(t *testing.T) {
	processor := NewSimpleProcessor(nil)

	// Valid JSON but no explanations
	response := `{
		"scores": {
			"mission_alignment": 3.0,
			"anti_challenge": 2.0,
			"strategic_fit": 1.5
		},
		"final_score": 6.5,
		"recommendation": "CONSIDER LATER"
	}`

	result, err := processor.Process(response, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Explanations == nil {
		t.Error("expected explanations map to be initialized")
	}
	if len(result.Explanations) != 0 {
		t.Errorf("expected empty explanations, got %d", len(result.Explanations))
	}
}

func TestSimpleProcessor_Process_ZeroScores(t *testing.T) {
	processor := NewSimpleProcessor(nil)

	response := `{
		"scores": {
			"mission_alignment": 0.0,
			"anti_challenge": 0.0,
			"strategic_fit": 0.0
		},
		"final_score": 0.0,
		"recommendation": "AVOID FOR NOW",
		"explanations": {}
	}`

	result, err := processor.Process(response, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.FinalScore != 0.0 {
		t.Errorf("expected FinalScore 0.0, got %v", result.FinalScore)
	}
}

// BenchmarkSimpleProcessor_Process_JSON benchmarks JSON parsing
func BenchmarkSimpleProcessor_Process_JSON(b *testing.B) {
	processor := NewSimpleProcessor(nil)
	response := `{
		"scores": {
			"mission_alignment": 3.5,
			"anti_challenge": 2.8,
			"strategic_fit": 2.0
		},
		"final_score": 8.3,
		"recommendation": "GOOD ALIGNMENT",
		"explanations": {"test": "value"}
	}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = processor.Process(response, "test idea")
	}
}

// BenchmarkSimpleProcessor_Process_Regex benchmarks regex extraction
func BenchmarkSimpleProcessor_Process_Regex(b *testing.B) {
	processor := NewSimpleProcessor(nil)
	response := `"mission_alignment": 3.2, "anti_challenge": 2.5, "strategic_fit": 1.8, "final_score": 7.5`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = processor.Process(response, "test idea")
	}
}

func TestSimpleProcessor_ExtractWithRegex_DecimalParsing(t *testing.T) {
	processor := NewSimpleProcessor(nil)

	// Test with various decimal formats
	tests := []struct {
		name     string
		response string
		wantNil  bool
	}{
		{
			name: "integers",
			response: `"mission_alignment": 3, "anti_challenge": 2, ` +
				`"strategic_fit": 1, "final_score": 6`,
			wantNil: false,
		},
		{
			name: "one decimal place",
			response: `"mission_alignment": 3.5, "anti_challenge": 2.5, ` +
				`"strategic_fit": 1.5, "final_score": 7.5`,
			wantNil: false,
		},
		{
			name: "two decimal places",
			response: `"mission_alignment": 3.25, "anti_challenge": 2.75, ` +
				`"strategic_fit": 1.50, "final_score": 7.50`,
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.extractWithRegex(tt.response)
			if tt.wantNil && result != nil {
				t.Error("expected nil result")
			}
			if !tt.wantNil && result == nil {
				t.Error("expected non-nil result")
			}
		})
	}
}

func TestProcessedResult_ExplanationsInitialized(t *testing.T) {
	processor := NewSimpleProcessor(nil)

	// Create a response struct with nil explanations
	type Response struct {
		Scores struct {
			MissionAlignment float64 `json:"mission_alignment"`
			AntiChallenge    float64 `json:"anti_challenge"`
			StrategicFit     float64 `json:"strategic_fit"`
		} `json:"scores"`
		FinalScore     float64           `json:"final_score"`
		Recommendation string            `json:"recommendation"`
		Explanations   map[string]string `json:"explanations"`
	}

	resp := Response{}
	resp.Scores.MissionAlignment = 3.0
	resp.Scores.AntiChallenge = 2.0
	resp.Scores.StrategicFit = 1.5
	resp.FinalScore = 6.5
	resp.Recommendation = "CONSIDER LATER"
	resp.Explanations = nil

	jsonBytes, _ := json.Marshal(resp)

	result, err := processor.Process(string(jsonBytes), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify explanations is initialized even if nil in JSON
	if result.Explanations == nil {
		t.Error("expected explanations to be initialized")
	}
}
