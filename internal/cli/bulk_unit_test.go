package cli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{
			name:     "30 days",
			input:    "30d",
			expected: 30 * 24 * time.Hour,
			wantErr:  false,
		},
		{
			name:     "6 hours",
			input:    "6h",
			expected: 6 * time.Hour,
			wantErr:  false,
		},
		{
			name:     "45 minutes",
			input:    "45m",
			expected: 45 * time.Minute,
			wantErr:  false,
		},
		{
			name:     "90 seconds",
			input:    "90s",
			expected: 90 * time.Second,
			wantErr:  false,
		},
		{
			name:     "invalid format - too short",
			input:    "d",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "invalid format - no number",
			input:    "xd",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "zero days",
			input:    "0d",
			expected: 0,
			wantErr:  false,
		},
		{
			name:     "large value",
			input:    "365d",
			expected: 365 * 24 * time.Hour,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDuration(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestCreateLLMManager(t *testing.T) {
	manager := createLLMManager()
	assert.NotNil(t, manager, "LLM manager should not be nil")

	// Verify that we can get providers
	providers := manager.GetProviders()
	assert.NotNil(t, providers, "Providers list should not be nil")
	assert.GreaterOrEqual(t, len(providers), 1, "Should have at least one provider (rule_based)")
}
