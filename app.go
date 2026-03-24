package ayaka

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/OddEer0/eelog"
	"github.com/OddEer0/errx"
)

type (
	Job[T any] interface {
		Name() string
		Init(ctx context.Context, container T) error
		Run(ctx context.Context, container T) error
	}

	Info struct {
		Name, Version, Description string
	}

	Hooks[T any] struct {
		BeforeInit func(ctx context.Context, job Job[T])
		AfterInit  func(ctx context.Context, job Job[T], err error)

		BeforeRun func(ctx context.Context, job Job[T])
		AfterRun  func(ctx context.Context, job Job[T], err error)
	}

	Options[T any] struct {
		Name, Description, Version    string
		StartTimeout, GracefulTimeout time.Duration
		Container                     T
		Logger                        eelog.Logger
		Hooks                         Hooks[T]
	}

	App[T any] struct {
		info Info

		jobs  map[string]Job[T]
		order []string

		container T
		logger    eelog.Logger

		startTimeout    time.Duration
		gracefulTimeout time.Duration

		ctx    context.Context
		cancel context.CancelFunc

		hooks Hooks[T]

		err error
	}
)

const (
	DefaultStartTimeout        = time.Second * 3
	DefaultGracefulTimeout     = time.Second * 10
	LogInfoInitAllJobsStarted  = "init all jobs started"
	LogInfoInitAllJobsFinished = "init all jobs finished"
	LogInfoRunAllJobsStarted   = "run all jobs started"
	LogInfoRunAllJobsFinished  = "run all jobs finished"
	LogWarnJobAlreadyExists    = "job already exists"

	defaultCapacity = 16
)

func NewApp[T any](opt Options[T]) *App[T] {
	var errRes error
	opt = opt.DefaultValue()
	err := opt.Validate()
	if err != nil {
		errRes = err
	}

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	result := &App[T]{
		info: Info{
			Name:        opt.Name,
			Description: opt.Description,
			Version:     opt.Version,
		},
		jobs:  make(map[string]Job[T], defaultCapacity),
		order: make([]string, 0, defaultCapacity),
		err:   errRes,

		startTimeout:    opt.StartTimeout,
		gracefulTimeout: opt.GracefulTimeout,

		container: opt.Container,
		logger:    opt.Logger,

		ctx:    ctx,
		cancel: cancel,

		hooks: opt.Hooks,
	}

	result.ctx = AppWithContext(ctx, result)

	return result
}

func (o Options[T]) DefaultValue() Options[T] {
	if o.Logger == nil {
		o.Logger = eelog.NoopLogger{}
	}
	if o.StartTimeout == 0 {
		o.StartTimeout = DefaultStartTimeout
	}
	if o.GracefulTimeout == 0 {
		o.GracefulTimeout = DefaultGracefulTimeout
	}

	return o
}

func (o Options[T]) Validate() error {
	if o.Name == "" ||
		o.Version == "" ||
		o.Description == "" {
		return ErrInvalidArgument
	}
	return nil
}

func (a *App[T]) Info() Info {
	return a.info
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

func (a *App[T]) Logger() eelog.Logger {
	return a.logger
}

func (a *App[T]) StartTimeout() time.Duration {
	return a.startTimeout
}

func (a *App[T]) GracefulTimeout() time.Duration {
	return a.gracefulTimeout
}

func (a *App[T]) Start() error {
	if a.err != nil {
		return a.err
	}
	if len(a.jobs) == 0 {
		a.err = ErrNoJobs
		return a.err
	}

	a.logger.Info(a.ctx, LogInfoInitAllJobsStarted)
	err := a.initJobs(a.ctx, a.order)
	if err != nil {
		return errx.Wrap(err, "[App] initJob")
	}
	a.logger.Info(a.ctx, LogInfoInitAllJobsFinished)

	a.logger.Info(a.ctx, LogInfoRunAllJobsStarted)
	err = a.runJobs(a.ctx, a.order)
	if err != nil {
		return errx.Wrap(err, "[App] runJob")
	}
	a.logger.Info(a.ctx, LogInfoRunAllJobsFinished)
	return nil
}

func (a *App[T]) WithJob(jobs ...Job[T]) *App[T] {
	if a.err != nil {
		return a
	}

	for _, job := range jobs {
		if _, ok := a.jobs[job.Name()]; ok {
			a.logger.Warn(a.ctx, LogWarnJobAlreadyExists,
				eelog.String(LogKeyJobName, job.Name()))
			continue
		}
		a.jobs[job.Name()] = job
		a.order = append(a.order, job.Name())
	}

	return a
}
