// Package errors provides user-friendly error handling with actionable suggestions.
package errors

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// FriendlyError wraps an error with user-friendly context and suggestions
type FriendlyError struct {
	Title       string   // Short description
	Cause       error    // Original error
	Explanation string   // What went wrong
	Suggestions []string // How to fix it
}

func (e *FriendlyError) Error() string {
	var b strings.Builder

	// Title in red
	b.WriteString(color.RedString("Error: %s\n", e.Title))

	// Cause if available
	if e.Cause != nil {
		b.WriteString(color.New(color.Faint).Sprintf("  Cause: %v\n", e.Cause))
	}

	// Explanation
	if e.Explanation != "" {
		b.WriteString(fmt.Sprintf("\n%s\n", e.Explanation))
	}

	// Suggestions
	if len(e.Suggestions) > 0 {
		b.WriteString(color.CyanString("\nSuggestions:\n"))
		for i, suggestion := range e.Suggestions {
			b.WriteString(fmt.Sprintf("  %d. %s\n", i+1, suggestion))
		}
	}

	return b.String()
}

// Unwrap returns the underlying error
func (e *FriendlyError) Unwrap() error {
	return e.Cause
}

// WrapError converts common errors into friendly errors
func WrapError(err error, context string) error {
	if err == nil {
		return nil
	}

	// If already a FriendlyError, return as is
	var fe *FriendlyError
	if errors.As(err, &fe) {
		return err
	}

	// Check for common error patterns
	errStr := err.Error()

	// Missing telos file
	if strings.Contains(errStr, "telos.md") && (strings.Contains(errStr, "no such file") || strings.Contains(errStr, "not found")) {
		return &FriendlyError{
			Title:       "Telos configuration file not found",
			Cause:       err,
			Explanation: "The telos.md file contains your personal goals, strategies, and failure patterns.",
			Suggestions: []string{
				"Run 'tm init' to create a template telos.md file",
				"Or create ~/.telos/telos.md manually with your goals",
				"See examples at: docs/examples/telos-samples/",
			},
		}
	}

	// Failed to parse telos
	if strings.Contains(errStr, "failed to parse telos") || strings.Contains(errStr, "parse telos.md") {
		return &FriendlyError{
			Title:       "Invalid telos.md file format",
			Cause:       err,
			Explanation: "The telos.md file could not be parsed. Check the format.",
			Suggestions: []string{
				"Verify telos.md uses proper markdown format",
				"Check for required sections: ## Goals, ## Strategies",
				"Run 'tm doctor' to check configuration health",
				"See examples at: docs/examples/telos-samples/",
			},
		}
	}

	// Missing API key
	if strings.Contains(errStr, "not available") && strings.Contains(errStr, "API") {
		providerName := extractProviderName(errStr)
		return &FriendlyError{
			Title:       fmt.Sprintf("%s API key not configured", providerName),
			Cause:       err,
			Explanation: "AI-powered analysis requires an API key for the selected provider.",
			Suggestions: getSuggestionsForProvider(providerName),
		}
	}

	// Database errors
	if strings.Contains(errStr, "database") || strings.Contains(errStr, "sqlite") || strings.Contains(errStr, "sql") {
		return handleDatabaseError(err, errStr)
	}

	// Empty content
	if strings.Contains(errStr, "empty") || (strings.Contains(errStr, "required") && !strings.Contains(errStr, "database")) {
		return &FriendlyError{
			Title:       "Missing required input",
			Cause:       err,
			Explanation: "You need to provide an idea to analyze.",
			Suggestions: []string{
				"Provide idea text: tm dump \"Your idea here\"",
				"Use --interactive for step-by-step input",
				"Or use --from-clipboard to paste from clipboard",
			},
		}
	}

	// Provider not available
	if strings.Contains(errStr, "provider") && (strings.Contains(errStr, "not available") || strings.Contains(errStr, "unavailable")) {
		return &FriendlyError{
			Title:       "LLM provider not available",
			Cause:       err,
			Explanation: "The requested LLM provider is not configured or unavailable.",
			Suggestions: []string{
				"Check which providers are available: tm llm",
				"Configure API keys: export OPENAI_API_KEY=sk-...",
				"Or use rule-based scoring: tm dump \"idea\" (no AI needed)",
				"Run 'tm doctor' to check LLM provider status",
			},
		}
	}

	// Connection errors
	if strings.Contains(errStr, "connection") || strings.Contains(errStr, "timeout") || strings.Contains(errStr, "unreachable") {
		return &FriendlyError{
			Title:       "Connection error",
			Cause:       err,
			Explanation: "Could not connect to the service.",
			Suggestions: []string{
				"Check your internet connection",
				"Verify the service URL is correct",
				"Check if firewall is blocking the connection",
				"Try again in a few moments",
			},
		}
	}

	// Generic wrapper with context
	return &FriendlyError{
		Title: context,
		Cause: err,
		Suggestions: []string{
			"Check the error message above for details",
			"Run 'tm doctor' to diagnose system health",
			"See documentation: docs/",
		},
	}
}

func extractProviderName(errStr string) string {
	errLower := strings.ToLower(errStr)
	if strings.Contains(errLower, "openai") {
		return "OpenAI"
	}
	if strings.Contains(errLower, "claude") {
		return "Claude"
	}
	if strings.Contains(errLower, "ollama") {
		return "Ollama"
	}
	return "LLM"
}

func getSuggestionsForProvider(provider string) []string {
	switch provider {
	case "OpenAI":
		return []string{
			"Set your API key: export OPENAI_API_KEY=sk-...",
			"Get an API key at: https://platform.openai.com/api-keys",
			"Or use --provider ollama for local LLM (free)",
			"Or use rule-based scoring: tm dump \"idea\" (no AI needed)",
		}
	case "Claude":
		return []string{
			"Set your API key: export CLAUDE_API_KEY=sk-ant-...",
			"Get an API key at: https://console.anthropic.com/",
			"Or use --provider openai instead",
			"Or use rule-based scoring: tm dump \"idea\" (no AI needed)",
		}
	case "Ollama":
		return []string{
			"Install Ollama from: https://ollama.ai",
			"Start Ollama: ollama serve",
			"Or use a cloud provider: --provider openai",
		}
	default:
		return []string{
			"Check which providers are available: tm llm",
			"Or use rule-based scoring (no API key needed)",
		}
	}
}

func handleDatabaseError(err error, errStr string) error {
	// Permission denied
	if strings.Contains(errStr, "permission denied") || strings.Contains(errStr, "access is denied") {
		return &FriendlyError{
			Title:       "Database permission denied",
			Cause:       err,
			Explanation: "Cannot access database file (permission denied).",
			Suggestions: []string{
				"Check file permissions: ls -la ~/.telos/",
				"Ensure directory is writable: chmod 755 ~/.telos/",
				"Or specify different path: tm dump \"idea\" --db /tmp/ideas.db",
			},
		}
	}

	// Locked database
	if strings.Contains(errStr, "locked") || strings.Contains(errStr, "busy") {
		return &FriendlyError{
			Title:       "Database is locked",
			Cause:       err,
			Explanation: "Another process is using the database.",
			Suggestions: []string{
				"Close other tm commands or applications using the database",
				"Wait a moment and try again",
				"If problem persists, check for stale lock files",
			},
		}
	}

	// Corrupted database
	if strings.Contains(errStr, "corrupt") || strings.Contains(errStr, "malformed") || strings.Contains(errStr, "not a database") {
		return &FriendlyError{
			Title:       "Database corruption detected",
			Cause:       err,
			Explanation: "The database file appears to be corrupted.",
			Suggestions: []string{
				"Restore from backup if available",
				"Run 'tm bulk export > backup.json' if database is partially readable",
				"Last resort: Delete ~/.telos/ideas.db and run 'tm init' (loses data)",
				"Run 'tm doctor' to check database health",
			},
		}
	}

	// Database not found
	if strings.Contains(errStr, "no such file") || strings.Contains(errStr, "not found") {
		return &FriendlyError{
			Title:       "Database not found",
			Cause:       err,
			Explanation: "The database file does not exist.",
			Suggestions: []string{
				"Run 'tm init' to initialize the database",
				"Check the database path: tm --help",
				"Or specify a custom path: tm --db /path/to/db dump \"idea\"",
			},
		}
	}

	// Generic database error
	return &FriendlyError{
		Title:       "Database error",
		Cause:       err,
		Explanation: "An error occurred while accessing the database.",
		Suggestions: []string{
			"Run 'tm doctor' to check database health",
			"Check disk space: df -h",
			"Check database file is not corrupted",
		},
	}
}

// ShowWarning displays a warning message
func ShowWarning(title string, message string) {
	color.Yellow("Warning: %s", title)
	if message != "" {
		fmt.Printf("  %s\n", message)
	}
}

// ShowSuccess displays a success message
func ShowSuccess(message string) {
	color.Green("✓ %s", message)
}

// ShowInfo displays an info message
func ShowInfo(message string) {
	color.Cyan("ℹ %s", message)
}
