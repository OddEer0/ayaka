package ayaka

import "context"

type Logger interface {
	Debug(ctx context.Context, message string, info map[string]any)
	Info(ctx context.Context, message string, info map[string]any)
	Warn(ctx context.Context, message string, info map[string]any)
	Error(ctx context.Context, message string, info map[string]any)
}

type Log struct {
	logger Logger
}

func (l *Log) Debug(ctx context.Context, message string, info map[string]any) {
	if l.logger != nil {
		l.logger.Debug(ctx, message, info)
	}
}

func (l *Log) Info(ctx context.Context, message string, info map[string]any) {
	if l.logger != nil {
		l.logger.Info(ctx, message, info)
	}
}

func (l *Log) Warn(ctx context.Context, message string, info map[string]any) {
	if l.logger != nil {
		l.logger.Warn(ctx, message, info)
	}
}

func (l *Log) Error(ctx context.Context, message string, info map[string]any) {
	if l.logger != nil {
		l.logger.Error(ctx, message, info)
	}
}
