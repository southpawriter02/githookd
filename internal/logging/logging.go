package logging

import (
	"fmt"
	"io"
	"log/slog"
	"strings"

	"githookd/internal/config"
)

// Setup initializes the global slog logger with the given level and writer.
// All log output goes to the provided writer (typically os.Stderr).
func Setup(level slog.Level, w io.Writer) {
	handler := slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: level,
	})
	slog.SetDefault(slog.New(handler))
}

// ParseLevel converts a string log level to slog.Level.
// An empty string defaults to slog.LevelWarn.
func ParseLevel(s string) (slog.Level, error) {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelWarn, fmt.Errorf("invalid log level %q: valid levels are debug, info, warn, error", s)
	}
}

// ConfigLevelToSlog maps a config.LogLevel enum to a slog.Level.
func ConfigLevelToSlog(l config.LogLevel) slog.Level {
	switch l {
	case config.LogDebug:
		return slog.LevelDebug
	case config.LogInfo:
		return slog.LevelInfo
	case config.LogWarn:
		return slog.LevelWarn
	case config.LogError:
		return slog.LevelError
	default:
		return slog.LevelWarn
	}
}
