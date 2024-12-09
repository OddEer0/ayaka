package ecosystem

import (
	"context"
	"github.com/rs/zerolog"
)

type ZerologLogger struct {
	logger *zerolog.Logger
}

func (z *ZerologLogger) Debug(ctx context.Context, message string, info map[string]any) {
	l := z.logger.Debug().Ctx(ctx)
	for k, v := range info {
		l = l.Interface(k, v)
	}
	l.Msg(message)
}

func (z *ZerologLogger) Info(ctx context.Context, message string, info map[string]any) {
	l := z.logger.Info().Ctx(ctx)
	for k, v := range info {
		l = l.Interface(k, v)
	}
	l.Msg(message)
}

func (z *ZerologLogger) Warn(ctx context.Context, message string, info map[string]any) {
	l := z.logger.Warn().Ctx(ctx)
	for k, v := range info {
		l = l.Interface(k, v)
	}
	l.Msg(message)
}

func (z *ZerologLogger) Error(ctx context.Context, message string, info map[string]any) {
	l := z.logger.Error().Ctx(ctx)
	for k, v := range info {
		l = l.Interface(k, v)
	}
	l.Msg(message)
}

func NewAppLoggerWithZerolog(logger *zerolog.Logger) *ZerologLogger {
	if logger == nil {
		logger = zerolog.DefaultContextLogger
	}
	return &ZerologLogger{logger: logger}
}
