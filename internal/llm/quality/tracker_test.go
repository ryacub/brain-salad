package quality

import (
	"sync"
	"testing"
	"time"
)

func TestNewSimpleTracker(t *testing.T) {
	tracker := NewSimpleTracker()
	if tracker == nil {
		t.Fatal("expected tracker to be non-nil")
		return
	}

	if tracker.records == nil {
		t.Error("expected records to be initialized")
	}

	if len(tracker.records) != 0 {
		t.Errorf("expected empty records, got %d", len(tracker.records))
	}
}

func TestSimpleTracker_Record(t *testing.T) {
	tracker := NewSimpleTracker()

	result := &SimpleResult{
		MissionAlignment: 3.5,
		AntiChallenge:    2.8,
		StrategicFit:     2.0,
		FinalScore:       8.3,
		Explanations: map[string]string{
			"mission":   "Good alignment with goals",
			"anti":      "Fast execution",
			"strategic": "Tech stack match",
		},
		Provider: "ollama",
	}

	metrics := tracker.Record(result)

	// Verify metrics are returned
	if metrics.Completeness == 0 {
		t.Error("expected non-zero completeness")
	}
	if metrics.Consistency == 0 {
		t.Error("expected non-zero consistency")
	}
	if metrics.Confidence == 0 {
		t.Error("expected non-zero confidence")
	}

	// Verify record was stored
	tracker.mu.RLock()
	recordCount := len(tracker.records)
	tracker.mu.RUnlock()

	if recordCount != 1 {
		t.Errorf("expected 1 record, got %d", recordCount)
	}
}

func TestSimpleTracker_Record_MultipleResults(t *testing.T) {
	tracker := NewSimpleTracker()

	results := []*SimpleResult{
		{
			MissionAlignment: 3.5,
			AntiChallenge:    2.8,
			StrategicFit:     2.0,
			FinalScore:       8.3,
			Explanations:     map[string]string{"test": "value"},
			Provider:         "ollama",
		},
		{
			MissionAlignment: 2.5,
			AntiChallenge:    1.8,
			StrategicFit:     1.0,
			FinalScore:       5.3,
			Explanations:     map[string]string{"test": "value"},
			Provider:         "rule_based",
		},
		{
			MissionAlignment: 4.0,
			AntiChallenge:    3.0,
			StrategicFit:     2.5,
			FinalScore:       9.5,
			Explanations:     map[string]string{"test": "value"},
			Provider:         "ollama",
		},
	}

	for _, result := range results {
		tracker.Record(result)
	}

	tracker.mu.RLock()
	recordCount := len(tracker.records)
	tracker.mu.RUnlock()

	if recordCount != 3 {
		t.Errorf("expected 3 records, got %d", recordCount)
	}
}

func TestSimpleTracker_GetAverage_EmptyTracker(t *testing.T) {
	tracker := NewSimpleTracker()

	avg := tracker.GetAverage()

	if avg.Completeness != 0 {
		t.Errorf("expected zero completeness for empty tracker, got %v", avg.Completeness)
	}
	if avg.Consistency != 0 {
		t.Errorf("expected zero consistency for empty tracker, got %v", avg.Consistency)
	}
	if avg.Confidence != 0 {
		t.Errorf("expected zero confidence for empty tracker, got %v", avg.Confidence)
	}
}

func TestSimpleTracker_GetAverage_SingleResult(t *testing.T) {
	tracker := NewSimpleTracker()

	result := &SimpleResult{
		MissionAlignment: 3.5,
		AntiChallenge:    2.8,
		StrategicFit:     2.0,
		FinalScore:       8.3,
		Explanations: map[string]string{
			"test": "This is a long explanation that should boost confidence score",
		},
		Provider: "ollama",
	}

	metrics := tracker.Record(result)
	avg := tracker.GetAverage()

	// With single result, average should equal the metrics
	if avg.Completeness != metrics.Completeness {
		t.Errorf("expected completeness %v, got %v", metrics.Completeness, avg.Completeness)
	}
	if avg.Consistency != metrics.Consistency {
		t.Errorf("expected consistency %v, got %v", metrics.Consistency, avg.Consistency)
	}
	if avg.Confidence != metrics.Confidence {
		t.Errorf("expected confidence %v, got %v", metrics.Confidence, avg.Confidence)
	}
}

func TestSimpleTracker_GetAverage_MultipleResults(t *testing.T) {
	tracker := NewSimpleTracker()

	results := []*SimpleResult{
		{
			MissionAlignment: 4.0,
			AntiChallenge:    3.5,
			StrategicFit:     2.5,
			FinalScore:       10.0,
			Explanations:     map[string]string{"key": "Very long and detailed explanation"},
			Provider:         "ollama",
		},
		{
			MissionAlignment: 0.0,
			AntiChallenge:    0.0,
			StrategicFit:     0.0,
			FinalScore:       0.0,
			Explanations:     map[string]string{},
			Provider:         "rule_based",
		},
	}

	for _, result := range results {
		tracker.Record(result)
	}

	avg := tracker.GetAverage()

	// Average should be between the two extremes
	if avg.Completeness < 0 || avg.Completeness > 1 {
		t.Errorf("completeness %v out of valid range [0,1]", avg.Completeness)
	}
	if avg.Consistency < 0 || avg.Consistency > 1 {
		t.Errorf("consistency %v out of valid range [0,1]", avg.Consistency)
	}
	if avg.Confidence < 0 || avg.Confidence > 1 {
		t.Errorf("confidence %v out of valid range [0,1]", avg.Confidence)
	}
}

func TestSimpleTracker_Record_ZeroScores(t *testing.T) {
	tracker := NewSimpleTracker()

	result := &SimpleResult{
		MissionAlignment: 0.0,
		AntiChallenge:    0.0,
		StrategicFit:     0.0,
		FinalScore:       0.0,
		Explanations:     map[string]string{},
		Provider:         "rule_based",
	}

	metrics := tracker.Record(result)

	// Zero scores are not considered as "having scores"
	// hasScores = false, hasExplanations = false, hasFinalScore = false
	// So completeness should be 0.0
	if metrics.Completeness != 0.0 {
		t.Errorf("expected completeness 0.0 for zero scores, got %v", metrics.Completeness)
	}
}

func TestSimpleTracker_Record_HasScoresDetection(t *testing.T) {
	tracker := NewSimpleTracker()

	tests := []struct {
		name   string
		result *SimpleResult
		want   float64 // expected minimum completeness
	}{
		{
			name: "has all scores",
			result: &SimpleResult{
				MissionAlignment: 3.0,
				AntiChallenge:    2.0,
				StrategicFit:     1.5,
				FinalScore:       6.5,
				Explanations:     map[string]string{"test": "value"},
				Provider:         "test",
			},
			want: 1.0,
		},
		{
			name: "has only final score",
			result: &SimpleResult{
				MissionAlignment: 0.0,
				AntiChallenge:    0.0,
				StrategicFit:     0.0,
				FinalScore:       5.0,
				Explanations:     map[string]string{},
				Provider:         "test",
			},
			want: 0.3,
		},
		{
			name: "has component scores but no final",
			result: &SimpleResult{
				MissionAlignment: 3.0,
				AntiChallenge:    2.0,
				StrategicFit:     1.5,
				FinalScore:       0.0,
				Explanations:     map[string]string{},
				Provider:         "test",
			},
			want: 0.4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := tracker.Record(tt.result)
			if metrics.Completeness < tt.want {
				t.Errorf("expected completeness >= %v, got %v", tt.want, metrics.Completeness)
			}
		})
	}
}

func TestSimpleTracker_ConcurrentAccess(t *testing.T) {
	tracker := NewSimpleTracker()

	// Test concurrent writes
	var wg sync.WaitGroup
	concurrency := 100

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			result := &SimpleResult{
				MissionAlignment: 3.0,
				AntiChallenge:    2.0,
				StrategicFit:     1.5,
				FinalScore:       6.5,
				Explanations:     map[string]string{"test": "value"},
				Provider:         "test",
			}

			tracker.Record(result)
		}(i)
	}

	wg.Wait()

	// Verify all records were stored
	tracker.mu.RLock()
	recordCount := len(tracker.records)
	tracker.mu.RUnlock()

	if recordCount != concurrency {
		t.Errorf("expected %d records, got %d", concurrency, recordCount)
	}

	// Test concurrent reads while writing
	done := make(chan bool)
	go func() {
		for i := 0; i < 10; i++ {
			result := &SimpleResult{
				MissionAlignment: 1.0,
				AntiChallenge:    1.0,
				StrategicFit:     1.0,
				FinalScore:       3.0,
				Explanations:     map[string]string{},
				Provider:         "test",
			}
			tracker.Record(result)
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	for i := 0; i < 10; i++ {
		_ = tracker.GetAverage()
		time.Sleep(1 * time.Millisecond)
	}

	<-done
}

func TestSimpleTracker_Record_Timestamp(t *testing.T) {
	tracker := NewSimpleTracker()

	before := time.Now()

	result := &SimpleResult{
		MissionAlignment: 3.0,
		AntiChallenge:    2.0,
		StrategicFit:     1.5,
		FinalScore:       6.5,
		Explanations:     map[string]string{},
		Provider:         "test",
	}

	tracker.Record(result)

	after := time.Now()

	tracker.mu.RLock()
	record := tracker.records[0]
	tracker.mu.RUnlock()

	if record.Timestamp.Before(before) || record.Timestamp.After(after) {
		t.Errorf("timestamp %v not within expected range [%v, %v]",
			record.Timestamp, before, after)
	}
}

func TestSimpleTracker_Record_ProviderTracking(t *testing.T) {
	tracker := NewSimpleTracker()

	providers := []string{"ollama", "rule_based", "claude", "ollama"}

	for _, provider := range providers {
		result := &SimpleResult{
			MissionAlignment: 3.0,
			AntiChallenge:    2.0,
			StrategicFit:     1.5,
			FinalScore:       6.5,
			Explanations:     map[string]string{},
			Provider:         provider,
		}
		tracker.Record(result)
	}

	tracker.mu.RLock()
	defer tracker.mu.RUnlock()

	if len(tracker.records) != 4 {
		t.Fatalf("expected 4 records, got %d", len(tracker.records))
	}

	for i, expectedProvider := range providers {
		if tracker.records[i].Provider != expectedProvider {
			t.Errorf("record %d: expected provider %s, got %s",
				i, expectedProvider, tracker.records[i].Provider)
		}
	}
}

func TestSimpleTracker_Record_ExplanationCalculation(t *testing.T) {
	tracker := NewSimpleTracker()

	tests := []struct {
		name          string
		explanations  map[string]string
		minConfidence float64
		maxConfidence float64
	}{
		{
			name:          "no explanations",
			explanations:  map[string]string{},
			minConfidence: 0.0,
			maxConfidence: 0.6,
		},
		{
			name: "short explanations",
			explanations: map[string]string{
				"key1": "Short",
			},
			minConfidence: 0.0,
			maxConfidence: 0.6,
		},
		{
			name: "long explanations",
			explanations: map[string]string{
				"key1": "This is a very long and detailed explanation that should boost the confidence score significantly",
			},
			minConfidence: 0.7,
			maxConfidence: 1.0,
		},
		{
			name: "explanations with qualifiers",
			explanations: map[string]string{
				"key1": "Maybe this could work possibly",
			},
			minConfidence: 0.0,
			maxConfidence: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &SimpleResult{
				MissionAlignment: 3.0,
				AntiChallenge:    2.0,
				StrategicFit:     1.5,
				FinalScore:       6.5,
				Explanations:     tt.explanations,
				Provider:         "test",
			}

			metrics := tracker.Record(result)

			if metrics.Confidence < tt.minConfidence || metrics.Confidence > tt.maxConfidence {
				t.Errorf("confidence %v not in expected range [%v, %v]",
					metrics.Confidence, tt.minConfidence, tt.maxConfidence)
			}
		})
	}
}

// BenchmarkSimpleTracker_Record benchmarks recording performance
func BenchmarkSimpleTracker_Record(b *testing.B) {
	tracker := NewSimpleTracker()
	result := &SimpleResult{
		MissionAlignment: 3.5,
		AntiChallenge:    2.8,
		StrategicFit:     2.0,
		FinalScore:       8.3,
		Explanations:     map[string]string{"test": "value"},
		Provider:         "ollama",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.Record(result)
	}
}

// BenchmarkSimpleTracker_GetAverage benchmarks average calculation
func BenchmarkSimpleTracker_GetAverage(b *testing.B) {
	tracker := NewSimpleTracker()

	// Pre-populate with some records
	for i := 0; i < 100; i++ {
		result := &SimpleResult{
			MissionAlignment: 3.0,
			AntiChallenge:    2.0,
			StrategicFit:     1.5,
			FinalScore:       6.5,
			Explanations:     map[string]string{},
			Provider:         "test",
		}
		tracker.Record(result)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.GetAverage()
	}
}
