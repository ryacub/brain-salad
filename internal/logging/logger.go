package logging

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Config holds the logging configuration
type Config struct {
	Level      string
	Format     string // "json" or "console"
	OutputPath string // file path or "stdout"
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
}

// NewLogger creates and configures a new zerolog logger
func NewLogger(cfg Config) zerolog.Logger {
	// Set log level
	level := parseLogLevel(cfg.Level)
	zerolog.SetGlobalLevel(level)

	// Configure output
	var output io.Writer
	if cfg.OutputPath == "stdout" || cfg.OutputPath == "" {
		output = os.Stdout
	} else {
		// Ensure directory exists
		dir := filepath.Dir(cfg.OutputPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			// Fallback to stdout if directory creation fails
			output = os.Stdout
		} else {
			// File output with rotation
			output = &lumberjack.Logger{
				Filename:   cfg.OutputPath,
				MaxSize:    cfg.MaxSizeMB,
				MaxBackups: cfg.MaxBackups,
				MaxAge:     cfg.MaxAgeDays,
				Compress:   true,
			}
		}
	}

	// Format
	if cfg.Format == "console" {
		output = zerolog.ConsoleWriter{Out: output, TimeFormat: time.RFC3339}
	}

	logger := zerolog.New(output).With().Timestamp().Caller().Logger().Level(level)

	// Set global logger
	log.Logger = logger

	return logger
}

// parseLogLevel converts a string log level to zerolog.Level
func parseLogLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}
