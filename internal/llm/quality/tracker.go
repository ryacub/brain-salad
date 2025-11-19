package quality

import (
	"sync"
	"time"
)

// SimpleResult is a simplified analysis result for quality tracking
type SimpleResult struct {
	MissionAlignment float64
	AntiChallenge    float64
	StrategicFit     float64
	FinalScore       float64
	Explanations     map[string]string
	Provider         string
}

// SimpleRecord represents a quality record
type SimpleRecord struct {
	Timestamp time.Time
	Provider  string
	Metrics   QualityMetrics
	RawScore  float64
}

// SimpleTracker tracks quality without depending on llm package
type SimpleTracker struct {
	records []SimpleRecord
	mu      sync.RWMutex
}

// NewSimpleTracker creates a new simple tracker
func NewSimpleTracker() *SimpleTracker {
	return &SimpleTracker{
		records: make([]SimpleRecord, 0),
	}
}

// Record records quality for a simple result
func (st *SimpleTracker) Record(result *SimpleResult) QualityMetrics {
	hasScores := result.MissionAlignment > 0 ||
		result.AntiChallenge > 0 ||
		result.StrategicFit > 0
	hasExplanations := len(result.Explanations) > 0
	hasFinalScore := result.FinalScore > 0

	completeness := CalculateCompleteness(hasScores, hasExplanations, hasFinalScore)

	sumOfComponents := result.MissionAlignment +
		result.AntiChallenge +
		result.StrategicFit
	consistency := CalculateConsistency(result.FinalScore, sumOfComponents)

	totalExplanationLength := 0
	hasQualifiers := false
	for _, exp := range result.Explanations {
		totalExplanationLength += len(exp)
		if containsQualifiers(exp) {
			hasQualifiers = true
		}
	}
	confidence := CalculateConfidence(totalExplanationLength, hasQualifiers)

	metrics := QualityMetrics{
		Completeness: completeness,
		Consistency:  consistency,
		Confidence:   confidence,
	}

	st.mu.Lock()
	st.records = append(st.records, SimpleRecord{
		Timestamp: time.Now(),
		Provider:  result.Provider,
		Metrics:   metrics,
		RawScore:  result.FinalScore,
	})
	st.mu.Unlock()

	return metrics
}

// GetAverage returns average quality metrics
func (st *SimpleTracker) GetAverage() QualityMetrics {
	st.mu.RLock()
	defer st.mu.RUnlock()

	if len(st.records) == 0 {
		return QualityMetrics{}
	}

	var sumCompleteness, sumConsistency, sumConfidence float64
	for _, record := range st.records {
		sumCompleteness += record.Metrics.Completeness
		sumConsistency += record.Metrics.Consistency
		sumConfidence += record.Metrics.Confidence
	}

	count := float64(len(st.records))
	return QualityMetrics{
		Completeness: sumCompleteness / count,
		Consistency:  sumConsistency / count,
		Confidence:   sumConfidence / count,
	}
}
