package ecosystem

import (
	"context"
	"log/slog"
	"os"
)

type SlogLogger struct {
	logger *slog.Logger
}

func (s *SlogLogger) Debug(ctx context.Context, message string, info map[string]any) {
	attrs := make([]any, 0, len(info))
	for k, v := range info {
		attrs = append(attrs, slog.Any(k, v))
	}
	s.logger.DebugContext(ctx, message, attrs...)
}

func (s *SlogLogger) Info(ctx context.Context, message string, info map[string]any) {
	attrs := make([]any, 0, len(info))
	for k, v := range info {
		attrs = append(attrs, slog.Any(k, v))
	}
	s.logger.InfoContext(ctx, message, attrs...)
}

func (s *SlogLogger) Warn(ctx context.Context, message string, info map[string]any) {
	attrs := make([]any, 0, len(info))
	for k, v := range info {
		attrs = append(attrs, slog.Any(k, v))
	}
	s.logger.WarnContext(ctx, message, attrs...)
}

func (s *SlogLogger) Error(ctx context.Context, message string, info map[string]any) {
	attrs := make([]any, 0, len(info))
	for k, v := range info {
		attrs = append(attrs, slog.Any(k, v))
	}
	s.logger.ErrorContext(ctx, message, attrs...)
}

func NewAppLoggerWithSlog(logger *slog.Logger) *SlogLogger {
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}
	return &SlogLogger{logger: logger}
}
