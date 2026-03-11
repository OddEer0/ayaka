package ayaka

import "context"

type Logger interface {
	Debug(ctx context.Context, message string, info map[string]any)
	Info(ctx context.Context, message string, info map[string]any)
	Warn(ctx context.Context, message string, info map[string]any)
	Error(ctx context.Context, message string, info map[string]any)
}

type NoopLogger struct{}

func (n NoopLogger) Debug(context.Context, string, map[string]any) {}

func (n NoopLogger) Info(context.Context, string, map[string]any) {}

func (n NoopLogger) Warn(context.Context, string, map[string]any) {}

func (n NoopLogger) Error(context.Context, string, map[string]any) {}
