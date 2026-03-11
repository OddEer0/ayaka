package ayaka

import (
	"context"
	"fmt"
	"sync"
	"time"

	"errors"
)

const (
	LogKeyInfoKey                    = "job_key"
	LogKeyInfoError                  = "job_error"
	LogKeyInfoPanic                  = "job_panic"
	LogMessageInitError              = "job init failed"
	LogMessageInitPanic              = "job init panic"
	LogMessageRunError               = "job run failed"
	LogMessageRunPanic               = "job run panic"
	LogMessageGracefulShutdownFailed = "graceful shutdown failed"
	FormatErrJobInitFailed           = "failed to initialize job '%s': %w"
	FormatErrJobInitPanic            = "panic in initialized job '%s': %v"
	FormatErrJobRunFailed            = "failed run job '%s': %w"
	FormatErrJobRunPanic             = "panic in runned job '%s': %v"
)

func (a *App[T]) initJob() error {
	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(a.ctx, a.Config().StartTimeout)
	defer cancel()

	sErr := newSingleError(cancel)
	stopChan := make(chan struct{})

	for key, job := range a.jobs {
		wg.Add(1)
		go func(ctx context.Context, key string, job Job[T]) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					a.logger.Error(ctx, LogMessageInitPanic, map[string]any{
						LogKeyInfoKey:   key,
						LogKeyInfoPanic: r,
					})
					sErr.add(fmt.Errorf(FormatErrJobInitPanic, key, r))
				}
			}()

			if err := job.Init(ctx, a.Container()); err != nil {
				a.logger.Error(ctx, LogMessageInitError, map[string]any{
					LogKeyInfoKey:   key,
					LogKeyInfoError: err.Error(),
				})
				sErr.add(fmt.Errorf(FormatErrJobInitFailed, key, err))
			}
		}(ctx, key, job)
	}

	go func() {
		wg.Wait()
		close(stopChan)
	}()

	select {
	case <-stopChan:
		return sErr.get()
	case <-ctx.Done():
		t := time.NewTimer(a.Config().GracefulTimeout)
		select {
		case <-t.C:
			a.logger.Warn(a.ctx, LogMessageGracefulShutdownFailed, nil)
			return ErrGracefulTimeout
		case <-stopChan:
			if errors.Is(ctx.Err(), context.Canceled) {
				return sErr.get()
			}
			return ctx.Err()
		}
	}
}

func (a *App[T]) runJob() error {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(a.ctx)
	defer cancel()

	sErr := newSingleError(cancel)
	stopChan := make(chan struct{})

	for key, job := range a.jobs {
		wg.Add(1)
		go func(ctx context.Context, key string, job Job[T]) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					a.logger.Error(ctx, LogMessageRunPanic, map[string]any{
						LogKeyInfoKey:   key,
						LogKeyInfoPanic: r,
					})
					sErr.add(fmt.Errorf(FormatErrJobRunPanic, key, r))
				}
			}()

			if err := job.Run(ctx, a.Container()); err != nil {
				a.logger.Error(ctx, LogMessageRunError, map[string]any{
					LogKeyInfoKey:   key,
					LogKeyInfoError: err.Error(),
				})
				sErr.add(fmt.Errorf(FormatErrJobRunFailed, key, err))
			}
		}(ctx, key, job)
	}

	go func() {
		wg.Wait()
		close(stopChan)
	}()

	select {
	case <-stopChan:
		return sErr.get()
	case <-ctx.Done():
		t := time.NewTimer(a.Config().GracefulTimeout)
		select {
		case <-t.C:
			a.logger.Warn(a.ctx, LogMessageGracefulShutdownFailed, nil)
			return ErrGracefulTimeout
		case <-stopChan:
			if errors.Is(ctx.Err(), context.Canceled) {
				return sErr.get()
			}
			return ctx.Err()
		}
	}
}
