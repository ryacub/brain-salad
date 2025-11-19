package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/rayyacub/telos-idea-matrix/internal/models"
)

// PromptTemplate is the template for LLM analysis prompts.
const PromptTemplate = `You are an expert at evaluating ideas against personal goals and values.

TELOS (Personal Goals & Values):
{{.TelosContent}}

IDEA TO EVALUATE:
{{.IdeaContent}}

TASK:
Analyze this idea and provide a detailed scoring breakdown based on the telos above.

SCORING FRAMEWORK:

1. Mission Alignment (0-4.0 points total - 40%):
   - Domain Expertise (0-1.2): Does this leverage existing skills and domain knowledge?
   - AI Alignment (0-1.5): How central is AI to this idea?
   - Execution Support (0-0.8): Can this be delivered quickly?
   - Revenue Potential (0-0.5): Is there a clear path to revenue?

2. Anti-Challenge Patterns (0-3.5 points total - 35%):
   - Avoid Context-Switching (0-1.2): Does this use your current stack?
   - Rapid Prototyping (0-1.0): Can you build an MVP quickly?
   - Accountability (0-0.8): Is there external accountability?
   - Income Anxiety (0-0.5): How quickly can this generate revenue?

3. Strategic Fit (0-2.5 points total - 25%):
   - Stack Compatibility (0-1.0): Enables flow state with your stack?
   - Shipping Habit (0-0.8): Creates reusable systems/code?
   - Public Accountability (0-0.4): Can you validate quickly?
   - Revenue Testing (0-0.3): Is this scalable (SaaS vs consulting)?

RESPONSE FORMAT:
Respond with valid JSON in this exact format:
{
  "scores": {
    "mission_alignment": 2.5,
    "anti_challenge": 2.0,
    "strategic_fit": 1.5
  },
  "final_score": 6.0,
  "recommendation": "CONSIDER LATER",
  "explanations": {
    "mission_alignment": "explanation here",
    "anti_challenge": "explanation here",
    "strategic_fit": "explanation here"
  }
}

IMPORTANT:
- Provide ONLY the JSON response, no additional text
- Ensure all scores are within their valid ranges
- final_score should be the sum of the three category scores
- recommendation should be one of: "PRIORITIZE NOW", "GOOD ALIGNMENT", "CONSIDER LATER", "AVOID FOR NOW"
`

// PromptData contains the data needed to build a prompt.
type PromptData struct {
	TelosContent string
	IdeaContent  string
}

// BuildAnalysisPrompt builds a prompt for LLM analysis.
// It takes the idea content and telos, and returns a formatted prompt.
func BuildAnalysisPrompt(ideaContent string, telos *models.Telos) (string, error) {
	if ideaContent == "" {
		return "", fmt.Errorf("idea content is required")
	}
	if telos == nil {
		return "", fmt.Errorf("telos is required")
	}

	// Convert telos to human-readable format
	telosContent := formatTelos(telos)

	// Create template data
	data := PromptData{
		TelosContent: telosContent,
		IdeaContent:  ideaContent,
	}

	// Parse and execute template
	tmpl, err := template.New("prompt").Parse(PromptTemplate)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}

// formatTelos converts a Telos struct to a human-readable string.
func formatTelos(telos *models.Telos) string {
	var buf bytes.Buffer

	// Goals
	if len(telos.Goals) > 0 {
		buf.WriteString("## Goals:\n")
		for _, goal := range telos.Goals {
			buf.WriteString(fmt.Sprintf("- %s: %s\n", goal.ID, goal.Description))
			if goal.Deadline != nil {
				buf.WriteString(fmt.Sprintf("  Deadline: %s\n", goal.Deadline.Format("2006-01-02")))
			}
		}
		buf.WriteString("\n")
	}

	// Strategies
	if len(telos.Strategies) > 0 {
		buf.WriteString("## Strategies:\n")
		for _, strategy := range telos.Strategies {
			buf.WriteString(fmt.Sprintf("- %s: %s\n", strategy.ID, strategy.Description))
		}
		buf.WriteString("\n")
	}

	// Stack
	if len(telos.Stack.Primary) > 0 || len(telos.Stack.Secondary) > 0 {
		buf.WriteString("## Tech Stack:\n")
		if len(telos.Stack.Primary) > 0 {
			buf.WriteString(fmt.Sprintf("- Primary: %s\n", strings.Join(telos.Stack.Primary, ", ")))
		}
		if len(telos.Stack.Secondary) > 0 {
			buf.WriteString(fmt.Sprintf("- Secondary: %s\n", strings.Join(telos.Stack.Secondary, ", ")))
		}
		buf.WriteString("\n")
	}

	// Failure Patterns
	if len(telos.FailurePatterns) > 0 {
		buf.WriteString("## Failure Patterns to Avoid:\n")
		for _, pattern := range telos.FailurePatterns {
			buf.WriteString(fmt.Sprintf("- %s: %s\n", pattern.Name, pattern.Description))
		}
		buf.WriteString("\n")
	}

	return buf.String()
}

// ParseLLMResponse parses the JSON response from an LLM.
func ParseLLMResponse(response string) (*LLMResponse, error) {
	// Extract JSON from response (handle cases where LLM adds extra text)
	jsonStr := extractJSON(response)
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON found in response")
	}

	var llmResp LLMResponse
	if err := json.Unmarshal([]byte(jsonStr), &llmResp); err != nil {
		return nil, fmt.Errorf("unmarshal JSON: %w", err)
	}

	// Validate response
	if err := llmResp.Validate(); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}

	return &llmResp, nil
}

// LLMResponse represents the structured response from an LLM.
type LLMResponse struct {
	Scores struct {
		MissionAlignment float64 `json:"mission_alignment"`
		AntiChallenge    float64 `json:"anti_challenge"`
		StrategicFit     float64 `json:"strategic_fit"`
	} `json:"scores"`
	FinalScore     float64           `json:"final_score"`
	Recommendation string            `json:"recommendation"`
	Explanations   map[string]string `json:"explanations"`
}

// Validate validates the LLM response.
func (r *LLMResponse) Validate() error {
	if r.Scores.MissionAlignment < 0 || r.Scores.MissionAlignment > 4.0 {
		return fmt.Errorf("mission_alignment must be between 0-4.0, got %f", r.Scores.MissionAlignment)
	}
	if r.Scores.AntiChallenge < 0 || r.Scores.AntiChallenge > 3.5 {
		return fmt.Errorf("anti_challenge must be between 0-3.5, got %f", r.Scores.AntiChallenge)
	}
	if r.Scores.StrategicFit < 0 || r.Scores.StrategicFit > 2.5 {
		return fmt.Errorf("strategic_fit must be between 0-2.5, got %f", r.Scores.StrategicFit)
	}
	if r.FinalScore < 0 || r.FinalScore > 10.0 {
		return fmt.Errorf("final_score must be between 0-10, got %f", r.FinalScore)
	}

	validRecommendations := map[string]bool{
		"PRIORITIZE NOW":  true,
		"GOOD ALIGNMENT":  true,
		"CONSIDER LATER":  true,
		"AVOID FOR NOW":   true,
	}
	if !validRecommendations[r.Recommendation] {
		return fmt.Errorf("invalid recommendation: %s", r.Recommendation)
	}

	return nil
}

// extractJSON extracts JSON from a string that might contain additional text.
func extractJSON(s string) string {
	// First, try to find JSON within code blocks
	if start := strings.Index(s, "```json"); start != -1 {
		start += 7 // Skip "```json"
		if end := strings.Index(s[start:], "```"); end != -1 {
			return strings.TrimSpace(s[start : start+end])
		}
	}

	// Try to find JSON object directly
	if start := strings.Index(s, "{"); start != -1 {
		braceCount := 0
		for i := start; i < len(s); i++ {
			switch s[i] {
			case '{':
				braceCount++
			case '}':
				braceCount--
				if braceCount == 0 {
					return s[start : i+1]
				}
			}
		}
	}

	// Return the original string if no JSON found
	return strings.TrimSpace(s)
}
