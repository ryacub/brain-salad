package processing

import (
	"encoding/json"
	"regexp"
	"strconv"
)

// ScoreBreakdown contains the three main scoring categories
type ScoreBreakdown struct {
	MissionAlignment float64
	AntiChallenge    float64
	StrategicFit     float64
}

// ProcessedResult represents the result after processing
type ProcessedResult struct {
	Scores         ScoreBreakdown
	FinalScore     float64
	Recommendation string
	Explanations   map[string]string
	Provider       string
	UsedFallback   bool
}

// FallbackFunc is called when processing fails
type FallbackFunc func(ideaContent string) (*ProcessedResult, error)

// SimpleProcessor handles LLM response processing
type SimpleProcessor struct {
	fallback FallbackFunc
}

// NewSimpleProcessor creates a new processor
func NewSimpleProcessor(fallbackFn FallbackFunc) *SimpleProcessor {
	return &SimpleProcessor{
		fallback: fallbackFn,
	}
}

// Process parses an LLM response and returns the result
func (sp *SimpleProcessor) Process(rawResponse string, ideaContent string) (*ProcessedResult, error) {
	// Try to parse JSON
	var jsonResp struct {
		Scores struct {
			MissionAlignment float64 `json:"mission_alignment"`
			AntiChallenge    float64 `json:"anti_challenge"`
			StrategicFit     float64 `json:"strategic_fit"`
		} `json:"scores"`
		FinalScore     float64           `json:"final_score"`
		Recommendation string            `json:"recommendation"`
		Explanations   map[string]string `json:"explanations"`
	}

	if err := json.Unmarshal([]byte(rawResponse), &jsonResp); err != nil {
		// Try regex extraction
		extracted := sp.extractWithRegex(rawResponse)
		if extracted != nil {
			return extracted, nil
		}

		// Use fallback
		if sp.fallback != nil {
			return sp.fallback(ideaContent)
		}

		return nil, err
	}

	result := &ProcessedResult{
		Scores: ScoreBreakdown{
			MissionAlignment: jsonResp.Scores.MissionAlignment,
			AntiChallenge:    jsonResp.Scores.AntiChallenge,
			StrategicFit:     jsonResp.Scores.StrategicFit,
		},
		FinalScore:     jsonResp.FinalScore,
		Recommendation: jsonResp.Recommendation,
		Explanations:   jsonResp.Explanations,
		UsedFallback:   false,
	}

	if result.Explanations == nil {
		result.Explanations = make(map[string]string)
	}

	// Validate
	if !sp.validate(result) {
		if sp.fallback != nil {
			result, err := sp.fallback(ideaContent)
			if err == nil {
				result.UsedFallback = true
			}
			return result, err
		}
	}

	return result, nil
}

// extractWithRegex tries to extract scores using regex
func (sp *SimpleProcessor) extractWithRegex(response string) *ProcessedResult {
	missionRe := regexp.MustCompile(`"mission_alignment":\s*(\d+\.?\d*)`)
	antiRe := regexp.MustCompile(`"anti_challenge":\s*(\d+\.?\d*)`)
	strategicRe := regexp.MustCompile(`"strategic_fit":\s*(\d+\.?\d*)`)
	finalRe := regexp.MustCompile(`"final_score":\s*(\d+\.?\d*)`)

	scores := make([]float64, 4)
	regexes := []*regexp.Regexp{missionRe, antiRe, strategicRe, finalRe}

	for i, re := range regexes {
		match := re.FindStringSubmatch(response)
		if len(match) < 2 {
			return nil
		}
		var err error
		scores[i], err = strconv.ParseFloat(match[1], 64)
		if err != nil {
			return nil
		}
	}

	return &ProcessedResult{
		Scores: ScoreBreakdown{
			MissionAlignment: scores[0],
			AntiChallenge:    scores[1],
			StrategicFit:     scores[2],
		},
		FinalScore:     scores[3],
		Recommendation: DetermineRecommendation(scores[3]),
		Explanations:   make(map[string]string),
		Provider:       "ollama_extracted",
	}
}

// validate checks score ranges
func (sp *SimpleProcessor) validate(result *ProcessedResult) bool {
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

// DetermineRecommendation maps score to recommendation
func DetermineRecommendation(score float64) string {
	if score >= 8.5 {
		return "PRIORITIZE NOW"
	} else if score >= 7.0 {
		return "GOOD ALIGNMENT"
	} else if score >= 5.0 {
		return "CONSIDER LATER"
	}
	return "AVOID FOR NOW"
}
