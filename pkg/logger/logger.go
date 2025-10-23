package logger

import (
	"io"
	"log/slog"
	"os"
)

// Logger defines the interface for structured logging
type Logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
	Warn(msg string, args ...any)
	Debug(msg string, args ...any)
	With(args ...any) Logger
}

// SlogLogger wraps slog.Logger to implement our Logger interface
type SlogLogger struct {
	logger *slog.Logger
}

// NewLogger creates a new structured logger instance
func NewLogger(level slog.Level, output io.Writer) Logger {
	if output == nil {
		output = os.Stdout
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewJSONHandler(output, opts)
	logger := slog.New(handler)

	return &SlogLogger{
		logger: logger,
	}
}

func (l *SlogLogger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *SlogLogger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

func (l *SlogLogger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *SlogLogger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

// With returns a new logger with the given key-value pairs added to context
func (l *SlogLogger) With(args ...any) Logger {
	return &SlogLogger{
		logger: l.logger.With(args...),
	}
}