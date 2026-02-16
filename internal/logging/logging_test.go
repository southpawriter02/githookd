package logging

import (
	"bytes"
	"log/slog"
	"testing"

	"githookd/internal/config"
)

func TestParseLevel_ValidLevels(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
		{"DEBUG", slog.LevelDebug},
		{"Info", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseLevel(tt.input)
			if err != nil {
				t.Fatalf("ParseLevel(%q) error = %v", tt.input, err)
			}
			if got != tt.expected {
				t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseLevel_EmptyDefaultsToWarn(t *testing.T) {
	got, err := ParseLevel("")
	if err != nil {
		t.Fatalf("ParseLevel('') error = %v", err)
	}
	if got != slog.LevelWarn {
		t.Errorf("ParseLevel('') = %v, want LevelWarn", got)
	}
}

func TestParseLevel_InvalidReturnsError(t *testing.T) {
	_, err := ParseLevel("verbose")
	if err == nil {
		t.Fatal("expected error for invalid level 'verbose'")
	}
}

func TestSetup_ConfiguresLogger(t *testing.T) {
	var buf bytes.Buffer
	Setup(slog.LevelDebug, &buf)

	slog.Debug("test debug message")
	if buf.Len() == 0 {
		t.Error("expected debug message to be written at debug level")
	}
}

func TestSetup_ErrorLevelSuppressesDebug(t *testing.T) {
	var buf bytes.Buffer
	Setup(slog.LevelError, &buf)

	slog.Debug("should not appear")
	slog.Info("should not appear")
	slog.Warn("should not appear")

	if buf.Len() != 0 {
		t.Error("expected no output at error level for debug/info/warn messages")
	}
}

func TestConfigLevelToSlog(t *testing.T) {
	tests := []struct {
		input    config.LogLevel
		expected slog.Level
	}{
		{config.LogDebug, slog.LevelDebug},
		{config.LogInfo, slog.LevelInfo},
		{config.LogWarn, slog.LevelWarn},
		{config.LogError, slog.LevelError},
	}

	for _, tt := range tests {
		got := ConfigLevelToSlog(tt.input)
		if got != tt.expected {
			t.Errorf("ConfigLevelToSlog(%v) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}
