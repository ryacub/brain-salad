package llm

import (
	"math"
	"testing"

	"github.com/ryacub/telos-idea-matrix/internal/models"
)

func TestConvertResultToAnalysis(t *testing.T) {
	tests := []struct {
		name   string
		result *AnalysisResult
	}{
		{
			name: "standard scores",
			result: &AnalysisResult{
				FinalScore: 7.5,
				Scores: ScoreBreakdown{
					MissionAlignment: 3.0,
					AntiChallenge:    2.5,
					StrategicFit:     2.0,
				},
			},
		},
		{
			name: "zero scores",
			result: &AnalysisResult{
				FinalScore: 0,
				Scores: ScoreBreakdown{
					MissionAlignment: 0,
					AntiChallenge:    0,
					StrategicFit:     0,
				},
			},
		},
		{
			name: "max scores",
			result: &AnalysisResult{
				FinalScore: 10.0,
				Scores: ScoreBreakdown{
					MissionAlignment: 4.0,
					AntiChallenge:    3.5,
					StrategicFit:     2.5,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := ConvertResultToAnalysis(tt.result)

			if analysis == nil {
				t.Fatal("Expected analysis to be non-nil")
				return
			}

			// Check final scores
			if analysis.FinalScore != tt.result.FinalScore {
				t.Errorf("FinalScore = %v, want %v", analysis.FinalScore, tt.result.FinalScore)
			}
			if analysis.RawScore != tt.result.FinalScore {
				t.Errorf("RawScore = %v, want %v", analysis.RawScore, tt.result.FinalScore)
			}

			// Check mission scores total
			if analysis.Mission.Total != tt.result.Scores.MissionAlignment {
				t.Errorf("Mission.Total = %v, want %v", analysis.Mission.Total, tt.result.Scores.MissionAlignment)
			}

			// Verify mission sub-scores sum to total (with floating point tolerance)
			missionSum := analysis.Mission.DomainExpertise + analysis.Mission.AIAlignment +
				analysis.Mission.ExecutionSupport + analysis.Mission.RevenuePotential
			if math.Abs(missionSum-analysis.Mission.Total) > 0.001 {
				t.Errorf("Mission sub-scores sum = %v, want %v", missionSum, analysis.Mission.Total)
			}

			// Check anti-challenge scores total
			if analysis.AntiChallenge.Total != tt.result.Scores.AntiChallenge {
				t.Errorf("AntiChallenge.Total = %v, want %v", analysis.AntiChallenge.Total, tt.result.Scores.AntiChallenge)
			}

			// Verify anti-challenge sub-scores sum to total
			antiSum := analysis.AntiChallenge.ContextSwitching + analysis.AntiChallenge.RapidPrototyping +
				analysis.AntiChallenge.Accountability + analysis.AntiChallenge.IncomeAnxiety
			if math.Abs(antiSum-analysis.AntiChallenge.Total) > 0.01 {
				t.Errorf("AntiChallenge sub-scores sum = %v, want %v", antiSum, analysis.AntiChallenge.Total)
			}

			// Check strategic scores total
			if analysis.Strategic.Total != tt.result.Scores.StrategicFit {
				t.Errorf("Strategic.Total = %v, want %v", analysis.Strategic.Total, tt.result.Scores.StrategicFit)
			}

			// Verify strategic sub-scores sum to total
			stratSum := analysis.Strategic.StackCompatibility + analysis.Strategic.ShippingHabit +
				analysis.Strategic.PublicAccountability + analysis.Strategic.RevenueTesting
			if math.Abs(stratSum-analysis.Strategic.Total) > 0.001 {
				t.Errorf("Strategic sub-scores sum = %v, want %v", stratSum, analysis.Strategic.Total)
			}
		})
	}
}

func TestAnalyzeWithProviderOverride(t *testing.T) {
	// Create manager with rule_based provider (always available)
	config := DefaultManagerConfig()
	manager := NewManager(config)

	// Create test telos
	telos := &models.Telos{
		Goals: []models.Goal{
			{ID: "goal-1", Description: "Build AI-powered productivity tools", Priority: 1},
		},
	}

	t.Run("default provider", func(t *testing.T) {
		analysis, err := manager.AnalyzeWithProviderOverride(
			"Build a task management app",
			"", // no provider override
			"", // no model override
			telos,
		)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if analysis == nil {
			t.Fatal("Expected analysis to be non-nil")
			return
		}

		if analysis.FinalScore < 0 || analysis.FinalScore > 10 {
			t.Errorf("FinalScore should be between 0-10, got %v", analysis.FinalScore)
		}
	})

	t.Run("with provider override", func(t *testing.T) {
		analysis, err := manager.AnalyzeWithProviderOverride(
			"Create an AI assistant",
			"rule_based", // explicit provider
			"",           // no model override
			telos,
		)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if analysis == nil {
			t.Fatal("Expected analysis to be non-nil")
			return
		}

		if analysis.FinalScore < 0 || analysis.FinalScore > 10 {
			t.Errorf("FinalScore should be between 0-10, got %v", analysis.FinalScore)
		}
	})

	t.Run("invalid provider", func(t *testing.T) {
		_, err := manager.AnalyzeWithProviderOverride(
			"Some idea",
			"nonexistent_provider",
			"",
			telos,
		)

		if err == nil {
			t.Error("Expected error for invalid provider, got nil")
		}
	})

	t.Run("model parameter reserved", func(t *testing.T) {
		// Model parameter should be ignored (reserved for future use)
		analysis, err := manager.AnalyzeWithProviderOverride(
			"Build something",
			"rule_based",
			"gpt-4", // should be ignored
			telos,
		)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if analysis == nil {
			t.Fatal("Expected analysis to be non-nil")
		}
	})
}

func TestConvertResultToAnalysis_ProportionalDistribution(t *testing.T) {
	// Test that the proportional distribution is correct
	result := &AnalysisResult{
		FinalScore: 8.0,
		Scores: ScoreBreakdown{
			MissionAlignment: 4.0, // max
			AntiChallenge:    3.5, // max
			StrategicFit:     2.5, // max
		},
	}

	analysis := ConvertResultToAnalysis(result)

	// Mission proportions: 0.30, 0.375, 0.20, 0.125 = 1.0
	expectedDomainExpertise := 4.0 * 0.30
	expectedAIAlignment := 4.0 * 0.375
	expectedExecutionSupport := 4.0 * 0.20
	expectedRevenuePotential := 4.0 * 0.125

	if math.Abs(analysis.Mission.DomainExpertise-expectedDomainExpertise) > 0.001 {
		t.Errorf("DomainExpertise = %v, want %v", analysis.Mission.DomainExpertise, expectedDomainExpertise)
	}
	if math.Abs(analysis.Mission.AIAlignment-expectedAIAlignment) > 0.001 {
		t.Errorf("AIAlignment = %v, want %v", analysis.Mission.AIAlignment, expectedAIAlignment)
	}
	if math.Abs(analysis.Mission.ExecutionSupport-expectedExecutionSupport) > 0.001 {
		t.Errorf("ExecutionSupport = %v, want %v", analysis.Mission.ExecutionSupport, expectedExecutionSupport)
	}
	if math.Abs(analysis.Mission.RevenuePotential-expectedRevenuePotential) > 0.001 {
		t.Errorf("RevenuePotential = %v, want %v", analysis.Mission.RevenuePotential, expectedRevenuePotential)
	}

	// Strategic proportions: 0.40, 0.32, 0.16, 0.12 = 1.0
	expectedStackCompatibility := 2.5 * 0.40
	expectedShippingHabit := 2.5 * 0.32
	expectedPublicAccountability := 2.5 * 0.16
	expectedRevenueTesting := 2.5 * 0.12

	if math.Abs(analysis.Strategic.StackCompatibility-expectedStackCompatibility) > 0.001 {
		t.Errorf("StackCompatibility = %v, want %v", analysis.Strategic.StackCompatibility, expectedStackCompatibility)
	}
	if math.Abs(analysis.Strategic.ShippingHabit-expectedShippingHabit) > 0.001 {
		t.Errorf("ShippingHabit = %v, want %v", analysis.Strategic.ShippingHabit, expectedShippingHabit)
	}
	if math.Abs(analysis.Strategic.PublicAccountability-expectedPublicAccountability) > 0.001 {
		t.Errorf("PublicAccountability = %v, want %v", analysis.Strategic.PublicAccountability, expectedPublicAccountability)
	}
	if math.Abs(analysis.Strategic.RevenueTesting-expectedRevenueTesting) > 0.001 {
		t.Errorf("RevenueTesting = %v, want %v", analysis.Strategic.RevenueTesting, expectedRevenueTesting)
	}
}
