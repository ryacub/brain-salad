package telos

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/rayyacub/telos-idea-matrix/internal/models"
)

// Parser parses telos.md files into Telos structs.
type Parser struct {
	goalRegex     *regexp.Regexp
	strategyRegex *regexp.Regexp
	deadlineRegex *regexp.Regexp
	patternRegex  *regexp.Regexp
}

// NewParser creates a new Telos parser with compiled regex patterns.
func NewParser() *Parser {
	return &Parser{
		// Matches: - G1: Description (Deadline: 2025-12-31)
		goalRegex: regexp.MustCompile(`^-\s+(G\d+):\s+(.+?)(?:\s+\(Deadline:\s+(.+?)\))?$`),
		// Matches: - S1: Description
		strategyRegex: regexp.MustCompile(`^-\s+(S\d+):\s+(.+)$`),
		// Matches: YYYY-MM-DD format
		deadlineRegex: regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})`),
		// Matches: - Name: Description
		patternRegex: regexp.MustCompile(`^-\s+([^:]+):\s+(.+)$`),
	}
}

// ParseFile parses a telos.md file and returns a Telos struct.
func (p *Parser) ParseFile(path string) (*models.Telos, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	telos := &models.Telos{
		LoadedAt: time.Now().UTC(),
	}

	scanner := bufio.NewScanner(file)
	var currentSection string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Detect sections (## Goals, ## Strategies, etc.)
		if strings.HasPrefix(line, "## ") {
			currentSection = strings.TrimPrefix(line, "## ")
			continue
		}

		// Parse content based on current section
		switch currentSection {
		case "Goals":
			if goal := p.parseGoal(line); goal != nil {
				telos.Goals = append(telos.Goals, *goal)
			}
		case "Strategies":
			if strategy := p.parseStrategy(line); strategy != nil {
				telos.Strategies = append(telos.Strategies, *strategy)
			}
		case "Stack":
			p.parseStack(line, &telos.Stack)
		case "Failure Patterns":
			if pattern := p.parsePattern(line); pattern != nil {
				telos.FailurePatterns = append(telos.FailurePatterns, *pattern)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Validate the parsed telos
	if err := telos.Validate(); err != nil {
		return nil, fmt.Errorf("invalid telos: %w", err)
	}

	return telos, nil
}

// parseGoal parses a goal line and returns a Goal struct.
// Expected format: - G1: Description (Deadline: YYYY-MM-DD)
func (p *Parser) parseGoal(line string) *models.Goal {
	matches := p.goalRegex.FindStringSubmatch(line)
	if len(matches) < 3 {
		return nil
	}

	goal := &models.Goal{
		ID:          matches[1],
		Description: strings.TrimSpace(matches[2]),
		Priority:    0, // Priority can be inferred from order if needed
	}

	// Parse deadline if present (matches[3])
	if len(matches) > 3 && matches[3] != "" {
		if deadline, err := time.Parse("2006-01-02", matches[3]); err == nil {
			goal.Deadline = &deadline
		}
	}

	return goal
}

// parseStrategy parses a strategy line and returns a Strategy struct.
// Expected format: - S1: Description
func (p *Parser) parseStrategy(line string) *models.Strategy {
	matches := p.strategyRegex.FindStringSubmatch(line)
	if len(matches) < 3 {
		return nil
	}

	return &models.Strategy{
		ID:          matches[1],
		Description: strings.TrimSpace(matches[2]),
	}
}

// parseStack parses a stack line and updates the Stack struct.
// Expected format:
//   - Primary: Go, TypeScript, PostgreSQL
//   - Secondary: Docker, Kubernetes
func (p *Parser) parseStack(line string, stack *models.Stack) {
	if strings.HasPrefix(line, "- Primary:") {
		techs := strings.TrimPrefix(line, "- Primary:")
		stack.Primary = parseTechList(techs)
	} else if strings.HasPrefix(line, "- Secondary:") {
		techs := strings.TrimPrefix(line, "- Secondary:")
		stack.Secondary = parseTechList(techs)
	}
}

// parsePattern parses a failure pattern line and returns a Pattern struct.
// Expected format: - Name: Description
func (p *Parser) parsePattern(line string) *models.Pattern {
	matches := p.patternRegex.FindStringSubmatch(line)
	if len(matches) < 3 {
		return nil
	}

	return &models.Pattern{
		Name:        strings.TrimSpace(matches[1]),
		Description: strings.TrimSpace(matches[2]),
		Keywords:    extractKeywords(matches[2]),
	}
}

// parseTechList parses a comma-separated list of technologies.
func parseTechList(text string) []string {
	var result []string
	parts := strings.Split(text, ",")
	for _, part := range parts {
		tech := strings.TrimSpace(part)
		if tech != "" {
			result = append(result, tech)
		}
	}
	return result
}

// extractKeywords extracts meaningful keywords from a description.
// Filters out common stopwords and short words.
func extractKeywords(text string) []string {
	// Common stopwords to filter out
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true,
		"or": true, "but": true, "in": true, "on": true,
		"at": true, "to": true, "for": true, "of": true,
		"with": true, "from": true, "before": true, "by": true,
		"as": true, "is": true, "was": true, "are": true,
		"it": true, "that": true, "this": true, "be": true,
		"been": true, "being": true, "have": true, "has": true,
		"had": true, "do": true, "does": true, "did": true,
		"will": true, "would": true, "should": true, "could": true,
		"than": true, "them": true, "then": true, "into": true,
	}

	words := strings.Fields(strings.ToLower(text))
	var keywords []string

	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?;:")

		// Include words longer than 3 characters that aren't stopwords
		if len(word) > 3 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}

	return keywords
}
