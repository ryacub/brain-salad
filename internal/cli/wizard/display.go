package wizard

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/term"
)

func init() {
	// Respect NO_COLOR environment variable for accessibility
	// See: https://no-color.org/
	if os.Getenv("NO_COLOR") != "" {
		color.NoColor = true
	}
}

var (
	headerColor   = color.New(color.FgCyan, color.Bold)
	promptColor   = color.New(color.FgYellow)
	optionColor   = color.New(color.FgWhite)
	successColor  = color.New(color.FgGreen)
	subtextColor  = color.New(color.FgHiBlack)
	errorColor    = color.New(color.FgRed)
	dividerColor  = color.New(color.FgHiBlack)
	progressColor = color.New(color.FgHiBlack)
)

// Runner executes the interactive wizard flow.
type Runner struct {
	reader *bufio.Reader
	isTTY  bool
}

// NewRunner creates a new wizard runner.
func NewRunner() *Runner {
	return &Runner{
		reader: bufio.NewReader(os.Stdin),
		isTTY:  term.IsTerminal(int(os.Stdin.Fd())),
	}
}

// IsTTY returns true if running in an interactive terminal.
func (r *Runner) IsTTY() bool {
	return r.isTTY
}

// Run executes the complete wizard flow and returns the collected answers.
func (r *Runner) Run() (*WizardAnswers, error) {
	// Check if running in interactive terminal
	if !r.isTTY {
		return nil, fmt.Errorf("wizard requires an interactive terminal\n\nTip: Run this command directly in your terminal, not through a pipe or script")
	}

	answers := &WizardAnswers{
		Answers: []Answer{},
		Goals:   []string{},
		Avoid:   []string{},
	}

	r.printHeader()

	// Ask discovery questions
	questions := Questions()
	totalSteps := len(questions) + 2 // +2 for goals and avoid prompts

	for i, q := range questions {
		r.printDivider()
		r.printProgress(i+1, totalSteps)
		answer, err := r.askQuestion(q, i+1, len(questions))
		if err != nil {
			return nil, err
		}
		answers.Answers = append(answers.Answers, *answer)
	}

	// Collect goals
	r.printDivider()
	r.printProgress(len(questions)+1, totalSteps)
	goals, err := r.collectGoals()
	if err != nil {
		return nil, err
	}
	answers.Goals = goals

	// Collect avoidances
	r.printDivider()
	r.printProgress(len(questions)+2, totalSteps)
	avoid, err := r.collectAvoid()
	if err != nil {
		return nil, err
	}
	answers.Avoid = avoid

	return answers, nil
}

// printHeader displays the wizard introduction.
func (r *Runner) printHeader() {
	fmt.Println()
	_, _ = headerColor.Println("Let's figure out how you evaluate ideas.")
	_, _ = subtextColor.Println("I'll show you some tradeoffs — pick what feels right.")
	fmt.Println()
}

// printDivider displays a visual separator.
func (r *Runner) printDivider() {
	_, _ = dividerColor.Println("─────────────────────────────────────────────────────")
	fmt.Println()
}

// printProgress displays the current step progress.
func (r *Runner) printProgress(current, total int) {
	_, _ = progressColor.Printf("Step %d of %d\n\n", current, total)
}

// askQuestion displays a question and collects the answer.
func (r *Runner) askQuestion(q Question, num, total int) (*Answer, error) {
	// Display question
	_, _ = promptColor.Printf("Q%d: %s\n", num, q.Text)
	if q.Subtext != "" {
		_, _ = subtextColor.Printf("    %s\n", q.Subtext)
	}
	fmt.Println()

	// Display options
	for _, opt := range q.Options {
		_, _ = optionColor.Printf("    [%s] %s\n", opt.Key, opt.Text)
		if opt.Description != "" {
			_, _ = subtextColor.Printf("        %s\n", opt.Description)
		}
	}
	fmt.Println()

	// Collect answer
	validKeys := make(map[string]bool)
	keyList := []string{}
	for _, opt := range q.Options {
		validKeys[strings.ToUpper(opt.Key)] = true
		keyList = append(keyList, opt.Key)
	}

	for {
		fmt.Print("> ")
		input, err := r.reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read input: %w", err)
		}

		input = strings.TrimSpace(strings.ToUpper(input))

		if validKeys[input] {
			return &Answer{
				QuestionID: q.ID,
				OptionKey:  input,
			}, nil
		}

		// Friendly error with valid options
		_, _ = errorColor.Printf("Hmm, '%s' isn't one of the options.\n", strings.ToLower(input))
		_, _ = subtextColor.Printf("Try typing %s\n\n", strings.Join(keyList, ", "))
	}
}

// collectGoals prompts user for their goals.
func (r *Runner) collectGoals() ([]string, error) {
	prompt := GetGoalPrompt()
	examples := GoalExamples()

	_, _ = promptColor.Printf("Q6: %s\n", prompt.Text)
	_, _ = subtextColor.Printf("    %s\n", prompt.Subtext)
	fmt.Println()

	// Show examples to guide the user
	_, _ = subtextColor.Println("    Examples:")
	for _, ex := range examples[:3] { // Show first 3 examples
		_, _ = subtextColor.Printf("      • %s\n", ex)
	}
	fmt.Println()

	const (
		maxRetries  = 3
		maxInputLen = 500
	)

	goals := []string{}
	retryCount := 0

	for i := 1; i <= prompt.MaxGoals; i++ {
		if i == 1 {
			_, _ = subtextColor.Printf("    Goal %d: ", i)
		} else if i <= prompt.MinGoals {
			_, _ = subtextColor.Printf("    Goal %d: ", i)
		} else {
			_, _ = subtextColor.Printf("    Goal %d (optional, Enter to skip): ", i)
		}
		fmt.Print("> ")

		input, err := r.reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read input: %w", err)
		}

		input = strings.TrimSpace(input)

		// Validate input length
		if len(input) > maxInputLen {
			_, _ = errorColor.Printf("Input too long (max %d characters). Please shorten.\n", maxInputLen)
			i-- // Retry this goal
			continue
		}

		if input == "" {
			if i <= prompt.MinGoals {
				retryCount++
				if retryCount >= maxRetries {
					return nil, fmt.Errorf("no goals provided after %d attempts", maxRetries)
				}
				_, _ = errorColor.Printf("Please enter at least one goal. (%d/%d attempts)\n", retryCount, maxRetries)
				i-- // Retry this goal
				continue
			}
			break // Optional goal, user pressed enter
		}

		retryCount = 0 // Reset on successful input
		goals = append(goals, input)
	}

	fmt.Println()
	return goals, nil
}

// collectAvoid prompts user for things to avoid.
func (r *Runner) collectAvoid() ([]string, error) {
	const maxInputLen = 500

	prompt := GetAvoidPrompt()
	examples := AvoidExamples()

	_, _ = promptColor.Printf("Q7: %s\n", prompt.Text)
	_, _ = subtextColor.Printf("    %s\n", prompt.Subtext)
	fmt.Println()

	// Show examples
	_, _ = subtextColor.Println("    Examples:")
	_, _ = subtextColor.Printf("      %s\n", strings.Join(examples[:3], ", "))
	fmt.Println()

	fmt.Print("> ")
	input, err := r.reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)

	// Validate input length
	if len(input) > maxInputLen {
		_, _ = errorColor.Printf("Input too long (max %d characters). Truncating.\n", maxInputLen)
		input = input[:maxInputLen]
	}

	if input == "" {
		_, _ = subtextColor.Println("    (skipped)")
		return []string{}, nil
	}

	// Split by comma and clean up
	parts := strings.Split(input, ",")
	avoid := []string{}
	for _, part := range parts {
		cleaned := strings.TrimSpace(part)
		if cleaned != "" {
			avoid = append(avoid, cleaned)
		}
	}

	fmt.Println()
	return avoid, nil
}

// PrintSummary displays what the wizard learned about the user.
func (r *Runner) PrintSummary(summaryLines []string) {
	fmt.Println()
	_, _ = headerColor.Println("Here's what I learned:")
	fmt.Println()

	for _, line := range summaryLines {
		_, _ = successColor.Printf("  • %s\n", line)
	}
	fmt.Println()
}

// PrintSuccess displays the completion message.
func (r *Runner) PrintSuccess(profilePath string) {
	_, _ = successColor.Printf("✓ Profile saved to %s\n", profilePath)
	fmt.Println()
	_, _ = subtextColor.Println("Run `brain-salad profile` to view your profile")
	_, _ = subtextColor.Println("Run `brain-salad profile reset` to start over")
	fmt.Println()
}

// PrintError displays an error message.
func (r *Runner) PrintError(message string) {
	_, _ = errorColor.Printf("✗ %s\n", message)
}

// Confirm asks for yes/no confirmation.
func (r *Runner) Confirm(prompt string) bool {
	_, _ = promptColor.Printf("%s [y/N]: ", prompt)

	input, err := r.reader.ReadString('\n')
	if err != nil {
		return false
	}

	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

// PrintProfilePreview displays a visual preview of the scoring weights.
func (r *Runner) PrintProfilePreview(priorities map[string]float64, goals, avoid []string) {
	fmt.Println()
	_, _ = headerColor.Println("Your Profile Preview")
	_, _ = dividerColor.Println("─────────────────────────────────────────────────────")
	fmt.Println()

	// Show goals
	if len(goals) > 0 {
		_, _ = promptColor.Println("Goals:")
		for _, goal := range goals {
			fmt.Printf("  • %s\n", goal)
		}
		fmt.Println()
	}

	// Show avoid
	if len(avoid) > 0 {
		_, _ = promptColor.Println("Avoiding:")
		for _, a := range avoid {
			fmt.Printf("  • %s\n", a)
		}
		fmt.Println()
	}

	// Show scoring weights as visual bars
	_, _ = promptColor.Println("Scoring Weights:")

	// Dimension labels (human-friendly)
	labels := map[string]string{
		"completion_likelihood": "Completion  ",
		"skill_fit":             "Skill Fit   ",
		"time_to_done":          "Timeline    ",
		"reward_alignment":      "Reward      ",
		"sustainability":        "Motivation  ",
		"avoidance_fit":         "Avoidance   ",
	}

	// Order for display
	order := []string{
		"completion_likelihood",
		"skill_fit",
		"time_to_done",
		"reward_alignment",
		"sustainability",
		"avoidance_fit",
	}

	for _, dim := range order {
		weight := priorities[dim]
		label := labels[dim]

		// Create visual bar (10 chars max, proportional to weight)
		barLen := int(weight * 40) // Scale to 40 chars max
		bar := strings.Repeat("█", barLen) + strings.Repeat("░", 8-barLen/5)

		// Truncate bar to reasonable length
		if len(bar) > 8 {
			bar = bar[:8]
		}

		fmt.Printf("  %s %s %.0f%%\n", label, bar, weight*100)
	}
	fmt.Println()
}

// ConfirmSave asks for confirmation before saving the profile.
func (r *Runner) ConfirmSave() bool {
	return r.Confirm("Save this profile?")
}
