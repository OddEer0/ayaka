package ayaka

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"sync"
	"time"
)

func (a *App) initJob(ctx context.Context) error {
	var wg sync.WaitGroup
	ctxStop, cancel := context.WithCancel(ctx)
	defer cancel()

	sErr := newSingleError(func() {
		cancel()
	})
	stopChan := make(chan struct{})

	for key, job := range a.jobs {
		wg.Add(1)
		go func(ctx context.Context, key string, job Job) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					a.logger.Error(ctx, "job init panic", map[string]any{
						"job_key":     key,
						"job_recover": r,
					})
					sErr.add(fmt.Errorf("panic in initialized job '%s': %v", key, r))
				}
			}()

			if err := job.Init(ctxStop, a.Dependency()); err != nil {
				a.logger.Error(ctx, "job init failed", map[string]any{
					"job_key":   key,
					"job_error": err.Error(),
				})
				sErr.add(fmt.Errorf("failed to initialize job '%s': %w", key, err))
			}
		}(a.ctx, key, job)
	}

	go func() {
		wg.Wait()
		close(stopChan)
	}()

	select {
	case <-stopChan:
		return errors.Wrap(sErr.get(), "[App.initJob]")
	case <-ctx.Done():
		t := time.NewTimer(a.Config().GracefulTimeout)
		select {
		case <-t.C:
			a.logger.Warn(a.ctx, "graceful shotdown failed", nil)
			return errors.Wrap(ErrGracefulTimeout, "[App.InitJob]")
		case <-stopChan:
			return errors.Wrap(ctx.Err(), "[App.InitJob]")
		}
	}
}

func (a *App) runJob(ctx context.Context) error {
	var wg sync.WaitGroup
	ctxStop, cancel := context.WithCancel(ctx)
	defer cancel()

	sErr := newSingleError(func() {
		cancel()
	})
	stopChan := make(chan struct{})

	for key, job := range a.jobs {
		wg.Add(1)
		go func(ctx context.Context, key string, job Job) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					a.logger.Error(ctx, "job run panic", map[string]any{
						"job_key":     key,
						"job_recover": r,
					})
					sErr.add(fmt.Errorf("panic in runned job '%s': %v", key, r))
				}
			}()

			if err := job.Run(ctxStop, a.Dependency()); err != nil {
				a.logger.Error(ctx, "job run failed", map[string]any{
					"job_key":   key,
					"job_error": err.Error(),
				})
				sErr.add(fmt.Errorf("failed run job '%s': %w", key, err))
			}
		}(a.ctx, key, job)
	}

	go func() {
		wg.Wait()
		close(stopChan)
	}()

	select {
	case <-stopChan:
		return errors.Wrap(sErr.get(), "[App.RunJob]")
	case <-ctx.Done():
		t := time.NewTimer(a.Config().GracefulTimeout)
		select {
		case <-t.C:
			a.logger.Warn(a.ctx, "graceful shotdown failed", nil)
			return errors.Wrap(ErrGracefulTimeout, "[App.RunJob]")
		case <-stopChan:
			return errors.Wrap(ctx.Err(), "[App.RunJob]")
		}
	}
}
