# Track 5C: Quality Metrics & Response Processing

**Phase**: 5 - LLM Integration
**Estimated Time**: 8-10 hours
**Dependencies**: 5A (needs types.go for LlmAnalysisResult)
**Can Run in Parallel**: Yes (after 5A creates types.go)

---

## Mission

You are implementing quality metrics tracking and LLM response processing for the Telos Idea Matrix Go application, following Test-Driven Development (TDD).

## Context

- The Rust implementation tracks LLM response quality in `src/quality_metrics_simple.rs`
- Response processing and validation in `src/response_processing.rs`
- Need to measure: completeness, consistency, confidence
- Must handle malformed LLM responses gracefully with fallback to rule-based scoring

## Reference Implementation

Review:
- `/home/user/brain-salad/src/quality_metrics_simple.rs` - Quality tracking
- `/home/user/brain-salad/src/response_processing.rs` - Response processing
- `/home/user/brain-salad/src/prompt_templates.rs` - Prompt management

## Your Task

Implement quality metrics and response processing using strict TDD methodology.

**IMPORTANT**: Wait for 5A to complete `types.go`. Once available, proceed in parallel with 5B.

## Directory Structure

Create files in `go/internal/llm/`:
- `quality/tracker.go` - Quality metrics tracking
- `quality/metrics.go` - Metric calculations
- `quality/tracker_test.go` - Quality tests
- `processing/processor.go` - Response processing
- `processing/validator.go` - Response validation
- `processing/processor_test.go` - Processing tests

## TDD Workflow (RED → GREEN → REFACTOR)

### STEP 1 - RED PHASE (Write Failing Tests)

Create `go/internal/llm/quality/tracker_test.go`:
- `TestQualityTracker_ScoreResponse()`
- `TestQualityTracker_ConfidenceLevel()`
- `TestQualityTracker_ConsistencyCheck()`
- `TestQualityTracker_GetAverageQuality()`

Create `go/internal/llm/processing/processor_test.go`:
- `TestProcessor_ParseLlmResponse()`
- `TestProcessor_ExtractScores()`
- `TestProcessor_HandleMalformedResponse()`
- `TestProcessor_FallbackToRuleBased()`

Run: `go test ./internal/llm/quality ./internal/llm/processing -v`
Expected: **ALL TESTS FAIL**

### STEP 2 - GREEN PHASE (Implement)

#### A. Implement `go/internal/llm/quality/metrics.go`:

```go
package quality

type QualityMetrics struct {
    Completeness float64 // 0.0-1.0: Are all scoring dimensions present?
    Consistency  float64 // 0.0-1.0: Do scores match explanations?
    Confidence   float64 // 0.0-1.0: Is LLM confident in analysis?
}

// CalculateCompleteness checks if all required fields are present
func CalculateCompleteness(hasScores, hasExplanations, hasFinalScore bool) float64 {
    score := 0.0
    if hasScores {
        score += 0.4
    }
    if hasExplanations {
        score += 0.3
    }
    if hasFinalScore {
        score += 0.3
    }
    return score
}

// CalculateConsistency checks if scores align with explanations
func CalculateConsistency(finalScore float64, sumOfComponents float64) float64 {
    if finalScore == 0 && sumOfComponents == 0 {
        return 1.0
    }
    
    diff := abs(finalScore - sumOfComponents)
    tolerance := 0.5 // Allow 0.5 point difference
    
    if diff <= tolerance {
        return 1.0
    }
    
    // Linear decay beyond tolerance
    consistency := 1.0 - (diff-tolerance)/10.0
    if consistency < 0 {
        return 0.0
    }
    return consistency
}

// CalculateConfidence based on explanation length and clarity
func CalculateConfidence(explanationLength int, hasQualifiers bool) float64 {
    confidence := 0.5 // Base confidence
    
    // Longer explanations = more confidence
    if explanationLength > 100 {
        confidence += 0.3
    } else if explanationLength > 50 {
        confidence += 0.2
    }
    
    // Qualifiers ("maybe", "possibly") reduce confidence
    if hasQualifiers {
        confidence -= 0.2
    }
    
    if confidence < 0 {
        return 0.0
    }
    if confidence > 1.0 {
        return 1.0
    }
    return confidence
}

func abs(x float64) float64 {
    if x < 0 {
        return -x
    }
    return x
}
```

#### B. Implement `go/internal/llm/quality/tracker.go`:

```go
package quality

import (
    "sync"
    "time"
    
    "github.com/rayyacub/telos-idea-matrix/internal/llm"
)

type QualityRecord struct {
    Timestamp time.Time
    Provider  string
    Metrics   QualityMetrics
    RawScore  float64
}

type QualityTracker struct {
    records []QualityRecord
    mu      sync.RWMutex
}

func NewQualityTracker() *QualityTracker {
    return &QualityTracker{
        records: make([]QualityRecord, 0),
    }
}

func (qt *QualityTracker) RecordAnalysis(result *llm.AnalysisResult) QualityMetrics {
    hasScores := result.Scores.MissionAlignment > 0 || 
                 result.Scores.AntiChallenge > 0 || 
                 result.Scores.StrategicFit > 0
    hasExplanations := len(result.Explanations) > 0
    hasFinalScore := result.FinalScore > 0

    completeness := CalculateCompleteness(hasScores, hasExplanations, hasFinalScore)
    
    sumOfComponents := result.Scores.MissionAlignment + 
                       result.Scores.AntiChallenge + 
                       result.Scores.StrategicFit
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
    
    qt.mu.Lock()
    qt.records = append(qt.records, QualityRecord{
        Timestamp: time.Now(),
        Provider:  result.Provider,
        Metrics:   metrics,
        RawScore:  result.FinalScore,
    })
    qt.mu.Unlock()
    
    return metrics
}

func (qt *QualityTracker) GetAverageQuality() QualityMetrics {
    qt.mu.RLock()
    defer qt.mu.RUnlock()
    
    if len(qt.records) == 0 {
        return QualityMetrics{}
    }
    
    var sumCompleteness, sumConsistency, sumConfidence float64
    for _, record := range qt.records {
        sumCompleteness += record.Metrics.Completeness
        sumConsistency += record.Metrics.Consistency
        sumConfidence += record.Metrics.Confidence
    }
    
    count := float64(len(qt.records))
    return QualityMetrics{
        Completeness: sumCompleteness / count,
        Consistency:  sumConsistency / count,
        Confidence:   sumConfidence / count,
    }
}

func containsQualifiers(text string) bool {
    qualifiers := []string{"maybe", "possibly", "perhaps", "might", "could be"}
    textLower := strings.ToLower(text)
    for _, q := range qualifiers {
        if strings.Contains(textLower, q) {
            return true
        }
    }
    return false
}
```

#### C. Implement `go/internal/llm/processing/processor.go`:

```go
package processing

import (
    "encoding/json"
    "fmt"
    "regexp"
    
    "github.com/rayyacub/telos-idea-matrix/internal/llm"
)

type Processor struct {
    fallbackProvider llm.Provider
}

func NewProcessor(fallback llm.Provider) *Processor {
    return &Processor{
        fallbackProvider: fallback,
    }
}

// ProcessResponse parses LLM JSON response with fallback on failure
func (p *Processor) ProcessResponse(rawResponse string, ideaContent string, telosPath string) (*llm.AnalysisResult, error) {
    // Try to parse as JSON
    var result llm.AnalysisResult
    if err := json.Unmarshal([]byte(rawResponse), &result); err != nil {
        // Malformed JSON, try to extract with regex
        extracted := p.extractWithRegex(rawResponse)
        if extracted != nil {
            return extracted, nil
        }
        
        // Complete failure, use fallback
        return p.useFallback(ideaContent, telosPath)
    }
    
    // Validate parsed result
    if !p.validateResult(&result) {
        return p.useFallback(ideaContent, telosPath)
    }
    
    return &result, nil
}

func (p *Processor) extractWithRegex(response string) *llm.AnalysisResult {
    // Try to extract scores even from malformed JSON
    missionRe := regexp.MustCompile(`"mission_alignment":\s*(\d+\.?\d*)`)
    antiChallengeRe := regexp.MustCompile(`"anti_challenge":\s*(\d+\.?\d*)`)
    strategicRe := regexp.MustCompile(`"strategic_fit":\s*(\d+\.?\d*)`)
    finalRe := regexp.MustCompile(`"final_score":\s*(\d+\.?\d*)`)
    
    matches := []*regexp.Regexp{missionRe, antiChallengeRe, strategicRe, finalRe}
    scores := make([]float64, 4)
    
    for i, re := range matches {
        match := re.FindStringSubmatch(response)
        if len(match) < 2 {
            return nil // Can't extract all scores
        }
        var err error
        scores[i], err = strconv.ParseFloat(match[1], 64)
        if err != nil {
            return nil
        }
    }
    
    return &llm.AnalysisResult{
        Scores: llm.ScoreBreakdown{
            MissionAlignment: scores[0],
            AntiChallenge:    scores[1],
            StrategicFit:     scores[2],
        },
        FinalScore:     scores[3],
        Recommendation: determineRecommendation(scores[3]),
        Explanations:   make(map[string]string),
        Provider:       "ollama_extracted",
    }
}

func (p *Processor) validateResult(result *llm.AnalysisResult) bool {
    // Check score ranges
    if result.Scores.MissionAlignment < 0 || result.Scores.MissionAlignment > 4.0 {
        return false
    }
    if result.Scores.AntiChallenge < 0 || result.Scores.AntiChallenge > 3.5 {
        return false
    }
    if result.Scores.StrategicFit < 0 || result.Scores.StrategicFit > 2.5 {
        return false
    }
    if result.FinalScore < 0 || result.FinalScore > 10.0 {
        return false
    }
    
    return true
}

func (p *Processor) useFallback(ideaContent string, telosPath string) (*llm.AnalysisResult, error) {
    req := llm.AnalysisRequest{
        IdeaContent: ideaContent,
        TelosPath:   telosPath,
    }
    return p.fallbackProvider.Analyze(req)
}

func determineRecommendation(score float64) string {
    if score >= 8.5 {
        return "PRIORITIZE NOW"
    } else if score >= 7.0 {
        return "GOOD ALIGNMENT"
    } else if score >= 5.0 {
        return "CONSIDER LATER"
    }
    return "AVOID FOR NOW"
}
```

Run: `go test ./internal/llm/quality ./internal/llm/processing -v`
Expected: **ALL TESTS PASS**

### STEP 3 - REFACTOR PHASE

- Add quality thresholds (reject low-quality responses)
- Optimize parsing for different LLM response formats
- Add retries on validation failure
- Extract common validation patterns

## Integration

1. Wire into Ollama provider (track quality after each analysis)
2. Add quality metrics to analytics endpoint
3. Add `tm llm quality` CLI command to show quality stats
4. Use processor for all LLM response handling

## Success Criteria

- ✅ All tests pass with >85% coverage
- ✅ Handles malformed responses gracefully
- ✅ Quality scoring accurate (validated against manual review)
- ✅ Fallback to rule-based scoring works
- ✅ Quality tracking provides useful insights

## Validation

```bash
# Unit tests
go test ./internal/llm/quality ./internal/llm/processing -v -cover

# Integration test with malformed response
# (manually inject bad JSON and verify fallback)
```

## Deliverables

- `go/internal/llm/quality/tracker.go`
- `go/internal/llm/quality/metrics.go`
- `go/internal/llm/quality/tracker_test.go`
- `go/internal/llm/processing/processor.go`
- `go/internal/llm/processing/validator.go`
- `go/internal/llm/processing/processor_test.go`
- Integration into Ollama provider
