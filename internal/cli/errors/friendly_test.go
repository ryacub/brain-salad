package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFriendlyError_Error(t *testing.T) {
	fe := &FriendlyError{
		Title:       "Test Error",
		Cause:       fmt.Errorf("underlying error"),
		Explanation: "This is a test error",
		Suggestions: []string{"Do this", "Or do that"},
	}

	errMsg := fe.Error()
	assert.Contains(t, errMsg, "Test Error")
	assert.Contains(t, errMsg, "underlying error")
	assert.Contains(t, errMsg, "This is a test error")
	assert.Contains(t, errMsg, "Do this")
	assert.Contains(t, errMsg, "Or do that")
}

func TestFriendlyError_Unwrap(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	fe := &FriendlyError{
		Title: "Wrapped Error",
		Cause: originalErr,
	}

	unwrapped := fe.Unwrap()
	assert.Equal(t, originalErr, unwrapped)
}

func TestWrapError_NilError(t *testing.T) {
	result := WrapError(nil, "test context")
	assert.Nil(t, result)
}

func TestWrapError_AlreadyFriendlyError(t *testing.T) {
	originalFE := &FriendlyError{
		Title: "Already Friendly",
	}

	result := WrapError(originalFE, "new context")
	assert.Equal(t, originalFE, result)
}

func TestWrapError_MissingTelosFile(t *testing.T) {
	tests := []struct {
		name   string
		errMsg string
	}{
		{
			name:   "no such file",
			errMsg: "open telos.md: no such file",
		},
		{
			name:   "not found",
			errMsg: "telos.md not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fmt.Errorf("%s", tt.errMsg)
			wrapped := WrapError(err, "test")

			var fe *FriendlyError
			require.True(t, errors.As(wrapped, &fe))
			assert.Equal(t, "Telos configuration file not found", fe.Title)
			assert.Contains(t, fe.Explanation, "telos.md")
			assert.Greater(t, len(fe.Suggestions), 0)
			assert.Contains(t, fe.Suggestions[0], "tm init")
		})
	}
}

func TestWrapError_InvalidTelosFormat(t *testing.T) {
	err := fmt.Errorf("failed to parse telos.md: invalid format")
	wrapped := WrapError(err, "test")

	var fe *FriendlyError
	require.True(t, errors.As(wrapped, &fe))
	assert.Equal(t, "Invalid telos.md file format", fe.Title)
	assert.Contains(t, fe.Explanation, "format")
	assert.Greater(t, len(fe.Suggestions), 0)
}

func TestWrapError_MissingAPIKey(t *testing.T) {
	tests := []struct {
		name             string
		errMsg           string
		expectedProvider string
		expectedInSuggestion string
	}{
		{
			name:             "openai",
			errMsg:           "OpenAI API not available",
			expectedProvider: "OpenAI",
			expectedInSuggestion: "OPENAI_API_KEY",
		},
		{
			name:             "claude",
			errMsg:           "Claude API not available",
			expectedProvider: "Claude",
			expectedInSuggestion: "CLAUDE_API_KEY",
		},
		{
			name:             "ollama",
			errMsg:           "Ollama API not available",
			expectedProvider: "Ollama",
			expectedInSuggestion: "ollama",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fmt.Errorf("%s", tt.errMsg)
			wrapped := WrapError(err, "test")

			var fe *FriendlyError
			require.True(t, errors.As(wrapped, &fe))
			assert.Contains(t, fe.Title, tt.expectedProvider)
			assert.Contains(t, fe.Explanation, "API key")
			assert.Greater(t, len(fe.Suggestions), 0)

			// Check that at least one suggestion contains the expected content
			found := false
			for _, suggestion := range fe.Suggestions {
				if strings.Contains(suggestion, tt.expectedInSuggestion) {
					found = true
					break
				}
			}
			assert.True(t, found, "Expected to find %s in suggestions", tt.expectedInSuggestion)
		})
	}
}

func TestWrapError_DatabaseErrors(t *testing.T) {
	tests := []struct {
		name          string
		errMsg        string
		expectedTitle string
	}{
		{
			name:          "permission denied",
			errMsg:        "database: permission denied",
			expectedTitle: "Database permission denied",
		},
		{
			name:          "locked database",
			errMsg:        "database is locked",
			expectedTitle: "Database is locked",
		},
		{
			name:          "corrupted database",
			errMsg:        "database disk image is corrupt",
			expectedTitle: "Database corruption detected",
		},
		{
			name:          "database not found",
			errMsg:        "database file: no such file",
			expectedTitle: "Database not found",
		},
		{
			name:          "generic database error",
			errMsg:        "sqlite error occurred",
			expectedTitle: "Database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fmt.Errorf("%s", tt.errMsg)
			wrapped := WrapError(err, "test")

			var fe *FriendlyError
			require.True(t, errors.As(wrapped, &fe))
			assert.Equal(t, tt.expectedTitle, fe.Title)
			assert.Greater(t, len(fe.Suggestions), 0)
		})
	}
}

func TestWrapError_EmptyContent(t *testing.T) {
	err := fmt.Errorf("content is required")
	wrapped := WrapError(err, "test")

	var fe *FriendlyError
	require.True(t, errors.As(wrapped, &fe))
	assert.Equal(t, "Missing required input", fe.Title)
	assert.Contains(t, fe.Explanation, "idea")
	assert.Greater(t, len(fe.Suggestions), 0)
}

func TestWrapError_ProviderNotAvailable(t *testing.T) {
	err := fmt.Errorf("provider not available")
	wrapped := WrapError(err, "test")

	var fe *FriendlyError
	require.True(t, errors.As(wrapped, &fe))
	assert.Equal(t, "LLM provider not available", fe.Title)
	assert.Greater(t, len(fe.Suggestions), 0)
}

func TestWrapError_ConnectionError(t *testing.T) {
	tests := []string{
		"connection timeout",
		"connection refused",
		"host unreachable",
	}

	for _, errMsg := range tests {
		t.Run(errMsg, func(t *testing.T) {
			err := fmt.Errorf("%s", errMsg)
			wrapped := WrapError(err, "test")

			var fe *FriendlyError
			require.True(t, errors.As(wrapped, &fe))
			assert.Equal(t, "Connection error", fe.Title)
			assert.Greater(t, len(fe.Suggestions), 0)
		})
	}
}

func TestWrapError_GenericError(t *testing.T) {
	err := fmt.Errorf("some random error")
	wrapped := WrapError(err, "Custom Context")

	var fe *FriendlyError
	require.True(t, errors.As(wrapped, &fe))
	assert.Equal(t, "Custom Context", fe.Title)
	assert.Equal(t, err, fe.Cause)
	assert.Greater(t, len(fe.Suggestions), 0)
	assert.Contains(t, fe.Suggestions[0], "Check the error message")
}

func TestExtractProviderName(t *testing.T) {
	tests := []struct {
		errStr   string
		expected string
	}{
		{"OpenAI API failed", "OpenAI"},
		{"OPENAI connection error", "OpenAI"},
		{"Claude error occurred", "Claude"},
		{"claude timeout", "Claude"},
		{"Ollama not responding", "Ollama"},
		{"ollama service down", "Ollama"},
		{"unknown provider error", "LLM"},
	}

	for _, tt := range tests {
		t.Run(tt.errStr, func(t *testing.T) {
			result := extractProviderName(tt.errStr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetSuggestionsForProvider(t *testing.T) {
	tests := []struct {
		provider              string
		shouldContain         string
	}{
		{"OpenAI", "OPENAI_API_KEY"},
		{"Claude", "CLAUDE_API_KEY"},
		{"Ollama", "ollama serve"},
		{"Unknown", "tm llm"},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			suggestions := getSuggestionsForProvider(tt.provider)
			assert.Greater(t, len(suggestions), 0)

			found := false
			for _, suggestion := range suggestions {
				if strings.Contains(suggestion, tt.shouldContain) {
					found = true
					break
				}
			}
			assert.True(t, found, "Expected to find %s in suggestions for %s", tt.shouldContain, tt.provider)
		})
	}
}

func TestHandleDatabaseError(t *testing.T) {
	tests := []struct {
		name          string
		errStr        string
		expectedTitle string
	}{
		{
			name:          "permission denied",
			errStr:        "permission denied",
			expectedTitle: "Database permission denied",
		},
		{
			name:          "access denied",
			errStr:        "access is denied",
			expectedTitle: "Database permission denied",
		},
		{
			name:          "locked",
			errStr:        "database is locked",
			expectedTitle: "Database is locked",
		},
		{
			name:          "busy",
			errStr:        "database is busy",
			expectedTitle: "Database is locked",
		},
		{
			name:          "corrupt",
			errStr:        "database corrupt",
			expectedTitle: "Database corruption detected",
		},
		{
			name:          "malformed",
			errStr:        "file is malformed",
			expectedTitle: "Database corruption detected",
		},
		{
			name:          "not a database",
			errStr:        "file is not a database",
			expectedTitle: "Database corruption detected",
		},
		{
			name:          "not found",
			errStr:        "no such file",
			expectedTitle: "Database not found",
		},
		{
			name:          "generic",
			errStr:        "some database error",
			expectedTitle: "Database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fmt.Errorf("%s", tt.errStr)
			result := handleDatabaseError(err, tt.errStr)

			var fe *FriendlyError
			require.True(t, errors.As(result, &fe))
			assert.Equal(t, tt.expectedTitle, fe.Title)
			assert.Greater(t, len(fe.Suggestions), 0)
		})
	}
}
