package logger

import (
	"context"
	"log/slog"
	"os"

	"taskflow/utils"
)

type contextKey string

const loggerKey contextKey = "logger_key"

// FromContext retrieves the request logger, or slog.Default() when none is present.
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

// NewContext stores a logger on the context.
func NewContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// Init configures the process-wide default logger.
func Init(config *utils.Config) {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(config.LogLevel)); err != nil {
		lvl = slog.LevelInfo
	}

	l := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: lvl,
	})).With("service", config.Name)

	slog.SetDefault(l)
}
