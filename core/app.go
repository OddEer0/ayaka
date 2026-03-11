package ayaka

import (
	"context"
	"errors"
)

type (
	Job[T any] interface {
		Init(ctx context.Context, container T) error
		Run(ctx context.Context, container T) error
	}

	JobEntry[T any] struct {
		Key string
		Job Job[T]
	}

	Info struct {
		Name, Version, Description string
	}

	ConfigInterceptor func(ctx context.Context, conf *Config) (*Config, error)

	App[T any] struct {
		info   Info
		config *Config
		jobs   map[string]Job[T]
		err    error
		ctx    context.Context

		container T

		configInterceptor ConfigInterceptor
		logger            Logger
	}

	ReadonlyApp[T any] struct {
		app *App[T]
	}
)

var (
	ErrIsInvalidArgument = errors.New("invallid argument")
)

func (a *App[T]) Info() Info {
	return a.info
}

func (a *App[T]) Config() *Config {
	return a.config
}

func (a *App[T]) Err() error {
	return a.err
}

func (a *App[T]) Container() T {
	return a.container
}

func (a *App[T]) Context() context.Context {
	return a.ctx
}

func (a *App[T]) Logger() Logger {
	return a.logger
}

func (r *ReadonlyApp[T]) Info() Info {
	return r.app.Info()
}

func (r *ReadonlyApp[T]) Context() context.Context {
	return r.app.Context()
}

func (r *ReadonlyApp[T]) Config() any {
	return r.app.Config()
}

func (r *ReadonlyApp[T]) Err() error {
	return r.app.Err()
}

func (r *ReadonlyApp[T]) Container() T {
	return r.app.Container()
}

func (r *ReadonlyApp[T]) Logger() Logger {
	return r.app.Logger()
}

type Options[T any] struct {
	Name, Description, Version string
	ConfigInterceptor          ConfigInterceptor
	Logger                     Logger
	Container                  T
}

func (o Options[T]) Validate() error {
	if o.Name == "" ||
		o.Version == "" ||
		o.Description == "" {
		return ErrIsInvalidArgument
	}
	return nil
}

func NewApp[T any](opt *Options[T]) *App[T] {
	var errRes error
	err := opt.Validate()
	if err != nil {
		errRes = err
	}
	var log Logger = NoopLogger{}
	if opt.Logger != nil {
		log = opt.Logger
	}

	result := &App[T]{
		info: Info{
			Name:        opt.Name,
			Description: opt.Description,
			Version:     opt.Version,
		},
		config: &Config{},
		jobs:   make(map[string]Job[T]),
		err:    errRes,

		container: opt.Container,

		configInterceptor: opt.ConfigInterceptor,
		logger:            log,
	}

	result.ctx = AppWithContext(context.Background(), result)

	return result
}
