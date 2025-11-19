package logging

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestLogger_NewLogger(t *testing.T) {
	cfg := Config{
		Level:      "info",
		Format:     "json",
		OutputPath: "stdout",
	}

	logger := NewLogger(cfg)
	if logger.GetLevel() != zerolog.InfoLevel {
		t.Errorf("Expected logger level to be InfoLevel, got %v", logger.GetLevel())
	}
}

func TestLogger_Levels(t *testing.T) {
	tests := []struct {
		name          string
		level         string
		expectedLevel zerolog.Level
	}{
		{"debug level", "debug", zerolog.DebugLevel},
		{"info level", "info", zerolog.InfoLevel},
		{"warn level", "warn", zerolog.WarnLevel},
		{"error level", "error", zerolog.ErrorLevel},
		{"default level", "invalid", zerolog.InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				Level:      tt.level,
				Format:     "json",
				OutputPath: "stdout",
			}

			logger := NewLogger(cfg)
			if logger.GetLevel() != tt.expectedLevel {
				t.Errorf("Expected level %v, got %v", tt.expectedLevel, logger.GetLevel())
			}
		})
	}
}

func TestLogger_StructuredFields(t *testing.T) {
	var buf bytes.Buffer

	logger := zerolog.New(&buf).With().Timestamp().Caller().Logger()

	logger.Info().
		Str("user_id", "123").
		Int("request_id", 456).
		Msg("test message")

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	if err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	if logEntry["user_id"] != "123" {
		t.Errorf("Expected user_id to be '123', got %v", logEntry["user_id"])
	}

	if logEntry["message"] != "test message" {
		t.Errorf("Expected message to be 'test message', got %v", logEntry["message"])
	}
}

func TestLogger_FileOutput(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	cfg := Config{
		Level:      "info",
		Format:     "json",
		OutputPath: logFile,
		MaxSizeMB:  10,
		MaxBackups: 3,
		MaxAgeDays: 7,
	}

	logger := NewLogger(cfg)
	logger.Info().Msg("test file output")

	// Wait a moment for file write
	time.Sleep(100 * time.Millisecond)

	// Verify file exists and has content
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if len(data) == 0 {
		t.Error("Log file is empty")
	}

	var logEntry map[string]interface{}
	err = json.Unmarshal(data, &logEntry)
	if err != nil {
		t.Fatalf("Failed to parse JSON log from file: %v", err)
	}

	if logEntry["message"] != "test file output" {
		t.Errorf("Expected message 'test file output', got %v", logEntry["message"])
	}
}

func TestLogger_JsonFormat(t *testing.T) {
	var buf bytes.Buffer

	// We'll manually create a logger writing to buffer for testing
	logger := zerolog.New(&buf).With().Timestamp().Logger()
	logger.Info().Str("format", "json").Msg("json test")

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	if err != nil {
		t.Fatalf("Expected valid JSON, got error: %v", err)
	}

	if logEntry["level"] != "info" {
		t.Errorf("Expected level 'info', got %v", logEntry["level"])
	}
}

func TestLogger_ContextIntegration(t *testing.T) {
	var buf bytes.Buffer

	logger := zerolog.New(&buf).With().
		Timestamp().
		Str("app", "telos-matrix").
		Str("version", "1.0.0").
		Logger()

	logger.Info().Msg("context test")

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if logEntry["app"] != "telos-matrix" {
		t.Errorf("Expected app 'telos-matrix', got %v", logEntry["app"])
	}

	if logEntry["version"] != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %v", logEntry["version"])
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected zerolog.Level
	}{
		{"debug", zerolog.DebugLevel},
		{"info", zerolog.InfoLevel},
		{"warn", zerolog.WarnLevel},
		{"error", zerolog.ErrorLevel},
		{"unknown", zerolog.InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseLogLevel(tt.input)
			if result != tt.expected {
				t.Errorf("parseLogLevel(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
