package ayaka

import (
	"context"
	validation "github.com/go-ozzo/ozzo-validation"
	"go.uber.org/dig"
)

type (
	Job interface {
		Init(ctx context.Context, container *dig.Container) error
		Run(ctx context.Context, container *dig.Container) error
	}

	JobEntry struct {
		Key string
		Job Job
	}

	Info struct {
		Name, Version, Description string
	}

	ConfigInterceptor func(ctx context.Context, conf *Config) (*Config, error)

	App struct {
		info   Info
		config *Config
		jobs   map[string]Job
		err    error
		ctx    context.Context

		di *dig.Container

		configInterceptor ConfigInterceptor
		logger            *Log
	}

	ReadonlyApp struct {
		app *App
	}
)

func (a *App) Info() Info {
	return a.info
}

func (a *App) Config() *Config {
	return a.config
}

func (a *App) Err() error {
	return a.err
}

func (a *App) Dependency() *dig.Container {
	return a.di
}

func (a *App) Context() context.Context {
	return a.ctx
}

func (r *ReadonlyApp) Info() Info {
	return r.app.Info()
}

func (r *ReadonlyApp) Context() context.Context {
	return r.app.Context()
}

func (r *ReadonlyApp) Config() any {
	return r.app.Config()
}

func (r *ReadonlyApp) Err() error {
	return r.app.Err()
}

func (r *ReadonlyApp) Dependency() *dig.Container {
	return r.app.Dependency()
}

type Options struct {
	Name, Description, Version string
	CoreConfigInterceptor      ConfigInterceptor
	Logger                     Logger
	Container                  *dig.Container
}

func (o Options) Validate() error {
	return validation.ValidateStruct(&o,
		validation.Field(&o.Name, validation.Required),
		validation.Field(&o.Description, validation.Required),
		validation.Field(&o.Version, validation.Required),
	)
}

func NewApp(opt *Options) *App {
	var errRes error
	err := opt.Validate()
	if err != nil {
		errRes = err
	}
	log := &Log{logger: opt.Logger}

	di := opt.Container
	if di == nil {
		di = dig.New()
	}
	err = di.Provide(func() Logger { return log })
	if err != nil {
		errRes = err
	}

	result := &App{
		info: Info{
			Name:        opt.Name,
			Description: opt.Description,
			Version:     opt.Version,
		},
		config: &Config{},
		jobs:   make(map[string]Job),
		err:    errRes,

		di: di,

		configInterceptor: opt.CoreConfigInterceptor,
		logger:            log,
	}

	result.ctx = AppWithContext(context.Background(), result)

	return result
}
