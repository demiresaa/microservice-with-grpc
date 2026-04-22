package logger

import (
	"log/slog"
	"os"
)

func New(serviceName string) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := slog.New(handler).With(
		"service", serviceName,
	)

	slog.SetDefault(logger)

	return logger
}
