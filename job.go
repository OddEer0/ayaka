package ayaka

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/OddEer0/eelog"
	"github.com/OddEer0/errx"
)

const (
	LogKeyJobName    = "job_name"
	LogKeyPanicValue = "job_panic"

	LogInfoJobInitStarted                 = "job init started"
	LogInfoJobInitFinished                = "job init finished"
	LogErrorJobInitFailure                = "job init failed"
	LogErrorJobInitPanic                  = "job init panic"
	LogInfoJobRunStarted                  = "job run stated"
	LogInfoJobRunFinished                 = "job run finished"
	LogErrorJobRunFailure                 = "job run failed"
	LogErrorJobRunPanic                   = "job run panic"
	LogWarnInitJobsGracefulShutdownFailed = "job init graceful shutdown failed"
	LogWarnRunJobsGracefulShutdownFailed  = "job run graceful shutdown failed"
)

func (a *App[T]) initJobs(ctx context.Context, order []string) error {
	ctx, cancel := context.WithTimeout(ctx, a.startTimeout)
	defer cancel()
	wg := &sync.WaitGroup{}

	stopChan := make(chan struct{})
	sErr := newSingleError()

	for _, name := range order {
		wg.Add(1)
		job := a.jobs[name]

		go initJob(ctx, job, wg, sErr, a.Logger(), cancel, a.hooks, a.Container())
	}

	go func() {
		wg.Wait()
		close(stopChan)
	}()

	select {
	case <-stopChan:
		return sErr.get()
	case <-ctx.Done():
		t := time.NewTimer(a.gracefulTimeout)
		defer t.Stop()
		select {
		case <-t.C:
			a.logger.Warn(ctx, LogWarnInitJobsGracefulShutdownFailed)
			return ErrGracefulTimeout
		case <-stopChan:
			if errx.Is(ctx.Err(), context.Canceled) {
				return sErr.get()
			}
			return ctx.Err()
		}
	}
}

func initJob[T any](
	ctx context.Context, job Job[T],
	wg *sync.WaitGroup,
	sErr *singleError,
	logger eelog.Logger,
	cancel context.CancelFunc,
	hooks Hooks[T],
	container T,
) {
	defer wg.Done()
	defer func() {
		if r := recover(); r != nil {
			logger.Error(ctx, LogErrorJobInitPanic,
				eelog.String(LogKeyJobName, job.Name()),
				eelog.Any(LogKeyPanicValue, r),
			)
			sErr.add(
				fmt.Errorf(
					`"%s": %w: %v`,
					job.Name(), ErrJobInitPanic, r,
				),
				cancel,
			)
		}
	}()

	logger.Info(ctx, LogInfoJobInitStarted,
		eelog.String(LogKeyJobName, job.Name()))

	if hooks.BeforeInit != nil {
		hooks.BeforeInit(ctx, job)
	}

	err := job.Init(ctx, container)

	if hooks.AfterInit != nil {
		hooks.AfterInit(ctx, job, err)
	}

	if err != nil {
		logger.Error(ctx, LogErrorJobInitFailure,
			eelog.String(LogKeyJobName, job.Name()),
			eelog.Err(err),
		)

		sErr.add(
			errx.Wrapf(
				errors.Join(ErrJobInitFailed, err),
				`"%s"`,
				job.Name(),
			),
			cancel,
		)
		return
	}

	logger.Info(ctx, LogInfoJobInitFinished,
		eelog.String(LogKeyJobName, job.Name()))
}

func (a *App[T]) runJobs(ctx context.Context, order []string) error {
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sErr := newSingleError()
	stopChan := make(chan struct{})

	for _, name := range order {
		wg.Add(1)
		job := a.jobs[name]

		go runJob(ctx, job, wg, sErr, cancel, a.Logger(), a.hooks, a.Container())
	}

	go func() {
		wg.Wait()
		close(stopChan)
	}()

	select {
	case <-stopChan:
		return sErr.get()
	case <-ctx.Done():
		t := time.NewTimer(a.gracefulTimeout)
		defer t.Stop()
		select {
		case <-t.C:
			a.logger.Warn(a.ctx, LogWarnRunJobsGracefulShutdownFailed)
			return ErrGracefulTimeout
		case <-stopChan:
			if errx.Is(ctx.Err(), context.Canceled) {
				return sErr.get()
			}
			return ctx.Err()
		}
	}
}

func runJob[T any](
	ctx context.Context, job Job[T],
	wg *sync.WaitGroup,
	sErr *singleError,
	cancel context.CancelFunc,
	logger eelog.Logger,
	hooks Hooks[T],
	container T,
) {
	defer wg.Done()
	defer func() {
		if r := recover(); r != nil {
			logger.Error(ctx, LogErrorJobRunPanic,
				eelog.String(LogKeyJobName, job.Name()),
				eelog.Any(LogKeyPanicValue, r),
			)

			sErr.add(
				fmt.Errorf(
					`"%s": %w: %v`,
					job.Name(), ErrJobRunPanic, r,
				),
				cancel,
			)
		}
	}()

	logger.Info(ctx, LogInfoJobRunStarted,
		eelog.String(LogKeyJobName, job.Name()))

	if hooks.BeforeRun != nil {
		hooks.BeforeRun(ctx, job)
	}

	err := job.Run(ctx, container)

	if hooks.AfterRun != nil {
		hooks.AfterRun(ctx, job, err)
	}

	if err != nil {
		logger.Error(ctx, LogErrorJobRunFailure,
			eelog.String(LogKeyJobName, job.Name()),
			eelog.Err(err),
		)

		sErr.add(
			errx.Wrapf(
				errors.Join(ErrJobRunFailed, err),
				`"%s"`,
				job.Name(),
			),
			cancel,
		)
		return
	}

	logger.Info(ctx, LogInfoJobRunFinished,
		eelog.String(LogKeyJobName, job.Name()))
}
