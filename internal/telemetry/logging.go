package telemetry

import (
	"log/slog"
	"os"
	"strings"

	"github.com/agentfence/agentfence/internal/config"
)

// NewLogger constructs the process logger from config.
func NewLogger(cfg config.LogConfig) *slog.Logger {
	options := &slog.HandlerOptions{
		Level: parseLevel(cfg.Level),
	}

	if strings.EqualFold(cfg.Format, "text") {
		return slog.New(slog.NewTextHandler(os.Stdout, options))
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, options))
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
